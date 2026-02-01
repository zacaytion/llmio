# Final Synthesis: Confirmed Findings for Go Rewrite

**Generated:** 2026-02-01
**Status:** Implementation-ready

---

## Purpose

This directory contains confirmed findings that align between our original research and the third-party discovery. These documents are **implementation-ready** and provide the definitive reference for the Go rewrite.

---

## Contents

### Third-Party Comparison (Original 7 Topics)

| File | Topic | Key Implementation Details |
|------|-------|---------------------------|
| [oauth_providers.md](oauth_providers.md) | OAuth Provider Configuration | 4 providers (Google, Generic OAuth2, SAML, Nextcloud), provider interface |
| [oauth_security.md](oauth_security.md) | OAuth Security Patterns | Session management, identity linking, **CSRF fix required** |
| [permission_flags.md](permission_flags.md) | Authorization Flags | 12 active flags, ability check patterns, admin override logic |
| [rate_limiting.md](rate_limiting.md) | Rate Limit Implementation | 3-layer architecture, Redis key patterns, proper 429 responses |
| [attachments_jsonb.md](attachments_jsonb.md) | JSONB Attachment Schema | Array default `[]`, 8 tables, attachment object structure |
| [stance_revision.md](stance_revision.md) | Voting Revision Logic | 15-minute threshold, 4-condition decision tree, partial unique index |
| [webhook_events.md](webhook_events.md) | Webhook Event System | 14 eligible events, 5 payload formats, delivery via Sidekiq |
| [CROSS_CUTTING_ANALYSIS.md](CROSS_CUTTING_ANALYSIS.md) | System-Wide Patterns | Implementation sequence, shared infrastructure, risk assessment |

### Additional Topics (From Baseline Research)

| File | Topic | Key Implementation Details |
|------|-------|---------------------------|
| [realtime_pubsub.md](realtime_pubsub.md) | Real-Time Pub/Sub | MessageChannelService, Redis channels, Hocuspocus auth |
| [search_indexing.md](search_indexing.md) | Full-Text Search | pg_search_documents, Searchable concern, query service |
| [infrastructure_jobs.md](infrastructure_jobs.md) | Background Jobs | 38 workers, River jobs, scheduled tasks |
| [counter_caches.md](counter_caches.md) | Counter Caches | 40+ counters, reconciliation strategy |

---

## Implementation Phases

Based on the cross-cutting analysis, implement in this order:

| Phase | Focus | Duration | Dependencies |
|-------|-------|----------|--------------|
| 1 | Core Infrastructure | 2 weeks | None |
| 2 | Authentication (OAuth) | 2 weeks | Phase 1 |
| 3 | Authorization (Permissions) | 2 weeks | Phase 2 |
| 4 | Rate Limiting | 1 week | Phase 1 |
| 5 | Voting & Stances | 1 week | Phase 3 |
| 6 | Webhooks | 2 weeks | Phase 5 |

---

## Critical Security Improvements

The Go implementation **MUST** address these gaps in the original Rails codebase:

### 1. OAuth State Parameter (CSRF Protection)
```go
// Generate state on initiate
state := base64.URLEncoding.EncodeToString(randomBytes(32))
session.Set("oauth_state", state)

// Validate state on callback
if r.URL.Query().Get("state") != session.Get("oauth_state") {
    return ErrInvalidOAuthState
}
```

### 2. Rate Limit Response Codes
```go
// Return proper 429 (not 500)
w.Header().Set("Retry-After", "60")
http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
```

### 3. Redis Key TTL
```go
// Always set expiration on rate limit keys
redis.SetEx(ctx, key, value, time.Hour)
```

---

## Shared Go Types

These types are used across multiple topics and should be defined once:

```go
// internal/types/ids.go
type UserID int64
type GroupID int64
type DiscussionID int64
type PollID int64
type StanceID int64
type ChatbotID int64
type EventID int64

// internal/types/jsonb.go
type Attachments []Attachment      // Default: []
type OptionScores map[string]int   // Default: {}
type CustomFields map[string]any   // Default: {}
```

---

## API Compatibility Checklist

The Vue frontend expects these specific behaviors:

- [ ] `attachments` fields serialize as `[]` (never `null`)
- [ ] `option_scores` fields serialize as `{}` (never `null`)
- [ ] All 12 permission flags present in group serializer
- [ ] `identityProviders` array in boot payload
- [ ] Error responses use `{"errors": {...}}` format
- [ ] Rate limit returns 429 with `Retry-After` header

---

## How to Use This Directory

1. **Reference during implementation:** Each file contains Go code examples
2. **Verify API compatibility:** Compare serializer output against Rails
3. **Track completion:** Mark items as implemented
4. **Update as questions resolve:** Incorporate answers from `final_follow_up/`

---

## Confidence Level

All findings in this directory are **HIGH CONFIDENCE** based on:
- Source code evidence with file/line references
- Migration history verification
- Cross-validation between our research and third-party discovery
- Schema dump confirmation
