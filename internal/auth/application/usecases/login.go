package usecases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type LoginUserRequest struct {
	Email      *string
	Phone      *string
	Password   string
	DeviceID   *string
	DeviceType *string
	DeviceName *string
	IPAddress  *string
	UserAgent  *string
}

type LoginUserResponse struct {
	User         *entities.User
	AccessToken  string
	RefreshToken string
	SessionID    uuid.UUID
}

type LoginUser struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	passwordSvc services.PasswordService
	tokenSvc    services.TokenService
	auditRepo   repositories.AuditRepository // Added
	cfg         *AppConfig 
}

func NewLoginUser(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	passwordSvc services.PasswordService,
	tokenSvc services.TokenService,
	auditRepo repositories.AuditRepository, // Added
	cfg *AppConfig,
) *LoginUser {
	return &LoginUser{
		userRepo: userRepo, sessionRepo: sessionRepo,
		passwordSvc: passwordSvc, tokenSvc: tokenSvc, auditRepo: auditRepo, cfg: cfg,
	}
}

func (uc *LoginUser) Execute(ctx context.Context, req LoginUserRequest) (*LoginUserResponse, error) {
	var user *entities.User
	var err error
	var identifier string

	if req.Email != nil {
		identifier = *req.Email
		user, err = uc.userRepo.FindByEmail(ctx, identifier)
	} else if req.Phone != nil {
		identifier = *req.Phone
		user, err = uc.userRepo.FindByPhone(ctx, identifier)
	} else {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now()

	if err != nil || user == nil {
		// Audit Log: Login Failed (User Not Found)
		metadata, _ := json.Marshal(map[string]string{"identifier": identifier, "reason": "not_found"})
		uc.auditRepo.Log(ctx, &entities.AuditLog{
			ID: uuid.New(), Action: domain.AuditActionLoginFailed,
			IPAddress: req.IPAddress, UserAgent: req.UserAgent, Metadata: metadata, CreatedAt: now,
		})
		return nil, domain.ErrInvalidCredentials 
	}

	if user.Status == domain.StatusLocked {
		return nil, domain.ErrAccountLocked
	}
	if user.Status == domain.StatusDisabled {
		return nil, domain.ErrAccountDisabled
	}

	if err := uc.passwordSvc.ComparePassword(user.PasswordHash, req.Password); err != nil {
		// Audit Log: Login Failed (Bad Password)
		metadata, _ := json.Marshal(map[string]string{"identifier": identifier, "reason": "invalid_password"})
		uc.auditRepo.Log(ctx, &entities.AuditLog{
			ID: uuid.New(), UserID: &user.ID, Action: domain.AuditActionLoginFailed,
			IPAddress: req.IPAddress, UserAgent: req.UserAgent, Metadata: metadata, CreatedAt: now,
		})
		return nil, domain.ErrInvalidCredentials
	}

	// Generate Tokens
	refreshToken, _ := uc.tokenSvc.GenerateRefreshToken()
	refreshHash, _ := uc.tokenSvc.HashToken(refreshToken)

	session := &entities.Session{
		ID: uuid.New(), UserID: user.ID, RefreshTokenHash: refreshHash,
		DeviceID: req.DeviceID, DeviceType: req.DeviceType, DeviceName: req.DeviceName,
		IPAddress: req.IPAddress, UserAgent: req.UserAgent, MFAVerified: false,
		CreatedAt: now, LastUsedAt: now, ExpiresAt: now.Add(uc.cfg.RefreshTTL),
	}
	uc.sessionRepo.Create(ctx, session)

	var roles []string
	for _, r := range user.Roles { roles = append(roles, string(r)) }
	accessToken, _ := uc.tokenSvc.GenerateAccessToken(user.ID, session.ID, roles, user.IsAdmin(), false)

	// Audit Log: Login Success
	uc.auditRepo.Log(ctx, &entities.AuditLog{
		ID: uuid.New(), UserID: &user.ID, Action: domain.AuditActionLoginSuccess,
		IPAddress: req.IPAddress, UserAgent: req.UserAgent, CreatedAt: now,
	})

	return &LoginUserResponse{
		User: user, AccessToken: accessToken, RefreshToken: refreshToken, SessionID: session.ID,
	}, nil
}