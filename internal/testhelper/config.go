package testhelper

import (
	"github.com/s-piazzano/cosmoria/internal/core"
)

func TestConfig() *core.Config {
	return &core.Config{
		JWTSecret:      "test-jwt-secret-32-bytes-long-for-hs256!!",
		JWTExpiry:      86400,
		AdminJWTSecret: "test-admin-jwt-secret-32-bytes-lon",
		AdminJWTExpiry: 3600,
		StoragePath:    "/tmp/cosmoria-test-files",
		DatabaseURL:    TestDBURL(),
	}
}
