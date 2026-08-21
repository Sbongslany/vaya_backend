package entities

import (
	"time"

	"github.com/google/uuid"
)

type DocumentStatus string

const (
	DocStatusPending  DocumentStatus = "PENDING"
	DocStatusApproved DocumentStatus = "APPROVED"
	DocStatusRejected DocumentStatus = "REJECTED"
)

type OnboardingStatus string

const (
	OnboardingNotStarted OnboardingStatus = "NOT_STARTED"
	OnboardingPending    OnboardingStatus = "PENDING_REVIEW"
	OnboardingApproved   OnboardingStatus = "APPROVED"
	OnboardingRejected   OnboardingStatus = "REJECTED"
)

type DriverDocument struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	DocumentType    string
	FileURL         string
	Status          DocumentStatus
	RejectionReason *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DriverKYCSummary struct {
	UserID           uuid.UUID
	Email            string
	OnboardingStatus OnboardingStatus
	TotalDocuments   int
	PendingCount     int
	ApprovedCount    int
	RejectedCount    int
}
