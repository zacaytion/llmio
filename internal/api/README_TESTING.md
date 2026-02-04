# API Integration Testing Patterns

This directory contains two patterns of integration tests. The **NEW pattern** should be used for all new tests.

## NEW Pattern (Recommended)

Files using the new pattern:
- `setup_integration_test.go` - TestMain with shared container
- `groups_integration_test.go` - Example tests using shared pattern
- `audit_integration_test.go` - Audit tests using shared pattern
- `workflow_integration_test.go` - Full workflow test

### Characteristics
- Uses `package api_test` (external/black-box testing)
- Single shared container per test run via `TestMain`
- Uses `testutil.GetPool()` for shared connection pool
- Uses `t.Cleanup(func() { testutil.Restore(t) })` for test isolation
- Fast: container + migrations run once, snapshot/restore between tests

### Template

```go
//go:build integration

package api_test

import (
    "testing"
    "github.com/zacaytion/llmio/internal/api"
    "github.com/zacaytion/llmio/internal/testutil"
)

func Test_Example(t *testing.T) {
    t.Cleanup(func() { testutil.Restore(t) })

    pool := testutil.GetPool()
    // Use pool...
}
```

## OLD Pattern (Deprecated)

Files using the old pattern:
- `groups_test.go`
- `memberships_test.go`

### Characteristics
- Uses `package api` (internal access)
- Creates new container per test via `testutil.SetupTestDB()`
- Slow: container + migrations run for EVERY test function
- Uses `defer setup.cleanup()` pattern

### Migration Guide

To migrate a test from OLD to NEW pattern:

1. Change `package api` to `package api_test`
2. Add `import "github.com/zacaytion/llmio/internal/api"`
3. Replace `testutil.SetupTestDB(ctx, t)` with `testutil.GetPool()`
4. Replace `defer setup.cleanup()` with `t.Cleanup(func() { testutil.Restore(t) })`
5. Prefix internal types with `api.` (e.g., `NewGroupHandler` → `api.NewGroupHandler`)
6. Remove container creation code from setup functions

## Running Tests

```bash
# Unit tests only (no database required)
go test ./internal/...

# Integration tests (requires Podman/Docker)
go test -tags=integration ./internal/api/...

# Specific test pattern
go test -tags=integration -run Test_CreateGroup ./internal/api/...
```
