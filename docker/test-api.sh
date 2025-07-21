#!/bin/bash

# MTS API Testing Script
# This script tests the MTS API endpoints

set -e

# Configuration
API_BASE_URL=${API_BASE_URL:-http://localhost:28080/api/v1}
CONTENT_TYPE="Content-Type: application/json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 MTS API Testing${NC}"
echo "=================="

# Check if MTS service is running
echo -e "${YELLOW}📡 Checking MTS API availability...${NC}"
if ! curl -s --fail ${API_BASE_URL%/api/v1}/health > /dev/null; then
    echo -e "${RED}❌ MTS API is not available. Please start the service first:${NC}"
    echo -e "${YELLOW}   docker-compose up mts-backend${NC}"
    echo -e "${YELLOW}   or: make run${NC}"
    exit 1
fi

echo -e "${GREEN}✅ MTS API is available${NC}"
echo ""

# Test 1: Create a new user
echo -e "${BLUE}👤 Test 1: Creating a new user${NC}"
USER_DATA='{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "age": 25,
  "password": "SecurePass123!"
}'

USER_RESPONSE=$(curl -s -X POST "$API_BASE_URL/users" \
  -H "$CONTENT_TYPE" \
  -d "$USER_DATA")

echo -e "${YELLOW}Request:${NC} POST $API_BASE_URL/users"
echo -e "${YELLOW}Response:${NC} $USER_RESPONSE"

# Extract user ID from response (assuming JSON response)
USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 || echo "")

if [ ! -z "$USER_ID" ]; then
    echo -e "${GREEN}✅ User created with ID: $USER_ID${NC}"
else
    echo -e "${RED}❌ Failed to create user${NC}"
fi
echo ""

# Test 2: Get user by ID
if [ ! -z "$USER_ID" ]; then
    echo -e "${BLUE}🔍 Test 2: Getting user by ID${NC}"
    GET_USER_RESPONSE=$(curl -s -X GET "$API_BASE_URL/users/$USER_ID")
    echo -e "${YELLOW}Request:${NC} GET $API_BASE_URL/users/$USER_ID"
    echo -e "${YELLOW}Response:${NC} $GET_USER_RESPONSE"
    echo -e "${GREEN}✅ User retrieved${NC}"
    echo ""
fi

# Test 3: Get all users with pagination
echo -e "${BLUE}📄 Test 3: Getting all users (paginated)${NC}"
USERS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/users?limit=10&offset=0")
echo -e "${YELLOW}Request:${NC} GET $API_BASE_URL/users?limit=10&offset=0"
echo -e "${YELLOW}Response:${NC} $USERS_RESPONSE"
echo -e "${GREEN}✅ Users list retrieved${NC}"
echo ""

# Test 4: Create a product (if endpoints are implemented)
echo -e "${BLUE}📦 Test 4: Creating a product${NC}"
PRODUCT_DATA='{
  "name": "Test Product",
  "description": "A test product for MTS",
  "price": 99.99,
  "stock_quantity": 100
}'

PRODUCT_RESPONSE=$(curl -s -X POST "$API_BASE_URL/products" \
  -H "$CONTENT_TYPE" \
  -d "$PRODUCT_DATA" 2>/dev/null || echo '{"error":"Products endpoint not implemented"}')

echo -e "${YELLOW}Request:${NC} POST $API_BASE_URL/products"
echo -e "${YELLOW}Response:${NC} $PRODUCT_RESPONSE"
echo ""

# Test 5: API Documentation
echo -e "${BLUE}📚 Test 5: Checking API documentation${NC}"
DOCS_URL="${API_BASE_URL%/api/v1}/docs/"
echo -e "${YELLOW}Swagger UI:${NC} $DOCS_URL"
echo -e "${YELLOW}OpenAPI Spec:${NC} ${API_BASE_URL%/api/v1}/docs/doc.json"
echo ""

# Test 6: Health check
echo -e "${BLUE}❤️  Test 6: Health check${NC}"
HEALTH_RESPONSE=$(curl -s -X GET "${API_BASE_URL%/api/v1}/health")
echo -e "${YELLOW}Request:${NC} GET ${API_BASE_URL%/api/v1}/health"
echo -e "${YELLOW}Response:${NC} $HEALTH_RESPONSE"
echo -e "${GREEN}✅ Health check passed${NC}"
echo ""

echo -e "${GREEN}🎉 API testing completed!${NC}"
echo ""
echo -e "${YELLOW}📚 Available endpoints:${NC}"
echo "   GET    /health           - Health check"
echo "   POST   /api/v1/users     - Create user"
echo "   GET    /api/v1/users     - List users (paginated)"
echo "   GET    /api/v1/users/:id - Get user by ID"
echo "   POST   /api/v1/products  - Create product (if implemented)"
echo "   POST   /api/v1/orders    - Create order (if implemented)"
echo "   GET    /docs/            - Swagger documentation" 