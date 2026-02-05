# API Integration Testing Patterns

This directory contains integration tests using a shared container pattern for fast execution.

## Integration Test Pattern

Files using the shared container pattern:
- `setup_integration_test.go` - TestMain with shared container
- `groups_integration_test.go` - Group API integration tests
- `memberships_integration_test.go` - Membership API integration tests
- `audit_integration_test.go` - Audit trail verification tests
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

## Companion Unit Tests

Files containing unit tests (no database required):
- `groups_test.go` - Unit tests for helper functions (e.g., `isUniqueViolation`)
- `memberships_test.go` - Unit tests for error detection functions (e.g., `isLastAdminTriggerError`)

These run with the standard `go test` command without any build tags.

## Running Tests

```bash
# Unit tests only (no database required)
go test ./internal/api/...

# Integration tests (requires Podman/Docker)
go test -tags=integration ./internal/api/...

# Specific test pattern
go test -tags=integration -run Test_CreateGroup ./internal/api/...

# All tests (unit + integration)
go test -tags=integration ./internal/api/...
```
