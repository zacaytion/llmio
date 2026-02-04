# Testing Strategy Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor test infrastructure to use package-level container sharing, build tags for test targeting, and proper test naming conventions.

**Architecture:** Extract generic test orchestration to `internal/testutil/`, keep postgres-specific container logic in `internal/db/testutil/`. Use `TestMain` pattern with functional options for container lifecycle. Three pgTap executor implementations with pg_prove as default.

**Tech Stack:** testcontainers-go, pgx/v5, tap13 (TAP parser), pg_prove, goose migrations

---

## Phase 1: Container Image & Dockerfile

### Task 1: Create new Dockerfile with pg_prove

**Files:**
- Create: `db/Dockerfile.pgtap`

**Step 1: Create the new Dockerfile**

```dockerfile
# PostgreSQL 18 with pgTap extension and pg_prove for schema testing
# Used by testcontainers for running pgTap tests via Go tests
FROM postgres:18

# Install pgTap extension and pg_prove (Perl TAP harness)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        postgresql-18-pgtap \
        libtap-parser-sourcehandler-pgtap-perl \
    && rm -rf /var/lib/apt/lists/*

# Copy init script to enable pgtap extension on database creation
COPY init-pgtap.sql /docker-entrypoint-initdb.d/
```

**Step 2: Build the image locally to verify**

Run: `docker build -t ghcr.io/zacaytion/lmmio-pg-tap:local -f db/Dockerfile.pgtap db/`
Expected: Build completes successfully

**Step 3: Test pg_prove is available**

Run: `docker run --rm ghcr.io/zacaytion/lmmio-pg-tap:local pg_prove --version`
Expected: `pg_prove 3.37` (or similar version)

**Step 4: Commit**

```bash
git add db/Dockerfile.pgtap
git commit -m "feat(test): add Dockerfile.pgtap with pg_prove support"
```

---

### Task 2: Update Makefile with docker image targets

**Files:**
- Modify: `Makefile`

**Step 1: Add variables for image tagging**

Add after line 8 (after `GO_SRC` definition):

```makefile
# Docker test image configuration
GIT_SHA    := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD | sed 's/\//-/g')
IMAGE_NAME := ghcr.io/zacaytion/lmmio-pg-tap
```

**Step 2: Add docker targets before the Help section**

Add before `##@ Help` (before line 170):

```makefile
##@ Docker Test Image

docker-test-build: ## Build lmmio-pg-tap image locally
	docker build -t $(IMAGE_NAME):latest -f db/Dockerfile.pgtap db/

docker-test-tag: docker-test-build ## Tag image with branch and SHA
	docker tag $(IMAGE_NAME):latest $(IMAGE_NAME):br-$(GIT_BRANCH)
	docker tag $(IMAGE_NAME):latest $(IMAGE_NAME):$(GIT_SHA)

docker-test-push: docker-test-tag ## Push image to ghcr.io (requires auth)
	docker push $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):br-$(GIT_BRANCH)
	docker push $(IMAGE_NAME):$(GIT_SHA)

docker-test-image: docker-test-push ## Build, tag, and push image
```

**Step 3: Update .PHONY list**

Update line 11-14 to include new targets:

```makefile
.PHONY: all help up down logs clean-volumes \
        build build-server build-migrate run-server run-migrate server migrate install tidy \
        test test-unit test-pgtap test-integration test-all coverage-view psql lint lint-fix lint-files lint-md lint-makefile lint-migrations lint-all fmt \
        clean clean-go-build clean-go-test clean-go-mod clean-go-fuzz clean-go-all \
        docker-test-build docker-test-tag docker-test-push docker-test-image
```

**Step 4: Test the build target**

Run: `make docker-test-build`
Expected: Image builds successfully

**Step 5: Commit**

```bash
git add Makefile
git commit -m "feat(make): add docker-test-* targets for test image"
```

---

## Phase 2: Add tap13 Dependency

### Task 3: Add tap13 TAP parser dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add the dependency**

Run: `go get github.com/mpontillo/tap13@latest`
Expected: Dependency added to go.mod

**Step 2: Tidy modules**

Run: `go mod tidy`
Expected: go.sum updated

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add tap13 TAP protocol parser"
```

---

## Phase 3: Create internal/testutil Package

### Task 4: Create options.go with functional options

**Files:**
- Create: `internal/testutil/options.go`

**Step 1: Create the options file**

```go
// Package testutil provides integration test orchestration.
package testutil

// Option configures integration test behavior.
type Option func(*IntegrationTestOptions)

// IntegrationTestOptions controls the test container lifecycle.
type IntegrationTestOptions struct {
	RunMigrations bool
	CreateSnapshot bool
	SkipIfNoDocker bool
}

// WithMigrations runs database migrations after container starts.
func WithMigrations() Option {
	return func(o *IntegrationTestOptions) {
		o.RunMigrations = true
	}
}

// WithSnapshot creates a database snapshot after migrations for fast restore.
func WithSnapshot() Option {
	return func(o *IntegrationTestOptions) {
		o.CreateSnapshot = true
	}
}

// SkipIfNoDocker skips tests if Docker/Podman is not available.
func SkipIfNoDocker() Option {
	return func(o *IntegrationTestOptions) {
		o.SkipIfNoDocker = true
	}
}

// applyOptions applies all options to the config.
func applyOptions(opts []Option) *IntegrationTestOptions {
	cfg := &IntegrationTestOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
```

**Step 2: Verify file compiles**

Run: `go build ./internal/testutil/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/testutil/options.go
git commit -m "feat(testutil): add functional options for integration tests"
```

---

### Task 5: Create integration.go with RunIntegrationTests

**Files:**
- Create: `internal/testutil/integration.go`

**Step 1: Create the integration test orchestrator**

```go
//go:build integration || pgtap

package testutil

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	dbtestutil "github.com/zacaytion/llmio/internal/db/testutil"
)

var (
	// Package-level container state (shared within a test package)
	sharedContainer *dbtestutil.PostgresContainer
	sharedPool      *pgxpool.Pool
	sharedConnStr   string
	containerMu     sync.Mutex
)

// RunIntegrationTests handles the container lifecycle for integration tests.
// Call this from TestMain in packages that need database access.
//
// Example:
//
//	func TestMain(m *testing.M) {
//	    os.Exit(testutil.RunIntegrationTests(m,
//	        testutil.WithMigrations(),
//	        testutil.WithSnapshot(),
//	    ))
//	}
func RunIntegrationTests(m *testing.M, opts ...Option) int {
	cfg := applyOptions(opts)
	ctx := context.Background()

	// Create container
	container, err := dbtestutil.NewPostgresContainer(ctx)
	if err != nil {
		if cfg.SkipIfNoDocker {
			// Can't easily skip from TestMain, so just return success
			return 0
		}
		_, _ = os.Stderr.WriteString("failed to create postgres container: " + err.Error() + "\n")
		return 1
	}

	// Store for package-level access
	containerMu.Lock()
	sharedContainer = container
	containerMu.Unlock()

	// Ensure cleanup
	defer func() {
		containerMu.Lock()
		if sharedPool != nil {
			sharedPool.Close()
			sharedPool = nil
		}
		sharedContainer = nil
		sharedConnStr = ""
		containerMu.Unlock()

		if err := container.Terminate(ctx); err != nil {
			_, _ = os.Stderr.WriteString("failed to terminate container: " + err.Error() + "\n")
		}
	}()

	// Run migrations if requested
	if cfg.RunMigrations {
		if err := container.RunMigrations(ctx); err != nil {
			_, _ = os.Stderr.WriteString("failed to run migrations: " + err.Error() + "\n")
			return 1
		}
	}

	// Create snapshot if requested
	if cfg.CreateSnapshot {
		if err := container.Snapshot(ctx); err != nil {
			_, _ = os.Stderr.WriteString("failed to create snapshot: " + err.Error() + "\n")
			return 1
		}
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to get connection string: " + err.Error() + "\n")
		return 1
	}

	containerMu.Lock()
	sharedConnStr = connStr
	containerMu.Unlock()

	// Create shared pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to create pool: " + err.Error() + "\n")
		return 1
	}

	containerMu.Lock()
	sharedPool = pool
	containerMu.Unlock()

	// Run all tests
	return m.Run()
}

// GetPool returns the shared connection pool.
// Must be called after RunIntegrationTests has started.
func GetPool() *pgxpool.Pool {
	containerMu.Lock()
	defer containerMu.Unlock()
	return sharedPool
}

// GetContainer returns the shared PostgresContainer.
// Must be called after RunIntegrationTests has started.
func GetContainer() *dbtestutil.PostgresContainer {
	containerMu.Lock()
	defer containerMu.Unlock()
	return sharedContainer
}

// GetConnectionString returns the shared connection string.
// Must be called after RunIntegrationTests has started.
func GetConnectionString() string {
	containerMu.Lock()
	defer containerMu.Unlock()
	return sharedConnStr
}

// Restore restores the database to the last snapshot.
// Use in t.Cleanup() at the start of each top-level test.
func Restore(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	containerMu.Lock()
	container := sharedContainer
	pool := sharedPool
	containerMu.Unlock()

	if container == nil {
		t.Fatal("Restore called but no container is running")
	}

	// Close existing pool connections before restore
	if pool != nil {
		pool.Close()
	}

	if err := container.Restore(ctx); err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}

	// Recreate pool after restore
	connStr := GetConnectionString()
	newPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to recreate pool after restore: %v", err)
	}

	containerMu.Lock()
	sharedPool = newPool
	containerMu.Unlock()
}

// ParallelEnabled returns true if parallel database mode is enabled.
// Currently always returns false (feature flag for future use).
func ParallelEnabled() bool {
	return os.Getenv("LLMIO_TEST_PARALLEL_DB") == "1"
}
```

**Step 2: Verify file compiles**

Run: `go build -tags=integration ./internal/testutil/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/testutil/integration.go
git commit -m "feat(testutil): add RunIntegrationTests orchestrator"
```

---

## Phase 4: Refactor db/testutil for Container-Only Logic

### Task 6: Update PostgresContainer to use pre-built image

**Files:**
- Modify: `internal/db/testutil/postgres.go`

**Step 1: Update NewPostgresContainer to use image instead of Dockerfile**

Replace lines 61-103 (the `NewPostgresContainer` function) with:

```go
// ImageName is the container image for testing (can be overridden for local builds).
var ImageName = "ghcr.io/zacaytion/lmmio-pg-tap:latest"

// NewPostgresContainer creates a new PostgreSQL container with pgTap extension enabled.
// The container uses a pre-built image from ghcr.io that includes PostgreSQL 18 + pgTap + pg_prove.
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	root, err := projectRoot()
	if err != nil {
		return nil, err
	}
	migrationsDir := filepath.Join(root, "db", "migrations")
	testsDir := filepath.Join(root, "db", "tests")

	req := testcontainers.ContainerRequest{
		Image: ImageName,
		Env: map[string]string{
			"POSTGRES_DB":       "loomio_test",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		ProviderType:     testcontainers.ProviderPodman,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	return &PostgresContainer{
		container:     container,
		migrationsDir: migrationsDir,
		testsDir:      testsDir,
	}, nil
}
```

**Step 2: Build the test image locally first**

Run: `make docker-test-build`
Expected: Image built successfully

**Step 3: Verify the container code compiles**

Run: `go build ./internal/db/testutil/...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/db/testutil/postgres.go
git commit -m "refactor(db/testutil): use pre-built image instead of Dockerfile"
```

---

### Task 7: Add pgTap executor interface and implementations

**Files:**
- Create: `internal/db/testutil/pgtap.go`

**Step 1: Create the executor interface and implementations**

```go
package testutil

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mpontillo/tap13"
)

// PgTapExecutor runs pgTap tests against a PostgreSQL database.
type PgTapExecutor interface {
	Run(ctx context.Context, t *testing.T, testFile string) error
}

// ExecutorOption configures a PgTapExecutor.
type ExecutorOption func(*executorConfig)

type executorConfig struct {
	usePsql bool
	usePgx  bool
	pool    *pgxpool.Pool
}

// WithPsqlExecutor uses psql to execute pgTap tests.
func WithPsqlExecutor() ExecutorOption {
	return func(c *executorConfig) {
		c.usePsql = true
		c.usePgx = false
	}
}

// WithPgxExecutor uses a pgx connection pool to execute pgTap tests.
func WithPgxExecutor(pool *pgxpool.Pool) ExecutorOption {
	return func(c *executorConfig) {
		c.usePgx = true
		c.usePsql = false
		c.pool = pool
	}
}

// NewPgTapExecutor creates a new executor for running pgTap tests.
// Default is pg_prove; use WithPsqlExecutor() or WithPgxExecutor() to override.
func NewPgTapExecutor(container *PostgresContainer, opts ...ExecutorOption) PgTapExecutor {
	cfg := &executorConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.usePsql {
		return &psqlExecutor{container: container}
	}
	if cfg.usePgx {
		return &pgxExecutor{container: container, pool: cfg.pool}
	}
	// Default: pg_prove
	return &pgProveExecutor{container: container}
}

// pgProveExecutor runs tests using pg_prove inside the container.
type pgProveExecutor struct {
	container *PostgresContainer
}

func (e *pgProveExecutor) Run(ctx context.Context, t *testing.T, testFile string) error {
	t.Helper()

	// Read test file content
	// #nosec G304 -- Test file paths are trusted, provided by test code only
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	// Write content to a temp file in container
	tempName := filepath.Base(testFile)
	exitCode, reader, err := e.container.container.Exec(ctx, []string{
		"sh", "-c", fmt.Sprintf("cat > /tmp/%s", tempName),
	})
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	_ = drainReader(reader)

	// Write the content via a second exec
	exitCode, reader, err = e.container.container.Exec(ctx, []string{
		"sh", "-c", fmt.Sprintf("echo '%s' > /tmp/%s",
			strings.ReplaceAll(string(content), "'", "'\"'\"'"), tempName),
	})
	if err != nil {
		return fmt.Errorf("failed to write test content: %w", err)
	}
	_ = drainReader(reader)

	// Run pg_prove
	exitCode, reader, err = e.container.container.Exec(ctx, []string{
		"pg_prove",
		"-U", "postgres",
		"-d", "loomio_test",
		"-v",
		fmt.Sprintf("/tmp/%s", tempName),
	})
	if err != nil {
		return fmt.Errorf("failed to run pg_prove: %w", err)
	}

	output := drainReader(reader)
	outputStr := string(output)

	if exitCode != 0 {
		t.Errorf("pg_prove failed (exit %d):\n%s", exitCode, outputStr)
		return fmt.Errorf("pg_prove failed with exit code %d", exitCode)
	}

	// Log output for debugging
	t.Logf("pg_prove output:\n%s", outputStr)
	return nil
}

// psqlExecutor runs tests using psql and parses TAP output.
type psqlExecutor struct {
	container *PostgresContainer
}

func (e *psqlExecutor) Run(ctx context.Context, t *testing.T, testFile string) error {
	t.Helper()

	// Read test file
	// #nosec G304 -- Test file paths are trusted, provided by test code only
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	// Execute via psql
	exitCode, reader, err := e.container.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-d", "loomio_test",
		"-v", "ON_ERROR_STOP=1",
		"-c", string(content),
	})
	if err != nil {
		return fmt.Errorf("failed to execute psql: %w", err)
	}

	output := drainReader(reader)
	outputStr := string(output)

	if exitCode != 0 {
		t.Errorf("psql failed (exit %d):\n%s", exitCode, outputStr)
		return fmt.Errorf("psql failed with exit code %d", exitCode)
	}

	// Parse TAP output
	return parseTAPOutput(t, outputStr)
}

// pgxExecutor runs tests using a pgx connection pool.
type pgxExecutor struct {
	container *PostgresContainer
	pool      *pgxpool.Pool
}

func (e *pgxExecutor) Run(ctx context.Context, t *testing.T, testFile string) error {
	t.Helper()

	// Read test file
	// #nosec G304 -- Test file paths are trusted, provided by test code only
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	// Execute the SQL and collect results
	rows, err := e.pool.Query(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute test SQL: %w", err)
	}
	defer rows.Close()

	var output strings.Builder
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		for _, v := range values {
			if s, ok := v.(string); ok {
				output.WriteString(s)
				output.WriteString("\n")
			}
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	return parseTAPOutput(t, output.String())
}

// parseTAPOutput parses TAP protocol output and reports results to testing.T.
func parseTAPOutput(t *testing.T, output string) error {
	t.Helper()

	reader := strings.NewReader(output)
	parser := tap13.NewParser(bufio.NewReader(reader))

	suite, err := parser.Suite()
	if err != nil && err != io.EOF {
		// If parsing fails, fall back to simple check
		if strings.Contains(output, "not ok") {
			t.Errorf("pgTap test had failures:\n%s", output)
			return fmt.Errorf("test failures detected")
		}
		t.Logf("pgTap output:\n%s", output)
		return nil
	}

	if suite == nil {
		// No TAP output parsed, check for simple failure indicators
		if strings.Contains(output, "not ok") {
			t.Errorf("pgTap test had failures:\n%s", output)
			return fmt.Errorf("test failures detected")
		}
		t.Logf("pgTap output:\n%s", output)
		return nil
	}

	// Report each test result
	var failures int
	for _, test := range suite.Tests {
		switch {
		case test.Skip:
			t.Logf("SKIP: %s - %s", test.Description, test.Directive)
		case test.Todo:
			t.Logf("TODO: %s - %s", test.Description, test.Directive)
		case !test.Ok:
			t.Errorf("FAIL: %s", test.Description)
			failures++
		default:
			t.Logf("PASS: %s", test.Description)
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d test(s) failed", failures)
	}
	return nil
}

// DiscoverPgTapTests finds all pgTap test files in db/tests/.
// This is a package-level function that doesn't require a container.
func DiscoverPgTapTests() ([]string, error) {
	root, err := projectRoot()
	if err != nil {
		return nil, err
	}
	testsDir := filepath.Join(root, "db", "tests")
	return filepath.Glob(filepath.Join(testsDir, "*.sql"))
}
```

**Step 2: Verify file compiles**

Run: `go build ./internal/db/testutil/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/db/testutil/pgtap.go
git commit -m "feat(db/testutil): add pgTap executor interface with three implementations"
```

---

## Phase 5: Update Makefile Test Targets

### Task 8: Add test-unit, test-integration, test-all targets

**Files:**
- Modify: `Makefile`

**Step 1: Update the Testing section (lines 87-114)**

Replace the Testing section with:

```makefile
##@ Testing

.var/coverage:
	@mkdir -p .var/coverage || (echo "ERROR: Cannot create .var/coverage directory" >&2; exit 1)

test-unit: .var/coverage ## Run unit tests (no containers, fast)
	go test -coverprofile=.var/coverage/coverage.out ./...

test-pgtap: ## Run pgTap schema validation tests (requires container)
	go test -v -tags=pgtap ./internal/db/...

test-integration: ## Run API/DB integration tests (requires container)
	go test -v -tags=integration ./...

test-all: test-pgtap test-integration ## Run all tests (pgTap first, then integration)

test: test-unit ## Alias for test-unit (default test target)

.var/coverage/coverage.out: test-unit

coverage-view: .var/coverage/coverage.out ## View coverage report in browser
	@test -s .var/coverage/coverage.out || (echo "ERROR: No coverage data. Run 'make test-unit' first." >&2; exit 1)
	go tool cover -html=.var/coverage/coverage.out
```

**Step 2: Update test-pgtap in Database section**

Remove the old `test-pgtap` from the Database section (it's now in Testing).

**Step 3: Verify Makefile syntax**

Run: `make help`
Expected: Shows all targets including new test targets

**Step 4: Commit**

```bash
git add Makefile
git commit -m "feat(make): add test-unit, test-integration, test-all targets"
```

---

## Phase 6: Refactor pgTap Tests

### Task 9: Create setup_pgtap_integration_test.go with TestMain

**Files:**
- Create: `internal/db/setup_pgtap_integration_test.go`

**Step 1: Create TestMain for pgTap tests**

```go
//go:build pgtap

package db_test

import (
	"os"
	"testing"

	"github.com/zacaytion/llmio/internal/testutil"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.RunIntegrationTests(m,
		testutil.WithMigrations(),
		// No snapshot needed - pgTap tests are read-only schema checks
	))
}
```

**Step 2: Verify file compiles**

Run: `go build -tags=pgtap ./internal/db/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/db/setup_pgtap_integration_test.go
git commit -m "feat(db): add TestMain for pgTap integration tests"
```

---

### Task 10: Rename and refactor pgtap_test.go

**Files:**
- Rename: `internal/db/pgtap_test.go` → `internal/db/pgtap_integration_test.go`
- Modify content

**Step 1: Rename the file**

Run: `git mv internal/db/pgtap_test.go internal/db/pgtap_integration_test.go`

**Step 2: Update the file content**

Replace entire content with:

```go
//go:build pgtap

package db_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/zacaytion/llmio/internal/testutil"
	dbtestutil "github.com/zacaytion/llmio/internal/db/testutil"
)

// Test_PgTap_SchemaValidation runs all pgTap schema tests.
// Each test file in db/tests/*.sql is executed as a subtest.
func Test_PgTap_SchemaValidation(t *testing.T) {
	ctx := context.Background()
	container := testutil.GetContainer()

	if container == nil {
		t.Fatal("container not initialized - TestMain may have failed")
	}

	// Create executor (default: pg_prove)
	executor := dbtestutil.NewPgTapExecutor(container)

	// Discover test files
	testFiles, err := dbtestutil.DiscoverPgTapTests()
	if err != nil {
		t.Fatalf("failed to discover pgTap tests: %v", err)
	}

	if len(testFiles) == 0 {
		t.Skip("no pgTap test files found in db/tests/")
	}

	// Run each test file as a subtest
	for _, testFile := range testFiles {
		testName := filepath.Base(testFile)
		t.Run(testName, func(t *testing.T) {
			if err := executor.Run(ctx, t, testFile); err != nil {
				t.Errorf("pgTap test failed: %v", err)
			}
		})
	}
}
```

**Step 3: Verify tests compile**

Run: `go build -tags=pgtap ./internal/db/...`
Expected: No errors

**Step 4: Run pgTap tests**

Run: `make test-pgtap`
Expected: Tests run and pass (container starts, migrations run, pgTap tests execute)

**Step 5: Commit**

```bash
git add internal/db/pgtap_integration_test.go
git commit -m "refactor(db): rename and update pgTap tests to use TestMain pattern"
```

---

## Phase 7: Refactor API Integration Tests

### Task 11: Create setup_integration_test.go for API package

**Files:**
- Create: `internal/api/setup_integration_test.go`

**Step 1: Create TestMain for API integration tests**

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

**Step 2: Verify file compiles**

Run: `go build -tags=integration ./internal/api/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/api/setup_integration_test.go
git commit -m "feat(api): add TestMain for integration tests"
```

---

### Task 12: Create groups_integration_test.go

**Files:**
- Create: `internal/api/groups_integration_test.go`

This task extracts integration tests from `groups_test.go` into a separate file with the `integration` build tag and `api_test` package.

**Step 1: Create the integration test file**

```go
//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zacaytion/llmio/internal/api"
	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
	"github.com/zacaytion/llmio/internal/testutil"
)

// testAPISetup holds shared test infrastructure for API integration tests.
type testAPISetup struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionStore
	mux      *http.ServeMux
}

// setupAPITest creates a test environment using the shared container.
func setupAPITest(t *testing.T) *testAPISetup {
	t.Helper()

	pool := testutil.GetPool()
	if pool == nil {
		t.Fatal("pool not initialized - TestMain may have failed")
	}

	queries := db.New(pool)
	sessions := auth.NewSessionStore()

	// Create handlers
	groupHandler := api.NewGroupHandler(pool, queries, sessions)
	membershipHandler := api.NewMembershipHandler(pool, queries, sessions)

	// Create Huma API
	mux := http.NewServeMux()
	humaAPI := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	groupHandler.RegisterRoutes(humaAPI)
	membershipHandler.RegisterRoutes(humaAPI)

	return &testAPISetup{
		pool:     pool,
		queries:  queries,
		sessions: sessions,
		mux:      mux,
	}
}

// createTestUser creates a test user directly in the database.
func (s *testAPISetup) createTestUser(t *testing.T, email, name string) *db.User {
	t.Helper()
	ctx := context.Background()

	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Name:         name,
		Username:     auth.GenerateUsername(name),
		PasswordHash: hash,
		Key:          auth.GeneratePublicKey(),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Mark user as verified
	_, err = s.pool.Exec(ctx, "UPDATE users SET email_verified = true WHERE id = $1", user.ID)
	if err != nil {
		t.Fatalf("failed to verify user: %v", err)
	}

	return &user
}

// createTestSession creates a session for a user.
func (s *testAPISetup) createTestSession(t *testing.T, userID int64) string {
	t.Helper()
	token, err := s.sessions.CreateSession(userID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return token
}

// Test_CreateGroup_ValidInput tests successful group creation.
func Test_CreateGroup_ValidInput(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	setup := setupAPITest(t)
	user := setup.createTestUser(t, "creator@example.com", "Creator")
	token := setup.createTestSession(t, user.ID)

	body := map[string]any{
		"name":        "Test Group",
		"handle":      "test-group",
		"description": "A test group",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["handle"] != "test-group" {
		t.Errorf("expected handle 'test-group', got %v", resp["handle"])
	}
}

// Test_CreateGroup_DuplicateHandle tests conflict on duplicate handle.
func Test_CreateGroup_DuplicateHandle(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	setup := setupAPITest(t)
	user := setup.createTestUser(t, "creator@example.com", "Creator")
	token := setup.createTestSession(t, user.ID)

	// Create first group
	body := map[string]any{
		"name":   "First Group",
		"handle": "duplicate-handle",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first group creation failed: %d - %s", w.Code, w.Body.String())
	}

	// Try to create second group with same handle
	body["name"] = "Second Group"
	bodyBytes, _ = json.Marshal(body)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w = httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d for duplicate handle, got %d: %s",
			http.StatusConflict, w.Code, w.Body.String())
	}
}

// Test_CreateGroup_Unauthenticated tests rejection without session.
func Test_CreateGroup_Unauthenticated(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	setup := setupAPITest(t)

	body := map[string]any{
		"name":   "Test Group",
		"handle": "test-group",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// No session cookie

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}
```

**Step 2: Verify file compiles**

Run: `go build -tags=integration ./internal/api/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/api/groups_integration_test.go
git commit -m "feat(api): add groups integration tests with new pattern"
```

---

### Task 13: Update groups_test.go to be unit tests only

**Files:**
- Modify: `internal/api/groups_test.go`

**Step 1: Keep only unit tests that don't need database**

This task removes the database-dependent tests (now in `groups_integration_test.go`) and keeps only true unit tests. If there are no pure unit tests, the file should contain placeholder tests or be removed.

For now, create a placeholder that documents the split:

```go
package api

// Unit tests for the api package.
// Integration tests that require database are in groups_integration_test.go
// with the 'integration' build tag and 'api_test' package.

// TODO: Add unit tests for GroupHandler that don't require database.
// Examples:
// - Input validation tests using mock queries
// - Response formatting tests
// - Error handling logic tests
```

If there are pure unit tests you can identify in the original file, keep those. Otherwise, this becomes documentation of the split.

**Step 2: Verify file compiles**

Run: `go build ./internal/api/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/api/groups_test.go
git commit -m "refactor(api): split groups_test.go - integration tests now separate"
```

---

## Phase 8: Rename Test Functions

### Task 14: Rename test functions to Test_XXX_Scenario pattern

**Files:**
- All `*_test.go` and `*_integration_test.go` files

This task systematically renames test functions from `TestXXX` to `Test_XXX_Scenario`.

**Step 1: Use sed to rename test functions**

Run:
```bash
find internal -name '*_test.go' -exec sed -i '' 's/func TestPgTap/func Test_PgTap_SchemaValidation/g' {} \;
find internal -name '*_test.go' -exec sed -i '' 's/func TestHashPassword/func Test_HashPassword/g' {} \;
find internal -name '*_test.go' -exec sed -i '' 's/func TestVerifyPassword/func Test_VerifyPassword/g' {} \;
find internal -name '*_test.go' -exec sed -i '' 's/func TestCreateGroup/func Test_CreateGroup/g' {} \;
# Continue for other test functions...
```

**Note:** This is a large refactor. The implementer should:
1. List all current test function names: `grep -r "^func Test" internal/ --include="*_test.go"`
2. Rename each to follow `Test_FunctionName_Scenario` pattern
3. Update any `t.Run()` names if needed

**Step 2: Verify tests still run**

Run: `go test ./... 2>&1 | head -50`
Expected: Tests compile and run (some may fail due to missing integration setup, that's expected)

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor(test): rename test functions to Test_XXX_Scenario pattern"
```

---

## Phase 9: Delete Old Files

### Task 15: Remove old Dockerfile.test

**Files:**
- Delete: `db/Dockerfile.test`

**Step 1: Remove the old Dockerfile**

Run: `git rm db/Dockerfile.test`

**Step 2: Commit**

```bash
git commit -m "chore(db): remove old Dockerfile.test (replaced by Dockerfile.pgtap)"
```

---

## Phase 10: Final Verification

### Task 16: Run full test suite

**Step 1: Build test image**

Run: `make docker-test-build`
Expected: Image builds successfully

**Step 2: Run unit tests**

Run: `make test-unit`
Expected: All unit tests pass

**Step 3: Run pgTap tests**

Run: `make test-pgtap`
Expected: All pgTap schema tests pass

**Step 4: Run integration tests**

Run: `make test-integration`
Expected: All integration tests pass

**Step 5: Run full test suite**

Run: `make test-all`
Expected: pgTap runs first, then integration tests, all pass

**Step 6: Final commit with test verification**

```bash
git add -A
git commit -m "test: verify full test suite passes with new infrastructure"
```

---

## Summary of Files Changed

### Created
- `db/Dockerfile.pgtap`
- `internal/testutil/options.go`
- `internal/testutil/integration.go`
- `internal/db/testutil/pgtap.go`
- `internal/db/setup_pgtap_integration_test.go`
- `internal/api/setup_integration_test.go`
- `internal/api/groups_integration_test.go`

### Modified
- `Makefile` (docker targets, test targets)
- `go.mod` / `go.sum` (tap13 dependency)
- `internal/db/testutil/postgres.go` (use pre-built image)
- `internal/api/groups_test.go` (unit tests only)
- All `*_test.go` files (rename functions)

### Renamed
- `internal/db/pgtap_test.go` → `internal/db/pgtap_integration_test.go`

### Deleted
- `db/Dockerfile.test`

---

## Dependencies

- `github.com/mpontillo/tap13` - TAP protocol parser
