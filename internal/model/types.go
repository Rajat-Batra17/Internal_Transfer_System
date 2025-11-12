package model

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// DecimalString wraps decimal.Decimal to handle JSON marshaling as strings.
// Prevents precision loss on monetary amounts during JSON serialization.
type DecimalString struct {
	decimal.Decimal
}

// UnmarshalJSON parses decimal from JSON string or number (prefers string).
func (d *DecimalString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		dec, err := decimal.NewFromString(s)
		if err != nil {
			return fmt.Errorf("invalid decimal string: %w", err)
		}
		d.Decimal = dec
		return nil
	}

	// Fallback to float if string parsing fails
	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		d.Decimal = decimal.NewFromFloat(f)
		return nil
	}

	return fmt.Errorf("invalid decimal value")
}

// MarshalJSON outputs decimal as JSON string to preserve precision.
func (d DecimalString) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// Incoming payload for POST /accounts
type CreateAccountRequest struct {
	AccountID      int64         `json:"account_id"`
	InitialBalance DecimalString `json:"initial_balance"`
}

// JSON returned by GET /accounts/{id}
type AccountResponse struct {
	AccountID int64         `json:"account_id"`
	Balance   DecimalString `json:"balance"`
}

// Incoming payload for POST /transactions
type TransactionRequest struct {
	SourceAccountID      int64         `json:"source_account_id"`
	DestinationAccountID int64         `json:"destination_account_id"`
	Amount               DecimalString `json:"amount"`
}
