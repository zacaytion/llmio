# Cross-Cutting Follow-up Analysis

**Generated:** 2026-02-01
**Topics Analyzed:** 7 comparisons across OAuth, Authorization, Rate Limiting, Data Models, and Events

---

## Executive Summary

This analysis identifies patterns that span multiple topic areas, revealing systemic gaps, dependencies, and priority questions that require resolution before Go implementation. The most critical finding is that **security-related topics share common architectural gaps** that compound risk when considered together.

---

## Cross-Cutting Discrepancy Patterns

### Pattern 1: Security Controls Operating in Isolation

**Affected Topics:** OAuth Security, Rate Limiting, Permission Flags

**Observation:** Security mechanisms exist but do not interact coherently:

| Mechanism | Scope | Gap |
|-----------|-------|-----|
| OAuth authentication | Identity providers | No CSRF protection (state parameter missing) |
| Rate limiting (Rack::Attack) | IP-based on endpoints | Does not consider authenticated user |
| Rate limiting (ThrottleService) | Per-user for invitations | Returns 500 instead of 429 |
| Permission flags | Group-level abilities | No audit trail for permission changes |

**Cross-cutting questions for third party:**
1. How do permission flags interact with OAuth/SSO flows? Can SSO users bypass permission restrictions?
2. Is there any connection between identity provider and rate limiting (e.g., trusted SSO users get higher limits)?
3. When a user's permissions change mid-session, how is this reflected in already-issued session tokens?

**Priority:** HIGH - These gaps may compound into privilege escalation or denial of service vulnerabilities.

---

### Pattern 2: Missing Error Standardization

**Affected Topics:** Rate Limiting, OAuth Security

**Observation:** Error handling is inconsistent across authentication and rate limiting:

| Error Condition | Current Response | Expected Response |
|-----------------|------------------|-------------------|
| Rate limit exceeded (Rack::Attack) | 429 with body "Retry later\n" | 429 with Retry-After header |
| Rate limit exceeded (ThrottleService) | 500 Internal Server Error | 429 with Retry-After header |
| OAuth CSRF attack | Silently links account | 403 Forbidden |
| Invalid OAuth state (if existed) | N/A (no state validation) | 403 with error message |

**Cross-cutting questions for third party:**
1. Was the 500 response from ThrottleService::LimitReached intentional (fail-closed) or a bug?
2. Are there frontend handlers for these error conditions? How does the Vue app respond to 500 vs 429?
3. Should the Go implementation standardize all rate limit responses?

**Priority:** MEDIUM - Affects user experience and debugging, but not security directly.

---

### Pattern 3: Redis Key Lifecycle Ambiguity

**Affected Topics:** Rate Limiting, OAuth Security (session storage)

**Observation:** Redis is used for multiple purposes with unclear TTL management:

| Usage | Redis Key Pattern | TTL | Management |
|-------|-------------------|-----|------------|
| Rate limiting (ThrottleService) | `THROTTLE-{PERIOD}-{key}-{id}` | **Unknown** | Manual reset via `ThrottleService.reset!` |
| Rate limiting (Rack::Attack) | Uses Rails.cache | Rack::Attack defaults | Automatic |
| Session storage | Unknown pattern | Session timeout | Unknown |
| OAuth state | **Not implemented** | N/A | N/A |

**Cross-cutting questions for third party:**
1. Is there a scheduled job that calls `ThrottleService.reset!`? If not, do counters persist indefinitely?
2. What is the actual cache store for Rack::Attack - is it shared with session storage?
3. When implementing OAuth state in Go, what TTL should be used?

**Priority:** MEDIUM - Memory management concern, not immediate functionality issue.

---

### Pattern 4: JSONB Schema Consistency

**Affected Topics:** Attachments JSONB, Stance Revision, Permission Flags

**Observation:** JSONB columns follow consistent patterns that should be preserved:

| Table | JSONB Column | Default | Structure |
|-------|--------------|---------|-----------|
| 8 tables | `attachments` | `[]` array | Array of file objects |
| `stances` | `option_scores` | `{}` object | Map of option_id -> score |
| `omniauth_identities` | `custom_fields` | `{}` object | Provider-specific data |

**Consistency finding:** Attachments use ARRAY, while other JSONB fields use OBJECT. This is intentional and should be preserved.

**Cross-cutting question for third party:**
1. Is there any validation that `attachments` elements conform to a schema, or is it free-form?

**Priority:** LOW - Well-documented, implementation-ready.

---

### Pattern 5: paper_trail Versioning Gaps

**Affected Topics:** Stance Revision, Permission Flags

**Observation:** paper_trail is used for audit trails but inconsistently:

| Model | paper_trail Used | Tracked Fields | Gap |
|-------|------------------|----------------|-----|
| Stance | Yes | `reason`, `option_scores`, `revoked_at`, `revoker_id`, `inviter_id`, `attachments` | Separate from `latest` flag mechanism |
| Group | Yes | Most permission flags | `members_can_add_guests` NOT tracked |
| Identity | Unknown | Unknown | Not investigated |

**Cross-cutting questions for third party:**
1. Why is `members_can_add_guests` excluded from paper_trail tracking on Group?
2. Should the Go implementation implement a unified audit system that covers both stance revisions and permission changes?
3. Is there a strategy for migrating paper_trail data to Go?

**Priority:** MEDIUM - Audit trail is important for compliance, but functionality works without it.

---

## Dependency Chains Requiring Clarification

### Chain 1: OAuth -> Permissions -> Webhooks

```
OAuth Login
    |
    v
Identity Created/Linked
    |
    v
User gets Membership in Group
    |
    v
Membership has Role (admin/member/guest)
    |
    v
Permission Flags control capabilities
    |
    v
Event published (e.g., new_discussion)
    |
    v
Webhook delivered (if chatbot subscribed)
```

**Unanswered questions:**
1. When a new user signs in via SSO and auto-joins a group, does this trigger a webhook event?
2. If a user is removed from a group, are pending webhook deliveries for their events cancelled?
3. How do guest permissions interact with webhook notifications?

### Chain 2: Rate Limiting -> API Calls -> Events -> Webhooks

```
API Request
    |
    v
Rack::Attack (IP-based)
    |
    v
Controller Action
    |
    v
ThrottleService (operation-specific)
    |
    v
Event Published
    |
    v
ChatbotService delivers webhook
```

**Unanswered questions:**
1. If a user hits a rate limit, is there any event published for monitoring?
2. Are webhook deliveries themselves rate limited? What prevents a flood of events from overwhelming external services?
3. Is there any backpressure mechanism if Sidekiq queues grow too large?

---

## Priority-Ranked Follow-up Questions for Third Party

### Critical (Must resolve before implementation)

| # | Question | Topics Affected | Why Critical |
|---|----------|-----------------|--------------|
| 1 | **Should the Go implementation fix the OAuth CSRF vulnerability?** | OAuth Security | Security - could enable account hijacking |
| 2 | **How do ThrottleService counters expire?** Is there scheduled reset? | Rate Limiting | Memory - counters may grow unbounded |
| 3 | **What is the complete SSO-only mode flow?** When email login is disabled, how do pending identities resolve? | OAuth Providers | Core onboarding path |

### High Priority (Blocks specific features)

| # | Question | Topics Affected | Impact |
|---|----------|-----------------|--------|
| 4 | Is `members_can_add_guests` intentionally excluded from paper_trail? | Permission Flags | Audit compliance |
| 5 | Why does `Null::Group` define `members_can_add_guests` in both true_methods and false_methods? | Permission Flags | Edge case behavior |
| 6 | What determines poll.discussion_id presence for standalone polls? | Stance Revision | Affects revision logic |
| 7 | Are there Sidekiq retry settings for webhook delivery? | Webhook Events | Reliability |

### Medium Priority (Implementation details)

| # | Question | Topics Affected | Impact |
|---|----------|-----------------|--------|
| 8 | What is the actual Rack::Attack cache store (Redis or Rails.cache)? | Rate Limiting | Configuration |
| 9 | Are bot API endpoints (`/api/b1/`, etc.) rate limited? | Rate Limiting | API security |
| 10 | Does Loomio sign outgoing webhook payloads (HMAC)? | Webhook Events | Integration security |
| 11 | What are all possible values for the attachment `icon` field? | Attachments JSONB | UI mapping |
| 12 | Is the unused `event` query in `stance_service.rb:34` dead code? | Stance Revision | Code cleanup |

### Low Priority (Nice to have)

| # | Question | Topics Affected | Impact |
|---|----------|-----------------|--------|
| 13 | Should `members_can_vote` column be preserved for compatibility? | Permission Flags | Migration |
| 14 | What is the purpose of `slack_community_id` column on users? | OAuth Providers | Schema |
| 15 | Is there a circuit breaker for consistently failing webhooks? | Webhook Events | Reliability |

---

## Confidence Misalignments

### Areas where third-party confidence may be overstated:

1. **Rate Limiting (claimed 5/5):** The investigation didn't verify:
   - Whether scheduled jobs reset ThrottleService counters
   - The actual cache store for Rack::Attack
   - Bot API rate limiting status

2. **OAuth Security (claimed 5/5):** The investigation didn't clarify:
   - Whether SAML signature disabling is intentional
   - Token refresh behavior for stored access tokens
   - The complete pending identity resolution flow

3. **Webhook Events (claimed 5/5):** The investigation noted but didn't resolve:
   - Whether HMAC signatures are used
   - HTTP timeout configuration
   - Matrix client caching eviction

### Areas where third-party confidence is well-justified:

1. **OAuth Providers (5/5):** Definitive source code evidence for exactly 4 providers
2. **Attachments JSONB (5/5):** Clear migration history and schema verification
3. **Permission Flags (5/5):** Complete enumeration with line number references

---

## Risk Assessment for Go Rewrite

### High Risk Areas (Complex, security-sensitive)

| Area | Risk Factors | Mitigation |
|------|--------------|------------|
| OAuth Implementation | CSRF vulnerability exists, custom implementation, 4 providers | Use golang.org/x/oauth2 with proper state handling |
| Rate Limiting | Multiple layers, unclear interactions | Implement single unified rate limiting middleware |
| Permission Flags | 12 flags with complex ability checks | Generate Go ability checks from documented rules |

### Medium Risk Areas (Behavioral complexity)

| Area | Risk Factors | Mitigation |
|------|--------------|------------|
| Stance Revision | 4-condition decision tree, two versioning mechanisms | Comprehensive unit tests, especially for standalone polls |
| Webhook Delivery | 5 payload formats, 2 delivery mechanisms | Template-based serializers, integration tests per format |

### Low Risk Areas (Well-understood)

| Area | Risk Factors | Mitigation |
|------|--------------|------------|
| Attachments JSONB | Consistent schema, well-documented | Standard Go JSONB handling |

---

## Summary

The cross-cutting analysis reveals that security-related topics (OAuth, Rate Limiting, Permissions) share common architectural patterns that require coordinated implementation in Go. The most critical gaps are:

1. **OAuth CSRF protection** - Must implement state parameter validation
2. **Rate limit error standardization** - Must return proper 429 responses with Retry-After
3. **Redis TTL management** - Must set expiration on all rate limit keys

The third party's confidence levels are generally well-justified by source code evidence, with minor gaps in rate limiting and webhook operational details that can be resolved during implementation rather than blocking synthesis.

---

## Addendum: Topics Not Covered in Third-Party Follow-Up

The following major topics were documented in baseline research (`research/investigation/`, `research/synthesis/`) but were **not included in the third-party follow-up investigation**. These require separate review.

### Critical Topics (Blocking Implementation)

| Topic | Baseline Source | Follow-up Document | Status |
|-------|-----------------|-------------------|--------|
| **Real-Time Pub/Sub** | synthesis/realtime_architecture.md | realtime_pubsub.md | NEW |
| **Full-Text Search** | schema_investigation.md (Section 9) | search_indexing.md | NEW |
| **Background Jobs** | investigation/jobs.md | infrastructure_jobs.md | NEW |

### Medium Priority Topics

| Topic | Baseline Source | Follow-up Document | Status |
|-------|-----------------|-------------------|--------|
| **Email/Mailer System** | initial_investigation_review.md (3.1) | missing_features.md | NEW |
| **File Storage** | initial_investigation_review.md (3.5) | missing_features.md | NEW |
| **Demo/RecordCloner** | initial_investigation_review.md (3.3) | missing_features.md | NEW |
| **Subscription/Billing** | initial_investigation_review.md (3.4) | missing_features.md | NEW |
| **Tasks System** | loomio_initial_investigation.md (7) | missing_features.md | NEW |
| **Mentions** | loomio_initial_investigation.md (4) | missing_features.md | NEW |
| **Counter Caches** | schema_investigation.md (5) | (See final_synthesis) | NEW |

### Unresolved Questions from Initial Investigation

From `research/investigation/questions.md`:

| Question | Priority | Status |
|----------|----------|--------|
| How is `pg_search_documents` populated? | HIGH | See search_indexing.md |
| What triggers search reindexing? | HIGH | See search_indexing.md |
| What are webhook permission values? | HIGH | NOT RESOLVED |
| Which events trigger Redis pub/sub? | HIGH | See realtime_pubsub.md |
| Guest migration incomplete (FIXME) | HIGH | See missing_features.md |

These new follow-up documents supplement the original 7-topic comparison with coverage of previously missing areas.
