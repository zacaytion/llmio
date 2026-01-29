# Loomio Go Rewrite Testing Strategy

**Date:** 2026-01-29
**Status:** Proposed

## Testing Philosophy

1. **Contract tests are critical** — The Vue frontend must work unchanged
2. **TDD for business logic** — Write tests first for domain rules
3. **Integration tests for confidence** — Test real database, real Redis
4. **Property tests for complex algorithms** — Event threading, voting tallies

## Test Categories

### 1. Contract Tests (API Compatibility)

**Purpose:** Ensure Go API responses match Rails exactly.

**Approach:**
1. Generate OpenAPI spec from Rails serializers
2. Run spec against both Rails and Go in CI
3. Fail build on any contract deviation

**Tools:**
- OpenAPI spec generation (custom script from serializers)
- Schemathesis or dredd for spec testing
- JSON diff for response comparison

**Coverage Target:** 100% of API endpoints

### 2. Unit Tests (Business Logic)

**Purpose:** Test domain rules in isolation.

**Approach:**
- TDD: Write failing test, implement, verify
- Mock external dependencies (database, Redis)
- Focus on poll voting algorithms, permission logic, event threading

**Tools:**
- testify for assertions
- gomock or moq for mocks

**Coverage Target:** 80% line coverage on domain packages

### 3. Integration Tests (Real Dependencies)

**Purpose:** Test with real PostgreSQL and Redis.

**Approach:**
- Use testcontainers-go for ephemeral databases
- Seed with representative data
- Test complete request flows

**Tools:**
- testcontainers-go
- httptest for HTTP testing

**Coverage Target:** All critical paths (auth, voting, permissions)

### 4. Property-Based Tests (Algorithms)

**Purpose:** Find edge cases in complex logic.

**Approach:**
- Event tree operations (threading)
- Vote counting algorithms
- Permission inheritance

**Tools:**
- rapid (Go property testing)

**Coverage Target:** All poll types, event threading

### 5. Load Tests (Performance)

**Purpose:** Verify Go meets or exceeds Rails performance.

**Approach:**
- Benchmark key endpoints
- Compare with Rails baseline
- Test WebSocket connection scaling

**Tools:**
- k6 or vegeta
- pprof for profiling

**Baseline Metrics (to measure from Rails):**
- API response time p95
- WebSocket messages/second
- Concurrent connections

### 6. Migration Tests (Data Integrity)

**Purpose:** Verify data migration correctness.

**Approach:**
- Run migrations on production data copy
- Compare record counts
- Validate relationships intact
- Check transformed fields

**Coverage Target:** All 56 tables

## Test Pyramid

```
        /\
       /  \      E2E (Frontend + Go) - Few, slow
      /----\
     /      \    Integration (Go + DB) - Medium
    /--------\
   /          \  Contract (API spec) - Many, critical
  /------------\
 /              \ Unit (Business logic) - Many, fast
/________________\
```

## CI Pipeline

```yaml
stages:
  - lint        # golangci-lint
  - unit        # go test ./... -short
  - contract    # API spec validation
  - integration # testcontainers tests
  - benchmark   # Performance comparison
```

## Test Data Strategy

1. **Factories:** Port FactoryBot patterns to Go
2. **Fixtures:** Export subset of production data (anonymized)
3. **Generators:** Property test generators for complex types

## Migration Testing Checklist

Before each module migration:
- [ ] Contract tests passing for all endpoints
- [ ] Unit test coverage > 80%
- [ ] Integration tests for critical paths
- [ ] Load test shows performance maintained
- [ ] Shadow traffic comparison complete
