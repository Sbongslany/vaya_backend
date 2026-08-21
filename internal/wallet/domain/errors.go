package domain

import "errors"

var (
	ErrWalletNotFound      = errors.New("wallet_not_found")
	ErrInsufficientBalance = errors.New("insufficient_balance")
	ErrInvalidAmount       = errors.New("invalid_amount")
	ErrWalletAlreadyExists = errors.New("wallet_already_exists")
	ErrUnauthorized        = errors.New("unauthorized")
)
