package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/auth/application/usecases"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/repositories"
	authMiddleware "github.com/yourorg/ehailing/backend/internal/auth/interfaces/http/middleware"
)

type AuthHandler struct {
	registerUC  *usecases.RegisterUser
	loginUC     *usecases.LoginUser
	refreshUC   *usecases.RefreshToken
	logoutUC    *usecases.LogoutUser
	logoutAllUC *usecases.LogoutAllUsers
	userRepo    repositories.UserRepository
}

func NewAuthHandler(
	register *usecases.RegisterUser,
	login *usecases.LoginUser,
	refresh *usecases.RefreshToken,
	logout *usecases.LogoutUser,
	logoutAll *usecases.LogoutAllUsers,
	userRepo repositories.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		registerUC:  register,
		loginUC:     login,
		refreshUC:   refresh,
		logoutUC:    logout,
		logoutAllUC: logoutAll,
		userRepo:    userRepo,
	}
}

// --- DTOs (Data Transfer Objects) for safe responses ---

type UserProfileResponse struct {
	ID              uuid.UUID  `json:"id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Email           *string    `json:"email,omitempty"`
	Phone           *string    `json:"phone,omitempty"`
	Status          string     `json:"status"`
	Roles           []string   `json:"roles"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
}

// --- Error Handling Helper ---

func handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user_already_exists"})
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrInvalidTokenFormat):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials_or_token"})
	case errors.Is(err, domain.ErrAccountLocked):
		c.JSON(http.StatusForbidden, gin.H{"error": "account_locked"})
	case errors.Is(err, domain.ErrAccountDisabled):
		c.JSON(http.StatusForbidden, gin.H{"error": "account_disabled"})
	case errors.Is(err, domain.ErrSessionExpired), errors.Is(err, domain.ErrSessionRevoked), errors.Is(err, domain.ErrInvalidRefreshToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session_invalid_or_expired"})
	case errors.Is(err, domain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}

// --- Public Endpoints ---

type RegisterRequest struct {
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Password  string  `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "details": err.Error()})
		return
	}

	if req.Email == nil && req.Phone == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email_or_phone_required"})
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	user, err := h.registerUC.Execute(c.Request.Context(), usecases.RegisterUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
		IPAddress: &ip,
		UserAgent: &ua,
	})

	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": user.ID,
		"status":  user.Status,
		"message": "registration_successful_pending_verification",
	})
}

type LoginRequest struct {
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Password   string  `json:"password" binding:"required"`
	DeviceID   *string `json:"device_id"`
	DeviceType *string `json:"device_type"`
	DeviceName *string `json:"device_name"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	res, err := h.loginUC.Execute(c.Request.Context(), usecases.LoginUserRequest{
		Email: req.Email, Phone: req.Phone, Password: req.Password,
		DeviceID: req.DeviceID, DeviceType: req.DeviceType, DeviceName: req.DeviceName,
		IPAddress: &ip, UserAgent: &ua,
	})

	if err != nil {
		handleDomainError(c, err)
		return
	}

	var roles []string
	for _, r := range res.User.Roles {
		roles = append(roles, string(r))
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"session_id":    res.SessionID,
		"expires_in":    900,
		"user": gin.H{ // NEW: Return user profile immediately
			"id":                res.User.ID,
			"first_name":        res.User.FirstName,
			"last_name":         res.User.LastName,
			"email":             res.User.Email,
			"phone":             res.User.Phone,
			"status":            string(res.User.Status),
			"roles":             roles,
			"email_verified_at": res.User.EmailVerifiedAt,
			"phone_verified_at": res.User.PhoneVerifiedAt,
		},
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	res, err := h.refreshUC.Execute(c.Request.Context(), usecases.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_in":    900,
	})
}

// --- Protected Endpoints ---

func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	// Map roles to string slice
	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, string(r))
	}

	// Return safe DTO (PasswordHash is excluded)
	c.JSON(http.StatusOK, UserProfileResponse{
		ID:              user.ID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Email:           user.Email,
		Phone:           user.Phone,
		Status:          string(user.Status),
		Roles:           roles,
		EmailVerifiedAt: user.EmailVerifiedAt,
		PhoneVerifiedAt: user.PhoneVerifiedAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionIDStr, exists := c.Get(authMiddleware.SessionIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	if err := h.logoutUC.Execute(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged_out_successfully"})
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userIDStr, exists := c.Get(authMiddleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	if err := h.logoutAllUC.Execute(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_logout_all"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all_sessions_revoked_successfully"})
}
