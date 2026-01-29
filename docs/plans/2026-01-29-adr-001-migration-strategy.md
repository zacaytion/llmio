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
