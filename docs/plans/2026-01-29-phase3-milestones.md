# Loomio Go Rewrite: Milestone Definitions

**Date:** 2026-01-29
**Status:** Proposed
**Phase:** 3 - Detailed Plan Creation

## Document Purpose

This document defines key milestones for the Loomio Go rewrite project. Each milestone includes:
- Target week/date
- Success criteria (measurable)
- Deliverables
- Go/No-Go decision criteria
- Dependencies

**Project start assumption:** Week 1 = T+0 (to be filled in when project kicks off)
**Total timeline:** 40 weeks (~10 months)

---

## Milestone Overview

```
Week    0         4         8        18        26        34        38    40
        │         │         │         │         │         │         │     │
        ▼         ▼         ▼         ▼         ▼         ▼         ▼     ▼
        ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬─────┐
        │   M1    │   M2    │   M3    │   M4    │   M5    │   M6    │ M7  │
        │Foundation│ Channel │Read API │ Write  │ Full    │ Beta   │Launch│
        │Complete │ Server  │Complete │ Ops    │ Feature │ Ready  │     │
        └─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────┘
```

---

## M1: Foundation Complete

**Target:** Week 4
**Phase:** 0 (Foundation & Infrastructure)

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| CI Pipeline | All workflows green | 100% pass |
| Code Coverage | Unit test coverage | > 70% |
| Database Layer | sqlc queries compile | All tables |
| HTTP Server | Health check responds | < 10ms p99 |
| Background Jobs | River processes test job | Success |

### Deliverables

- [ ] Go module with standard project structure
- [ ] GitHub Actions CI pipeline (lint, test, build)
- [ ] Docker image build and push
- [ ] PostgreSQL connection pool configured
- [ ] sqlc generating code for core tables
- [ ] Chi router with middleware stack
- [ ] River job queue operational
- [ ] Health check endpoints (/health, /ready)
- [ ] Structured logging with request IDs
- [ ] Configuration via environment variables

### Go/No-Go Criteria

**Go:** All criteria met, team comfortable with Go patterns
**No-Go conditions:**
- CI pipeline unstable
- Database connectivity issues
- Team blocked on Go learning curve

### Dependencies

- PostgreSQL 17 available
- Redis available
- GitHub Actions configured
- Docker registry access

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |

---

## M2: Channel Server Ready

**Target:** Week 8
**Phase:** 1 (Channel Server Migration)

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| WebSocket Connections | Concurrent connections | > 1000 |
| Message Latency | Pub/sub to client | < 50ms p95 |
| Matrix Bot | Test messages sent | Success |
| Load Test | vs Node.js baseline | >= 100% |
| Integration Tests | All channel tests | 100% pass |

### Deliverables

- [ ] WebSocket server (nhooyr.io/websocket)
- [ ] Room-based broadcasting (user-{id}, group-{id})
- [ ] Redis pub/sub subscription (/records channel)
- [ ] User authentication via Redis session lookup
- [ ] Matrix bot (mautrix/go) operational
- [ ] Connection management (join, leave, reconnect)
- [ ] Load test results documented
- [ ] Shadow deployment plan ready

### Go/No-Go Criteria

**Go:** Performance meets or exceeds Node.js, all tests pass
**No-Go conditions:**
- Performance regression vs Node.js
- WebSocket stability issues
- Matrix bot failures

### Dependencies

- M1 complete
- Redis pub/sub configured
- Matrix homeserver access (for bot testing)

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |

---

## M3: Read API Complete

**Target:** Week 18
**Phase:** 2 (Read-Only API Endpoints)

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| API Endpoints | Read endpoints implemented | 100% (30+) |
| Contract Tests | JSON response match | 100% pass |
| Response Time | API p95 latency | < 100ms |
| Auth Flows | Login methods working | All (password, magic link, OAuth) |
| Test Coverage | Unit + integration | > 80% |

### Deliverables

- [ ] Authentication (JWT, sessions, magic links)
- [ ] Users API (profile, search, detail)
- [ ] Groups API (list, detail, subgroups, memberships)
- [ ] Discussions API (list, detail, events/timeline)
- [ ] Polls API (list, detail, stances, outcomes)
- [ ] Supporting APIs (notifications, tags, translations)
- [ ] All serializers match Rails exactly
- [ ] Contract test suite complete
- [ ] Performance baseline documented

### Go/No-Go Criteria

**Go:** All read endpoints pass contract tests, frontend can render
**No-Go conditions:**
- Contract test failures
- Performance regressions
- Event threading incorrect

### Dependencies

- M1 complete
- Full database access
- Rails serializer documentation

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |
| QA Lead | | | |

---

## M4: Write Operations Complete

**Target:** Week 26
**Phase:** 3 (Write Endpoints by Domain)

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Write Endpoints | All write operations | 100% |
| Permission Coverage | Endpoints with auth | 100% |
| Voting Accuracy | All poll types correct | 100% |
| Notification Delivery | Events → notifications | Working |
| Test Coverage | Including write paths | > 80% |

### Deliverables

- [ ] Permission system (extracted and implemented)
- [ ] Users write (registration, profile update, delete)
- [ ] Groups write (create, update, memberships, archive)
- [ ] Discussions write (create, edit, comments with threading)
- [ ] Polls write (create all types, voting, outcomes)
- [ ] Notification generation and delivery
- [ ] Event position_key generation correct
- [ ] Permission test suite complete

### Go/No-Go Criteria

**Go:** Full CRUD functionality working, permissions enforced
**No-Go conditions:**
- Permission gaps (security risk)
- Voting algorithm errors
- Comment threading broken

### Dependencies

- M3 complete
- CanCanCan rules extracted
- Event threading logic documented

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |
| Security Lead | | | |

---

## M5: Full Feature Parity

**Target:** Week 34
**Phase:** 3 (continued) + partial Phase 4

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Feature Coverage | vs Rails functionality | 100% |
| Email System | All mailers working | 100% |
| File Handling | Upload/download working | All storage backends |
| Search | Full-text search accuracy | Rails parity |
| SSO | SAML + OAuth providers | All supported |

### Deliverables

- [ ] Email system (templates, transactional, notifications, digests)
- [ ] Inbound email parsing (ActionMailbox replacement)
- [ ] File uploads (S3, GCS compatible)
- [ ] Image processing (thumbnails, avatars)
- [ ] Full-text search (PostgreSQL FTS)
- [ ] SAML SSO implementation
- [ ] OAuth providers (Google, Facebook, etc.)
- [ ] All API endpoints complete
- [ ] Feature comparison document

### Go/No-Go Criteria

**Go:** No functional gaps vs Rails, all integrations working
**No-Go conditions:**
- Missing critical features
- Integration failures (email, storage, SSO)
- Performance regressions

### Dependencies

- M4 complete
- Email infrastructure (SMTP)
- Storage credentials (S3/GCS)
- IdP access for SSO testing

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |
| QA Lead | | | |

---

## M6: Beta Ready

**Target:** Week 38
**Phase:** 4 (Integration & Polish)

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Contract Tests | All endpoints | 100% pass |
| E2E Tests | Critical paths | 100% pass |
| Security Audit | OWASP checklist | No critical issues |
| Performance | vs Rails baseline | >= 100% |
| Data Migration | Dry run complete | Success |

### Deliverables

- [ ] Full API contract verification
- [ ] E2E test suite (Playwright or similar)
- [ ] Security audit report
- [ ] Performance benchmark results
- [ ] Data migration scripts tested
- [ ] Rollback procedures documented
- [ ] Monitoring dashboards configured
- [ ] Alerting thresholds set
- [ ] Shadow deployment plan finalized

### Go/No-Go Criteria

**Go:** System passes all verification, ready for real traffic
**No-Go conditions:**
- Security vulnerabilities
- Performance regressions
- Data migration failures
- Contract test failures

### Dependencies

- M5 complete
- Anonymized production data copy
- Staging environment matching production
- Security review resources

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |
| Security Lead | | | |
| Operations | | | |

---

## M7: Production Launch

**Target:** Week 40
**Phase:** 5 (Testing & Launch)

### Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Traffic Migration | Go handling traffic | 100% |
| Error Rate | Post-migration | < 0.1% |
| User Satisfaction | Support tickets | Baseline or better |
| Rails Sunset | Rails containers removed | Complete |
| Documentation | All guides updated | 100% |

### Deliverables

- [ ] Shadow mode deployment complete
- [ ] Canary deployment (1% traffic) stable
- [ ] Gradual rollout complete (1%→10%→50%→100%)
- [ ] Monitoring confirms stability
- [ ] Rails containers removed
- [ ] API documentation (OpenAPI)
- [ ] Deployment runbook
- [ ] Developer setup guide
- [ ] Self-hosted user migration guide
- [ ] Post-launch retrospective

### Go/No-Go Criteria

**Go:** Stable at each rollout stage, no user impact
**No-Go conditions:**
- Error rate spike
- Performance degradation
- User-reported issues
- Monitoring gaps

### Dependencies

- M6 complete
- Production deployment access
- On-call rotation established
- Communication plan for users

### Stakeholder Sign-off

| Role | Name | Decision | Date |
|------|------|----------|------|
| Tech Lead | | | |
| Product Owner | | | |
| Operations | | | |
| Executive Sponsor | | | |

---

## Milestone Risk Matrix

| Milestone | Primary Risks | Mitigation |
|-----------|---------------|------------|
| M1 | Team Go learning (Risk #6) | Training in first 2 weeks |
| M2 | Real-time degradation (Risk #5) | Keep Node.js fallback |
| M3 | Event threading (Risk #3) | Property-based testing |
| M4 | Permission gaps (Risk #4) | Security review gate |
| M5 | Email complexity (Risk #8) | Keep Haraka SMTP |
| M6 | Data migration (Risk #1) | Multiple dry runs |
| M7 | Timeline overrun (Risk #9) | 40% buffer in estimates |

---

## Timeline Visualization

```
2026-Q2        2026-Q3        2026-Q4        2027-Q1
Apr May Jun    Jul Aug Sep    Oct Nov Dec    Jan Feb Mar
 │   │   │      │   │   │      │   │   │      │   │   │
 │ M1│  M2│     │ M3│   │      │ M4│  M5│     │ M6│ M7│
 │ ▼ │  ▼ │     │ ▼ │   │      │ ▼ │  ▼ │     │ ▼ │ ▼ │
 └───┴────┴─────┴───┴───┴──────┴───┴────┴─────┴───┴───┘
 W1-4 W5-8      W9-18          W19-26 W27-34   W35-40
```

*(Actual dates to be determined at project kickoff)*

---

## Milestone Review Process

### At Each Milestone

1. **Review meeting** (2 hours)
   - Present deliverables
   - Review success criteria
   - Discuss issues and blockers
   - Go/No-Go decision

2. **Documentation update**
   - Update this document with actuals
   - Record lessons learned
   - Adjust future estimates if needed

3. **Stakeholder communication**
   - Send milestone report
   - Highlight risks and mitigations
   - Next milestone preview

### Go/No-Go Decision Authority

| Decision | Authority |
|----------|-----------|
| Go (all criteria met) | Tech Lead |
| Conditional Go (minor gaps) | Tech Lead + Product Owner |
| No-Go (blocking issues) | Escalate to Executive Sponsor |

---

## Appendix: Milestone Dependency Graph

```
M1 (Foundation)
 │
 ├──► M2 (Channel Server)
 │     │
 │     ▼
 │    [Can shadow deploy channel server]
 │
 └──► M3 (Read API)
       │
       ▼
      M4 (Write Operations)
       │
       ▼
      M5 (Feature Parity)
       │
       ▼
      M6 (Beta Ready)
       │
       ▼
      M7 (Production Launch)
```

Note: M2 can proceed in parallel with M3 after M1 is complete.
