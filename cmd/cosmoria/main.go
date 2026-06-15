// @title Cosmoria API
// @version 0.1.0
// @description Backend engine for building multi-tenant SaaS applications on PostgreSQL.
// @termsOfService https://cosmoria.dev/terms

// @contact.name Cosmoria Support
// @contact.url https://cosmoria.dev/support
// @contact.email support@cosmoria.dev

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

// @securityDefinitions.apikey AdminBearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the admin JWT token.

package main

func main() {
	Run()
}
