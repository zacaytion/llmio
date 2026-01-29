# ADR-003: Go Technology Stack Selection

**Date:** 2026-01-29
**Status:** Proposed

## Context

Need to select Go ecosystem components for the Loomio backend rewrite. Selection criteria:
1. **Maturity:** Production-proven, active maintenance
2. **Performance:** Suitable for concurrent API workload
3. **Developer experience:** Reasonable learning curve
4. **Ecosystem:** Good middleware and tooling

## Decisions

### Web Framework: Chi

**Selected:** `github.com/go-chi/chi/v5`

**Rationale:**
- Lightweight, composable middleware
- stdlib `net/http` compatible
- No magic, explicit routing
- Battle-tested at scale

**Alternatives considered:**
- Gin: More opinionated, larger dependency tree
- Echo: Similar to Chi, less stdlib-aligned
- Fiber: fasthttp-based, not net/http compatible

### Database: sqlc + pgx

**Selected:** `github.com/sqlc-dev/sqlc` with `github.com/jackc/pgx/v5`

**Rationale:**
- Type-safe SQL, generated Go code
- Full PostgreSQL feature support (JSONB, arrays, citext)
- No ORM magic, explicit queries
- pgx is fastest PostgreSQL driver

**Alternatives considered:**
- GORM: Magic-heavy, harder to debug complex queries
- sqlx: Good but manual, sqlc generates code
- ent: Schema-first, larger learning curve

### Authentication: Custom JWT + bcrypt

**Selected:** Custom implementation with:
- `github.com/golang-jwt/jwt/v5` for tokens
- `golang.org/x/crypto/bcrypt` for passwords

**Rationale:**
- Loomio has specific auth flows (magic links, SAML, OAuth)
- Custom allows exact feature parity
- Simpler than adapting full auth framework

### Background Jobs: River

**Selected:** `github.com/riverqueue/river`

**Rationale:**
- PostgreSQL-backed (no Redis required for jobs)
- Transactional job enqueuing
- Type-safe job definitions
- Active development, good docs

**Alternatives considered:**
- Asynq: Redis-backed, adds dependency
- Machinery: Heavier, more complex

### Testing: testify

**Selected:** `github.com/stretchr/testify`

**Rationale:**
- De facto standard for assertions
- Suite support for setup/teardown
- Mock generation
- Familiar to most Go developers

## Summary Table

| Category | Selection | Package |
|----------|-----------|---------|
| Web Framework | Chi | github.com/go-chi/chi/v5 |
| Database | sqlc + pgx | github.com/sqlc-dev/sqlc |
| Auth | Custom JWT | github.com/golang-jwt/jwt/v5 |
| Background Jobs | River | github.com/riverqueue/river |
| Testing | testify | github.com/stretchr/testify |
| WebSocket | websocket | nhooyr.io/websocket |
| Redis | go-redis | github.com/redis/go-redis/v9 |
| Logging | slog | stdlib (log/slog) |
| Config | env | github.com/caarlos0/env/v10 |

## Consequences

**Positive:**
- Minimal dependencies, stdlib-aligned
- Type-safe database access
- PostgreSQL-native job queue reduces infrastructure

**Negative:**
- More boilerplate than full frameworks
- sqlc requires writing SQL (learning curve for Rails developers)
