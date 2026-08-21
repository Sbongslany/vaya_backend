package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/services"
	"github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/paystack"
	"github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/wallet/interfaces/http/handlers"
)

type WalletContainer struct {
	Handler       *handlers.WalletHandler
	SplitTripFare *usecases.SplitTripFare
}

func WireWallet(pgPool *pgxpool.Pool, paystackSecretKey string) *WalletContainer {
	// Repositories
	walletRepo := postgres.NewWalletRepository(pgPool)
	ledgerRepo := postgres.NewLedgerRepository(pgPool)
	payoutRepo := postgres.NewPayoutRepository(pgPool)

	// Services
	commissionSvc := services.NewCommissionService(0.20)
	paystackTransferSvc := paystack.NewPaystackTransferService(paystackSecretKey)

	// Use cases
	getWalletUC := usecases.NewGetWallet(walletRepo)
	getHistoryUC := usecases.NewGetLedgerHistory(walletRepo, ledgerRepo)
	adminTopupUC := usecases.NewAdminTopup(walletRepo, ledgerRepo)
	splitTripFareUC := usecases.NewSplitTripFare(walletRepo, ledgerRepo, commissionSvc)
	requestPayoutUC := usecases.NewRequestPayout(walletRepo, payoutRepo, ledgerRepo, paystackTransferSvc)
	getPayoutHistoryUC := usecases.NewGetPayoutHistory(payoutRepo)
	handleTransferHookUC := usecases.NewHandleTransferWebhook(payoutRepo)
	listBanksUC := usecases.NewListBanks(paystackTransferSvc)

	// Handler
	handler := handlers.NewWalletHandler(
		getWalletUC,
		getHistoryUC,
		adminTopupUC,
		requestPayoutUC,
		getPayoutHistoryUC,
		handleTransferHookUC,
		listBanksUC,
		paystackTransferSvc,
	)

	return &WalletContainer{
		Handler:       handler,
		SplitTripFare: splitTripFareUC,
	}
}
