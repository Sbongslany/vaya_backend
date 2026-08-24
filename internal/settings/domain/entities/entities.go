package entities

import (
	"time"

	"github.com/google/uuid"
)

type VehicleType struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	BaseFare   float64
	PerKmRate  float64
	PerMinRate float64
	IsActive   bool
	CreatedAt  time.Time
}

type PlatformSettings struct {
	ID                   uuid.UUID
	CommissionPercentage float64
	CancellationFee      float64
	UpdatedAt            time.Time
}
