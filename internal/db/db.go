package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes the pgx connection pool.
// It applies connection pooling settings as required in Phase 3.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	// Phase 3: Connection Pooling
	// Optimizing the pool for concurrent bulk inserts.
	config.MaxConns = 100
	config.MinConns = 10
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	log.Println("Database connection pool established successfully.")
	return pool, nil
}
