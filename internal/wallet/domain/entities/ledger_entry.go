package entities

import (
	"time"

	"github.com/google/uuid"
)

type LedgerEntryType string

const (
	LedgerEntryCredit LedgerEntryType = "CREDIT"
	LedgerEntryDebit  LedgerEntryType = "DEBIT"
)

type LedgerReferenceType string

const (
	RefTripFare           LedgerReferenceType = "TRIP_FARE"
	RefPlatformCommission LedgerReferenceType = "PLATFORM_COMMISSION"
	RefAdminTopup         LedgerReferenceType = "ADMIN_TOPUP"
	RefRefund             LedgerReferenceType = "REFUND"
	RefPayout             LedgerReferenceType = "PAYOUT"
	RefPaystackDeposit    LedgerReferenceType = "PAYSTACK_DEPOSIT"
	RefPromotionCredit    LedgerReferenceType = "PROMOTION_CREDIT"
)

type LedgerEntry struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	EntryType     LedgerEntryType
	Amount        float64
	BalanceAfter  float64
	ReferenceType *LedgerReferenceType
	ReferenceID   *uuid.UUID
	Description   string
	CreatedBy     *uuid.UUID
	CreatedAt     time.Time
}
