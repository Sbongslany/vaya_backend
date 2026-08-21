package entities

import (
	"time"

	"github.com/google/uuid"
)

type TicketCategory string

const (
	CategoryLostItem    TicketCategory = "LOST_ITEM"
	CategoryOvercharged TicketCategory = "OVERCHARGED"
	CategoryNoShow      TicketCategory = "DRIVER_NO_SHOW"
	CategoryUnsafe      TicketCategory = "UNSAFE_DRIVING"
	CategoryOther       TicketCategory = "OTHER"
)

type TicketStatus string

const (
	StatusOpen       TicketStatus = "OPEN"
	StatusInProgress TicketStatus = "IN_PROGRESS"
	StatusResolved   TicketStatus = "RESOLVED"
	StatusClosed     TicketStatus = "CLOSED"
)

type SupportTicket struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TripID        *uuid.UUID
	Category      TicketCategory
	Subject       string
	Description   string
	Status        TicketStatus
	AdminAssigned *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TicketComment struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
}

type Refund struct {
	ID          uuid.UUID
	TicketID    uuid.UUID
	UserID      uuid.UUID
	Amount      float64
	Reason      string
	ProcessedBy *uuid.UUID
	CreatedAt   time.Time
}
