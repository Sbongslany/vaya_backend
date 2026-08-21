package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/support/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/support/domain"
	"github.com/yourorg/ehailing/backend/internal/support/domain/entities"
)

type SupportHandler struct {
	createTicketUC   *usecases.CreateTicket
	getUserTicketsUC *usecases.GetUserTickets
	addCommentUC     *usecases.AddComment
	resolveTicketUC  *usecases.ResolveTicket
}

func NewSupportHandler(
	createTicketUC *usecases.CreateTicket,
	getUserTicketsUC *usecases.GetUserTickets,
	addCommentUC *usecases.AddComment,
	resolveTicketUC *usecases.ResolveTicket,
) *SupportHandler {
	return &SupportHandler{
		createTicketUC:   createTicketUC,
		getUserTicketsUC: getUserTicketsUC,
		addCommentUC:     addCommentUC,
		resolveTicketUC:  resolveTicketUC,
	}
}

type CreateTicketRequest struct {
	TripID      string `json:"trip_id"`
	Category    string `json:"category" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func (h *SupportHandler) CreateTicket(c *gin.Context) {
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

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	var tripID *uuid.UUID
	if req.TripID != "" {
		parsedTripID, err := uuid.Parse(req.TripID)
		if err == nil {
			tripID = &parsedTripID
		}
	}

	ticket, err := h.createTicketUC.Execute(c.Request.Context(), usecases.CreateTicketInput{
		UserID:      userID,
		TripID:      tripID,
		Category:    entities.TicketCategory(req.Category),
		Subject:     req.Subject,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_ticket"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ticket": ticket})
}

func (h *SupportHandler) GetMyTickets(c *gin.Context) {
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

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	tickets, err := h.getUserTicketsUC.Execute(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

type AddCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *SupportHandler) AddComment(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authorID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ticketID, err := uuid.Parse(c.Param("ticketId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_ticket_id"})
		return
	}

	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	comment, err := h.addCommentUC.Execute(c.Request.Context(), usecases.AddCommentInput{
		TicketID: ticketID,
		AuthorID: authorID,
		Content:  req.Content,
	})
	if err != nil {
		if err == domain.ErrTicketNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket_not_found"})
			return
		}
		if err == domain.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_add_comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

type ResolveTicketRequest struct {
	Resolution   string  `json:"resolution" binding:"required"`
	RefundAmount float64 `json:"refund_amount"`
	RefundReason string  `json:"refund_reason"`
}

func (h *SupportHandler) ResolveTicket(c *gin.Context) {
	adminIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ticketID, err := uuid.Parse(c.Param("ticketId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_ticket_id"})
		return
	}

	var req ResolveTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err = h.resolveTicketUC.Execute(c.Request.Context(), usecases.ResolveTicketInput{
		TicketID:     ticketID,
		AdminID:      adminID,
		Resolution:   req.Resolution,
		RefundAmount: req.RefundAmount,
		RefundReason: req.RefundReason,
	})
	if err != nil {
		if err == domain.ErrTicketNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket_not_found"})
			return
		}
		if err == domain.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_already_resolved"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_resolve_ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket_resolved"})
}
