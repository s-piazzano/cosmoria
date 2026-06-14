package middleware

import (
	"net/http"
	"strings"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
)

var adminPublicRoutes = []string{
	"/api/admin/setup",
	"/api/admin/login",
}

func AdminAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range adminPublicRoutes {
				if r.URL.Path == p {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := adminauth.ValidateToken(tokenStr, secret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
				return
			}

			ctx := adminauth.WithAdminAuth(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
