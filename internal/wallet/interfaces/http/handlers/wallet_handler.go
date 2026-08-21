package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/wallet/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/wallet/domain"
	"github.com/yourorg/ehailing/backend/internal/wallet/infrastructure/paystack"
)

type WalletHandler struct {
	getWallet          *usecases.GetWallet
	getHistory         *usecases.GetLedgerHistory
	adminTopup         *usecases.AdminTopup
	requestPayout      *usecases.RequestPayout
	getPayoutHistory   *usecases.GetPayoutHistory
	handleTransferHook *usecases.HandleTransferWebhook
	listBanks          *usecases.ListBanks
	paystackTransfer   *paystack.PaystackTransferService
}

func NewWalletHandler(
	getWallet *usecases.GetWallet,
	getHistory *usecases.GetLedgerHistory,
	adminTopup *usecases.AdminTopup,
	requestPayout *usecases.RequestPayout,
	getPayoutHistory *usecases.GetPayoutHistory,
	handleTransferHook *usecases.HandleTransferWebhook,
	listBanks *usecases.ListBanks,
	paystackTransfer *paystack.PaystackTransferService,
) *WalletHandler {
	return &WalletHandler{
		getWallet:          getWallet,
		getHistory:         getHistory,
		adminTopup:         adminTopup,
		requestPayout:      requestPayout,
		getPayoutHistory:   getPayoutHistory,
		handleTransferHook: handleTransferHook,
		listBanks:          listBanks,
		paystackTransfer:   paystackTransfer,
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

type RequestPayoutRequest struct {
	Amount            float64 `json:"amount" binding:"required"`
	BankName          string  `json:"bank_name" binding:"required"`
	BankAccountNumber string  `json:"bank_account_number" binding:"required"`
	BankCode          string  `json:"bank_code" binding:"required"`
}

func (h *WalletHandler) RequestPayout(c *gin.Context) {
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

	var req RequestPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	payout, err := h.requestPayout.Execute(c.Request.Context(), usecases.RequestPayoutInput{
		UserID:            userID,
		Amount:            req.Amount,
		BankName:          req.BankName,
		BankAccountNumber: req.BankAccountNumber,
		BankCode:          req.BankCode,
	})
	if err != nil {
		handleWalletError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"payout": payout, "message": "payout_initiated"})
}

func (h *WalletHandler) GetPayoutHistory(c *gin.Context) {
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

	payouts, err := h.getPayoutHistory.Execute(c.Request.Context(), userID, limit, offset)
	if err != nil {
		handleWalletError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"payouts": payouts})
}

func (h *WalletHandler) ListBanks(c *gin.Context) {
	country := c.DefaultQuery("country", "south africa")

	banks, err := h.listBanks.Execute(c.Request.Context(), country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_banks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"banks": banks})
}

// HandleTransferWebhook receives webhook events from Paystack for transfers
// This endpoint is PUBLIC (no auth) because Paystack calls it directly
func (h *WalletHandler) HandleTransferWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	signature := c.GetHeader("x-paystack-signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
		return
	}

	if !h.paystackTransfer.VerifyWebhookSignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	var event struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
		} `json:"data"`
	}

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	if event.Event != "transfer.success" && event.Event != "transfer.failed" && event.Event != "transfer.reversed" {
		c.JSON(http.StatusOK, gin.H{"message": "event ignored"})
		return
	}

	if err := h.handleTransferHook.Execute(c.Request.Context(), event.Data.Reference, event.Data.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer processed"})
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
