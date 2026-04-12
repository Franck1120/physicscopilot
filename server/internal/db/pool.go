// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

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

// NewPool creates a PostgreSQL connection pool. It reads DATABASE_URL from the
// environment. If DATABASE_URL is not set, an error is returned.
// Pool size is controlled via DB_POOL_MAX_CONNS (default 50) and
// DB_POOL_MIN_CONNS (default 5).
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

// envInt32 reads an environment variable as int32. Returns fallback when the
// variable is unset or not a positive integer.
func envInt32(key string, fallback int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			return int32(n)
		}
	}
	return fallback
}
