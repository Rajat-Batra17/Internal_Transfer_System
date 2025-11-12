# Makefile
BINARY=internal-transfers
IMAGE=internal-transfers:local

.PHONY: help setup run build test test-integration test-api docker-build docker-run clean

help:
	@echo "Internal Transfers System - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make setup            - Start DB, apply migrations, create .env"
	@echo "  make run              - Start the server"
	@echo "  make test             - Run unit tests"
	@echo "  make test-integration - Run integration tests (requires DB)"
	@echo "  make test-api         - Run API curl tests (requires running server)"
	@echo "  make build            - Build the binary"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run Docker container"
	@echo "  make clean            - Stop containers and remove generated files"
	@echo ""

setup:
	@bash scripts/setup.sh

run:
	go run ./cmd/server

test:
	@echo "🧪 Running unit tests..."
	@go test ./... -v

test-integration:
	@echo "🧪 Running integration tests..."
	@go test ./internal/store -v -tags=integration

test-api:
	@bash scripts/test-api.sh

build:
	go build -o $(BINARY) ./cmd/server

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(IMAGE)

clean:
	@echo "🧹 Cleaning up..."
	@docker compose down
	@rm -f .env
	@echo "✅ Cleanup complete"
