package domain

type Role string

const (
	RolePassenger     Role = "PASSENGER"
	RoleDriver        Role = "DRIVER"
	RoleAdmin         Role = "ADMIN"
	RoleSuperAdmin    Role = "SUPER_ADMIN"
	RoleSupportAdmin  Role = "SUPPORT_ADMIN"
	RoleSafetyAdmin   Role = "SAFETY_ADMIN"
	RoleFinanceAdmin  Role = "FINANCE_ADMIN"
)

type UserStatus string

const (
	StatusPendingVerification UserStatus = "PENDING_VERIFICATION"
	StatusActive              UserStatus = "ACTIVE"
	StatusLocked              UserStatus = "LOCKED"
	StatusDisabled            UserStatus = "DISABLED"
	StatusDeleted             UserStatus = "DELETED"
)

type OTPPurpose string

const (
	OTPPurposePhoneVerification OTPPurpose = "PHONE_VERIFICATION"
	OTPPurposeEmailVerification OTPPurpose = "EMAIL_VERIFICATION"
	OTPPurposeLogin             OTPPurpose = "LOGIN_VERIFICATION"
	OTPPurposePasswordReset     OTPPurpose = "PASSWORD_RESET"
	OTPPurposeMFASetup          OTPPurpose = "MFA_SETUP"
)

type OTPChannel string

const (
	OTPChannelSMS   OTPChannel = "SMS"
	OTPChannelEmail OTPChannel = "EMAIL"
)

type AuditAction string

const (
	AuditActionRegisterStarted       AuditAction = "REGISTER_STARTED"
	AuditActionRegisterCompleted     AuditAction = "REGISTER_COMPLETED"
	AuditActionLoginSuccess          AuditAction = "LOGIN_SUCCESS"
	AuditActionLoginFailed           AuditAction = "LOGIN_FAILED"
	AuditActionOTPRequested          AuditAction = "OTP_REQUESTED"
	AuditActionOTPVerified           AuditAction = "OTP_VERIFIED"
	AuditActionOTPFailed             AuditAction = "OTP_FAILED"
	AuditActionEmailVerificationSent AuditAction = "EMAIL_VERIFICATION_SENT"
	AuditActionEmailVerified         AuditAction = "EMAIL_VERIFIED"
	AuditActionPasswordResetRequest  AuditAction = "PASSWORD_RESET_REQUESTED"
	AuditActionPasswordResetComplete AuditAction = "PASSWORD_RESET_COMPLETED"
	AuditActionPasswordResetFailed   AuditAction = "PASSWORD_RESET_FAILED"
	AuditActionSessionCreated        AuditAction = "SESSION_CREATED"
	AuditActionSessionRevoked        AuditAction = "SESSION_REVOKED"
	AuditActionLogout                AuditAction = "LOGOUT"
	AuditActionLogoutAll             AuditAction = "LOGOUT_ALL"
	AuditActionMFASuccess            AuditAction = "MFA_SUCCESS"
	AuditActionMFAFailed             AuditAction = "MFA_FAILED"
	AuditActionAccessDenied          AuditAction = "ACCESS_DENIED"
	AuditActionRateLimited           AuditAction = "RATE_LIMITED"
)