package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/services"
)

type AuthMiddleware struct {
	tokenSvc services.TokenService
}

func NewAuthMiddleware(tokenSvc services.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokenSvc: tokenSvc}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_authorization_header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_authorization_header_format"})
			return
		}

		tokenString := parts[1]
		claims, err := m.tokenSvc.ValidateAccessToken(tokenString)
		if err != nil {
			// Map specific domain errors to standard unauthorized response
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
			return
		}

		// Inject claims into Gin context for handlers to use
		c.Set(UserIDKey, claims.UserID.String())
		c.Set(SessionIDKey, claims.SessionID.String())
		c.Set(RolesKey, claims.Roles)
		c.Set(IsAdminKey, claims.IsAdmin)
		c.Set(MFAVerifiedKey, claims.MFAVerified)

		c.Next()
	}
}

// RequireRole is an optional authorization middleware to enforce specific roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...domain.Role) gin.HandlerFunc {
	roleMap := make(map[domain.Role]struct{})
	for _, r := range requiredRoles {
		roleMap[r] = struct{}{}
	}

	return func(c *gin.Context) {
		rolesInterface, exists := c.Get(RolesKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		userRoles, ok := rolesInterface.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
			return
		}

		authorized := false
		for _, r := range userRoles {
			if _, required := roleMap[domain.Role(r)]; required {
				authorized = true
				break
			}
		}

		if !authorized {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_permissions"})
			return
		}

		c.Next()
	}
}
