# Research: Groups & Memberships

**Feature**: 004-groups-memberships
**Date**: 2026-02-02
**Status**: Complete

## Research Tasks

| Topic | Status | Decision |
|-------|--------|----------|
| Audit logging pattern (supa_audit) | ✅ Resolved | Use supi_audit pattern with xact_id for transaction correlation |
| PostgreSQL session variables for actor context | ✅ Resolved | SET LOCAL app.current_user_id at transaction start |
| Go/pgx transaction patterns | ✅ Resolved | pgx.BeginTxFunc for automatic commit/rollback |
| Existing codebase patterns | ✅ Resolved | Follow auth.go handler structure, DTO patterns |

---

## 1. Audit Logging Pattern

### Decision: supi_audit (fork of supa_audit)

**Source**: https://github.com/patte/supi_audit (fork of supabase/supa_audit)

**Why supi_audit over supa_audit**:
1. **Simple migration** - No PostgreSQL extension required, just SQL migrations
2. **Transaction correlation** - `xact_id` column tracks which changes occurred in the same transaction
3. **Simpler record_id** - Uses single-column UUID PKs directly instead of deriving UUID v5

### audit.record_version Schema

```sql
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TYPE audit.operation AS ENUM ('INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'SNAPSHOT');

CREATE TABLE audit.record_version (
    id              BIGSERIAL PRIMARY KEY,
    record_id       UUID,                    -- Primary key of the record (for UUID PKs)
    old_record_id   UUID,                    -- Previous record_id (for UPDATE/DELETE)
    op              audit.operation NOT NULL,
    ts              TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    xact_id         BIGINT NOT NULL DEFAULT txid_current(), -- Transaction correlation
    table_oid       OID NOT NULL,
    table_schema    NAME NOT NULL,
    table_name      NAME NOT NULL,
    record          JSONB,                   -- New state (INSERT/UPDATE)
    old_record      JSONB,                   -- Old state (UPDATE/DELETE)
    actor_id        BIGINT,                  -- User who made the change (from session var)

    -- Constraints
    CHECK (COALESCE(record_id, old_record_id) IS NOT NULL OR op IN ('TRUNCATE', 'SNAPSHOT')),
    CHECK (op IN ('INSERT', 'UPDATE', 'SNAPSHOT') = (record IS NOT NULL)),
    CHECK (op IN ('UPDATE', 'DELETE') = (old_record IS NOT NULL))
);

-- Indexes
CREATE INDEX record_version_record_id ON audit.record_version(record_id) WHERE record_id IS NOT NULL;
CREATE INDEX record_version_ts ON audit.record_version USING BRIN(ts);
CREATE INDEX record_version_table_oid ON audit.record_version(table_oid);
CREATE INDEX record_version_xact_id ON audit.record_version(xact_id);
CREATE INDEX record_version_actor_id ON audit.record_version(actor_id) WHERE actor_id IS NOT NULL;
```

**Key additions from supi_audit**:
- `xact_id` - Correlates all changes in a single transaction
- `actor_id` - Our custom column for user tracking (vs auth_uid which requires Supabase auth)
- BRIN index on `ts` - 99% smaller than B-tree for append-only audit logs

### Trigger Function Pattern

```sql
CREATE OR REPLACE FUNCTION audit.insert_update_delete_trigger()
RETURNS TRIGGER
SECURITY DEFINER
LANGUAGE plpgsql
AS $$
DECLARE
    pkey_cols TEXT[] := audit.primary_key_columns(TG_RELID);
    record_jsonb JSONB := to_jsonb(NEW);
    old_record_jsonb JSONB := to_jsonb(OLD);
    v_actor_id BIGINT;
BEGIN
    -- Get actor from session variable (NULL if not set)
    v_actor_id := NULLIF(current_setting('app.current_user_id', true), '')::BIGINT;

    INSERT INTO audit.record_version (
        record_id,
        old_record_id,
        op,
        table_oid,
        table_schema,
        table_name,
        record,
        old_record,
        actor_id
    ) VALUES (
        CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN (NEW.id)::TEXT::UUID END,
        CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN (OLD.id)::TEXT::UUID END,
        TG_OP::audit.operation,
        TG_RELID,
        TG_TABLE_SCHEMA,
        TG_TABLE_NAME,
        CASE WHEN TG_OP != 'DELETE' THEN record_jsonb END,
        CASE WHEN TG_OP != 'INSERT' THEN old_record_jsonb END,
        v_actor_id
    );

    RETURN COALESCE(NEW, OLD);
END;
$$;
```

**Note**: For BIGINT primary keys (our case), we skip UUID v5 derivation and use the ID directly cast to UUID where needed, or store as BIGINT in a separate column. Since our `groups.id` and `memberships.id` are BIGSERIAL, we'll adapt the pattern.

### Alternatives Rejected

| Alternative | Why Rejected |
|-------------|--------------|
| Application-layer logging | TC-001 requires triggers; app bugs could skip audit |
| supa_audit extension | Requires extension installation; supi_audit is simpler |
| pgaudit extension | Logs SQL statements, not row-level JSONB snapshots |

---

## 2. PostgreSQL Session Variables

### Decision: SET LOCAL app.current_user_id

**Pattern**: Set at transaction start, automatically reset on commit/rollback

### Implementation

**In trigger function**:
```sql
-- Get actor with graceful fallback (NULL if not set)
v_actor_id := NULLIF(current_setting('app.current_user_id', true), '')::BIGINT;
```

**In Go handler** (before mutations):
```go
return pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
    // Set actor context FIRST
    _, err := tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", userID)
    if err != nil {
        return err
    }

    // Execute mutations - triggers will capture actor_id
    group, err := queries.WithTx(tx).CreateGroup(ctx, params)
    if err != nil {
        return err
    }

    // Automatic commit on return nil
    return nil
})
```

### Why SET LOCAL

| Feature | SET | SET LOCAL |
|---------|-----|-----------|
| Scope | Session | Transaction only |
| Connection pool safety | ❌ Leaks to next request | ✅ Auto-resets |
| Requires explicit cleanup | Yes | No |

---

## 3. Go/pgx Transaction Patterns

### Decision: pgx.BeginTxFunc with queries.WithTx(tx)

**Pattern from existing codebase** (internal/db/db.go:28-32):
```go
// WithTx creates a new Queries instance using a transaction
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
    return &Queries{db: tx}
}
```

### Full Transaction Pattern

```go
func (h *GroupHandler) handleCreateGroup(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
    // 1. Authenticate user (existing pattern from auth.go)
    session, found := h.sessions.Get(input.Cookie)
    if !found {
        return nil, huma.Error401Unauthorized("Not authenticated")
    }

    // 2. Execute in transaction
    var result *db.Group
    err := pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
        // Set audit context
        _, err := tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", session.UserID)
        if err != nil {
            return err
        }

        txQueries := h.queries.WithTx(tx)

        // Create group
        result, err = txQueries.CreateGroup(ctx, db.CreateGroupParams{
            Name:        input.Body.Name,
            Handle:      generateHandle(input.Body.Name),
            CreatedByID: session.UserID,
        })
        if err != nil {
            return err
        }

        // Create admin membership
        _, err = txQueries.CreateMembership(ctx, db.CreateMembershipParams{
            GroupID:   result.ID,
            UserID:    session.UserID,
            Role:      "admin",
            InviterID: session.UserID,
        })
        return err
    })

    if err != nil {
        LogDBError(ctx, "CreateGroup", err)
        return nil, huma.Error500InternalServerError("Database error")
    }

    return &CreateGroupOutput{Body: struct{ Group GroupDTO }{Group: GroupDTOFromGroup(result)}}, nil
}
```

---

## 4. Existing Codebase Patterns

### Handler Structure (internal/api/auth.go)

```go
type GroupHandler struct {
    pool     *pgxpool.Pool  // For transactions
    queries  *db.Queries    // For read operations
    sessions *auth.SessionStore
}

func NewGroupHandler(pool *pgxpool.Pool, queries *db.Queries, sessions *auth.SessionStore) *GroupHandler {
    return &GroupHandler{pool: pool, queries: queries, sessions: sessions}
}

func (h *GroupHandler) RegisterRoutes(api huma.API) {
    huma.Register(api, huma.Operation{
        OperationID:   "createGroup",
        Method:        http.MethodPost,
        Path:          "/api/v1/groups",
        Tags:          []string{"Groups"},
        DefaultStatus: http.StatusCreated,
    }, h.handleCreateGroup)
}
```

### DTO Pattern (internal/api/dto.go)

```go
type GroupDTO struct {
    ID          int64      `json:"id"`
    Name        string     `json:"name"`
    Handle      string     `json:"handle"`
    Description string     `json:"description,omitempty"`
    ParentID    *int64     `json:"parent_id,omitempty"`
    ArchivedAt  *time.Time `json:"archived_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
}

func GroupDTOFromGroup(g *db.Group) GroupDTO {
    dto := GroupDTO{
        ID:          g.ID,
        Name:        g.Name,
        Handle:      g.Handle,
        Description: g.Description,
        CreatedAt:   g.CreatedAt.Time,
    }
    if g.ParentID.Valid {
        dto.ParentID = &g.ParentID.Int64
    }
    if g.ArchivedAt.Valid {
        dto.ArchivedAt = &g.ArchivedAt.Time
    }
    return dto
}
```

### Error Handling Pattern

```go
// Check for specific error types
if err != nil {
    if db.IsNotFound(err) {
        return nil, huma.Error404NotFound("Group not found")
    }
    LogDBError(ctx, "GetGroupByID", err)
    return nil, huma.Error500InternalServerError("Database error")
}

// Handle unique constraint violations
if isUniqueViolation(err, "groups_handle_key") {
    return nil, huma.Error409Conflict("Handle already taken")
}
```

---

## 5. BRIN vs B-tree Index Decision

### Decision: BRIN for timestamps, B-tree for lookups

| Column | Index Type | Rationale |
|--------|------------|-----------|
| `ts` | BRIN | Append-only, 99% smaller than B-tree |
| `record_id` | B-tree (partial) | Point lookups for history |
| `table_oid` | B-tree | Filter by table |
| `actor_id` | B-tree (partial) | Filter by user |
| `xact_id` | B-tree | Correlate transaction changes |

**BRIN effectiveness check**:
```sql
-- Verify high correlation (should be > 0.9 for audit tables)
SELECT correlation FROM pg_stats
WHERE tablename = 'record_version' AND attname = 'ts';
```

---

## 6. Last-Admin Protection Trigger

### Decision: BEFORE UPDATE/DELETE trigger on memberships

```sql
CREATE OR REPLACE FUNCTION prevent_last_admin_removal()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    admin_count INTEGER;
BEGIN
    -- Only check for admin role changes or deletions
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND OLD.role = 'admin' AND NEW.role != 'admin') THEN
        SELECT COUNT(*) INTO admin_count
        FROM memberships
        WHERE group_id = OLD.group_id
          AND role = 'admin'
          AND accepted_at IS NOT NULL
          AND id != OLD.id;  -- Exclude the record being modified

        IF admin_count = 0 THEN
            RAISE EXCEPTION 'Cannot remove or demote the last administrator of a group';
        END IF;
    END IF;

    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

CREATE TRIGGER memberships_last_admin_protection
    BEFORE UPDATE OR DELETE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION prevent_last_admin_removal();
```

### Why Database Trigger

| Approach | Pros | Cons |
|----------|------|------|
| App-layer check | Easier to test | Race conditions possible |
| Database trigger | Atomic, no races | Harder to test |

**Decision**: TC-005 requires DB trigger. Use pgTap for testing.

---

## Sources

- [supi_audit (fork)](https://github.com/patte/supi_audit) - Simplified audit pattern with xact_id
- [supa_audit (original)](https://github.com/supabase/supa_audit) - JSONB audit logging
- [Postgres Auditing in 150 lines of SQL](https://supabase.com/blog/postgres-audit)
- [BRIN Index Performance](https://pganalyze.com/blog/5mins-postgres-BRIN-index)
- [PostgreSQL SET command](https://www.postgresql.org/docs/current/sql-set.html)
- [pgx v5 transactions](https://pkg.go.dev/github.com/jackc/pgx/v5)
