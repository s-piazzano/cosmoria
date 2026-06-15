package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/s-piazzano/cosmoria/internal/app"
	"github.com/s-piazzano/cosmoria/internal/configfile"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/db"
	"github.com/s-piazzano/cosmoria/internal/mcp"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Cosmoria — backend engine for multi-tenant SaaS applications

Usage:
  cosmoria serve              Start the server (default command)
  cosmoria dev                Start with hot reload (watch .go files)
  cosmoria init               Generate .env, docker-compose.yml, Dockerfile
  cosmoria migrate new <name> Create a new migration pair
  cosmoria migrate up         Run pending migrations
  cosmoria migrate down       Revert last migration
  cosmoria mcp                Start MCP server (stdin/stdout JSON-RPC)
`)
}

func Run() {
	if len(os.Args) < 2 {
		runServe()
		return
	}

	switch os.Args[1] {
	case "serve":
		runServe()
	case "dev":
		runDev()
	case "init":
		runInit()
	case "migrate":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "new":
			if len(os.Args) < 4 {
				fmt.Fprintln(os.Stderr, "usage: cosmoria migrate new <name>")
				os.Exit(1)
			}
			runMigrateNew(os.Args[3])
		case "up":
			runMigrateUp()
		case "down":
			runMigrateDown()
		default:
			usage()
			os.Exit(1)
		}
	case "mcp":
		runMCP()
	default:
		usage()
		os.Exit(1)
	}
}

func runServe() {
	cfg := core.LoadConfig()
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

	if err := configfile.ApplyIfPresent(pool, cfg); err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := app.Serve(cfg, pool); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func runInit() {
	// .env
	envContent := `# Cosmoria Configuration
# Copy to .env and adjust as needed.

DATABASE_URL=postgres://localhost:5432/cosmoria?sslmode=disable
PORT=8080
AUTO_MIGRATE=true

# Optional: set these for production to persist tokens across restarts
# JWT_SECRET=your-256-bit-secret
# ADMIN_JWT_SECRET=your-256-bit-admin-secret
`
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		log.Fatalf("write .env: %v", err)
	}
	log.Println("created .env")

	// docker-compose.yml
	composeContent := `version: "3.8"

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: cosmoria
      POSTGRES_USER: cosmoria
      POSTGRES_PASSWORD: cosmoria
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  cosmoria:
    build: .
    ports:
      - "${PORT:-8080}:${PORT:-8080}"
    environment:
      DATABASE_URL: postgres://cosmoria:cosmoria@postgres:5432/cosmoria?sslmode=disable
      PORT: "${PORT:-8080}"
    depends_on:
      - postgres

volumes:
  pgdata:
`
	if err := os.WriteFile("docker-compose.yml", []byte(composeContent), 0644); err != nil {
		log.Fatalf("write docker-compose.yml: %v", err)
	}
	log.Println("created docker-compose.yml")

	// Dockerfile
	dockerfileContent := `FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cosmoria ./cmd/cosmoria/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /cosmoria /cosmoria
COPY --from=builder /src/db/migrations /db/migrations
EXPOSE 8080
CMD ["/cosmoria"]
`
	if err := os.WriteFile("Dockerfile", []byte(dockerfileContent), 0644); err != nil {
		log.Fatalf("write Dockerfile: %v", err)
	}
	log.Println("created Dockerfile")

	log.Println("init complete. run: docker compose up")
}

func runMigrateNew(name string) {
	if name == "" {
		log.Fatal("migration name is required")
	}

	now := time.Now().UTC()
	ts := now.Format("20060102150405")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	up := filepath.Join("db/migrations", fmt.Sprintf("%s_%s.up.sql", ts, name))
	down := filepath.Join("db/migrations", fmt.Sprintf("%s_%s.down.sql", ts, name))

	if err := os.WriteFile(up, []byte("-- migrate:up\n"), 0644); err != nil {
		log.Fatalf("create up file: %v", err)
	}
	if err := os.WriteFile(down, []byte("-- migrate:down\n"), 0644); err != nil {
		log.Fatalf("create down file: %v", err)
	}

	log.Printf("created migration: %s", filepath.Base(up))
	log.Printf("created migration: %s", filepath.Base(down))

	// update stubs generator
	if _, err := strconv.Atoi(name); err == nil {
		log.Printf("migration files created in db/migrations/")
	}
}

func runMigrateUp() {
	cfg := core.LoadConfig()
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(pool, "db/migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")
}

func runMCP() {
	cfg := core.LoadConfig()
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	server := mcp.NewServer(pool, cfg)
	if err := server.Run(); err != nil {
		log.Fatalf("mcp: %v", err)
	}
}

func runMigrateDown() {
	cfg := core.LoadConfig()
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := db.MigrateDown(pool, "db/migrations"); err != nil {
		log.Fatalf("migrations down: %v", err)
	}
	log.Println("migration reverted")
}
