package redis

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
)

type DriverStateManagerAdapter struct {
	repo *DriverStateRepository
}

func NewDriverStateManagerAdapter(repo *DriverStateRepository) *DriverStateManagerAdapter {
	return &DriverStateManagerAdapter{repo: repo}
}

func (a *DriverStateManagerAdapter) MarkBusy(ctx context.Context, driverID string) error {
	return a.repo.SetStatus(ctx, driverID, entities.DriverStatusBusy)
}

func (a *DriverStateManagerAdapter) MarkOnline(ctx context.Context, driverID string) error {
	return a.repo.SetStatus(ctx, driverID, entities.DriverStatusOnline)
}
