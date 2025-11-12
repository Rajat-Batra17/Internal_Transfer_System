package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/you/internal-transfers/internal/api"
	"github.com/you/internal-transfers/internal/store"
)

type Config struct {
	PostgresDSN string
	Port        string
	ReqTimeout  time.Duration
}

func loadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("info: .env not loaded: %v (continuing with environment variables)", err)
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return nil, errors.New("POSTGRES_DSN is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	reqTimeout := 5 * time.Second
	if s := os.Getenv("REQ_TIMEOUT_SEC"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			reqTimeout = time.Duration(v) * time.Second
		}
	}

	return &Config{
		PostgresDSN: dsn,
		Port:        port,
		ReqTimeout:  reqTimeout,
	}, nil
}

func main() {

	// Loading required config
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Connecting to Database
	ctx := context.Background()
	pool, err := store.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	// Initializing HTTP API and Router
	s := store.NewStore(pool)
	a := api.New(s)

	// Router and routes
	r := setupRouter(a, pool)

	// Configuring HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server and wait for shutdown
	serverErr := startServer(srv)
	shutdownOnSignal(srv, serverErr)
	log.Println("server gracefully stopped")
}

// startServer starts the HTTP server in a goroutine and returns a channel receiving any server error.
func startServer(srv *http.Server) <-chan error {
	ch := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", srv.Addr)
		ch <- srv.ListenAndServe()
	}()
	return ch
}

// shutdownOnSignal waits for an OS signal or server error and performs a graceful shutdown.
func shutdownOnSignal(srv *http.Server, serverErr <-chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("shutdown signal received")
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

// setupRouter configures middleware, health endpoints and application routes.
func setupRouter(a *api.API, pool *pgxpool.Pool) *mux.Router {
	r := mux.NewRouter()
	r.Use(api.LoggingMiddleware)

	// Health endpoints
	r.HandleFunc("/healthz", api.HealthHandler).Methods(http.MethodGet)
	r.HandleFunc("/readyz", api.ReadyHandler(pool)).Methods(http.MethodGet)

	// Application routes
	a.RegisterRoutes(r)

	return r
}
