package entities

import (
	"time"

	"github.com/google/uuid"
)

// --- Phase 17: Existing Admin Entities ---

type UserRole string

const (
	RolePassenger UserRole = "PASSENGER"
	RoleDriver    UserRole = "DRIVER"
	RoleAdmin     UserRole = "ADMIN"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusBanned    UserStatus = "BANNED"
)

type UserSummary struct {
	ID        uuid.UUID
	Email     string
	Role      UserRole
	Status    UserStatus
	CreatedAt time.Time
}

type TripSummary struct {
	ID            uuid.UUID
	PassengerID   uuid.UUID
	DriverID      *uuid.UUID
	Status        string
	TripType      string
	EstimatedFare float64
	FinalFare     *float64
	CreatedAt     time.Time
}

type PlatformStats struct {
	TotalUsers      int64
	TotalDrivers    int64
	TotalPassengers int64
	TotalTrips      int64
	ActiveTrips     int64
	TotalRevenue    float64
	TotalCommission float64
}

type FinancialSummary struct {
	TotalGrossFare     float64
	TotalCommission    float64
	TotalDriverPayouts float64
	TotalRefunds       float64
}

// --- Phase B: New Live Operations Entities ---

type AdminAuditLog struct {
	ID           uuid.UUID
	AdminID      uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Details      string
	CreatedAt    time.Time
}

type LiveTrip struct {
	ID             uuid.UUID
	Status         string
	PassengerID    uuid.UUID
	DriverID       *uuid.UUID
	PickupAddress  string
	DropoffAddress string
	PickupLat      float64
	PickupLng      float64
	DropoffLat     float64
	DropoffLng     float64
	EstimatedFare  float64
	CreatedAt      time.Time
}

type LiveDriver struct {
	ID        uuid.UUID
	Email     string
	Status    string
	Latitude  float64
	Longitude float64
}

type LiveSOS struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	TriggeredBy uuid.UUID
	Status      string
	TriggeredAt time.Time
}

// --- Phase C: Payout Approval Entities ---

type PayoutSummary struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	Amount        float64
	BankName      string
	AccountNumber string
	Status        string
	CreatedAt     time.Time
}

type PayoutDetails struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	Amount        float64
	BankCode      string
	AccountNumber string
	Status        string
}

type AdminSummary struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Role      string
	Status    string
	CreatedAt time.Time
}
