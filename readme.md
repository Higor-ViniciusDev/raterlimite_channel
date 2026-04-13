# Rate Limiter — Sliding Window with Penalty System

A production-ready HTTP rate limiting middleware built in Go, supporting **per-IP** and **per-Token** strategies with configurable sliding windows, penalty blocks, and a concurrent worker pool for high-throughput environments.

Built as part of the [Full Cycle Go Expert](https://fullcycle.com.br/) post-graduation program.

---

## Features

- **Dual strategy**: limits requests by IP address or API token (`API-KEY` header), with token taking priority
- **Sliding window**: request counters reset after a configurable TTL (default: 1 second)
- **Penalty system**: once the limit is exceeded, the client is blocked for a configurable penalty duration — all subsequent requests return `429` until the block expires
- **Worker pool**: requests are processed through a buffered channel and a pool of goroutines, preventing contention under high load
- **In-memory counters**: fast, mutex-protected counters with automatic expiry — no per-request Redis round-trip for counting
- **Token lifecycle**: JWT tokens are created via REST, stored in Redis, and auto-deleted on expiry using a timer-based expirer
- **Structured logging**: JSON logs via `zap`
- **Load testing**: ready-made `Makefile` targets using [Vegeta](https://github.com/tsenart/vegeta)

---

## Architecture

```
HTTP Request
     │
     ▼
RateLimiterMiddleware
     │
     ├── PolicyUsecase.Resolver()
     │       ├── API-KEY header present → TokenStrategy
     │       └── No header             → IPStrategy
     │
     ├── strategy.GenerateKey()        (validates token in Redis if applicable)
     │
     ├── RateLimitMessage → InputChan (buffered)
     │
     └── Worker Pool (N goroutines)
             │
             ├── In-memory Counter (mutex-protected)
             │       ├── Blocked? → reply 429
             │       ├── TTL expired? → reset counter
             │       └── Count >= Limit → apply penalty, reply 429
             │
             └── OK → reply nil → next.ServeHTTP()
                          │
                          └── SaveRequestInfo() → Redis (async, success only)
```

### Request flow — token with penalty

```
t=0.0s  Req #1–10  ✅ 200 OK   (within limit)
t=0.0s  Req #11    🚫 429      (limit exceeded → penalty activated: 10s)
t=0.1s  Req #12+   🚫 429      (penalty active)
t=1.0s  TTL expires             (counter would reset, but penalty is still active)
t=2–9s  all reqs   🚫 429      (penalty active)
t=10s+  penalty expires         (back to normal)
```

---

## Tech Stack

| Layer        | Technology                              |
|--------------|-----------------------------------------|
| Language     | Go 1.24                                 |
| HTTP router  | [chi v5](https://github.com/go-chi/chi) |
| Cache / store| Redis (`go-redis/v9`)                   |
| Auth         | JWT (`go-chi/jwtauth`, HS256)           |
| Logging      | `go.uber.org/zap`                       |
| Testing      | `testing`, `miniredis`, `testify`       |
| Load testing | [Vegeta](https://github.com/tsenart/vegeta) |
| Container    | Docker / Docker Compose                 |

---

## Getting Started

### Prerequisites

- Go 1.24+
- Docker (for Redis)
- [Vegeta](https://github.com/tsenart/vegeta) *(optional — for load tests)*

### 1. Start Redis

```bash
docker-compose up -d
```

### 2. Configure environment

Copy and edit the environment file if needed:

```bash
# cmd/ratelimite/.env
```

Default values are ready to use out of the box (see [Configuration](#configuration) below).

### 3. Run the application

```bash
go run cmd/ratelimite/main.go
```

Server starts at `http://localhost:8080`.

---

## API Endpoints

### `POST /tolken` — Create a token

```bash
curl -X POST http://localhost:8080/tolken
```

**Response `200 OK`:**
```json
{
  "tolken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

The token is stored in Redis and automatically deleted after `TOLKEN_EXPIRATION` seconds.

---

### `GET /` — Rate-limited endpoint

```bash
# Request limited by IP
curl http://localhost:8080/

# Request limited by token
curl http://localhost:8080/ -H "API-KEY: <token>"
```

**Possible responses:**

| Status | Meaning |
|--------|---------|
| `200 OK` | Request allowed |
| `429 Too Many Requests` | Rate limit exceeded or penalty active |
| `400 Bad Request` | Token not found or invalid |

---

## Configuration

All settings live in `cmd/ratelimite/.env`:

```env
# Redis
REDIS_URL=localhost
REDIS_PORT=6379

# Token lifecycle
TOLKEN_EXPIRATION=8        # Token TTL in seconds
JWT_SECRET=secret

# Rate limits
REQUEST_PER_SECOND_IP=5    # Max requests per second per IP
REQUEST_PER_SECOND_TOLKEN=10  # Max requests per second per token

# Penalty durations (applied when limit is exceeded)
TIME_UNLOCKED_NEW_REQUEST_IP=1        # IP block: 1 second
TIME_UNLOCKED_NEW_REQUEST_TOLKEN=10   # Token block: 10 seconds

# Sliding window TTL (how often counters reset)
TLL_KEY_IP=1               # IP counter resets every 1s
TLL_KEY_TOLKEN=1           # Token counter resets every 1s

# Worker pool
WORKER_POOL_SIZE=5         # Number of goroutines processing the queue
SIZE_BUFFER_CHANNEL=1000   # Buffered channel size
```

---

## Running Tests

### Unit & integration tests

```bash
go test ./...
```

The middleware integration test (`internal/middleware/rate_limiter_middleware_test.go`) spins up a real HTTP test server with [miniredis](https://github.com/alicebob/miniredis) — no external dependencies needed.

```bash
go test ./internal/middleware/... -v -run Test_RateLimiterMiddleware_IP
```

### Load tests (requires Vegeta)

```bash
# IP rate limit: 10 req/s for 2s → expect ~10 allowed, ~10 blocked
make test-ip

# Token rate limit: 15 req/s for 10s → expect ~10–15 allowed, ~135–140 blocked
make test-token
```

**Expected result for `make test-token`:**

```
Status Codes:
  200: ~10–15   (6–10%)   ✅ Allowed
  429: ~60–65   (40–43%)  🚫 Blocked by rate limiter
  400: ~74      (~49%)    ⚠️  Token expired (TOLKEN_EXPIRATION=8s, test runs 10s)
```

> The `400` responses after ~8 seconds are expected: the token expires before the test finishes. This is by design and demonstrates the token lifecycle.

---

## Project Structure

```
.
├── cmd/
│   └── ratelimite/
│       ├── main.go               # Entry point, dependency wiring
│       └── .env                  # Environment configuration
│
├── internal/
│   ├── entity/
│   │   ├── request_info_entity/  # RequestInfo domain entity
│   │   └── tolken_entity/        # Token domain entity + repository interface
│   │
│   ├── infra/
│   │   ├── api/
│   │   │   ├── controller/       # HTTP handlers
│   │   │   └── web/              # HTTP server (chi router wrapper)
│   │   └── repository/           # Redis implementations
│   │
│   ├── middleware/
│   │   ├── rate_limiter_middleware.go       # HTTP middleware
│   │   └── rate_limiter_middleware_test.go  # Integration tests
│   │
│   ├── ratelimiter/
│   │   └── local_rate_limiter.go # Worker pool + in-memory counter logic
│   │
│   └── usecase/
│       ├── expire_usecase/       # Timer-based token expiry
│       ├── policy_usecase/       # Strategy resolver + RateLimitStrategy interface
│       ├── strategy_usecase/     # IP and Token strategy implementations
│       └── tolken_usecase/       # Token creation use case
│
├── configuration/
│   ├── database/                 # Redis connection factory
│   ├── logger/                   # Zap logger setup
│   └── rest_err/                 # HTTP error response helpers
│
├── docker-compose.yaml
├── Makefile
└── go.mod
```

---

## Design Decisions

**Why in-memory counters instead of Redis INCR per request?**
Redis round-trips on every request add latency under high load. The worker pool serializes counter access via a mutex-protected map, which is faster for single-instance deployments. For a horizontally scaled setup, replacing the counter backend with a Redis pipeline (`INCR` + `EXPIRE`) or a sliding log in Redis would be the natural next step.

**Why a buffered channel + worker pool?**
This decouples HTTP handler threads from rate limiting logic, preventing a slow Redis write (for `SaveRequestInfo`) from blocking the response. The buffer absorbs traffic spikes, and the pool bounds goroutine count.

**Strategy pattern for IP vs. Token**
`PolicyUsecase.Resolver()` picks the right strategy at runtime. Adding a new limiting dimension (e.g., by user ID or API plan) only requires implementing the `RateLimitStrategy` interface — the middleware and worker are unchanged.

---

## License

MIT
