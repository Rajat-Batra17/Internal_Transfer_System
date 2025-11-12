//go:build integration
// +build integration

package store

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
)

// NOTE:
// - Ensure your Postgres container is running (docker compose up -d db).
// - Ensure migrations were applied (migrations/0001_init.sql).
// - Run: go test ./internal/store -v -tags=integration

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@localhost:5432/transfers?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	t.Cleanup(func() { pool.Close() })

	// cleaning tables to keep test repeatable
	if _, err := pool.Exec(ctx, "DELETE FROM transactions"); err != nil {
		t.Fatalf("failed to clear transactions: %v", err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM accounts"); err != nil {
		t.Fatalf("failed to clear accounts: %v", err)
	}

	return NewStore(pool)
}

func TestConcurrentTransfers(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	// create accounts with large starting balances
	err := s.CreateAccount(ctx, 1, decimal.NewFromInt(1_000_000))
	if err != nil {
		t.Fatalf("CreateAccount 1 failed: %v", err)
	}
	err = s.CreateAccount(ctx, 2, decimal.NewFromInt(1_000_000))
	if err != nil {
		t.Fatalf("CreateAccount 2 failed: %v", err)
	}

	const numTransfers = 500
	amount := decimal.NewFromFloat(1.23)

	var wg sync.WaitGroup
	wg.Add(numTransfers * 2)

	for i := 0; i < numTransfers; i++ {
		// 1 -> 2
		go func() {
			defer wg.Done()
			_ = s.Transfer(ctx, 1, 2, amount)
		}()
		// 2 -> 1
		go func() {
			defer wg.Done()
			_ = s.Transfer(ctx, 2, 1, amount)
		}()
	}

	wg.Wait()

	acc1, err := s.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("GetAccount 1 failed: %v", err)
	}
	acc2, err := s.GetAccount(ctx, 2)
	if err != nil {
		t.Fatalf("GetAccount 2 failed: %v", err)
	}

	total := acc1.Add(acc2)
	expected := decimal.NewFromInt(2_000_000)

	if !total.Equal(expected) {
		t.Fatalf("total mismatch: got %s want %s", total.String(), expected.String())
	}

	if acc1.IsNegative() || acc2.IsNegative() {
		t.Fatalf("negative balance found: a1=%s a2=%s", acc1.String(), acc2.String())
	}
}
