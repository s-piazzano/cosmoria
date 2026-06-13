package main

import (
	"log"

	"github.com/s-piazzano/cosmoria/internal/api"
	"github.com/s-piazzano/cosmoria/internal/api/handlers"
	"github.com/s-piazzano/cosmoria/internal/api/middleware"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/db"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

func main() {
	cfg := core.LoadConfig()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := db.Migrate(pool, "db/migrations"); err != nil {
			log.Fatalf("migrations: %v", err)
		}
	}

	authService := auth.NewService(pool, cfg)
	authHandler := &handlers.AuthHandler{Service: authService}

	tenantService := tenant.NewService(pool)
	tenantHandler := &handlers.TenantHandler{Service: tenantService}

	router := api.NewRouter()
	router.HandleFunc("POST /api/auth/signup", authHandler.Signup)
	router.HandleFunc("POST /api/auth/login", authHandler.Login)

	router.HandleFunc("POST /api/projects/{pid}/tenants", tenantHandler.Create)
	router.HandleFunc("GET /api/projects/{pid}/tenants", tenantHandler.List)
	router.HandleFunc("GET /api/projects/{pid}/tenants/{tid}", tenantHandler.Get)
	router.HandleFunc("DELETE /api/projects/{pid}/tenants/{tid}", tenantHandler.Delete)
	router.HandleFunc("POST /api/projects/{pid}/tenants/{tid}/users", tenantHandler.AssignUser)
	router.HandleFunc("DELETE /api/projects/{pid}/tenants/{tid}/users/{uid}", tenantHandler.RemoveUser)

	mw := middleware.Chain(router,
		middleware.Logging(),
		middleware.Auth(cfg.JWTSecret),
		middleware.Tenant(tenantService),
	)

	app := core.NewApp(cfg, pool, mw)

	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
