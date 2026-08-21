package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/support/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/support/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/support/interfaces/http/handlers"
	walletDep "github.com/yourorg/ehailing/backend/internal/wallet/dependency"
)

type SupportContainer struct {
	Handler *handlers.SupportHandler
}

func WireSupport(pgPool *pgxpool.Pool, walletContainer *walletDep.WalletContainer) *SupportContainer {
	// Repositories
	ticketRepo := postgres.NewTicketRepository(pgPool)
	commentRepo := postgres.NewCommentRepository(pgPool)
	refundRepo := postgres.NewRefundRepository(pgPool)

	// Wallet creditor adapter (for processing refunds)
	walletCreditor := walletDep.NewWalletCreditorAdapter(walletContainer.AdminTopupUC)

	// Use cases
	createTicketUC := usecases.NewCreateTicket(ticketRepo)
	getUserTicketsUC := usecases.NewGetUserTickets(ticketRepo)
	addCommentUC := usecases.NewAddComment(ticketRepo, commentRepo)
	resolveTicketUC := usecases.NewResolveTicket(ticketRepo, commentRepo, refundRepo, walletCreditor)

	// Handler
	handler := handlers.NewSupportHandler(
		createTicketUC,
		getUserTicketsUC,
		addCommentUC,
		resolveTicketUC,
	)

	return &SupportContainer{
		Handler: handler,
	}
}
