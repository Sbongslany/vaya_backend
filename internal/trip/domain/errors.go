package domain

import "errors"

var (
	ErrTripNotFound           = errors.New("trip_not_found")
	ErrInvalidStateTransition = errors.New("invalid_state_transition")
	ErrOfferNotFound          = errors.New("offer_not_found")
	ErrUnauthorized           = errors.New("unauthorized_action")
	ErrInvalidCoordinates     = errors.New("invalid_coordinates")
	ErrActiveTripExists       = errors.New("active_trip_exists")
	ErrInvalidOfferFare       = errors.New("invalid_offer_fare")
)
