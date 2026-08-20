package services

import (
	"fmt"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type StateMachine struct{}

func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

var ValidTransitions = map[entities.TripStatus][]entities.TripStatus{
	// Normal trip flow
	entities.StatusRequested:         {entities.StatusOffersReceived, entities.StatusDriverAssigned, entities.StatusQuoteGenerated, entities.StatusCancelledByPassenger, entities.StatusCancelledBySystem},
	entities.StatusOffersReceived:    {entities.StatusDriverAssigned, entities.StatusDriverSelected, entities.StatusCancelledByPassenger, entities.StatusCancelledBySystem},
	entities.StatusDriverAssigned:    {entities.StatusDriverEnRoute, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver, entities.StatusCancelledBySystem},
	entities.StatusDriverEnRoute:     {entities.StatusDriverArrived, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver, entities.StatusCancelledBySystem},
	entities.StatusDriverArrived:     {entities.StatusTripStartPending, entities.StatusTripInProgress, entities.StatusTripStarted, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver},
	entities.StatusTripStartPending:  {entities.StatusTripStarted, entities.StatusTripInProgress},
	entities.StatusTripStarted:       {entities.StatusTripInProgress, entities.StatusOutboundInProgress},
	entities.StatusTripInProgress:    {entities.StatusArrivedAtDest, entities.StatusTripCompleted},
	entities.StatusArrivedAtDest:     {entities.StatusTripCompleted},
	entities.StatusTripCompleted:     {entities.StatusPaymentProcessing},
	entities.StatusPaymentProcessing: {entities.StatusPaymentCompleted},
	entities.StatusPaymentCompleted:  {entities.StatusRatingPending},
	entities.StatusRatingPending:     {entities.StatusClosed},
	// Long-distance trip flow
	entities.StatusQuoteGenerated:     {entities.StatusSearchingDrivers, entities.StatusCancelledByPassenger},
	entities.StatusSearchingDrivers:   {entities.StatusOffersReceived, entities.StatusCancelledByPassenger},
	entities.StatusDriverSelected:     {entities.StatusDriverConfirmed, entities.StatusDriverAssigned, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver},
	entities.StatusDriverConfirmed:    {entities.StatusScheduled, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver},
	entities.StatusScheduled:          {entities.StatusDriverEnRoute, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver},
	entities.StatusOutboundInProgress: {entities.StatusDestinationReached},
	entities.StatusDestinationReached: {entities.StatusTripCompleted, entities.StatusDriverRetained},
	entities.StatusDriverRetained:     {entities.StatusReturnScheduled},
	entities.StatusReturnScheduled:    {entities.StatusReturnStarted},
	entities.StatusReturnStarted:      {entities.StatusReturnInProgress},
	entities.StatusReturnInProgress:   {entities.StatusFinalDestination},
	entities.StatusFinalDestination:   {entities.StatusTripCompleted},
}

func (sm *StateMachine) CanTransition(current, next entities.TripStatus) bool {
	allowed, exists := ValidTransitions[current]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}

func (sm *StateMachine) Transition(current, next entities.TripStatus) error {
	if !sm.CanTransition(current, next) {
		return fmt.Errorf("cannot transition from %s to %s", current, next)
	}
	return nil
}