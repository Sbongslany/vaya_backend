package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
)

type SessionHandler struct {
	listUC   *usecases.ListSessions
	revokeUC *usecases.RevokeSession
}

func NewSessionHandler(list *usecases.ListSessions, revoke *usecases.RevokeSession) *SessionHandler {
	return &SessionHandler{listUC: list, revokeUC: revoke}
}

type SessionResponse struct {
	ID          uuid.UUID `json:"id"`
	DeviceID    *string   `json:"device_id,omitempty"`
	DeviceType  *string   `json:"device_type,omitempty"`
	DeviceName  *string   `json:"device_name,omitempty"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	UserAgent   *string   `json:"user_agent,omitempty"`
	MFAVerified bool      `json:"mfa_verified"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsCurrent   bool      `json:"is_current"`
}

func (h *SessionHandler) List(c *gin.Context) {
	userIDStr, _ := c.Get(authMiddleware.UserIDKey)
	currentSessionIDStr, _ := c.Get(authMiddleware.SessionIDKey)
	
	userID, _ := uuid.Parse(userIDStr.(string))
	currentSessionID, _ := uuid.Parse(currentSessionIDStr.(string))

	sessions, err := h.listUC.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_sessions"})
		return
	}

	var response []SessionResponse
	for _, s := range sessions {
		response = append(response, SessionResponse{
			ID:          s.ID,
			DeviceID:    s.DeviceID,
			DeviceType:  s.DeviceType,
			DeviceName:  s.DeviceName,
			IPAddress:   s.IPAddress,
			UserAgent:   s.UserAgent,
			MFAVerified: s.MFAVerified,
			CreatedAt:   s.CreatedAt,
			LastUsedAt:  s.LastUsedAt,
			ExpiresAt:   s.ExpiresAt,
			IsCurrent:   s.ID == currentSessionID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": response})
}

func (h *SessionHandler) Revoke(c *gin.Context) {
	userIDStr, _ := c.Get(authMiddleware.UserIDKey)
	userID, _ := uuid.Parse(userIDStr.(string))

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id"})
		return
	}

	err = h.revokeUC.Execute(c.Request.Context(), userID, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session_not_found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_revoke_session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session_revoked_successfully"})
}