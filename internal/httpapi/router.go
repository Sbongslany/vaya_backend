package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	adminDep "github.com/yourorg/ehailing/backend/internal/admin/dependency"
	"github.com/yourorg/ehailing/backend/internal/auth/dependency"
	chatDep "github.com/yourorg/ehailing/backend/internal/chat/dependency"
	"github.com/yourorg/ehailing/backend/internal/config"
	driverDep "github.com/yourorg/ehailing/backend/internal/driver/dependency"
	geofenceDep "github.com/yourorg/ehailing/backend/internal/geofence/dependency"
	"github.com/yourorg/ehailing/backend/internal/http/handlers"
	"github.com/yourorg/ehailing/backend/internal/http/middleware"
	kycDep "github.com/yourorg/ehailing/backend/internal/kyc/dependency"
	promoDep "github.com/yourorg/ehailing/backend/internal/promotions/dependency"
	safetyDep "github.com/yourorg/ehailing/backend/internal/safety/dependency"
	supportDep "github.com/yourorg/ehailing/backend/internal/support/dependency"
	tripDep "github.com/yourorg/ehailing/backend/internal/trip/dependency"
	walletDep "github.com/yourorg/ehailing/backend/internal/wallet/dependency"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// This imports the generated docs folder (we will create it in Step 4)
	_ "github.com/yourorg/ehailing/backend/docs"

	settingsDep "github.com/yourorg/ehailing/backend/internal/settings/dependency"
)

func NewRouter(
	log *slog.Logger,
	postgres *pgxpool.Pool,
	redisClient *redis.Client,
	cfg *config.Config,
	authContainer *dependency.AuthContainer,
	tripContainer *tripDep.TripContainer,
	promosContainer *promoDep.PromotionsContainer,
	driverContainer *driverDep.DriverContainer,
	walletContainer *walletDep.WalletContainer,
	chatContainer *chatDep.ChatContainer,
	supportContainer *supportDep.SupportContainer,
	adminContainer *adminDep.AdminContainer,
	kycContainer *kycDep.KYCContainer,
	geofenceContainer *geofenceDep.GeofenceContainer,
	safetyContainer *safetyDep.SafetyContainer,
	settingsContainer *settingsDep.SettingsContainer,

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

		// Document Upload Routes
		driver.GET("/onboarding/documents/signature", authContainer.DriverHandler.GetUploadSignature)
		driver.POST("/onboarding/documents", authContainer.DriverHandler.SubmitDocument)
	}

	// ==========================================
	// TRIP ROUTES (Protected)
	// ==========================================
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
		trips.POST("/multi-stop", tripContainer.Handler.CreateMultiStopTrip)

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

		// Events & History
		trips.GET("/:id/events", tripContainer.EventHandler.GetTripHistory)
		trips.GET("/:id/history", tripContainer.EventHandler.GetTripHistory)

		// WebSocket & Devices
		trips.GET("/ws", tripContainer.WSHandler.ServeWS)
		trips.POST("/devices/token", tripContainer.DeviceHandler.RegisterToken)

		// Ratings
		trips.GET("/ratings/:userId", tripContainer.RatingHandler.GetUserRating)

		// Promotions (passenger)
		trips.POST("/promotions/validate", promosContainer.PassengerHandler.ValidatePromoCode)
		trips.GET("/promotions/my-redemptions", promosContainer.PassengerHandler.GetMyRedemptions)

		// Payment
		trips.POST("/:id/pay/initiate", tripContainer.PaymentHandler.InitiatePayment)

		trips.GET("/:id/route", tripContainer.Handler.GetTripRoute)
		trips.POST("/route/calculate", tripContainer.Handler.CalculateRoute)

		trips.GET("/surge", tripContainer.Handler.GetSurgeMultiplier)
		trips.GET("/surge/heatmap", tripContainer.Handler.GetSurgeHeatmap)
	}

	// ==========================================
	// SAFETY ROUTES
	// ==========================================

	// Protected Safety Routes (Passenger/Driver)
	safetyRoutes := engine.Group("/api/v1/safety")
	safetyRoutes.Use(authContainer.Middleware.Authenticate())
	{
		safetyRoutes.POST("/sos", safetyContainer.Handler.TriggerSOS)
		safetyRoutes.POST("/share-link", safetyContainer.Handler.GenerateShareLink)
	}

	// Public Safety Routes (Unauthenticated tracking for trusted contacts)
	engine.GET("/track/:token", safetyContainer.Handler.ViewSharedTrip)

	// ==========================================
	// ADMIN PROMOTIONS ROUTES (Protected + Admin Role)
	// ==========================================
	adminPromos := engine.Group("/api/v1/admin/promotions")
	adminPromos.Use(authContainer.Middleware.Authenticate())
	{
		adminPromos.POST("", promosContainer.AdminHandler.CreatePromotion)
		adminPromos.GET("", promosContainer.AdminHandler.ListPromotions)
		adminPromos.GET("/:id", promosContainer.AdminHandler.GetPromotion)
		adminPromos.PUT("/:id", promosContainer.AdminHandler.UpdatePromotion)
		adminPromos.POST("/:id/activate", promosContainer.AdminHandler.ActivatePromotion)
		adminPromos.POST("/:id/pause", promosContainer.AdminHandler.PausePromotion)
	}

	// ==========================================
	// DRIVER REALTIME & STATE ROUTES (Protected)
	// ==========================================
	driverState := engine.Group("/api/v1/driver")
	driverState.Use(authContainer.Middleware.Authenticate())
	{
		driverState.POST("/online", driverContainer.Handler.GoOnline)
		driverState.POST("/offline", driverContainer.Handler.GoOffline)
		driverState.POST("/location", driverContainer.Handler.UpdateLocation)
		driverState.GET("/nearby", driverContainer.Handler.GetNearbyDrivers)
	}

	// ==========================================
	// WALLET ROUTES (Protected)
	// ==========================================
	walletRoutes := engine.Group("/api/v1/wallet")
	walletRoutes.Use(authContainer.Middleware.Authenticate())
	{
		walletRoutes.GET("", walletContainer.Handler.GetMyWallet)
		walletRoutes.GET("/history", walletContainer.Handler.GetMyLedgerHistory)
		walletRoutes.GET("/payouts", walletContainer.Handler.GetPayoutHistory)
		walletRoutes.POST("/payouts/request", walletContainer.Handler.RequestPayout)
		walletRoutes.GET("/banks", walletContainer.Handler.ListBanks)
	}

	// ==========================================
	// ADMIN WALLET ROUTES (Protected + Admin Role)
	// ==========================================
	adminWalletRoutes := engine.Group("/api/v1/admin/wallet")
	adminWalletRoutes.Use(authContainer.Middleware.Authenticate())
	{
		adminWalletRoutes.POST("/topup", walletContainer.Handler.AdminTopup)
	}

	// ==========================================
	// CHAT & CALL ROUTES (Protected)
	// ==========================================
	chatRoutes := engine.Group("/api/v1/chat")
	chatRoutes.Use(authContainer.Middleware.Authenticate())
	{
		chatRoutes.GET("/ws", chatContainer.Handler.ServeWS)
		chatRoutes.GET("/:tripId/history", chatContainer.Handler.GetChatHistory)
		chatRoutes.POST("/send", chatContainer.Handler.SendMessage)
	}

	callRoutes := engine.Group("/api/v1/calls")
	callRoutes.Use(authContainer.Middleware.Authenticate())
	{
		callRoutes.POST("/initiate", chatContainer.Handler.InitiateCall)
		callRoutes.POST("/:sessionId/status", chatContainer.Handler.UpdateCallStatus)
	}

	// ==========================================
	// PAYSTACK WEBHOOKS (PUBLIC - No Auth)
	// ==========================================
	engine.POST("/api/v1/payments/paystack/webhook", tripContainer.PaymentHandler.HandlePaystackWebhook)
	engine.POST("/api/v1/payments/paystack/transfer-webhook", walletContainer.Handler.HandleTransferWebhook)

	// ==========================================
	// SUPPORT TICKET ROUTES (Protected)
	// ==========================================
	supportRoutes := engine.Group("/api/v1/support")
	supportRoutes.Use(authContainer.Middleware.Authenticate())
	{
		supportRoutes.POST("/tickets", supportContainer.Handler.CreateTicket)
		supportRoutes.GET("/tickets", supportContainer.Handler.GetMyTickets)
		supportRoutes.POST("/tickets/:ticketId/comments", supportContainer.Handler.AddComment)
	}

	// Admin Support Routes
	adminSupportRoutes := engine.Group("/api/v1/admin/support")
	adminSupportRoutes.Use(authContainer.Middleware.Authenticate())
	{
		adminSupportRoutes.POST("/tickets/:ticketId/resolve", supportContainer.Handler.ResolveTicket)
	}

	// ==========================================
	// ADMIN DASHBOARD ROUTES (Protected + Admin Role)
	// ==========================================
	adminDashboard := engine.Group("/api/v1/admin/dashboard")
	adminDashboard.Use(
		authContainer.Middleware.Authenticate(),
		authContainer.AdminMiddleware.RequireAdmin(),
	)
	{
		// Phase 17: Existing Dashboard Stats & Management
		adminDashboard.GET("/overview", adminContainer.Handler.GetPlatformOverview)
		adminDashboard.GET("/financials", adminContainer.Handler.GetFinancialSummary)
		adminDashboard.GET("/users", adminContainer.Handler.ListUsers)
		adminDashboard.PATCH("/users/:userId/status", adminContainer.Handler.UpdateUserStatus)
		adminDashboard.GET("/trips", adminContainer.Handler.ListAllTrips)

		// Phase B: Live Operations & Trip Intervention (NEW)
		adminDashboard.GET("/live-map", adminContainer.Handler.GetLiveMap)
		adminDashboard.POST("/trips/force-cancel", adminContainer.Handler.ForceCancelTrip)
		adminDashboard.POST("/trips/force-complete", adminContainer.Handler.ForceCompleteTrip)
		adminDashboard.GET("/safety/sos/active", adminContainer.Handler.GetActiveSOS)
	}

	// ==========================================
	// ADMIN KYC ROUTES (Protected + Admin Role)
	// ==========================================
	adminKYC := engine.Group("/api/v1/admin/kyc")
	adminKYC.Use(
		authContainer.Middleware.Authenticate(),
		authContainer.AdminMiddleware.RequireAdmin(),
	)
	{
		adminKYC.GET("/pending", kycContainer.Handler.ListPendingKYC)
		adminKYC.GET("/drivers/:userId/documents", kycContainer.Handler.GetDriverDocuments)
		adminKYC.POST("/documents/:documentId/review", kycContainer.Handler.ReviewDocument)
	}

	// ==========================================
	// ADMIN GEOFENCE ROUTES (Protected + Admin Role)
	// ==========================================
	adminGeofence := engine.Group("/api/v1/admin/geofences")
	adminGeofence.Use(
		authContainer.Middleware.Authenticate(),
		authContainer.AdminMiddleware.RequireAdmin(),
	)
	{
		adminGeofence.POST("", geofenceContainer.Handler.CreateGeofence)
		adminGeofence.GET("", geofenceContainer.Handler.ListGeofences)
		adminGeofence.GET("/check", geofenceContainer.Handler.CheckLocation)
	}

	adminZones := engine.Group("/api/v1/admin/zones")
	adminZones.Use(
		authContainer.Middleware.Authenticate(),
		authContainer.AdminMiddleware.RequireAdmin(),
	)
	{
		adminZones.POST("/assign", geofenceContainer.Handler.AssignDriverToZone)
		adminZones.DELETE("/drivers/:driverId/zones/:zoneId", geofenceContainer.Handler.RemoveDriverFromZone)
	}

	// ==========================================
	// ADMIN SETTINGS & VEHICLES ROUTES
	// ==========================================
	adminSettings := engine.Group("/api/v1/admin/settings")
	adminSettings.Use(
		authContainer.Middleware.Authenticate(),
		authContainer.AdminMiddleware.RequireAdmin(),
	)
	{
		adminSettings.GET("", settingsContainer.Handler.GetPlatformSettings)
		adminSettings.PUT("", settingsContainer.Handler.UpdatePlatformSettings)
	}

	adminVehicles := engine.Group("/api/v1/admin/vehicles")
	adminVehicles.Use(
		authContainer.Middleware.Authenticate(),
		authContainer.AdminMiddleware.RequireAdmin(),
	)
	{
		adminVehicles.POST("", settingsContainer.Handler.CreateVehicleType)
		adminVehicles.GET("", settingsContainer.Handler.ListVehicleTypes)
		adminVehicles.PUT("/:id", settingsContainer.Handler.UpdateVehicleType)
		adminVehicles.DELETE("/:id", settingsContainer.Handler.DeleteVehicleType)
	}

	// ==========================================
	// OPENAPI DOCUMENTATION
	// ==========================================
	// ==========================================
	// SWAGGER DOCUMENTATION UI
	// ==========================================
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return engine
}
