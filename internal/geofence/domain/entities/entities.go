package entities

import (
	"time"

	"github.com/google/uuid"
)

type GeofenceType string

const (
	GeofenceCity       GeofenceType = "CITY"
	GeofenceAirport    GeofenceType = "AIRPORT"
	GeofenceRestricted GeofenceType = "RESTRICTED"
	GeofenceSurge      GeofenceType = "SURGE_ZONE"
)

type Geofence struct {
	ID          uuid.UUID
	Name        string
	Type        GeofenceType
	Coordinates string // WKT format: "((lng1 lat1), (lng2 lat2), ...)"
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ZoneAssignment struct {
	ID        uuid.UUID
	DriverID  uuid.UUID
	ZoneID    uuid.UUID
	Status    string
	CreatedAt time.Time
}
