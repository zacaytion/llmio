package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunPgTapTest executes a single pgTap test file against the container.
// The test file should contain TAP-formatted output from pgTap functions.
func (p *PostgresContainer) RunPgTapTest(ctx context.Context, t *testing.T, testFile string) {
	t.Helper()

	// Read the test file
	// #nosec G304 -- Test file paths are trusted, provided by test code only
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read pgTap test file %s: %v", testFile, err)
	}

	// Execute via psql in the container
	exitCode, reader, err := p.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-d", "loomio_test",
		"-v", "ON_ERROR_STOP=1",
		"-c", string(content),
	})
	if err != nil {
		t.Fatalf("failed to execute pgTap test: %v", err)
	}

	// Read output
	output := make([]byte, 0, 4096)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	outputStr := string(output)

	// Parse TAP output for failures
	if exitCode != 0 {
		t.Errorf("pgTap test exited with code %d:\n%s", exitCode, outputStr)
		return
	}

	// Check for TAP failure indicators
	if strings.Contains(outputStr, "not ok") {
		t.Errorf("pgTap test had failures:\n%s", outputStr)
		return
	}

	// Log successful output for debugging
	t.Logf("pgTap test passed:\n%s", outputStr)
}

// RunPgTapTestFile is a convenience function that creates a container, runs migrations,
// and executes a single pgTap test file. Suitable for isolated test file execution.
func RunPgTapTestFile(ctx context.Context, t *testing.T, testFile string) {
	t.Helper()

	pg, err := NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	if err := pg.RunMigrations(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	pg.RunPgTapTest(ctx, t, testFile)
}

// RunAllPgTapTests runs all pgTap test files in the tests/pgtap directory.
// Each test file runs in a subtest with shared container (migrations run once).
func RunAllPgTapTests(ctx context.Context, t *testing.T) {
	t.Helper()

	// Locate project root
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	pgtapDir := filepath.Join(projectRoot, "tests", "pgtap")

	// Find all pgTap test files
	testFiles, err := filepath.Glob(filepath.Join(pgtapDir, "*.sql"))
	if err != nil {
		t.Fatalf("failed to find pgTap tests: %v", err)
	}

	if len(testFiles) == 0 {
		t.Skip("no pgTap test files found in tests/pgtap/")
	}

	// Create container and run migrations once
	pg, err := NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	if err := pg.RunMigrations(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Run each test file as a subtest
	for _, testFile := range testFiles {
		testName := filepath.Base(testFile)
		t.Run(testName, func(t *testing.T) {
			pg.RunPgTapTest(ctx, t, testFile)
		})
	}
}

// PgTapTestSuite provides a reusable container for multiple pgTap tests.
// Use this when you want to control container lifecycle across multiple test functions.
type PgTapTestSuite struct {
	Container *PostgresContainer
	ctx       context.Context
}

// NewPgTapTestSuite creates a new test suite with a fresh container and migrations.
func NewPgTapTestSuite(ctx context.Context, t *testing.T) *PgTapTestSuite {
	t.Helper()

	pg, err := NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	if err := pg.RunMigrations(ctx); err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container after migration error: %v", termErr)
		}
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create snapshot after migrations for fast restore
	if err := pg.Snapshot(ctx); err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container after snapshot error: %v", termErr)
		}
		t.Fatalf("failed to create snapshot: %v", err)
	}

	return &PgTapTestSuite{
		Container: pg,
		ctx:       ctx,
	}
}

// Run executes a pgTap test file and restores state after.
func (s *PgTapTestSuite) Run(t *testing.T, testFile string) {
	t.Helper()

	// Restore to clean state before test
	if err := s.Container.Restore(s.ctx); err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}

	s.Container.RunPgTapTest(s.ctx, t, testFile)
}

// Close terminates the container.
func (s *PgTapTestSuite) Close(t *testing.T) {
	t.Helper()

	if err := s.Container.Terminate(s.ctx); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

// SetupTestDB is a helper for Go tests that need a database connection.
// It creates a container, runs migrations, and returns a connection string.
// The cleanup function should be deferred to terminate the container.
func SetupTestDB(ctx context.Context, t *testing.T) (connStr string, cleanup func()) {
	t.Helper()

	pg, err := NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	if err := pg.RunMigrations(ctx); err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to run migrations: %v", err)
	}

	connStr, err = pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to get connection string: %v", err)
	}

	cleanup = func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return connStr, cleanup
}

// SetupTestDBWithSnapshot is like SetupTestDB but also creates a snapshot.
// Returns additional restore function for fast cleanup between tests.
func SetupTestDBWithSnapshot(ctx context.Context, t *testing.T) (connStr string, restore func(), cleanup func()) {
	t.Helper()

	pg, err := NewPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	if err := pg.RunMigrations(ctx); err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to run migrations: %v", err)
	}

	if err := pg.Snapshot(ctx); err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to create snapshot: %v", err)
	}

	connStr, err = pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		if termErr := pg.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate container: %v", termErr)
		}
		t.Fatalf("failed to get connection string: %v", err)
	}

	restore = func() {
		if err := pg.Restore(ctx); err != nil {
			t.Errorf("failed to restore snapshot: %v", err)
		}
	}

	cleanup = func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return connStr, restore, cleanup
}

// NewPoolFromConnStr creates a pgxpool.Pool from a connection string.
func NewPoolFromConnStr(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, connStr)
}

// ExecSQL executes arbitrary SQL against the container for test setup.
func (p *PostgresContainer) ExecSQL(ctx context.Context, sql string) error {
	exitCode, reader, err := p.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-d", "loomio_test",
		"-v", "ON_ERROR_STOP=1",
		"-c", sql,
	})
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	// Drain output
	buf := make([]byte, 1024)
	var output []byte
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	if exitCode != 0 {
		return fmt.Errorf("SQL execution failed (exit %d): %s", exitCode, string(output))
	}

	return nil
}
