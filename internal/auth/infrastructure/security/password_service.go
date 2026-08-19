package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"golang.org/x/crypto/argon2"
)

type Argon2PasswordService struct{}

func NewPasswordService() *Argon2PasswordService {
	return &Argon2PasswordService{}
}

func (s *Argon2PasswordService) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Argon2id parameters: memory=64MB, iterations=3, parallelism=2, keyLen=32
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", saltB64, hashB64), nil
}

func (s *Argon2PasswordService) ComparePassword(hashedPassword, password string) error {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return domain.ErrInvalidCredentials
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	// Use same parameters to generate comparison hash
	comparisonHash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)

	if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
		return nil
	}

	return domain.ErrInvalidCredentials
}
