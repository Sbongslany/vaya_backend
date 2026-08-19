package sms

import (
	"context"
	"fmt"
	"log/slog"
)

type SMSProvider interface {
	SendSMS(ctx context.Context, to string, message string) error
}

// ConsoleProvider logs the SMS to the terminal for local development.
// In production, you would implement a TwilioProvider or AWS SNSProvider here.
type ConsoleProvider struct {
	log *slog.Logger
}

func NewConsoleProvider(log *slog.Logger) *ConsoleProvider {
	return &ConsoleProvider{log: log}
}

func (p *ConsoleProvider) SendSMS(ctx context.Context, to string, message string) error {
	p.log.Info("📱 [MOCK SMS PROVIDER]", "to", to, "message", message)
	fmt.Printf("\n📱 SMS to %s: %s\n\n", to, message)
	return nil
}
