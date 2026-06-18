package core

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const defaultPort = 8080

const defaultDatabaseURL = "postgres://localhost:5432/cosmoria?sslmode=disable"

type Config struct {
	Port             int
	DatabaseURL      string
	AutoMigrate      bool
	JWTSecret        string
	JWTExpiry        int64
	AdminJWTSecret   string
	AdminJWTExpiry   int64
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3Bucket         string
	S3Region         string
	S3UseSSL         bool
	StoragePath      string
	WSAllowedOrigins []string
	MaxUploadSize    int64
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

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func LoadConfig() *Config {
	loadEnvFile(".env")

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

	var wsAllowedOrigins []string
	if v := os.Getenv("WS_ALLOWED_ORIGINS"); v != "" {
		wsAllowedOrigins = strings.Split(v, ",")
		for i := range wsAllowedOrigins {
			wsAllowedOrigins[i] = strings.TrimSpace(wsAllowedOrigins[i])
		}
	}

	maxUploadSize := int64(100 * 1024 * 1024) // 100MB default
	if v := os.Getenv("MAX_UPLOAD_SIZE"); v != "" {
		if s, err := strconv.ParseInt(v, 10, 64); err == nil && s > 0 {
			maxUploadSize = s
		}
	}

	if os.Getenv("ENV") == "production" {
		if jwtSecret == os.Getenv("JWT_SECRET") && jwtSecret == "" {
			slog.Error("JWT_SECRET must be set in production")
			os.Exit(1)
		}
		if adminJwtSecret == os.Getenv("ADMIN_JWT_SECRET") && adminJwtSecret == "" {
			slog.Error("ADMIN_JWT_SECRET must be set in production")
			os.Exit(1)
		}
	}

	return &Config{
		Port:             port,
		DatabaseURL:      databaseURL,
		AutoMigrate:      os.Getenv("AUTO_MIGRATE") != "false",
		JWTSecret:        jwtSecret,
		JWTExpiry:        jwtExpiry,
		AdminJWTSecret:   adminJwtSecret,
		AdminJWTExpiry:   adminJwtExpiry,
		S3Endpoint:       getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:      os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:      os.Getenv("S3_SECRET_KEY"),
		S3Bucket:         getEnv("S3_BUCKET", "cosmoria"),
		S3Region:         getEnv("S3_REGION", "us-east-1"),
		S3UseSSL:         os.Getenv("S3_USE_SSL") == "true",
		StoragePath:      getEnv("STORAGE_PATH", "./data/files"),
		WSAllowedOrigins: wsAllowedOrigins,
		MaxUploadSize:    maxUploadSize,
	}
}
