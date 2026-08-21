package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/driver/domain"
	"github.com/yourorg/ehailing/backend/internal/driver/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/driver/domain/repositories"
)

type GoOnline struct {
	stateRepo repositories.DriverStateRepository
}

func NewGoOnline(stateRepo repositories.DriverStateRepository) *GoOnline {
	return &GoOnline{stateRepo: stateRepo}
}

func (uc *GoOnline) Execute(ctx context.Context, driverID string) error {
	status, err := uc.stateRepo.GetStatus(ctx, driverID)
	if err != nil {
		return err
	}
	if status == entities.DriverStatusOnline || status == entities.DriverStatusBusy {
		return domain.ErrAlreadyOnline
	}

	return uc.stateRepo.SetStatus(ctx, driverID, entities.DriverStatusOnline)
}