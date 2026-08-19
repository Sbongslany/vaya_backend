package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type JWTTokenService struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) *JWTTokenService {
	return &JWTTokenService{cfg: cfg}
}

func (s *JWTTokenService) GenerateAccessToken(userID, sessionID uuid.UUID, roles []string, isAdmin bool, mfaVerified bool) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":          userID.String(),
		"sid":          sessionID.String(),
		"roles":        roles,
		"is_admin":     isAdmin,
		"mfa_verified": mfaVerified,
		"iat":          now.Unix(),
		"exp":          now.Add(s.cfg.JWTAccessTTL).Unix(),
		"iss":          s.cfg.JWTIssuer,
		"aud":          s.cfg.JWTAudience,
		"token_type":   "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTAccessSecret))
}

func (s *JWTTokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *JWTTokenService) ValidateAccessToken(tokenString string) (*services.AccessClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTAccessSecret), nil
	}, jwt.WithIssuer(s.cfg.JWTIssuer), jwt.WithAudience(s.cfg.JWTAudience))

	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidTokenFormat
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrInvalidTokenFormat
	}

	if claims["token_type"] != "access" {
		return nil, domain.ErrInvalidTokenFormat
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, domain.ErrInvalidTokenFormat
	}
	
	sessionID, err := uuid.Parse(claims["sid"].(string))
	if err != nil {
		return nil, domain.ErrInvalidTokenFormat
	}

	var roles []string
	if r, ok := claims["roles"].([]interface{}); ok {
		for _, role := range r {
			if str, ok := role.(string); ok {
				roles = append(roles, str)
			}
		}
	}

	isAdmin, _ := claims["is_admin"].(bool)
	mfaVerified, _ := claims["mfa_verified"].(bool)
	expUnix := int64(claims["exp"].(float64))

	return &services.AccessClaims{
		UserID:      userID,
		SessionID:   sessionID,
		Roles:       roles,
		IsAdmin:     isAdmin,
		MFAVerified: mfaVerified,
		ExpiresAt:   time.Unix(expUnix, 0),
	}, nil
}

func (s *JWTTokenService) HashToken(token string) (string, error) {
	h := sha256.New()
	h.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}

func (s *JWTTokenService) GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *JWTTokenService) GenerateMFATicket(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":        userID.String(),
		"purpose":    "admin_mfa_ticket",
		"iat":        now.Unix(),
		"exp":        now.Add(5 * time.Minute).Unix(),
		"iss":        s.cfg.JWTIssuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTAccessSecret))
}

func (s *JWTTokenService) ValidateMFATicket(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTAccessSecret), nil
	}, jwt.WithIssuer(s.cfg.JWTIssuer))

	if err != nil || !token.Valid {
		return uuid.Nil, domain.ErrInvalidTokenFormat
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["purpose"] != "admin_mfa_ticket" {
		return uuid.Nil, domain.ErrInvalidTokenFormat
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return uuid.Nil, domain.ErrInvalidTokenFormat
	}

	return userID, nil
}