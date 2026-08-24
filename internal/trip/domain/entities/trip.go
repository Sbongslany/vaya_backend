package entities

import (
	"time"

	"github.com/google/uuid"
)

type TripType string

const (
	TripTypeNormal       TripType = "NORMAL"
	TripTypeLongDistance TripType = "LONG_DISTANCE"
)

type TripStatus string

const (
	StatusRequested            TripStatus = "REQUESTED"
	StatusSearchingDrivers     TripStatus = "SEARCHING_DRIVERS"
	StatusOffersReceived       TripStatus = "OFFERS_RECEIVED"
	StatusDriverSelected       TripStatus = "DRIVER_SELECTED"
	StatusDriverAssigned       TripStatus = "DRIVER_ASSIGNED"
	StatusDriverEnRoute        TripStatus = "DRIVER_EN_ROUTE"
	StatusDriverArrived        TripStatus = "DRIVER_ARRIVED"
	StatusTripStartPending     TripStatus = "TRIP_START_PENDING"
	StatusTripStarted          TripStatus = "TRIP_STARTED"
	StatusTripInProgress       TripStatus = "TRIP_IN_PROGRESS"
	StatusArrivedAtDest        TripStatus = "ARRIVED_AT_DESTINATION"
	StatusTripCompleted        TripStatus = "TRIP_COMPLETED"
	StatusPaymentProcessing    TripStatus = "PAYMENT_PROCESSING"
	StatusPaymentCompleted     TripStatus = "PAYMENT_COMPLETED"
	StatusRatingPending        TripStatus = "RATING_PENDING"
	StatusClosed               TripStatus = "CLOSED"
	StatusCancelledByPassenger TripStatus = "CANCELLED_BY_PASSENGER"
	StatusCancelledByDriver    TripStatus = "CANCELLED_BY_DRIVER"
	StatusCancelledBySystem    TripStatus = "CANCELLED_BY_SYSTEM"
	// Long-distance specific
	StatusQuoteGenerated     TripStatus = "QUOTE_GENERATED"
	StatusDriverConfirmed    TripStatus = "DRIVER_CONFIRMED"
	StatusScheduled          TripStatus = "SCHEDULED"
	StatusOutboundInProgress TripStatus = "OUTBOUND_IN_PROGRESS"
	StatusDestinationReached TripStatus = "DESTINATION_REACHED"
	StatusDriverRetained     TripStatus = "DRIVER_RETAINED"
	StatusReturnScheduled    TripStatus = "RETURN_SCHEDULED"
	StatusReturnStarted      TripStatus = "RETURN_STARTED"
	StatusReturnInProgress   TripStatus = "RETURN_IN_PROGRESS"
	StatusFinalDestination   TripStatus = "FINAL_DESTINATION"
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

type LongDistanceType string

const (
	LongDistanceOneWay   LongDistanceType = "ONE_WAY"
	LongDistanceReturn   LongDistanceType = "RETURN"
	LongDistanceMultiDay LongDistanceType = "MULTI_DAY"
)

type Trip struct {
	ID                   uuid.UUID
	PassengerID          uuid.UUID
	DriverID             *uuid.UUID
	VehicleID            *uuid.UUID
	TripType             TripType
	Status               TripStatus
	StartPIN             string
	PickupLatitude       float64
	PickupLongitude      float64
	PickupAddress        string
	DropoffLatitude      float64
	DropoffLongitude     float64
	DropoffAddress       string
	EstimatedFare        float64
	FinalFare            *float64
	Currency             string
	DistanceKM           *float64
	LongDistanceType     *LongDistanceType
	ScheduledDeparture   *time.Time
	ScheduledReturn      *time.Time
	TripDurationDays     *int
	CancellationReason   *string
	CancelledBy          *uuid.UUID
	CancelledAt          *time.Time
	CancellationFee      *float64
	PromotionID          *uuid.UUID
	DiscountAmount       float64
	RoutePolyline        *string
	RouteDurationMinutes *int
	RouteDistanceKM      *float64
	SurgeMultiplier      *float64
	ScheduledPickupTime  *time.Time
	Waypoints            []*Waypoint
	CreatedAt            time.Time
	UpdatedAt            time.Time
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
