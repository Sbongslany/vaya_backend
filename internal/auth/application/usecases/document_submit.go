package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
)

type SubmitDocumentRequest struct {
	UserID    uuid.UUID
	VehicleID *uuid.UUID
	DocType   domain.DocumentType
	FileKey   string // Cloudinary public_id
	FileURL   string // Cloudinary secure_url
}

type SubmitDocument struct {
	driverRepo repositories.DriverRepository
	docRepo    repositories.DocumentRepository
}

func NewSubmitDocument(driverRepo repositories.DriverRepository, docRepo repositories.DocumentRepository) *SubmitDocument {
	return &SubmitDocument{driverRepo: driverRepo, docRepo: docRepo}
}

func (uc *SubmitDocument) Execute(ctx context.Context, req SubmitDocumentRequest) error {
	// 1. Verify the user has an active driver profile
	profile, err := uc.driverRepo.GetProfileByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 2. Create the document record
	now := time.Now()
	doc := &entities.DriverDocument{
		ID:              uuid.New(),
		DriverProfileID: profile.ID,
		VehicleID:       req.VehicleID,
		DocType:         req.DocType,
		FileKey:         req.FileKey,
		FileURL:         req.FileURL,
		Status:          domain.DocStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return uc.docRepo.Create(ctx, doc)
}
