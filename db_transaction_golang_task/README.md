# Bank API

Basic Fiber v3 API.

## Structure

```text
cmd/api/                Server entry point
internal/config/        Environment configuration
internal/database/      Database connections
internal/handlers/      HTTP handlers
internal/routes/        Route registration
internal/models/        Database models
internal/repositories/  Database queries
internal/services/      Business logic
```

## Run locally

Use the Go version in `go.mod`.

```sh
go run ./cmd/api
curl http://localhost:3000/
```

Reference: [Fiber documentation](https://docs.gofiber.io/).
