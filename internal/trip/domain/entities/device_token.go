package entities

import (
	"time"

	"github.com/google/uuid"
)

type DeviceToken struct {
	UserID     uuid.UUID
	Token      string
	DeviceType string // "IOS" or "ANDROID"
	CreatedAt  time.Time
}
