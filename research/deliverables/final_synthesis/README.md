# Final Synthesis: Confirmed Findings

**Generated:** 2026-02-01
**Status:** Implementation-ready

---

## Purpose

This directory contains confirmed findings that align between our original research and the third-party discovery. These documents provide the definitive reference for understanding the Loomio codebase architecture and behavior.

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
| [infrastructure_jobs.md](infrastructure_jobs.md) | Background Jobs | 38 workers, scheduled tasks |
| [counter_caches.md](counter_caches.md) | Counter Caches | 40+ counters, reconciliation strategy |

---

## Critical Security Improvements

The implementation **MUST** address these gaps in the original Rails codebase:

### 1. OAuth State Parameter (CSRF Protection)

The current implementation lacks state parameter validation. Any reimplementation must:
- Generate secure random state on OAuth initiation
- Store state in session
- Validate state on callback before processing

### 2. Rate Limit Response Codes

The ThrottleService returns 500 instead of proper 429 responses. Correct implementation should:
- Return 429 Too Many Requests
- Include `Retry-After` header

### 3. Redis Key TTL

Rate limit keys may not expire. Always set TTL on Redis keys.

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

1. **Reference during implementation:** Each file contains detailed specifications
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
