package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
	"github.com/yourorg/ehailing/backend/internal/promotions/application/usecases"
)

type PassengerPromotionHandler struct {
	validatePromo  *usecases.ValidatePromoCode
	getRedemptions *usecases.GetUserRedemptions
}

func NewPassengerPromotionHandler(
	validatePromo *usecases.ValidatePromoCode,
	getRedemptions *usecases.GetUserRedemptions,
) *PassengerPromotionHandler {
	return &PassengerPromotionHandler{
		validatePromo:  validatePromo,
		getRedemptions: getRedemptions,
	}
}

type ValidatePromoRequest struct {
	Code     string  `json:"code" binding:"required"`
	TripFare float64 `json:"trip_fare" binding:"required"`
}

func (h *PassengerPromotionHandler) ValidatePromoCode(c *gin.Context) {
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

	var req ValidatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	result, err := h.validatePromo.Execute(c.Request.Context(), usecases.ValidatePromoCodeInput{
		Code:     req.Code,
		UserID:   userID,
		TripFare: req.TripFare,
	})
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"code":       result.Promotion.Code,
		"name":       result.Promotion.Name,
		"discount":   result.Discount,
		"final_fare": result.FinalFare,
	})
}

func (h *PassengerPromotionHandler) GetMyRedemptions(c *gin.Context) {
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

	redemptions, err := h.getRedemptions.Execute(c.Request.Context(), userID, limit, offset)
	if err != nil {
		handlePromoError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"redemptions": redemptions})
}
