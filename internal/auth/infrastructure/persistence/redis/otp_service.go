package redis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type OTpService struct {
	client *redis.Client
}

func NewOTPService(client *redis.Client) *OTpService {
	return &OTpService{client: client}
}

func (s *OTpService) GenerateOTP(length int) (string, error) {
	otp := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp[i] = byte('0' + n.Int64())
	}
	return string(otp), nil
}

func hashOTP(plainOTP string) string {
	h := sha256.Sum256([]byte(plainOTP))
	return hex.EncodeToString(h[:])
}

func (s *OTpService) StoreOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose, plainOTP string, ttl time.Duration) error {
	key := fmt.Sprintf("otp:data:%s:%s", purpose, identifier)
	hashed := hashOTP(plainOTP)

	// Reset attempts when a new OTP is stored
	attemptsKey := fmt.Sprintf("otp:attempts:%s:%s", purpose, identifier)
	s.client.Del(ctx, attemptsKey)

	return s.client.Set(ctx, key, hashed, ttl).Err()
}

func (s *OTpService) VerifyOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose, plainOTP string) error {
	key := fmt.Sprintf("otp:data:%s:%s", purpose, identifier)

	storedHash, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return domain.ErrOTPNotFound
	}
	if err != nil {
		return err
	}

	if storedHash != hashOTP(plainOTP) {
		return domain.ErrOTPInvalid
	}

	return nil
}

func (s *OTpService) InvalidateOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose) error {
	key := fmt.Sprintf("otp:data:%s:%s", purpose, identifier)
	return s.client.Del(ctx, key).Err()
}

func (s *OTpService) IsInCooldown(ctx context.Context, identifier string, purpose domain.OTPPurpose) (bool, error) {
	key := fmt.Sprintf("otp:cooldown:%s:%s", purpose, identifier)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *OTpService) SetCooldown(ctx context.Context, identifier string, purpose domain.OTPPurpose, cooldown time.Duration) error {
	key := fmt.Sprintf("otp:cooldown:%s:%s", purpose, identifier)
	return s.client.Set(ctx, key, "1", cooldown).Err()
}

func (s *OTpService) IncrementAttempts(ctx context.Context, identifier string, purpose domain.OTPPurpose, maxAttempts int) (int, error) {
	key := fmt.Sprintf("otp:attempts:%s:%s", purpose, identifier)

	attempts, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiration on the attempts key so it cleans up automatically
	s.client.Expire(ctx, key, 15*time.Minute)

	if int(attempts) > maxAttempts {
		s.InvalidateOTP(ctx, identifier, purpose)
		return int(attempts), domain.ErrOTPMaxAttemptsExceeded
	}

	return int(attempts), nil
}

func (s *OTpService) ResetAttempts(ctx context.Context, identifier string, purpose domain.OTPPurpose) error {
	key := fmt.Sprintf("otp:attempts:%s:%s", purpose, identifier)
	return s.client.Del(ctx, key).Err()
}
