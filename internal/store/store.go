package store

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Errors returned by store operations
var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
)

// Store wraps a pgxpool.Pool
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a new Store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// CreateAccount inserts a new account with initial balance.
func (s *Store) CreateAccount(ctx context.Context, accountID int64, initial decimal.Decimal) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO accounts (account_id, balance) VALUES ($1, $2)`, accountID, initial.String())
	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}

// GetAccount fetches the current balance for accountID.
func (s *Store) GetAccount(ctx context.Context, accountID int64) (decimal.Decimal, error) {
	var balStr string
	err := s.pool.QueryRow(ctx, `SELECT balance::text FROM accounts WHERE account_id = $1`, accountID).Scan(&balStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, ErrAccountNotFound
		}
		return decimal.Zero, fmt.Errorf("get account: %w", err)
	}
	d, err := decimal.NewFromString(balStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("parse balance: %w", err)
	}
	return d, nil
}

// Transfer performs an atomic transfer from srcID -> dstID of amount.
func (s *Store) Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
	// having some validations upfront
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}

	// No-op when transferring to the same account. Prevents double-lock/update bug.
	if srcID == dstID {
		return nil
	}

	// Begin a DB transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Ensure rollback if not committed
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// To avoid deadlocks, locking rows in ascending order of account_id.
	ids := []int64{srcID, dstID}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	// Fetch balances FOR UPDATE in deterministic order
	balances := make(map[int64]decimal.Decimal, 2)
	for _, id := range ids {
		var balStr string
		row := tx.QueryRow(ctx, `SELECT balance::text FROM accounts WHERE account_id = $1 FOR UPDATE`, id)
		if err := row.Scan(&balStr); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, _ = tx.Exec(ctx, `INSERT INTO transactions (source_account_id, destination_account_id, amount, status, error_message) VALUES ($1,$2,$3,$4,$5)`,
					srcID, dstID, amount.String(), "failed", "account not found")
				return ErrAccountNotFound
			}
			return fmt.Errorf("select balance for account %d: %w", id, err)
		}
		dec, err := decimal.NewFromString(balStr)
		if err != nil {
			return fmt.Errorf("parse balance for account %d: %w", id, err)
		}
		balances[id] = dec
	}

	// Map balances to source/dest
	srcBal, ok1 := balances[srcID]
	dstBal, ok2 := balances[dstID]
	if !ok1 || !ok2 {
		_, _ = tx.Exec(ctx, `INSERT INTO transactions (source_account_id, destination_account_id, amount, status, error_message) VALUES ($1,$2,$3,$4,$5)`,
			srcID, dstID, amount.String(), "failed", "account not found")
		return ErrAccountNotFound
	}

	// Check sufficient funds
	if srcBal.LessThan(amount) {
		_, _ = tx.Exec(ctx, `INSERT INTO transactions (source_account_id, destination_account_id, amount, status, error_message) VALUES ($1,$2,$3,$4,$5)`,
			srcID, dstID, amount.String(), "failed", "insufficient funds")
		return ErrInsufficientFunds
	}

	newSrc := srcBal.Sub(amount)
	newDst := dstBal.Add(amount)

	// Update account balances
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = $1 WHERE account_id = $2`, newSrc.String(), srcID); err != nil {
		return fmt.Errorf("update src balance: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = $1 WHERE account_id = $2`, newDst.String(), dstID); err != nil {
		return fmt.Errorf("update dst balance: %w", err)
	}

	// Insert succeeded transaction row
	if _, err := tx.Exec(ctx, `INSERT INTO transactions (source_account_id, destination_account_id, amount, status) VALUES ($1,$2,$3,$4)`,
		srcID, dstID, amount.String(), "succeeded"); err != nil {
		return fmt.Errorf("insert transaction log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
