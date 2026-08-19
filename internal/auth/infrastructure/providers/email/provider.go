package email

import (
	"context"
	"fmt"
	"log/slog"
)

type EmailProvider interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
}

// ConsoleProvider logs the Email to the terminal for local development.
// In production, you would implement a SendGridProvider or AWS SESProvider here.
type ConsoleProvider struct {
	log *slog.Logger
}

func NewConsoleProvider(log *slog.Logger) *ConsoleProvider {
	return &ConsoleProvider{log: log}
}

func (p *ConsoleProvider) SendEmail(ctx context.Context, to string, subject string, body string) error {
	p.log.Info("📧 [MOCK EMAIL PROVIDER]", "to", to, "subject", subject, "body", body)
	fmt.Printf("\n📧 Email to %s\nSubject: %s\nBody: %s\n\n", to, subject, body)
	return nil
}
