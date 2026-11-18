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
# Set JWT_SECRET in .env to a random value of at least 32 bytes.
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

The account table is created automatically at startup. Register a user to create an account. Authentication and account ownership checks are not yet implemented, so this development endpoint accepts any account ID.

## Registration

```sh
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alex","email":"alex@example.com","password":"example-password"}'
```

Registration returns HTTP 201 with the user and a new active account with zero balance. Passwords are hashed with bcrypt and excluded from responses. User and account creation happen in one database transaction. Emails are stored in lowercase; duplicates return HTTP 409. Invalid input returns HTTP 400. Passwords must contain at least 8 characters and at most 72 bytes.

## Login

```sh
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alex@example.com","password":"example-password"}'
```

Login returns an `access_token`, `token_type` (`Bearer`), `expires_in` (3600 seconds), and user details without the password. Invalid credentials return HTTP 401; invalid input returns HTTP 400. The signed JWT contains the user ID in `sub` and expires after one hour.

`JWT_SECRET` must be configured in `.env` or the environment and contain at least 32 bytes. The existing account endpoint is still public; The deposit endpoint uses JWT middleware. The account information endpoint is still public.

## JWT middleware

`internal/middleware.JWT(jwtSecret)` validates `Authorization: Bearer <access_token>` headers. It verifies the HS256 signature, issuer, expiration, and user ID, returning HTTP 401 for invalid tokens. Handlers can read the verified user ID as `c.Locals("user_id").(uint)`.

The deposit endpoint uses this middleware. To protect another route later:

```go
app.Get("/api/v1/accounts/:id", middleware.JWT(jwtSecret), handlers.AccountInfo(db))
```

Authentication identifies the user. Account ownership must also be checked separately when protecting account handlers.

## Deposit

Log in, then use the returned `access_token`:

```sh
curl -X POST http://localhost:3000/api/v1/accounts/deposit \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"amount":10050}'
```

Amounts are integer paise: `10050` adds ₹100.50. The account is selected from the authenticated user ID; the request does not choose an account. The account must be active.

A successful deposit returns HTTP 201 with `account_id`, the updated `balance`, and the deposit `transaction`. The balance update and history record commit together, with a row lock to serialize simultaneous deposits. Failed operations roll back both changes.

Invalid amounts return HTTP 400, missing/invalid tokens return 401, missing accounts return 404, inactive accounts return 409, and database errors return 500. This training endpoint simulates adding funds.

Deposit integration tests use a temporary PostgreSQL schema that is removed afterwards:

```sh
TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/handlers -run TestDeposit -count=1
```

Set `TEST_DATABASE_URL` to your local PostgreSQL URL; the test database user needs permission to create schemas. These tests are skipped when it is unset.

## Withdraw

```sh
curl -X POST http://localhost:3000/api/v1/accounts/withdraw \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"amount":2500}'
```

This withdraws ₹25 from the signed-in user's active account. The amount must be a positive integer in paise. The handler begins a database transaction, locks the account, checks the balance, deducts the amount, inserts a withdrawal record, and commits. An error rolls back both changes.

Success returns HTTP 201 with `account_id`, the remaining `balance`, and the `transaction`. Insufficient funds or an inactive account returns HTTP 409. Other error statuses match the deposit endpoint.

Run both PostgreSQL integration tests by setting `TEST_DATABASE_URL` and running `go test ./internal/handlers -run 'TestDeposit|TestWithdraw' -count=1`.
