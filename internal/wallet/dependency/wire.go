package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain/services"
	"github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/wallet/interfaces/http/handlers"
)

type WalletContainer struct {
	Handler       *handlers.WalletHandler
	SplitTripFare *usecases.SplitTripFare
}

func WireWallet(pgPool *pgxpool.Pool) *WalletContainer {
	// Repositories
	walletRepo := postgres.NewWalletRepository(pgPool)
	ledgerRepo := postgres.NewLedgerRepository(pgPool)

	// Services (20% commission rate — configurable via env in production)
	commissionSvc := services.NewCommissionService(0.20)

	// Use cases
	getWalletUC := usecases.NewGetWallet(walletRepo)
	getHistoryUC := usecases.NewGetLedgerHistory(walletRepo, ledgerRepo)
	adminTopupUC := usecases.NewAdminTopup(walletRepo, ledgerRepo)
	splitTripFareUC := usecases.NewSplitTripFare(walletRepo, ledgerRepo, commissionSvc)

	// Handler
	handler := handlers.NewWalletHandler(getWalletUC, getHistoryUC, adminTopupUC)

	return &WalletContainer{
		Handler:       handler,
		SplitTripFare: splitTripFareUC,
	}
}
