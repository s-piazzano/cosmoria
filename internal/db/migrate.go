package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func Migrate(pool *pgxpool.Pool, migrationsPath string) error {
	conn := stdlib.OpenDBFromPool(pool)
	defer conn.Close()

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("db: migration driver: %w", err)
	}

	src, err := (&file.File{}).Open(migrationsPath)
	if err != nil {
		return fmt.Errorf("db: migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("file", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("db: migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("db: migration up: %w", err)
	}

	return nil
}
