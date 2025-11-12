#!/bin/bash
set -e

echo "🚀 Internal Transfers System - Quick Start"
echo ""

# Check if Docker is running
if ! docker ps > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

echo "📦 Starting PostgreSQL container..."
docker compose up -d

echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 3

echo "📂 Applying migrations..."
CONTAINER=$(docker compose ps -q db)
if [ -z "$CONTAINER" ]; then
    echo "❌ Failed to get PostgreSQL container ID."
    exit 1
fi

docker cp migrations/0001_init.sql "$CONTAINER:/tmp/0001_init.sql"
docker exec -i "$CONTAINER" psql -U test -d transfers -f /tmp/0001_init.sql

echo "✅ Database setup complete"
echo ""
echo "🔧 Creating .env file..."
if [ ! -f .env ]; then
    cat > .env << EOF
POSTGRES_DSN=postgres://test:test@localhost:5432/transfers?sslmode=disable
PORT=8080
REQ_TIMEOUT_SEC=10
EOF
    echo "✅ .env created"
else
    echo "✅ .env already exists"
fi

echo ""
echo "🎯 Ready to start the server!"
echo ""
echo "Run the following command to start:"
echo "  go run ./cmd/server"
echo ""
echo "Or use the test script:"
echo "  bash scripts/test-api.sh"
echo ""
