package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/cloudinary"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type GenerateUploadSignatureRequest struct {
	UserID  uuid.UUID
	DocType string
}

type GenerateUploadSignatureResponse struct {
	CloudName string `json:"cloud_name"`
	APIKey    string `json:"api_key"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
	Folder    string `json:"folder"`
}

type GenerateUploadSignature struct {
	driverRepo    repositories.DriverRepository
	cloudinarySvc *cloudinary.CloudinaryService
	cfg           *config.Config
}

func NewGenerateUploadSignature(driverRepo repositories.DriverRepository, cloudinarySvc *cloudinary.CloudinaryService, cfg *config.Config) *GenerateUploadSignature {
	return &GenerateUploadSignature{driverRepo: driverRepo, cloudinarySvc: cloudinarySvc, cfg: cfg}
}

func (uc *GenerateUploadSignature) Execute(ctx context.Context, req GenerateUploadSignatureRequest) (*GenerateUploadSignatureResponse, error) {
	// 1. Verify the user has an active driver profile
	profile, err := uc.driverRepo.GetProfileByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// 2. Generate timestamp and signature
	timestamp := time.Now().Unix()
	folder := "driver_onboarding/" + profile.ID.String()
	signature := uc.cloudinarySvc.GenerateSignature(folder, timestamp)

	return &GenerateUploadSignatureResponse{
		CloudName: uc.cfg.CloudinaryCloudName,
		APIKey:    uc.cfg.CloudinaryAPIKey,
		Timestamp: timestamp,
		Signature: signature,
		Folder:    folder,
	}, nil
}
