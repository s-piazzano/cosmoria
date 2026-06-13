package core

import (
	"os"
	"strconv"
)

const defaultPort = 8080

type Config struct {
	Port        int
	DatabaseURL string
}

func LoadConfig() *Config {
	port := defaultPort
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
	}
	return &Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
