package dependency

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/persistence/postgres"
	pgRedis "github.com/yourorg/ehailing/backend/internal/auth/infrastructure/persistence/redis"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/cloudinary"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/email"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/providers/sms"
	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/security"
	"github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/handlers"
	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/config"
)

type AuthContainer struct {
	Handler             *handlers.AuthHandler
	OTPHandler          *handlers.OTPHandler
	VerificationHandler *handlers.VerificationHandler
	PasswordHandler     *handlers.PasswordHandler
	SessionHandler      *handlers.SessionHandler
	AdminHandler        *handlers.AdminAuthHandler
	DriverHandler       *handlers.DriverHandler
	Middleware          *authMiddleware.AuthMiddleware
	AdminMiddleware     *authMiddleware.AdminMiddleware
	RateLimitMiddleware *authMiddleware.RateLimitMiddleware
}

func WireAuth(pool *pgxpool.Pool, redisClient *redis.Client, cfg *config.Config, log *slog.Logger) *AuthContainer {
	// Repositories
	userRepo := postgres.NewUserRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	emailTokenRepo := postgres.NewEmailVerificationRepository(pool)
	passwordResetRepo := postgres.NewPasswordResetRepository(pool)
	mfaRepo := postgres.NewMFARepository(pool)
	auditRepo := postgres.NewAuditRepository(pool)
	driverRepo := postgres.NewDriverRepository(pool)
	docReqRepo := postgres.NewDocumentRequirementRepository(pool)
	docRepo := postgres.NewDocumentRepository(pool) // NEW

	// Services
	passwordSvc := security.NewPasswordService()
	tokenSvc := security.NewTokenService(cfg)
	otpSvc := pgRedis.NewOTPService(redisClient)
	mfaSvc := security.NewMFAService(cfg)
	rateLimiter := pgRedis.NewRateLimiter(redisClient)
	cloudinarySvc := cloudinary.NewCloudinaryService(cfg) // NEW

	// Providers
	// Providers
	smsProvider := sms.NewConsoleProvider(log)
	emailProvider := email.NewGmailSMTPProvider(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFromEmail,
		log,
	)

	appCfg := &usecases.AppConfig{RefreshTTL: cfg.JWTRefreshTTL}

	// Auth Use Cases
	registerUC := usecases.NewRegisterUser(userRepo, passwordSvc, auditRepo)
	loginUC := usecases.NewLoginUser(userRepo, sessionRepo, passwordSvc, tokenSvc, auditRepo, appCfg)
	refreshUC := usecases.NewRefreshToken(userRepo, sessionRepo, tokenSvc, appCfg)
	logoutUC := usecases.NewLogoutUser(sessionRepo)
	logoutAllUC := usecases.NewLogoutAllUsers(sessionRepo)

	// OTP Use Cases
	requestOTPUC := usecases.NewRequestOTP(otpSvc, smsProvider, emailProvider, cfg)
	verifyOTPUC := usecases.NewVerifyOTP(otpSvc, cfg)

	// Verification Use Cases
	verifyPhoneUC := usecases.NewVerifyPhone(userRepo, otpSvc, cfg)
	requestEmailUC := usecases.NewRequestEmailVerification(userRepo, emailTokenRepo, tokenSvc, emailProvider, cfg)
	verifyEmailTokenUC := usecases.NewVerifyEmailToken(userRepo, emailTokenRepo)

	// Password Use Cases
	forgotPasswordUC := usecases.NewForgotPassword(userRepo, passwordResetRepo, tokenSvc, smsProvider, emailProvider)
	resetPasswordUC := usecases.NewResetPassword(userRepo, passwordResetRepo, sessionRepo, passwordSvc)

	// Session Use Cases
	listSessionsUC := usecases.NewListSessions(sessionRepo)
	revokeSessionUC := usecases.NewRevokeSession(sessionRepo)

	// Admin Use Cases
	adminLoginUC := usecases.NewAdminLogin(userRepo, mfaRepo, passwordSvc, tokenSvc)
	adminMFASetupUC := usecases.NewAdminMFASetup(userRepo, mfaRepo, mfaSvc)
	adminMFAConfirmUC := usecases.NewAdminMFAConfirm(mfaRepo, mfaSvc)
	adminMFAVerifyUC := usecases.NewAdminMFAVerify(userRepo, mfaRepo, sessionRepo, mfaSvc, tokenSvc, cfg)

	// Driver Use Cases
	initOnboardingUC := usecases.NewInitiateDriverOnboarding(driverRepo)
	updateProfileUC := usecases.NewUpdateDriverProfile(driverRepo)
	createVehicleUC := usecases.NewCreateVehicle(driverRepo)
	getOnboardingUC := usecases.NewGetOnboardingStatus(driverRepo, docReqRepo)

	// Document Use Cases (NEW)
	generateSignatureUC := usecases.NewGenerateUploadSignature(driverRepo, cloudinarySvc, cfg)
	submitDocumentUC := usecases.NewSubmitDocument(driverRepo, docRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(
		registerUC, loginUC, refreshUC, logoutUC, logoutAllUC, userRepo,
	)
	otpHandler := handlers.NewOTPHandler(requestOTPUC, verifyOTPUC)
	verificationHandler := handlers.NewVerificationHandler(verifyPhoneUC, requestEmailUC, verifyEmailTokenUC)
	passwordHandler := handlers.NewPasswordHandler(forgotPasswordUC, resetPasswordUC)
	sessionHandler := handlers.NewSessionHandler(listSessionsUC, revokeSessionUC)
	adminHandler := handlers.NewAdminAuthHandler(adminLoginUC, adminMFASetupUC, adminMFAConfirmUC, adminMFAVerifyUC)
	driverHandler := handlers.NewDriverHandler(
		initOnboardingUC, updateProfileUC, createVehicleUC, getOnboardingUC,
		generateSignatureUC, submitDocumentUC, // NEW
	)

	// Middleware
	middleware := authMiddleware.NewAuthMiddleware(tokenSvc)
	adminMiddleware := authMiddleware.NewAdminMiddleware(tokenSvc)
	rateLimitMiddleware := authMiddleware.NewRateLimitMiddleware(rateLimiter)

	return &AuthContainer{
		Handler:             authHandler,
		OTPHandler:          otpHandler,
		VerificationHandler: verificationHandler,
		PasswordHandler:     passwordHandler,
		SessionHandler:      sessionHandler,
		AdminHandler:        adminHandler,
		DriverHandler:       driverHandler,
		Middleware:          middleware,
		AdminMiddleware:     adminMiddleware,
		RateLimitMiddleware: rateLimitMiddleware,
	}
}
