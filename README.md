# KeyForge 🗝️

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/PostgreSQL-16--Alpine-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Redis-7--Alpine-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis">
  <img src="https://img.shields.io/badge/CI%2FCD-Passing-success?style=for-the-badge&logo=githubactions&logoColor=white" alt="CI/CD">
  <img src="https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge" alt="License">
</p>

<p align="center">
  <b>High-performance API Key lifecycle management and distributed rate-limiting microservice engineered in Go.</b>
</p>

---

## 🚀 Overview

**KeyForge** is a secure, high-throughput API key lifecycle management microservice. Designed following strict **Hexagonal (Ports & Adapters) Architecture**, KeyForge isolates core business logic from external frameworks and delivery mechanisms. It uses **PostgreSQL** for durable persistence, **Redis** for sub-millisecond authentication caching and distributed rate limiting, and an optimized multi-tier cryptographic validation pipeline.

---

## 💎 Core Features

- **Multi-Tier Key Validation:** Generates API keys embedded with an IEEE 802.3 CRC32 checksum (`<prefix>_<random>_<checksum>`) for ultra-fast, zero-allocation pre-validation before database or cache queries.
- **Cryptographic Security:** Stores only HMAC-SHA256 digests; plain keys are returned strictly once upon generation and never persisted.
- **Distributed Rate Limiting:** Enforces per-client quotas backed by Redis using sliding window counters and token-bucket primitives.
- **Hexagonal Architecture:** Strict separation across Domain, Ports, Application Services, Handlers, and Storage Adapters.
- **Distroless Containerization:** Minimal Docker deployment footprint based on `gcr.io/distroless/static-debian12` running as `nonroot`.
- **Automated CI Pipeline:** GitHub Actions workflow executing unit tests, integration tests, race condition detection (`-race`), and atomic coverage reporting against live PostgreSQL and Redis services.

---

## 📊 Tech Stack & Active Dependencies

| Component | Technology | Version | Purpose |
| :--- | :--- | :--- | :--- |
| **Language** | Go (Golang) | `1.26.0` | High-concurrency core runtime |
| **HTTP Router** | `github.com/go-chi/chi/v5` | `v5.2.1` | Idiomatic, low-overhead HTTP routing and middleware chaining |
| **Primary Database** | PostgreSQL (`jackc/pgx/v5`) | `v5.7.2` (Engine: 16-alpine) | High-performance PostgreSQL connection pooling (`pgxpool`) |
| **In-Memory Cache** | Redis (`redis/go-redis/v9`) | `v9.7.1` (Engine: 7-alpine) | Key auth caching and distributed rate limiting |
| **ID Generation** | `segmentio/ksuid` | `v1.0.4` | K-Sortable, 27-character globally unique identifiers |
| **Configuration** | `joho/godotenv` | `v1.5.1` | Local environment file loading |
| **Base Image** | Google Distroless | `static-debian12` | Minimal, non-root attack surface production container |
| **Performance Testing** | Grafana k6 | `latest` | High-concurrency load testing and threshold assertions |

---

## 🛠️ Project Structure

```
.
├── .github/workflows/
│   └── ci.yml                     # Automated CI pipeline (linting, tests with -race, build)
├── cmd/
│   └── keyforge/
│       └── main.go                # Application composition root, dependency wiring & graceful shutdown
├── config/
│   └── config.go                  # Configuration package (internal loaders)
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP delivery controllers (HandleCreateKey, HandleRevokeKey)
│   │   └── middlewares/           # HTTP middlewares (Auth validation, Redis rate limiter)
│   ├── core/
│   │   ├── domain/                # Pure enterprise domain entities (APIKey, Client, KeyStatus)
│   │   ├── ports/                 # Hexagonal interfaces (KeyRepository, ClientRepository, CacheRepository)
│   │   └── services/              # Application services & business logic orchestrator (KeyService)
│   ├── security/                  # Cryptographic routines (HMAC-SHA256, CRC32 checksum, CSPRNG)
│   └── storage/
│       ├── postgres/              # PostgreSQL persistence adapter (pgxpool implementation)
│       └── redis/                 # Redis cache & rate-limiting adapter (go-redis implementation)
├── migrations/
│   └── 0001_init.sql              # Relational schema migrations (clients and api_keys tables & indexes)
├── scripts/
│   └── load_test.js               # k6 distributed load testing script
├── .env.example                   # Baseline environment variable template
├── docker-compose.yml             # Orchestration for API, PostgreSQL 16, and Redis 7
├── Dockerfile                     # Multi-stage distroless build recipe
└── Makefile                       # Build, run, test, and container lifecycle targets
```

---

## ⚙️ Environment Configuration

Copy `.env.example` to `.env` in the project root:

```bash
cp .env.example .env
```

| Variable | Default (Local / Compose) | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | Port on which the HTTP server listens. |
| `DB_URL` | `postgres://keyforge:forgepass@localhost:5432/keyforge_db?sslmode=disable` | PostgreSQL connection DSN (`postgres:5432` inside docker network). |
| `REDIS_URL` | `localhost:6380` (Host via Compose) / `redis:6379` (Container) | Redis host and port (`localhost:6379` if running native Redis). |
| `HMAC_SECRET` | `super-secret-keyforge-hmac-key` | Secret key used to generate HMAC-SHA256 hashes of API tokens. |

> [!NOTE]
> In `docker-compose.yml`, the Redis service maps container port `6379` to host port `6380` to avoid collisions with local Redis instances. Use `REDIS_URL=localhost:6380` when running the application natively against Compose-managed dependencies.

---

## 🔌 API Endpoints Reference

### Public & Health Endpoints

#### `GET /health`
Liveness and readiness probe for orchestration platforms (Kubernetes, Docker).

- **Authentication:** None (Public)
- **Response `200 OK`:**
  ```json
  {
    "status": "healthy"
  }
  ```

---

### Administrative Endpoints (`/api/v1/admin`)

#### `POST /api/v1/admin/keys`
Generates a new API key for an existing client. Calculates the CRC32 checksum, hashes the key using HMAC-SHA256, and saves the hashed record in PostgreSQL.

- **Authentication:** Admin / Internal Network
- **Headers:** `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "client_id": "client_01",
    "prefix": "prod_live",
    "ttl_hours": 720
  }
  ```
- **Response `201 Created`:**
  ```json
  {
    "id": "2tABC123xyz...",
    "plain_key": "prod_live_4k8zN9xP2mL7vQ5R1tY3wE6bC0jS8uD4_a1b2c3d4",
    "expires_at": "2026-12-05T20:00:00Z"
  }
  ```
- **Error Responses:**
  - `400 Bad Request`: Missing or invalid input payload.
  - `404 Not Found`: `client_id` does not exist in the database.
  - `500 Internal Server Error`: Persistence failure.

#### `DELETE /api/v1/admin/keys/{id}`
Revokes an existing active API key by its KSUID identifier.

- **Authentication:** Admin / Internal Network
- **URL Parameters:** `id` (string, required) — Key identifier.
- **Response `204 No Content`**
- **Error Responses:**
  - `400 Bad Request`: Missing key ID.
  - `404 Not Found`: Key ID not found.

---

### Protected Gateway Endpoints (`/api/v1`)

#### `GET /api/v1/resource`
Example protected endpoint demonstrating the middleware pipeline:
1. **CRC32 Checksum Verification:** Verifies the format and mathematical integrity of the token prefix and payload without external I/O.
2. **Auth Cache / HMAC Verification:** Checks Redis for a cached `auth:hash:<plainKey>` entry; on miss, computes HMAC-SHA256, verifies status in PostgreSQL, and caches the result for 5 minutes.
3. **Redis Rate Limiting:** Verifies the client's per-minute quota (`ratelimit:<clientID>`) and returns HTTP 429 if the request threshold is exceeded.

- **Authentication:** `Authorization: Bearer <plain_key>`
- **Response `200 OK`:**
  ```json
  {
    "message": "access granted",
    "client_id": "client_01"
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Missing `Authorization` header, invalid CRC32 checksum, expired key, or revoked key.
  - `429 Too Many Requests`: Client exceeded its configured requests-per-minute quota.

---

## 📈 Performance & Load Testing

A high-concurrency performance benchmark is provided via [k6](https://k6.io/) at `scripts/load_test.js`.

The test ramps up from 100 to 1,000 concurrent Virtual Users (VUs) and asserts the following service-level objectives:
- **95th percentile latency (p95):** `< 20ms`
- **99th percentile latency (p99):** `< 50ms`
- **Request failure rate:** `< 1%`

### Running with k6 CLI:

```bash
# 1. Generate an active key for a seeded client (e.g., client_01)
API_KEY=$(curl -s -X POST http://localhost:8080/api/v1/admin/keys \
  -H "Content-Type: application/json" \
  -d '{"client_id":"client_01","prefix":"prod_live","ttl_hours":24}' | jq -r '.plain_key')

# 2. Run the load test against the protected endpoint
k6 run -e BASE_URL=http://localhost:8080 -e API_KEY=$API_KEY scripts/load_test.js
```

### Running via Docker:

```bash
docker run --rm -i \
  -e BASE_URL=http://host.docker.internal:8080 \
  -e API_KEY=$API_KEY \
  grafana/k6 run - <scripts/load_test.js
```

---

## 🚀 Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.26+](https://golang.org/dl/) (for local host compilation and testing)
- `curl` and `jq` (optional, for CLI testing)

---

### Deployment via Docker Compose

To spin up PostgreSQL 16, Redis 7, and the KeyForge API service simultaneously:

```bash
# Using Makefile
make up

# Or directly via Docker Compose
docker-compose up --build -d
```

Check the health status of running services:

```bash
docker-compose ps
```

Verify the API is running:

```bash
curl -i http://localhost:8080/health
```

Stop and destroy containers and volumes:

```bash
make down
# Equivalent to: docker-compose down -v
```

---

### Local Native Execution

If you prefer executing the Go process directly on the host while running databases in Docker:

1. **Start database dependencies:**
   ```bash
   docker-compose up -d postgres redis
   ```

2. **Execute schema migrations manually (if not auto-initialized):**
   ```bash
   PGPASSWORD=forgepass psql -h localhost -p 5432 -U keyforge -d keyforge_db -f migrations/0001_init.sql
   ```

3. **Configure local environment variables (`.env`):**
   ```env
   PORT=8080
   DB_URL=postgres://keyforge:forgepass@localhost:5432/keyforge_db?sslmode=disable
   REDIS_URL=localhost:6380
   HMAC_SECRET=super-secret-keyforge-hmac-key
   ```

4. **Build and run the application:**
   ```bash
   make run
   # Equivalent to:
   # go build -ldflags="-w -s" -o bin/keyforge ./cmd/keyforge
   # ./bin/keyforge
   ```

---

## 🧪 Testing & Quality Assurance

### Run Unit and Integration Tests

Execute test packages with race detector enabled (on supported platforms):

```bash
make test
# Equivalent to:
go test -v -race ./...
```

> [!NOTE]
> When executing on Windows environments lacking a 64-bit GCC toolchain for CGo race detection, run tests without `-race`:
> ```powershell
> go test -v ./...
> ```

### Generate Code Coverage Profile

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

## 📄 License

Distributed under the **MIT License**. See `LICENSE` for details.
