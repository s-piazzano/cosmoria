package api

import (
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/api/handlers"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	mux := http.NewServeMux()
	r := &Router{mux: mux}
	r.registerHealth()
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerHealth() {
	r.mux.HandleFunc("GET /health", handlers.Health)
}
