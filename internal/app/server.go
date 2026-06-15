package app

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/api"
	"github.com/s-piazzano/cosmoria/internal/api/handlers"
	"github.com/s-piazzano/cosmoria/internal/api/middleware"
	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/collections"
	"github.com/s-piazzano/cosmoria/internal/core"
	_ "github.com/s-piazzano/cosmoria/docs"
	"github.com/s-piazzano/cosmoria/internal/rbac"
	"github.com/s-piazzano/cosmoria/internal/records"
	"github.com/s-piazzano/cosmoria/internal/tenant"
	httpSwagger "github.com/swaggo/http-swagger"
)

func BuildHandler(cfg *core.Config, pool *pgxpool.Pool) http.Handler {
	authService := auth.NewService(pool, cfg)
	authHandler := &handlers.AuthHandler{Service: authService}

	tenantService := tenant.NewService(pool)
	tenantHandler := &handlers.TenantHandler{Service: tenantService}

	adminService := adminauth.NewService(pool, cfg)
	adminHandler := &handlers.AdminHandler{Service: adminService}

	rbacService := rbac.NewService(pool)
	rolesHandler := &handlers.RolesHandler{Service: rbacService}

	collectionsService := collections.NewService(pool)
	collectionsHandler := &handlers.CollectionsHandler{Service: collectionsService}

	recordsService := records.NewService(pool, collectionsService)
	recordsHandler := &handlers.RecordsHandler{Service: recordsService}

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

	router.Handle("POST /api/projects/{pid}/tenants/{tid}/collections/{cid}/records",
		middleware.RequirePermission(rbacService, "records", "create")(http.HandlerFunc(recordsHandler.Create)))
	router.Handle("GET /api/projects/{pid}/tenants/{tid}/collections/{cid}/records",
		middleware.RequirePermission(rbacService, "records", "read")(http.HandlerFunc(recordsHandler.List)))
	router.Handle("GET /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}",
		middleware.RequirePermission(rbacService, "records", "read")(http.HandlerFunc(recordsHandler.Get)))
	router.Handle("PUT /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}",
		middleware.RequirePermission(rbacService, "records", "update")(http.HandlerFunc(recordsHandler.Update)))
	router.Handle("DELETE /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}",
		middleware.RequirePermission(rbacService, "records", "delete")(http.HandlerFunc(recordsHandler.Delete)))

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

	router.HandleFunc("POST /api/admin/projects/{pid}/collections", collectionsHandler.Create)
	router.HandleFunc("GET /api/admin/projects/{pid}/collections", collectionsHandler.List)
	router.HandleFunc("GET /api/admin/projects/{pid}/collections/{cid}", collectionsHandler.Get)
	router.HandleFunc("PUT /api/admin/projects/{pid}/collections/{cid}", collectionsHandler.UpdateSchema)
	router.HandleFunc("DELETE /api/admin/projects/{pid}/collections/{cid}", collectionsHandler.Delete)

	router.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))
	router.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/doc.json", http.StatusMovedPermanently)
	})

	mw := middleware.Chain(router,
		middleware.Logging(),
		middleware.Auth(cfg.JWTSecret),
		middleware.AdminAuth(cfg.AdminJWTSecret),
		middleware.Tenant(tenantService),
	)

	return mw
}

func Serve(cfg *core.Config, pool *pgxpool.Pool) error {
	handler := BuildHandler(cfg, pool)
	app := core.NewApp(cfg, pool, handler)
	return app.Run()
}
