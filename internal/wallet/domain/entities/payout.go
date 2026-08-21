package entities

import (
	"time"

	"github.com/google/uuid"
)

type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "PENDING"
	PayoutStatusProcessing PayoutStatus = "PROCESSING"
	PayoutStatusCompleted  PayoutStatus = "COMPLETED"
	PayoutStatusFailed     PayoutStatus = "FAILED"
	PayoutStatusCancelled  PayoutStatus = "CANCELLED"
)

type Payout struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Amount              float64
	Currency            string
	Status              PayoutStatus
	BankName            string
	BankAccountNumber   string
	BankAccountName     string
	PaystackTransferRef *string
	PaystackTransferID  *string
	FailureReason       *string
	RequestedAt         time.Time
	ProcessedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
