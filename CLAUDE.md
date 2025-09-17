# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a scalable e-commerce microservices system built with Go, using the Fiber web framework. The system consists of multiple services orchestrated with Docker Compose and routed through Kong API Gateway.

## Service Architecture

The system follows a microservices architecture with the following services:

- **Kong API Gateway** (Port 3000): Entry point for all API requests with rate limiting, authentication, and routing
- **User Service** (Port 3003): User management, authentication, RBAC (Role-Based Access Control)
- **Product Service** (Port 3004): Product catalog, categories, inventory management
- **Shopping Cart Service** (Port 3005): Shopping cart operations and calculations
- **Crypto Service** (Port 3002): Encryption/decryption utilities
- **Store Service**: (In development)

Each service has its own PostgreSQL database and Redis cache instance.

## Development Commands

### Building and Running
```bash
# Start all services with Docker Compose
docker-compose up -d

# Build and start specific service
docker-compose up --build user-service

# Stop all services
docker-compose down

# View logs for specific service
docker-compose logs -f user-service
```

### Individual Service Development
Each service can be developed independently:
```bash
cd user-service/
go mod tidy
go run main.go

# Or for other services:
cd product-service/
go run main.go
```

### Testing
Currently no test framework is configured. Check individual service directories for test files ending in `_test.go`.

## Code Architecture Patterns

### Domain-Driven Design (Clean Architecture)
Each service follows this structure:
```
service-name/
├── internal/
│   ├── domain/
│   │   ├── entities/        # Core business entities
│   │   ├── repositories/    # Repository interfaces
│   │   └── services/        # Business logic interfaces
│   ├── application/
│   │   ├── dto/            # Data Transfer Objects
│   │   └── services/       # Business logic implementation
│   ├── infrastructure/
│   │   ├── db/             # Database connections (PostgreSQL, Redis)
│   │   ├── repositories/   # Repository implementations
│   │   └── seed/          # Database seeders
│   ├── interfaces/
│   │   └── http/
│   │       ├── handlers/   # HTTP request handlers
│   │       └── routes/     # Route definitions
│   ├── config/            # Configuration management
│   └── utils/             # Utility functions
├── cmd/
│   └── main.go           # Alternative entry point
└── main.go               # Main entry point
```

### Key Technologies and Libraries
- **Web Framework**: Fiber v2 (Express.js-like for Go)
- **ORM**: GORM v1.30+ with PostgreSQL driver
- **Database**: PostgreSQL 16 Alpine
- **Cache**: Redis 7 Alpine
- **Authentication**: JWT tokens (golang-jwt/jwt/v5)
- **Decimal handling**: shopspring/decimal for financial calculations
- **UUID**: google/uuid for primary keys

### Important Conventions
1. **Entity IDs**: All entities use UUID strings as primary keys with auto-generation
2. **Soft Deletes**: Entities use GORM's `DeletedAt` for soft deletion
3. **JSON Tags**: All entity fields include JSON tags, sensitive fields use `json:"-"`
4. **Table Names**: Explicit table naming using `TableName()` methods
5. **RBAC**: User service implements role-based access control with permissions
6. **Environment Configuration**: Each service uses separate `.env` files

### Database Management
- **Migration**: Located in `internal/infrastructure/db/migration.go`
- **Seeding**: Located in `internal/infrastructure/seed/seeder.go`
- **Connection**: PostgreSQL and Redis connections configured in `internal/infrastructure/db/`

### Kong API Gateway Configuration
- Custom authentication plugins for JWT validation
- Service-specific rate limiting (Redis-backed)
- RBAC enforcement at gateway level
- Public/private route separation
- CORS and security headers configured globally

### Inter-Service Communication
Services communicate through HTTP calls. The shopping cart service calls the product service to validate products and get pricing information.

## Important Notes
- All services use Go 1.24.6
- Database passwords and Redis passwords should be configured via environment variables
- The system uses Docker internal networking for service-to-service communication
- Kong routes are configured with specific rate limits per endpoint type (auth, public, admin)
- Services expose different ports internally but are accessed through Kong on port 3000