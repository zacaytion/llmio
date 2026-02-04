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
