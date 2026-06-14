package core

import (
	"os"
	"strconv"
)

const defaultPort = 8080

type Config struct {
	Port            int
	DatabaseURL     string
	AutoMigrate     bool
	JWTSecret       string
	JWTExpiry       int64
	AdminJWTSecret  string
	AdminJWTExpiry  int64
}

func LoadConfig() *Config {
	port := defaultPort
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
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
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		AutoMigrate:    os.Getenv("AUTO_MIGRATE") != "false",
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiry:      jwtExpiry,
		AdminJWTSecret: os.Getenv("ADMIN_JWT_SECRET"),
		AdminJWTExpiry: adminJwtExpiry,
	}
}
