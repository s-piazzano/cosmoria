package middleware_test

import (
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/auth"
)

func injectUserClaims(r *http.Request, userID, projectID string) *http.Request {
	return r.WithContext(auth.WithAuth(r.Context(), &auth.Claims{
		UserID:    userID,
		ProjectID: projectID,
	}))
}
