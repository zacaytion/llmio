// Package testutil provides database test utilities with isolated containers.
//
// PostgresContainer provides isolated PostgreSQL containers for testing using testcontainers-go.
// It uses a pre-built image from ghcr.io that includes PostgreSQL 18 + pgTap extension.
//
// Usage:
//
//	ctx := context.Background()
//	pg, err := testutil.NewPostgresContainer(ctx)
//	if err != nil { ... }
//	defer pg.Terminate(ctx)
//
//	// Run migrations
//	if err := pg.RunMigrations(ctx); err != nil { ... }
//
//	// Get connection for tests
//	connStr, err := pg.ConnectionString(ctx)
//
//	// Snapshot/restore for fast test reset
//	if err := pg.Snapshot(ctx); err != nil { ... }
//	t.Cleanup(func() { pg.Restore(ctx) })
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
)

// ImageName is the container image for testing (can be overridden for local builds).
var ImageName = "ghcr.io/zacaytion/llmio-pg-tap:latest"

// PostgresContainer wraps a testcontainers PostgreSQL instance with migration support.
type PostgresContainer struct {
	container testcontainers.Container

	migrationsDir string
	testsDir      string
}

// projectRoot returns the absolute path to the project root directory.
func projectRoot() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}
	// internal/db/testutil/postgres.go -> project root is ../../..
	return filepath.Join(filepath.Dir(currentFile), "..", "..", ".."), nil
}

// NewPostgresContainer creates a new PostgreSQL container with pgTap extension enabled.
// The container uses a pre-built image from ghcr.io which includes PostgreSQL 18 + pgTap.
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

// ConnectionString returns the connection string for the container database.
func (p *PostgresContainer) ConnectionString(ctx context.Context, opts ...string) (string, error) {
	host, err := p.container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := p.container.MappedPort(ctx, "5432")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/loomio_test", host, port.Port())
	if len(opts) > 0 {
		connStr += "?" + strings.Join(opts, "&")
	}
	return connStr, nil
}

// RunMigrations executes all goose migrations against the container database.
func (p *PostgresContainer) RunMigrations(ctx context.Context) error {
	connStr, err := p.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, p.migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Snapshot creates a database snapshot for fast restore between tests.
// This is implemented by creating a template database.
func (p *PostgresContainer) Snapshot(ctx context.Context) error {
	// Create template from current state
	exitCode, reader, err := p.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-c", "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'loomio_test' AND pid <> pg_backend_pid();",
	})
	if err != nil {
		return fmt.Errorf("failed to terminate connections: %w", err)
	}
	drainReader(reader)
	if exitCode != 0 {
		return fmt.Errorf("terminate connections failed with exit code %d", exitCode)
	}

	// Drop existing template if it exists (separate command to avoid transaction issues)
	exitCode, reader, err = p.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-c", "DROP DATABASE IF EXISTS loomio_test_template;",
	})
	if err != nil {
		return fmt.Errorf("failed to drop template: %w", err)
	}
	drainReader(reader)
	// Ignore exit code - DROP IF EXISTS may fail if DB doesn't exist

	// Create template database from current state
	exitCode, reader, err = p.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-c", "CREATE DATABASE loomio_test_template TEMPLATE loomio_test;",
	})
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	output := drainReader(reader)
	if exitCode != 0 {
		return fmt.Errorf("create template failed with exit code %d: %s", exitCode, string(output))
	}

	return nil
}

// Restore restores the database to the last snapshot.
// Use in t.Cleanup() to reset state between tests.
func (p *PostgresContainer) Restore(ctx context.Context) error {
	// Terminate connections to loomio_test
	exitCode, reader, err := p.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-c", "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'loomio_test' AND pid <> pg_backend_pid();",
	})
	if err != nil {
		return fmt.Errorf("failed to terminate connections: %w", err)
	}
	drainReader(reader)
	if exitCode != 0 {
		return fmt.Errorf("terminate connections failed with exit code %d", exitCode)
	}

	// Drop existing test database (separate command to avoid transaction issues)
	exitCode, reader, err = p.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-c", "DROP DATABASE IF EXISTS loomio_test;",
	})
	if err != nil {
		return fmt.Errorf("failed to drop test database: %w", err)
	}
	drainReader(reader)
	if exitCode != 0 {
		return fmt.Errorf("drop database failed with exit code %d", exitCode)
	}

	// Recreate from template
	exitCode, reader, err = p.container.Exec(ctx, []string{
		"psql",
		"-U", "postgres",
		"-c", "CREATE DATABASE loomio_test TEMPLATE loomio_test_template;",
	})
	if err != nil {
		return fmt.Errorf("failed to restore from template: %w", err)
	}
	output := drainReader(reader)
	if exitCode != 0 {
		return fmt.Errorf("restore failed with exit code %d: %s", exitCode, string(output))
	}

	return nil
}

// Terminate stops and removes the container.
// Always defer this after creating a container.
func (p *PostgresContainer) Terminate(ctx context.Context) error {
	return p.container.Terminate(ctx)
}

// DB returns a new database connection to the container.
// Caller is responsible for closing the connection.
func (p *PostgresContainer) DB(ctx context.Context) (*sql.DB, error) {
	connStr, err := p.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

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
	exitCode, reader, err := p.container.Exec(ctx, []string{
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

// ExecSQL executes arbitrary SQL against the container for test setup.
func (p *PostgresContainer) ExecSQL(ctx context.Context, sql string) error {
	exitCode, reader, err := p.container.Exec(ctx, []string{
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
	output := drainReader(reader)

	if exitCode != 0 {
		return fmt.Errorf("SQL execution failed (exit %d): %s", exitCode, string(output))
	}

	return nil
}

// drainReader reads all content from reader and returns it.
func drainReader(reader interface{ Read([]byte) (int, error) }) []byte {
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
	return output
}

// DiscoverPgTapTests finds all pgTap test files in db/tests/.
func (p *PostgresContainer) DiscoverPgTapTests() ([]string, error) {
	return filepath.Glob(filepath.Join(p.testsDir, "*.sql"))
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
