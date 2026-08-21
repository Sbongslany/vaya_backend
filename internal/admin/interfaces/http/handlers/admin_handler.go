package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/admin/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
)

type AdminHandler struct {
	getOverviewUC      *usecases.GetPlatformOverview
	getFinancialUC     *usecases.GetFinancialSummary
	listUsersUC        *usecases.ListUsers
	updateUserStatusUC *usecases.UpdateUserStatus
	listTripsUC        *usecases.ListAllTrips
}

func NewAdminHandler(
	getOverviewUC *usecases.GetPlatformOverview,
	getFinancialUC *usecases.GetFinancialSummary,
	listUsersUC *usecases.ListUsers,
	updateUserStatusUC *usecases.UpdateUserStatus,
	listTripsUC *usecases.ListAllTrips,
) *AdminHandler {
	return &AdminHandler{
		getOverviewUC:      getOverviewUC,
		getFinancialUC:     getFinancialUC,
		listUsersUC:        listUsersUC,
		updateUserStatusUC: updateUserStatusUC,
		listTripsUC:        listTripsUC,
	}
}

func (h *AdminHandler) GetPlatformOverview(c *gin.Context) {
	stats, err := h.getOverviewUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (h *AdminHandler) GetFinancialSummary(c *gin.Context) {
	summary, err := h.getFinancialUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_financials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"financials": summary})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	roleStr := c.Query("role")
	statusStr := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var role *entities.UserRole
	if roleStr != "" {
		r := entities.UserRole(roleStr)
		role = &r
	}

	var status *entities.UserStatus
	if statusStr != "" {
		s := entities.UserStatus(statusStr)
		status = &s
	}

	users, err := h.listUsersUC.Execute(c.Request.Context(), usecases.ListUsersInput{
		Role:   role,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	err = h.updateUserStatusUC.Execute(c.Request.Context(), usecases.UpdateUserStatusInput{
		UserID: userID,
		Status: entities.UserStatus(req.Status),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user_status_updated"})
}

func (h *AdminHandler) ListAllTrips(c *gin.Context) {
	statusStr := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var status *string
	if statusStr != "" {
		status = &statusStr
	}

	trips, err := h.listTripsUC.Execute(c.Request.Context(), usecases.ListAllTripsInput{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_trips"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trips": trips})
}
