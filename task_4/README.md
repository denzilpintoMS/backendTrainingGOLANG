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

The app loads `.env` automatically; existing environment variables take precedence. `DATABASE_URL` and `JWT_SECRET` are required, and `PORT` defaults to `8080`. Startup checks the connection and runs GORM AutoMigrate for User and Book. This is intended for local development; use versioned migrations before deploying. There is no in-memory fallback when PostgreSQL is unavailable.

Docker stores database files in a named volume, so they survive container restarts. Stop it with `docker compose down`. Local credentials in `.env` are ignored by Git; `.env.example` is safe to commit.

## API

| Method | Route | Behavior |
| --- | --- | --- |
| POST | `/api/v1/books` | Create a book with a bearer JWT; returns 201 |
| GET | `/api/v1/books/all` | List books ordered by ID |
| GET | `/api/v1/books/:id` | Get a book; returns 404 if missing |

```sh
curl -X POST http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"The Go Programming Language","author":"Alan Donovan and Brian Kernighan","desc":"A Go reference"}'

curl http://localhost:8080/api/v1/books/all
curl http://localhost:8080/api/v1/books/1
```

`name` is required. `author` and `desc` are optional. Unknown fields, invalid JSON, and invalid IDs return 400. Request bodies are limited to 1 MiB. An empty list preserves the original response: `{"message":"no books available"}`. Database failures return a generic 500 JSON response, with details logged on the server. Database operations have a five-second timeout.

## Code layout

```text
main.go                         # Entry point; calls app.Run
internal/
  app/
    app.go                      # Builds Fiber app; manages startup/shutdown
    app_test.go                 # Book API tests and test helpers
    auth_test.go                # Registration/login tests
    postgres_integration_test.go
  auth/
    jwt.go                      # Signs and validates one-hour access tokens
  middleware/
    jwt.go                      # Requires a bearer token for protected routes
  config/
    config.go                   # Loads .env and environment configuration
  routes/
    routes.go                   # Registers all /api/v1 endpoints
  handler/
    book.go                     # Book HTTP handlers and validation
    auth.go                     # Auth HTTP handlers and bcrypt checks
    errors.go                   # Shared HTTP error responses
  model/
    book.go                     # GORM Book model
    user.go                     # GORM User model
  repository/
    connection.go               # PostgreSQL/GORM connection setup
    postgres.go                 # AutoMigrate and book persistence
    book.go                     # BookRepository interface
    user.go                     # UserRepository interface and persistence
```

Requests flow from routes to handlers to repositories. Handlers depend on repository interfaces, and database queries stay in the repository package. `app.New` accepts repositories so API tests can supply fakes. Run `go run .` from the project root so configuration loads the local `.env` file.

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

Users and books are independent until the app's ownership or borrowing workflow is defined. Registration and credential verification are available through the auth endpoints below. Login returns the user and a signed JWT. Book reads are public; book creation requires the bearer token.

Potential next steps:

- Add refresh tokens or logout/revocation if required by the login workflow.
- For a lending app, add BookCopy and Loan models to track physical copies, borrowers, due dates, and returns.
- Add pagination for lists and versioned migrations before deployment.

## Registration and login

- `POST /api/v1/auth/register`: accepts `name`, `email`, and `password`; returns 201 with the public user, or 409 for an existing email.
- `POST /api/v1/auth/login`: accepts `email` and `password`; returns 200 with the public user, or 401 for invalid credentials. Successful login includes `access_token`, `token_type` (`Bearer`), and `expires_in` (3600 seconds).

Emails are trimmed and normalized to lowercase. Registration requires a nonblank name, a valid email, and a password of at least 8 characters and at most 72 bytes. Passwords are preserved exactly, hashed using bcrypt at its default cost, and excluded from JSON responses. Existing users without a password hash cannot log in.

```sh
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

## JWT authorization

The local `.env` and `.env.example` configure `JWT_SECRET=mindstack`. Tokens are signed with HS256 and contain the user ID (`sub`), issuer, issue time, and a one-hour expiry. Verification checks the signing algorithm, signature, issuer, expiry, and a positive user ID. Missing or invalid tokens on protected routes return 401.

Registration and login are public. `GET /api/v1/books/all` and `GET /api/v1/books/:id` are public. `POST /api/v1/books` requires `Authorization: Bearer <access_token>`. Set `TOKEN` to the `access_token` from the login response before running the book creation example:

```sh
export TOKEN='paste-the-access-token-from-login'
```

Only login issues tokens. Registration returns the public user without a token. The middleware makes the authenticated ID available as `c.Locals("user_id")`.

Request validation uses `github.com/go-playground/validator/v10`. Define request
structs and their `validate` tags in `internal/request/types.go`, then call
`request.Decode(c.Body(), &input)` in the handler and pass errors to
`request.Error(c, err)`. The shared decoder rejects unknown fields and multiple
JSON values, normalizes supported request types, and validates their tags.
Optional `normalize()` methods trim names and normalize email addresses before
validation; passwords are preserved exactly. Custom `request_email` and
`maxbytes` tags preserve mailbox validation and bcrypt's 72-byte password limit.
Validation failures return HTTP 400 with an `error` string naming the fields.
