# PostgreSQL Security Architecture Design

**Status**: Brainstorming complete - ready for user approval
**Scope**: Roles, RLS, schema namespaces, GDPR compliance for Loomio rewrite

## Summary

Add a comprehensive PostgreSQL security architecture to the Loomio rewrite with:

1. **Three schemas** for access tier isolation
2. **Service-specific roles** with principle of least privilege
3. **Row-Level Security** for Organization-based multi-tenancy
4. **GDPR-compliant** user deletion via privileged purge functions

This is a **significant infrastructure addition** that should be implemented as a separate feature (e.g., `007-postgres-security`) after the current core domain features (004-006).

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Tenant boundary | Organization (new entity) | Above Groups; users can belong to multiple Orgs (federated) |
| Schema strategy | Access tiers | `loomio_public`, `loomio_private`, `loomio_internal` |
| Role model | Service-specific | `llmio_server`, `llmio_river_worker`, `llmio_migrater`, `llmio_admin` |
| RLS context | App-set session vars | `app.current_user_id`, `app.current_org_id` |
| Admin bypass | BYPASSRLS role | `llmio_admin` role bypasses RLS for support/admin tools |
| Delete policy | Table-specific | Soft-delete for content; hard-delete for memberships |
| GDPR handling | Anonymization | `gdpr_purge_user()` function anonymizes content, deletes user |

---

## Schema Structure

```
loomio_public/          -- API-exposed tables (RLS-protected)
├── organizations
├── organization_memberships
├── users
├── groups
├── memberships
├── discussions
├── comments
├── discussion_readers
└── events

loomio_private/         -- App-internal (no RLS, trusted code only)
├── sessions
├── jobs (River queue)
└── email_queue

loomio_internal/        -- Admin/migrations only
├── audit_log
├── schema_migrations
└── gdpr_purge_user() function
```

---

## Role Hierarchy

```sql
-- Base roles (nologin, privilege grouping)
loomio_public_reader    -- SELECT on loomio_public
loomio_public_writer    -- INSERT/UPDATE on loomio_public
loomio_delete_memberships -- DELETE on specific tables only
loomio_private_reader   -- SELECT on loomio_private
loomio_private_writer   -- INSERT/UPDATE/DELETE on loomio_private
loomio_internal_admin   -- DDL on loomio_internal

-- Login roles (actual connections)
llmio_server            -- public RW + private RW + delete memberships
llmio_river_worker      -- private RW + public RO
llmio_migrater          -- internal admin (DDL only, no data)
llmio_analytics_ro      -- public RO + BYPASSRLS
llmio_admin             -- SUPERUSER + BYPASSRLS
```

---

## Organization Model (New)

```sql
CREATE TABLE loomio_public.organizations (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name            TEXT NOT NULL,
    handle          CITEXT NOT NULL UNIQUE,
    archived_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loomio_public.organization_memberships (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES loomio_public.organizations(id),
    user_id         BIGINT NOT NULL REFERENCES loomio_public.users(id),
    role            TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, user_id)
);

-- Groups now require an Organization
ALTER TABLE loomio_public.groups
  ADD COLUMN org_id BIGINT NOT NULL REFERENCES loomio_public.organizations(id);
```

**Personal content**: Users get a "personal organization" auto-created for direct discussions.

---

## RLS Policies

```sql
-- Enable and force RLS on all public tables
ALTER TABLE loomio_public.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE loomio_public.organizations FORCE ROW LEVEL SECURITY;
-- (repeat for all tables)

-- Organization policy: user must be a member
CREATE POLICY org_member_policy ON loomio_public.organizations
  FOR ALL
  USING (
    id IN (
      SELECT organization_id
      FROM loomio_public.organization_memberships
      WHERE user_id = (SELECT current_setting('app.current_user_id')::bigint)
    )
  );

-- Groups policy: org context must match
CREATE POLICY groups_org_policy ON loomio_public.groups
  FOR ALL
  USING (
    org_id = (SELECT current_setting('app.current_org_id')::bigint)
  );

-- Discussions inherit from group
CREATE POLICY discussions_org_policy ON loomio_public.discussions
  FOR ALL
  USING (
    group_id IN (
      SELECT id FROM loomio_public.groups
      WHERE org_id = (SELECT current_setting('app.current_org_id')::bigint)
    )
  );
```

**Go integration:**
```go
func SetRequestContext(ctx context.Context, conn *pgx.Conn, userID, orgID int64) error {
    _, err := conn.Exec(ctx, `
        SELECT set_config('app.current_user_id', $1::text, true);
        SELECT set_config('app.current_org_id', $2::text, true);
    `, userID, orgID)
    return err
}
```

---

## Delete Privilege Matrix

| Table | Normal DELETE | GDPR Purge | Notes |
|-------|---------------|------------|-------|
| `users` | ❌ | ✅ hard-delete | Via `gdpr_purge_user()` |
| `organizations` | ❌ | N/A | Soft-delete via `archived_at` |
| `organization_memberships` | ✅ | ✅ | Can leave/be removed |
| `groups` | ❌ | N/A | Soft-delete via `archived_at` |
| `memberships` | ✅ | ✅ | Can leave group |
| `discussions` | ❌ | Anonymized | `author_id=NULL`, description cleared |
| `comments` | ❌ | Anonymized | `author_id=NULL`, `body='[deleted]'` |
| `events` | ❌ | `user_id=NULL` | Append-only, immutable |
| `discussion_readers` | ✅ | ✅ | Read state resetable |
| `sessions` (private) | ✅ | ✅ | Logout = delete |
| `jobs` (private) | ✅ | N/A | Completed jobs purged |

---

## GDPR Purge Function

```sql
CREATE OR REPLACE FUNCTION loomio_internal.gdpr_purge_user(p_user_id BIGINT)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
  -- Audit the purge request
  INSERT INTO loomio_internal.audit_log (action, target_type, target_id, performed_by, performed_at)
  VALUES ('gdpr_purge', 'user', p_user_id, current_setting('app.admin_user_id')::bigint, NOW());

  -- Anonymize content (preserve thread structure)
  UPDATE loomio_public.comments
  SET body = '[deleted]', author_id = NULL
  WHERE author_id = p_user_id;

  UPDATE loomio_public.discussions
  SET description = '[deleted]'
  WHERE author_id = p_user_id;

  -- Anonymize events
  UPDATE loomio_public.events
  SET user_id = NULL
  WHERE user_id = p_user_id;

  -- Remove memberships
  DELETE FROM loomio_public.memberships WHERE user_id = p_user_id;
  DELETE FROM loomio_public.organization_memberships WHERE user_id = p_user_id;

  -- Purge sessions
  DELETE FROM loomio_private.sessions WHERE user_id = p_user_id;

  -- Delete user record
  DELETE FROM loomio_public.users WHERE id = p_user_id;
END;
$$;

REVOKE ALL ON FUNCTION loomio_internal.gdpr_purge_user FROM PUBLIC;
GRANT EXECUTE ON FUNCTION loomio_internal.gdpr_purge_user TO llmio_admin;
```

---

## Impact on Existing Features

### Feature 004 (Groups & Memberships)

**Changes required:**
1. Add `organizations` and `organization_memberships` tables
2. Add `org_id` column to `groups` table
3. Move tables to `loomio_public` schema
4. Add RLS policies

### Feature 001 (User Auth)

**Changes required:**
1. Move `users` table to `loomio_public` schema
2. Move `sessions` to `loomio_private` schema (or keep in-memory for MVP)
3. Add RLS policy for users table
4. Update `SetRequestContext()` call in auth middleware

### Migration Strategy

Since this is a foundational change, implement as **Feature 007** after completing 004-006:
1. Create schemas and roles (migration)
2. Migrate existing tables to new schemas
3. Add RLS policies
4. Update Go code to set context
5. Test thoroughly with permission matrix

---

## Files to Create/Modify

**New migrations:**
- `migrations/007_create_schemas.sql` - Schema creation
- `migrations/008_create_roles.sql` - Role hierarchy
- `migrations/009_create_organizations.sql` - Organization tables
- `migrations/010_enable_rls.sql` - RLS policies
- `migrations/011_gdpr_functions.sql` - Purge functions

**Go code changes:**
- `internal/db/context.go` - `SetRequestContext()` function
- `internal/api/middleware/rls.go` - Middleware to set context per-request
- `internal/config/database.go` - Multiple connection configs per role

**Configuration:**
- `.env.example` - Add role-specific DB credentials
- `docker-compose.yml` - Database initialization with roles

---

## Verification

1. **Schema isolation test**: Connect as `llmio_server`, verify cannot access `loomio_internal`
2. **RLS test**: Query as user A, verify cannot see user B's org data
3. **Delete privilege test**: Attempt DELETE on `events` table, verify denied
4. **GDPR purge test**: Run purge function, verify user data anonymized/deleted
5. **Bypass test**: Connect as `llmio_admin`, verify RLS bypassed

---

## Constitution Compliance

| Principle | Status | Notes |
|-----------|--------|-------|
| Test-First (I) | ✅ | pgTap tests for schema/roles/RLS |
| Huma-First (II) | ✅ | No API changes |
| Security-First (III) | ✅ | This IS the security enhancement |
| Type Safety (IV) | ✅ | sqlc regenerated for new schemas |
| Simplicity (V) | ⚠️ | Adds complexity, but justified for security |

**Complexity justification**: This adds significant infrastructure but provides defense-in-depth that prevents catastrophic data leaks if Go code has authorization bugs.

---

## Resolved Decisions

| Question | Decision |
|----------|----------|
| Personal org creation | At registration (every user gets personal org immediately) |
| Implementation timing | Feature 007 after completing 004-006 |
| Cross-org users | Yes, user can have different roles in different orgs |

## Open Questions (for Feature 007 planning)

1. **Org switching UI**: How does frontend handle org context switching?
2. **Analytics bypass**: Should `llmio_analytics_ro` see ALL orgs or require explicit access grants?
3. **Personal org naming**: What's the handle format for personal orgs? (e.g., `~username` or `personal-{user_id}`)
