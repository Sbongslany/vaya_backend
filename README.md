(Phase 1 - 12)
Production-ready, enterprise-grade authentication and identity system:
Clean Architecture: Strict separation of Domain, Use Cases, Infrastructure, and HTTP layers.
Core Auth: Secure registration, login, and JWT access + opaque refresh token rotation.
Verification: Redis-backed OTP (with SMS/Email abstractions) and secure email verification tokens.
Recovery: Forgot/Reset password flows that automatically revoke all active sessions.
Session Management: Multi-device support with the ability to view and revoke specific sessions.
Admin Security: Isolated admin routes with strict TOTP Multi-Factor Authentication (MFA) and AES-256-GCM encrypted secrets.
Security Hardening: Redis rate limiting to prevent brute-force attacks and PostgreSQL audit logging for all security events.# vaya_backend
