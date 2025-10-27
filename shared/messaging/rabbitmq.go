package messaging

import (
	"context"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %v", uri)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("fail to create channel: %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.SetupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("fail to setup exchanges and queues: %v", err)
	}

	return rmq, nil
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	err := r.Channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if err != nil {
		return fmt.Errorf("failed to set Qos: %v", err)
	}

	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)

			if err := handler(ctx, msg); err != nil {
				log.Fatalf("Error handling message: %v", err)

				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("Error sending Nack: %v", nackErr)
				}

				continue
			}
			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("Error sending Ack: %v", ackErr)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routeKey string, message string) error {
	log.Printf("Publishing mesage with route key %s: %s", routeKey, message)
	return r.Channel.PublishWithContext(ctx,
		TripExchange,       // exchange
		routeKey, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(message),
			DeliveryMode: amqp.Persistent,
		})
}

func (r *RabbitMQ) SetupExchangesAndQueues() error {
	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	if err := r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated,
			contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{contracts.DriverCmdTripRequest},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, messageType := range messageTypes {
		if err := r.Channel.QueueBind(
			q.Name,      // queue name
			messageType, // routing key
			exchange,    // exchange
			false,
			nil,
		); err != nil {
			log.Fatal(err)
			return err
		}
	}

	return nil

}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.Channel != nil {
		r.Channel.Close()
	}
}
