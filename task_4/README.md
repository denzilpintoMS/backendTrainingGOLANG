# Books API

A small Go API built with Fiber v3. Books live in a mutex-protected in-memory store and are lost when the server restarts. No database or configuration is needed.

Run the server on port 8080:

```sh
go run .
```

| Method | Route | Behavior |
| --- | --- | --- |
| POST | `/books` | Create a book; returns 201 |
| GET | `/books/all` | List books ordered by ID |
| GET | `/books/:id` | Get a book; returns 404 if missing |

Create a book:

```sh
curl -X POST http://localhost:8080/books \
  -H 'Content-Type: application/json' \
  -d '{"name":"The Go Programming Language","author":"Alan Donovan and Brian Kernighan","desc":"A Go reference"}'

curl http://localhost:8080/books/all
curl http://localhost:8080/books/1
```

`name` is required. `author` and `desc` are optional. Unknown fields, invalid JSON, and invalid IDs return 400. Request bodies are limited to 1 MiB. An empty list preserves the original response: `{"message":"no books available"}`.

`main.go` wires the Fiber routes, `handlers.go` handles HTTP validation and responses, and `store.go` contains the in-memory storage methods. Database storage can be added later by replacing those methods.

Run checks:

```sh
go test -race ./...
go vet ./...
```
