# Final Follow-Up Questions for Third Party

**Generated:** 2026-02-01
**Status:** Ready for third-party review

---

## Purpose

This directory contains questions, discrepancies, and clarification requests identified by comparing third-party discovery findings against our original research. These documents represent **the last opportunity** to ask the third party for additional information before they deliver the OpenAPI spec.

---

## Contents

### Third-Party Comparison (Original 7 Topics)

| File | Topic | Priority Issues |
|------|-------|-----------------|
| [oauth_providers.md](oauth_providers.md) | OAuth Provider Configuration | SSO-only mode flow, pending identity resolution |
| [oauth_security.md](oauth_security.md) | OAuth Security Patterns | CSRF vulnerability, token refresh behavior |
| [permission_flags.md](permission_flags.md) | Authorization Flags | `members_can_add_guests` paper_trail exclusion, Null::Group contradiction |
| [rate_limiting.md](rate_limiting.md) | Rate Limit Implementation | ThrottleService counter expiration, cache store configuration |
| [attachments_jsonb.md](attachments_jsonb.md) | JSONB Attachment Schema | Attachment element validation |
| [stance_revision.md](stance_revision.md) | Voting Revision Logic | Standalone poll behavior, unused query in service |
| [webhook_events.md](webhook_events.md) | Webhook Event System | HMAC signatures, retry configuration, Matrix caching |
| [CROSS_CUTTING_ANALYSIS.md](CROSS_CUTTING_ANALYSIS.md) | System-Wide Patterns | Security control interactions, Redis TTL management |

### Additional Topics (Not in Third-Party Follow-Up)

| File | Topic | Priority Issues |
|------|-------|-----------------|
| [realtime_pubsub.md](realtime_pubsub.md) | Real-Time Pub/Sub | Event → pub/sub mapping, room routing logic |
| [search_indexing.md](search_indexing.md) | Full-Text Search | pg_search population, reindex triggers |
| [infrastructure_jobs.md](infrastructure_jobs.md) | Background Jobs | 38 workers, retry config, scheduled tasks |
| [missing_features.md](missing_features.md) | Omitted Features | Email, storage, demo, billing, tasks, mentions |

---

## Critical Questions (Must Resolve Before Implementation)

These are the highest-priority questions extracted from all topic analyses:

### 1. OAuth CSRF Vulnerability [HIGH - Security]
**Question:** Should the Go implementation fix the missing OAuth state parameter validation?
**Impact:** Account hijacking risk if not addressed
**Reference:** `app/controllers/identities_controller.rb`

### 2. ThrottleService Counter Expiration [HIGH - Operations]
**Question:** How do ThrottleService Redis counters expire? Is there a scheduled job that calls `ThrottleService.reset!`?
**Impact:** Potential unbounded memory growth
**Reference:** `app/services/throttle_service.rb`

### 3. SSO-Only Mode Flow [HIGH - Core Feature]
**Question:** When email login is disabled via `disableEmailLogin`, what is the complete flow for pending identities?
**Impact:** Blocks understanding of core onboarding path
**Reference:** `app/controllers/identities_controller.rb`, `app/models/login_token.rb`

---

## How to Use This Directory

1. **Third-party review:** Send these documents to the third party for their response
2. **Prioritize by criticality:** Focus on HIGH priority items first
3. **Track resolutions:** Update documents as answers are received
4. **Feed into implementation:** Resolved questions should be incorporated into `final_synthesis/`

---

## Confidence Assessment

Areas where third-party confidence levels may be overstated:
- Rate Limiting (claimed 5/5) - Counter expiration unverified
- OAuth Security (claimed 5/5) - Token refresh behavior unclear
- Webhook Events (claimed 5/5) - HMAC and retry config unverified

Areas where confidence is well-justified:
- OAuth Providers (5/5) - Definitive source code evidence
- Attachments JSONB (5/5) - Clear migration history
- Permission Flags (5/5) - Complete enumeration with line references
