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
