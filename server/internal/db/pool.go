// Package db provides PostgreSQL connectivity primitives for the
// PhysicsCopilot server. It exposes a connection-pool constructor and
// repository types for sessions and messages.
package db

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns = 50
	defaultMinConns = 5
)

// NewPool creates and validates a pgxpool.Pool for PostgreSQL.
//
// The DSN is read from DATABASE_URL; an error is returned immediately if the
// variable is absent or empty so that the caller can decide whether to abort
// or continue without a database (the server supports running without
// persistence when DATABASE_URL is unset).
//
// Pool sizing is controlled by two environment variables:
//   - DB_POOL_MAX_CONNS: maximum number of open connections (default 50).
//   - DB_POOL_MIN_CONNS: minimum number of idle connections (default 5).
//
// After the pool is created, a Ping is performed to verify reachability.
// If the Ping fails, the pool is closed and the error is returned so that
// callers are not handed a broken pool.
//
// For horizontal deployments with many instances, reduce DB_POOL_MAX_CONNS
// to stay within Postgres's max_connections limit. See docs/SCALING.md.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	cfg.MaxConns = envInt32("DB_POOL_MAX_CONNS", defaultMaxConns)
	cfg.MinConns = envInt32("DB_POOL_MIN_CONNS", defaultMinConns)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// envInt32 reads key from the environment and parses it as a positive int32.
// Returns fallback when the variable is absent, empty, unparseable, or
// non-positive. Used to configure pool sizes from environment variables
// without panicking on malformed input.
func envInt32(key string, fallback int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			return int32(n)
		}
	}
	return fallback
}
