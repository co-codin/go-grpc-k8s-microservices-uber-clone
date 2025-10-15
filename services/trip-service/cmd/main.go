package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
)

func main() {
	ctx := context.Background()
	inmemRepo := repository.NewInmemRepository()
	service := service.NewService(inmemRepo)

	fare := &domain.RideFareModel{
		UserID: "11",
	}
	t, err := service.CreateTrip(ctx, fare)
	if err != nil {
		log.Println(err)
	}

	log.Println(t)
}
