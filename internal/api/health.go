package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler returns 200 OK when server is alive.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ReadyHandler returns a handler that checks DB pool connectivity.
func ReadyHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			http.Error(w, "db not configured", http.StatusServiceUnavailable)
			return
		}
		// Simple ping using Ping context with short timeout
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
