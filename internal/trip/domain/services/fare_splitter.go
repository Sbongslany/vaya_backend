package services

import (
	"context"

	"github.com/google/uuid"
)

// FareSplitter abstracts the wallet commission engine so the trip module
// doesn't depend directly on the wallet module.
type FareSplitter interface {
	SplitFare(ctx context.Context, tripID, passengerID, driverID uuid.UUID, fare float64) error
}
