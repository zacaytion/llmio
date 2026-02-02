# Tasks: User Authentication

**Input**: Design documents from `/specs/001-user-auth/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth.yaml

**Tests**: Required per constitution (Principle I: Test-First Development)

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Paths follow constitution project structure: `cmd/`, `internal/`, `migrations/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and Go module setup

- [x] T001 Create project directory structure per plan.md: `cmd/server/`, `internal/api/`, `internal/auth/`, `internal/db/`, `migrations/`, `openapi/`
- [x] T002 Initialize Go module with `go mod init` and add dependencies: huma/v2, pgx/v5, argon2, goose/v3
- [x] T003 [P] Create sqlc.yaml configuration in repository root
- [x] T004 [P] Copy OpenAPI contract from specs/001-user-auth/contracts/auth.yaml to openapi/paths/auth.yaml

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema, password hashing, and session infrastructure that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Create database migration migrations/001_create_users.sql with schema from data-model.md
- [x] T006 Write pgTap tests for users table schema in migrations/001_create_users_test.sql
- [x] T007 Create SQL queries in internal/db/queries/users.sql (CreateUser, GetUserByEmail, GetUserByID, UsernameExists, EmailExists)
- [x] T008 Run `sqlc generate` to create Go types in internal/db/
- [x] T009 [P] Write unit tests for password hashing in internal/auth/password_test.go (hash, verify, timing)
- [x] T010 [P] Implement Argon2id password hashing in internal/auth/password.go with OWASP parameters
- [x] T011 [P] Write unit tests for username generation in internal/auth/username_test.go
- [x] T012 [P] Implement username slug generation in internal/auth/username.go
- [x] T013 [P] Write unit tests for public key generation in internal/auth/key_test.go
- [x] T014 [P] Implement public URL key generation in internal/auth/key.go
- [x] T015 [P] Write unit tests for session store in internal/auth/session_test.go (create, get, delete, expiry)
- [x] T016 [P] Implement in-memory session store with sync.Map in internal/auth/session.go
- [x] T017 Create UserDTO and error response types in internal/api/dto.go
- [x] T018 Create database connection pool helper in internal/db/pool.go
- [x] T019 Create main server entrypoint with Huma router in cmd/server/main.go

**Checkpoint**: Foundation ready - all infrastructure in place, tests pass

---

## Phase 3: User Story 1 - User Registration (Priority: P1) MVP

**Goal**: New users can create accounts with email, name, and password

**Independent Test**: Submit registration with valid data → user record created with hashed password

### Tests for User Story 1 (TDD - Write First, Must Fail)

- [x] T020 [P] [US1] Write API test for successful registration in internal/api/auth_test.go (TestRegisterSuccess)
- [x] T021 [P] [US1] Write API test for duplicate email rejection in internal/api/auth_test.go (TestRegisterDuplicateEmail)
- [x] T022 [P] [US1] Write API test for password validation in internal/api/auth_test.go (TestRegisterPasswordValidation)
- [x] T023 [P] [US1] Write API test for name required in internal/api/auth_test.go (TestRegisterNameRequired)
- [x] T024 [P] [US1] Write integration test for user creation in internal/db/users_test.go (TestCreateUser) - verify username and key are generated

### Implementation for User Story 1

- [x] T025 [US1] Implement registration request/response types in internal/api/auth.go (RegistrationRequest, RegistrationResponse)
- [x] T026 [US1] Implement registration handler in internal/api/auth.go (CreateRegistration)
- [x] T027 [US1] Add registration validation: email format, password length, password match, name required
- [x] T028 [US1] Wire registration endpoint POST /api/v1/registrations in cmd/server/main.go
- [x] T029 [US1] Run all US1 tests and verify they pass

**Checkpoint**: Registration works independently - users can create accounts

---

## Phase 4: User Story 2 - User Login (Priority: P1)

**Goal**: Registered users with verified emails can log in and receive session cookie

**Independent Test**: Log in with valid credentials → session created, cookie set, user data returned

### Tests for User Story 2 (TDD - Write First, Must Fail)

- [x] T030 [P] [US2] Write API test for successful login in internal/api/auth_test.go (TestLoginSuccess)
- [x] T031 [P] [US2] Write API test for invalid credentials (wrong password) in internal/api/auth_test.go (TestLoginWrongPassword)
- [x] T032 [P] [US2] Write API test for invalid credentials (unknown email) in internal/api/auth_test.go (TestLoginUnknownEmail)
- [x] T033 [P] [US2] Write API test for unverified email rejection in internal/api/auth_test.go (TestLoginUnverifiedEmail)
- [x] T034 [P] [US2] Write API test for deactivated account rejection in internal/api/auth_test.go (TestLoginDeactivatedAccount)
- [x] T035 [P] [US2] Write timing test to verify no account enumeration in internal/api/auth_test.go (TestLoginTimingConsistency)

### Implementation for User Story 2

- [x] T036 [US2] Implement login request/response types in internal/api/auth.go (LoginRequest, LoginResponse)
- [x] T037 [US2] Implement login handler in internal/api/auth.go (CreateSession) with constant-time checks
- [x] T038 [US2] Add cookie setting with HttpOnly, Secure, SameSite=Lax, 7-day expiry
- [x] T039 [US2] Wire login endpoint POST /api/v1/sessions in cmd/server/main.go
- [x] T040 [US2] Run all US2 tests and verify they pass

**Checkpoint**: Login works independently - users can authenticate

---

## Phase 5: User Story 3 - User Logout (Priority: P2)

**Goal**: Logged-in users can log out, terminating their session

**Independent Test**: Authenticated user calls logout → session deleted, cookie cleared

### Tests for User Story 3 (TDD - Write First, Must Fail)

- [x] T041 [P] [US3] Write API test for successful logout in internal/api/auth_test.go (TestLogoutSuccess)
- [x] T042 [P] [US3] Write API test for logout without session in internal/api/auth_test.go (TestLogoutUnauthenticated)
- [x] T043 [P] [US3] Write test for session invalidation after logout in internal/api/auth_test.go (TestSessionInvalidAfterLogout)

### Implementation for User Story 3

- [x] T044 [US3] Implement logout handler in internal/api/auth.go (DestroySession)
- [x] T045 [US3] Implement cookie clearing with Max-Age=0
- [x] T046 [US3] Wire logout endpoint DELETE /api/v1/sessions in cmd/server/main.go
- [x] T047 [US3] Run all US3 tests and verify they pass

**Checkpoint**: Logout works independently - users can end sessions

---

## Phase 6: User Story 4 - Session Persistence (Priority: P2)

**Goal**: Sessions persist across requests; expired sessions are rejected

**Independent Test**: Make authenticated request → session validated from cookie, user data returned

### Tests for User Story 4 (TDD - Write First, Must Fail)

- [x] T048 [P] [US4] Write API test for get current user in internal/api/auth_test.go (TestGetCurrentUser)
- [x] T049 [P] [US4] Write API test for invalid/missing session in internal/api/auth_test.go (TestGetCurrentUserUnauthenticated)
- [x] T050 [P] [US4] Write API test for expired session rejection in internal/api/auth_test.go (TestExpiredSessionRejected)
- [x] T051 [P] [US4] Write test for session expiry cleanup in internal/auth/session_test.go (TestSessionCleanup)

### Implementation for User Story 4

- [x] T052 [US4] Implement auth middleware to extract and validate session cookie in internal/api/middleware.go
- [x] T053 [US4] Implement get current user handler in internal/api/auth.go (GetCurrentSession)
- [x] T054 [US4] Wire GET /api/v1/sessions/me endpoint with auth middleware in cmd/server/main.go
- [x] T055 [US4] Implement background session cleanup goroutine in internal/auth/session.go
- [x] T056 [US4] Run all US4 tests and verify they pass

**Checkpoint**: Session persistence works - multi-tab and page refresh maintain auth

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and hardening

- [x] T057 [P] Add request logging middleware in internal/api/middleware.go
- [x] T058 [P] Add input sanitization for email (lowercase, trim) across all endpoints
- [x] T059 Run full test suite: `go test ./... -v`
- [x] T060 Validate against quickstart.md curl examples
- [x] T061 Run golangci-lint and fix any issues
- [x] T062 Verify all acceptance scenarios from spec.md pass manually

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational completion
- **User Story 2 (Phase 4)**: Depends on Foundational completion (can parallel with US1 if staffed)
- **User Story 3 (Phase 5)**: Depends on US2 (needs login to test logout)
- **User Story 4 (Phase 6)**: Depends on US2 (needs session to test persistence)
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

```
              ┌────────────────────────────────┐
              │        Foundational            │
              │          (Phase 2)             │
              └────────────────────────────────┘
                    │                │
         ┌─────────┘                └─────────┐
         ▼                                    ▼
┌─────────────────┐                 ┌─────────────────┐
│  US1 Register   │                 │   US2 Login     │
│   (Phase 3)     │                 │   (Phase 4)     │
└─────────────────┘                 └─────────────────┘
                                           │
                          ┌────────────────┼────────────────┐
                          ▼                                 ▼
               ┌─────────────────┐               ┌─────────────────┐
               │   US3 Logout    │               │  US4 Persistence│
               │   (Phase 5)     │               │   (Phase 6)     │
               └─────────────────┘               └─────────────────┘
```

### Within Each User Story

1. Write all tests (marked [P] can run in parallel)
2. Verify tests fail (TDD Red phase)
3. Implement code in order (non-[P] tasks are sequential)
4. Run tests until all pass (TDD Green phase)
5. Commit and move to next story

### Parallel Opportunities

**Phase 1 (Setup)**:
- T003, T004 can run in parallel

**Phase 2 (Foundational)** - After T005-T008 (schema/sqlc):
- T009+T010 (password), T011+T012 (username), T013+T014 (key), T015+T016 (session) - all 4 pairs can run in parallel

**Phase 3-6 (User Stories)** - Within each story:
- All test tasks marked [P] can run in parallel
- Implementation follows test completion

**Cross-Story**:
- US1 (Phase 3) and US2 (Phase 4) can run in parallel after Foundational
- US3 (Phase 5) and US4 (Phase 6) can run in parallel after US2

---

## Parallel Example: Foundational Phase

```bash
# After T005-T008 complete (schema must exist for all following):

# Launch all auth utilities in parallel:
Task: "Write unit tests for password hashing in internal/auth/password_test.go"
Task: "Write unit tests for username generation in internal/auth/username_test.go"
Task: "Write unit tests for public key generation in internal/auth/key_test.go"
Task: "Write unit tests for session store in internal/auth/session_test.go"

# Then implement each (can also parallel):
Task: "Implement Argon2id password hashing in internal/auth/password.go"
Task: "Implement username slug generation in internal/auth/username.go"
Task: "Implement public URL key generation in internal/auth/key.go"
Task: "Implement in-memory session store in internal/auth/session.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T019)
3. Complete Phase 3: User Story 1 - Registration (T020-T029)
4. **STOP and VALIDATE**: Test registration independently
5. Deploy/demo if ready - users can create accounts

### Incremental Delivery

| Milestone | Stories Complete | Value Delivered |
|-----------|-----------------|-----------------|
| MVP | US1 | Users can register |
| Alpha | US1 + US2 | Users can register AND log in |
| Beta | US1 + US2 + US3 | Full auth flow with logout |
| 1.0 | All stories | Session persistence, ready for production |

### Task Count Summary

| Phase | Tasks | Parallel Opportunities |
|-------|-------|----------------------|
| Setup | 4 | 2 |
| Foundational | 15 | 8 |
| US1 Registration | 10 | 5 (tests) |
| US2 Login | 11 | 6 (tests) |
| US3 Logout | 7 | 3 (tests) |
| US4 Persistence | 9 | 4 (tests) |
| Polish | 6 | 2 |
| **Total** | **62** | **30** |

---

## Notes

- All tests follow Go table-driven test patterns per constitution
- Password never logged - even in test failures
- Session tokens use constant-time comparison (subtle.ConstantTimeCompare)
- Commit after each completed task or logical group
- Run `go test ./... -v` frequently to catch regressions
