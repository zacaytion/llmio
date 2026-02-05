# Quickstart: Discussions & Comments

**Feature**: 005-discussions | **Time**: ~15 minutes

## Prerequisites

- Feature 004 (Groups & Memberships) complete
- PostgreSQL running (`make up`)
- Go 1.25+ installed

## 1. Run Migrations

```bash
make migrate ARGS="up"
```

Creates tables: `discussions`, `comments`, `discussion_readers`

## 2. Start the Server

```bash
make server
```

Server listens on port 8080 (or `PORT` env var).

## 3. Create a Discussion

```bash
# Authenticate first (from Feature 001)
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}' \
  | jq -r '.token')

# Create discussion in a group
curl -X POST http://localhost:8080/discussions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Q1 Planning",
    "description": "Let us discuss our Q1 priorities.",
    "group_id": 1
  }'
```

## 4. Add a Comment

```bash
curl -X POST http://localhost:8080/comments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "discussion_id": 1,
    "body": "I think we should focus on user retention."
  }'
```

## 5. Reply to a Comment

```bash
curl -X POST http://localhost:8080/comments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "discussion_id": 1,
    "parent_id": 1,
    "body": "Good point! What metrics should we track?"
  }'
```

## 6. View Discussion with Comments

```bash
curl http://localhost:8080/discussions/1 \
  -H "Authorization: Bearer $TOKEN"
```

Returns discussion details, comments (flat list with parent_id references), and your read state.

## 7. Close a Discussion

```bash
curl -X POST http://localhost:8080/discussions/1/close \
  -H "Authorization: Bearer $TOKEN"
```

New comments are now blocked.

## Run Tests

```bash
# Go tests
go test ./internal/discussion/... -v

# pgTap schema tests
make test-pgtap
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/api/discussions.go` | Huma operations for discussions |
| `internal/api/comments.go` | Huma operations for comments |
| `internal/discussion/service.go` | Discussion domain logic |
| `internal/discussion/permissions.go` | Permission checks |
| `migrations/007_create_discussions.sql` | Schema |
