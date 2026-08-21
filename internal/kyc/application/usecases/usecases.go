package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/kyc/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/kyc/domain/repositories"
)

// --- List Pending KYC Drivers ---

type ListPendingKYC struct {
	kycRepo repositories.KYCRepository
}

func NewListPendingKYC(kycRepo repositories.KYCRepository) *ListPendingKYC {
	return &ListPendingKYC{kycRepo: kycRepo}
}

func (uc *ListPendingKYC) Execute(ctx context.Context, limit, offset int) ([]*entities.DriverKYCSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.kycRepo.ListPendingDrivers(ctx, limit, offset)
}

// --- Get Driver Documents ---

type GetDriverDocuments struct {
	kycRepo repositories.KYCRepository
}

func NewGetDriverDocuments(kycRepo repositories.KYCRepository) *GetDriverDocuments {
	return &GetDriverDocuments{kycRepo: kycRepo}
}

func (uc *GetDriverDocuments) Execute(ctx context.Context, userID uuid.UUID) ([]*entities.DriverDocument, error) {
	return uc.kycRepo.GetDocumentsByUserID(ctx, userID)
}

// --- Review Document (Approve/Reject) ---

type ReviewDocumentInput struct {
	DocumentID uuid.UUID
	Status     entities.DocumentStatus
	Reason     *string
}

type ReviewDocument struct {
	kycRepo repositories.KYCRepository
}

func NewReviewDocument(kycRepo repositories.KYCRepository) *ReviewDocument {
	return &ReviewDocument{kycRepo: kycRepo}
}

func (uc *ReviewDocument) Execute(ctx context.Context, input ReviewDocumentInput) error {
	doc, err := uc.kycRepo.GetDocumentByID(ctx, input.DocumentID)
	if err != nil {
		return err
	}
	if doc == nil {
		return entities.ErrDocumentNotFound
	}

	// Update the specific document status
	if err := uc.kycRepo.UpdateDocumentStatus(ctx, input.DocumentID, input.Status, input.Reason); err != nil {
		return err
	}

	// Recalculate the driver's overall onboarding status
	total, pending, approved, rejected, err := uc.kycRepo.CountDocumentStatuses(ctx, doc.UserID)
	if err != nil {
		return err
	}

	var newStatus entities.OnboardingStatus
	switch {
	case rejected > 0:
		newStatus = entities.OnboardingRejected
	case pending > 0:
		newStatus = entities.OnboardingPending
	case approved == total && total > 0:
		newStatus = entities.OnboardingApproved
	default:
		newStatus = entities.OnboardingPending
	}

	return uc.kycRepo.UpdateUserOnboardingStatus(ctx, doc.UserID, newStatus)
}
