# AeroMentor Wingman Backend

The Go backend for the AeroMentor platform. It provides the core REST API for managing courses, lessons, file uploads, AI-assisted chat, quizzes, and analytics. The architecture follows Clean Architecture / Domain-Driven Design principles, with all external dependencies (database, storage, cache, message broker) abstracted behind ports and swapped in via infrastructure adapters.

## Tech Stack

| Concern | Library/Tool |
|---|---|
| Language | Go 1.25 |
| HTTP Router | go-chi/chi v5 |
| Database | PostgreSQL 15 via pgx v5 |
| SQL Codegen | sqlc v1.31 |
| Migrations | golang-migrate v4 |
| Object Storage | MinIO (minio-go v7) + AWS S3 SDK v2 |
| Resumable Uploads | tusd v2 (TUS protocol) |
| Message Broker | NATS JetStream |
| Vector Database | Qdrant |
| Cache / Rate Limiting | Redis 7 via go-redis v9 |
| Auth | JWT via golang-jwt v5 |
| Config | Viper |
| Logging | Uber Zap |

## Directory Structure

```text
wingman-backend/
├── cmd/
│   └── server/                  # main.go - application entrypoint and wiring
├── docs/                        # ROADMAP and other documentation
├── internal/
│   ├── config/                  # Viper-based config loading (.env / env vars)
│   ├── domain/                  # Core domain models (Course, Lesson, File, Chat, Quiz, etc.)
│   ├── port/                    # Interfaces for storage, cache, queue, RAG client
│   ├── handler/                 # Chi router, middleware, HTTP controllers
│   └── infra/
│       ├── minio/               # ObjectStorage adapter (presign upload/view, delete)
│       ├── nats/                # JetStream publisher and subscriber helpers
│       ├── postgres/            # sqlc-generated queries, repos, migrations
│       ├── redis/               # Cache adapter and token-bucket rate limiter
│       ├── rag/                 # HTTP client for the external RAG/AI service
│       └── tus/                 # Resumable upload handler (post-finish hook -> NATS)
├── tests/                       # Mocks and integration tests
├── Dockerfile                   # Multi-stage build (Go builder + Alpine runtime)
├── docker-compose.yml           # Local dev stack (Postgres, Redis, MinIO, NATS, Qdrant)
├── Makefile                     # Dev, build, test, migrate, sqlc, mock targets
└── go.mod
```

## Features

- Course and lesson management with ownership checks
- File upload flow: presigned MinIO PUT URL returned to the frontend, no file bytes through the server
- Resumable large-file uploads via the TUS protocol (tusd)
- Async document ingestion: after upload, a NATS JetStream message is published to `rag.ingest.request` and picked up by an external AI service
- AI chat and quiz generation via REST + NATS (`rag.ingest.done`, `quiz.generate.*`)
- Analytics tracking
- JWT authentication with role-based access control
- Redis token-bucket rate limiting (60 req/min per user by default)
- All list endpoints return empty arrays instead of `null` (sqlc `emit_empty_slices`)

## Getting Started

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev/) CLI (only needed when modifying SQL queries)
- golangci-lint (only needed for linting)

### Local Dev Setup

**1. Start the infrastructure stack**

```bash
docker compose up -d
```

This starts:
- Postgres 15 on `localhost:5432` (user/pass/db: `aeromentor`)
- Redis 7 on `localhost:6379`
- MinIO on `localhost:9000` (API) and `localhost:9001` (console), root user/pass: `minioadmin`
- NATS 2.10 with JetStream on `localhost:4222`
- Qdrant on `localhost:6333`

**2. Configure environment variables**

Create a `.env` file in the project root (Viper will pick it up automatically):

```env
DATABASE_URL=postgres://aeromentor:aeromentor@localhost:5432/aeromentor?sslmode=disable

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=course-files

NATS_URL=nats://localhost:4222

REDIS_DB=0

JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
```

**3. Run migrations**

```bash
make migrate-up
```

**4. Start the server**

```bash
make dev
```

The server listens on `http://localhost:8080`.

### Docker Build

The `Dockerfile` is a two-stage build. The builder stage compiles the binary with CGO disabled; the runtime stage is a minimal Alpine image with `ca-certificates` and `tzdata`.

```bash
docker build -t wingman-backend .
```

## Development Commands

| Command | Description |
|---|---|
| `make dev` | Kill any existing process on :8080, then run the server |
| `make build` | Compile binary to `bin/server` |
| `make test` | Run all tests with race detector |
| `make lint` | Run golangci-lint |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back the last migration |
| `make sqlc` | Regenerate Go code from SQL queries |
| `make mock` | Regenerate mocks from port interfaces using mockery |
| `make tidy` | Run go mod tidy |

## Contributing

**Adding a new feature:**
1. Define the domain model in `internal/domain/`.
2. Add the persistence layer in `internal/infra/postgres/` (write the SQL query, run `make sqlc`, implement the repo).
3. Implement the application logic in a service under the appropriate layer.
4. Wire up the HTTP handler in `internal/handler/` and register the route in `internal/handler/router.go`.

**Modifying the database schema:**
1. Add a new migration file in `internal/infra/postgres/migrations/`.
2. Update or add queries in `internal/infra/postgres/query/`.
3. Run `make sqlc` to regenerate the Go code.
