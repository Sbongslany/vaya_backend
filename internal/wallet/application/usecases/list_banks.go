package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/paystack"
)

type ListBanks struct {
	paystackService *paystack.PaystackTransferService
}

func NewListBanks(paystackService *paystack.PaystackTransferService) *ListBanks {
	return &ListBanks{paystackService: paystackService}
}

func (uc *ListBanks) Execute(ctx context.Context, country string) ([]paystack.Bank, error) {
	resp, err := uc.paystackService.ListBanks(country)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
