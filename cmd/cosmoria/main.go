package main

import (
	"log"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/api"
	"github.com/s-piazzano/cosmoria/internal/api/handlers"
	"github.com/s-piazzano/cosmoria/internal/api/middleware"
	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/db"
	"github.com/s-piazzano/cosmoria/internal/rbac"
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
	if cfg.AdminJWTSecret == "" {
		log.Fatal("ADMIN_JWT_SECRET is required")
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

	adminService := adminauth.NewService(pool, cfg)
	adminHandler := &handlers.AdminHandler{Service: adminService}

	rbacService := rbac.NewService(pool)
	rolesHandler := &handlers.RolesHandler{Service: rbacService}

	router := api.NewRouter()
	router.HandleFunc("POST /api/auth/signup", authHandler.Signup)
	router.HandleFunc("POST /api/auth/login", authHandler.Login)

	router.Handle("POST /api/projects/{pid}/tenants",
		middleware.RequirePermission(rbacService, "tenants", "create")(http.HandlerFunc(tenantHandler.Create)))
	router.Handle("GET /api/projects/{pid}/tenants",
		middleware.RequirePermission(rbacService, "tenants", "read")(http.HandlerFunc(tenantHandler.List)))
	router.Handle("GET /api/projects/{pid}/tenants/{tid}",
		middleware.RequirePermission(rbacService, "tenants", "read")(http.HandlerFunc(tenantHandler.Get)))
	router.Handle("DELETE /api/projects/{pid}/tenants/{tid}",
		middleware.RequirePermission(rbacService, "tenants", "delete")(http.HandlerFunc(tenantHandler.Delete)))
	router.Handle("POST /api/projects/{pid}/tenants/{tid}/users",
		middleware.RequirePermission(rbacService, "tenants", "update")(http.HandlerFunc(tenantHandler.AssignUser)))
	router.Handle("DELETE /api/projects/{pid}/tenants/{tid}/users/{uid}",
		middleware.RequirePermission(rbacService, "tenants", "delete")(http.HandlerFunc(tenantHandler.RemoveUser)))

	router.HandleFunc("POST /api/admin/setup", adminHandler.Setup)
	router.HandleFunc("POST /api/admin/login", adminHandler.Login)
	router.HandleFunc("POST /api/admin/projects", adminHandler.CreateProject)
	router.HandleFunc("GET /api/admin/projects", adminHandler.ListProjects)

	router.HandleFunc("POST /api/admin/projects/{pid}/admin-roles", adminHandler.AssignRole)
	router.HandleFunc("GET /api/admin/projects/{pid}/admin-roles", adminHandler.ListRoles)
	router.HandleFunc("DELETE /api/admin/projects/{pid}/admin-roles/{aid}", adminHandler.RemoveRole)

	router.HandleFunc("POST /api/admin/projects/{pid}/roles", rolesHandler.CreateRole)
	router.HandleFunc("GET /api/admin/projects/{pid}/roles", rolesHandler.ListRoles)
	router.HandleFunc("DELETE /api/admin/projects/{pid}/roles/{rid}", rolesHandler.DeleteRole)
	router.HandleFunc("POST /api/admin/projects/{pid}/roles/{rid}/permissions", rolesHandler.SetPermission)
	router.HandleFunc("DELETE /api/admin/projects/{pid}/roles/{rid}/permissions", rolesHandler.RemovePermission)
	router.HandleFunc("GET /api/admin/projects/{pid}/roles/{rid}/permissions", rolesHandler.ListPermissions)
	router.HandleFunc("POST /api/admin/projects/{pid}/users/{uid}/role", rolesHandler.AssignUserRole)
	router.HandleFunc("GET /api/admin/projects/{pid}/users/{uid}/role", rolesHandler.GetUserRole)
	router.HandleFunc("DELETE /api/admin/projects/{pid}/users/{uid}/role", rolesHandler.RemoveUserRole)

	mw := middleware.Chain(router,
		middleware.Logging(),
		middleware.Auth(cfg.JWTSecret),
		middleware.AdminAuth(cfg.AdminJWTSecret),
		middleware.Tenant(tenantService),
	)

	app := core.NewApp(cfg, pool, mw)

	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
