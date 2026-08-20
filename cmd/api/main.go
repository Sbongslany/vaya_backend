package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/yourorg/ehailing/backend/internal/auth/dependency"
	"github.com/yourorg/ehailing/backend/internal/config"
	"github.com/yourorg/ehailing/backend/internal/database"
	httpapi "github.com/yourorg/ehailing/backend/internal/httpapi"
	"github.com/yourorg/ehailing/backend/internal/logger"
	promoDep "github.com/yourorg/ehailing/backend/internal/promotions/dependency"
	tripDep "github.com/yourorg/ehailing/backend/internal/trip/dependency"
	"github.com/yourorg/ehailing/backend/migrations"

)

func main() {
	// Load environment variables from .env.
	// This must happen before config.Load().
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: .env file not found: %v\n", err)
	}

	// Load and validate application configuration.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger.
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
		os.Exit(1)
	}

	log = log.With("service", "ehailing-auth-api")

	// Run database migrations before starting the application.
	if err := runMigrations(cfg.Postgres.DSN()); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create application context with graceful shutdown support.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Connect to PostgreSQL.
	pgPool, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	// Connect to Redis.
	redisClient, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	// Wire authentication dependencies.
	authContainer := dependency.WireAuth(pgPool, redisClient, cfg, log)

	// Wire promotions dependencies
	promosContainer := promoDep.WirePromotions(pgPool)

	// Wire trip dependencies (now passes promosContainer)
	tripContainer := tripDep.WireTrip(pgPool, promosContainer)


	// Create HTTP router.
	engine := httpapi.NewRouter(
		log,
		pgPool,
		redisClient,
		cfg,
		authContainer,
		tripContainer, // <-- ADD THIS LINE
		promosContainer, // <-- ADD THIS
	)

	// Configure HTTP server.
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Channel used to capture fatal HTTP server errors.
	serverErrCh := make(chan error, 1)

	// Start HTTP server.
	go func() {
		log.Info(
			"starting api server",
			"port", cfg.HTTPPort,
			"env", cfg.Env,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	// Wait for either a shutdown signal or a fatal server error.
	select {
	case err := <-serverErrCh:
		log.Error("api server failed", "error", err)
		os.Exit(1)

	case <-ctx.Done():
		stop()
		log.Info("shutdown signal received")
	}

	// Gracefully shut down the HTTP server.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		_ = server.Close()
	}

	log.Info("api server stopped")
}

// runMigrations opens a temporary PostgreSQL connection,
// verifies the connection, and executes pending migrations.
func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf(
			"open migration database connection: %w",
			err,
		)
	}

	defer func() {
		_ = db.Close()
	}()

	// Migrations should run using a single connection.
	db.SetMaxOpenConns(1)

	// Verify that PostgreSQL is reachable.
	if err := db.Ping(); err != nil {
		return fmt.Errorf(
			"ping migration database: %w",
			err,
		)
	}

	// Run pending migrations.
	return migrations.Up(db)
}
