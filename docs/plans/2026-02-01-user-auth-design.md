# User Authentication Design

**Date**: 2026-02-01
**Status**: Ready for specification
**Scope**: Email/password authentication with session management for Loomio rewrite MVP

## Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Auth methods | Email/password only | Foundation first; OAuth/magic links deferred |
| Session storage | In-memory (Go map) | Simpler for MVP; upgrade to DB/Redis later |
| Password hashing | Argon2id | Modern, memory-hard, resists GPU attacks |
| Argon2id params | 19 MiB, 2 iter, 1 parallel | User-specified; balances security and latency |
| Security features | Basic only | No lockout, no breach checking initially |
| Session lifetime | Fixed 7 days | No "remember me" option for MVP |
| Response shape | `{ users, memberships, groups }` | Matches existing Loomio API; empty arrays until features exist |

## API Endpoints

### POST /api/v1/registrations

Create new user account.

**Request**:
```json
{
  "user": {
    "email": "user@example.com",
    "name": "Jane Doe",
    "password": "secretpassword",
    "password_confirmation": "secretpassword"
  }
}
```

**Response** (200):
```json
{
  "users": [{
    "id": 1,
    "email": "user@example.com",
    "name": "Jane Doe",
    "username": "janedoe",
    "email_verified": false
  }]
}
```

**Errors**: 422 (validation failed)

### POST /api/v1/sessions

Authenticate user and create session.

**Request**:
```json
{
  "user": {
    "email": "user@example.com",
    "password": "secretpassword"
  }
}
```

**Response** (200):
```json
{
  "users": [{
    "id": 1,
    "email": "user@example.com",
    "name": "Jane Doe",
    "username": "janedoe",
    "email_verified": true,
    "secret_token": "uuid-for-websocket-auth",
    "has_password": true
  }],
  "memberships": [],
  "groups": []
}
```

**Cookie**: `session=<token>; HttpOnly; Secure; SameSite=Lax; Max-Age=604800`

**Errors**: 401 (invalid credentials)

### DELETE /api/v1/sessions

Destroy current session.

**Response** (200):
```json
{
  "success": "ok"
}
```

## Data Model

### Users Table

```sql
CREATE TABLE users (
  id                BIGSERIAL PRIMARY KEY,
  email             CITEXT NOT NULL UNIQUE,
  name              VARCHAR(255),
  username          VARCHAR(255) UNIQUE,
  key               VARCHAR(8) UNIQUE,              -- public URL key

  password_hash     TEXT,                           -- Argon2id, NULL if OAuth-only
  email_verified    BOOLEAN NOT NULL DEFAULT FALSE,
  secret_token      UUID NOT NULL DEFAULT gen_random_uuid(),  -- WebSocket auth

  deactivated_at    TIMESTAMPTZ,
  bot               BOOLEAN NOT NULL DEFAULT FALSE,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email_verified ON users(email_verified);
```

### Session Store (In-Memory)

```go
type Session struct {
    UserID    int64
    ExpiresAt time.Time
    UserAgent string
    IPAddress string
}

// Thread-safe map: token (string) -> Session
// Background goroutine cleans expired sessions
// All sessions lost on server restart
```

## Authentication Flows

### Registration

1. Validate email format, uniqueness (case-insensitive via CITEXT)
2. Validate name present
3. Validate password >= 8 chars, matches confirmation
4. Hash password with Argon2id
5. Generate username from name/email
6. Generate 8-char public key
7. Insert user with `email_verified=false`
8. Return user data (no auto-login)

### Login

1. Find user by email (case-insensitive)
2. Verify: user exists AND `email_verified=true` AND `deactivated_at=NULL`
3. Verify password against Argon2id hash
4. Generate 32-byte random session token
5. Store in memory: `token -> { user_id, expires_at, user_agent, ip }`
6. Set HttpOnly secure cookie
7. Return user + empty memberships/groups arrays

### Logout

1. Extract token from cookie
2. Delete from in-memory session store
3. Clear cookie
4. Return success

### Session Middleware

1. Extract session token from cookie
2. Look up in memory store
3. Check not expired
4. If valid: attach user_id to request context
5. If invalid: return 401 or continue as anonymous (depending on route)

## Implementation Structure

```
internal/
  auth/
    auth.go           # RegisterHandler, LoginHandler, LogoutHandler, Middleware
    password.go       # HashPassword, VerifyPassword (Argon2id)
    session.go        # Store, Session, NewStore, Create, Get, Delete

migrations/
  001_create_users.sql
```

## Future Enhancements (Out of Scope)

- OAuth/SAML SSO
- Passwordless magic links
- Account lockout after failed attempts
- Password breach checking (HaveIBeenPwned)
- "Remember me" option
- Database-backed sessions (for revocation, multi-device view)
- Email verification flow

## References

- Existing API: `discovery/openapi/paths/auth.yaml`
- User model: `discovery/specifications/models/user.md`
- Schema: `discovery/schema_dump.sql` (users table at line 2278)
- Constitution: `.specify/memory/constitution.md`
