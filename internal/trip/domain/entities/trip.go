package entities

import (
	"time"
	"github.com/google/uuid"
)

// --- ENUMS & CONSTANTS ---

type TripType string
const (
	TripTypeNormal       TripType = "NORMAL"
	TripTypeLongDistance TripType = "LONG_DISTANCE"
)

type TripStatus string
const (
	StatusRequested          TripStatus = "REQUESTED"
	StatusSearchingDrivers   TripStatus = "SEARCHING_DRIVERS"
	StatusOffersReceived     TripStatus = "OFFERS_RECEIVED"
	StatusDriverSelected     TripStatus = "DRIVER_SELECTED"
	StatusDriverAssigned     TripStatus = "DRIVER_ASSIGNED"
	StatusDriverEnRoute      TripStatus = "DRIVER_EN_ROUTE"
	StatusDriverArrived      TripStatus = "DRIVER_ARRIVED"
	StatusTripStartPending   TripStatus = "TRIP_START_PENDING"
	StatusTripStarted        TripStatus = "TRIP_STARTED"
	StatusTripInProgress     TripStatus = "TRIP_IN_PROGRESS"
	StatusArrivedAtDest      TripStatus = "ARRIVED_AT_DESTINATION"
	StatusTripCompleted      TripStatus = "TRIP_COMPLETED"
	StatusPaymentProcessing  TripStatus = "PAYMENT_PROCESSING"
	StatusPaymentCompleted   TripStatus = "PAYMENT_COMPLETED"
	StatusRatingPending      TripStatus = "RATING_PENDING"
	StatusClosed             TripStatus = "CLOSED"
	StatusCancelledByPassenger TripStatus = "CANCELLED_BY_PASSENGER"
	StatusCancelledByDriver    TripStatus = "CANCELLED_BY_DRIVER"
	StatusCancelledBySystem    TripStatus = "CANCELLED_BY_SYSTEM"
)

type OfferType string
const (
	OfferTypeNormalFare OfferType = "NORMAL_FARE"
	OfferTypeOffer      OfferType = "OFFER"
)

type OfferStatus string
const (
	OfferStatusPending  OfferStatus = "PENDING"
	OfferStatusAccepted OfferStatus = "ACCEPTED"
	OfferStatusRejected OfferStatus = "REJECTED"
	OfferStatusExpired  OfferStatus = "EXPIRED"
)

// --- ENTITIES ---

type Trip struct {
	ID               uuid.UUID
	PassengerID      uuid.UUID
	DriverID         *uuid.UUID
	VehicleID        *uuid.UUID
	TripType         TripType
	Status           TripStatus
	PickupLatitude   float64
	PickupLongitude  float64
	PickupAddress    string
	DropoffLatitude  float64
	DropoffLongitude float64
	DropoffAddress   string
	EstimatedFare    float64
	FinalFare        *float64
	Currency         string
	DistanceKM       *float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TripOffer struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	DriverID    uuid.UUID
	OfferType   OfferType
	OfferedFare float64
	Status      OfferStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}