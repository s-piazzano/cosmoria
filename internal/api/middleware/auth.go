package middleware

import (
	"net/http"
	"strings"

	"github.com/s-piazzano/cosmoria/internal/auth"
)

var adminPrefix = "/api/admin/"

var publicRoutes = []string{
	"/health",
	"/api/auth/signup",
	"/api/auth/login",
	"/openapi.json",
}

type AuthMiddleware struct {
	JWTSecret  string
	ApiKeySvc  *auth.ApiKeyService
}

func Auth(secret string, apiKeySvc *auth.ApiKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range publicRoutes {
				if r.URL.Path == p {
					next.ServeHTTP(w, r)
					return
				}
			}

			if strings.HasPrefix(r.URL.Path, adminPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasSuffix(r.URL.Path, "/ws") {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/docs") {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := auth.ValidateToken(tokenStr, secret)
				if err != nil {
					writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
					return
				}
				ctx := auth.WithAuth(r.Context(), claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			apiKey := r.Header.Get("X-Api-Key")
			if apiKey != "" && apiKeySvc != nil {
				claims, err := apiKeySvc.Validate(r.Context(), apiKey)
				if err != nil {
					writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_api_key"})
					return
				}
				ctx := auth.WithAuth(r.Context(), claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		})
	}
}
