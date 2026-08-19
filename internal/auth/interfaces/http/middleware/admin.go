package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type AdminMiddleware struct {
	tokenSvc services.TokenService
}

func NewAdminMiddleware(tokenSvc services.TokenService) *AdminMiddleware {
	return &AdminMiddleware{tokenSvc: tokenSvc}
}

// ValidateMFATicket extracts and validates the short-lived MFA ticket 
// used during the MFA Setup and Confirm flows.
func (m *AdminMiddleware) ValidateMFATicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_mfa_ticket"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_mfa_ticket_format"})
			return
		}

		userID, err := m.tokenSvc.ValidateMFATicket(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_mfa_ticket"})
			return
		}

		c.Set(UserIDKey, userID.String())
		c.Next()
	}
}

// RequireAdmin ensures the user is an Admin AND has passed MFA for this session.
func (m *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get(IsAdminKey)
		if !exists || !isAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin_access_required"})
			return
		}

		mfaVerified, exists := c.Get(MFAVerifiedKey)
		// Note: We need to add MFAVerifiedKey to keys.go. See step 4.
		if !exists || !mfaVerified.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "mfa_verification_required_for_admin_access"})
			return
		}

		c.Next()
	}
}