# 🏦 Internal Transfers System (Golang + Postgres)

This project implements a simple **Internal Transfers System** using **Go** and **PostgreSQL**.  
It exposes RESTful HTTP APIs for creating accounts, checking balances, and processing money transfers between accounts.

---

## 🚀 Features

- Create new accounts with an initial balance  
- Fetch account balance by ID  
- Transfer money between two accounts (with validation and atomic updates)  
- PostgreSQL persistence using Docker  
- Clean modular project structure (`internal/api`, `internal/store`, `cmd/server`)

---
## Prerequisites
Make sure you have:
- [Go 1.21+](https://go.dev/doc/install)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- `curl` (for testing HTTP endpoints)

---

## ⚡ Quick Start (One Command)

```bash
bash scripts/setup.sh && go run ./cmd/server
```

This will:
1. Start PostgreSQL in Docker
2. Apply database migrations
3. Create `.env` file with default config
4. Display instructions to run the server

Then in another terminal, test the API:
```bash
bash scripts/test-api.sh
```
---

## 🔧 Manual Setup Instructions
### 1️⃣ Start PostgreSQL with Docker

```bash
docker compose up -d
```

This launches a Postgres 15 instance at port 5432 with default creds:
- **user:** test
- **password:** test
- **database:** transfers

### 2️⃣  Run Database Migrations

```bash
CONTAINER=$(docker compose ps -q db)
docker cp migrations/0001_init.sql $CONTAINER:/tmp/0001_init.sql
docker exec -it $CONTAINER psql -U test -d transfers -f /tmp/0001_init.sql
```

### 3️⃣  Create `.env` File

```bash
cat > .env << EOF
POSTGRES_DSN=postgres://test:test@localhost:5432/transfers?sslmode=disable
PORT=8080
REQ_TIMEOUT_SEC=10
EOF
```

### 4️⃣  Run the Server

```bash
go run ./cmd/server
```

✅ You should see:
```
server listening on :8080
```

---

## � API Endpoints

### Create Account
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 100, "initial_balance": "1000.00"}'
```

### Get Account Balance
```bash
curl http://localhost:8080/accounts/100
```

### Transfer Money
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 100, "destination_account_id": 200, "amount": "50.25"}'
```

### Health Check
```bash
curl http://localhost:8080/healthz
```

---

## 📂 Project Structure

```
internal-transfers/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/                     # HTTP handlers
│   ├── model/                   # Request/response types
│   └── store/                   # Database layer
├── migrations/                  # SQL migration scripts
├── scripts/
│   ├── setup.sh                # One-command setup
│   └── test-api.sh             # API test with curl
├── docker-compose.yml
├── Makefile
├── .env.example
└── README.md
```

---

## 📋 Using Makefile 

```bash
# See all available commands
make help

# One-command setup
make setup

# Run server
make run

# Test API endpoints
make test-api

# Run unit tests
make test

# Run integration tests (requires DB)
make test-integration

# Clean up (stop containers, remove .env)
make clean
```
---


## 🧼 Clean Up

Stop containers and remove generated files:
```bash
make clean
```

Or manually:
```bash
docker compose down
rm .env
```

---

## 🧰 Tech Stack

- **Language:** Go 1.21+
- **Database:** PostgreSQL 15
- **Libraries:** 
  - [shopspring/decimal](https://github.com/shopspring/decimal) — Precise decimal arithmetic
  - [gorilla/mux](https://github.com/gorilla/mux) — HTTP router
  - [pgx](https://github.com/jackc/pgx) — PostgreSQL driver



---

## 📚 For Reviewers

1. **Quick review:** Run `bash scripts/setup.sh && go run ./cmd/server` then `bash scripts/test-api.sh`
2. **Code walk-through:** Start with `cmd/server/main.go` → `internal/api/handler.go` → `internal/store/store.go`
3. **Run tests:** `make test-integration` (requires DB running)

---


