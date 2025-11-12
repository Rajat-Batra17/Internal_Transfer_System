package model

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidAccountID      = errors.New("account_id must be non-zero")
	ErrInvalidInitialBalance = errors.New("initial_balance must be >= 0")
	ErrInvalidAmount         = errors.New("amount must be > 0")
	ErrSameSourceDestination = errors.New("source and destination must differ")
)

// ValidateCreateAccount validates CreateAccountRequest
func (r *CreateAccountRequest) Validate() error {
	if r.AccountID == 0 {
		return ErrInvalidAccountID
	}
	if r.InitialBalance.IsNegative() {
		return ErrInvalidInitialBalance
	}
	return nil
}

// ValidateTransaction validates TransactionRequest
func (r *TransactionRequest) Validate() error {
	if r.SourceAccountID == 0 || r.DestinationAccountID == 0 {
		return ErrInvalidAccountID
	}
	if r.SourceAccountID == r.DestinationAccountID {
		return ErrSameSourceDestination
	}
	if !r.Amount.GreaterThan(decimal.Zero) {
		return ErrInvalidAmount
	}
	return nil
}
