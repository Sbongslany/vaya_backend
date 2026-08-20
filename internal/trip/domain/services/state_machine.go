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
	entities.StatusRequested:         {entities.StatusOffersReceived, entities.StatusDriverAssigned, entities.StatusCancelledByPassenger, entities.StatusCancelledBySystem},
	entities.StatusOffersReceived:    {entities.StatusDriverAssigned, entities.StatusCancelledByPassenger, entities.StatusCancelledBySystem},
	entities.StatusDriverAssigned:    {entities.StatusDriverEnRoute, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver, entities.StatusCancelledBySystem},
	entities.StatusDriverEnRoute:     {entities.StatusDriverArrived, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver, entities.StatusCancelledBySystem},
	entities.StatusDriverArrived:     {entities.StatusTripStartPending, entities.StatusCancelledByPassenger, entities.StatusCancelledByDriver},
	entities.StatusTripStartPending:  {entities.StatusTripStarted},
	entities.StatusTripStarted:       {entities.StatusTripInProgress},
	entities.StatusTripInProgress:    {entities.StatusArrivedAtDest},
	entities.StatusArrivedAtDest:     {entities.StatusTripCompleted},
	entities.StatusTripCompleted:     {entities.StatusPaymentProcessing},
	entities.StatusPaymentProcessing: {entities.StatusPaymentCompleted},
	entities.StatusPaymentCompleted:  {entities.StatusRatingPending},
	entities.StatusRatingPending:     {entities.StatusClosed},
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