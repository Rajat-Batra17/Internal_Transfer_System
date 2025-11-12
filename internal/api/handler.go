package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"

	"github.com/you/internal-transfers/internal/model"
	"github.com/you/internal-transfers/internal/store"
)

// interface for store operations
type StoreAPI interface {
	CreateAccount(ctx context.Context, accountID int64, initial decimal.Decimal) error
	GetAccount(ctx context.Context, accountID int64) (decimal.Decimal, error)
	Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error
}

// API holds the store and request timeout
type API struct {
	store      StoreAPI
	reqTimeout time.Duration
}

// New creates an API instance
func New(s StoreAPI) *API {
	return &API{
		store:      s,
		reqTimeout: 5 * time.Second,
	}
}

// RegisterRoutes registers HTTP routes onto the router.
func (a *API) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/accounts", a.CreateAccount).Methods(http.MethodPost)
	r.HandleFunc("/accounts/{id}", a.GetAccount).Methods(http.MethodGet)
	r.HandleFunc("/transactions", a.CreateTransaction).Methods(http.MethodPost)
}

// writeJSON writes a JSON response with proper headers
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("encode response: %v", err)
		}
	}
}

// CreateAccount creates a new account
func (a *API) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), a.reqTimeout)
	defer cancel()

	if err := a.store.CreateAccount(ctx, req.AccountID, req.InitialBalance.Decimal); err != nil {
		log.Printf("create account failed: accountID=%d, error=%v", req.AccountID, err)
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetAccount retrieves account balance by ID
func (a *API) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid account id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), a.reqTimeout)
	defer cancel()

	bal, err := a.store.GetAccount(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrAccountNotFound) {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		log.Printf("get account failed: accountID=%d, error=%v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := model.AccountResponse{
		AccountID: id,
		Balance:   model.DecimalString{Decimal: bal},
	}
	writeJSON(w, http.StatusOK, resp)
}

// CreateTransaction transfers money between accounts
func (a *API) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req model.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), a.reqTimeout)
	defer cancel()

	if err := a.store.Transfer(ctx, req.SourceAccountID, req.DestinationAccountID, req.Amount.Decimal); err != nil {
		switch {
		case errors.Is(err, store.ErrAccountNotFound):
			http.Error(w, "account not found", http.StatusNotFound)
		case errors.Is(err, store.ErrInsufficientFunds):
			http.Error(w, "insufficient funds", http.StatusConflict)
		default:
			log.Printf("transfer failed: src=%d, dst=%d, amount=%s, error=%v",
				req.SourceAccountID, req.DestinationAccountID, req.Amount.String(), err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
