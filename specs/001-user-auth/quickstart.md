# Quickstart: User Authentication

**Feature**: 001-user-auth | **Branch**: `001-user-auth`

## Prerequisites

- Go 1.25+ installed
- PostgreSQL 18 running locally
- `mise` for version management (see project README)

## Initial Setup

```bash
# Switch to feature branch
git checkout 001-user-auth

# Install tools
mise install

# Create database
createdb loomio_development
```

## Running Tests

```bash
# Unit tests
go test ./internal/auth/... -v

# Integration tests (requires database)
go test ./internal/db/... -v

# API tests
go test ./internal/api/... -v

# All tests
go test ./... -v
```

## Running the Server

```bash
# Start server (port 8080)
go run ./cmd/server

# Or with live reload (if using air)
air
```

## API Examples

### Register a User

```bash
curl -X POST http://localhost:8080/api/v1/registrations \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User",
    "password": "password123",
    "password_confirmation": "password123"
  }'
```

Expected response (201):
```json
{
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User",
    "username": "test-user",
    "email_verified": false,
    "key": "a1b2c3d4e5f6g7h8",
    "created_at": "2026-02-01T12:00:00Z"
  }
}
```

### Log In

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

Note: For login to work, the user must have `email_verified = true`. For testing:

```sql
UPDATE users SET email_verified = true WHERE email = 'test@example.com';
```

### Check Current User

```bash
curl http://localhost:8080/api/v1/sessions/me \
  -b cookies.txt
```

### Log Out

```bash
curl -X DELETE http://localhost:8080/api/v1/sessions \
  -b cookies.txt
```

## Development Workflow

This feature follows TDD (Test-Driven Development) as mandated by the constitution.

### 1. Schema First

```bash
# Edit migration
vim migrations/001_create_users.sql

# Run migrations
go run ./cmd/migrate up
```

### 2. Generate Database Code

```bash
# Edit queries
vim internal/db/queries/users.sql

# Generate Go code
sqlc generate
```

### 3. Write Tests First

```bash
# Write failing test
vim internal/auth/password_test.go

# Run test (should fail)
go test ./internal/auth/... -v -run TestHashPassword
```

### 4. Implement

```bash
# Implement to pass test
vim internal/auth/password.go

# Run test (should pass)
go test ./internal/auth/... -v -run TestHashPassword
```

### 5. Refactor

Review code for clarity, then run all tests to ensure nothing broke.

## Key Files

| File | Purpose |
|------|---------|
| `migrations/001_create_users.sql` | Users table schema |
| `internal/db/queries/users.sql` | SQL queries for sqlc |
| `internal/auth/password.go` | Argon2id hashing |
| `internal/auth/session.go` | In-memory session store |
| `internal/api/auth.go` | HTTP handlers |
| `specs/001-user-auth/contracts/auth.yaml` | OpenAPI specification |

## Debugging

### Session Issues

Sessions are in-memory and lost on server restart. Check:

```bash
# View server logs
go run ./cmd/server 2>&1 | grep session
```

### Database Issues

```bash
# Connect to database
psql loomio_development

# Check users table
SELECT id, email, email_verified, deactivated_at FROM users;
```

### Password Hashing

To verify a password hash is valid Argon2id:

```go
// Hash format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
fmt.Println(strings.HasPrefix(hash, "$argon2id$"))
```
