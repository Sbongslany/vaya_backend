package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
)

type WalletHandler struct {
	getWallet  *usecases.GetWallet
	getHistory *usecases.GetLedgerHistory
	adminTopup *usecases.AdminTopup
}

func NewWalletHandler(
	getWallet *usecases.GetWallet,
	getHistory *usecases.GetLedgerHistory,
	adminTopup *usecases.AdminTopup,
) *WalletHandler {
	return &WalletHandler{
		getWallet:  getWallet,
		getHistory: getHistory,
		adminTopup: adminTopup,
	}
}

func (h *WalletHandler) GetMyWallet(c *gin.Context) {
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

	wallet, err := h.getWallet.Execute(c.Request.Context(), userID)
	if err != nil {
		handleWalletError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet": wallet})
}

func (h *WalletHandler) GetMyLedgerHistory(c *gin.Context) {
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

	entries, err := h.getHistory.Execute(c.Request.Context(), userID, limit, offset)
	if err != nil {
		handleWalletError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

type AdminTopupRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description"`
}

func (h *WalletHandler) AdminTopup(c *gin.Context) {
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

	var req AdminTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}

	wallet, err := h.adminTopup.Execute(c.Request.Context(), usecases.AdminTopupInput{
		UserID:      targetUserID,
		Amount:      req.Amount,
		Description: req.Description,
		AdminID:     adminID,
	})
	if err != nil {
		handleWalletError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet": wallet, "message": "topup_successful"})
}

func handleWalletError(c *gin.Context, err error) {
	switch err {
	case domain.ErrWalletNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet_not_found"})
	case domain.ErrInsufficientBalance:
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient_balance"})
	case domain.ErrInvalidAmount:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_amount"})
	case domain.ErrWalletAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "wallet_already_exists"})
	case domain.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}
