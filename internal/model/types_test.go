package model

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestDecimalString_UnmarshalJSON_String(t *testing.T) {
	var d DecimalString
	err := json.Unmarshal([]byte(`"100.23344"`), &d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.Equal(decimal.RequireFromString("100.23344")) {
		t.Fatalf("expected 100.23344, got %s", d.String())
	}
}

func TestDecimalString_UnmarshalJSON_Number(t *testing.T) {
	var d DecimalString
	err := json.Unmarshal([]byte(`100.5`), &d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.Equal(decimal.NewFromFloat(100.5)) {
		t.Fatalf("expected 100.5, got %s", d.String())
	}
}

func TestCreateAccountRequest_Validate(t *testing.T) {
	r := CreateAccountRequest{
		AccountID:      0,
		InitialBalance: DecimalString{decimal.NewFromInt(0)},
	}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for zero account id")
	}

	r.AccountID = 1
	r.InitialBalance = DecimalString{decimal.NewFromInt(-1)}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for negative initial balance")
	}

	r.InitialBalance = DecimalString{decimal.NewFromInt(100)}
	if err := r.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionRequest_Validate(t *testing.T) {
	r := TransactionRequest{
		SourceAccountID:      1,
		DestinationAccountID: 1,
		Amount:               DecimalString{decimal.NewFromInt(10)},
	}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error when source == destination")
	}

	r.DestinationAccountID = 2
	r.Amount = DecimalString{decimal.NewFromInt(0)}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for zero amount")
	}
}

// TestDecimalString_MarshalJSON tests JSON marshaling with string output
func TestDecimalString_MarshalJSON(t *testing.T) {
	d := DecimalString{decimal.RequireFromString("123.45")}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `"123.45"`
	if string(b) != expected {
		t.Fatalf("expected %s, got %s", expected, string(b))
	}
}

// TestDecimalString_UnmarshalJSON_Invalid tests invalid decimal string
func TestDecimalString_UnmarshalJSON_Invalid(t *testing.T) {
	var d DecimalString
	err := json.Unmarshal([]byte(`"not_a_number"`), &d)
	if err == nil {
		t.Fatalf("expected error for invalid decimal string")
	}
}

// TestCreateAccountRequest_Validate_MissingFields tests with minimal valid data
func TestCreateAccountRequest_Validate_MissingFields(t *testing.T) {
	// Valid case
	r := CreateAccountRequest{
		AccountID:      1,
		InitialBalance: DecimalString{decimal.NewFromInt(0)},
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected no error for valid account with zero balance, got %v", err)
	}
}

// TestTransactionRequest_Validate_ZeroSourceAccount tests zero source account ID
func TestTransactionRequest_Validate_ZeroSourceAccount(t *testing.T) {
	r := TransactionRequest{
		SourceAccountID:      0,
		DestinationAccountID: 2,
		Amount:               DecimalString{decimal.NewFromInt(10)},
	}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for zero source account ID")
	}
}

// TestTransactionRequest_Validate_ZeroDestinationAccount tests zero destination account ID
func TestTransactionRequest_Validate_ZeroDestinationAccount(t *testing.T) {
	r := TransactionRequest{
		SourceAccountID:      1,
		DestinationAccountID: 0,
		Amount:               DecimalString{decimal.NewFromInt(10)},
	}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for zero destination account ID")
	}
}

// TestTransactionRequest_Validate_NegativeAmount tests negative transfer amount
func TestTransactionRequest_Validate_NegativeAmount(t *testing.T) {
	r := TransactionRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               DecimalString{decimal.NewFromInt(-10)},
	}
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for negative amount")
	}
}

// TestCreateAccountRequest_ZeroBalance tests with zero initial balance (valid)
func TestCreateAccountRequest_ZeroBalance(t *testing.T) {
	r := CreateAccountRequest{
		AccountID:      100,
		InitialBalance: DecimalString{decimal.NewFromInt(0)},
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected no error for zero initial balance, got %v", err)
	}
}

// TestDecimalString_Roundtrip tests marshaling and unmarshaling
func TestDecimalString_Roundtrip(t *testing.T) {
	original := DecimalString{decimal.RequireFromString("999.9999")}

	// Marshal to JSON
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Unmarshal back
	var restored DecimalString
	if err := json.Unmarshal(b, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !restored.Equal(original.Decimal) {
		t.Fatalf("roundtrip failed: expected %s, got %s", original.String(), restored.String())
	}
}
