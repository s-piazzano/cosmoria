package main

import (
	"log"

	"github.com/s-piazzano/cosmoria/internal/api"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/db"
)

func main() {
	cfg := core.LoadConfig()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	router := api.NewRouter()
	app := core.NewApp(cfg, pool, router)

	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
