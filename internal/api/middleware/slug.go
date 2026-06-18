package middleware

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}

func ResolveProjectSlug(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if !strings.HasPrefix(path, "/api/admin/projects/") {
				next.ServeHTTP(w, r)
				return
			}

			prefix := "/api/admin/projects/"
			rest := path[len(prefix):]
			pid := strings.SplitN(rest, "/", 2)[0]

			if pid == "" || isUUID(pid) {
				next.ServeHTTP(w, r)
				return
			}

			var projectID string
			err := pool.QueryRow(r.Context(), `SELECT id FROM projects WHERE slug = $1`, pid).Scan(&projectID)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "project_not_found"})
				return
			}

			r.URL.Path = prefix + projectID + rest[len(pid):]

			next.ServeHTTP(w, r)
		})
	}
}
