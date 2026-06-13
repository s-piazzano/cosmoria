package middleware

import (
	"encoding/json"
	"net/http"
)

func Chain(handler http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range mws {
		handler = mw(handler)
	}
	return handler
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	b, _ := json.Marshal(v)
	w.Write(b)
}
