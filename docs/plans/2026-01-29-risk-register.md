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
