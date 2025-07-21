#!/bin/bash

# MTS Database Initialization Script
# This script initializes the PostgreSQL database for MTS

set -e

# Database configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-25432}
DB_USER=${DB_USER:-mts}
DB_PASSWORD=${DB_PASSWORD:-mts_password_2024}
DB_NAME=${DB_NAME:-mts}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 MTS Database Initialization${NC}"
echo "=================================="

# Check if PostgreSQL is running
echo -e "${YELLOW}📡 Checking PostgreSQL connection...${NC}"
export PGPASSWORD=$DB_PASSWORD

if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
    echo -e "${RED}❌ PostgreSQL is not ready. Please start the database first:${NC}"
    echo -e "${YELLOW}   docker-compose up postgres-mts${NC}"
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is ready${NC}"

# Create database if it doesn't exist
echo -e "${YELLOW}🗄️  Creating database '$DB_NAME' if it doesn't exist...${NC}"
createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME 2>/dev/null || echo "Database '$DB_NAME' already exists"

# Run migrations
echo -e "${YELLOW}📄 Running database migrations...${NC}"
cd ../backend

# Check if goose is installed
if ! command -v goose &> /dev/null; then
    echo -e "${RED}❌ goose migration tool is not installed${NC}"
    echo -e "${YELLOW}Installing goose...${NC}"
    go install github.com/pressly/goose/v3/cmd/goose@latest
fi

# Run migrations
DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
cd migration/postgres
goose postgres "$DB_URL" up

echo -e "${GREEN}✅ Database initialization completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📊 Database connection details:${NC}"
echo "   Host: $DB_HOST"
echo "   Port: $DB_PORT"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"
echo ""
echo -e "${YELLOW}🌐 pgAdmin access:${NC}"
echo "   URL: http://localhost:25050"
echo "   Email: admin@mts.local"
echo "   Password: admin123" 