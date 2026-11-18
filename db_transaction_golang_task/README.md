# Bank API

Basic Fiber v3 API with PostgreSQL and GORM. Database name: `bank_db`.

## Structure

```text
cmd/api/                Server entry point
internal/config/        Environment configuration
internal/database/      PostgreSQL and GORM connection pool
internal/handlers/      HTTP handlers
internal/routes/        Route registration
internal/models/        Database models
internal/repositories/  Database queries
internal/services/      Business logic
```

## Run locally

Use the Go version in `go.mod` and Docker Compose.

```sh
cp .env.example .env
docker compose up -d --wait
go run ./cmd/api
```

Compose creates `bank_db` on first startup. If PostgreSQL already runs locally, create `bank_db` there and update `DATABASE_URL` instead of starting Compose. Environment variables take precedence over `.env`.

```sh
curl http://localhost:3000/
curl http://localhost:3000/api/v1/health
```

The health endpoint checks the database and returns HTTP 503 when unavailable. Startup requires a working database connection. Add models and explicit migrations when introducing application tables.

```sh
go test ./...
go vet ./...
docker compose down
```

The Compose credentials are local development defaults. Configure credentials and SSL in `DATABASE_URL` for other environments.

References: [Fiber](https://docs.gofiber.io/), [GORM PostgreSQL setup](https://gorm.io/docs/connecting_to_the_database.html).
