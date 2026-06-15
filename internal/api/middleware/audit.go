package middleware

import (
	"net/http"
	"strings"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/audit"
	"github.com/s-piazzano/cosmoria/internal/auth"
)

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Audit(logger *audit.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			aw := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(aw, r)

			if aw.statusCode < 200 || aw.statusCode >= 300 {
				return
			}

			resource, action, resourceID := auditInfo(r.Pattern, r)
			if resource == "" {
				return
			}

			projectID := r.PathValue("pid")
			var userID string
			if claims := auth.GetAuth(r.Context()); claims != nil {
				userID = claims.UserID
			} else if adminClaims := adminauth.GetAdminAuth(r.Context()); adminClaims != nil {
				userID = adminClaims.AdminUserID
			}
			if userID == "" || projectID == "" {
				return
			}

			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = strings.Split(xff, ",")[0]
			}

			logger.Log(r.Context(), projectID, userID, action, resource, resourceID, nil, ip)
		})
	}
}

func auditInfo(pattern string, r *http.Request) (resource, action string, resourceID *string) {
	if pattern == "" {
		return "", "", nil
	}

	method := strings.SplitN(pattern, " ", 2)[0]
	path := strings.SplitN(pattern, " ", 2)[1]

	var rid string

	switch {
	case strings.Contains(path, "/tenants/{tid}/files/{fid}"):
		rid = r.PathValue("fid")
		switch method {
		case "GET":
			return "files", "read", &rid
		case "DELETE":
			return "files", "delete", &rid
		}
	case strings.Contains(path, "/tenants/{tid}/files"):
		switch method {
		case "POST":
			return "files", "create", nil
		case "GET":
			return "files", "read", nil
		}

	case strings.Contains(path, "/collections/{cid}/records/{rid}"):
		rid = r.PathValue("rid")
		switch method {
		case "PUT":
			return "records", "update", &rid
		case "DELETE":
			return "records", "delete", &rid
		case "GET":
			return "records", "read", &rid
		}
	case strings.Contains(path, "/collections/{cid}/records"):
		switch method {
		case "POST":
			return "records", "create", nil
		case "GET":
			return "records", "read", nil
		}

	case strings.HasPrefix(path, "/api/admin/projects/{pid}/collections/{cid}"):
		rid = r.PathValue("cid")
		switch method {
		case "PUT":
			return "collections", "update", &rid
		case "DELETE":
			return "collections", "delete", &rid
		case "GET":
			return "collections", "read", &rid
		}
	case strings.HasPrefix(path, "/api/admin/projects/{pid}/collections"):
		switch method {
		case "POST":
			return "collections", "create", nil
		case "GET":
			return "collections", "read", nil
		}

	case strings.HasPrefix(path, "/api/projects/{pid}/tenants/{tid}") &&
		strings.Contains(path, "/users/{uid}"):
		rid = r.PathValue("uid")
		switch method {
		case "POST":
			return "tenants", "update", &rid
		case "DELETE":
			return "tenants", "delete", &rid
		}
	case strings.HasPrefix(path, "/api/projects/{pid}/tenants/{tid}"):
		rid = r.PathValue("tid")
		switch method {
		case "DELETE":
			return "tenants", "delete", &rid
		}
	case strings.HasPrefix(path, "/api/projects/{pid}/tenants"):
		switch method {
		case "POST":
			return "tenants", "create", nil
		case "GET":
			return "tenants", "read", nil
		}
	}

	return "", "", nil
}
