package core

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"strconv"
)

const defaultPort = 8080

const defaultDatabaseURL = "postgres://localhost:5432/cosmoria?sslmode=disable"

type Config struct {
	Port           int
	DatabaseURL    string
	AutoMigrate    bool
	JWTSecret      string
	JWTExpiry      int64
	AdminJWTSecret string
	AdminJWTExpiry int64
	S3Endpoint     string
	S3AccessKey    string
	S3SecretKey    string
	S3Bucket       string
	S3Region       string
	S3UseSSL       bool
}

func generateSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func LoadConfig() *Config {
	port := defaultPort
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
		slog.Warn("DATABASE_URL not set, using default", "url", defaultDatabaseURL)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = generateSecret()
		slog.Warn("JWT_SECRET not set, using generated key — tokens invalid after restart")
	}

	adminJwtSecret := os.Getenv("ADMIN_JWT_SECRET")
	if adminJwtSecret == "" {
		adminJwtSecret = generateSecret()
		slog.Warn("ADMIN_JWT_SECRET not set, using generated key — tokens invalid after restart")
	}

	jwtExpiry := int64(86400)
	if v := os.Getenv("JWT_EXPIRY"); v != "" {
		if e, err := strconv.ParseInt(v, 10, 64); err == nil && e > 0 {
			jwtExpiry = e
		}
	}

	adminJwtExpiry := int64(3600)
	if v := os.Getenv("ADMIN_JWT_EXPIRY"); v != "" {
		if e, err := strconv.ParseInt(v, 10, 64); err == nil && e > 0 {
			adminJwtExpiry = e
		}
	}

	return &Config{
		Port:           port,
		DatabaseURL:    databaseURL,
		AutoMigrate:    os.Getenv("AUTO_MIGRATE") != "false",
		JWTSecret:      jwtSecret,
		JWTExpiry:      jwtExpiry,
		AdminJWTSecret: adminJwtSecret,
		AdminJWTExpiry: adminJwtExpiry,
		S3Endpoint:     getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:    os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:    os.Getenv("S3_SECRET_KEY"),
		S3Bucket:       getEnv("S3_BUCKET", "cosmoria"),
		S3Region:       getEnv("S3_REGION", "us-east-1"),
		S3UseSSL:       os.Getenv("S3_USE_SSL") == "true",
	}
}
