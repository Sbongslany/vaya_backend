package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	pgtest "github.com/testcontainers/testcontainers-go/modules/postgres"
	redistest "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yourorg/ehailing/backend/internal/auth/dependency"
	"github.com/yourorg/ehailing/backend/internal/config"
	"github.com/yourorg/ehailing/backend/internal/database"
	"github.com/yourorg/ehailing/backend/internal/httpapi"
	"github.com/yourorg/ehailing/backend/migrations"
)

type TestSuite struct {
	PgPool      *pgxpool.Pool
	RedisClient *redis.Client
	Engine      *gin.Engine
	Cfg         *config.Config
}

func SetupSuite(t *testing.T) *TestSuite {
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	// 1. Start Postgres Container
	pgContainer, err := pgtest.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		pgtest.WithDatabase("testdb"),
		pgtest.WithUsername("testuser"),
		pgtest.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	// 2. Start Redis Container
	redisContainer, err := redistest.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	t.Cleanup(func() { redisContainer.Terminate(ctx) })

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")

	// 3. Run Migrations
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", pgHost, pgPort.Port())
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open sql db for migrations: %v", err)
	}
	defer sqlDB.Close()

	if err := migrations.Up(sqlDB); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// 4. Build Config
	cfg := &config.Config{
		Env:      "test",
		HTTPPort: "0",
		LogLevel: "error",
		Postgres: config.PostgresConfig{
			Host: pgHost, Port: pgPort.Port(), User: "testuser", Password: "testpass", DB: "testdb", SSLMode: "disable",
			MaxConns: 5, MinConns: 1, MaxConnLifetime: 5 * time.Minute, MaxConnIdleTime: 1 * time.Minute, HealthCheckPeriod: 10 * time.Second,
		},
		Redis: config.RedisConfig{
			Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()), DB: 0,
			PoolSize: 5, MinIdleConns: 1, DialTimeout: 5 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second,
		},
		JWTAccessSecret:   "test-access-secret-min-32-chars-long",
		JWTRefreshSecret:  "test-refresh-secret-min-32-chars-long",
		JWTIssuer:         "test-issuer",
		JWTAudience:       "test-audience",
		JWTAccessTTL:      15 * time.Minute,
		JWTRefreshTTL:     24 * time.Hour,
		MFAEncryptionKey:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		OTPLength:         6,
		OTPTTL:            5 * time.Minute,
		OTPMaxAttempts:    5,
		OTPResendCooldown: 60 * time.Second,
	}

	// 5. Connect App DB & Redis
	pgPool, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		t.Fatalf("failed to connect to test postgres: %v", err)
	}
	t.Cleanup(func() { pgPool.Close() })

	redisClient, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		t.Fatalf("failed to connect to test redis: %v", err)
	}
	t.Cleanup(func() { redisClient.Close() })

	// 6. Wire Dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	authContainer := dependency.WireAuth(pgPool, redisClient, cfg, logger)
	engine := httpapi.NewRouter(logger, pgPool, redisClient, cfg, authContainer)

	return &TestSuite{
		PgPool:      pgPool,
		RedisClient: redisClient,
		Engine:      engine,
		Cfg:         cfg,
	}
}
