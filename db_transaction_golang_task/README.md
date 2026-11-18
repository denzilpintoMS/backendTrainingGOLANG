# Bank API

A simple Fiber v3 API using GORM and PostgreSQL. Each user has one account. Money is stored as integer paise: `2500` means ₹25.

## Run locally

Use the Go version in `go.mod` and Docker Compose.

```sh
cp .env.example .env
# Set JWT_SECRET in .env to a random secret of at least 32 bytes.
docker compose up -d --wait
go run ./cmd/api
```

PostgreSQL runs at `localhost:15432`, with database `bank_db` and local username/password `postgres`. Docker stores its data in a persistent volume. `docker compose down` preserves the data; `docker compose down -v` deletes it.

The API reads `DATABASE_URL`, `JWT_SECRET`, and `PORT` from the environment or `.env`; environment variables take precedence. Startup checks the database connection and runs AutoMigrate for User, Account, and Transaction.

## Structure

```text
cmd/api/                Server entry point
internal/config/        Environment configuration
internal/database/      PostgreSQL connection and AutoMigrate
internal/models/        User, Account, Transaction
internal/handlers/      Request validation and database operations
internal/middleware/    JWT authentication
internal/routes/        Route registration
internal/validator/     Request struct validation
examples/queries.sql    SQL practice examples
```

The repository and service folders are reserved for future use. Current logic stays in handlers to keep this exercise simple.

## Endpoints

| Method | Path | JWT required |
| --- | --- | --- |
| GET | `/` | No |
| GET | `/api/v1/health` | No |
| POST | `/api/v1/auth/register` | No |
| POST | `/api/v1/auth/login` | No |
| GET | `/api/v1/accounts/:id` | Yes, account owner only |
| POST | `/api/v1/accounts/deposit` | Yes |
| POST | `/api/v1/accounts/withdraw` | Yes |
| POST | `/api/v1/accounts/transfer` | Yes |

Register a user and automatically create an active account with zero balance:

```sh
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alex","email":"alex@example.com","password":"example-password"}'
```

Passwords must have at least 8 characters and at most 72 bytes. They are hashed with bcrypt and excluded from responses. Emails are stored in lowercase; duplicates return HTTP 409. User and account creation commit together.

Log in to get an `access_token` valid for one hour:

```sh
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alex@example.com","password":"example-password"}'
```

Send `Authorization: Bearer <access_token>` for all account routes. JWT middleware checks the HS256 signature, issuer, expiration, and user ID, then stores the ID in `c.Locals("user_id")`.

```sh
curl http://localhost:3000/api/v1/accounts/1 \
  -H 'Authorization: Bearer <access_token>'

curl -X POST http://localhost:3000/api/v1/accounts/deposit \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"amount":10000}'

curl -X POST http://localhost:3000/api/v1/accounts/withdraw \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"amount":2500}'

curl -X POST http://localhost:3000/api/v1/accounts/transfer \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"to_account_number":"receiver-account-number","amount":2500}'
```

All money endpoints use the authenticated user's account and require positive integer amounts and active accounts. Withdrawals and transfers check sufficient funds. Transfers to the same account are rejected. Deposits simulate adding funds for this exercise.

Each operation begins a database transaction, locks the account(s), updates balances, inserts one transaction record, and commits. `defer tx.Rollback()` undoes changes when an error occurs before commit. Transfers lock both accounts in ID order to avoid conflicting lock orders.

Success returns HTTP 201 with `account_id`, the updated sender/owner `balance`, and `transaction`. Invalid inputs return 400, invalid/missing JWTs return 401, missing or unowned accounts return 404, insufficient funds or inactive accounts return 409, and database failures return 500. Account info and health return 200 on success; health returns 503 if the database is unavailable.

## Checks and SQL practice

```sh
go test ./...
go vet ./...

TEST_DATABASE_URL='postgres://postgres:postgres@localhost:15432/bank_db?sslmode=disable' go test ./... -count=1

docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d bank_db < examples/queries.sql
```

Database integration tests use temporary schemas and remove them afterwards. They cover ownership, money movements, failed operations, rollback, and concurrent requests. They are skipped when `TEST_DATABASE_URL` is unset; the test database user needs permission to create schemas.

The SQL examples demonstrate CRUD, WHERE, ORDER BY, GROUP BY, inner/left joins, primary/foreign keys, and one-to-many relationships using temporary practice tables. The script rolls back its changes and leaves API records untouched. Many-to-many relationships are not needed for this design.

References: [Fiber](https://docs.gofiber.io/), [GORM](https://gorm.io/docs/).
