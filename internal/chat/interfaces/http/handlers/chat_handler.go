package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/chat/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/chat/domain"
	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/chat/infrastructure/websocket"
)

type ChatHandler struct {
	chatHub        *websocket.ChatHub
	sendMessageUC  *usecases.SendMessage
	getHistoryUC   *usecases.GetChatHistory
	initiateCallUC *usecases.InitiateCall
	updateCallUC   *usecases.UpdateCallStatus
}

func NewChatHandler(
	chatHub *websocket.ChatHub,
	sendMessageUC *usecases.SendMessage,
	getHistoryUC *usecases.GetChatHistory,
	initiateCallUC *usecases.InitiateCall,
	updateCallUC *usecases.UpdateCallStatus,
) *ChatHandler {
	return &ChatHandler{
		chatHub:        chatHub,
		sendMessageUC:  sendMessageUC,
		getHistoryUC:   getHistoryUC,
		initiateCallUC: initiateCallUC,
		updateCallUC:   updateCallUC,
	}
}

// ServeWS upgrades HTTP to WebSocket for real-time chat/call signaling
func (h *ChatHandler) ServeWS(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tripIDStr := c.Query("trip_id")
	if tripIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_trip_id"})
		return
	}
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	h.chatHub.ServeWS(c.Writer, c.Request, userID, tripID)
}

func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("tripId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	messages, err := h.getHistoryUC.Execute(c.Request.Context(), tripID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

type SendMessageRequest struct {
	TripID     string `json:"trip_id" binding:"required"`
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	senderID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}
	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_receiver_id"})
		return
	}

	message, err := h.sendMessageUC.Execute(c.Request.Context(), usecases.SendMessageInput{
		TripID:     tripID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_send_message"})
		return
	}

	// Broadcast to WebSocket
	h.chatHub.Broadcast(tripID, "chat_message", message)

	c.JSON(http.StatusCreated, gin.H{"message": message})
}

type InitiateCallRequest struct {
	TripID     string `json:"trip_id" binding:"required"`
	ReceiverID string `json:"receiver_id" binding:"required"`
}

func (h *ChatHandler) InitiateCall(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	callerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req InitiateCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}
	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_receiver_id"})
		return
	}

	result, err := h.initiateCallUC.Execute(c.Request.Context(), usecases.InitiateCallInput{
		TripID:     tripID,
		CallerID:   callerID,
		ReceiverID: receiverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_initiate_call"})
		return
	}

	// Broadcast incoming call to the receiver via WebSocket
	h.chatHub.Broadcast(tripID, "incoming_call", map[string]interface{}{
		"session_id": result.SessionID,
		"caller_id":  callerID,
		"token":      result.Token,
	})

	c.JSON(http.StatusCreated, gin.H{
		"session_id": result.SessionID,
		"token":      result.Token,
	})
}

type UpdateCallStatusRequest struct {
	Status   string `json:"status" binding:"required"`
	Duration int    `json:"duration"`
}

func (h *ChatHandler) UpdateCallStatus(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id"})
		return
	}

	var req UpdateCallStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err = h.updateCallUC.Execute(c.Request.Context(), usecases.UpdateCallStatusInput{
		SessionID: sessionID,
		Status:    entities.CallStatus(req.Status),
		Duration:  req.Duration,
	})
	if err != nil {
		if err == domain.ErrCallNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "call_session_not_found"})
			return
		}
		if err == domain.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status_transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_call"})
		return
	}

	// Note: In a production app, you would fetch the trip_id here to broadcast.
	// For now, the mobile app handles the UI state locally upon receiving the HTTP 200.

	c.JSON(http.StatusOK, gin.H{"message": "call_status_updated"})
}
