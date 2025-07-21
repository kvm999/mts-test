# 🐳 MTS Docker Infrastructure

This directory contains the Docker infrastructure for the MTS (Mobile TeleSystems) test assignment project.

## 📋 Overview

The Docker setup includes:
- **PostgreSQL 16** - Primary database for users, products, and orders
- **pgAdmin 4** - Web-based PostgreSQL administration tool
- **MTS Backend** - The main application service

## 🚀 Quick Start

### 1. Start Infrastructure Only
```bash
# Start only the database
docker-compose up postgres-mts

# Start database with pgAdmin
docker-compose up postgres-mts pgadmin
```

### 2. Start Everything
```bash
# Start all services (database, pgAdmin, and MTS backend)
docker-compose up

# Start in background
docker-compose up -d
```

### 3. Initialize Database
```bash
# Run the database initialization script
./docker/init-db.sh
```

### 4. Test API
```bash
# Test the MTS API endpoints
./docker/test-api.sh
```

## 🔧 Configuration

### Environment Variables

Copy the example environment file and customize:
```bash
cp docker/env.example .env
# Edit .env with your specific configuration
```

### Key Configuration Files

- `docker/postgres.env` - PostgreSQL configuration
- `docker-compose.yaml` - Docker Compose configuration
- `backend/config.yaml` - MTS application configuration

## 🌐 Service Access

| Service | URL | Credentials |
|---------|-----|-------------|
| MTS API | http://localhost:28080 | - |
| Swagger UI | http://localhost:28080/docs/ | - |
| pgAdmin | http://localhost:25050 | admin@mts.local / admin123 |
| PostgreSQL | localhost:25432 | mts / mts_password_2024 |

## 📊 Database Information

### Connection Details
- **Host**: localhost (from host machine) / postgres-mts (from containers)
- **Port**: 25432 (host) / 5432 (container)
- **Database**: mts
- **Username**: mts
- **Password**: mts_password_2024

### Tables
- `users` - User accounts with authentication
- `products` - Product catalog with inventory
- `orders` - Customer orders
- `order_items` - Order line items with product snapshots

## 🛠️ Development Commands

### Docker Operations
```bash
# Build only the MTS backend image
docker-compose build mts-backend

# View logs
docker-compose logs -f mts-backend

# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ data loss)
docker-compose down -v

# Restart a specific service
docker-compose restart mts-backend
```

### Database Operations
```bash
# Connect to PostgreSQL
docker-compose exec postgres-mts psql -U mts -d mts

# Create database backup
docker-compose exec postgres-mts pg_dump -U mts mts > backup.sql

# Restore database backup
docker-compose exec -T postgres-mts psql -U mts -d mts < backup.sql

# Run database migrations
make migrate-up
```

### API Testing
```bash
# Test API endpoints
./docker/test-api.sh

# Create a test user
curl -X POST http://localhost:28080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "age": 25,
    "password": "SecurePass123!"
  }'

# Get all users
curl http://localhost:28080/api/v1/users?limit=10&offset=0
```

## 🔍 Troubleshooting

### Common Issues

#### 1. Port Already in Use
```bash
# Check what's using the port
lsof -i :25432
# or
netstat -tulpn | grep 25432

# Stop the conflicting service or change ports in docker-compose.yaml
```

#### 2. Database Connection Failed
```bash
# Check if PostgreSQL container is running
docker-compose ps postgres-mts

# Check PostgreSQL logs
docker-compose logs postgres-mts

# Wait for PostgreSQL to be ready
docker-compose exec postgres-mts pg_isready -U mts
```

#### 3. MTS Backend Won't Start
```bash
# Check application logs
docker-compose logs mts-backend

# Rebuild the image
docker-compose build --no-cache mts-backend

# Check configuration
docker-compose exec mts-backend cat /app/config.yaml
```

#### 4. Migrations Failed
```bash
# Check migration status
make migrate-status

# Reset database and run migrations again
make db-reset
make migrate-up
```

### Health Checks

```bash
# Check all container health
docker-compose ps

# Check MTS API health
curl http://localhost:28080/health

# Check PostgreSQL health
docker-compose exec postgres-mts pg_isready -U mts
```

## 📁 Directory Structure

```
docker/
├── README.md           # This file
├── postgres.env        # PostgreSQL environment variables
├── env.example         # Example environment configuration
├── init-db.sh         # Database initialization script
├── test-api.sh         # API testing script
└── postgres-mts/       # PostgreSQL data volume (created automatically)
```

## 🔐 Security Notes

### Development vs Production

This configuration is designed for development. For production:

1. **Change default passwords** in `postgres.env`
2. **Use environment variables** instead of hardcoded values
3. **Enable SSL** for database connections
4. **Use secrets management** for sensitive data
5. **Add proper network security** and firewall rules
6. **Enable monitoring and logging**

### Exposed Ports

All services expose ports on localhost only. In production, consider:
- Using reverse proxy (nginx, traefik)
- Enabling TLS/SSL
- Implementing rate limiting
- Adding authentication middleware

## 📝 Maintenance

### Data Persistence

- PostgreSQL data is persisted in `./docker/postgres-mts/` volume
- pgAdmin settings are persisted in `./docker/pgadmin/` volume
- Remove volumes only if you want to reset all data

### Updates

```bash
# Update base images
docker-compose pull

# Rebuild with latest changes
docker-compose build --no-cache

# Update dependencies
cd backend && go mod tidy
```

## 🤝 Contributing

When modifying the Docker infrastructure:

1. Test changes locally first
2. Update this README if adding new services
3. Maintain compatibility with existing Makefile commands
4. Document any new environment variables 