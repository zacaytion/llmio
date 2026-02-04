//go:build integration

package api_test

import (
	"os"
	"testing"

	"github.com/zacaytion/llmio/internal/testutil"
)

func Test_Main(m *testing.M) {
	os.Exit(testutil.RunIntegrationTests(m,
		testutil.WithMigrations(),
		testutil.WithSnapshot(),
	))
}
