package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/promotions/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain"
	"github.com/yourorg/ehailing/backend/internal/promotions/domain/entities"
)

type AdminPromotionHandler struct {
	createPromo   *usecases.CreatePromotion
	updatePromo   *usecases.UpdatePromotion
	activatePromo *usecases.ActivatePromotion
	pausePromo    *usecases.PausePromotion
	listPromos    *usecases.ListPromotions
	getPromo      *usecases.GetPromotion
}

func NewAdminPromotionHandler(
	createPromo *usecases.CreatePromotion,
	updatePromo *usecases.UpdatePromotion,
	activatePromo *usecases.ActivatePromotion,
	pausePromo *usecases.PausePromotion,
	listPromos *usecases.ListPromotions,
	getPromo *usecases.GetPromotion,
) *AdminPromotionHandler {
	return &AdminPromotionHandler{
		createPromo:   createPromo,
		updatePromo:   updatePromo,
		activatePromo: activatePromo,
		pausePromo:    pausePromo,
		listPromos:    listPromos,
		getPromo:      getPromo,
	}
}

type CreatePromotionRequest struct {
	Code              string    `json:"code" binding:"required"`
	Name              string    `json:"name" binding:"required"`
	Description       string    `json:"description"`
	DiscountType      string    `json:"discount_type" binding:"required"`
	DiscountValue     float64   `json:"discount_value" binding:"required"`
	MaxDiscountAmount *float64  `json:"max_discount_amount"`
	MinTripFare       float64   `json:"min_trip_fare"`
	UsageLimit        *int      `json:"usage_limit"`
	PerUserLimit      int       `json:"per_user_limit"`
	ValidFrom         time.Time `json:"valid_from" binding:"required"`
	ValidUntil        time.Time `json:"valid_until" binding:"required"`
}

func (h *AdminPromotionHandler) CreatePromotion(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	adminID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	perUserLimit := req.PerUserLimit
	if perUserLimit <= 0 {
		perUserLimit = 1
	}

	promo, err := h.createPromo.Execute(c.Request.Context(), usecases.CreatePromotionInput{
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		DiscountType:      entities.DiscountType(req.DiscountType),
		DiscountValue:     req.DiscountValue,
		MaxDiscountAmount: req.MaxDiscountAmount,
		MinTripFare:       req.MinTripFare,
		UsageLimit:        req.UsageLimit,
		PerUserLimit:      perUserLimit,
		ValidFrom:         req.ValidFrom,
		ValidUntil:        req.ValidUntil,
		CreatedBy:         adminID,
	})
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"promotion": promo})
}

func (h *AdminPromotionHandler) ListPromotions(c *gin.Context) {
	statusStr := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var status *entities.PromotionStatus
	if statusStr != "" {
		s := entities.PromotionStatus(statusStr)
		status = &s
	}

	promos, err := h.listPromos.Execute(c.Request.Context(), usecases.ListPromotionsInput{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotions": promos})
}

func (h *AdminPromotionHandler) GetPromotion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_promotion_id"})
		return
	}

	promo, err := h.getPromo.Execute(c.Request.Context(), id)
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotion": promo})
}

type UpdatePromotionRequest struct {
	Code              string    `json:"code" binding:"required"`
	Name              string    `json:"name" binding:"required"`
	Description       string    `json:"description"`
	DiscountType      string    `json:"discount_type" binding:"required"`
	DiscountValue     float64   `json:"discount_value" binding:"required"`
	MaxDiscountAmount *float64  `json:"max_discount_amount"`
	MinTripFare       float64   `json:"min_trip_fare"`
	UsageLimit        *int      `json:"usage_limit"`
	PerUserLimit      int       `json:"per_user_limit"`
	ValidFrom         time.Time `json:"valid_from" binding:"required"`
	ValidUntil        time.Time `json:"valid_until" binding:"required"`
}

func (h *AdminPromotionHandler) UpdatePromotion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_promotion_id"})
		return
	}

	var req UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	promo, err := h.updatePromo.Execute(c.Request.Context(), usecases.UpdatePromotionInput{
		ID:                id,
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		DiscountType:      entities.DiscountType(req.DiscountType),
		DiscountValue:     req.DiscountValue,
		MaxDiscountAmount: req.MaxDiscountAmount,
		MinTripFare:       req.MinTripFare,
		UsageLimit:        req.UsageLimit,
		PerUserLimit:      req.PerUserLimit,
		ValidFrom:         req.ValidFrom,
		ValidUntil:        req.ValidUntil,
	})
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotion": promo})
}

func (h *AdminPromotionHandler) ActivatePromotion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_promotion_id"})
		return
	}

	promo, err := h.activatePromo.Execute(c.Request.Context(), id)
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotion": promo})
}

func (h *AdminPromotionHandler) PausePromotion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_promotion_id"})
		return
	}

	promo, err := h.pausePromo.Execute(c.Request.Context(), id)
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotion": promo})
}

func handlePromoError(c *gin.Context, err error) {
	switch err {
	case domain.ErrPromotionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion_not_found"})
	case domain.ErrPromotionInactive:
		c.JSON(http.StatusBadRequest, gin.H{"error": "promotion_inactive"})
	case domain.ErrPromotionExpired:
		c.JSON(http.StatusBadRequest, gin.H{"error": "promotion_expired"})
	case domain.ErrPromotionCodeExists:
		c.JSON(http.StatusConflict, gin.H{"error": "promotion_code_exists"})
	case domain.ErrUsageLimitReached:
		c.JSON(http.StatusBadRequest, gin.H{"error": "usage_limit_reached"})
	case domain.ErrPerUserLimitReached:
		c.JSON(http.StatusBadRequest, gin.H{"error": "per_user_limit_reached"})
	case domain.ErrFareBelowMinimum:
		c.JSON(http.StatusBadRequest, gin.H{"error": "fare_below_minimum"})
	case domain.ErrInvalidDiscount:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_discount_value"})
	case domain.ErrAlreadyRedeemed:
		c.JSON(http.StatusConflict, gin.H{"error": "already_redeemed"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}
