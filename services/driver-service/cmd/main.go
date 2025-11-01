package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/driver-service/internal/infrastructure/grpc"
	driverMessaging "ride-sharing/services/driver-service/internal/infrastructure/messaging"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {
	RABBITMQ_URI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")

	tracerCfg := tracing.Config{
		ServiceName:    "driver-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JaegerEndpoint", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatal("Failed to init tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer sh(ctx)
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

	rabbitmq, err := messaging.NewRabbitMQ(RABBITMQ_URI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpc.NewGrpcHandler(grpcServer, svc)

	consumer := driverMessaging.NewTripConsumer(rabbitmq, svc)

	go func() {
		if err := consumer.Listen(); err != nil {
			log.Printf("Error starting trip consumer: %v", err)
		}
	}()

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
