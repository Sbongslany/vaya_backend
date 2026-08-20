package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type ProcessPaymentInput struct {
	TripID      uuid.UUID
	PassengerID uuid.UUID
	Method      entities.PaymentMethod
}

type ProcessPayment struct {
	tripRepo        repositories.TripRepository
	paymentRepo     repositories.PaymentRepository
	paymentProvider services.PaymentProvider
	stateMachine    *services.StateMachine
}

func NewProcessPayment(
	tripRepo repositories.TripRepository,
	paymentRepo repositories.PaymentRepository,
	paymentProvider services.PaymentProvider,
	stateMachine *services.StateMachine,
) *ProcessPayment {
	return &ProcessPayment{
		tripRepo:        tripRepo,
		paymentRepo:     paymentRepo,
		paymentProvider: paymentProvider,
		stateMachine:    stateMachine,
	}
}

func (uc *ProcessPayment) Execute(ctx context.Context, input ProcessPaymentInput) (*entities.Payment, error) {
	trip, err := uc.tripRepo.GetByID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, domain.ErrTripNotFound
	}

	if trip.PassengerID != input.PassengerID {
		return nil, domain.ErrUnauthorized
	}

	if trip.Status != entities.StatusTripCompleted {
		return nil, domain.ErrInvalidStateTransition
	}

	existingPayment, err := uc.paymentRepo.GetByTripID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if existingPayment != nil {
		return nil, domain.ErrAlreadyPaid
	}

	amount := trip.EstimatedFare
	if trip.FinalFare != nil {
		amount = *trip.FinalFare
	}

	now := time.Now()
	payment := &entities.Payment{
		ID:          uuid.New(),
		TripID:      input.TripID,
		PassengerID: input.PassengerID,
		Amount:      amount,
		Currency:    trip.Currency,
		Method:      input.Method,
		Status:      entities.PaymentStatusProcessing,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// TRIP_COMPLETED -> PAYMENT_PROCESSING
	if err := uc.stateMachine.Transition(trip.Status, entities.StatusPaymentProcessing); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusPaymentProcessing); err != nil {
		return nil, err
	}

	// Charge via payment provider
	if err := uc.paymentProvider.Charge(ctx, amount, trip.Currency, input.Method); err != nil {
		_ = uc.paymentRepo.UpdateStatus(ctx, payment.ID, entities.PaymentStatusFailed)
		return nil, err
	}

	if err := uc.paymentRepo.UpdateStatus(ctx, payment.ID, entities.PaymentStatusCompleted); err != nil {
		return nil, err
	}
	payment.Status = entities.PaymentStatusCompleted

	// PAYMENT_PROCESSING -> PAYMENT_COMPLETED
	if err := uc.stateMachine.Transition(entities.StatusPaymentProcessing, entities.StatusPaymentCompleted); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusPaymentCompleted); err != nil {
		return nil, err
	}

	// PAYMENT_COMPLETED -> RATING_PENDING
	if err := uc.stateMachine.Transition(entities.StatusPaymentCompleted, entities.StatusRatingPending); err != nil {
		return nil, domain.ErrInvalidStateTransition
	}
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusRatingPending); err != nil {
		return nil, err
	}

	return payment, nil
}
