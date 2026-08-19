package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env                string
	HTTPPort           string
	LogLevel           string
	CORSAllowedOrigins []string

	Postgres PostgresConfig
	Redis    RedisConfig

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTIssuer        string
	JWTAudience      string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	OTPLength         int
	OTPTTL            time.Duration
	OTPMaxAttempts    int
	OTPResendCooldown time.Duration

	MFAEncryptionKey string

	// Cloudinary
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

type PostgresConfig struct {
	Host              string
	Port              string
	User              string
	Password          string
	DB                string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func (c PostgresConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   "/" + c.DB,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() (*Config, error) {
	env := getEnv("APP_ENV", "development")
	httpPort := getEnv("HTTP_PORT", "8080")
	logLevel := getEnv("LOG_LEVEL", "info")

	// -------------------------------------------------------------------------
	// HTTP
	// -------------------------------------------------------------------------

	if err := validatePort(httpPort); err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	corsAllowedOrigins := parseCORS(
		getEnv("CORS_ALLOWED_ORIGINS", ""),
	)

	// -------------------------------------------------------------------------
	// PostgreSQL
	// -------------------------------------------------------------------------

	pgHost, err := requiredEnv("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}

	pgPort, err := requiredEnv("POSTGRES_PORT")
	if err != nil {
		return nil, err
	}

	if err := validatePort(pgPort); err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
	}

	pgUser, err := requiredEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	pgPassword, err := requiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	pgDB, err := requiredEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	pgSSLMode := getEnv("POSTGRES_SSLMODE", "disable")

	pgMaxConns, err := getInt32("POSTGRES_MAX_CONNS", 20)
	if err != nil {
		return nil, err
	}

	pgMinConns, err := getInt32("POSTGRES_MIN_CONNS", 2)
	if err != nil {
		return nil, err
	}

	pgMaxConnLifetime, err := getDuration(
		"POSTGRES_MAX_CONN_LIFETIME",
		30*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	pgMaxConnIdleTime, err := getDuration(
		"POSTGRES_MAX_CONN_IDLE_TIME",
		5*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	pgHealthCheckPeriod, err := getDuration(
		"POSTGRES_HEALTH_CHECK_PERIOD",
		30*time.Second,
	)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// Redis
	// -------------------------------------------------------------------------

	redisHost, err := requiredEnv("REDIS_HOST")
	if err != nil {
		return nil, err
	}

	redisPort, err := requiredEnv("REDIS_PORT")
	if err != nil {
		return nil, err
	}

	if err := validatePort(redisPort); err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	redisDB, err := getInt("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}

	redisPoolSize, err := getInt("REDIS_POOL_SIZE", 20)
	if err != nil {
		return nil, err
	}

	redisMinIdleConns, err := getInt(
		"REDIS_MIN_IDLE_CONNS",
		5,
	)
	if err != nil {
		return nil, err
	}

	redisDialTimeout, err := getDuration(
		"REDIS_DIAL_TIMEOUT",
		5*time.Second,
	)
	if err != nil {
		return nil, err
	}

	redisReadTimeout, err := getDuration(
		"REDIS_READ_TIMEOUT",
		3*time.Second,
	)
	if err != nil {
		return nil, err
	}

	redisWriteTimeout, err := getDuration(
		"REDIS_WRITE_TIMEOUT",
		3*time.Second,
	)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// JWT Authentication
	// -------------------------------------------------------------------------

	jwtAccessSecret, err := requiredEnv("JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}

	jwtRefreshSecret, err := requiredEnv("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, err
	}

	jwtIssuer := getEnv(
		"JWT_ISSUER",
		"ehailing-api",
	)

	jwtAudience := getEnv(
		"JWT_AUDIENCE",
		"ehailing-apps",
	)

	jwtAccessTTL, err := getDuration(
		"JWT_ACCESS_TTL",
		15*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	jwtRefreshTTL, err := getDuration(
		"JWT_REFRESH_TTL",
		720*time.Hour,
	)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// OTP
	// -------------------------------------------------------------------------

	otpLength, err := getInt(
		"OTP_LENGTH",
		6,
	)
	if err != nil {
		return nil, err
	}

	otpTTL, err := getDuration(
		"OTP_TTL",
		5*time.Minute,
	)
	if err != nil {
		return nil, err
	}

	otpMaxAttempts, err := getInt(
		"OTP_MAX_ATTEMPTS",
		5,
	)
	if err != nil {
		return nil, err
	}

	otpResendCooldown, err := getDuration(
		"OTP_RESEND_COOLDOWN",
		60*time.Second,
	)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// MFA Encryption
	// -------------------------------------------------------------------------

	mfaEncryptionKey, err := requiredEnv("MFA_ENCRYPTION_KEY")
	if err != nil {
		return nil, err
	}

	if len(mfaEncryptionKey) != 64 {
		return nil, fmt.Errorf(
			"MFA_ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes)",
		)
	}

	// -------------------------------------------------------------------------
	// Cloudinary
	// -------------------------------------------------------------------------

	cloudinaryCloudName, err := requiredEnv("CLOUDINARY_CLOUD_NAME")
	if err != nil {
		return nil, err
	}

	cloudinaryAPIKey, err := requiredEnv("CLOUDINARY_API_KEY")
	if err != nil {
		return nil, err
	}

	cloudinaryAPISecret, err := requiredEnv("CLOUDINARY_API_SECRET")
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// Build configuration
	// -------------------------------------------------------------------------

	cfg := &Config{
		Env:                env,
		HTTPPort:           httpPort,
		LogLevel:           logLevel,
		CORSAllowedOrigins: corsAllowedOrigins,

		Postgres: PostgresConfig{
			Host:              pgHost,
			Port:              pgPort,
			User:              pgUser,
			Password:          pgPassword,
			DB:                pgDB,
			SSLMode:           pgSSLMode,
			MaxConns:          pgMaxConns,
			MinConns:          pgMinConns,
			MaxConnLifetime:   pgMaxConnLifetime,
			MaxConnIdleTime:   pgMaxConnIdleTime,
			HealthCheckPeriod: pgHealthCheckPeriod,
		},

		Redis: RedisConfig{
			Addr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password:     os.Getenv("REDIS_PASSWORD"),
			DB:           redisDB,
			PoolSize:     redisPoolSize,
			MinIdleConns: redisMinIdleConns,
			DialTimeout:  redisDialTimeout,
			ReadTimeout:  redisReadTimeout,
			WriteTimeout: redisWriteTimeout,
		},

		JWTAccessSecret:  jwtAccessSecret,
		JWTRefreshSecret: jwtRefreshSecret,
		JWTIssuer:        jwtIssuer,
		JWTAudience:      jwtAudience,
		JWTAccessTTL:     jwtAccessTTL,
		JWTRefreshTTL:    jwtRefreshTTL,

		OTPLength:         otpLength,
		OTPTTL:            otpTTL,
		OTPMaxAttempts:    otpMaxAttempts,
		OTPResendCooldown: otpResendCooldown,

		MFAEncryptionKey: mfaEncryptionKey,

		CloudinaryCloudName: cloudinaryCloudName,
		CloudinaryAPIKey:    cloudinaryAPIKey,
		CloudinaryAPISecret: cloudinaryAPISecret,
	}

	return cfg, nil
}

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return value, nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return value, nil
}

func getInt32(key string, fallback int32) (int32, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return int32(value), nil
}

func getDuration(
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be a valid duration, example: 30s, 5m, 1h",
			key,
		)
	}

	return value, nil
}

func validatePort(port string) error {
	value, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be numeric")
	}

	if value <= 0 || value > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

func parseCORS(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
