package security

import (
	"encoding/hex"
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type MFAService struct {
	cfg *config.Config
}

func NewMFAService(cfg *config.Config) *MFAService {
	return &MFAService{cfg: cfg}
}

func (s *MFAService) GenerateSecret(accountName string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.JWTIssuer,
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *MFAService) EncryptSecret(plainSecret string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(s.cfg.MFAEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid MFA encryption key format: %w", err)
	}
	return Encrypt([]byte(plainSecret), keyBytes)
}

func (s *MFAService) DecryptSecret(encryptedSecret []byte) (string, error) {
	keyBytes, err := hex.DecodeString(s.cfg.MFAEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("invalid MFA encryption key format: %w", err)
	}
	plainBytes, err := Decrypt(encryptedSecret, keyBytes)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}

func (s *MFAService) ValidateTOTP(secret string, code string) bool {
	return totp.Validate(code, secret)
}

// Ensure MFAService implements the domain interface if we had one. 
// For now, we use it directly in use cases.
var _ services.MFAService = (*MFAService)(nil)