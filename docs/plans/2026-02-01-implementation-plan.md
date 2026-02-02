# Loomio Rewrite: Implementation Plan

## Phase 0: Scaffold & Verify Stack (Proof of Concept)

**Goal:** Validate the technology stack works together before committing to full implementation.

### 0.1 Project Initialization

```
loomio/
├── cmd/server/main.go          # Entrypoint
├── internal/
│   ├── api/                    # Huma operations
│   ├── db/                     # sqlc generated + migrations
│   └── config/                 # Environment config
├── migrations/                 # SQL files (embedded via goose)
├── openapi/                    # Source specs (copy from discovery/)
├── generated/                  # oapi-codegen output
├── web/                        # SvelteKit app
├── sqlc.yaml
├── go.mod
└── Makefile                    # Build commands
```

Tasks:
- [ ] Initialize Go module (`go mod init`)
- [ ] Set up Makefile with targets: `generate`, `migrate`, `serve`, `test`
- [ ] Configure sqlc.yaml for pgx/v5
- [ ] Set up goose with embedded migrations
- [ ] Initialize Huma with net/http router
- [ ] Add health check endpoint (`GET /healthz`)

### 0.2 First Migration & Query

Create minimal schema to prove sqlc + goose + pgx work together.

Migration (`migrations/00001_create_users.sql`):
```sql
-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email CITEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
```

sqlc query (`internal/db/queries/users.sql`):
```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (email, name) VALUES ($1, $2) RETURNING *;
```

Tasks:
- [ ] Create migration file
- [ ] Create sqlc queries file
- [ ] Run `sqlc generate`
- [ ] Verify generated Go code compiles
- [ ] Write table-driven test for CreateUser + GetUser

### 0.3 First Huma Endpoint

Implement one endpoint to verify oapi-codegen → Huma integration.

```go
// GET /api/v1/users/:id
type GetUserInput struct {
    ID int64 `path:"id"`
}

type GetUserOutput struct {
    Body UserResponse
}

type UserResponse struct {
    User generated.User `json:"user"`
}
```

Tasks:
- [ ] Run oapi-codegen on `openapi/components/schemas/user.yaml`
- [ ] Create Huma operation using generated types
- [ ] Wire up sqlc queries to handler
- [ ] Verify Huma generates OpenAPI spec matching source
- [ ] Test endpoint returns JSON in expected format

### 0.4 SvelteKit Integration

Prove the full type chain works: OpenAPI → Go → Huma → OpenAPI → TypeScript → Svelte.

Tasks:
- [ ] Initialize SvelteKit project in `web/`
- [ ] Export Huma's generated OpenAPI spec to `web/src/lib/api/openapi.json`
- [ ] Generate TypeScript types with openapi-typescript
- [ ] Create typed API client (fetch wrapper)
- [ ] Build simple page that fetches and displays a user
- [ ] Verify TypeScript catches type mismatches at compile time

### 0.5 Phase 0 Verification Checklist

Before proceeding:
- [ ] `make migrate` runs goose migrations successfully
- [ ] `make generate` runs sqlc and oapi-codegen without errors
- [ ] `make test` passes all Go tests
- [ ] `make serve` starts server on :8080
- [ ] `GET /healthz` returns 200
- [ ] `GET /api/v1/users/:id` returns JSON with correct structure
- [ ] SvelteKit dev server proxies to Go backend
- [ ] TypeScript types match Go response structure
- [ ] Changing OpenAPI spec → regenerate → TypeScript error if frontend doesn't match

---

## Phase 1: Core Infrastructure

### 1.1 Database Schema Migration

Port the PostgreSQL schema from `discovery/schema_dump.sql` to goose migrations.

**Migration order (respecting foreign keys):**

1. Extensions & types
   - Enable citext, hstore, pgcrypto, pg_stat_statements
   - Create custom enums if any

2. Core identity
   - `users` (48+ columns, complex)
   - `groups` (hierarchical, 12 permission flags)
   - `memberships` (join table with roles)

3. Content
   - `discussions`
   - `comments`
   - `reactions`
   - `documents` (file metadata)

4. Decisions
   - `polls`
   - `poll_options`
   - `stances`
   - `stance_choices`
   - `outcomes`

5. Activity
   - `events` (42 types)
   - `notifications`

6. Supporting
   - `tags`, `taggings`
   - `webhooks`, `chatbots`
   - `login_tokens`, `omniauth_identities`
   - `received_emails`

7. Indexes
   - Composite indexes for common queries
   - Full-text search indexes (pg_search)

Tasks:
- [ ] Extract CREATE TABLE statements from schema_dump.sql
- [ ] Split into logical migration files (one per domain)
- [ ] Add corresponding DOWN migrations
- [ ] Run full migration on fresh database
- [ ] Verify schema matches original with `pg_dump --schema-only` diff

### 1.2 Configuration System

Environment-based configuration with sensible defaults.

```go
type Config struct {
    // Server
    Port            int    `env:"PORT" default:"8080"`
    Environment     string `env:"ENV" default:"development"`

    // Database
    DatabaseURL     string `env:"DATABASE_URL" required:"true"`

    // Redis
    RedisURL        string `env:"REDIS_URL" required:"true"`

    // Sessions
    SessionSecret   string `env:"SESSION_SECRET" required:"true"`
    SessionMaxAge   int    `env:"SESSION_MAX_AGE" default:"604800"` // 7 days

    // Storage (S3-compatible)
    StorageEndpoint string `env:"STORAGE_ENDPOINT"`
    StorageBucket   string `env:"STORAGE_BUCKET"`
    StorageKey      string `env:"STORAGE_KEY"`
    StorageSecret   string `env:"STORAGE_SECRET"`

    // Email
    SMTPHost        string `env:"SMTP_HOST"`
    SMTPPort        int    `env:"SMTP_PORT" default:"587"`
    SMTPUser        string `env:"SMTP_USER"`
    SMTPPassword    string `env:"SMTP_PASSWORD"`
    EmailFrom       string `env:"EMAIL_FROM"`
}
```

Tasks:
- [ ] Create config package with env parsing
- [ ] Add validation for required fields
- [ ] Create .env.example with all variables
- [ ] Wire config into main.go

### 1.3 Session Management

Custom session implementation (cookie + Redis/Postgres store).

```go
type Session struct {
    ID        string
    UserID    int64
    Data      map[string]any
    CreatedAt time.Time
    ExpiresAt time.Time
}

type SessionStore interface {
    Get(ctx context.Context, id string) (*Session, error)
    Create(ctx context.Context, userID int64) (*Session, error)
    Update(ctx context.Context, session *Session) error
    Delete(ctx context.Context, id string) error
    DeleteForUser(ctx context.Context, userID int64) error
}
```

Tasks:
- [ ] Define Session type and store interface
- [ ] Implement Redis-backed store
- [ ] Implement Postgres-backed store (fallback)
- [ ] Create Huma middleware for session handling
- [ ] Add secure cookie configuration (HttpOnly, Secure, SameSite)
- [ ] Write tests for session lifecycle

### 1.4 Background Jobs (River)

Set up River for async task processing.

Initial job types:
- `SendEmailJob` - transactional emails
- `ProcessWebhookJob` - outbound webhook delivery
- `CleanupSessionsJob` - periodic session expiry

Tasks:
- [ ] Initialize River with pgx pool
- [ ] Create job registration system
- [ ] Implement SendEmailJob with go-mail
- [ ] Add job processing to server startup
- [ ] Write tests with River's test helpers

---

## Phase 2: Authentication

### 2.1 Password Authentication

Port Devise's password authentication.

Endpoints:
- `POST /api/v1/sessions` - login
- `DELETE /api/v1/sessions` - logout
- `POST /api/v1/users` - registration
- `POST /api/v1/users/password` - password reset request
- `PUT /api/v1/users/password` - password reset confirm

Tasks:
- [ ] Implement bcrypt password hashing (matching Devise cost factor)
- [ ] Create login endpoint with session creation
- [ ] Create logout endpoint with session destruction
- [ ] Create registration endpoint
- [ ] Implement password reset flow with login tokens
- [ ] Add rate limiting for auth endpoints
- [ ] Write integration tests for full auth flow

### 2.2 Login Tokens (Passwordless)

One-time use tokens for email-based login.

Tasks:
- [ ] Create login_tokens table migration
- [ ] Implement token generation (secure random)
- [ ] Create email sending job for login links
- [ ] Implement token verification endpoint
- [ ] Add token expiry and single-use enforcement
- [ ] Test token lifecycle

### 2.3 OAuth Providers

Support for external identity providers.

Providers (in priority order):
1. Google
2. Microsoft
3. Slack
4. SAML (generic)
5. Nextcloud

Tasks:
- [ ] Create OAuth middleware for Huma
- [ ] Implement OAuth2 authorization code flow
- [ ] Create omniauth_identities table
- [ ] Link OAuth identities to users (create or connect)
- [ ] Handle OAuth callback errors gracefully
- [ ] Implement SAML assertion parsing
- [ ] Write tests with mock OAuth server

---

## Phase 3: Core Domain (Read Path)

Implement read-only endpoints first to verify data access patterns.

### 3.1 Users API

Endpoints:
- `GET /api/v1/profile` - current user
- `GET /api/v1/users/:id` - public profile
- `GET /api/v1/memberships` - current user's memberships

Tasks:
- [ ] Generate types from openapi/components/schemas/user.yaml
- [ ] Create sqlc queries for user reads
- [ ] Implement profile endpoint with session user
- [ ] Implement public user endpoint
- [ ] Add sideloading for memberships
- [ ] Test response structure matches spec

### 3.2 Groups API (Read)

Endpoints:
- `GET /api/v1/groups` - list user's groups
- `GET /api/v1/groups/:id` - group details
- `GET /api/v1/groups/:id/memberships` - group members
- `GET /api/v1/groups/:id/subgroups` - child groups

Tasks:
- [ ] Generate types from openapi/components/schemas/group.yaml
- [ ] Create sqlc queries for group reads
- [ ] Implement permission checking (membership required)
- [ ] Handle hierarchical groups (parent/child)
- [ ] Add sideloading for memberships, users
- [ ] Test permission boundaries

### 3.3 Discussions API (Read)

Endpoints:
- `GET /api/v1/discussions` - list discussions
- `GET /api/v1/discussions/:id` - discussion details
- `GET /api/v1/discussions/:id/comments` - comments thread

Tasks:
- [ ] Generate types from openapi/components/schemas/discussion.yaml
- [ ] Create sqlc queries for discussion reads
- [ ] Implement threaded comment loading
- [ ] Handle reader tracking (last read position)
- [ ] Add sideloading for comments, users, polls
- [ ] Test pagination

### 3.4 Polls API (Read)

Endpoints:
- `GET /api/v1/polls` - list polls
- `GET /api/v1/polls/:id` - poll details with options
- `GET /api/v1/polls/:id/stances` - votes

Tasks:
- [ ] Generate types from openapi/components/schemas/poll.yaml
- [ ] Create sqlc queries for poll reads
- [ ] Handle multiple poll types (proposal, ranked choice, dot vote, etc.)
- [ ] Implement anonymous vote handling
- [ ] Add sideloading for options, stances, users
- [ ] Test vote privacy rules

### 3.5 Events & Notifications (Read)

Endpoints:
- `GET /api/v1/events` - activity feed
- `GET /api/v1/notifications` - user notifications

Tasks:
- [ ] Generate types from event.yaml, notification.yaml
- [ ] Create sqlc queries with polymorphic eventable loading
- [ ] Implement notification filtering (unread, type)
- [ ] Handle 42 event types with discriminator
- [ ] Add sideloading for eventable records
- [ ] Test event sequence ordering

---

## Phase 4: Core Domain (Write Path)

### 4.1 Groups API (Write)

Endpoints:
- `POST /api/v1/groups` - create group
- `PATCH /api/v1/groups/:id` - update group
- `DELETE /api/v1/groups/:id` - archive group

Tasks:
- [ ] Implement GroupService with event publishing
- [ ] Handle 12 permission flag updates
- [ ] Create membership on group creation
- [ ] Publish group_create, group_update events
- [ ] Validate handle uniqueness
- [ ] Test permission checks for updates

### 4.2 Discussions API (Write)

Endpoints:
- `POST /api/v1/discussions` - create discussion
- `PATCH /api/v1/discussions/:id` - update discussion
- `DELETE /api/v1/discussions/:id` - discard discussion
- `POST /api/v1/discussions/:id/move` - move to different group
- `POST /api/v1/discussions/:id/close` - close discussion
- `POST /api/v1/discussions/:id/reopen` - reopen discussion

Tasks:
- [ ] Implement DiscussionService with event publishing
- [ ] Handle rich text (markdown/html) with attachments
- [ ] Create reader record on discussion creation
- [ ] Implement move logic (permission check on target group)
- [ ] Publish discussion_* events
- [ ] Test state transitions

### 4.3 Comments API

Endpoints:
- `POST /api/v1/comments` - create comment
- `PATCH /api/v1/comments/:id` - update comment
- `DELETE /api/v1/comments/:id` - delete comment
- `POST /api/v1/reactions` - add reaction
- `DELETE /api/v1/reactions/:id` - remove reaction

Tasks:
- [ ] Implement CommentService with threading
- [ ] Handle reply-to relationships
- [ ] Update discussion last_activity timestamps
- [ ] Publish comment_* events
- [ ] Implement reaction emoji handling
- [ ] Test threading logic

### 4.4 Polls API (Write)

Endpoints:
- `POST /api/v1/polls` - create poll
- `PATCH /api/v1/polls/:id` - update poll
- `DELETE /api/v1/polls/:id` - delete poll
- `POST /api/v1/polls/:id/close` - close poll
- `POST /api/v1/polls/:id/reopen` - reopen poll

Tasks:
- [ ] Implement PollService with type-specific validation
- [ ] Handle poll option creation/updates
- [ ] Enforce closing_at timestamps
- [ ] Publish poll_* events
- [ ] Test each poll type

### 4.5 Stances (Voting)

Endpoints:
- `POST /api/v1/stances` - cast vote
- `PATCH /api/v1/stances/:id` - update vote

Tasks:
- [ ] Implement StanceService with vote revision rules
- [ ] Enforce 15-minute throttle for revisions
- [ ] Handle ranked choice, dot vote, etc.
- [ ] Implement anonymous voting
- [ ] Update poll counts after voting
- [ ] Publish stance_* events
- [ ] Test vote revision logic

### 4.6 Memberships

Endpoints:
- `POST /api/v1/memberships` - add member
- `PATCH /api/v1/memberships/:id` - update role
- `DELETE /api/v1/memberships/:id` - remove member
- `POST /api/v1/membership_requests` - request to join
- `POST /api/v1/membership_requests/:id/approve` - approve request

Tasks:
- [ ] Implement MembershipService
- [ ] Handle role transitions (member → admin → coordinator)
- [ ] Implement membership request workflow
- [ ] Publish membership_* events
- [ ] Test permission inheritance

---

## Phase 5: Real-time & Notifications

### 5.1 Event Publishing

Set up Redis pub/sub for real-time updates.

Tasks:
- [ ] Create EventPublisher that writes to Redis channels
- [ ] Define channel naming scheme (e.g., `group:123`, `user:456`)
- [ ] Publish events from all *Service classes
- [ ] Add event sequence numbers for ordering
- [ ] Test pub/sub connectivity

### 5.2 SSE Endpoint

Server-sent events for real-time updates.

Endpoint:
- `GET /api/v1/events/stream` - SSE stream for current user

Tasks:
- [ ] Create SSE handler with Redis subscription
- [ ] Filter events by user's subscribed channels
- [ ] Handle connection keep-alive
- [ ] Implement reconnection with Last-Event-ID
- [ ] Test with SvelteKit EventSource client

### 5.3 Notification Delivery

Email and push notifications.

Tasks:
- [ ] Create NotificationService with decision tree
- [ ] Implement email templates (using go templates)
- [ ] Queue notification emails via River
- [ ] Mark notifications as read via API
- [ ] Test notification preferences

---

## Phase 6: Frontend (SvelteKit)

### 6.1 Project Structure

```
web/
├── src/
│   ├── lib/
│   │   ├── api/          # Generated types + fetch client
│   │   ├── components/   # Reusable UI components
│   │   ├── stores/       # Svelte stores (session, notifications)
│   │   └── utils/        # Helpers
│   ├── routes/
│   │   ├── +layout.svelte
│   │   ├── +page.svelte         # Home/dashboard
│   │   ├── login/
│   │   ├── groups/
│   │   │   ├── +page.svelte     # Group list
│   │   │   └── [id]/
│   │   │       ├── +page.svelte # Group detail
│   │   │       └── discussions/
│   │   ├── discussions/
│   │   │   └── [id]/
│   │   │       └── +page.svelte # Discussion thread
│   │   └── polls/
│   │       └── [id]/
│   │           └── +page.svelte # Poll voting
│   └── app.html
├── static/
├── tests/                # Playwright tests
├── svelte.config.js
├── vite.config.ts
└── package.json
```

### 6.2 API Client

Tasks:
- [ ] Set up openapi-typescript code generation
- [ ] Create typed fetch wrapper with error handling
- [ ] Implement request/response interceptors
- [ ] Add session token handling
- [ ] Create reactive stores for cached data

### 6.3 Core Pages

Build pages in this order:

1. **Auth pages** - Login, register, password reset
2. **Dashboard** - User's groups and recent activity
3. **Group pages** - List, detail, members, settings
4. **Discussion pages** - List, thread view, comment form
5. **Poll pages** - List, voting UI, results

Tasks per page:
- [ ] Create route with load function (SSR)
- [ ] Build UI components
- [ ] Connect to API client
- [ ] Add loading/error states
- [ ] Write Playwright test

### 6.4 Real-time Integration

Tasks:
- [ ] Create EventSource store for SSE connection
- [ ] Update stores reactively on events
- [ ] Handle reconnection gracefully
- [ ] Show connection status indicator
- [ ] Test with simulated events

### 6.5 File Uploads

Tasks:
- [ ] Integrate Uppy component
- [ ] Create presigned URL endpoint in Go
- [ ] Handle upload progress UI
- [ ] Display uploaded files in discussions/comments
- [ ] Test upload flow end-to-end

---

## Phase 7: Advanced Features

### 7.1 Search

- [ ] Implement full-text search with pg_search
- [ ] Create search endpoint with filters
- [ ] Build search UI in SvelteKit
- [ ] Add search result highlighting

### 7.2 Webhooks

- [ ] Implement outbound webhook delivery via River
- [ ] Support Slack, Discord, Microsoft, generic webhook formats
- [ ] Add webhook management UI
- [ ] Implement retry logic with exponential backoff

### 7.3 Chatbots

- [ ] Implement Matrix bot protocol
- [ ] Create chatbot registration endpoints
- [ ] Handle incoming messages
- [ ] Test with Matrix test server

### 7.4 Templates

- [ ] Implement discussion templates
- [ ] Implement poll templates
- [ ] Create template management UI
- [ ] Test template application

---

## Phase 8: Production Readiness

### 8.1 Security Hardening

- [ ] Audit all endpoints for authorization
- [ ] Implement rate limiting (Rack::Attack equivalent)
- [ ] Add CSRF protection
- [ ] Review OAuth/SAML implementations
- [ ] Address 11 security issues from discovery

### 8.2 Performance

- [ ] Add database connection pooling (pgxpool)
- [ ] Implement query optimization for N+1s
- [ ] Add Redis caching for hot paths
- [ ] Profile and optimize slow endpoints

### 8.3 Observability

- [ ] Add structured logging (slog)
- [ ] Implement request tracing
- [ ] Add Prometheus metrics
- [ ] Create health check endpoints

### 8.4 Deployment

- [ ] Create Dockerfile (multi-stage build)
- [ ] Set up CI/CD pipeline
- [ ] Create Kubernetes manifests / Docker Compose
- [ ] Document deployment process
- [ ] Create database backup strategy

---

## Verification Checkpoints

After each phase, verify:

| Phase | Verification |
|-------|--------------|
| 0 | Stack POC works: Go → OpenAPI → TypeScript → Svelte |
| 1 | Full schema migrated, sessions work, River processes jobs |
| 2 | Can register, login, logout; OAuth works with one provider |
| 3 | Can read all core data; sideloading works |
| 4 | Can create/update all core data; events publish |
| 5 | SSE updates arrive in browser; notifications send |
| 6 | Full UI works; matches original Loomio UX |
| 7 | Search, webhooks, chatbots functional |
| 8 | Passes security audit; ready for production traffic |
