package usecases

import (
	"context"
	"fmt"

	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/email"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/sms"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type RequestOTP struct {
	otpSvc        services.OTPService
	smsProvider   sms.SMSProvider
	emailProvider email.EmailProvider
	cfg           *config.Config
}

func NewRequestOTP(otpSvc services.OTPService, smsProv sms.SMSProvider, emailProv email.EmailProvider, cfg *config.Config) *RequestOTP {
	return &RequestOTP{otpSvc: otpSvc, smsProvider: smsProv, emailProvider: emailProv, cfg: cfg}
}

func (uc *RequestOTP) Execute(ctx context.Context, identifier string, purpose domain.OTPPurpose, channel domain.OTPChannel) error {
	inCooldown, err := uc.otpSvc.IsInCooldown(ctx, identifier, purpose)
	if err != nil {
		return err
	}
	if inCooldown {
		return domain.ErrOTPCooldownActive
	}

	otp, err := uc.otpSvc.GenerateOTP(uc.cfg.OTPLength)
	if err != nil {
		return err
	}

	if err := uc.otpSvc.StoreOTP(ctx, identifier, purpose, otp, uc.cfg.OTPTTL); err != nil {
		return err
	}

	if err := uc.otpSvc.SetCooldown(ctx, identifier, purpose, uc.cfg.OTPResendCooldown); err != nil {
		return err
	}

	message := fmt.Sprintf("Your verification code is: %s. It expires in %d minutes.", otp, int(uc.cfg.OTPTTL.Minutes()))

	if channel == domain.OTPChannelSMS {
		return uc.smsProvider.SendSMS(ctx, identifier, message)
	}

	subject := "Your Verification Code"
	return uc.emailProvider.SendEmail(ctx, identifier, subject, message)
}
