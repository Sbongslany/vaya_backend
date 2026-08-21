package entities

import "time"

type DriverLocation struct {
	DriverID  string
	Latitude  float64
	Longitude float64
	Heading   float64 // degrees 0-360
	Speed     float64 // m/s
	UpdatedAt time.Time
}