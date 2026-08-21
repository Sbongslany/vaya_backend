package entities

import (
	"time"

	"github.com/google/uuid"
)

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
