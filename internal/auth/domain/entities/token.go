package entities

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (t *EmailVerificationToken) IsUsed() bool {
	return t.UsedAt != nil
}

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

type MFASecret struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	SecretEncrypted []byte
	Method          string
	IsEnabled       bool
	ConfirmedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}