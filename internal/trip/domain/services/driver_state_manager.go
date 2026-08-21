package services

import "context"

// DriverStateManager abstracts driver availability so the trip module
// doesn't depend directly on the driver module's Redis implementation.
type DriverStateManager interface {
	MarkBusy(ctx context.Context, driverID string) error
	MarkOnline(ctx context.Context, driverID string) error
}
