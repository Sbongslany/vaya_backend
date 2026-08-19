package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type RegisterUserRequest struct {
	FirstName string
	LastName  string
	Email     *string
	Phone     *string
	Password  string
	IPAddress *string // Added for auditing
	UserAgent *string // Added for auditing
}

type RegisterUser struct {
	userRepo    repositories.UserRepository
	passwordSvc services.PasswordService
	auditRepo   repositories.AuditRepository // Added
}

func NewRegisterUser(userRepo repositories.UserRepository, passwordSvc services.PasswordService, auditRepo repositories.AuditRepository) *RegisterUser {
	return &RegisterUser{userRepo: userRepo, passwordSvc: passwordSvc, auditRepo: auditRepo}
}

func (uc *RegisterUser) Execute(ctx context.Context, req RegisterUserRequest) (*entities.User, error) {
	if req.Email == nil && req.Phone == nil {
		return nil, domain.ErrInvalidTokenFormat 
	}

	hash, err := uc.passwordSvc.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &entities.User{
		ID:           uuid.New(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hash,
		Status:       domain.StatusPendingVerification,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Audit Log: Registration Completed
	uc.auditRepo.Log(ctx, &entities.AuditLog{
		ID:        uuid.New(),
		UserID:    &user.ID,
		Action:    domain.AuditActionRegisterCompleted,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		CreatedAt: now,
	})

	return user, nil
}