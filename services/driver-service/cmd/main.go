package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/driver-service/internal/infrastructure/grpc"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/env"
	"syscall"

	grpcserver "google.golang.org/grpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

var GrpcAddr = ":9092"

func main() {
	RABBITMQ_URI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	svc := service.NewService()

	conn, err := amqp.Dial(RABBITMQ_URI)
	if err != nil {
		log.Fatal("failed to connect to rabbitmq")
		return
	}
	defer conn.Close()

	grpcServer := grpcserver.NewServer()
	grpc.NewGrpcHandler(grpcServer, svc)

	log.Printf("Starting gRPC server driver service on port: %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the server")
	grpcServer.GracefulStop()

}
