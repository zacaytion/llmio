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
