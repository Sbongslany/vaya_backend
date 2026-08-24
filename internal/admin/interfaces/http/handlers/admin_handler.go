package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/admin/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/admin/domain/entities"
	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
)

type AdminHandler struct {
	getOverviewUC      *usecases.GetPlatformOverview
	getFinancialUC     *usecases.GetFinancialSummary
	listUsersUC        *usecases.ListUsers
	updateUserStatusUC *usecases.UpdateUserStatus
	listTripsUC        *usecases.ListAllTrips
	getLiveMapUC       *usecases.GetLiveMap
	forceCancelUC      *usecases.ForceCancelTrip
	forceCompleteUC    *usecases.ForceCompleteTrip
	getActiveSOSUC     *usecases.GetActiveSOS
}

func NewAdminHandler(
	getOverviewUC *usecases.GetPlatformOverview,
	getFinancialUC *usecases.GetFinancialSummary,
	listUsersUC *usecases.ListUsers,
	updateUserStatusUC *usecases.UpdateUserStatus,
	listTripsUC *usecases.ListAllTrips,
	getLiveMapUC *usecases.GetLiveMap,
	forceCancelUC *usecases.ForceCancelTrip,
	forceCompleteUC *usecases.ForceCompleteTrip,
	getActiveSOSUC *usecases.GetActiveSOS,
) *AdminHandler {
	return &AdminHandler{
		getOverviewUC:      getOverviewUC,
		getFinancialUC:     getFinancialUC,
		listUsersUC:        listUsersUC,
		updateUserStatusUC: updateUserStatusUC,
		listTripsUC:        listTripsUC,
		getLiveMapUC:       getLiveMapUC,
		forceCancelUC:      forceCancelUC,
		forceCompleteUC:    forceCompleteUC,
		getActiveSOSUC:     getActiveSOSUC,
	}
}

// --- Phase 17: Existing Dashboard Handlers ---

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	role := c.Query("role")
	status := c.Query("status")

	users, err := h.listUsersUC.Execute(c.Request.Context(), usecases.ListUsersInput{
		Limit: limit, Offset: offset, Role: role, Status: status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

type UpdateUserStatusRequest struct {
	Status entities.UserStatus `json:"status" binding:"required"`
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
		UserID: userID, Status: req.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user_status_updated"})
}

func (h *AdminHandler) ListAllTrips(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")

	trips, err := h.listTripsUC.Execute(c.Request.Context(), usecases.ListAllTripsInput{
		Limit: limit, Offset: offset, Status: status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_trips"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trips": trips})
}

// --- Phase B: New Live Operations Handlers ---

func (h *AdminHandler) GetLiveMap(c *gin.Context) {
	trips, drivers, err := h.getLiveMapUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_live_map"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trips": trips, "drivers": drivers})
}

type ForceCancelTripRequest struct {
	TripID string `json:"trip_id" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

func (h *AdminHandler) ForceCancelTrip(c *gin.Context) {
	adminIDStr, _ := c.Get(authMiddleware.UserIDKey)
	adminID, _ := uuid.Parse(adminIDStr.(string))

	var req ForceCancelTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	err = h.forceCancelUC.Execute(c.Request.Context(), usecases.ForceCancelTripInput{
		AdminID: adminID, TripID: tripID, Reason: req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_force_cancel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "trip_force_cancelled"})
}

type ForceCompleteTripRequest struct {
	TripID string `json:"trip_id" binding:"required"`
}

func (h *AdminHandler) ForceCompleteTrip(c *gin.Context) {
	adminIDStr, _ := c.Get(authMiddleware.UserIDKey)
	adminID, _ := uuid.Parse(adminIDStr.(string))

	var req ForceCompleteTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_trip_id"})
		return
	}

	err = h.forceCompleteUC.Execute(c.Request.Context(), usecases.ForceCompleteTripInput{
		AdminID: adminID, TripID: tripID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_force_complete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "trip_force_completed"})
}

func (h *AdminHandler) GetActiveSOS(c *gin.Context) {
	alerts, err := h.getActiveSOSUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_sos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
