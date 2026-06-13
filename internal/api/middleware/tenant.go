package middleware

import (
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

func Tenant(svc *tenant.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range publicRoutes {
				if r.URL.Path == p {
					next.ServeHTTP(w, r)
					return
				}
			}

			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims := auth.GetAuth(r.Context())
			if claims == nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ok, err := svc.HasAccess(r.Context(), claims.UserID, tenantID, claims.ProjectID)
			if err != nil || !ok {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "access_denied"})
				return
			}

			ctx := tenant.WithTenant(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
