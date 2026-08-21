package domain

import "errors"

var (
	ErrMessageNotFound = errors.New("message_not_found")
	ErrCallNotFound    = errors.New("call_session_not_found")
	ErrInvalidStatus   = errors.New("invalid_call_status_transition")
	ErrUnauthorized    = errors.New("unauthorized")
)
