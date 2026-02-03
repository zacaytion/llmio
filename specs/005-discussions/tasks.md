# Tasks: Discussions & Comments

**Input**: Design documents from `/specs/005-discussions/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md, quickstart.md

**Tests**: This project follows TDD (per constitution). pgTap for schema tests, Go table-driven tests for API/service logic.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Structure)

**Purpose**: Create directories and initial file scaffolding

**Prerequisite Check** (before any tasks):
- [ ] T000 Verify Feature 004 migrations exist: `ls migrations/002_create_groups.sql migrations/003_create_memberships.sql` - if missing, sync worktree with main branch or complete Feature 004 first

- [ ] T001 Create `internal/discussion/` directory for domain logic
- [ ] T002 [P] Create `internal/db/queries/` directory for sqlc query files
- [ ] T003 [P] Create placeholder files: `internal/discussion/service.go`, `internal/discussion/comment_service.go`, `internal/discussion/permissions.go`

---

## Phase 2: Foundational (Database Schema & Core Types)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Schema Tests (pgTap) - Write First, Must Fail

- [ ] T004 [P] Create pgTap test for discussions table schema in `tests/pgtap/004_discussions_test.sql`
- [ ] T005 [P] Create pgTap test for comments table schema in `tests/pgtap/005_comments_test.sql`
- [ ] T006 [P] Create pgTap test for discussion_readers table schema in `tests/pgtap/006_discussion_readers_test.sql`

### Migrations - Make Schema Tests Pass

- [ ] T007 Create migration `migrations/004_create_discussions.sql` (discussions table per data-model.md)
- [ ] T008 Create migration `migrations/005_create_comments.sql` (comments table with self-referential FK)
- [ ] T009 Create migration `migrations/006_create_discussion_readers.sql` (read tracking table)
- [ ] T010 Run `make test-pgtap` to verify all schema tests pass

### sqlc Queries - Foundational

- [ ] T011 [P] Create `internal/db/queries/discussions.sql` with basic CRUD queries (insert, get by ID, update, delete)
- [ ] T012 [P] Create `internal/db/queries/comments.sql` with basic CRUD queries
- [ ] T013 [P] Create `internal/db/queries/discussion_readers.sql` with upsert and get queries
- [ ] T014 Run `sqlc generate` and verify generated code in `internal/db/`

### Permission Infrastructure

- [ ] T015 Implement `CanUserAccessDiscussion` in `internal/discussion/permissions.go` (checks group membership or direct participant)
- [ ] T016 Write unit tests for `CanUserAccessDiscussion` in `internal/discussion/permissions_test.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Start a Group Discussion (Priority: P1)

**Goal**: Users can create discussions within groups subject to permission flags

**Independent Test**: Create a group, add a member, verify they can create a discussion with title and description. Admins bypass permission flags.

### Tests for User Story 1

> **NOTE**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T017 [US1] Write service unit tests for `DiscussionService.Create` in `internal/discussion/service_test.go` covering:
  - Member creates discussion when `members_can_start_discussions` enabled
  - Member denied when flag disabled
  - Admin creates regardless of flag
  - Title required validation
- [ ] T018 [US1] Write API handler tests for `POST /discussions` in `internal/api/discussions_test.go`

### Implementation for User Story 1

- [ ] T019 [US1] Add `CanUserCreateDiscussion` permission check in `internal/discussion/permissions.go`
- [ ] T020 [US1] Add sqlc query `CreateDiscussion` in `internal/db/queries/discussions.sql`
- [ ] T021 [US1] Implement `DiscussionService.Create` in `internal/discussion/service.go` (ensure `private=true` default per FR-016)
- [ ] T022 [US1] Define Huma request/response types for discussions in `internal/api/discussions.go`
- [ ] T023 [US1] Implement `POST /discussions` Huma operation in `internal/api/discussions.go`
- [ ] T024 [US1] Register discussions routes in API setup (integrate with existing `internal/api/` structure)
- [ ] T025 [US1] Run tests: `go test ./internal/discussion/... ./internal/api/... -v`

**Checkpoint**: User Story 1 complete - users can create group discussions

---

## Phase 4: User Story 2 - Reply with Comments (Priority: P2)

**Goal**: Users can add comments to discussions, with nested threading up to max_depth

**Independent Test**: Create a discussion, add a comment, add a nested reply, verify depth enforcement at max_depth.

### Tests for User Story 2

- [ ] T026 [US2] Write service unit tests for `CommentService.Create` in `internal/discussion/comment_service_test.go` covering:
  - Top-level comment creation
  - Nested reply creation with correct depth
  - Depth flattening at max_depth
  - Closed discussion rejection
- [ ] T027 [US2] Write service unit tests for `CommentService.Update` (edit comment, sets edited_at)
- [ ] T028 [US2] Write API handler tests for comment endpoints in `internal/api/comments_test.go`

### Implementation for User Story 2

- [ ] T029 [US2] Add `CanUserCreateComment` permission check in `internal/discussion/permissions.go`
- [ ] T030 [US2] Add sqlc queries for comments in `internal/db/queries/comments.sql`:
  - `CreateComment` (with depth calculation)
  - `GetCommentByID`
  - `GetCommentsByDiscussion` (ordered by created_at)
  - `UpdateCommentBody` (sets edited_at)
- [ ] T031 [US2] Run `sqlc generate` to regenerate types
- [ ] T032 [US2] Implement `CommentService.Create` in `internal/discussion/comment_service.go` (includes depth validation against max_depth)
- [ ] T033 [US2] Implement `CommentService.Update` in `internal/discussion/comment_service.go`
- [ ] T034 [US2] Define Huma request/response types for comments in `internal/api/comments.go`
- [ ] T035 [US2] Implement `POST /comments` Huma operation in `internal/api/comments.go`
- [ ] T036 [US2] Implement `PATCH /comments/{id}` Huma operation in `internal/api/comments.go`
- [ ] T037 [US2] Implement `GET /discussions/{id}` to return discussion with comments in `internal/api/discussions.go`
- [ ] T038 [US2] Run tests: `go test ./internal/discussion/... ./internal/api/... -v`

**Checkpoint**: User Story 2 complete - users can comment on and reply within discussions

---

## Phase 5: User Story 3 - Close and Reopen Discussion (Priority: P3)

**Goal**: Discussion authors and admins can close/reopen discussions to control comment flow

**Independent Test**: Create discussion, close it, verify comments blocked, reopen, verify comments allowed.

### Tests for User Story 3

- [ ] T039 [US3] Write service unit tests for `DiscussionService.Close` and `DiscussionService.Reopen` in `internal/discussion/service_test.go`
- [ ] T040 [US3] Write API handler tests for close/reopen endpoints in `internal/api/discussions_test.go`

### Implementation for User Story 3

- [ ] T041 [US3] Add `CanUserCloseDiscussion` permission check in `internal/discussion/permissions.go`
- [ ] T042 [US3] Add sqlc queries `CloseDiscussion` and `ReopenDiscussion` in `internal/db/queries/discussions.sql`
- [ ] T043 [US3] Run `sqlc generate`
- [ ] T044 [US3] Implement `DiscussionService.Close` in `internal/discussion/service.go`
- [ ] T045 [US3] Implement `DiscussionService.Reopen` in `internal/discussion/service.go`
- [ ] T046 [US3] Implement `POST /discussions/{id}/close` Huma operation in `internal/api/discussions.go`
- [ ] T047 [US3] Implement `POST /discussions/{id}/reopen` Huma operation in `internal/api/discussions.go`
- [ ] T048 [US3] Run tests: `go test ./internal/discussion/... ./internal/api/... -v`

**Checkpoint**: User Story 3 complete - discussion lifecycle management works

---

## Phase 6: User Story 4 - Direct Discussions (Priority: P4)

**Goal**: Users can create discussions without a group, with explicit participant list

**Independent Test**: Create direct discussion with participants, verify only participants can access, verify adding new participant grants access.

### Tests for User Story 4

- [ ] T049 [US4] Write service unit tests for direct discussion creation in `internal/discussion/service_test.go`
- [ ] T050 [US4] Write service unit tests for `DiscussionService.AddParticipant` in `internal/discussion/service_test.go`
- [ ] T051 [US4] Write API handler tests for participant management in `internal/api/discussions_test.go`

### Implementation for User Story 4

- [ ] T052 [US4] Update `CanUserAccessDiscussion` to check `discussion_readers.participant` for direct discussions
- [ ] T053 [US4] Add sqlc query `AddParticipant` in `internal/db/queries/discussion_readers.sql`
- [ ] T054 [US4] Run `sqlc generate`
- [ ] T055 [US4] Update `DiscussionService.Create` to handle `participant_ids` for direct discussions
- [ ] T056 [US4] Implement `DiscussionService.AddParticipant` in `internal/discussion/service.go`
- [ ] T057 [US4] Implement `POST /discussions/{id}/participants` Huma operation in `internal/api/discussions.go`
- [ ] T058 [US4] Run tests: `go test ./internal/discussion/... ./internal/api/... -v`

**Checkpoint**: User Story 4 complete - direct discussions work without groups

---

## Phase 7: User Story 5 - Read Tracking (Priority: P5)

**Goal**: System tracks per-user read state and notification volume preferences

**Independent Test**: Open discussion, verify last_read_at recorded, add new comment, verify unread detection, set volume to mute.

### Tests for User Story 5

- [ ] T059 [US5] Write service unit tests for `DiscussionService.UpdateReadState` in `internal/discussion/service_test.go`
- [ ] T060 [US5] Write API handler tests for read state endpoint in `internal/api/discussions_test.go`

### Implementation for User Story 5

- [ ] T061 [US5] Add sqlc query `UpsertReadState` in `internal/db/queries/discussion_readers.sql` (uses ON CONFLICT per research.md)
- [ ] T062 [US5] Run `sqlc generate`
- [ ] T063 [US5] Implement `DiscussionService.UpdateReadState` in `internal/discussion/service.go`
- [ ] T064 [US5] Implement `PUT /discussions/{id}/read-state` Huma operation in `internal/api/discussions.go`
- [ ] T065 [US5] Update `GET /discussions/{id}` to include reader's read state in response
- [ ] T066 [US5] Run tests: `go test ./internal/discussion/... ./internal/api/... -v`

**Checkpoint**: User Story 5 complete - read tracking works

---

## Phase 8: Comment Soft Delete & Discussion CRUD

**Goal**: Complete remaining CRUD operations for discussions and comments

### Tests

- [ ] T067 Write service unit tests for `CommentService.Delete` (soft delete) in `internal/discussion/comment_service_test.go`
- [ ] T068 Write service unit tests for `DiscussionService.Update` and `DiscussionService.Delete` in `internal/discussion/service_test.go`
- [ ] T069 Write API handler tests for DELETE endpoints in `internal/api/comments_test.go` and `internal/api/discussions_test.go`

### Implementation

- [ ] T070 Add `CanUserDeleteComment` permission check in `internal/discussion/permissions.go` (author or admin)
- [ ] T071 Add sqlc query `SoftDeleteComment` in `internal/db/queries/comments.sql`
- [ ] T072 Run `sqlc generate`
- [ ] T073 Implement `CommentService.Delete` in `internal/discussion/comment_service.go` (soft delete sets discarded_at)
- [ ] T074 Update `GetCommentsByDiscussion` query to return `"[deleted]"` for body when discarded_at is set
- [ ] T075 Implement `DELETE /comments/{id}` Huma operation in `internal/api/comments.go`
- [ ] T076 Implement `PATCH /discussions/{id}` Huma operation (update title/description)
- [ ] T077 Implement `DELETE /discussions/{id}` Huma operation (hard delete cascades)
- [ ] T078 Run tests: `go test ./internal/discussion/... ./internal/api/... -v`

**Checkpoint**: All CRUD operations complete

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, edge cases, and quality improvements

- [ ] T079 Add validation for title length (1-255 chars) using go-playground/validator in request types
- [ ] T080 Add validation for comment body (non-empty) using go-playground/validator
- [ ] T081 Verify closed discussion enforcement blocks all comment creation (SC-007)
- [ ] T082 Verify soft-deleted comments retain children and display "[deleted]" (SC-006)
- [ ] T083 Run full test suite: `go test ./... -v`
- [ ] T084 Run `golangci-lint run ./...` and fix any issues
- [ ] T085 Validate quickstart.md scenarios work end-to-end
- [ ] T086 Run `make test-pgtap` to verify all schema tests still pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - US1 (Phase 3): No dependency on other stories
  - US2 (Phase 4): Depends on US1 (needs discussions to exist)
  - US3 (Phase 5): Depends on US2 (needs comments to verify blocking)
  - US4 (Phase 6): Depends on US1 (extends discussion creation)
  - US5 (Phase 7): Depends on US1 (needs discussions to track)
- **CRUD Completion (Phase 8)**: Depends on US2 for comment context
- **Polish (Phase 9)**: Depends on all prior phases

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - **MVP**
- **User Story 2 (P2)**: Best after US1 (uses discussions created)
- **User Story 3 (P3)**: Best after US2 (tests comment blocking)
- **User Story 4 (P4)**: Can start after US1 (independent participant logic)
- **User Story 5 (P5)**: Can start after US1 (reads discussion state)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Permission checks before service logic
- sqlc queries before service implementation
- Service before API handlers
- Run `sqlc generate` after adding queries
- Run tests to verify after each story

### Parallel Opportunities

**Phase 2 Parallel**:
- T004, T005, T006 (pgTap tests) can run in parallel
- T011, T012, T013 (sqlc queries) can run in parallel

**Cross-Story Parallel** (after Foundational):
- US4 and US5 can be implemented in parallel (both only depend on US1)

---

## Parallel Example: Phase 2 Foundation

```bash
# Launch all pgTap tests together:
Task: "Create pgTap test for discussions table schema in tests/pgtap/004_discussions_test.sql"
Task: "Create pgTap test for comments table schema in tests/pgtap/005_comments_test.sql"
Task: "Create pgTap test for discussion_readers table schema in tests/pgtap/006_discussion_readers_test.sql"

# After migrations, launch all sqlc query files together:
Task: "Create internal/db/queries/discussions.sql with basic CRUD queries"
Task: "Create internal/db/queries/comments.sql with basic CRUD queries"
Task: "Create internal/db/queries/discussion_readers.sql with upsert and get queries"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (group discussions)
4. **STOP and VALIDATE**: Test US1 independently via quickstart.md steps 1-3
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational -> Foundation ready
2. Add User Story 1 -> Test independently -> Deploy/Demo (MVP!)
3. Add User Story 2 -> Test independently -> Users can comment
4. Add User Story 3 -> Test independently -> Discussion lifecycle
5. Add User Story 4 -> Test independently -> Direct discussions
6. Add User Story 5 -> Test independently -> Read tracking
7. Complete Phase 8 -> Full CRUD
8. Complete Phase 9 -> Production ready

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (required first)
3. After US1 complete:
   - Developer A: User Story 2
   - Developer B: User Story 4 (parallel - different concern)
   - Developer C: User Story 5 (parallel - different concern)
4. US3 after US2 (needs comments)
5. Phase 8-9 together

---

## Notes

- [P] tasks = different files, no dependencies within phase
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD per constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Feature 004 (Groups & Memberships) must be complete first (dependency)

---

## Summary

| Metric | Count |
|--------|-------|
| Total Tasks | 87 |
| Phase 1 (Setup) | 4 |
| Phase 2 (Foundational) | 13 |
| Phase 3 (US1) | 9 |
| Phase 4 (US2) | 13 |
| Phase 5 (US3) | 10 |
| Phase 6 (US4) | 10 |
| Phase 7 (US5) | 8 |
| Phase 8 (CRUD) | 12 |
| Phase 9 (Polish) | 8 |

| User Story | Task Count | Parallel Opportunities |
|------------|------------|------------------------|
| Setup | 4 | 2 |
| Foundational | 13 | 6 |
| US1 (Group Discussions) | 9 | 0 (sequential TDD) |
| US2 (Comments) | 13 | 0 (sequential TDD) |
| US3 (Close/Reopen) | 10 | 0 (sequential TDD) |
| US4 (Direct Discussions) | 10 | 0 (sequential TDD) |
| US5 (Read Tracking) | 8 | 0 (sequential TDD) |
| CRUD Completion | 12 | 0 (sequential TDD) |
| Polish | 8 | 2 |

**MVP Scope**: Phases 1-3 (Setup + Foundational + US1) = 26 tasks
