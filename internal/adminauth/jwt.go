package adminauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	AdminUserID string `json:"admin_user_id"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(claims AdminClaims, secret string, expirySeconds int64) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(expirySeconds) * time.Second))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("adminauth: sign token: %w", err)
	}
	return signed, nil
}

func ValidateToken(tokenStr, secret string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("adminauth: validate token: %w", err)
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("adminauth: invalid token")
	}

	return claims, nil
}
