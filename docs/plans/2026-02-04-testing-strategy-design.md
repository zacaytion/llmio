# Testing Strategy Design

Date: 2026-02-04

## Overview

This document defines the testing infrastructure for the llmio project, focusing on:
- Testcontainers lifecycle management (package-level sharing, not per-test)
- pgTap schema validation tests
- Build tags for test targeting
- Test naming conventions

## Container Image Strategy

### Image Details

- **Registry:** `ghcr.io/zacaytion/lmmio-pg-tap`
- **Base:** `postgres:18` (Debian)
- **Additions:** pgTap extension, pg_prove command

### Tag Strategy

| Tag | Description |
|-----|-------------|
| `:latest` | Most recent main branch build |
| `:br-<branch>` | Per-branch (e.g., `:br-main`, `:br-005-discussions`) |
| `:<sha>` | Git short SHA for exact commit traceability |

### Dockerfile (`db/Dockerfile.pgtap`)

```dockerfile
FROM postgres:18
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        postgresql-18-pgtap \
        libtap-parser-sourcehandler-pgtap-perl \
    && rm -rf /var/lib/apt/lists/*
```

### What the Image Provides

- PostgreSQL 18 with pgTap extension available
- `pg_prove` command for TAP test execution
- No migrations baked in (migrations run at test time via Go)

## Build Tags

### Tag Definitions

| Tag | Purpose | Files |
|-----|---------|-------|
| `pgtap` | Schema validation tests | `*_integration_test.go` in `internal/db/` |
| `integration` | API/DB integration tests | `*_integration_test.go` in other packages |
| (none) | Unit tests | `*_test.go` without build tag |

### Test Execution Modes

```bash
go test ./...                           # Unit tests only (fast, no containers)
go test -tags=pgtap ./...               # pgTap schema tests only
go test -tags=integration ./...         # API/DB integration tests only
go test -tags=pgtap,integration ./...   # All container tests
```

### Ordering Constraint

pgTap tests validate the schema before any Go code runs against it. The Makefile enforces this ordering for the full suite.

## Test Naming Convention

All tests follow `Test_FunctionName_Scenario` pattern:

```go
// Good
func Test_CreateGroup_ReturnsConflictOnDuplicateHandle(t *testing.T)
func Test_HashPassword_ReturnsValidArgon2Hash(t *testing.T)
func Test_PgTap_UsersTableHasCorrectColumns(t *testing.T)

// Bad (old style)
func TestCreateGroup(t *testing.T)
func TestCreateGroupConflict(t *testing.T)
```

## Test Package Organization

### Package Naming

- **Unit tests:** `package xxx` (internal access to unexported symbols)
- **Integration tests:** `package xxx_test` (external, public API only)

### File Organization

```
internal/api/
├── groups.go
├── groups_test.go              # Unit tests, package api
├── groups_integration_test.go  # Integration tests, package api_test
└── setup_integration_test.go   # TestMain, package api_test
```

### Benefits of `_test` Packages for Integration Tests

- Forces integration tests to use exported API only
- Catches accidentally unexported dependencies
- Clear separation between unit tests (internal) and integration tests (black-box)

## Container Lifecycle Management

### Package Structure

| Package | Purpose |
|---------|---------|
| `internal/testutil/` | Generic integration test orchestration |
| `internal/db/testutil/` | Postgres/pgTap container wrapper |

### `internal/testutil/integration.go`

```go
// Options for RunIntegrationTests
type Option func(*IntegrationTestOptions)

type IntegrationTestOptions struct {
    WithMigrations bool
    WithSnapshot   bool
    SkipIfNoDocker bool
}

func WithMigrations() Option { ... }
func WithSnapshot() Option { ... }
func SkipIfNoDocker() Option { ... }

// Main entry point - handles container lifecycle
func RunIntegrationTests(m *testing.M, opts ...Option) int {
    // 1. Parse options
    // 2. Create PostgresContainer via db/testutil
    // 3. Optionally run migrations
    // 4. Optionally create snapshot
    // 5. Store container reference for test access
    // 6. Run tests: code := m.Run()
    // 7. Terminate container
    // 8. Return code
}

// Access for tests
func GetPool() *pgxpool.Pool { ... }
func GetContainer() *dbtestutil.PostgresContainer { ... }
func Restore(t *testing.T) { ... }
```

### `internal/db/testutil/postgres.go`

```go
// PostgresContainer wraps testcontainers postgres with pgTap support
type PostgresContainer struct { ... }

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error)
func (c *PostgresContainer) ConnectionString(ctx context.Context) (string, error)
func (c *PostgresContainer) RunMigrations(ctx context.Context) error
func (c *PostgresContainer) Snapshot(ctx context.Context) error
func (c *PostgresContainer) Restore(ctx context.Context) error
func (c *PostgresContainer) Terminate(ctx context.Context) error
```

### TestMain Pattern

Each package with integration tests has a `TestMain`:

```go
//go:build integration

package api_test

import (
    "os"
    "testing"

    "github.com/zacaytion/llmio/internal/testutil"
)

func TestMain(m *testing.M) {
    os.Exit(testutil.RunIntegrationTests(m,
        testutil.WithMigrations(),
        testutil.WithSnapshot(),
    ))
}
```

### Reset Strategy

- Top-level `Test_*` functions get fresh snapshot (via `t.Cleanup`)
- Subtests within `t.Run()` share state
- Sequential execution within a package (no `t.Parallel()` for integration tests)

## pgTap Integration

### Execution Strategy Interface

```go
type PgTapExecutor interface {
    Run(ctx context.Context, t *testing.T, testFile string) error
}

// Three implementations
type PgProveExecutor struct { ... }  // Default - exec pg_prove in container
type PsqlExecutor struct { ... }     // Shell out to psql
type PgxExecutor struct { ... }      // Native Go via pgx connection

// Factory function with default
func NewPgTapExecutor(container *PostgresContainer, opts ...ExecutorOption) PgTapExecutor

// Options to override
func WithPsqlExecutor() ExecutorOption { ... }
func WithPgxExecutor(pool *pgxpool.Pool) ExecutorOption { ... }
```

### TAP Parsing

- **pg_prove executor:** pg_prove handles TAP aggregation, parse its summary output
- **psql/pgx executors:** Parse raw TAP output with `github.com/mpontillo/tap13` library

### TAP Result Mapping

| TAP Result | Go Test Behavior |
|------------|------------------|
| `ok` | `t.Run()` passes |
| `not ok` | `t.Run()` fails with `t.Error()` |
| `ok # SKIP` | `t.Run()` calls `t.Skip()` |
| `ok # TODO` | `t.Run()` passes with `t.Log("TODO: ...")` |
| `Bail out!` | `t.Fatal()` stops all tests |

### pgTap Test Entry Point

```go
//go:build pgtap

package db_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/zacaytion/llmio/internal/testutil"
    dbtestutil "github.com/zacaytion/llmio/internal/db/testutil"
)

func TestMain(m *testing.M) {
    os.Exit(testutil.RunIntegrationTests(m,
        testutil.WithMigrations(),
        // No snapshot - pgTap tests are read-only
    ))
}

func Test_PgTap_SchemaValidation(t *testing.T) {
    ctx := context.Background()
    container := testutil.GetContainer()
    executor := dbtestutil.NewPgTapExecutor(container)

    testFiles, err := dbtestutil.DiscoverPgTapTests()
    if err != nil {
        t.Fatalf("discovering pgtap tests: %v", err)
    }

    for _, file := range testFiles {
        t.Run(filepath.Base(file), func(t *testing.T) {
            executor.Run(ctx, t, file)
        })
    }
}
```

## Makefile Targets

### Test Targets

```makefile
test-unit:          ## Run unit tests (no containers)
	go test ./...

test-pgtap:         ## Run pgTap schema validation tests
	go test -v -tags=pgtap ./internal/db/...

test-integration:   ## Run API/DB integration tests
	go test -v -tags=integration ./...

test-all: test-pgtap test-integration  ## Run all tests (pgTap first, then integration)
```

### Docker Image Targets

```makefile
GIT_SHA    := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD | sed 's/\//-/g')
IMAGE_NAME := ghcr.io/zacaytion/lmmio-pg-tap

docker-test-build:  ## Build lmmio-pg-tap image locally
	docker build -t $(IMAGE_NAME):latest -f db/Dockerfile.pgtap db/

docker-test-push:   ## Push image to ghcr.io
	docker push $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):br-$(GIT_BRANCH)
	docker push $(IMAGE_NAME):$(GIT_SHA)

docker-test-image: docker-test-build docker-test-push  ## Build and push
```

## Future Consideration: Parallel with DB-per-Test

### Current Default

Sequential execution for integration tests (no `t.Parallel()`).

### Feature Flag

`LLMIO_TEST_PARALLEL_DB=1` environment variable.

### Behavior

- **Flag unset/false:** Sequential execution, shared container with snapshot/restore
- **Flag set to `1`:** Parallel execution with database-per-test

### Implementation Sketch

```go
// In internal/testutil/integration.go

func WithParallelDatabases() Option {
    // Only enabled if LLMIO_TEST_PARALLEL_DB=1
    // Creates database pool instead of single DB with snapshot
}

// In test files - pattern stays the same
func Test_CreateGroup_Something(t *testing.T) {
    db := testutil.AcquireTestDB(t)  // Sequential: returns shared DB
                                      // Parallel: creates/acquires isolated DB
    t.Cleanup(func() { testutil.ReleaseTestDB(t, db) })

    // When parallel enabled, test can call t.Parallel()
    if testutil.ParallelEnabled() {
        t.Parallel()
    }

    // ... test code unchanged
}
```

### What Gets Built Now

- The `AcquireTestDB` / `ReleaseTestDB` abstraction
- Sequential implementation behind it
- Environment variable check (returns false, no-op)

### What Gets Built Later

- Parallel database pool implementation
- Pre-warming N databases at `TestMain`
- Actual `t.Parallel()` enablement

### When to Revisit

- Test suite time exceeds acceptable threshold
- CI pipeline becomes a bottleneck

## Migration Plan

### Files to Rename/Move

| Current | New |
|---------|-----|
| `internal/db/testutil/postgres.go` | Keep, refactor to container-only logic |
| `db/Dockerfile.test` | `db/Dockerfile.pgtap` |
| (new) | `internal/testutil/integration.go` |

### Test Files to Refactor

| File | Changes |
|------|---------|
| `internal/api/*_test.go` | Split: unit tests stay `package api`, integration tests move to `package api_test` with `//go:build integration` |
| `internal/db/pgtap_test.go` | Rename to `pgtap_integration_test.go`, add `//go:build pgtap`, change to `package db_test` |
| All test functions | Rename from `TestXXX` to `Test_XXX_Scenario` |

### New Files to Create

| File | Purpose |
|------|---------|
| `internal/testutil/integration.go` | `RunIntegrationTests`, options, container lifecycle |
| `internal/testutil/options.go` | Functional options definitions |
| `internal/api/setup_integration_test.go` | `TestMain` for API integration tests |
| `internal/db/setup_integration_test.go` | `TestMain` for pgTap tests |
| `db/Dockerfile.pgtap` | Postgres 18 + pgTap + pg_prove |

### Makefile Changes

- Rename `test` → `test-unit`
- Update `test-pgtap` to use `-tags=pgtap`
- Add `test-integration`, `test-all`
- Add `docker-test-build`, `docker-test-push`, `docker-test-image`

### Dependencies to Add

- `github.com/mpontillo/tap13` - TAP protocol parser

## References

### pgTap

- [pgTap Documentation](https://pgtap.org/documentation.html)
- [pg_prove](https://pgtap.org/pg_prove.html)
- [TAP Protocol](https://testanything.org/)

### Testcontainers

- [Testcontainers for Go](https://golang.testcontainers.org/)
- [PostgreSQL Module](https://golang.testcontainers.org/modules/postgres/)
- [Garbage Collector (Ryuk)](https://golang.testcontainers.org/features/garbage_collector/)

### Go Testing

- [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests)
- [Scaling Acceptance Tests](https://quii.gitbook.io/learn-go-with-tests/testing-fundamentals/scaling-acceptance-tests)
- [Test Naming](https://bitfieldconsulting.com/posts/test-names)

### GitHub Container Registry

- [Working with the Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
