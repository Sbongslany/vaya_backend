package dependency

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
)

type FareSplitterAdapter struct {
	splitFareUC *usecases.SplitTripFare
}

func NewFareSplitterAdapter(splitFareUC *usecases.SplitTripFare) *FareSplitterAdapter {
	return &FareSplitterAdapter{splitFareUC: splitFareUC}
}

func (a *FareSplitterAdapter) SplitFare(ctx context.Context, tripID, passengerID, driverID uuid.UUID, fare float64) error {
	return a.splitFareUC.Execute(ctx, usecases.SplitTripFareInput{
		TripID:      tripID,
		PassengerID: passengerID,
		DriverID:    driverID,
		Fare:        fare,
	})
}
