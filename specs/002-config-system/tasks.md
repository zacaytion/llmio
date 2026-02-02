# Tasks: Configuration System

**Input**: Design documents from `/specs/002-config-system/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: Included per Constitution Principle I (Test-First Development)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go module**: `cmd/`, `internal/` at repository root
- Config files at repository root

---

## Phase 1: Setup (Dependencies & Project Structure)

**Purpose**: Add dependencies and create new package directories

- [x] T001 Add Viper dependency with `go get github.com/spf13/viper`
- [x] T002 [P] Create directory `internal/config/`
- [x] T003 [P] Create directory `internal/logging/`
- [x] T004 [P] Update `.gitignore` to add `config.yaml` and `config.local.yaml`

---

## Phase 2: Foundational (Config Structs & Core Loading)

**Purpose**: Core config infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational

- [x] T005 [P] Write failing test for Config struct existence in `internal/config/config_test.go`
- [x] T006 [P] Write failing test for Load() with defaults in `internal/config/config_test.go`
- [x] T007 [P] Write failing test for DatabaseConfig.DSN() method in `internal/config/config_test.go`

### Implementation for Foundational

- [x] T008 Create Config, DatabaseConfig, ServerConfig, SessionConfig, LoggingConfig structs in `internal/config/config.go`
- [x] T009 Implement DatabaseConfig.DSN() method in `internal/config/config.go`
- [x] T010 Implement setDefaults() function with all default values in `internal/config/config.go`
- [x] T011 Implement Load() function with Viper file discovery in `internal/config/config.go`
- [x] T012 Run tests to verify foundational config works: `go test ./internal/config/... -v`

**Checkpoint**: Foundation ready - config structs exist and Load() returns defaults

---

## Phase 3: User Story 1 - Developer Configures Local Environment (Priority: P1) 🎯 MVP

**Goal**: Developer creates config.yaml with custom values and server uses them

**Independent Test**: Create config file → Start server → Verify server uses config values

### Tests for User Story 1

- [x] T013 [P] [US1] Write failing test for YAML file loading in `internal/config/config_test.go`
- [x] T014 [P] [US1] Write failing test for NewPoolFromConfig() in `internal/db/pool_test.go`
- [x] T015 [P] [US1] Write failing test for NewSessionStoreWithConfig() in `internal/auth/session_test.go`

### Implementation for User Story 1

- [x] T016 [US1] Add YAML file reading to Load() in `internal/config/config.go`
- [x] T017 [US1] Add NewPoolFromConfig() function in `internal/db/pool.go`
- [x] T018 [US1] Add duration field to SessionStore and NewSessionStoreWithConfig() in `internal/auth/session.go`
- [x] T019 [US1] Create `config.example.yaml` with all settings and placeholder values
- [x] T020 [US1] Rewrite `cmd/server/main.go` with Cobra root command and config loading
- [x] T021 [US1] Update server to use NewPoolFromConfig() and NewSessionStoreWithConfig()
- [x] T022 [US1] Run tests and verify server starts with config file: `go test ./... -v`

**Checkpoint**: Server starts with `--config config.yaml` using configured values

---

## Phase 4: User Story 2 - Developer Runs Tests Against Test Database (Priority: P1)

**Goal**: Developer uses config.test.yaml to run tests against isolated loomio_test database

**Independent Test**: Create test config → Run migrate → Run tests → Verify test DB used

### Tests for User Story 2

- [x] T023 [P] [US2] Write test verifying config.test.yaml loads different database name in `internal/config/config_test.go`

### Implementation for User Story 2

- [x] T024 [US2] Create `config.test.yaml` pointing to loomio_test database
- [x] T025 [US2] Rewrite `cmd/migrate/main.go` with Cobra (up/down/status subcommands) and config loading
- [x] T026 [US2] Add --config flag to migrate command for test config support
- [x] T027 [US2] Verify migrate works with test config: `go run ./cmd/migrate --config config.test.yaml status`

**Checkpoint**: Migrations run against test database when using config.test.yaml

---

## Phase 5: User Story 3 - Developer Overrides Config via Command Line (Priority: P2)

**Goal**: CLI flags override config file values (--port, --log-level, etc.)

**Independent Test**: Start server with --port 9000 → Verify server listens on 9000 despite config file

### Tests for User Story 3

- [x] T028 [P] [US3] Write test verifying CLI flag overrides config file value in `internal/config/config_test.go`

### Implementation for User Story 3

- [x] T029 [US3] Add all server CLI flags (--port, --db-*, --session-*, --log-*) in `cmd/server/main.go`
- [x] T030 [US3] Bind CLI flags to Viper for priority override in `cmd/server/main.go`
- [x] T031 [US3] Add database CLI flags to migrate command in `cmd/migrate/main.go`
- [x] T032 [US3] Verify CLI override: `go run ./cmd/server --port 9000` uses port 9000

**Checkpoint**: CLI flags take precedence over config file values

---

## Phase 6: User Story 4 - Developer Uses Environment Variables (Priority: P2)

**Goal**: Environment variables override config file but are overridden by CLI flags

**Independent Test**: Set LOOMIO_SERVER_PORT=9000 → Start server → Verify port 9000

### Tests for User Story 4

- [x] T033 [P] [US4] Write test for env var override (LOOMIO_*) in `internal/config/config_test.go`

### Implementation for User Story 4

- [x] T034 [US4] Add SetEnvPrefix("LOOMIO") and AutomaticEnv() to Load() in `internal/config/config.go`
- [x] T035 [US4] Verify env var override: `LOOMIO_SERVER_PORT=9000 go run ./cmd/server` uses port 9000

**Checkpoint**: Environment variables (LOOMIO_*) override config file values

---

## Phase 7: User Story 5 - Developer Views Structured Logs (Priority: P3)

**Goal**: Configurable structured logging with slog (JSON by default, level filtering)

**Independent Test**: Set log format to JSON → Make request → Verify JSON log output

### Tests for User Story 5

- [x] T036 [P] [US5] Write failing test for logging.Setup() with JSON format in `internal/logging/logging_test.go`
- [x] T037 [P] [US5] Write failing test for log level filtering in `internal/logging/logging_test.go`
- [x] T038 [P] [US5] Write failing test for log file fallback to stdout in `internal/logging/logging_test.go`

### Implementation for User Story 5

- [x] T039 [US5] Create logging.Setup() function in `internal/logging/logging.go`
- [x] T040 [US5] Implement JSON and text format handlers in `internal/logging/logging.go`
- [x] T041 [US5] Implement log level parsing (debug/info/warn/error) in `internal/logging/logging.go`
- [x] T042 [US5] Implement file output with fallback to stdout on error in `internal/logging/logging.go`
- [x] T043 [US5] Replace log.Printf with slog calls in `internal/api/logging.go`
- [x] T044 [US5] Replace log.Printf with slog.Info in `internal/api/middleware.go`
- [x] T045 [US5] Call logging.Setup() in server startup in `cmd/server/main.go`
- [x] T046 [US5] Run tests and verify JSON logs: `go test ./internal/logging/... -v`

**Checkpoint**: Server outputs structured JSON logs by default, respects log level config

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [x] T047 Run full test suite: `go test ./... -v`
- [x] T048 Run linter: `golangci-lint run ./...`
- [x] T049 Verify server starts with defaults: `go run ./cmd/server`
- [x] T050 Verify server starts with config: `go run ./cmd/server --config config.example.yaml`
- [x] T051 Verify migrate with test config: `go run ./cmd/migrate --config config.test.yaml status`
- [x] T052 Run quickstart.md validation scenarios manually
- [x] T053 Remove old getEnv() helper from `cmd/server/main.go` if still present
- [x] T054 Final commit with all changes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 and can proceed in parallel
  - US3 and US4 are P2 and depend on US1 (need server rewrite)
  - US5 is P3 and can proceed independently after Foundation
- **Polish (Phase 8)**: Depends on all user stories being complete
- **Code Review Fixes (Phase 9)**: Depends on Phase 8 completion - addresses review findings before merge
- **Linter Configuration (Phase 10)**: Depends on Phase 9 completion - stricter linting
- **PR Review Findings (Phase 11)**: Depends on Phase 10 completion - comprehensive review fixes before merge

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Can run in parallel with US1
- **User Story 3 (P2)**: Depends on US1 (needs Cobra server from US1)
- **User Story 4 (P2)**: Depends on US1 (needs Load() function from US1)
- **User Story 5 (P3)**: Can start after Foundational - only needs logging package

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Config package changes before consumer changes (db/pool, auth/session)
- Core implementation before CLI integration
- Story complete before moving to next priority

### Parallel Opportunities

- T002, T003, T004 (Setup) can run in parallel
- T005, T006, T007 (Foundation tests) can run in parallel
- T013, T014, T015 (US1 tests) can run in parallel
- T036, T037, T038 (US5 tests) can run in parallel
- US1 and US2 can be worked on in parallel (both P1)
- US5 (logging) can be worked on in parallel with US3/US4

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Write failing test for YAML file loading in internal/config/config_test.go"
Task: "Write failing test for NewPoolFromConfig() in internal/db/pool_test.go"
Task: "Write failing test for NewSessionStoreWithConfig() in internal/auth/session_test.go"
```

## Parallel Example: User Story 5

```bash
# Launch all tests for User Story 5 together (T036, T037, T038):
Task: "Write failing test for logging.Setup() with JSON format in internal/logging/logging_test.go"
Task: "Write failing test for log level filtering in internal/logging/logging_test.go"
Task: "Write failing test for log file fallback to stdout in internal/logging/logging_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready - developers can now use config files!

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Config files work!
3. Add User Story 2 → Test independently → Test database works!
4. Add User Story 3 → Test independently → CLI flags work!
5. Add User Story 4 → Test independently → Env vars work!
6. Add User Story 5 → Test independently → Structured logs!
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 + User Story 3 (sequential - US3 needs US1)
   - Developer B: User Story 2
   - Developer C: User Story 4 + User Story 5 (US4 needs US1, then US5 is independent)
3. Stories complete and integrate independently

---

## Phase 9: Code Review Fixes

**Input**: Code review findings from 002-config-system implementation (commit 237173f)

**Purpose**: Address issues discovered during code review before merge

### Phase 9.1: Important Fixes (Should Fix Before Merge)

#### Issue 1: File Handle Leak in Logging Package

**Problem**: When logging to a file, the file handle is never closed. Could lead to "too many open files" errors.

**File**: `internal/logging/logging.go:62-74`

- [x] T055 [P] Write test for logging.Setup() returning closeable resource in `internal/logging/logging_test.go`
- [x] T056 Document single-call pattern OR return file handle for cleanup in `internal/logging/logging.go`
- [x] T057 If returning handle: Update `cmd/server/main.go` to close log file on shutdown

#### Issue 2: Missing Test for Invalid YAML Error Handling

**Problem**: Spec edge case "invalid YAML syntax → clear error message" has no unit test coverage.

**File**: `internal/config/config_test.go`

- [x] T058 [P] Write test `TestLoad_InvalidYAML` verifying error contains file info in `internal/config/config_test.go`

#### Issue 3: Document Reduced Migrate Subcommands

**Problem**: Original migrate had `up-by-one`, `up-to`, `down-to`, `redo`, `reset`. New only has basic commands.

**File**: `cmd/migrate/main.go`

- [x] T059 [P] Add comment in `cmd/migrate/main.go` documenting intentional simplification
- [x] T060 [P] Update `specs/002-config-system/quickstart.md` to reflect available migrate subcommands

**Checkpoint**: Important issues resolved - safe to merge

---

### Phase 9.2: Minor Improvements (Nice to Have)

**Purpose**: Code quality improvements that don't affect functionality

#### Issue 4: Inconsistent Error Return in Server

**Problem**: `runServer` returns error but goroutine exits directly instead of propagating.

**File**: `cmd/server/main.go:168-171`

- [x] T061 [P] Refactor server error handling to use error channel in `cmd/server/main.go`

#### Issue 5: Package-Level Config Variables

**Problem**: Package-level mutable state (`cfgFile`, `cfg`, `v`) makes testing harder.

**Files**: `cmd/server/main.go:21-23`, `cmd/migrate/main.go:18-20`

- [ ] T062 [P] Refactor `cmd/server/main.go` to pass config through function parameters (deferred - minor improvement)
- [ ] T063 [P] Refactor `cmd/migrate/main.go` to pass config through function parameters (deferred - minor improvement)

#### Issue 6: DSN Password Special Characters

**Problem**: Passwords with special characters (spaces, `=`, quotes) could break DSN format.

**File**: `internal/config/config.go:35-42`

- [x] T064 [P] Write test for DSN with special characters in password in `internal/config/config_test.go`
- [x] T065 Update `DatabaseConfig.DSN()` to handle special characters in `internal/config/config.go`

---

### Phase 9.3: Verification

**Purpose**: Ensure all fixes work correctly

- [x] T066 Run full test suite: `go test ./... -v`
- [x] T067 Run linter: `golangci-lint run ./...`
- [x] T068 Verify server starts with config file: `go run ./cmd/server --config config.example.yaml`
- [x] T069 Commit fixes with message: `fix(config): address code review findings`

**Checkpoint**: All code review fixes complete

---

### Phase 9 Dependencies

- **Phase 9.1 (Important)**: No dependencies - can start immediately after Phase 8
- **Phase 9.2 (Minor)**: Can start after Phase 9.1 or in parallel if different files
- **Phase 9.3 (Verification)**: Depends on all fix phases being complete

### Phase 9 Parallel Opportunities

- T055, T058, T059, T060 can run in parallel (different files)
- T061, T062, T063, T064 can run in parallel (different files)
- T056 depends on T055 (test first)
- T065 depends on T064 (test first)
- T057 depends on T056 (implementation first)

### Phase 9 Implementation Strategy

**Recommended Order**:

1. **T058** - Quick win: Add missing invalid YAML test
2. **T059, T060** - Quick win: Document migrate simplification
3. **T055, T056** - File handle: Write test, then document/fix
4. **T057** - If needed: Update server to close log file
5. **Minor fixes** - T061-T065 as time permits
6. **T066-T069** - Verify and commit

**MVP (Minimum to Merge)**: Complete T055-T060 (Important fixes) to address all "Should Fix" issues from code review.

---

## Phase 10: Linter Configuration & Cleanup

**Input**: Additional linter enablement and godoclint fixes

**Purpose**: Enable stricter linting and fix resulting issues

### Phase 10.1: Linter Configuration

- [x] T070 [P] Enable errorlint in `.golangci.yml` for proper error wrapping checks
- [x] T071 [P] Enable sloglint in `.golangci.yml` for slog best practices
- [x] T072 [P] Enable iface in `.golangci.yml` for interface analysis
- [x] T073 [P] Enable godoclint in `.golangci.yml` with relaxed settings (max-len: 100, no require-doc)
- [x] T074 [P] Add errcheck type assertion checking in `.golangci.yml`

### Phase 10.2: Godoclint Fixes

- [x] T075 Remove duplicate package doc from `internal/api/logging.go` (keep doc in dto.go)

### Phase 10.3: Error Handling Fixes

- [x] T076 Fix `internal/auth/key_test.go` to use `errors.Is()` instead of direct comparison

### Phase 10.4: Documentation Updates

- [x] T077 Update CLAUDE.md with learnings from config system implementation

### Phase 10.5: Verification

- [x] T078 Run linter: `golangci-lint run ./...`
- [x] T079 Commit changes with message: `chore: enable additional linters and fix godoclint issues`

**Checkpoint**: All linter issues resolved with stricter configuration

---

## Phase 11: PR Review Findings

**Input**: Comprehensive PR review findings from multi-agent analysis

**Purpose**: Address critical and important issues discovered during PR review before merge

### Phase 11.1: Critical Fixes (Must Fix Before Merge)

#### Issue 1: Silent Fallback on Log File Failure

**Problem**: When configured log file can't be opened (permissions, disk full, path doesn't exist), silently falls back to stdout. User configuration is ignored without clear indication.

**File**: `internal/logging/logging.go:107-118`

- [x] T080 [P] Write test for log file open failure returning error in `internal/logging/logging_test.go`
- [x] T081 Change `getWriterWithCleanup()` to return error instead of silent fallback in `internal/logging/logging.go`
- [x] T082 Update callers to handle log file error (fail startup or make fallback opt-in) in `cmd/server/main.go`

#### Issue 2: Silent Fallback for Migrations Directory

**Problem**: When migrations directory isn't found, silently defaults to `"migrations"` relative path. Running from wrong directory → silent failure with "no migrations to apply".

**File**: `cmd/migrate/main.go:180-196`

- [x] T083 [P] Write test for `findMigrationsDir()` returning error when not found in `cmd/migrate/main_test.go`
- [x] T084 Change `findMigrationsDir()` to return `(string, error)` in `cmd/migrate/main.go`
- [x] T085 Update callers to fail with clear error message in `cmd/migrate/main.go`

**Checkpoint**: Critical issues resolved - silent failures eliminated

---

### Phase 11.2: Important Fixes (Should Fix Before Merge)

#### Issue 3: No Graceful Shutdown for Cleanup Goroutine

**Problem**: `startSessionCleanup` goroutine has no mechanism to stop during shutdown. Minor resource leak.

**File**: `cmd/server/main.go:143-144, 208-218`

- [x] T086 [P] Add context parameter to `startSessionCleanup()` for graceful shutdown in `cmd/server/main.go`
- [x] T087 Cancel cleanup context in server shutdown handler in `cmd/server/main.go`

#### Issue 4: Discarded BindPFlag Errors

**Problem**: 16+ `_ = v.BindPFlag(...)` calls. Flag name typos would silently fail.

**Files**: `cmd/server/main.go:93-118`, `cmd/migrate/main.go:69-74`

- [x] T088 [P] Create helper function to collect BindPFlag errors in `cmd/server/main.go`
- [x] T089 [P] Create helper function to collect BindPFlag errors in `cmd/migrate/main.go`
- [x] T090 Fail startup if any flag binding fails in both commands

#### Issue 5: Silent filepath.Abs Error

**Problem**: Error discarded in `findMigrationsDir()`; returns empty string on failure.

**File**: `cmd/migrate/main.go:190-191`

- [x] T091 Handle `filepath.Abs()` error in `findMigrationsDir()` in `cmd/migrate/main.go` (covered by T084)

#### Issue 6: Invalid Log Level Defaults Silently

**Problem**: Typo like `level: deubg` → silently uses `info`.

**File**: `internal/logging/logging.go:67-79`

- [x] T092 [P] Write test for invalid log level warning in `internal/logging/logging_test.go`
- [x] T093 Add warning log for invalid log level in `parseLevel()` in `internal/logging/logging.go`

#### Issue 7: Invalid Log Format Defaults Silently

**Problem**: `format: tect` → silently uses JSON.

**File**: `internal/logging/logging.go:125-137`

- [x] T094 [P] Write test for invalid log format warning in `internal/logging/logging_test.go`
- [x] T095 Add warning log for invalid log format in `createHandler()` in `internal/logging/logging.go`

**Checkpoint**: Important issues resolved - warnings added for silent defaults

---

### Phase 11.3: Test Coverage Gaps

#### Issue 8: Missing Full Priority Chain Test

**Problem**: No test verifying CLI > env > file > defaults all together.

**File**: `internal/config/config_test.go`

- [x] T096 [P] Write test `TestLoad_FullPriorityChain` verifying CLI > env > file > defaults in `internal/config/config_test.go`

#### Issue 9: Missing Env Overrides Config File Test

**Problem**: Env > defaults tested, but not env > file.

**File**: `internal/config/config_test.go`

- [x] T097 [P] Write test `TestLoad_EnvOverridesConfigFile` in `internal/config/config_test.go`

#### Issue 10: Missing File Handle Closure Verification

**Problem**: Cleanup function tested, but not that file is actually closed.

**File**: `internal/logging/logging_test.go`

- [x] T098 [P] Write test verifying file handle is closed after cleanup in `internal/logging/logging_test.go`

**Checkpoint**: Test coverage gaps addressed

---

### Phase 11.4: Type Design Improvements (Follow-up PR)

**Purpose**: Improve type safety and validation - can be deferred to follow-up PR

#### Issue 11: No Validation in Config Types

**Problem**: All config types lack validation; invalid states are representable (e.g., `Port: -1`, `MinConns > MaxConns`).

**Files**: `internal/config/config.go`

- [x] T099 [P] Add `Validate() error` method to `DatabaseConfig` in `internal/config/config.go`
- [x] T100 [P] Add `Validate() error` method to `ServerConfig` in `internal/config/config.go`
- [x] T101 [P] Add `Validate() error` method to `SessionConfig` in `internal/config/config.go`
- [x] T102 [P] Add `Validate() error` method to `LoggingConfig` in `internal/config/config.go`
- [x] T103 Add `Validate() error` method to `Config` that calls sub-config validators in `internal/config/config.go`
- [x] T104 Call `Validate()` in `Load()` before returning config in `internal/config/config.go`

#### Issue 12: Stringly-Typed Enumerations

**Problem**: SSLMode, LogLevel, LogFormat should be custom types with validation.

**Files**: `internal/config/config.go`, `internal/logging/logging.go`

- [x] T105 [P] Create `SSLMode` type with constants and `Valid()` method in `internal/config/config.go`
- [x] T106 [P] Create `LogLevel` type with constants in `internal/config/config.go`
- [x] T107 [P] Create `LogFormat` type with constants in `internal/config/config.go`

**Checkpoint**: Type design improvements complete (follow-up PR)

---

### Phase 11.5: Verification

**Purpose**: Ensure all fixes work correctly

- [x] T108 Run full test suite: `go test ./... -v`
- [x] T109 Run linter: `golangci-lint run ./...`
- [x] T110 Verify server starts with defaults: `go run ./cmd/server`
- [x] T111 Verify server starts with config file: `go run ./cmd/server --config config.example.yaml`
- [x] T112 Verify migrate with test config: `go run ./cmd/migrate --config config.test.yaml status`
- [x] T113 Commit fixes with message: `fix(config): address PR review findings`

**Checkpoint**: All PR review fixes complete

---

### Phase 11 Dependencies

- **Phase 11.1 (Critical)**: No dependencies - must complete before merge
- **Phase 11.2 (Important)**: Can run in parallel with 11.1 (different files mostly)
- **Phase 11.3 (Tests)**: Can run in parallel with 11.1 and 11.2
- **Phase 11.4 (Type Design)**: Can be deferred to follow-up PR
- **Phase 11.5 (Verification)**: Depends on 11.1, 11.2, 11.3 completion

### Phase 11 Parallel Opportunities

- T080, T083, T086, T088, T089, T092, T094, T096, T097, T098 can run in parallel (different files or test files)
- T081 depends on T080 (test first)
- T082 depends on T081 (implementation first)
- T084 depends on T083 (test first)
- T085 depends on T084 (implementation first)
- T093 depends on T092 (test first)
- T095 depends on T094 (test first)
- T099-T102 can run in parallel (different methods)
- T103 depends on T099-T102 (sub-validators first)
- T104 depends on T103 (Validate method first)
- T105, T106, T107 can run in parallel (different files/types)

### Phase 11 Implementation Strategy

**Recommended Order (MVP to Merge)**:

1. **Critical (must fix)**: T080-T085 - Eliminate silent failures
2. **Important (should fix)**: T086-T095 - Graceful shutdown, flag binding, warnings
3. **Test gaps**: T096-T098 - Improve coverage
4. **Verify**: T108-T113 - Run tests and commit

**Deferred to Follow-up PR**: T099-T107 (Type design improvements)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD per Constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- Phase 9 tasks (T055-T069) are code review fixes - minor improvements (Phase 9.2) can be deferred to follow-up PR
