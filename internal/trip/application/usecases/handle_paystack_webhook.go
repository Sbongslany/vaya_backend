package usecases

import (
	"context"
	"fmt"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
)

type HandlePaystackWebhook struct {
	paymentRepo     repositories.PaymentRepository
	tripRepo        repositories.TripRepository
	paystackService *payment.PaystackService
	stateMachine    *services.StateMachine
}

func NewHandlePaystackWebhook(
	paymentRepo repositories.PaymentRepository,
	tripRepo repositories.TripRepository,
	paystackService *payment.PaystackService,
	stateMachine *services.StateMachine,
) *HandlePaystackWebhook {
	return &HandlePaystackWebhook{
		paymentRepo:     paymentRepo,
		tripRepo:        tripRepo,
		paystackService: paystackService,
		stateMachine:    stateMachine,
	}
}

func (uc *HandlePaystackWebhook) Execute(ctx context.Context, reference string, eventStatus string) error {
	// 1. Find the payment by Paystack reference
	paymentRecord, err := uc.paymentRepo.GetByReference(ctx, reference)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if paymentRecord == nil {
		return fmt.Errorf("payment not found for reference: %s", reference)
	}

	// 2. Verify with Paystack API to confirm the payment is genuine
	verifyResp, err := uc.paystackService.VerifyTransaction(reference)
	if err != nil {
		return fmt.Errorf("failed to verify transaction: %w", err)
	}

	if verifyResp.Data.Status != "success" {
		// Payment failed — update status
		if err := uc.paymentRepo.UpdateStatus(ctx, paymentRecord.ID, entities.PaymentStatusFailed); err != nil {
			return err
		}
		return nil
	}

	// 3. Payment successful — update payment status
	if err := uc.paymentRepo.UpdateStatus(ctx, paymentRecord.ID, entities.PaymentStatusCompleted); err != nil {
		return err
	}

	// 4. Transition trip from PAYMENT_PROCESSING → PAYMENT_COMPLETED → RATING_PENDING
	trip, err := uc.tripRepo.GetByID(ctx, paymentRecord.TripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return fmt.Errorf("trip not found: %s", paymentRecord.TripID)
	}

	// PAYMENT_PROCESSING → PAYMENT_COMPLETED
	if err := uc.stateMachine.Transition(trip.Status, entities.StatusPaymentCompleted); err != nil {
		return fmt.Errorf("invalid transition to PAYMENT_COMPLETED: %w", err)
	}
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusPaymentCompleted); err != nil {
		return err
	}

	// PAYMENT_COMPLETED → RATING_PENDING
	if err := uc.stateMachine.Transition(entities.StatusPaymentCompleted, entities.StatusRatingPending); err != nil {
		return fmt.Errorf("invalid transition to RATING_PENDING: %w", err)
	}
	if err := uc.tripRepo.UpdateStatus(ctx, trip.ID, entities.StatusRatingPending); err != nil {
		return err
	}

	return nil
}
