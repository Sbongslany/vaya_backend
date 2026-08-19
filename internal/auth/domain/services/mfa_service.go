package services

import "github.com/pquerna/otp"

type MFAService interface {
	GenerateSecret(accountName string) (*otp.Key, error)
	EncryptSecret(plainSecret string) ([]byte, error)
	DecryptSecret(encryptedSecret []byte) (string, error)
	ValidateTOTP(secret string, code string) bool
}
