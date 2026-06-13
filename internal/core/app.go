package core

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	config  *Config
	pool    *pgxpool.Pool
	handler http.Handler
}

func NewApp(cfg *Config, pool *pgxpool.Pool, handler http.Handler) *App {
	return &App{config: cfg, pool: pool, handler: handler}
}

func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.config.Port)
	log.Printf("cosmoria starting on %s", addr)
	return http.ListenAndServe(addr, a.handler)
}

func (a *App) Shutdown() {
	if a.pool != nil {
		a.pool.Close()
	}
}
