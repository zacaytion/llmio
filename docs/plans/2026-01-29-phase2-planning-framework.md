# Phase 2: Planning Framework Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create the strategic planning framework documents that will guide the Loomio Rails→Go rewrite.

**Architecture:** Phase 2 builds on the Discovery Report (Phase 1) to produce decision matrices, risk assessments, and planning templates. Each deliverable is a standalone document that can be reviewed independently.

**Tech Stack:** Markdown documents, Mermaid diagrams (where helpful)

---

## Task 1: Migration Strategy Decision Matrix

**Files:**
- Create: `docs/plans/2026-01-29-adr-001-migration-strategy.md`

**Step 1: Create the ADR document with context and options**

Write an ADR that evaluates the three migration strategies against Loomio-specific factors from the Discovery Report.

```markdown
# ADR-001: Migration Strategy Selection

**Date:** 2026-01-29
**Status:** Proposed

## Context

Loomio is a 10+ year old Rails 8 application with:
- 1,546 Ruby files, 56 database tables, 941 migrations
- Well-decoupled Vue 3 frontend (API-driven)
- Separate Node.js channel server for real-time features
- Active user base requiring minimal disruption

We must choose how to approach the Rails→Go rewrite.

## Options Evaluated

### Option A: Big Bang Rewrite
Complete the entire Go backend before any production deployment.

**Pros:**
- Clean slate, no dual maintenance
- Simpler infrastructure during development

**Cons:**
- 12-18 months without production feedback
- High risk if estimates are wrong
- Team morale risk from long cycles

### Option B: Strangler Fig Pattern
Gradually migrate endpoints behind a routing layer, running both systems in production.

**Pros:**
- Incremental delivery and feedback
- Lower risk per deployment
- Can pause/adjust based on learnings

**Cons:**
- Complex routing infrastructure needed
- Dual maintenance during transition
- Session/auth sharing between Rails and Go

### Option C: Hybrid Approach (Recommended)
Rewrite independent modules as Go microservices, running alongside Rails until feature parity, then cutover.

**Pros:**
- Balance of speed and risk
- Clear module boundaries from Discovery Report
- Can validate Go stack early with low-risk modules

**Cons:**
- Still requires some dual infrastructure
- Module boundaries must be carefully chosen

## Decision

**Recommended: Option C (Hybrid Approach)** with the following strategy:

1. **Phase 1:** Rewrite channel server components (records.js, bots.js) in Go (~100 lines each, low risk)
2. **Phase 2:** Rewrite read-only API endpoints (groups, discussions listing)
3. **Phase 3:** Rewrite write endpoints by domain (users, then groups, then discussions)
4. **Phase 4:** Final cutover when feature parity achieved

Preserve the Hocuspocus collaborative editing server as Node.js (Y.js ecosystem is JS-native).

## Consequences

**Positive:**
- Early production validation of Go stack
- Reduced risk through incremental delivery
- Clear rollback path at each phase

**Negative:**
- Requires routing layer for gradual migration
- Some dual maintenance during transition period

## Alternatives Considered

Options A and B as described above.
```

**Step 2: Review the ADR for completeness**

Verify:
- [ ] Context reflects Discovery Report findings
- [ ] All three options have balanced pros/cons
- [ ] Decision includes concrete phases
- [ ] Consequences are realistic

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-adr-001-migration-strategy.md
git commit -m "docs: add ADR-001 migration strategy decision"
```

---

## Task 2: Channel Server Architecture Decision

**Files:**
- Create: `docs/plans/2026-01-29-adr-002-channel-server-migration.md`

**Step 1: Create the ADR for channel server migration**

```markdown
# ADR-002: Channel Server Migration Strategy

**Date:** 2026-01-29
**Status:** Proposed

## Context

Loomio's real-time features are provided by `loomio_channel_server`, a ~150 LOC Node.js service with three components:

| Component | LOC | Complexity | Purpose |
|-----------|-----|------------|---------|
| records.js | ~50 | Low | Socket.io live updates via Redis pub/sub |
| bots.js | ~50 | Low | Matrix chat integration |
| hocuspocus.mjs | ~50 | High | Y.js collaborative editing |

## Options Evaluated

### Option A: Keep Entire Channel Server as Node.js
**Pros:** Minimal risk, Hocuspocus is mature
**Cons:** Two languages in production

### Option B: Rewrite Entire Channel Server in Go
**Pros:** Single language
**Cons:** Y.js/CRDT implementation is complex, high risk

### Option C: Hybrid (Recommended)
Rewrite records.js and bots.js in Go; keep Hocuspocus as Node.js.

**Pros:**
- Simple components get Go benefits
- Complex CRDT stays in mature ecosystem
- Reduces but doesn't eliminate Node.js

**Cons:**
- Still two languages (but minimal Node.js footprint)

## Decision

**Option C (Hybrid):**

1. Rewrite `records.js` as Go service using:
   - `gorilla/websocket` or `nhooyr.io/websocket`
   - `go-redis` for pub/sub
   - Room-based broadcasting pattern

2. Rewrite `bots.js` as Go service using:
   - `mautrix/go` for Matrix SDK
   - Same Redis subscription pattern

3. Keep `hocuspocus.mjs` as Node.js service:
   - Y.js ecosystem is JavaScript-native
   - Hocuspocus is well-tested
   - Collaborative editing is the most complex feature

## Consequences

**Positive:**
- Go handles majority of real-time traffic
- Hocuspocus stability preserved
- Clear separation of concerns

**Negative:**
- Must maintain small Node.js deployment
- Two container images for channel functionality

## Go Package Selections

| Purpose | Package | Rationale |
|---------|---------|-----------|
| WebSocket | nhooyr.io/websocket | Modern, context-aware, better than gorilla |
| Redis | go-redis/redis/v9 | Standard choice, pub/sub support |
| Matrix | mautrix/go | Official Matrix Go SDK |
```

**Step 2: Verify the ADR**

- [ ] References Discovery Report channel server analysis
- [ ] Package selections are justified
- [ ] Migration phases are clear

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-adr-002-channel-server-migration.md
git commit -m "docs: add ADR-002 channel server migration strategy"
```

---

## Task 3: Go Stack Selection ADR

**Files:**
- Create: `docs/plans/2026-01-29-adr-003-go-stack-selection.md`

**Step 1: Create the Go stack ADR**

```markdown
# ADR-003: Go Technology Stack Selection

**Date:** 2026-01-29
**Status:** Proposed

## Context

Need to select Go ecosystem components for the Loomio backend rewrite. Selection criteria:
1. **Maturity:** Production-proven, active maintenance
2. **Performance:** Suitable for concurrent API workload
3. **Developer experience:** Reasonable learning curve
4. **Ecosystem:** Good middleware and tooling

## Decisions

### Web Framework: Chi

**Selected:** `github.com/go-chi/chi/v5`

**Rationale:**
- Lightweight, composable middleware
- stdlib `net/http` compatible
- No magic, explicit routing
- Battle-tested at scale

**Alternatives considered:**
- Gin: More opinionated, larger dependency tree
- Echo: Similar to Chi, less stdlib-aligned
- Fiber: fasthttp-based, not net/http compatible

### Database: sqlc + pgx

**Selected:** `github.com/sqlc-dev/sqlc` with `github.com/jackc/pgx/v5`

**Rationale:**
- Type-safe SQL, generated Go code
- Full PostgreSQL feature support (JSONB, arrays, citext)
- No ORM magic, explicit queries
- pgx is fastest PostgreSQL driver

**Alternatives considered:**
- GORM: Magic-heavy, harder to debug complex queries
- sqlx: Good but manual, sqlc generates code
- ent: Schema-first, larger learning curve

### Authentication: Custom JWT + bcrypt

**Selected:** Custom implementation with:
- `github.com/golang-jwt/jwt/v5` for tokens
- `golang.org/x/crypto/bcrypt` for passwords

**Rationale:**
- Loomio has specific auth flows (magic links, SAML, OAuth)
- Custom allows exact feature parity
- Simpler than adapting full auth framework

### Background Jobs: River

**Selected:** `github.com/riverqueue/river`

**Rationale:**
- PostgreSQL-backed (no Redis required for jobs)
- Transactional job enqueuing
- Type-safe job definitions
- Active development, good docs

**Alternatives considered:**
- Asynq: Redis-backed, adds dependency
- Machinery: Heavier, more complex

### Testing: testify

**Selected:** `github.com/stretchr/testify`

**Rationale:**
- De facto standard for assertions
- Suite support for setup/teardown
- Mock generation
- Familiar to most Go developers

## Summary Table

| Category | Selection | Package |
|----------|-----------|---------|
| Web Framework | Chi | github.com/go-chi/chi/v5 |
| Database | sqlc + pgx | github.com/sqlc-dev/sqlc |
| Auth | Custom JWT | github.com/golang-jwt/jwt/v5 |
| Background Jobs | River | github.com/riverqueue/river |
| Testing | testify | github.com/stretchr/testify |
| WebSocket | websocket | nhooyr.io/websocket |
| Redis | go-redis | github.com/redis/go-redis/v9 |
| Logging | slog | stdlib (log/slog) |
| Config | env | github.com/caarlos0/env/v10 |

## Consequences

**Positive:**
- Minimal dependencies, stdlib-aligned
- Type-safe database access
- PostgreSQL-native job queue reduces infrastructure

**Negative:**
- More boilerplate than full frameworks
- sqlc requires writing SQL (learning curve for Rails developers)
```

**Step 2: Verify selections**

- [ ] Each selection has clear rationale
- [ ] Alternatives are documented
- [ ] Matches Discovery Report's gem→Go mapping needs

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-adr-003-go-stack-selection.md
git commit -m "docs: add ADR-003 Go technology stack selection"
```

---

## Task 4: Risk Register

**Files:**
- Create: `docs/plans/2026-01-29-risk-register.md`

**Step 1: Create risk register with top 20 risks**

```markdown
# Loomio Rewrite Risk Register

**Date:** 2026-01-29
**Status:** Active

## Risk Matrix

| Impact ↓ / Probability → | Low | Medium | High |
|--------------------------|-----|--------|------|
| **High** | Monitor | Mitigate | Prevent |
| **Medium** | Accept | Monitor | Mitigate |
| **Low** | Accept | Accept | Monitor |

## Top 10 Risks (Priority Order)

### Risk 1: Data Migration Failure
**Probability:** Medium | **Impact:** High | **Action:** Mitigate

**Description:** Data loss or corruption during PostgreSQL schema migration or data transformation.

**Mitigation:**
1. Develop migration scripts with rollback capability
2. Test migrations on production data copy
3. Run parallel systems during transition
4. Maintain backup before each migration step

**Early Warning:** Migration script failures in staging

---

### Risk 2: API Contract Drift
**Probability:** High | **Impact:** High | **Action:** Prevent

**Description:** Go API responses differ from Rails, breaking Vue frontend.

**Mitigation:**
1. Generate OpenAPI spec from Rails serializers
2. Contract testing against spec in Go CI
3. Shadow traffic comparison before cutover
4. Frontend integration tests in CI

**Early Warning:** Contract test failures

---

### Risk 3: Event Threading Logic Errors
**Probability:** Medium | **Impact:** High | **Action:** Mitigate

**Description:** The Event model's tree structure and position_key logic is complex. Incorrect reimplementation breaks comment threading.

**Mitigation:**
1. Extract all Event logic into dedicated document
2. Write comprehensive test cases before implementation
3. Side-by-side comparison testing with production data

**Early Warning:** Thread ordering differences in tests

---

### Risk 4: Permission System Gaps
**Probability:** Medium | **Impact:** High | **Action:** Mitigate

**Description:** CanCanCan ability rules are complex. Missing or incorrect permissions could expose private data.

**Mitigation:**
1. Extract complete ability matrix from Rails
2. Security-focused code review for all permission code
3. Permission test suite covering all endpoints
4. Security audit before production

**Early Warning:** Permission test failures

---

### Risk 5: Real-time Feature Degradation
**Probability:** Medium | **Impact:** Medium | **Action:** Monitor

**Description:** WebSocket reliability or performance worse than current Node.js implementation.

**Mitigation:**
1. Load test WebSocket implementation early
2. Keep Node.js channel server as fallback
3. Gradual rollout with monitoring

**Early Warning:** Latency metrics, connection drops

---

### Risk 6: Team Go Learning Curve
**Probability:** Medium | **Impact:** Medium | **Action:** Mitigate

**Description:** Team unfamiliar with Go patterns, leading to non-idiomatic code or slow velocity.

**Mitigation:**
1. Go training before implementation starts
2. Pair programming with Go-experienced developers
3. Code review focused on Go idioms
4. Start with simpler modules

**Early Warning:** Velocity slowdown, review feedback

---

### Risk 7: PostgreSQL Feature Incompatibility
**Probability:** Low | **Impact:** High | **Action:** Monitor

**Description:** Some PostgreSQL features (JSONB, citext, arrays) harder to work with in Go than ActiveRecord.

**Mitigation:**
1. sqlc handles most PostgreSQL types
2. pgx supports all PostgreSQL features
3. Early POC for complex queries

**Early Warning:** Compile errors in sqlc, runtime type errors

---

### Risk 8: Email Integration Complexity
**Probability:** Medium | **Impact:** Medium | **Action:** Monitor

**Description:** ActionMailbox email parsing is complex. Reply-by-email could break.

**Mitigation:**
1. Document all email parsing rules
2. Comprehensive test suite for email parsing
3. Keep Haraka SMTP server, change only handler

**Early Warning:** Parsing failures in staging

---

### Risk 9: Timeline Overrun
**Probability:** High | **Impact:** Medium | **Action:** Accept

**Description:** Estimates based on incomplete understanding; actual work exceeds plan.

**Mitigation:**
1. 30-50% buffer in estimates
2. Prioritize core features for MVP
3. Regular scope reviews
4. Adjust timeline rather than cut quality

**Early Warning:** Sprint velocity below plan

---

### Risk 10: User Adoption Resistance
**Probability:** Low | **Impact:** Medium | **Action:** Monitor

**Description:** Users notice bugs or behavior changes, lose trust.

**Mitigation:**
1. Extensive beta testing period
2. Clear communication about rewrite
3. Quick bug response during rollout
4. Easy rollback capability

**Early Warning:** Support ticket volume, community feedback

---

## Risks 11-20 (Summary)

| # | Risk | P | I | Action |
|---|------|---|---|--------|
| 11 | Third-party integration breaks (OAuth, S3) | L | M | Test early |
| 12 | Search functionality regression | M | M | Benchmark pg_search |
| 13 | Internationalization issues | L | M | Preserve i18n keys |
| 14 | Background job reliability | M | M | River testing |
| 15 | Memory/performance regression | M | M | Benchmark vs Rails |
| 16 | SAML SSO complexity | M | M | Early POC |
| 17 | File upload handling | L | L | ActiveStorage patterns |
| 18 | Soft delete consistency | L | M | Global middleware |
| 19 | Audit logging gaps | L | M | Match paper_trail |
| 20 | Documentation lag | M | L | Doc alongside code |

## Review Schedule

- **Weekly:** Review top 10 risks
- **Bi-weekly:** Full register review
- **Monthly:** Risk retrospective
```

**Step 2: Verify risk register**

- [ ] Top 10 risks have full detail
- [ ] Probabilities and impacts are realistic
- [ ] Mitigations are actionable
- [ ] Connects to Discovery Report findings

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-risk-register.md
git commit -m "docs: add risk register with top 20 risks"
```

---

## Task 5: Testing Strategy Framework

**Files:**
- Create: `docs/plans/2026-01-29-testing-strategy.md`

**Step 1: Create testing strategy document**

```markdown
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
```

**Step 2: Verify testing strategy**

- [ ] Covers all test types needed
- [ ] Tools are Go ecosystem standard
- [ ] Coverage targets are realistic
- [ ] CI pipeline is defined

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-testing-strategy.md
git commit -m "docs: add testing strategy framework"
```

---

## Task 6: Team & Resource Planning Template

**Files:**
- Create: `docs/plans/2026-01-29-team-resource-plan.md`

**Step 1: Create team planning template**

```markdown
# Loomio Rewrite Team & Resource Plan

**Date:** 2026-01-29
**Status:** Template (to be filled with actual team data)

## Required Skill Sets

| Skill | Required Level | Purpose |
|-------|----------------|---------|
| Go | Intermediate+ | Backend development |
| PostgreSQL | Intermediate | Database work, migrations |
| Ruby/Rails | Basic | Understanding existing code |
| Vue.js | Basic | Frontend debugging |
| WebSocket | Basic | Real-time features |
| DevOps | Intermediate | CI/CD, deployment |

## Team Composition Options

### Option A: Dedicated Rewrite Team (Recommended)

| Role | Count | Responsibility |
|------|-------|----------------|
| Tech Lead | 1 | Architecture, decisions, code review |
| Senior Go Dev | 2 | Core implementation |
| Mid Go Dev | 2 | Feature implementation |
| QA Engineer | 1 | Testing strategy, automation |
| DevOps | 0.5 | CI/CD, deployment |

**Total:** 6.5 FTE
**Duration:** 12-18 months

### Option B: Part-Time Migration

| Role | Allocation | Responsibility |
|------|------------|----------------|
| Existing team | 50% | Rewrite alongside maintenance |

**Total:** Varies by team size
**Duration:** 18-24 months (longer due to context switching)

### Option C: External Augmentation

| Role | Source | Responsibility |
|------|--------|----------------|
| Go consultants | Contract | Initial architecture, training |
| Existing team | Internal | Implementation after training |

**Total:** 2 contractors + internal team
**Duration:** 14-18 months

## Training Plan

### Phase 1: Go Fundamentals (Week 1-2)
- Tour of Go (all team members)
- Effective Go reading
- Go by Example exercises
- Internal code review sessions

### Phase 2: Stack Training (Week 3-4)
- Chi routing patterns
- sqlc + pgx database patterns
- River job processing
- Testing with testify

### Phase 3: Loomio Domain (Week 5-6)
- Rails codebase walkthrough
- Discovery Report review
- Domain model deep-dive
- Permission system review

## Resource Requirements

### Development Environment
- [ ] Go 1.22+ installed
- [ ] PostgreSQL 17 local instance
- [ ] Redis local instance
- [ ] Docker for testcontainers
- [ ] IDE with Go support (GoLand, VS Code)

### Infrastructure
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Staging environment (mirrors production)
- [ ] Load testing environment
- [ ] Anonymized production data copy

### Budget Considerations

| Category | Estimate | Notes |
|----------|----------|-------|
| Personnel | $X/month | Based on team composition |
| Infrastructure | $X/month | Staging, CI minutes |
| Training | $X one-time | Courses, materials |
| Tools | $X/month | IDE licenses, monitoring |

## Communication Plan

| Meeting | Frequency | Participants | Purpose |
|---------|-----------|--------------|---------|
| Daily standup | Daily | Dev team | Progress, blockers |
| Tech review | Weekly | Tech lead + seniors | Architecture decisions |
| Stakeholder update | Bi-weekly | All + stakeholders | Progress report |
| Retrospective | Per sprint | Dev team | Process improvement |

## Decision Authority

| Decision Type | Authority | Escalation |
|---------------|-----------|------------|
| Code style | Any developer | Tech lead |
| Package selection | Tech lead | Team consensus |
| Architecture | Tech lead + seniors | ADR process |
| Timeline changes | Tech lead | Stakeholders |
| Scope changes | Stakeholders | - |
```

**Step 2: Verify template**

- [ ] Covers all planning aspects
- [ ] Options provide flexibility
- [ ] Training plan is realistic
- [ ] Communication plan is clear

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-team-resource-plan.md
git commit -m "docs: add team and resource planning template"
```

---

## Task 7: Deployment & Rollout Plan

**Files:**
- Create: `docs/plans/2026-01-29-deployment-rollout-plan.md`

**Step 1: Create deployment plan**

```markdown
# Loomio Go Deployment & Rollout Plan

**Date:** 2026-01-29
**Status:** Proposed

## Current Deployment Architecture

From Discovery Report, Loomio uses Docker Compose on single host:
- nginx-proxy (SSL termination)
- app (Rails/Puma)
- worker (Sidekiq)
- channels (Socket.io)
- hocuspocus (Y.js)
- db (PostgreSQL 17)
- redis

## Target Architecture (Go)

### Phase 1: Parallel Systems

```
┌─────────────────────────────────────────────────┐
│                  nginx-proxy                     │
│         (routes based on path/feature flag)      │
└─────────────────┬───────────────────────────────┘
                  │
      ┌───────────┼───────────┐
      │           │           │
      ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────────┐
│  Rails  │ │   Go    │ │  channels   │
│  (old)  │ │  (new)  │ │  (Go/Node)  │
└────┬────┘ └────┬────┘ └─────────────┘
     │           │
     └─────┬─────┘
           ▼
     ┌──────────┐
     │ PostgreSQL│
     │  (shared) │
     └──────────┘
```

### Phase 2: Go Primary

```
┌─────────────────────────────────────────────────┐
│                  nginx-proxy                     │
└─────────────────┬───────────────────────────────┘
                  │
      ┌───────────┼───────────┐
      │           │           │
      ▼           ▼           ▼
┌─────────┐ ┌─────────────┐ ┌────────────┐
│   Go    │ │  channels   │ │ hocuspocus │
│   API   │ │    (Go)     │ │  (Node.js) │
└────┬────┘ └──────┬──────┘ └────────────┘
     │             │
     └──────┬──────┘
            ▼
     ┌──────────┐
     │ PostgreSQL│
     └──────────┘
```

## Rollout Strategy

### Stage 1: Shadow Mode (2 weeks)
- Deploy Go alongside Rails
- Mirror traffic to Go (write to logs only)
- Compare responses for discrepancies
- No user impact

### Stage 2: Canary (2 weeks)
- Route 1% of read-only traffic to Go
- Monitor error rates and latency
- Automatic rollback on anomalies

### Stage 3: Gradual Rollout (4 weeks)
- Increase Go traffic: 1% → 10% → 50% → 100%
- Each increase after 1 week of stability
- Maintain Rails as hot standby

### Stage 4: Rails Sunset (2 weeks)
- Go handles 100% of traffic
- Rails kept running but not receiving traffic
- Final data verification
- Rails containers removed

## Feature Flag Strategy

Use environment-based routing in nginx:

```nginx
# Example: Route /api/v1/groups to Go
location /api/v1/groups {
    if ($go_enabled = "true") {
        proxy_pass http://go-api:8080;
    }
    proxy_pass http://rails-app:3000;
}
```

Feature flags in database for fine-grained control:
- Per-endpoint toggles
- Per-user toggles (for beta testers)
- Percentage-based rollout

## Monitoring Requirements

### Metrics to Track
- Request latency (p50, p95, p99)
- Error rate (5xx, 4xx)
- Database query time
- WebSocket connection count
- Memory and CPU usage

### Alerting Thresholds
- Error rate > 1%: Page on-call
- p95 latency > 500ms: Warning
- p99 latency > 2s: Page on-call
- Memory > 80%: Warning

### Dashboards
- Side-by-side Rails vs Go metrics
- User-facing error tracking (Sentry)
- Database performance

## Rollback Procedures

### Automatic Rollback Triggers
- Error rate > 5% for 5 minutes
- p99 latency > 5s for 5 minutes
- Health check failures

### Manual Rollback Steps
1. Set feature flag to route all traffic to Rails
2. Verify Rails health
3. Investigate Go issues
4. Do not remove Go containers (for debugging)

### Rollback Time Target
- Automatic: < 1 minute
- Manual: < 5 minutes

## Infrastructure Changes

### New Containers
- `loomio/loomio-go:latest` - Go API server
- `loomio/channels-go:latest` - Go channel server (records + bots)

### Retained Containers
- `loomio/loomio_channel_server` - Hocuspocus only

### Removed Containers (after full migration)
- `loomio/loomio` - Rails app
- Sidekiq worker (replaced by River in Go)

## Database Migration

### Schema Changes
- Minimal schema changes needed
- Go uses same PostgreSQL database
- sqlc generates types from existing schema

### Data Migration
- No data migration needed (same schema)
- Application-level compatibility only

## Self-Hosted User Communication

1. **Announcement:** 3 months before release
2. **Beta period:** Opt-in for testing
3. **Release notes:** Detailed upgrade guide
4. **Docker image tags:** Maintain Rails image for 6 months
5. **Support:** Dedicated channel for migration issues
```

**Step 2: Verify deployment plan**

- [ ] Builds on Discovery Report deployment analysis
- [ ] Rollout stages are conservative
- [ ] Rollback procedures are clear
- [ ] Monitoring is comprehensive

**Step 3: Commit**

```bash
git add docs/plans/2026-01-29-deployment-rollout-plan.md
git commit -m "docs: add deployment and rollout plan"
```

---

## Task 8: Update CLAUDE.md and Cross-References

**Files:**
- Modify: `CLAUDE.md`
- Modify: `META_PLAN.md` (add references to new documents)

**Step 1: Update CLAUDE.md progress tracker**

Change Phase 2 status from "Next" to "Complete" and add document references.

**Step 2: Add cross-references in META_PLAN.md**

Add links to new documents in Phase 2 section.

**Step 3: Commit**

```bash
git add CLAUDE.md META_PLAN.md
git commit -m "docs: update progress tracker and cross-references for Phase 2"
```

---

## Summary

| Task | Deliverable | Est. Effort |
|------|-------------|-------------|
| 1 | ADR-001: Migration Strategy | 15 min |
| 2 | ADR-002: Channel Server | 15 min |
| 3 | ADR-003: Go Stack Selection | 15 min |
| 4 | Risk Register | 20 min |
| 5 | Testing Strategy | 15 min |
| 6 | Team Resource Plan | 15 min |
| 7 | Deployment Rollout Plan | 20 min |
| 8 | Cross-reference updates | 10 min |

**Total:** ~2 hours of focused work

---

## Verification Checklist

After completing all tasks:

- [ ] All 7 new documents created in `docs/plans/`
- [ ] Each document has clear structure and actionable content
- [ ] ADRs follow standard format (Context, Decision, Consequences)
- [ ] Risk register has probability/impact ratings
- [ ] All documents committed to git
- [ ] CLAUDE.md progress updated
- [ ] META_PLAN.md has cross-references
