# Tasks: Configuration System - Code Review Fixes

**Input**: Code review findings from 002-config-system implementation
**Prerequisites**: Completed 002-config-system implementation (commit 237173f)

**Tests**: Included per Constitution Principle I (Test-First Development)

**Organization**: Tasks grouped by severity (Important fixes first, then Minor improvements)

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions

- **Go module**: `cmd/`, `internal/` at repository root

---

## Phase 1: Important Fixes (Should Fix Before Merge)

**Purpose**: Address issues that could cause problems in production or reduce maintainability

### Issue 1: File Handle Leak in Logging Package

**Problem**: When logging to a file, the file handle is never closed. Could lead to "too many open files" errors.

**File**: `internal/logging/logging.go:62-74`

- [ ] T001 [P] Write test for logging.Setup() returning closeable resource in `internal/logging/logging_test.go`
- [ ] T002 Document single-call pattern OR return file handle for cleanup in `internal/logging/logging.go`
- [ ] T003 If returning handle: Update `cmd/server/main.go` to close log file on shutdown

### Issue 2: Missing Test for Invalid YAML Error Handling

**Problem**: Spec edge case "invalid YAML syntax → clear error message" has no unit test coverage.

**File**: `internal/config/config_test.go`

- [ ] T004 [P] Write test `TestLoad_InvalidYAML` verifying error contains file info in `internal/config/config_test.go`

### Issue 3: Document Reduced Migrate Subcommands

**Problem**: Original migrate had `up-by-one`, `up-to`, `down-to`, `redo`, `reset`. New only has basic commands.

**File**: `cmd/migrate/main.go`

- [ ] T005 [P] Add comment in `cmd/migrate/main.go` documenting intentional simplification
- [ ] T006 [P] Update `specs/002-config-system/quickstart.md` to reflect available migrate subcommands

**Checkpoint**: Important issues resolved - safe to merge

---

## Phase 2: Minor Improvements (Nice to Have)

**Purpose**: Code quality improvements that don't affect functionality

### Issue 4: Inconsistent Error Return in Server

**Problem**: `runServer` returns error but goroutine exits directly instead of propagating.

**File**: `cmd/server/main.go:168-171`

- [ ] T007 [P] Refactor server error handling to use error channel in `cmd/server/main.go`

### Issue 5: Package-Level Config Variables

**Problem**: Package-level mutable state (`cfgFile`, `cfg`, `v`) makes testing harder.

**Files**: `cmd/server/main.go:21-23`, `cmd/migrate/main.go:18-20`

- [ ] T008 [P] Refactor `cmd/server/main.go` to pass config through function parameters
- [ ] T009 [P] Refactor `cmd/migrate/main.go` to pass config through function parameters

### Issue 6: DSN Password Special Characters

**Problem**: Passwords with special characters (spaces, `=`, quotes) could break DSN format.

**File**: `internal/config/config.go:35-42`

- [ ] T010 [P] Write test for DSN with special characters in password in `internal/config/config_test.go`
- [ ] T011 Update `DatabaseConfig.DSN()` to handle special characters in `internal/config/config.go`

---

## Phase 3: Verification

**Purpose**: Ensure all fixes work correctly

- [ ] T012 Run full test suite: `go test ./... -v`
- [ ] T013 Run linter: `golangci-lint run ./...`
- [ ] T014 Verify server starts with config file: `go run ./cmd/server --config config.example.yaml`
- [ ] T015 Commit fixes with message: `fix(config): address code review findings`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Important)**: No dependencies - can start immediately
- **Phase 2 (Minor)**: Can start after Phase 1 or in parallel if different files
- **Phase 3 (Verification)**: Depends on all fix phases being complete

### Parallel Opportunities

- T001, T004, T005, T006 can run in parallel (different files)
- T007, T008, T009, T010 can run in parallel (different files)
- T002 depends on T001 (test first)
- T011 depends on T010 (test first)
- T003 depends on T002 (implementation first)

---

## Implementation Strategy

### Recommended Order

1. **T004** - Quick win: Add missing invalid YAML test
2. **T005, T006** - Quick win: Document migrate simplification
3. **T001, T002** - File handle: Write test, then document/fix
4. **T003** - If needed: Update server to close log file
5. **Minor fixes** - T007-T011 as time permits
6. **T012-T015** - Verify and commit

### MVP (Minimum to Merge)

Complete T001-T006 (Important fixes) to address all "Should Fix" issues from code review.

---

## Notes

- TDD: Write tests before implementation (T001 before T002, T010 before T011)
- Commit after each logical group of fixes
- Minor improvements (Phase 2) can be deferred to a follow-up PR if needed
