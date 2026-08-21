package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/trip/infrastructure/payment"
	walletUseCases "github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
	walletDomain "github.com/yourorg/ehailing/backend/internal/wallet/domain"
)

type InitiateTripPaymentInput struct {
	TripID         uuid.UUID
	PassengerID    uuid.UUID
	Method         entities.PaymentMethod
	PassengerEmail string
}

type InitiateTripPaymentResult struct {
	PaymentID        uuid.UUID
	Method           entities.PaymentMethod
	AuthorizationURL string // Only for CARD payments
	Status           entities.PaymentStatus
}

type InitiateTripPayment struct {
	tripRepo        repositories.TripRepository
	paymentRepo     repositories.PaymentRepository
	paystackService *payment.PaystackService
	walletBalanceUC *walletUseCases.GetWallet
	callbackURL     string
}

func NewInitiateTripPayment(
	tripRepo repositories.TripRepository,
	paymentRepo repositories.PaymentRepository,
	paystackService *payment.PaystackService,
	walletBalanceUC *walletUseCases.GetWallet,
	callbackURL string,
) *InitiateTripPayment {
	return &InitiateTripPayment{
		tripRepo:        tripRepo,
		paymentRepo:     paymentRepo,
		paystackService: paystackService,
		walletBalanceUC: walletBalanceUC,
		callbackURL:     callbackURL,
	}
}

func (uc *InitiateTripPayment) Execute(ctx context.Context, input InitiateTripPaymentInput) (*InitiateTripPaymentResult, error) {
	// 1. Get the trip
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

	// 2. Check if payment already exists
	existingPayment, err := uc.paymentRepo.GetByTripID(ctx, input.TripID)
	if err != nil {
		return nil, err
	}
	if existingPayment != nil && existingPayment.Status == entities.PaymentStatusCompleted {
		return nil, domain.ErrAlreadyPaid
	}

	// 3. Calculate the amount (subtract any promo discount)
	amount := trip.EstimatedFare
	if trip.FinalFare != nil {
		amount = *trip.FinalFare
	}

	now := time.Now()
	paymentID := uuid.New()

	// 4. Handle based on payment method
	switch input.Method {
	case entities.PaymentMethodCash:
		return uc.processCashPayment(ctx, paymentID, input, trip, amount, now)
	case entities.PaymentMethodWallet:
		return uc.processWalletPayment(ctx, paymentID, input, trip, amount, now)
	case entities.PaymentMethodCard:
		return uc.processCardPayment(ctx, paymentID, input, trip, amount, now)
	default:
		return nil, fmt.Errorf("unsupported payment method: %s", input.Method)
	}
}

func (uc *InitiateTripPayment) processCashPayment(
	ctx context.Context,
	paymentID uuid.UUID,
	input InitiateTripPaymentInput,
	trip *entities.Trip,
	amount float64,
	now time.Time,
) (*InitiateTripPaymentResult, error) {
	payment := &entities.Payment{
		ID:          paymentID,
		TripID:      input.TripID,
		PassengerID: input.PassengerID,
		Amount:      amount,
		Currency:    trip.Currency,
		Method:      entities.PaymentMethodCash,
		Status:      entities.PaymentStatusCompleted,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &InitiateTripPaymentResult{
		PaymentID: paymentID,
		Method:    entities.PaymentMethodCash,
		Status:    entities.PaymentStatusCompleted,
	}, nil
}

func (uc *InitiateTripPayment) processWalletPayment(
	ctx context.Context,
	paymentID uuid.UUID,
	input InitiateTripPaymentInput,
	trip *entities.Trip,
	amount float64,
	now time.Time,
) (*InitiateTripPaymentResult, error) {
	// Check wallet balance
	wallet, err := uc.walletBalanceUC.Execute(ctx, input.PassengerID)
	if err != nil {
		if err == walletDomain.ErrWalletNotFound {
			return nil, fmt.Errorf("wallet not found — please top up or use another payment method")
		}
		return nil, err
	}

	if wallet.Balance < amount {
		return nil, fmt.Errorf("insufficient wallet balance: available R%.2f, required R%.2f", wallet.Balance, amount)
	}

	// Create payment record
	payment := &entities.Payment{
		ID:          paymentID,
		TripID:      input.TripID,
		PassengerID: input.PassengerID,
		Amount:      amount,
		Currency:    trip.Currency,
		Method:      entities.PaymentMethodWallet,
		Status:      entities.PaymentStatusCompleted,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// Note: Wallet debit will be handled in a follow-up step with a proper
	// wallet debit use case. For now, the payment is marked as completed.

	return &InitiateTripPaymentResult{
		PaymentID: paymentID,
		Method:    entities.PaymentMethodWallet,
		Status:    entities.PaymentStatusCompleted,
	}, nil
}

func (uc *InitiateTripPayment) processCardPayment(
	ctx context.Context,
	paymentID uuid.UUID,
	input InitiateTripPaymentInput,
	trip *entities.Trip,
	amount float64,
	now time.Time,
) (*InitiateTripPaymentResult, error) {
	// Generate unique reference
	reference := fmt.Sprintf("VAYA_%s_%d", input.TripID.String()[:8], now.Unix())

	// Initialize Paystack transaction
	paystackResp, err := uc.paystackService.InitializeTransaction(payment.InitializeTransactionRequest{
		Email:       input.PassengerEmail,
		Amount:      amount,
		Currency:    trip.Currency,
		Reference:   reference,
		CallbackURL: uc.callbackURL,
		Metadata: map[string]interface{}{
			"trip_id":      input.TripID.String(),
			"passenger_id": input.PassengerID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Paystack transaction: %w", err)
	}

	// Create payment record
	payment := &entities.Payment{
		ID:                       paymentID,
		TripID:                   input.TripID,
		PassengerID:              input.PassengerID,
		Amount:                   amount,
		Currency:                 trip.Currency,
		Method:                   entities.PaymentMethodCard,
		Status:                   entities.PaymentStatusProcessing,
		PaystackReference:        &paystackResp.Data.Reference,
		PaystackAuthorizationURL: &paystackResp.Data.AuthorizationURL,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &InitiateTripPaymentResult{
		PaymentID:        paymentID,
		Method:           entities.PaymentMethodCard,
		AuthorizationURL: paystackResp.Data.AuthorizationURL,
		Status:           entities.PaymentStatusProcessing,
	}, nil
}
