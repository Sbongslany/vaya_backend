package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/kyc/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/kyc/infrastructure/persistence/postgres"
	"github.com/yourorg/ehailing/backend/internal/kyc/interfaces/http/handlers"
)

type KYCContainer struct {
	Handler *handlers.KYCHandler
}

func WireKYC(pgPool *pgxpool.Pool) *KYCContainer {
	kycRepo := postgres.NewKYCRepository(pgPool)

	listPendingUC := usecases.NewListPendingKYC(kycRepo)
	getDocumentsUC := usecases.NewGetDriverDocuments(kycRepo)
	reviewDocumentUC := usecases.NewReviewDocument(kycRepo)

	handler := handlers.NewKYCHandler(
		listPendingUC,
		getDocumentsUC,
		reviewDocumentUC,
	)

	return &KYCContainer{
		Handler: handler,
	}
}
