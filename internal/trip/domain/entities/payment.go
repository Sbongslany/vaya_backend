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
	ID                       uuid.UUID
	TripID                   uuid.UUID
	PassengerID              uuid.UUID
	Amount                   float64
	Currency                 string
	Method                   PaymentMethod
	Status                   PaymentStatus
	PaystackReference        *string
	PaystackAuthorizationURL *string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}
