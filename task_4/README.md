# Books API

A small Go API built with Fiber v3, GORM, and PostgreSQL. Handlers depend on a repository interface; GORM queries, database connections, and model migrations live in `internal/repository`.

## Local setup

Copy the environment template:

```sh
cp .env.example .env
```

For a local database using Docker:

```sh
docker compose up -d --wait
```

Or set `DATABASE_URL` in `.env` to an existing PostgreSQL database. The database must already exist, and the configured user needs permission to migrate the users and books tables and read/write their rows. Follow your database provider's TLS settings for remote connections; `sslmode=disable` in the example is for local development.

Start the server:

```sh
go run .
```

The app loads `.env` automatically; existing environment variables take precedence. `DATABASE_URL` is required, and `PORT` defaults to `8080`. Startup checks the connection and runs GORM AutoMigrate for User and Book. This is intended for local development; use versioned migrations before deploying. There is no in-memory fallback when PostgreSQL is unavailable.

Docker stores database files in a named volume, so they survive container restarts. Stop it with `docker compose down`. Local credentials in `.env` are ignored by Git; `.env.example` is safe to commit.

## API

| Method | Route | Behavior |
| --- | --- | --- |
| POST | `/api/v1/books` | Create a book; returns 201 |
| GET | `/api/v1/books/all` | List books ordered by ID |
| GET | `/api/v1/books/:id` | Get a book; returns 404 if missing |

```sh
curl -X POST http://localhost:8080/api/v1/books \
  -H 'Content-Type: application/json' \
  -d '{"name":"The Go Programming Language","author":"Alan Donovan and Brian Kernighan","desc":"A Go reference"}'

curl http://localhost:8080/api/v1/books/all
curl http://localhost:8080/api/v1/books/1
```

`name` is required. `author` and `desc` are optional. Unknown fields, invalid JSON, and invalid IDs return 400. Request bodies are limited to 1 MiB. An empty list preserves the original response: `{"message":"no books available"}`. Database failures return a generic 500 JSON response, with details logged on the server. Database operations have a five-second timeout.

## Code layout

- `main.go`: environment loading, startup, route wiring, and shutdown.
- `handlers.go`: HTTP validation and responses through `BookRepository`.
- `internal/model/book.go`: book fields, primary key, and timestamps.
- `internal/model/user.go`: user name, case-insensitive unique email, primary key, and timestamps.
- `internal/repository/book.go`: repository interface and not-found error.
- `internal/repository/postgres.go`: GORM setup, PostgreSQL connection pool, AutoMigrate, and book queries.

## Checks

```sh
go test -race ./...
go vet ./...
```

Handler tests use a fake repository without requiring a database. To also run the PostgreSQL integration test, provide a URL for a local test database:

```sh
TEST_DATABASE_URL='postgres://books:books_local@localhost:5432/books?sslmode=disable' \
  go test -race -count=1 -v ./...
```

The integration test creates and removes a unique test schema. The test user needs schema creation permission. It verifies migration from the original books schema, GORM round trips, timestamps, user constraints, API responses, missing books, canceled queries, and persistence after reopening the repository. Without `TEST_DATABASE_URL`, this test is skipped.

## Models and next steps

`Book` has ID, name, author, description, and creation/update timestamps. The existing book JSON response is unchanged; timestamps remain internal. `User` has ID, name, email, and creation/update timestamps. Both models reject blank names at the database level, and user emails must be nonblank and unique without regard to letter case.

Users and books are independent until the app's ownership or borrowing workflow is defined. Registration and credential verification are available through the auth endpoints below. Login returns the user and a success message; persistent sessions, tokens, and protected book routes can be added next.

Potential next steps:

- Add sessions or tokens and protect routes when the authorization workflow is defined.
- For a lending app, add BookCopy and Loan models to track physical copies, borrowers, due dates, and returns.
- Add pagination for lists and versioned migrations before deployment.

## Registration and login

- `POST /api/v1/auth/register`: accepts `name`, `email`, and `password`; returns 201 with the public user, or 409 for an existing email.
- `POST /api/v1/auth/login`: accepts `email` and `password`; returns 200 with the public user, or 401 for invalid credentials. It verifies credentials without creating a session or issuing a token.

Emails are trimmed and normalized to lowercase. Registration requires a nonblank name, a valid email, and a password of at least 8 characters and at most 72 bytes. Passwords are preserved exactly, hashed using bcrypt at its default cost, and excluded from JSON responses. Existing users without a password hash cannot log in.

```sh
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secret123"}'
```
