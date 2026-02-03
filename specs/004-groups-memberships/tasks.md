# Tasks: Groups & Memberships

**Input**: Design documents from `/specs/004-groups-memberships/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/groups.yaml

**Tests**: Tests are included as this is a TDD project (per constitution principle I: Test-First Development). pgTap tests for database triggers/constraints, Go table-driven tests for API.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This project uses the Go internal package structure:
- **API handlers**: `internal/api/`
- **Database queries**: `queries/` (sqlc)
- **Database models**: `internal/db/` (sqlc-generated)
- **Migrations**: `migrations/`
- **Database tests**: `tests/pgtap/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Audit infrastructure and core schema that ALL user stories depend on

### Test Infrastructure (testcontainers-go)

- [ ] T000a Add testcontainers-go dependency: `go get github.com/testcontainers/testcontainers-go`
- [ ] T000b Add testcontainers postgres module: `go get github.com/testcontainers/testcontainers-go/modules/postgres`
- [ ] T000c [P] Create internal/testutil/postgres.go with PostgresContainer helper
- [ ] T000d [P] Create internal/testutil/pgtap.go with RunPgTapTests helper for isolated pgTap execution

### Database Migrations

- [x] T001 Create audit schema migration in migrations/002_create_audit_schema.sql
- [x] T002 Write pgTap tests for audit schema in tests/pgtap/002_audit_schema_test.sql (depends on T001)
- [x] T003 Create groups table migration in migrations/003_create_groups.sql
- [x] T004 Write pgTap tests for groups constraints in tests/pgtap/003_groups_test.sql (depends on T003)
- [x] T005 Create memberships table migration in migrations/004_create_memberships.sql
- [x] T006 Write pgTap tests for memberships last-admin trigger in tests/pgtap/004_memberships_test.sql (depends on T005)
- [x] T007 Create audit trigger enablement migration in migrations/005_enable_auditing.sql
- [x] T008 Run migrations and verify all pgTap tests pass

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T009 Create sqlc queries for groups in queries/groups.sql
- [x] T010 [P] Create sqlc queries for memberships in queries/memberships.sql
- [x] T011 Run sqlc generate to create internal/db/groups.sql.go and internal/db/memberships.sql.go
- [x] T012 Create GroupDTO and MembershipDTO in internal/api/dto.go
- [x] T013 [P] Create handle generation utility function in internal/api/groups.go
- [x] T014 [P] Create transaction helper for audit context (SET LOCAL) in internal/db/audit.go
- [x] T015 Create authorization helpers (isAdmin, canInvite) in internal/api/authorization.go
- [ ] T015a [P] Implement handleListMemberships handler in internal/api/memberships.go
- [ ] T015b [P] Implement handleGetMembership handler in internal/api/memberships.go
- [ ] T015c [P] Register GET /api/v1/groups/{id}/memberships route
- [ ] T015d [P] Register GET /api/v1/memberships/{id} route

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Create a Group (Priority: P1)

**Goal**: A user creates a group with name/description and becomes its first admin

**Independent Test**: Create a group via POST /api/v1/groups → verify group created and creator is admin

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T016 [US1] Write table-driven tests for createGroup handler in internal/api/groups_test.go
- [ ] T017 [US1] Test: authenticated user creates group → 201 + group returned with auto-generated handle
- [ ] T018 [US1] Test: authenticated user creates group with custom handle → 201 + handle preserved
- [ ] T019 [US1] Test: handle conflict → 409 Conflict
- [ ] T020 [US1] Test: unauthenticated → 401 Unauthorized
- [ ] T020a [US1] Test: handle auto-generated from name with spaces → "my group" becomes "my-group"
- [ ] T020b [US1] Test: handle auto-generated from name with special chars → "Team @#$% 2026" becomes "team-2026"
- [ ] T020c [US1] Test: handle auto-generated collision retry → if "climate-team" exists, try "climate-team-1", "climate-team-2", etc.
- [ ] T020d [US1] Test: empty name rejected with 422 (handle generation requires name)

### Implementation for User Story 1

- [ ] T021 [US1] Create GroupHandler struct in internal/api/groups.go
- [ ] T022 [US1] Implement NewGroupHandler constructor in internal/api/groups.go
- [ ] T023 [US1] Implement handleCreateGroup handler in internal/api/groups.go
- [ ] T024 [US1] Handle auto-generation of handle from name (slugify)
- [ ] T025 [US1] Create admin membership for creator in same transaction
- [ ] T026 [US1] Set audit context (app.current_user_id) before mutations
- [ ] T027 [US1] Handle unique constraint violation (handle taken) → 409
- [ ] T028 [US1] Register POST /api/v1/groups route in internal/api/groups.go
- [ ] T029 [US1] Register GroupHandler in cmd/server/main.go
- [ ] T030 [US1] Run tests and verify all pass

**Checkpoint**: User Story 1 complete - users can create groups and become admins

---

## Phase 4: User Story 2 - Invite Members to a Group (Priority: P1)

**Goal**: An admin invites users to join a group; invited users can accept

**Independent Test**: Admin invites user → user sees invitation → user accepts → becomes active member

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T031 [US2] Write table-driven tests for inviteMember handler in internal/api/memberships_test.go
- [ ] T032 [US2] Test: admin invites user → 201 + pending membership created
- [ ] T033 [US2] Test: non-admin without permission → 403 Forbidden
- [ ] T034 [US2] Test: invite already-member → 409 Conflict
- [ ] T035 [US2] Test: acceptInvitation → membership.accepted_at set
- [ ] T036 [US2] Test: accept non-existent invitation → 404 Not Found
- [ ] T036a [US2] Test: invite non-existent user → 404 Not Found
- [ ] T037 [US2] Test: accept someone else's invitation → 403 Forbidden

### Implementation for User Story 2

- [ ] T038 [US2] Create MembershipHandler struct in internal/api/memberships.go
- [ ] T039 [US2] Implement NewMembershipHandler constructor in internal/api/memberships.go
- [ ] T040 [US2] Implement handleInviteMember handler in internal/api/memberships.go
- [ ] T041 [US2] Check inviter authorization (admin-only for now; members_can_add_members flag check added in T076/US4)
- [ ] T042 [US2] Handle duplicate membership constraint → 409
- [ ] T043 [US2] Implement handleAcceptInvitation handler in internal/api/memberships.go
- [ ] T044 [US2] Verify current user is the invited user before accepting
- [ ] T045 [US2] Implement handleListMyInvitations handler in internal/api/memberships.go
- [ ] T046 [US2] Register POST /api/v1/groups/{id}/memberships route
- [ ] T047 [US2] Register POST /api/v1/memberships/{id}/accept route
- [ ] T048 [US2] Register GET /api/v1/users/me/invitations route
- [ ] T049 [US2] Register MembershipHandler in cmd/server/main.go
- [ ] T050 [US2] Run tests and verify all pass

**Checkpoint**: User Story 2 complete - admins can invite members, users can accept invitations

---

## Phase 5: User Story 3 - Manage Group Members (Priority: P2)

**Goal**: Admins can promote/demote/remove members; last-admin protection enforced

**Independent Test**: Admin promotes member → member becomes admin; demote last admin → 409 blocked

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T051 [US3] Write table-driven tests for promoteMember handler in internal/api/memberships_test.go
- [ ] T052 [US3] Test: admin promotes member → role becomes admin
- [ ] T053 [US3] Test: non-admin tries to promote → 403 Forbidden
- [ ] T054 [US3] Test: admin demotes other admin → role becomes member (verify membership still exists, only role changed)
- [ ] T055 [US3] Test: demote last admin → 409 Conflict (DB trigger enforced)
- [ ] T056 [US3] Test: remove member → membership deleted
- [ ] T057 [US3] Test: remove last admin → 409 Conflict

### Implementation for User Story 3

- [ ] T058 [US3] Implement handlePromoteMember handler in internal/api/memberships.go
- [ ] T059 [US3] Implement handleDemoteMember handler in internal/api/memberships.go
- [ ] T060 [US3] Handle last-admin trigger error → 409 Conflict with clear message
- [ ] T061 [US3] Implement handleRemoveMember (DELETE) handler in internal/api/memberships.go
- [ ] T062 [US3] Register POST /api/v1/memberships/{id}/promote route
- [ ] T063 [US3] Register POST /api/v1/memberships/{id}/demote route
- [ ] T064 [US3] Register DELETE /api/v1/memberships/{id} route
- [ ] T065 [US3] Run tests and verify all pass

**Checkpoint**: User Story 3 complete - full member management with last-admin protection

---

## Phase 6: User Story 4 - Configure Group Permissions (Priority: P2)

**Goal**: Admins configure 11 permission flags; enforcement happens immediately

**Scope Note**: This feature enforces `members_can_add_members` (T076) and `members_can_create_subgroups` (T087). The remaining 9 flags are stored and configurable but enforcement is deferred:

| Flag | Deferred To | Reason |
|------|-------------|--------|
| `members_can_add_guests` | Feature 005 | Guests are discussion-level access |
| `members_can_start_discussions` | Feature 005 | Requires Discussion entity |
| `members_can_raise_motions` | Feature 006 | Requires Poll entity |
| `members_can_edit_discussions` | Feature 005 | Requires Discussion entity |
| `members_can_edit_comments` | Feature 005 | Requires Comment entity |
| `members_can_delete_comments` | Feature 005 | Requires Comment entity |
| `members_can_announce` | Feature 005 | Requires notification system |
| `admins_can_edit_user_content` | Feature 005 | Requires content ownership model |
| `parent_members_can_see_discussions` | Feature 005 | Requires Discussion entity |

**Verification**: T071 tests that getGroup returns all 11 flags; enforcement tests for deferred flags will be added in their respective features.

**Independent Test**: Admin toggles members_can_add_members → member invite succeeds/fails accordingly

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T066 [US4] Write table-driven tests for updateGroup handler in internal/api/groups_test.go
- [ ] T067 [US4] Test: admin updates permission flags → flags saved
- [ ] T068 [US4] Test: non-admin tries to update → 403 Forbidden
- [ ] T069 [US4] Test: members_can_add_members=false → member invite blocked
- [ ] T070 [US4] Test: members_can_add_members=true → member invite allowed
- [ ] T070a [US4] Test: admin can invite even when members_can_add_members=false (FR-022 admin bypass)
- [ ] T071 [US4] Test: getGroup returns all 11 permission flags

### Implementation for User Story 4

- [ ] T072 [US4] Implement handleUpdateGroup (PATCH) handler in internal/api/groups.go
- [ ] T073 [US4] Support partial updates (only update provided fields)
- [ ] T074 [US4] Implement handleGetGroup handler in internal/api/groups.go
- [ ] T075 [US4] Return GroupDetailDTO with all permission flags and counts
- [ ] T076 [US4] Update inviteMember to check members_can_add_members flag (Note: handler created in US2/T040; this task adds permission flag check as enhancement)
- [ ] T077 [US4] Register PATCH /api/v1/groups/{id} route
- [ ] T078 [US4] Register GET /api/v1/groups/{id} route
- [ ] T079 [US4] Run tests and verify all pass

**Checkpoint**: User Story 4 complete - permission configuration and enforcement working

---

## Phase 7: User Story 5 - Create Subgroups (Priority: P3)

**Goal**: Admins create subgroups linked to parent; subgroups can optionally inherit permissions

**Independent Test**: Create subgroup under parent → parent_id set correctly; list subgroups works

### Tests for User Story 5

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T080 [US5] Write table-driven tests for createSubgroup handler in internal/api/groups_test.go
- [ ] T081 [US5] Test: admin creates subgroup → parent_id set correctly
- [ ] T082 [US5] Test: member with permission creates subgroup → allowed
- [ ] T083 [US5] Test: member without permission → 403 Forbidden
- [ ] T083a [US5] Test: admin can create subgroup even when members_can_create_subgroups=false (FR-022 admin bypass)
- [ ] T084 [US5] Test: listSubgroups returns child groups
- [ ] T085 [US5] Test: subgroup cannot be its own parent (self-ref blocked)
- [ ] T085a [US5] Test: subgroup with inherit_permissions=true copies parent permission flags
- [ ] T085b [US5] Test: subgroup with inherit_permissions=false uses default permission flags

### Implementation for User Story 5

- [ ] T086 [US5] Implement handleCreateSubgroup handler in internal/api/groups.go
- [ ] T087 [US5] Check authorization (admin OR members_can_create_subgroups)
- [ ] T088 [US5] Set parent_id and copy permissions if requested
- [ ] T089 [US5] Implement handleListSubgroups handler in internal/api/groups.go
- [ ] T090 [US5] Register POST /api/v1/groups/{id}/subgroups route
- [ ] T091 [US5] Register GET /api/v1/groups/{id}/subgroups route
- [ ] T092 [US5] Run tests and verify all pass

**Checkpoint**: User Story 5 complete - subgroup hierarchy working

---

## Phase 8: User Story 6 - Archive a Group (Priority: P3)

**Goal**: Admins archive/unarchive groups; archived groups hidden from lists but accessible directly

**Independent Test**: Archive group → hidden from list; access by ID/handle still works; unarchive restores

### Tests for User Story 6

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T093 [US6] Write table-driven tests for archiveGroup handler in internal/api/groups_test.go
- [ ] T094 [US6] Test: admin archives group → archived_at timestamp set
- [ ] T095 [US6] Test: non-admin tries to archive → 403 Forbidden
- [ ] T096 [US6] Test: archived group excluded from listGroups (default)
- [ ] T097 [US6] Test: archived group included with include_archived=true
- [ ] T098 [US6] Test: unarchive sets archived_at to NULL
- [ ] T099 [US6] Test: archived group accessible via getGroup/getGroupByHandle
- [ ] T099a [US6] Test: subgroup with archived parent shows parent relationship as archived

### Implementation for User Story 6

- [ ] T100 [US6] Implement handleArchiveGroup handler in internal/api/groups.go
- [ ] T101 [US6] Implement handleUnarchiveGroup handler in internal/api/groups.go
- [ ] T102 [US6] Implement handleListGroups with include_archived filter
- [ ] T103 [US6] Implement handleGetGroupByHandle handler in internal/api/groups.go
- [ ] T103a [US6] Include parent_archived indicator in GroupDetailDTO when parent group is archived (for T099a test)
- [ ] T104 [US6] Register POST /api/v1/groups/{id}/archive route
- [ ] T105 [US6] Register POST /api/v1/groups/{id}/unarchive route
- [ ] T106 [US6] Register GET /api/v1/groups route
- [ ] T107 [US6] Register GET /api/v1/groups/handle/{handle} route
- [ ] T108 [US6] Run tests and verify all pass

**Checkpoint**: User Story 6 complete - full group lifecycle management

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Integration testing, audit verification, and cleanup

- [ ] T109 [P] Write integration test: full group creation → invite → accept → promote workflow
- [ ] T110 [P] Write audit verification tests in internal/api/audit_test.go:
  - T110a: Verify audit record created for group creation (INSERT)
  - T110b: Verify audit record created for membership invite (INSERT)
  - T110c: Verify audit record created for membership accept (UPDATE)
  - T110d: Verify audit record created for membership promote (UPDATE)
  - T110e: Verify audit record created for membership demote (UPDATE)
  - T110f: Verify audit record created for membership remove (DELETE)
  - T110g: Verify actor_id matches the authenticated user who performed each mutation
  - T110h: Verify xact_id correlates related changes within a single transaction (e.g., createGroup + createMembership)
  - T110i: Verify record/old_record JSONB contains expected field values for each action type
- [ ] T111 Verify all sqlc queries handle error cases correctly (db.IsNotFound usage)
- [ ] T112 Review and update internal/api/dto.go with any missing conversions
- [ ] T113 Run full test suite: go test ./... -v
- [ ] T114 Run quickstart.md validation (manual or automated curl tests)
- [ ] T115 Update CLAUDE.md if any new patterns or gotchas discovered

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - migrations and pgTap tests first
- **Foundational (Phase 2)**: Depends on Setup - sqlc codegen requires migrations
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - User stories can proceed in priority order (P1 → P2 → P3)
  - Within same priority: US1 and US2 are both P1, but US2 needs US1 for group creation
- **Polish (Phase 9)**: Depends on all user stories

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - no other story dependencies
- **User Story 2 (P1)**: Needs US1 complete (requires existing group to invite into)
- **User Story 3 (P2)**: Needs US2 complete (requires existing memberships to manage)
- **User Story 4 (P2)**: Needs US2 complete (T076 enhances inviteMember handler created in US2)
- **User Story 5 (P3)**: Needs US1 complete (requires parent group)
- **User Story 6 (P3)**: Needs US1 complete (requires group to archive)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models/queries before handlers
- Handlers before route registration
- Story complete before moving to next priority

### Parallel Opportunities

**Parallel WRITING (multiple developers/agents):**
- Phase 2 sqlc queries for groups/memberships can be written in parallel
- Phase 2 DTOs, handle generation, and audit helper can be written in parallel
- Within each story: test files can be written in parallel (T016-T020 can all be written simultaneously)
- US4, US5, US6 could theoretically be written in parallel if team has capacity

**TDD Discipline (per Constitution I):**
- All tests within a user story MUST be written and verified to FAIL before any implementation begins
- "Parallel test writing" means multiple test files authored concurrently, NOT tests written alongside implementation
- Example: T016-T020 (US1 tests) can be written in parallel by different agents, but all must fail before T021-T030 (US1 implementation) begins

**Parallel EXECUTION (testcontainers):**
- Each pgTap test file runs in its own isolated PostgreSQL container
- Go API tests use container snapshots for fast reset between test cases
- Test files for different user stories can execute in parallel (separate containers)

---

## Parallel Example: Phase 1 Setup

```bash
# Launch all pgTap tests in parallel (after migrations):
Task: "Write pgTap tests for audit schema in tests/pgtap/002_audit_schema_test.sql"
Task: "Write pgTap tests for groups constraints in tests/pgtap/003_groups_test.sql"
Task: "Write pgTap tests for memberships last-admin trigger in tests/pgtap/004_memberships_test.sql"
```

## Parallel Example: Phase 2 Foundational

```bash
# Launch parallel tasks (no dependencies on each other):
Task: "Create sqlc queries for groups in queries/groups.sql"
Task: "Create sqlc queries for memberships in queries/memberships.sql"

# After sqlc generate:
Task: "Create handle generation utility function in internal/api/groups.go"
Task: "Create transaction helper for audit context in internal/db/audit.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + 2 Only)

1. Complete Phase 1: Setup (migrations + pgTap)
2. Complete Phase 2: Foundational (sqlc + DTOs + helpers)
3. Complete Phase 3: User Story 1 (create group)
4. Complete Phase 4: User Story 2 (invite members)
5. **STOP and VALIDATE**: Test groups/memberships independently
6. Deploy/demo if ready - users can create groups and invite members

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → Users can create groups (MVP!)
3. Add User Story 2 → Users can invite and accept memberships
4. Add User Story 3 → Full member management
5. Add User Story 4 → Permission configuration
6. Add User Story 5 → Subgroup hierarchy
7. Add User Story 6 → Group archival
8. Each story adds value without breaking previous stories

### Suggested MVP Scope

**Minimum Viable**: User Stories 1 + 2 (P1 priority)
- Users can create groups
- Admins can invite members
- Users can accept invitations

This delivers immediate collaboration value with ~50 tasks.

---

## Task Summary

| Phase | User Story | Task Count |
|-------|-----------|------------|
| Phase 1 | Setup (incl. testcontainers) | 12 |
| Phase 2 | Foundational | 11 |
| Phase 3 | US1 - Create Group | 19 |
| Phase 4 | US2 - Invite Members | 21 |
| Phase 5 | US3 - Manage Members | 15 |
| Phase 6 | US4 - Configure Permissions | 15 |
| Phase 7 | US5 - Create Subgroups | 16 |
| Phase 8 | US6 - Archive Group | 18 |
| Phase 9 | Polish | 12 |
| **Total** | | **139** |

---

## Notes

- [P] tasks = different files, no dependencies
- **Test task pattern**: Each user story has a "write table-driven tests" task (e.g., T016) followed by individual test case tasks (e.g., T017-T020). This granularity aids progress tracking; in practice, all test cases for a story are typically written together in a single test file.
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All mutations use transaction with SET LOCAL for audit context
- DB triggers handle last-admin protection (not app layer)
- **Test isolation**: All database tests use testcontainers-go for isolated PostgreSQL containers. The `internal/testutil/postgres.go` helper provides:
  - `NewPostgresContainer(ctx)` - creates a fresh container with migrations applied
  - `Container.Snapshot(ctx)` / `Container.Restore(ctx)` - fast state reset between tests
  - `Container.RunPgTap(ctx, testFile)` - executes pgTap tests in the container
- **Parallel pgTap**: pgTap test *files* (T002, T004, T006) can be written in parallel, but each test file runs against its own container sequentially after its migration dependency completes
