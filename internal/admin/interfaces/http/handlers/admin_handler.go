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
	// Phase 17: Dashboard
	getOverviewUC      *usecases.GetPlatformOverview
	getFinancialUC     *usecases.GetFinancialSummary
	listUsersUC        *usecases.ListUsers
	updateUserStatusUC *usecases.UpdateUserStatus
	listTripsUC        *usecases.ListAllTrips

	// Phase B: Live Operations
	getLiveMapUC    *usecases.GetLiveMap
	forceCancelUC   *usecases.ForceCancelTrip
	forceCompleteUC *usecases.ForceCompleteTrip
	getActiveSOSUC  *usecases.GetActiveSOS

	// Phase C: Financial Controls (Payouts)
	getPendingPayoutsUC     *usecases.GetPendingPayouts
	processPayoutApprovalUC *usecases.ProcessPayoutApproval
	rejectPayoutUC          *usecases.RejectPayout

	// Phase 3: Admin Management
	adminMgmtUCs *usecases.AdminManagementUCs
}

func NewAdminHandler(
	// Phase 17
	getOverviewUC *usecases.GetPlatformOverview,
	getFinancialUC *usecases.GetFinancialSummary,
	listUsersUC *usecases.ListUsers,
	updateUserStatusUC *usecases.UpdateUserStatus,
	listTripsUC *usecases.ListAllTrips,
	// Phase B
	getLiveMapUC *usecases.GetLiveMap,
	forceCancelUC *usecases.ForceCancelTrip,
	forceCompleteUC *usecases.ForceCompleteTrip,
	getActiveSOSUC *usecases.GetActiveSOS,
	// Phase C
	getPendingPayoutsUC *usecases.GetPendingPayouts,
	processPayoutApprovalUC *usecases.ProcessPayoutApproval,
	rejectPayoutUC *usecases.RejectPayout,
	adminMgmtUCs *usecases.AdminManagementUCs, // <-- ADD THIS

) *AdminHandler {
	return &AdminHandler{
		getOverviewUC:           getOverviewUC,
		getFinancialUC:          getFinancialUC,
		listUsersUC:             listUsersUC,
		updateUserStatusUC:      updateUserStatusUC,
		listTripsUC:             listTripsUC,
		getLiveMapUC:            getLiveMapUC,
		forceCancelUC:           forceCancelUC,
		forceCompleteUC:         forceCompleteUC,
		getActiveSOSUC:          getActiveSOSUC,
		getPendingPayoutsUC:     getPendingPayoutsUC,
		processPayoutApprovalUC: processPayoutApprovalUC,
		rejectPayoutUC:          rejectPayoutUC,
		adminMgmtUCs:            adminMgmtUCs, // <-- ADD THIS
	}
}

// --- Phase 17: Existing Dashboard Handlers ---

func (h *AdminHandler) GetPlatformOverview(c *gin.Context) {
	stats, err := h.getOverviewUC.Execute(c.Request.Context())
	if err != nil {
		// We added "details" to expose the exact database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_fetch_stats",
			"details": err.Error(),
		})
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
		// We added "details" to expose the exact database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_fetch_users",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_trips", "details": err.Error()})
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

// --- Phase C: Payout Approval Handlers ---

func (h *AdminHandler) GetPendingPayouts(c *gin.Context) {
	payouts, err := h.getPendingPayoutsUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_payouts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payouts": payouts})
}

type PayoutActionRequest struct {
	PayoutID string `json:"payout_id" binding:"required"`
	Reason   string `json:"reason"`
}

func (h *AdminHandler) ApprovePayout(c *gin.Context) {
	adminIDStr, _ := c.Get(authMiddleware.UserIDKey)
	adminID, _ := uuid.Parse(adminIDStr.(string))

	var req PayoutActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	payoutID, err := uuid.Parse(req.PayoutID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payout_id"})
		return
	}

	// This now triggers the Paystack API and updates the DB
	if err := h.processPayoutApprovalUC.Execute(c.Request.Context(), usecases.ProcessPayoutApprovalInput{
		AdminID: adminID, PayoutID: payoutID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_process_payout", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payout_approved_and_transfer_initiated"})
}

func (h *AdminHandler) RejectPayout(c *gin.Context) {
	adminIDStr, _ := c.Get(authMiddleware.UserIDKey)
	adminID, _ := uuid.Parse(adminIDStr.(string))

	var req PayoutActionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request_or_missing_reason"})
		return
	}

	payoutID, err := uuid.Parse(req.PayoutID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payout_id"})
		return
	}

	if err := h.rejectPayoutUC.Execute(c.Request.Context(), usecases.RejectPayoutInput{
		AdminID: adminID, PayoutID: payoutID, Reason: req.Reason,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_reject_payout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payout_rejected"})
}

// --- Phase 3: Admin Management Handlers ---

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	admins, err := h.adminMgmtUCs.ListAdmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_fetch_admins"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"admins": admins})
}

type CreateAdminRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Role      string `json:"role" binding:"required"`
}

func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tempPwd, err := h.adminMgmtUCs.CreateAdmin(c.Request.Context(), usecases.CreateAdminInput{
		FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, Role: req.Role,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_admin", "details": err.Error()})
		return
	}

	// Return the temporary password so the Super Admin can share it securely
	c.JSON(http.StatusCreated, gin.H{
		"message":            "admin_created_successfully",
		"temporary_password": tempPwd,
		"instruction":        "Share this password securely with the new admin. They should change it on first login.",
	})
}

type UpdateAdminRoleRequest struct {
	TargetAdminID string `json:"target_admin_id" binding:"required"`
	Role          string `json:"role" binding:"required"`
}

func (h *AdminHandler) UpdateAdminRole(c *gin.Context) {
	adminIDStr, _ := c.Get(authMiddleware.UserIDKey)
	actingAdminID, _ := uuid.Parse(adminIDStr.(string))

	var req UpdateAdminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	targetID, err := uuid.Parse(req.TargetAdminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_target_id"})
		return
	}

	if err := h.adminMgmtUCs.UpdateAdminRole(c.Request.Context(), usecases.UpdateAdminRoleInput{
		ActingAdminID: actingAdminID, TargetAdminID: targetID, Role: req.Role,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "admin_role_updated"})
}

type UpdateAdminStatusRequest struct {
	TargetAdminID string `json:"target_admin_id" binding:"required"`
	Status        string `json:"status" binding:"required"` // ACTIVE or DISABLED
}

func (h *AdminHandler) UpdateAdminStatus(c *gin.Context) {
	adminIDStr, _ := c.Get(authMiddleware.UserIDKey)
	actingAdminID, _ := uuid.Parse(adminIDStr.(string))

	var req UpdateAdminStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	targetID, err := uuid.Parse(req.TargetAdminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_target_id"})
		return
	}

	if err := h.adminMgmtUCs.UpdateAdminStatus(c.Request.Context(), usecases.UpdateAdminStatusInput{
		ActingAdminID: actingAdminID, TargetAdminID: targetID, Status: req.Status,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "admin_status_updated"})
}
