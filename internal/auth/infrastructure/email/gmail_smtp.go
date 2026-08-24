package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
)

type GmailSMTPProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
	logger   *slog.Logger
}

func NewGmailSMTPProvider(host, port, username, password, from string, logger *slog.Logger) *GmailSMTPProvider {
	return &GmailSMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		logger:   logger,
	}
}

func (p *GmailSMTPProvider) SendEmail(ctx context.Context, to, subject, body string) error {
	// 1. PRINT TO CONSOLE FOR TESTING
	p.logger.Info("📧 SENDING EMAIL (TESTING MODE)", "to", to, "subject", subject, "body", body)

	// If SMTP credentials are empty, just skip sending (acts as a mock)
	if p.username == "" || p.password == "" {
		p.logger.Warn("SMTP not configured. Skipping actual email send.")
		return nil
	}

	// 2. ACTUALLY SEND VIA GMAIL SMTP
	addr := fmt.Sprintf("%s:%s", p.host, p.port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, p.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer c.Close()

	// Start TLS (Required by Gmail)
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: p.host}
		if err = c.StartTLS(config); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Authenticate
	auth := smtp.PlainAuth("", p.username, p.password, p.host)
	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	// Set Sender and Recipient
	if err = c.Mail(p.from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %w", err)
	}

	// Send the Email Body
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		p.from, to, subject, body)

	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close email body: %w", err)
	}

	return c.Quit()
}
