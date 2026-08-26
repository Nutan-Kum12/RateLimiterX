# RateLimiterX 

A **production-grade, distributed, extensible rate-limiting framework** and example API server written in Go.

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/Nutan-Kum12/RateLimiterX/actions/workflows/ci.yml/badge.svg)](https://github.com/Nutan-Kum12/RateLimiterX/actions/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](docker/docker-compose.yml)

## ✨ Features

- **2 Production-Ready Algorithms**: Fixed Window & Token Bucket (fully tracked; Sliding Window & Sliding Log available locally)
- **`pkg/ratelimiter` Embeddable Library**: Drop-in Gin middleware you can import directly — supports **Fixed Window** and **Token Bucket**
- **Redis-Backed Distributed State**: Atomic Lua scripts ensure correctness across multiple instances
- **JWT Authentication**: Register & Login with access (15 min) and refresh (7 day) tokens
- **Tier-Based Rate Policies**: `free` and `premium` tiers with independent algorithm + limit configs
- **Strategy + Factory Pattern**: Drop-in addition of new algorithms with one function
- **Prometheus Metrics**: Request counters, latency histograms, rate-limit hit/allow ratio
- **Grafana Dashboard**: Pre-built JSON dashboard included
- **Structured Logging**: Zap-based request logging with request IDs
- **Docker Deployment**: Full stack (API + MySQL + Redis + Prometheus + Grafana) via Docker Compose
- **Repository Pattern**: Clean MySQL data access via `database/sql`
- **CI Pipeline**: GitHub Actions — lint, unit tests, build, Docker push

## 🏗️ Architecture

```
HTTP Request
     │
     ▼
┌─────────────────────────────────────────────┐
│            Gin Middleware Pipeline           │
│  LoggingMiddleware → MetricsMiddleware       │
│       → AuthMiddleware (JWT)                 │
│       → RateLimitMiddleware (tier-aware)     │
└────────────────────┬────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────┐
│              Limiter Manager                 │
│  tier → policy lookup → factory → Limiter   │
│                                              │
│  Limiter interface                           │
│  ├── FixedWindowLimiter  (Lua atomic INCR)  │
│  └── TokenBucketLimiter  (Lua refill logic) │
└────────────────────┬────────────────────────┘
                     │
                     ▼
               ┌───────────┐
               │   Redis   │
               │ (Lua scripts, atomic ops)
               └───────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- Docker & Docker Compose

### Using Docker (Recommended)

```bash
git clone https://github.com/Nutan-Kum12/RateLimiterX.git
cd RateLimiterX

# Copy environment file and set secrets
cp .env.example .env

# Start the full stack
docker compose -f docker/docker-compose.yml up -d
```

| Service | URL |
|---------|-----|
| API | http://localhost:8080 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 (admin / admin) |

### Local Development

```bash
cp .env.example .env
# Fill in JWT secrets, MySQL and Redis credentials in .env

go run ./cmd/server
```

## 📡 API Endpoints

### Public

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Service health check (MySQL + Redis status) |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint |
| `POST` | `/api/v1/auth/register` | Register a new user (assigned `free` tier) |
| `POST` | `/api/v1/auth/login` | Login — returns `access_token` + `refresh_token` |

### Protected (JWT required + rate limited)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/users/me` | Get current authenticated user's profile |

### Example Flow

```bash
# 1. Register
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' | jq

# 2. Login and capture token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  | jq -r '.data.access_token')

# 3. Call protected endpoint
curl -s http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Rate Limit Response Headers

Every protected response includes:

```
X-RateLimit-Limit:     10
X-RateLimit-Remaining: 7
X-RateLimit-Reset:     1753000000
```

When the limit is exceeded (HTTP `429 Too Many Requests`):

```
Retry-After: 45
```

```json
{
  "success": false,
  "message": "rate limit exceeded",
  "error": "too many requests, please try again later"
}
```

## ⚡ Rate-Limiting Algorithms

| Algorithm | Status | `pkg/` Support | Best For | Memory |
|-----------|--------|---------------|----------|--------|
| **Fixed Window** | ✅ Tracked | ✅ Yes | Simple, low-overhead limiting | Very Low |
| **Token Bucket** | ✅ Tracked | ✅ Yes | Bursty traffic with burst capacity | Low |
| **Sliding Window** | 🔧 Local only | ❌ No | Smooth boundary limiting | Low |
| **Sliding Log** | 🔧 Local only | ❌ No | Highest accuracy, per-request timestamps | High |

All algorithms use **atomic Lua scripts** in Redis — no race conditions across instances.

## ⚙️ Configuration

All configuration lives in `config/config.yaml`. Environment variables prefixed with `RATELIMITERX_` override any value (e.g. `RATELIMITERX_JWT_ACCESS_SECRET`).

```yaml
server:
  port: 8080
  mode: release          # debug | release | test

database:
  host: localhost
  port: 3306
  user: root
  password: ""
  name: ratelimiterx

redis:
  addr: localhost:6379
  password: ""
  db: 0

jwt:
  access_secret: ""      # Required — set via .env or env var
  refresh_secret: ""     # Required — set via .env or env var
  access_ttl: 15m
  refresh_ttl: 168h      # 7 days

tiers:
  free:
    algorithm: fixed_window
    limit: 10             # 10 requests per minute
    window: 1m
    burst: 0

  premium:
    algorithm: token_bucket
    limit: 100            # 100 requests per minute
    window: 1m
    burst: 20             # Burst capacity of 20
```

## 📦 pkg/ratelimiter — Embeddable Library

The `pkg/ratelimiter` package is a **standalone, importable Go library** that wraps the core rate-limiting engine into a clean public API. It accepts only the two production-ready algorithms: **Fixed Window** and **Token Bucket**.

```go
import "github.com/Nutan-Kum12/RateLimiterX/pkg/ratelimiter"

// Fixed Window — simple, low-overhead
limiter, err := ratelimiter.New(
    ratelimiter.WithRedis("localhost:6379", "", 0),
    ratelimiter.WithAlgorithm("fixed_window"), // or "token_bucket"
    ratelimiter.WithLimit(100),
    ratelimiter.WithWindow(time.Minute),
)

// Token Bucket — allows bursts
limiter, err := ratelimiter.New(
    ratelimiter.WithRedis("localhost:6379", "", 0),
    ratelimiter.WithAlgorithm("token_bucket"),
    ratelimiter.WithLimit(100),
    ratelimiter.WithWindow(time.Minute),
    ratelimiter.WithBurst(20), // extra burst capacity
)

// Attach as Gin middleware
router := gin.Default()
router.Use(limiter.Middleware(ratelimiter.KeyByIP))       // by client IP
router.Use(limiter.Middleware(ratelimiter.KeyByUserID))   // by authenticated user

// Or call directly (non-Gin / gRPC / etc.)
result, err := limiter.Allow(ctx, "user-42")
if !result.Allowed {
    // reject request
}
```

> **Note** — `sliding_window` and `sliding_log` are **not** accepted by `pkg/ratelimiter`; they are available inside `internal/limiter` for local experimentation only.

## 🧪 Testing

Tests use [miniredis](https://github.com/alicebob/miniredis) — no external Redis required.

```bash
# Unit tests (Fixed Window + Token Bucket)
go test -v -race ./tests/unit/...

# Integration tests (middleware pipeline)
go test -v -race ./tests/integration/...

# pkg/ratelimiter library tests
go test -v -race ./pkg/ratelimiter/...

# All tests
go test -v -race ./...
```

### CI Pipeline (GitHub Actions)

| Job | What it does |
|-----|-------------|
| **Lint** | `golangci-lint` with errcheck, govet, staticcheck, gosec, gocritic, noctx |
| **Unit Tests** | Fixed window & token bucket with `-race` flag |
| **Integration Tests** | Auth + rate-limit middleware with miniredis |
| **Pkg Tests** | `pkg/ratelimiter` library — fixed window & token bucket with `-race` flag |
| **Build** | Cross-compiles `linux/amd64` binary |
| **Docker** | Builds & pushes image to GHCR on `main` |

## 📊 Monitoring

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin / admin)

Pre-built dashboard tracks:
- Request rate by endpoint and status code
- Rate-limit allow vs. deny ratio by tier
- Request latency (p50 / p95 / p99)
- Redis operation latency

## 📂 Project Structure

```
RateLimiterX/
├── cmd/server/              # main.go — application entrypoint
├── config/
│   └── config.yaml          # Default configuration
├── docker/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── grafana/             # Pre-built Grafana dashboard JSON
├── internal/
│   ├── api/
│   │   ├── handler/         # HTTP handlers (auth, user, health)
│   │   ├── middleware/       # Auth, rate-limit, logging, metrics
│   │   └── router.go        # Gin route registration
│   ├── auth/                # JWT token generation & validation
│   ├── configs/             # Config loading (Viper + godotenv)
│   ├── dto/                 # Request / response data types
│   ├── limiter/             # Rate-limiting engine
│   │   ├── strategy.go      # Limiter interface + Result type
│   │   ├── factory.go       # Algorithm factory (NewLimiter)
│   │   ├── manager.go       # Per-tier limiter management
│   │   ├── fixed_window.go  # Fixed Window algorithm
│   │   └── token_bucket.go  # Token Bucket algorithm
│   ├── logger/              # Zap structured logger (singleton)
│   ├── metrics/             # Prometheus counters & histograms
│   ├── model/               # Domain models (User, RateLimitPolicy)
│   ├── mysql/               # MySQL connection setup
│   ├── redis/               # Redis client setup
│   ├── repository/          # Data access layer (UserRepository)
│   └── service/             # Business logic (AuthService, UserService)
├── migrations/              # SQL schema (users table)
├── pkg/
│   └── ratelimiter/         # Embeddable public library (Fixed Window + Token Bucket)
│       ├── ratelimiter.go   # RateLimiter struct, New(), Allow(), Middleware()
│       ├── options.go       # Functional options (WithRedis, WithAlgorithm, …)
│       └── ratelimiter_test.go
├── tests/
│   ├── unit/                # Fixed window & token bucket tests
│   └── integration/         # Middleware pipeline tests
├── .golangci.yml            # Lint configuration
└── .github/workflows/ci.yml # CI pipeline
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -m 'feat: add my feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Open a Pull Request against `main`

## 📄 License

This project is licensed under the MIT License.
