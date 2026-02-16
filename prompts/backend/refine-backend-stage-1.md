# Backend Stage 1 — Project Scaffold & Infrastructure

## 1. Create Backend Project

- Create the `backend/` directory at the project root (`/pointofsale/backend/`)
- Initialize a Go module: `go mod init github.com/pointofsale/backend` (or appropriate module path)
- Use **Go 1.24** (latest stable)
- Use **Chi v5** as the HTTP router

### Project Structure (Flat/Simple)

```
backend/
├── cmd/
│   └── server/
│       └── main.go            # Entry point: load config, init DB, start server
├── config/
│   └── config.go              # Environment config loader
├── handlers/                   # HTTP handlers (one file per domain)
│   └── health.go              # Health check endpoint
├── middleware/                  # Custom middleware (auth, CORS, logging, etc.)
├── models/                     # GORM model structs
├── repositories/               # Database queries (GORM-based)
├── services/                   # Business logic layer
├── migrations/                 # Versioned SQL migration files (up/down)
├── seeds/                      # Database seed data
├── routes/
│   └── routes.go              # Route definitions, group all endpoints here
├── utils/                      # Shared helpers (password hashing, JWT, etc.)
├── Dockerfile                  # Multi-stage build for Go backend
├── .env.example                # Template for environment variables
├── .air.toml                   # Hot-reload config for development
└── go.mod
```

## 2. Database — PostgreSQL

- Use **PostgreSQL 17** (latest stable)
- Run as a Docker service
- Use **GORM** as the ORM for database queries, model definitions, and relationships
- Use **goose** for versioned SQL migrations (NOT GORM AutoMigrate)
  - All migrations are manual `.sql` files in `backend/migrations/`
  - Support both `up` and `down` migrations for rollback
  - GORM is for querying only; schema changes go through goose
- Write seed files in `backend/seeds/` to populate initial data
- Database credentials and connection string configured via environment variables

## 3. Cache — Redis

- Use **Redis 7** (latest stable)
- Run as a Docker service
- Used for:
  - JWT refresh token storage (with TTL matching token expiry)
  - Session/token blacklisting on logout
  - Future: caching, rate limiting

## 4. Mail — Mailpit

- Use **Mailpit** (latest)
- Run as a Docker service
- SMTP on port `1025`, Web UI on port `8025`
- Used for development email testing (password reset, notifications, etc.)

## 5. Docker & Docker Compose

- Place `docker-compose.yml` at **project root** (`/pointofsale/docker-compose.yml`)
- Orchestrate ALL services:

| Service      | Image / Build        | Ports          | Notes                              |
|--------------|----------------------|----------------|------------------------------------|
| `backend`    | Build from `backend/Dockerfile` | `8080:8080`    | Multi-stage build, hot-reload with Air in dev |
| `frontend`   | Build from `frontend/Dockerfile` | `3000:3000`    | Next.js dev server                 |
| `postgres`   | `postgres:17-alpine` | `5432:5432`    | Named volume for data persistence  |
| `redis`      | `redis:7-alpine`     | `6379:6379`    | Named volume for data persistence  |
| `mailpit`    | `axllent/mailpit`    | `1025, 8025:8025` | SMTP + Web UI                  |

- Create a shared Docker network so services can communicate by service name
- Use named volumes for PostgreSQL and Redis data persistence
- Use `.env` file at project root for shared environment variables
- Create `frontend/Dockerfile` for the Next.js app (dev mode with hot-reload)
- Create `backend/Dockerfile` as multi-stage:
  - **Dev stage**: Use Air for hot-reload during development
  - **Prod stage**: Compile binary, use minimal `scratch` or `alpine` image

## 6. Authentication

- **Password hashing**: Argon2id (use `golang.org/x/crypto/argon2`)
  - Parameters: memory=64MB, iterations=3, parallelism=4, saltLength=16, keyLength=32
- **JWT**: Access + Refresh token flow
  - **Access token**: Short-lived (15 minutes), signed with HS256
  - **Refresh token**: Long-lived (7 days), stored in Redis with TTL
  - Claims: `user_id`, `role`, `exp`, `iat`, `jti` (unique token ID for revocation)
  - On logout: blacklist both tokens in Redis
  - Refresh endpoint to issue new access token using valid refresh token

## 7. Middleware

Set up the following Chi middleware:

- **CORS**: Allow requests from `http://localhost:3000` (frontend), configurable via env
- **Structured Logging**: Use Go's `slog` (standard library) with JSON output
  - Log: request method, path, status code, duration, request ID
- **Request ID**: Generate unique ID per request for tracing
- **Recoverer**: Catch panics and return 500
- **Auth middleware**: Validate JWT access token, inject user context

## 8. Environment Configuration

Use a `.env` file loaded via `github.com/joho/godotenv`.

Required variables:

```env
# Server
APP_ENV=development
APP_PORT=8080
FRONTEND_URL=http://localhost:3000

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=pointofsale
DB_PASSWORD=secret
DB_NAME=pointofsale
DB_SSLMODE=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_ACCESS_SECRET=your-access-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Mail (Mailpit)
SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_FROM=noreply@pointofsale.local
```

## 9. Initial Endpoints

Scaffold these starter endpoints to verify the setup works:

```
GET  /health              → Health check (DB + Redis ping)
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

Route all API endpoints under `/api/v1/` prefix for versioning.

## 10. Go Dependencies

Key packages to install:

| Package | Purpose |
|---------|---------|
| `github.com/go-chi/chi/v5` | HTTP router |
| `github.com/go-chi/cors` | CORS middleware |
| `gorm.io/gorm` | ORM |
| `gorm.io/driver/postgres` | GORM PostgreSQL driver |
| `github.com/pressly/goose/v3` | SQL migrations |
| `github.com/redis/go-redis/v9` | Redis client |
| `github.com/golang-jwt/jwt/v5` | JWT handling |
| `golang.org/x/crypto` | Argon2id password hashing |
| `github.com/joho/godotenv` | .env file loader |
| `github.com/air-verse/air` | Hot-reload for dev (installed in Dockerfile) |

## Deliverables

After completing this stage, I should be able to:

1. Run `docker compose up` from the project root and have all 5 services start
2. Hit `GET http://localhost:8080/health` and get a successful response
3. See Mailpit UI at `http://localhost:8025`
4. Connect to PostgreSQL at `localhost:5432`
5. Connect to Redis at `localhost:6379`
6. Frontend still works at `http://localhost:3000`
7. Hot-reload works for both frontend and backend code changes
