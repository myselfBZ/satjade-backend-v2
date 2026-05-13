//go:build !prod

package store

import (
    "context"
    "os"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/joho/godotenv"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
    t.Helper()

    // Load .env.test from project root
    godotenv.Load("../../.env.test") // adjust path depth as needed

    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Fatal("TEST_DATABASE_URL not set")
    }

    pool, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        t.Fatalf("failed to connect: %v", err)
    }

    t.Cleanup(func() { pool.Close() })
    return pool
}
