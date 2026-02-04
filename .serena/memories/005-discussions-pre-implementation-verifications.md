# Pre-Implementation Verifications for Feature 005 (Discussions & Comments)

**Created**: 2026-02-03
**Context**: Verifications to run before `/speckit.implement`

## 1. Database Schema Validation
- Verify `data-model.md` SQL DDL is syntactically correct by dry-running against PostgreSQL
- Check that FK references (users, groups, memberships) match Feature 004's actual schema
- Validate index naming conventions match existing migrations

## 2. Contract Validation
- Lint OpenAPI YAML files in `contracts/` for spec compliance
- Verify all referenced schemas (`$ref`) resolve correctly
- Check for missing error response codes (e.g., 422 for validation errors)

## 3. Dependency Verification
- Confirm Feature 004 migrations actually exist in the main branch
- Verify `memberships` table has the `admin` column referenced by permission checks
- Check that `groups.members_can_start_discussions` flag exists

## 4. Task Dependency Graph Validation
- Verify no circular dependencies in task ordering
- Confirm parallel tasks (`[P]`) truly have no file conflicts
- Validate that test tasks precede implementation tasks in each phase

## 5. Existing Codebase Compatibility
- Check `internal/api/` structure matches plan.md assumptions
- Verify `sqlc.yaml` configuration supports new query files
- Confirm goose migration numbering won't conflict

## 6. Security Review
- Verify all mutation endpoints have authorization checks in tasks
- Check for missing permission scenarios (e.g., archived group handling)
- Validate IDOR protection is addressed for all entity access

## 7. Test Infrastructure
- Verify pgTap is installed and `make test-pgtap` works
- Check that `go test ./...` passes on current codebase
- Confirm testcontainers setup works for integration tests

## Commands for Quick Verification

```bash
# Dependency check
ls migrations/002_create_groups.sql migrations/003_create_memberships.sql

# Check memberships schema for admin column
grep -n "admin" migrations/003_create_memberships.sql

# Check groups schema for permission flag
grep -n "members_can_start_discussions" migrations/002_create_groups.sql

# Verify sqlc config
cat sqlc.yaml

# Check migration numbering
ls -la migrations/*.sql | tail -5

# Test infrastructure
make test-pgtap
go test ./... -v -count=1
```

## Status
- [x] Run before implementation (2026-02-03)
- [ ] All checks pass

## Remediation Completed (2026-02-03)

### Fixed Issues

1. **Migration Numbering Conflict (RESOLVED)**
   - Feature 004 uses migrations 002-005 (audit, groups, memberships, auditing)
   - Updated plan.md and tasks.md: Feature 005 now uses migrations 006-008
   - Updated pgTap test file names to match (006, 007, 008)

2. **Missing 422 Validation Errors (RESOLVED)**
   - Added 422 responses to discussions.yaml (createDiscussion, updateDiscussion)
   - Added 422 responses to comments.yaml (createComment, updateComment)

3. **Reordered HTTP status codes** to follow convention (400, 401, 403, 404, 422)

### Remaining Blocker

**Feature 004 Branch Not Merged**
- Feature 004 migrations exist in `004-groups-memberships` branch (not main)
- Options:
  1. Wait for 004 PR to merge, then rebase 005 on main
  2. Rebase 005 on 004-groups-memberships branch for parallel development
  3. Cherry-pick 004 migrations into 005 branch (not recommended - causes merge conflicts)

**Recommendation**: Rebase 005-discussions on 004-groups-memberships branch to enable parallel development. When 004 merges to main, 005 will cleanly rebase.

### Infrastructure Status

- Go tests: PASS (except pre-existing config test failures documented in CLAUDE.md)
- pgTap: Requires PostgreSQL container running (`make up` first)
- sqlc config: Correctly configured for `internal/db/queries/` directory
