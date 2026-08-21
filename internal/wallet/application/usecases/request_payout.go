package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/paystack"
)

type RequestPayoutInput struct {
	UserID            uuid.UUID
	Amount            float64
	BankName          string
	BankAccountNumber string
	BankCode          string
}

type RequestPayout struct {
	walletRepo      repositories.WalletRepository
	payoutRepo      repositories.PayoutRepository
	ledgerRepo      repositories.LedgerRepository
	paystackService *paystack.PaystackTransferService
}

func NewRequestPayout(
	walletRepo repositories.WalletRepository,
	payoutRepo repositories.PayoutRepository,
	ledgerRepo repositories.LedgerRepository,
	paystackService *paystack.PaystackTransferService,
) *RequestPayout {
	return &RequestPayout{
		walletRepo:      walletRepo,
		payoutRepo:      payoutRepo,
		ledgerRepo:      ledgerRepo,
		paystackService: paystackService,
	}
}

func (uc *RequestPayout) Execute(ctx context.Context, input RequestPayoutInput) (*entities.Payout, error) {
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// 1. Check wallet balance
	wallet, err := uc.walletRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, domain.ErrWalletNotFound
	}

	if wallet.Balance < input.Amount {
		return nil, domain.ErrInsufficientBalance
	}

	// 2. Resolve account number with Paystack
	resolveResp, err := uc.paystackService.ResolveAccountNumber(input.BankAccountNumber, input.BankCode)
	if err != nil {
		return nil, err
	}

	// 3. Create or get transfer recipient
	recipientResp, err := uc.paystackService.CreateTransferRecipient(
		resolveResp.Data.AccountName,
		input.BankAccountNumber,
		input.BankCode,
		wallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	// 4. Initiate the transfer
	transferResp, err := uc.paystackService.InitiateTransfer(
		input.Amount,
		wallet.Currency,
		"Driver payout",
		recipientResp.Data.RecipientCode,
	)
	if err != nil {
		return nil, err
	}

	// 5. Create payout record
	now := time.Now()
	payoutID := uuid.New()
	transferRef := transferResp.Data.Reference
	transferID := string(rune(transferResp.Data.ID))

	payout := &entities.Payout{
		ID:                  payoutID,
		UserID:              input.UserID,
		Amount:              input.Amount,
		Currency:            wallet.Currency,
		Status:              entities.PayoutStatusProcessing,
		BankName:            input.BankName,
		BankAccountNumber:   input.BankAccountNumber,
		BankAccountName:     resolveResp.Data.AccountName,
		PaystackTransferRef: &transferRef,
		PaystackTransferID:  &transferID,
		RequestedAt:         now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := uc.payoutRepo.Create(ctx, payout); err != nil {
		return nil, err
	}

	// 6. Debit the wallet
	newBalance := wallet.Balance - input.Amount
	if err := uc.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
		return nil, err
	}

	// 7. Create ledger entry
	refType := entities.RefPayout
	ledgerEntry := &entities.LedgerEntry{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		EntryType:     entities.LedgerEntryDebit,
		Amount:        input.Amount,
		BalanceAfter:  newBalance,
		ReferenceType: &refType,
		ReferenceID:   &payoutID,
		Description:   "Payout to bank account",
		CreatedAt:     now,
	}

	if err := uc.ledgerRepo.Create(ctx, ledgerEntry); err != nil {
		return nil, err
	}

	return payout, nil
}
