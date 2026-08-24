package entities

import (
	"time"

	"github.com/google/uuid"
)

type SOSStatus string

const (
	SOSStatusActive     SOSStatus = "ACTIVE"
	SOSStatusResolved   SOSStatus = "RESOLVED"
	SOSStatusFalseAlarm SOSStatus = "FALSE_ALARM"
)

type SOSAlert struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	TriggeredBy uuid.UUID
	Status      SOSStatus
	TriggeredAt time.Time
	ResolvedAt  *time.Time
	ResolvedBy  *uuid.UUID
}

type TripShareToken struct {
	ID        uuid.UUID
	TripID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SharedTripView is a sanitized DTO returned to unauthenticated users
type SharedTripView struct {
	TripID         uuid.UUID          `json:"trip_id"`
	Status         string             `json:"status"`
	PickupAddress  string             `json:"pickup_address"`
	DropoffAddress string             `json:"dropoff_address"`
	DriverName     string             `json:"driver_name,omitempty"`
	VehiclePlate   string             `json:"vehicle_plate,omitempty"`
	RoutePolyline  *string            `json:"route_polyline"`
	PickupLatLng   map[string]float64 `json:"pickup_lat_lng"`
	DropoffLatLng  map[string]float64 `json:"dropoff_lat_lng"`
}
