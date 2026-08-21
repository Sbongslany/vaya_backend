package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/services"
)

type SplitTripFareInput struct {
	TripID      uuid.UUID
	PassengerID uuid.UUID
	DriverID    uuid.UUID
	Fare        float64
}

type SplitTripFare struct {
	walletRepo    repositories.WalletRepository
	ledgerRepo    repositories.LedgerRepository
	commissionSvc *services.CommissionService
}

func NewSplitTripFare(
	walletRepo repositories.WalletRepository,
	ledgerRepo repositories.LedgerRepository,
	commissionSvc *services.CommissionService,
) *SplitTripFare {
	return &SplitTripFare{
		walletRepo:    walletRepo,
		ledgerRepo:    ledgerRepo,
		commissionSvc: commissionSvc,
	}
}

// getOrCreateWallet ensures a wallet exists for the user (auto-creates if missing)
func (uc *SplitTripFare) getOrCreateWallet(ctx context.Context, userID uuid.UUID) (*entities.Wallet, error) {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	now := time.Now()
	wallet = &entities.Wallet{
		ID:        uuid.New(),
		UserID:    userID,
		Balance:   0,
		Currency:  "ZAR",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (uc *SplitTripFare) Execute(ctx context.Context, input SplitTripFareInput) error {
	if input.Fare <= 0 {
		return domain.ErrInvalidAmount
	}

	commission, driverEarnings := uc.commissionSvc.Calculate(input.Fare)
	now := time.Now()

	// 1. Credit driver wallet
	driverWallet, err := uc.getOrCreateWallet(ctx, input.DriverID)
	if err != nil {
		return err
	}

	driverNewBalance := driverWallet.Balance + driverEarnings
	if err := uc.walletRepo.UpdateBalance(ctx, driverWallet.ID, driverNewBalance); err != nil {
		return err
	}

	refType := entities.RefTripFare
	driverEntry := &entities.LedgerEntry{
		ID:            uuid.New(),
		WalletID:      driverWallet.ID,
		EntryType:     entities.LedgerEntryCredit,
		Amount:        driverEarnings,
		BalanceAfter:  driverNewBalance,
		ReferenceType: &refType,
		ReferenceID:   &input.TripID,
		Description:   "Trip fare earnings",
		CreatedAt:     now,
	}
	if err := uc.ledgerRepo.Create(ctx, driverEntry); err != nil {
		return err
	}

	// 2. Credit platform wallet with commission
	if commission > 0 {
		platformWallet, err := uc.walletRepo.GetPlatformWallet(ctx)
		if err != nil {
			return err
		}
		if platformWallet != nil {
			platformNewBalance := platformWallet.Balance + commission
			if err := uc.walletRepo.UpdateBalance(ctx, platformWallet.ID, platformNewBalance); err != nil {
				return err
			}

			commRefType := entities.RefPlatformCommission
			platformEntry := &entities.LedgerEntry{
				ID:            uuid.New(),
				WalletID:      platformWallet.ID,
				EntryType:     entities.LedgerEntryCredit,
				Amount:        commission,
				BalanceAfter:  platformNewBalance,
				ReferenceType: &commRefType,
				ReferenceID:   &input.TripID,
				Description:   "Platform commission",
				CreatedAt:     now,
			}
			if err := uc.ledgerRepo.Create(ctx, platformEntry); err != nil {
				return err
			}
		}
	}

	return nil
}
