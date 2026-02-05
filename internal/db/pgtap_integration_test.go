//go:build pgtap

package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	dbtestutil "github.com/zacaytion/llmio/internal/db/testutil"
	"github.com/zacaytion/llmio/internal/testutil"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.RunIntegrationTests(m,
		testutil.WithMigrations(),
		// No snapshot needed - pgTap tests are read-only schema checks
	))
}

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
