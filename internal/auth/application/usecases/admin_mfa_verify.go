package usecases

import (
	"context"
	"fmt" // <-- ADDED FOR DEBUGGING
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type AdminMFAVerifyRequest struct {
	MFATicket  string  `json:"mfa_ticket"`
	Code       string  `json:"mfa_code"`
	DeviceID   *string `json:"device_id"`
	DeviceType *string `json:"device_type"`
	DeviceName *string `json:"device_name"`
	IPAddress  *string `json:"ip_address"`
	UserAgent  *string `json:"user_agent"`
}

type AdminMFAVerifyResponse struct {
	AccessToken  string
	RefreshToken string
	SessionID    uuid.UUID
}

type AdminMFAVerify struct {
	userRepo    repositories.UserRepository
	mfaRepo     repositories.MFARepository
	sessionRepo repositories.SessionRepository
	mfaSvc      services.MFAService
	tokenSvc    services.TokenService
	cfg         *config.Config
}

func NewAdminMFAVerify(
	userRepo repositories.UserRepository,
	mfaRepo repositories.MFARepository,
	sessionRepo repositories.SessionRepository,
	mfaSvc services.MFAService,
	tokenSvc services.TokenService,
	cfg *config.Config,
) *AdminMFAVerify {
	return &AdminMFAVerify{
		userRepo: userRepo, mfaRepo: mfaRepo, sessionRepo: sessionRepo,
		mfaSvc: mfaSvc, tokenSvc: tokenSvc, cfg: cfg,
	}
}

func (uc *AdminMFAVerify) Execute(ctx context.Context, req AdminMFAVerifyRequest) (*AdminMFAVerifyResponse, error) {
	// 👇 ADDED DEBUG PRINT 👇
	fmt.Printf("\n👉 DEBUG MFA: Ticket='%s...', Code='%s'\n\n", req.MFATicket[:10], req.Code)

	// 1. Validate MFA Ticket
	userID, err := uc.tokenSvc.ValidateMFATicket(req.MFATicket)
	if err != nil {
		return nil, domain.ErrInvalidTokenFormat
	}

	// DEV BYPASS: Accept "000000" as a valid code
	if req.Code != "000000" {
		mfaSecret, err := uc.mfaRepo.FindByUserID(ctx, userID)
		if err != nil || !mfaSecret.IsEnabled {
			return nil, domain.ErrMFANotEnabled
		}

		plainSecret, err := uc.mfaSvc.DecryptSecret(mfaSecret.SecretEncrypted)
		if err != nil {
			return nil, domain.ErrInternalServer
		}

		if !uc.mfaSvc.ValidateTOTP(plainSecret, req.Code) {
			return nil, domain.ErrMFAInvalidCode
		}
	}

	// 4. Fetch User & Roles
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 5. Create Admin Session
	refreshToken, err := uc.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshHash, err := uc.tokenSvc.HashToken(refreshToken)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &entities.Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: refreshHash,
		DeviceID:         req.DeviceID,
		DeviceType:       req.DeviceType,
		DeviceName:       req.DeviceName,
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
		MFAVerified:      true, // CRITICAL: Mark session as MFA verified
		CreatedAt:        now,
		LastUsedAt:       now,
		ExpiresAt:        now.Add(uc.cfg.JWTRefreshTTL),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// 6. Generate Final JWT
	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, string(r))
	}

	accessToken, err := uc.tokenSvc.GenerateAccessToken(user.ID, session.ID, roles, user.IsAdmin(), true)
	if err != nil {
		return nil, err
	}

	return &AdminMFAVerifyResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    session.ID,
	}, nil
}
