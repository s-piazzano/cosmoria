package middleware

import (
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/rbac"
)

func RequirePermission(svc *rbac.Service, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := auth.GetAuth(r.Context())
			if claims == nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ok, err := svc.CheckAccess(r.Context(), claims.UserID, claims.ProjectID, resource, action)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "access_check_failed"})
				return
			}
			if !ok {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
