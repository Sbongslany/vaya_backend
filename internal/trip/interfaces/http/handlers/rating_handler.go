package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/application/usecases"
)

type RatingHandler struct {
	getUserRating *usecases.GetUserRating
}

func NewRatingHandler(getUserRating *usecases.GetUserRating) *RatingHandler {
	return &RatingHandler{getUserRating: getUserRating}
}

func (h *RatingHandler) GetUserRating(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}

	summary, err := h.getUserRating.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_get_rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      summary.UserID,
		"rating_avg":   summary.RatingAvg,
		"rating_count": summary.RatingCount,
	})
}