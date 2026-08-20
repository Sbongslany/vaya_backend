package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/yourorg/ehailing/backend/internal/auth/dependency"
	"github.com/yourorg/ehailing/backend/internal/config"
	"github.com/yourorg/ehailing/backend/internal/http/handlers"
	"github.com/yourorg/ehailing/backend/internal/http/middleware"
	tripDep "github.com/yourorg/ehailing/backend/internal/trip/dependency"
)

func NewRouter(
	log *slog.Logger,
	postgres *pgxpool.Pool,
	redisClient *redis.Client,
	cfg *config.Config,
	authContainer *dependency.AuthContainer,
	tripContainer *tripDep.TripContainer, // <-- ADD THIS LINE

) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Global Middleware
	engine.Use(
		middleware.RequestID(),
		middleware.Logging(log),
		middleware.Recovery(log),
		middleware.CORS(cfg.CORSAllowedOrigins),
	)

	// Health Routes
	healthHandler := handlers.NewHealthHandler(postgres, redisClient)
	engine.GET("/health", healthHandler.Health)
	engine.GET("/readiness", healthHandler.Readiness)

	// 404 Handler
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "route_not_found",
			"message": "the requested route does not exist",
		})
	})

	// ==========================================
	// PASSENGER / DRIVER AUTH ROUTES
	// ==========================================
	auth := engine.Group("/api/v1/auth")
	{
		// Rate Limit Definitions
		ipLimit := authContainer.RateLimitMiddleware.Limit("ip", 60, 1*time.Minute, func(c *gin.Context) string {
			return c.ClientIP()
		})

		registerLimit := authContainer.RateLimitMiddleware.Limit("register", 5, 1*time.Minute, func(c *gin.Context) string {
			return c.ClientIP()
		})

		loginLimit := authContainer.RateLimitMiddleware.Limit("login", 10, 1*time.Minute, func(c *gin.Context) string {
			return c.ClientIP()
		})

		otpLimit := authContainer.RateLimitMiddleware.Limit("otp", 5, 1*time.Minute, func(c *gin.Context) string {
			return c.ClientIP()
		})

		// Public Routes with Rate Limits
		auth.POST("/register", registerLimit, authContainer.Handler.Register)
		auth.POST("/login", loginLimit, authContainer.Handler.Login)
		auth.POST("/refresh", ipLimit, authContainer.Handler.Refresh)
		auth.POST("/otp/request", otpLimit, authContainer.OTPHandler.Request)
		auth.POST("/otp/verify", ipLimit, authContainer.OTPHandler.Verify)
		auth.POST("/verify/email", ipLimit, authContainer.VerificationHandler.VerifyEmailToken)
		auth.POST("/forgot-password", ipLimit, authContainer.PasswordHandler.ForgotPassword)
		auth.POST("/reset-password", ipLimit, authContainer.PasswordHandler.ResetPassword)

		// Protected Routes (Requires standard JWT)
		protected := auth.Group("")
		protected.Use(authContainer.Middleware.Authenticate())
		{
			protected.GET("/me", authContainer.Handler.GetMe)
			protected.POST("/logout", authContainer.Handler.Logout)
			protected.POST("/logout-all", authContainer.Handler.LogoutAll)

			protected.POST("/verify/phone", authContainer.VerificationHandler.VerifyPhone)
			protected.POST("/verify/email/request", authContainer.VerificationHandler.RequestEmailVerification)

			protected.GET("/sessions", authContainer.SessionHandler.List)
			protected.DELETE("/sessions/:session_id", authContainer.SessionHandler.Revoke)
		}
	}

	// ==========================================
	// ADMIN AUTH ROUTES
	// ==========================================
	adminAuth := engine.Group("/api/v1/admin/auth")
	{
		adminIpLimit := authContainer.RateLimitMiddleware.Limit("admin_ip", 20, 1*time.Minute, func(c *gin.Context) string {
			return c.ClientIP()
		})

		// Public Admin Routes
		adminAuth.POST("/login", adminIpLimit, authContainer.AdminHandler.Login)
		adminAuth.POST("/mfa/verify", adminIpLimit, authContainer.AdminHandler.MFAVerify)

		// Protected Admin Setup Routes (Requires MFA Ticket)
		adminMFASetup := adminAuth.Group("")
		adminMFASetup.Use(authContainer.AdminMiddleware.ValidateMFATicket())
		{
			adminMFASetup.POST("/mfa/setup", authContainer.AdminHandler.MFASetup)
			adminMFASetup.POST("/mfa/confirm", authContainer.AdminHandler.MFAConfirm)
		}

		// Protected Admin Dashboard Routes (Requires Standard JWT + Admin Role + MFA Verified)
		adminProtected := adminAuth.Group("")
		adminProtected.Use(
			authContainer.Middleware.Authenticate(),
			authContainer.AdminMiddleware.RequireAdmin(),
		)
		{
			adminProtected.GET("/me", authContainer.Handler.GetMe)
			adminProtected.POST("/logout", authContainer.Handler.Logout)
		}
	}

	// ==========================================
	// DRIVER ONBOARDING ROUTES (Protected)
	// ==========================================
	driver := engine.Group("/api/v1/driver")
	driver.Use(authContainer.Middleware.Authenticate())
	{
		driver.POST("/onboarding/init", authContainer.DriverHandler.InitOnboarding)
		driver.PUT("/onboarding/profile", authContainer.DriverHandler.UpdateProfile)
		driver.POST("/onboarding/vehicle", authContainer.DriverHandler.CreateVehicle)
		driver.GET("/onboarding/status", authContainer.DriverHandler.GetOnboardingStatus)

		// Document Upload Routes (NEW)
		driver.GET("/onboarding/documents/signature", authContainer.DriverHandler.GetUploadSignature)
		driver.POST("/onboarding/documents", authContainer.DriverHandler.SubmitDocument)
	}

	// Trip Routes (protected)
	trips := engine.Group("/api/v1/trips")
	trips.Use(authContainer.Middleware.Authenticate())
	{
		trips.POST("", tripContainer.Handler.CreateTrip)
		trips.GET("/nearby", tripContainer.Handler.GetNearbyTrips)
		trips.POST("/long-distance", tripContainer.Handler.CreateLongDistanceTrip)
		trips.GET("/long-distance/open", tripContainer.Handler.GetOpenLongDistanceTrips)
		trips.GET("/:id", tripContainer.Handler.GetTrip)
		trips.POST("/:id/offers", tripContainer.Handler.SubmitTripOffer)
		trips.GET("/:id/offers", tripContainer.Handler.GetTripOffers)
		trips.POST("/:id/offers/:offerId/accept", tripContainer.Handler.AcceptTripOffer)
		trips.POST("/:id/confirm", tripContainer.Handler.ConfirmTripAssignment)
		trips.POST("/:id/arrive", tripContainer.Handler.ArriveAtPickup)
		trips.POST("/:id/start", tripContainer.Handler.StartTrip)
		trips.POST("/:id/complete", tripContainer.Handler.CompleteTrip)
		trips.POST("/:id/pay", tripContainer.Handler.ProcessPayment)
		trips.POST("/:id/rate", tripContainer.Handler.SubmitRating)
		trips.POST("/:id/cancel", tripContainer.Handler.CancelTrip)
		// Long-distance execution (outbound)
		trips.POST("/:id/long-distance/publish", tripContainer.Handler.PublishLongDistanceTrip)
		trips.POST("/:id/long-distance/confirm", tripContainer.Handler.ConfirmLongDistanceAssignment)
		trips.POST("/:id/long-distance/schedule", tripContainer.Handler.ScheduleLongDistanceTrip)
		trips.POST("/:id/long-distance/depart", tripContainer.Handler.DepartForPickup)
		trips.POST("/:id/long-distance/outbound/begin", tripContainer.Handler.BeginOutbound)
		trips.POST("/:id/long-distance/outbound/arrive", tripContainer.Handler.ReachOutboundDestination)
		trips.POST("/:id/long-distance/outbound/resolve", tripContainer.Handler.ResolveOutboundArrival)
		// Long-distance execution (return)
		trips.POST("/:id/long-distance/return/schedule", tripContainer.Handler.ScheduleReturn)
		trips.POST("/:id/long-distance/return/start", tripContainer.Handler.StartReturn)
		trips.POST("/:id/long-distance/return/begin", tripContainer.Handler.BeginReturnInProgress)
		trips.POST("/:id/long-distance/return/arrive", tripContainer.Handler.ReachFinalDestination)
		trips.POST("/:id/long-distance/return/complete", tripContainer.Handler.CompleteLongDistanceTrip)
		trips.GET("/:id/events", tripContainer.EventHandler.GetTripHistory)
		trips.GET("/:id/history", tripContainer.EventHandler.GetTripHistory)
	}
	// ==========================================
	// OPENAPI DOCUMENTATION (Only registered once)
	// ==========================================
	engine.GET("/api/v1/docs/openapi.yaml", func(c *gin.Context) {
		c.File("api/openapi/auth.yaml")
	})

	return engine
}
