package entities

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "CASH"
	PaymentMethodCard   PaymentMethod = "CARD"
	PaymentMethodWallet PaymentMethod = "WALLET"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusCompleted  PaymentStatus = "COMPLETED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	PassengerID uuid.UUID
	Amount      float64
	Currency    string
	Method      PaymentMethod
	Status      PaymentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TripRating struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	RaterID     uuid.UUID
	RatedUserID uuid.UUID
	Rating      int
	Comment     string
	CreatedAt   time.Time
}
