# KeyForge 🗝️

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/PostgreSQL-Ready-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Redis-Caching-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis">
  <img src="https://img.shields.io/badge/CI%2FEC-Passing-success?style=for-the-badge&logo=githubactions&logoColor=white" alt="CI/CD">
  <img src="https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge" alt="License">
</p>

<p align="center">
  <b>High-performance API Key management and distributed rate-limiting microservice engineered in Go.</b>
</p>

---

## 🚀 Overview

**KeyForge** is a secure, enterprise-grade backend microservice designed for high-throughput API key lifecycle management (generation, cryptographic hashing, instant pre-validation, and revocation). Built with a strict **Hexagonal (Clean) Architecture**, it completely decouples business logic from external frameworks, leveraging **PostgreSQL** for durable storage and **Redis** for sub-millisecond distributed rate limiting.

---

## 💎 Core Features

- **Advanced Cryptography:** Generates secure keys embedded with CRC32 checksums for fast client-side pre-validation and secures them via HMAC-SHA256 hashing before persistence.
- **Hexagonal Architecture:** Strict separation of concerns (Domain, Ports, Services, Handlers) ensuring modularity, easy testability, and high maintainability.
- **Distributed Rate Limiting:** High-concurrency request throttling powered by Redis to protect downstream infrastructure from abuse or denial-of-service vectors.
- **Automated CI/CD Pipeline:** Fully integrated GitHub Actions workflow executing automated unit tests, integration tests, and race-condition detection (`-race`) on every push.
- **Robust Test Coverage:** Comprehensive test suites covering core business rules, middlewares, cryptographic security layers, and storage drivers using isolated dependency injection and mocks.

---

## 📊 Tech Stack & Architecture

| Layer / Component | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go (Golang) | High concurrency, fast execution, memory safety |
| **HTTP Routing** | `go-chi/chi` | Lightweight, idiomatic, and fast HTTP request router |
| **Primary Database** | PostgreSQL (`pgx`) | Durable relational storage for clients and hashed keys |
| **Cache & Throttling** | Redis (`go-redis`) | Distributed rate-limiting counters and fast lookup cache |
| **Containerization** | Docker / Distroless | Multi-stage secure builds with minimal attack surface |
| **Testing & CI** | Go Test & GitHub Actions | Automated continuous integration and race-free validation |

---

## 🛠️ Project Structure

    ├── cmd/keyforge/            # Application entrypoint (main.go initialization)
    ├── internal/
    │   ├── api/                 # HTTP delivery layer (Controllers/Handlers, Middlewares)
    │   ├── core/                # Domain business logic (Entities, Ports, Services)
    │   ├── security/            # Cryptographic primitives (HMAC-SHA256, Checksums, RNG)
    │   └── storage/             # Infrastructure persistence adapters (PostgreSQL, Redis)
    ├── config/                  # Environment configuration loaders
    └── .github/workflows/       # Automated CI/CD integration pipelines

---

## ⚙️ Environment Configuration

Create a `.env` file in the root directory based on the following template:

PORT=8080
DB_URL=postgres://keyforge:forgepass@localhost:5433/keyforge_db?sslmode=disable
REDIS_URL=127.0.0.1:6379
RATE_LIMIT_DEFAULT=100

---

## 🔌 API Endpoints Reference

| Method | Endpoint | Description | Protected / Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/keys` | Generates a new API key for a specified client. | Internal / Admin |
| `POST` | `/api/v1/keys/verify` | Validates an incoming API key and evaluates rate limits. | Bearer Token / Public Proxy |
| `DELETE` | `/api/v1/keys/{id}` | Revokes an existing active API key. | Internal / Admin |

---

## 📈 Performance & Load Testing

KeyForge is designed for low latency under high concurrency. You can run performance and stress benchmarks using **k6**:

```bash
k6 run ./load-tests/load_test.js
```

---

## 🚀 Getting Started

### Prerequisites
- Docker & Docker Compose installed on your machine.
- Go 1.26 or higher (for local development).

### Deployment via Docker Compose
To spin up the entire architecture locally with a single command:

```bash
docker-compose up --build -d
```

### Running the Test Suite
To execute all unit and integration tests across packages with the race detector enabled:

```bash
go test -v -race ./...
```

To generate and analyze code statement coverage locally:

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📄 License

Distributed under the **MIT License**. See `LICENSE` for more information.
