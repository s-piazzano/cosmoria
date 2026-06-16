package testhelper

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/core"
)

func UserJWT(t testing.TB, cfg *core.Config, userID, projectID string) string {
	t.Helper()

	token, err := auth.GenerateToken(auth.Claims{
		UserID:    userID,
		ProjectID: projectID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "cosmoria-test",
		},
	}, cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("testhelper: generate user JWT: %v", err)
	}
	return token
}

func AdminJWT(t testing.TB, cfg *core.Config, adminID string) string {
	t.Helper()

	token, err := adminauth.GenerateToken(adminauth.AdminClaims{
		AdminUserID: adminID,
		Role:    "super_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "cosmoria-test",
		},
	}, cfg.AdminJWTSecret, cfg.AdminJWTExpiry)
	if err != nil {
		t.Fatalf("testhelper: generate admin JWT: %v", err)
	}
	return token
}
