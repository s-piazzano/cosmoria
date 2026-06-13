package main

import (
	"log"

	"github.com/s-piazzano/cosmoria/internal/api"
	"github.com/s-piazzano/cosmoria/internal/core"
)

func main() {
	cfg := core.LoadConfig()
	router := api.NewRouter()
	app := core.NewApp(cfg, router)

	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
