package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	DeviceID         *string
	DeviceType       *string
	DeviceName       *string
	IPAddress        *string
	UserAgent        *string
	MFAVerified      bool
	CreatedAt        time.Time
	LastUsedAt       time.Time
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}