# Cross-Cutting Synthesis: Architectural Patterns

**Generated:** 2026-02-01
**Topics Synthesized:** 7 (OAuth Providers, OAuth Security, Permission Flags, Rate Limiting, Attachments JSONB, Stance Revision, Webhook Events)

---

## Executive Summary

This document consolidates confirmed architectural patterns across all topic investigations and provides implementation sequencing recommendations. The synthesis reveals a coherent event-driven architecture with clear separation of concerns, but with notable security gaps that must be addressed.

---

## Confirmed Architectural Patterns

### Pattern 1: Event-Driven Notification System

**Spans:** Webhook Events, Stance Revision, Permission Flags

All significant state changes in Loomio flow through a common event system:

```
State Change (e.g., vote cast)
    |
    v
Event Published (Events::StanceCreated)
    |
    +---> In-App Notification
    +---> Email Notification
    +---> Webhook/Chatbot Notification
```

**Confirmed behaviors:**
- 42 event kinds total, 14 exposed in webhook UI
- Events reference an `eventable` (the changed object)
- Events have `kind` string for routing
- Sidekiq async workers handle delivery

---

### Pattern 2: Three-Layer Rate Limiting

**Spans:** Rate Limiting, OAuth Security

Loomio implements defense-in-depth rate limiting:

| Layer | Scope | Purpose |
|-------|-------|---------|
| Rack::Attack | IP + Endpoint | API flood protection |
| ThrottleService | User + Operation | Business limit enforcement |
| Devise Lockable | User + Failed Logins | Account protection |

---

### Pattern 3: Session-Based Authentication State

**Spans:** OAuth Providers, OAuth Security

Authentication uses server-side sessions for multi-step flows:

| Session Key | Purpose | Lifecycle |
|-------------|---------|-----------|
| `back_to` | Return URL after OAuth | Set on initiate, cleared on callback |
| `pending_identity_id` | Unlinked OAuth identity | Set when no user match, cleared on link |
| `oauth_state` | CSRF protection (MISSING) | Should be: set on initiate, validated on callback |

---

### Pattern 4: Permission Check Hierarchy

**Spans:** Permission Flags, Webhook Events

Authorization follows a consistent pattern:

```
Admin Override
    |
    v
Can always perform action?
    |-- Yes --> ALLOW
    |-- No  --> Check member permission
                    |
                    v
                Is user a member?
                    |-- No  --> DENY
                    |-- Yes --> Is permission flag enabled?
                                    |-- No  --> DENY
                                    |-- Yes --> ALLOW
```

**12 permission flags** control this hierarchy across groups, discussions, polls, and comments.

---

### Pattern 5: JSONB Column Usage

**Spans:** Attachments JSONB, Stance Revision

JSONB is used for flexible metadata storage with consistent patterns:

| Column | Default | Type | Purpose |
|--------|---------|------|---------|
| `attachments` | `[]` | Array | File metadata (8 tables) |
| `option_scores` | `{}` | Object | Vote values on stances |
| `custom_fields` | `{}` | Object | Provider-specific identity data |

**Critical rule:** Arrays default to `[]`, objects default to `{}`. API output must match (never `null`).

---

## Implementation Sequence Recommendations

### Phase 1: Core Infrastructure

Build foundational components that other features depend on:

| Component | Priority | Dependencies |
|-----------|----------|--------------|
| Session management | P0 | None |
| Redis client wrapper | P0 | None |
| JSONB types (Attachments, OptionScores) | P0 | None |
| Event model and publishing | P0 | None |
| User model with lockout | P0 | None |

**Deliverables:**
- Session middleware and store
- Redis connection pool and helpers
- Custom JSONB types
- Event types and publisher
- User with authentication fields

### Phase 2: Authentication

Implement secure OAuth with CSRF protection:

| Component | Priority | Dependencies |
|-----------|----------|--------------|
| OAuth provider interface | P0 | Session |
| Google provider | P0 | OAuth interface |
| Generic OAuth provider | P0 | OAuth interface |
| Nextcloud provider | P1 | OAuth interface |
| SAML handler | P1 | Session |
| Identity model and repository | P0 | None |

**Critical: Implement OAuth state parameter (missing in Rails):**
- Generate state on initiate
- Store in session
- Validate state on callback

**Deliverables:**
- OAuth provider interface and implementations
- SAML handler
- Identity model

### Phase 3: Authorization

Implement permission system:

| Component | Priority | Dependencies |
|-----------|----------|--------------|
| Group model with permission flags | P0 | None |
| Membership model | P0 | Group |
| Ability system | P0 | Group, Membership |
| Discussion abilities | P0 | Ability |
| Poll abilities | P0 | Ability |
| Comment abilities | P0 | Ability |

**All 12 permission flags must be implemented:**
- MembersCanAddMembers
- MembersCanAddGuests
- MembersCanAnnounce
- MembersCanCreateSubgroups
- MembersCanStartDiscussions
- MembersCanEditDiscussions
- MembersCanEditComments
- MembersCanDeleteComments
- MembersCanRaiseMotions
- AdminsCanEditUserContent
- ParentMembersCanSeeDiscussions
- (MembersCanVote - OMIT, unused legacy)

**Deliverables:**
- Group model with permissions
- Membership model
- Ability checker per model

### Phase 4: Rate Limiting

Implement all three rate limiting layers:

| Component | Priority | Dependencies |
|-----------|----------|--------------|
| Rate limit middleware | P0 | None |
| Endpoint limit configuration | P0 | Middleware |
| ThrottleService | P0 | Redis |
| User lockout integration | P1 | User model |

**Improvements over Rails:**
1. Return 429 with Retry-After header (not 500)
2. Set TTL on all Redis keys
3. Add X-RateLimit-* headers

**Deliverables:**
- HTTP middleware
- ThrottleService equivalent

### Phase 5: Voting and Stances

Implement stance revision logic:

| Component | Priority | Dependencies |
|-----------|----------|--------------|
| Poll model | P0 | Group, Discussion |
| Stance model | P0 | Poll |
| StanceService with revision logic | P0 | Stance, Events |
| Partial unique index | P0 | Schema |

**Critical: Implement 15-minute threshold correctly:**
- ALL four conditions must be true for new record creation
- is_update AND poll_in_discussion AND scores_changed AND time_exceeded

**Deliverables:**
- Poll model
- Stance model
- Update logic service

### Phase 6: Webhooks

Implement chatbot/webhook delivery:

| Component | Priority | Dependencies |
|-----------|----------|--------------|
| Chatbot model | P0 | Group |
| ChatbotService | P0 | Events, Chatbot |
| Webhook serializers (5 formats) | P0 | Service |
| Job for delivery | P0 | Service |
| Matrix Redis pub/sub | P1 | Service |

**14 webhook-eligible events to support:**
- new_discussion, discussion_edited, new_comment
- poll_created, poll_edited, poll_closing_soon
- poll_expired, poll_closed_by_user, poll_reopened
- stance_created, stance_updated
- outcome_created, outcome_updated, outcome_review_due

**Deliverables:**
- Chatbot model
- Delivery logic service
- 5 format serializers
- Delivery job

---

## Risk Areas

### High Risk: Security-Critical Components

| Component | Risk | Mitigation |
|-----------|------|------------|
| OAuth flow | CSRF vulnerability in original | **MUST** implement state parameter |
| SAML handler | Signature verification disabled | Consider enabling, document decision |
| Rate limiting | ThrottleService returns 500 | Return proper 429 with Retry-After |
| Session management | No OAuth state tracking | Use secure random state, validate on callback |

### Medium Risk: Complex Business Logic

| Component | Risk | Mitigation |
|-----------|------|------------|
| Stance revision | 4-condition decision tree | Comprehensive unit tests, especially edge cases |
| Permission checks | 12 flags with complex interactions | Generate tests from documented ability rules |
| Webhook delivery | 5 serialization formats | Template-based approach, integration tests |

### Low Risk: Well-Documented Data Patterns

| Component | Risk | Mitigation |
|-----------|------|------------|
| Attachments JSONB | Schema well-defined | Standard JSONB handling |
| Event publishing | Clear event types | Type-safe event constants |
| Identity storage | Simple CRUD | Standard repository pattern |

---

## API Compatibility Requirements

### Frontend Boot Payload

The Vue frontend expects specific JSON structure:

```json
{
    "identityProviders": [
        {"name": "google", "href": "/google/oauth"},
        {"name": "saml", "href": "/saml/oauth"}
    ],
    "features": {
        "disableEmailLogin": true
    }
}
```

### Serializer Output

All API serializers must match Rails output. Key fields:

| Model | Critical Fields |
|-------|-----------------|
| Group | All 12 permission flags, `new_threads_max_depth`, `new_threads_newest_first` |
| Discussion | `attachments` as array (never null) |
| Stance | `option_scores` as object, `latest` boolean |
| Chatbot | `event_kinds` as array, `webhook_kind` enum |

### Error Response Format

```json
{
    "errors": {
        "base": ["Rate limit exceeded. Please try again later."],
        "email": ["has already been taken"]
    }
}
```

---

## Summary

The Loomio codebase demonstrates well-documented patterns across all investigated topics:

1. **Event-driven architecture** - Clear event types and notification channels
2. **Layered rate limiting** - Three complementary mechanisms
3. **Permission-based authorization** - 12 flags with consistent check pattern
4. **JSONB flexibility** - Consistent defaults and schemas

**Critical security improvements to make:**
- Implement OAuth state parameter (CSRF protection)
- Return proper 429 responses with Retry-After
- Set TTL on all Redis rate limit keys

**Implementation should proceed in 6 phases**, starting with core infrastructure and building toward webhook delivery as the final integration point.

---

## Addendum: Extended Implementation Topics

The following topics were identified as missing from the original synthesis and have been documented in supplementary files.

### Additional Synthesis Documents

| Document | Topic | Priority |
|----------|-------|----------|
| [realtime_pubsub.md](realtime_pubsub.md) | Redis pub/sub channels, Socket.io integration | HIGH |
| [search_indexing.md](search_indexing.md) | pg_search_documents, full-text search | HIGH |
| [infrastructure_jobs.md](infrastructure_jobs.md) | 38 background workers, job definitions | HIGH |
| [counter_caches.md](counter_caches.md) | 40+ counter cache columns, reconciliation | MEDIUM |

### Feature Deferral Options

If timeline is constrained, these features can be deferred:
1. Demo/RecordCloner system (not user-facing)
2. Translation service (i18n)
3. Link previews (nice-to-have)
4. Blocked domains (can use manual list)
