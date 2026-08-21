package entities

import (
	"time"

	"github.com/google/uuid"
)

type CallStatus string

const (
	CallStatusInitiated CallStatus = "INITIATED"
	CallStatusRinging   CallStatus = "RINGING"
	CallStatusConnected CallStatus = "CONNECTED"
	CallStatusEnded     CallStatus = "ENDED"
	CallStatusMissed    CallStatus = "MISSED"
	CallStatusDeclined  CallStatus = "DECLINED"
)

type CallSession struct {
	ID              uuid.UUID
	TripID          uuid.UUID
	CallerID        uuid.UUID
	ReceiverID      uuid.UUID
	Status          CallStatus
	StartedAt       *time.Time
	EndedAt         *time.Time
	DurationSeconds int
	CreatedAt       time.Time
}
