package testutil

import (
	"context"
	"fmt"
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
	_, reader, err := e.container.container.Exec(ctx, []string{
		"sh", "-c", fmt.Sprintf("cat > /tmp/%s", tempName),
	})
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	_ = drainReader(reader)

	// Write the content via a second exec
	_, reader, err = e.container.container.Exec(ctx, []string{
		"sh", "-c", fmt.Sprintf("echo '%s' > /tmp/%s",
			strings.ReplaceAll(string(content), "'", "'\"'\"'"), tempName),
	})
	if err != nil {
		return fmt.Errorf("failed to write test content: %w", err)
	}
	_ = drainReader(reader)

	// Run pg_prove
	exitCode, reader, err := e.container.container.Exec(ctx, []string{
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

	// Split output into lines for tap13 parser
	lines := strings.Split(output, "\n")
	results := tap13.Parse(lines)

	if !results.FoundTapData {
		// No TAP output parsed, check for simple failure indicators
		if strings.Contains(output, "not ok") {
			t.Errorf("pgTap test had failures:\n%s", output)
			return fmt.Errorf("test failures detected")
		}
		t.Logf("pgTap output:\n%s", output)
		return nil
	}

	// Check for bail out
	if results.BailOut {
		t.Fatalf("Bail out! %s", results.BailOutReason)
	}

	// Report each test result
	var failures int
	for _, test := range results.Tests {
		switch {
		case test.Skipped:
			t.Logf("SKIP: %s - %s", test.Description, test.DirectiveText)
		case test.Todo:
			t.Logf("TODO: %s - %s", test.Description, test.DirectiveText)
		case test.Failed:
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
