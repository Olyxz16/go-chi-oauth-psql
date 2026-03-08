package config

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DSN builds a PostgreSQL connection string from config.
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// NewPostgresPool creates a pgxpool connection pool.
func NewPostgresPool(cfg *PostgresConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to reach postgres: %w", err)
	}

	return pool, nil
}

// Migrate runs idempotent DDL migrations.
func Migrate(pool *pgxpool.Pool) error {
	ctx := context.Background()

	log.Println("Running database migrations...")

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id         UUID        PRIMARY KEY,
			email      TEXT        NOT NULL UNIQUE,
			provider   TEXT        NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	log.Println("Migrations complete.")
	return nil
}
