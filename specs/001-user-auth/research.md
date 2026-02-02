# Research: User Authentication

**Feature**: 001-user-auth | **Date**: 2026-02-01

## Technical Decisions

### 1. Password Hashing: Argon2id

**Decision**: Use Argon2id with OWASP-recommended parameters

**Rationale**:
- Argon2id is the winner of the Password Hashing Competition (2015) and combines resistance to both GPU attacks (Argon2d) and side-channel attacks (Argon2i)
- Specified in the feature requirements
- Go stdlib provides `golang.org/x/crypto/argon2` which is well-maintained

**Parameters** (OWASP 2023 recommendations):
```go
// Minimum recommended for web apps
Memory:      64 * 1024  // 64 MiB
Iterations:  3          // Time cost
Parallelism: 4          // Threads
KeyLength:   32         // 256-bit hash
SaltLength:  16         // 128-bit salt (crypto/rand)
```

**Alternatives Considered**:
- bcrypt: Limited to 72-byte passwords, slower to evolve parameters
- scrypt: Good but Argon2id has better side-channel resistance
- PBKDF2: Still acceptable but weaker against modern GPUs

### 2. Session Storage: In-Memory Map

**Decision**: Use sync.Map with session tokens stored as HTTP-only cookies

**Rationale**:
- Spec explicitly states "sessions stored in memory, lost on restart (acceptable for MVP)"
- Constitution Principle V (Simplicity & YAGNI) - don't add Redis complexity until needed
- sync.Map provides concurrent-safe access without external dependencies

**Implementation**:
```go
type SessionStore struct {
    sessions sync.Map // map[sessionToken]Session
}

type Session struct {
    UserID    int64
    CreatedAt time.Time
    ExpiresAt time.Time
    UserAgent string
    IPAddress string
}
```

**Session Token**:
- 32 bytes of crypto/rand, base64url encoded (43 characters)
- Constant-time comparison via subtle.ConstantTimeCompare

**Cookie Settings**:
- `HttpOnly: true` - prevents XSS access
- `Secure: true` - HTTPS only (configurable for dev)
- `SameSite: Lax` - CSRF protection for GET requests
- `Path: /` - available to all routes
- `MaxAge: 7 * 24 * 60 * 60` - 7 days

**Alternatives Considered**:
- Redis: Overkill for MVP; adds infrastructure complexity
- PostgreSQL sessions: Extra queries on every request; in-memory is faster
- JWT: Stateless but harder to invalidate; in-memory allows immediate logout

### 3. Account Enumeration Prevention

**Decision**: Constant-time responses with identical error messages

**Rationale**:
- FR-011 requires preventing account enumeration
- SC-006 requires consistent timing (<3s for invalid attempts)

**Implementation Strategy**:
1. Always perform password hash even if user not found (dummy hash)
2. Return identical "Invalid credentials" message for:
   - User not found
   - Wrong password
   - Unverified email
   - Deactivated account
3. Use subtle.ConstantTimeCompare for token comparisons

**Response Timing**:
```go
// Always hash even if user not found
if user == nil {
    // Hash against dummy to maintain consistent timing
    argon2.IDKey([]byte(password), dummySalt, 3, 64*1024, 4, 32)
    return ErrInvalidCredentials
}
```

### 4. Username Generation

**Decision**: Generate from name with numeric suffix for uniqueness

**Rationale**:
- FR-014 requires unique username generation
- Following existing Loomio pattern observed in schema

**Algorithm**:
```
1. Slugify name: "John Doe" → "john-doe"
2. If taken, append incrementing suffix: "john-doe-1", "john-doe-2"
3. Fallback to email prefix if name produces empty slug
4. Maximum length: 40 characters
```

### 5. Email Case Insensitivity

**Decision**: Store emails lowercase; compare case-insensitively

**Rationale**:
- Edge case in spec: "System treats email as case-insensitive"
- PostgreSQL citext extension used in original schema
- For rewrite: lowercase on input, use LOWER() in queries

**Implementation**:
```sql
-- Migration uses citext or explicit lowercasing
CREATE TABLE users (
    email TEXT NOT NULL UNIQUE,
    -- Constraint: CHECK (email = LOWER(email))
);

-- Query pattern
SELECT * FROM users WHERE LOWER(email) = LOWER($1);
-- Or with index: CREATE INDEX users_email_lower ON users(LOWER(email));
```

### 6. Public URL Key Generation

**Decision**: Use crypto/rand for 128-bit base64url key

**Rationale**:
- FR-015 requires unique public URL key
- Matches pattern from original schema (`key` column)
- Used for public profile URLs without exposing internal IDs

**Implementation**:
```go
// 16 bytes = 128 bits, base64url = 22 characters
key := make([]byte, 16)
crypto/rand.Read(key)
return base64.RawURLEncoding.EncodeToString(key)
```

## Clarifications Resolved

### Q1: Session vs JWT Token?

**Answer**: In-memory sessions with HTTP-only cookies

Per spec: "Sessions are stored in memory and will be lost on server restart (acceptable for MVP)"

This means traditional sessions, not JWTs. Benefits:
- Immediate logout capability (delete from map)
- No token refresh complexity
- Simpler security model

### Q2: How to handle email verification status?

**Answer**: Manual verification for MVP, boolean flag in database

Per spec assumption: "users will be manually marked as verified for MVP"

- `email_verified BOOLEAN DEFAULT FALSE`
- Admin can set to TRUE via database or admin endpoint (future feature)
- Login checks this flag per FR-012

### Q3: What response format for login success?

**Answer**: Follow existing Loomio pattern from OpenAPI discovery

From `discovery/openapi/paths/auth.yaml`:
```yaml
responses:
  '200':
    content:
      application/json:
        schema:
          properties:
            users: [CurrentUser]
            # For MVP, return just the user
```

Simplified for MVP:
```json
{
  "user": {
    "id": 123,
    "email": "user@example.com",
    "name": "John Doe",
    "username": "john-doe",
    "email_verified": true,
    "created_at": "2026-02-01T12:00:00Z"
  }
}
```

### Q4: What specific Huma patterns to follow?

**Answer**: Standard Huma operations with typed request/response

```go
type LoginRequest struct {
    Body struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=1"`
    }
}

type LoginResponse struct {
    Body struct {
        User UserDTO `json:"user"`
    }
    SetCookie http.Cookie
}

func (api *AuthAPI) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // Implementation
}
```

### Q5: Thread safety for session store?

**Answer**: sync.Map with cleanup goroutine

```go
// Background cleanup every hour
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        store.sessions.Range(func(key, value any) bool {
            session := value.(Session)
            if time.Now().After(session.ExpiresAt) {
                store.sessions.Delete(key)
            }
            return true
        })
    }
}()
```

## Security Considerations

### Password Storage

1. **Never log passwords** - not in errors, not in request logs
2. **Salt per password** - 16 bytes from crypto/rand
3. **Hash on every attempt** - prevents timing attacks
4. **No password hints** - not storing any reversible info

### Session Security

1. **Token entropy** - 256 bits (32 bytes) sufficient for unguessable
2. **Secure cookie flags** - HttpOnly, Secure, SameSite
3. **Session binding** - Store IP/UserAgent for audit (not enforcement in MVP)
4. **Logout invalidation** - Delete from store immediately

### Database Security

1. **Parameterized queries** - sqlc enforces this
2. **No email enumeration** - registration returns success even if email exists?
   - Actually: return error for duplicate (spec requires it) but identical timing

## Dependencies

| Package | Purpose | Version |
|---------|---------|---------|
| `golang.org/x/crypto/argon2` | Password hashing | Latest |
| `github.com/danielgtaylor/huma/v2` | HTTP framework | v2.x |
| `github.com/jackc/pgx/v5` | PostgreSQL driver | v5.x |
| `github.com/sqlc-dev/sqlc` | Query generation | v1.x |
| `github.com/pressly/goose/v3` | Migrations | v3.x |

## Test Strategy

### Unit Tests
- Password hashing/verification
- Username generation
- Session token generation
- Session expiry logic

### Integration Tests
- User creation with unique email constraint
- User lookup by email (case insensitive)
- Full login flow with database

### API Tests (via httptest)
- Registration success/failure scenarios
- Login success/failure scenarios
- Logout with session invalidation
- Protected endpoint access
