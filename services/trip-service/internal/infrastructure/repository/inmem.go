package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, rideFare *domain.RideFareModel) error {
	r.rideFares[rideFare.ID.Hex()] = rideFare
	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, fareID string) (*domain.RideFareModel, error) {
	rideFare, ok := r.rideFares[fareID]
	if !ok {
		return nil, fmt.Errorf("fare does not exist with ID: %s", fareID)
	}

	return rideFare, nil
}
