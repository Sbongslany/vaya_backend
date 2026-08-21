package domain

import "errors"

var (
	ErrDriverNotFound  = errors.New("driver_not_found")
	ErrAlreadyOnline   = errors.New("driver_already_online")
	ErrAlreadyOffline  = errors.New("driver_already_offline")
	ErrInvalidLocation = errors.New("invalid_location_coordinates")
	ErrNoActiveTrip    = errors.New("no_active_trip_for_location_share")
)
