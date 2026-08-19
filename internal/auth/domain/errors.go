package domain

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountLocked          = errors.New("account is locked")
	ErrAccountDisabled        = errors.New("account is disabled")
	ErrAccountPending         = errors.New("account is pending verification")
	
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionExpired         = errors.New("session has expired")
	ErrSessionRevoked         = errors.New("session has been revoked")
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrRefreshTokenReused     = errors.New("refresh token reuse detected")

	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenExpired           = errors.New("token has expired")
	ErrTokenAlreadyUsed       = errors.New("token has already been used")
	ErrInvalidTokenFormat     = errors.New("invalid token format")

	ErrOTPNotFound            = errors.New("otp not found")
	ErrOTPExpired             = errors.New("otp has expired")
	ErrOTPInvalid             = errors.New("invalid otp")
	ErrOTPMaxAttemptsExceeded = errors.New("maximum otp attempts exceeded")
	ErrOTPCooldownActive      = errors.New("otp resend cooldown is active")

	ErrMFANotEnabled          = errors.New("mfa is not enabled for this user")
	ErrMFAInvalidCode         = errors.New("invalid mfa code")
	ErrMFAAlreadyEnabled      = errors.New("mfa is already enabled")

	ErrUnauthorized           = errors.New("unauthorized")
	ErrForbidden              = errors.New("forbidden: insufficient permissions")
	ErrRateLimited            = errors.New("rate limit exceeded")
	ErrInternalServer         = errors.New("internal server error")
)