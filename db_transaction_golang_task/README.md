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

Compose creates `bank_db` on first startup and exposes PostgreSQL at `localhost:15432` (port `5432` inside the container). If PostgreSQL already runs locally, create `bank_db` there and update `DATABASE_URL` instead of starting Compose. Environment variables take precedence over `.env`.

```sh
curl http://localhost:3000/
curl http://localhost:3000/api/v1/health
```

The health endpoint checks the database and returns HTTP 503 when unavailable. Startup requires a working database connection. Startup runs GORM AutoMigrate for the User and Account models before accepting requests. Migration errors prevent the server from starting.

```sh
go test ./...
go vet ./...
docker compose down
```

The Compose credentials are local development defaults. Configure credentials and SSL in `DATABASE_URL` for other environments.

References: [Fiber](https://docs.gofiber.io/), [GORM PostgreSQL setup](https://gorm.io/docs/connecting_to_the_database.html).

## Account information

```sh
curl http://localhost:3000/api/v1/accounts/1
```

`GET /api/v1/accounts/:id` returns the account ID, user ID, account number, balance in paise, status, creation time, and deletion time. It returns HTTP 400 for invalid IDs, 404 for missing or soft-deleted accounts, and 500 for database errors.

The account table is created automatically at startup; account creation is not yet implemented, so testing this endpoint requires existing account records. Authentication and account ownership checks are not yet implemented, so this development endpoint accepts any account ID.
