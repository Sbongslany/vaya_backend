package usecases

import (
	"context"
	"time"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

// AppConfig holds use-case specific configuration
type AppConfig struct {
	RefreshTTL time.Duration
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

type RefreshToken struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	tokenSvc    services.TokenService
	cfg         *AppConfig
}

func NewRefreshToken(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	tokenSvc services.TokenService,
	cfg *AppConfig,
) *RefreshToken {
	return &RefreshToken{userRepo: userRepo, sessionRepo: sessionRepo, tokenSvc: tokenSvc, cfg: cfg}
}

func (uc *RefreshToken) Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	hash, err := uc.tokenSvc.HashToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	session, err := uc.sessionRepo.FindByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if session.IsRevoked() {
		return nil, domain.ErrRefreshTokenReused
	}

	if session.IsExpired(time.Now()) {
		return nil, domain.ErrSessionExpired
	}

	// Rotate Refresh Token
	newRefreshToken, err := uc.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newHash, err := uc.tokenSvc.HashToken(newRefreshToken)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session.RefreshTokenHash = newHash
	session.LastUsedAt = now
	session.ExpiresAt = now.Add(uc.cfg.RefreshTTL)

	// Persist the updated session (rotated refresh token and new expiration)
	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, string(r))
	}

	// Note: We pass session.MFAVerified here so the new access token retains the MFA status
	accessToken, err := uc.tokenSvc.GenerateAccessToken(user.ID, session.ID, roles, user.IsAdmin(), session.MFAVerified)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
