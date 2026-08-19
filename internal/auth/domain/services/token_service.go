package services

import (
	"time"

	"github.com/google/uuid"
)

type AccessClaims struct {
	UserID      uuid.UUID
	SessionID   uuid.UUID
	Roles       []string
	IsAdmin     bool
	MFAVerified bool // NEW: Tracks if this session passed MFA
	ExpiresAt   time.Time
}

type TokenService interface {
	GenerateAccessToken(userID, sessionID uuid.UUID, roles []string, isAdmin bool, mfaVerified bool) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	HashToken(token string) (string, error)
	GenerateSecureToken() (string, error)
	GenerateMFATicket(userID uuid.UUID) (string, error)
	ValidateMFATicket(tokenString string) (uuid.UUID, error)
}