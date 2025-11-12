#!/bin/bash

# Internal Transfers System - API Test Script
# This script demonstrates all endpoints with sample curl commands.
# Ensure the server is running before executing: go run ./cmd/server

set -e

BASE_URL="http://localhost:8080"
ACCOUNT_1=100
ACCOUNT_2=200

echo "🧪 Internal Transfers System - API Tests"
echo "Base URL: $BASE_URL"
echo ""

# Check if server is running
if ! curl -s "$BASE_URL/healthz" > /dev/null; then
    echo "❌ Server is not running. Start it with: go run ./cmd/server"
    exit 1
fi

echo "✅ Server is healthy"
echo ""

# 1. Health Check
echo "1️⃣  Health Check"
echo "   GET /healthz"
curl -X GET "$BASE_URL/healthz"
echo ""
echo ""

# 2. Create Account 1
echo "2️⃣  Create Account (ID: $ACCOUNT_1, Balance: 1000)"
echo "   POST /accounts"
curl -X POST "$BASE_URL/accounts" \
  -H "Content-Type: application/json" \
  -d '{"account_id": '$ACCOUNT_1', "initial_balance": "1000.00"}'
echo ""
echo ""

# 3. Create Account 2
echo "3️⃣  Create Account (ID: $ACCOUNT_2, Balance: 500)"
echo "   POST /accounts"
curl -X POST "$BASE_URL/accounts" \
  -H "Content-Type: application/json" \
  -d '{"account_id": '$ACCOUNT_2', "initial_balance": "500.00"}'
echo ""
echo ""

# 4. Get Account 1 Balance
echo "4️⃣  Get Account Balance (ID: $ACCOUNT_1)"
echo "   GET /accounts/$ACCOUNT_1"
curl -X GET "$BASE_URL/accounts/$ACCOUNT_1"
echo ""
echo ""

# 5. Get Account 2 Balance
echo "5️⃣  Get Account Balance (ID: $ACCOUNT_2)"
echo "   GET /accounts/$ACCOUNT_2"
curl -X GET "$BASE_URL/accounts/$ACCOUNT_2"
echo ""
echo ""

# 6. Transfer Money (Account 1 -> Account 2)
echo "6️⃣  Transfer Money (From: $ACCOUNT_1, To: $ACCOUNT_2, Amount: 100.50)"
echo "   POST /transactions"
curl -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": '$ACCOUNT_1', "destination_account_id": '$ACCOUNT_2', "amount": "100.50"}'
echo ""
echo ""

# 7. Get Updated Balances
echo "7️⃣  Get Updated Balance (ID: $ACCOUNT_1, Should be 899.50)"
echo "   GET /accounts/$ACCOUNT_1"
curl -X GET "$BASE_URL/accounts/$ACCOUNT_1"
echo ""
echo ""

echo "8️⃣  Get Updated Balance (ID: $ACCOUNT_2, Should be 600.50)"
echo "   GET /accounts/$ACCOUNT_2"
curl -X GET "$BASE_URL/accounts/$ACCOUNT_2"
echo ""
echo ""

# 9. Test Insufficient Funds
echo "9️⃣  Test Insufficient Funds (Transfer: 1000 from Account 1)"
echo "   POST /transactions (Expected: 409 Conflict)"
curl -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": '$ACCOUNT_1', "destination_account_id": '$ACCOUNT_2', "amount": "1000.00"}'
echo ""
echo ""

# 10. Test Invalid Account
echo "🔟 Test Invalid Account (ID: 999)"
echo "   GET /accounts/999 (Expected: 404 Not Found)"
curl -X GET "$BASE_URL/accounts/999"
echo ""
echo ""

echo "✅ All tests completed!"
echo ""
echo "📊 Summary:"
echo "   - Created 2 accounts"
echo "   - Transferred 100.50 from account $ACCOUNT_1 to $ACCOUNT_2"
echo "   - Tested insufficient funds error"
echo "   - Tested invalid account error"
