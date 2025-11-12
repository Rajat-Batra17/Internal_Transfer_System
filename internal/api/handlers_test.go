package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"

	"github.com/you/internal-transfers/internal/model"
	"github.com/you/internal-transfers/internal/store"
)

// MockStore implements StoreAPI for testing
type MockStore struct {
	CreateAccountFunc func(ctx context.Context, accountID int64, initial decimal.Decimal) error
	GetAccountFunc    func(ctx context.Context, accountID int64) (decimal.Decimal, error)
	TransferFunc      func(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error
}

func (m *MockStore) CreateAccount(ctx context.Context, accountID int64, initial decimal.Decimal) error {
	if m.CreateAccountFunc != nil {
		return m.CreateAccountFunc(ctx, accountID, initial)
	}
	return nil
}

func (m *MockStore) GetAccount(ctx context.Context, accountID int64) (decimal.Decimal, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, accountID)
	}
	return decimal.Zero, nil
}

func (m *MockStore) Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
	if m.TransferFunc != nil {
		return m.TransferFunc(ctx, srcID, dstID, amount)
	}
	return nil
}

// TestCreateAccount_Success tests successful account creation
func TestCreateAccount_Success(t *testing.T) {
	mockStore := &MockStore{
		CreateAccountFunc: func(ctx context.Context, accountID int64, initial decimal.Decimal) error {
			return nil
		},
	}
	api := New(mockStore)

	body := []byte(`{"account_id": 100, "initial_balance": "1000.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateAccount(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

// TestCreateAccount_InvalidJSON tests malformed JSON
func TestCreateAccount_InvalidJSON(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateAccount_ZeroAccountID tests validation: account_id cannot be zero
func TestCreateAccount_ZeroAccountID(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	body := []byte(`{"account_id": 0, "initial_balance": "1000.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("account_id must be non-zero")) {
		t.Fatalf("expected error message about account_id, got: %s", w.Body.String())
	}
}

// TestCreateAccount_NegativeBalance tests validation: initial_balance cannot be negative
func TestCreateAccount_NegativeBalance(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	body := []byte(`{"account_id": 100, "initial_balance": "-50.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetAccount_Success tests successful balance retrieval
func TestGetAccount_Success(t *testing.T) {
	mockStore := &MockStore{
		GetAccountFunc: func(ctx context.Context, accountID int64) (decimal.Decimal, error) {
			if accountID == 100 {
				return decimal.RequireFromString("1000.50"), nil
			}
			return decimal.Zero, store.ErrAccountNotFound
		},
	}
	api := New(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/accounts/100", nil)
	w := httptest.NewRecorder()

	// Use gorilla/mux to match the route
	r := mux.NewRouter()
	r.HandleFunc("/accounts/{id}", api.GetAccount).Methods(http.MethodGet)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp model.AccountResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.AccountID != 100 {
		t.Fatalf("expected account_id 100, got %d", resp.AccountID)
	}

	expected := decimal.RequireFromString("1000.50")
	if !resp.Balance.Equal(expected) {
		t.Fatalf("expected balance 1000.50, got %s", resp.Balance.String())
	}
}

// TestGetAccount_InvalidID tests with non-numeric account ID
func TestGetAccount_InvalidID(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/accounts/abc", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/accounts/{id}", api.GetAccount).Methods(http.MethodGet)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetAccount_NotFound tests when account doesn't exist
func TestGetAccount_NotFound(t *testing.T) {
	mockStore := &MockStore{
		GetAccountFunc: func(ctx context.Context, accountID int64) (decimal.Decimal, error) {
			return decimal.Zero, store.ErrAccountNotFound
		},
	}
	api := New(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/accounts/999", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/accounts/{id}", api.GetAccount).Methods(http.MethodGet)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestCreateTransaction_Success tests successful transfer
func TestCreateTransaction_Success(t *testing.T) {
	mockStore := &MockStore{
		TransferFunc: func(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
			return nil
		},
	}
	api := New(mockStore)

	body := []byte(`{"source_account_id": 100, "destination_account_id": 200, "amount": "50.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestCreateTransaction_InvalidJSON tests malformed JSON
func TestCreateTransaction_InvalidJSON(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateTransaction_SameAccount tests validation: source and destination must differ
func TestCreateTransaction_SameAccount(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	body := []byte(`{"source_account_id": 100, "destination_account_id": 100, "amount": "50.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateTransaction_ZeroAmount tests validation: amount must be positive
func TestCreateTransaction_ZeroAmount(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)

	body := []byte(`{"source_account_id": 100, "destination_account_id": 200, "amount": "0"}`)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateTransaction_InsufficientFunds tests transfer with insufficient balance
func TestCreateTransaction_InsufficientFunds(t *testing.T) {
	mockStore := &MockStore{
		TransferFunc: func(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
			return store.ErrInsufficientFunds
		},
	}
	api := New(mockStore)

	body := []byte(`{"source_account_id": 100, "destination_account_id": 200, "amount": "50.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

// TestCreateTransaction_AccountNotFound tests transfer when account doesn't exist
func TestCreateTransaction_AccountNotFound(t *testing.T) {
	mockStore := &MockStore{
		TransferFunc: func(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
			return store.ErrAccountNotFound
		},
	}
	api := New(mockStore)

	body := []byte(`{"source_account_id": 100, "destination_account_id": 200, "amount": "50.00"}`)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
