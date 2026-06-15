package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func migrateInstance(pool *pgxpool.Pool, migrationsPath string) (*migrate.Migrate, error) {
	conn := stdlib.OpenDBFromPool(pool)
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("db: migration driver: %w", err)
	}

	src, err := (&file.File{}).Open(migrationsPath)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("db: migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("file", src, "postgres", driver)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("db: migration instance: %w", err)
	}

	return m, nil
}

func Migrate(pool *pgxpool.Pool, migrationsPath string) error {
	m, err := migrateInstance(pool, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("db: migration up: %w", err)
	}
	return nil
}

func MigrateDown(pool *pgxpool.Pool, migrationsPath string) error {
	m, err := migrateInstance(pool, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("db: migration down: %w", err)
	}
	return nil
}
