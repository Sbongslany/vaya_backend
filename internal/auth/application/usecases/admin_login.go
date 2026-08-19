package usecases

import (
	"context"
	"errors"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type AdminLoginRequest struct {
	Email    string
	Password string
}

type AdminLoginResponse struct {
	MFATicket  string `json:"mfa_ticket"`
	MFAEnabled bool   `json:"mfa_enabled"`
}

type AdminLogin struct {
	userRepo    repositories.UserRepository
	mfaRepo     repositories.MFARepository
	passwordSvc services.PasswordService
	tokenSvc    services.TokenService
}

func NewAdminLogin(
	userRepo repositories.UserRepository,
	mfaRepo repositories.MFARepository,
	passwordSvc services.PasswordService,
	tokenSvc services.TokenService,
) *AdminLogin {
	return &AdminLogin{
		userRepo: userRepo, mfaRepo: mfaRepo,
		passwordSvc: passwordSvc, tokenSvc: tokenSvc,
	}
}

func (uc *AdminLogin) Execute(ctx context.Context, req AdminLoginRequest) (*AdminLoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err == domain.ErrUserNotFound {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	// Check Admin Role
	if !user.IsAdmin() {
		return nil, domain.ErrForbidden
	}

	if user.Status != domain.StatusActive {
		return nil, domain.ErrAccountLocked
	}

	if err := uc.passwordSvc.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check MFA Status
	mfaSecret, err := uc.mfaRepo.FindByUserID(ctx, user.ID)
	mfaEnabled := false
	if err == nil && mfaSecret.IsEnabled {
		mfaEnabled = true
	} else if err != nil && !errors.Is(err, domain.ErrMFANotEnabled) {
		return nil, err
	}

	// Generate short-lived MFA ticket (valid for 5 minutes)
	// We will add GenerateMFATicket to TokenService
	mfaTicket, err := uc.tokenSvc.GenerateMFATicket(user.ID)
	if err != nil {
		return nil, err
	}

	return &AdminLoginResponse{
		MFATicket:  mfaTicket,
		MFAEnabled: mfaEnabled,
	}, nil
}