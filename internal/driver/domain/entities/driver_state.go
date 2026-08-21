package entities

import "time"

type DriverStatus string

const (
	DriverStatusOffline DriverStatus = "OFFLINE"
	DriverStatusOnline  DriverStatus = "ONLINE"
	DriverStatusBusy    DriverStatus = "BUSY" // On a trip
)

type DriverState struct {
	DriverID  string
	Status    DriverStatus
	UpdatedAt time.Time
}
