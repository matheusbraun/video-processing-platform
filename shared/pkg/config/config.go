package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	ServerPort string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// RabbitMQ
	RabbitMQURL string

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3UploadsBucket    string
	S3ProcessedBucket  string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// Service URLs
	AuthServiceURL    string
	APIGatewayURL     string
	StorageServiceURL string

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	// HTTP Client
	HTTPClientTimeout    time.Duration
	HTTPClientRetryCount int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	jwtAccessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	jwtRefreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	httpTimeout, err := time.ParseDuration(getEnv("HTTP_CLIENT_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_CLIENT_TIMEOUT: %w", err)
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	retryCount, err := strconv.Atoi(getEnv("HTTP_CLIENT_RETRY_COUNT", "3"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_CLIENT_RETRY_COUNT: %w", err)
	}

	return &Config{
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		RedisURL:             getEnv("REDIS_URL", ""),
		RabbitMQURL:          getEnv("RABBITMQ_URL", ""),
		AWSRegion:            getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:       getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:   getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3UploadsBucket:      getEnv("S3_UPLOADS_BUCKET", ""),
		S3ProcessedBucket:    getEnv("S3_PROCESSED_BUCKET", ""),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:      jwtAccessExpiry,
		JWTRefreshExpiry:     jwtRefreshExpiry,
		AuthServiceURL:       getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		APIGatewayURL:        getEnv("API_GATEWAY_URL", "http://api-gateway:8080"),
		StorageServiceURL:    getEnv("STORAGE_SERVICE_URL", "http://storage-service:8080"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             smtpPort,
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		HTTPClientTimeout:    httpTimeout,
		HTTPClientRetryCount: retryCount,
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
