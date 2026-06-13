package core

import (
	"fmt"
	"log"
	"net/http"
)

type App struct {
	config  *Config
	handler http.Handler
}

func NewApp(cfg *Config, handler http.Handler) *App {
	return &App{config: cfg, handler: handler}
}

func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.config.Port)
	log.Printf("cosmoria starting on %s", addr)
	return http.ListenAndServe(addr, a.handler)
}
