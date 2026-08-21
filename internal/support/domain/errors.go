package domain

import "errors"

var (
	ErrTicketNotFound = errors.New("ticket_not_found")
	ErrRefundNotFound = errors.New("refund_not_found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidStatus  = errors.New("invalid_ticket_status_transition")
	ErrInvalidAmount  = errors.New("invalid_refund_amount")
)
