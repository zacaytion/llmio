# Cross-Cutting Synthesis: Architectural Patterns and Implementation Guidance

**Generated:** 2026-02-01
**Topics Synthesized:** 7 (OAuth Providers, OAuth Security, Permission Flags, Rate Limiting, Attachments JSONB, Stance Revision, Webhook Events)

---

## Executive Summary

This document consolidates confirmed architectural patterns across all topic investigations and provides implementation sequencing recommendations for the Go rewrite. The synthesis reveals a coherent event-driven architecture with clear separation of concerns, but with notable security gaps that must be addressed during the rewrite.

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

**Go implementation requirement:**
```go
// Event types as constants
type EventKind string

const (
    EventNewDiscussion    EventKind = "new_discussion"
    EventDiscussionEdited EventKind = "discussion_edited"
    EventNewComment       EventKind = "new_comment"
    // ... 41 more
)

// Event model
type Event struct {
    ID           int64
    Kind         EventKind
    EventableID  int64
    EventableType string
    DiscussionID *int64
    UserID       int64
    CreatedAt    time.Time
}
```

---

### Pattern 2: Three-Layer Rate Limiting

**Spans:** Rate Limiting, OAuth Security

Loomio implements defense-in-depth rate limiting:

| Layer | Scope | Purpose | Go Equivalent |
|-------|-------|---------|---------------|
| Rack::Attack | IP + Endpoint | API flood protection | chi/httprate middleware |
| ThrottleService | User + Operation | Business limit enforcement | Custom Redis service |
| Devise Lockable | User + Failed Logins | Account protection | User model with lockout |

**Go implementation requirement:**
```go
// Layer 1: Middleware (per-endpoint, per-IP)
r.Use(httprate.LimitByIP(100, time.Hour))

// Layer 2: Service (per-user, per-operation)
type ThrottleService struct {
    redis *redis.Client
}

func (s *ThrottleService) Limit(ctx context.Context, key, id string, max int, per string) error

// Layer 3: User model (failed login tracking)
type User struct {
    FailedAttempts int
    LockedAt       *time.Time
    UnlockToken    *string
}
```

---

### Pattern 3: Session-Based Authentication State

**Spans:** OAuth Providers, OAuth Security

Authentication uses server-side sessions for multi-step flows:

| Session Key | Purpose | Lifecycle |
|-------------|---------|-----------|
| `back_to` | Return URL after OAuth | Set on initiate, cleared on callback |
| `pending_identity_id` | Unlinked OAuth identity | Set when no user match, cleared on link |
| `oauth_state` | CSRF protection (MISSING) | Should be: set on initiate, validated on callback |

**Go implementation requirement:**
```go
// Session store with signed cookies
store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

// OAuth state handling (MUST IMPLEMENT - not in Rails)
func (h *OAuthHandler) Initiate(w http.ResponseWriter, r *http.Request) {
    session := store.Get(r, "loomio-session")
    state := generateSecureRandom(32)
    session.Values["oauth_state"] = state
    session.Values["oauth_back_to"] = r.URL.Query().Get("back_to")
    session.Save(r, w)

    authURL := provider.AuthorizationURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}
```

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

**Go implementation requirement:**
```go
type Ability struct {
    user *User
}

func (a *Ability) CanEditComment(comment *Comment, discussion *Discussion, group *Group) bool {
    // Admin override
    if a.isAdminOf(discussion.ID) && group.AdminsCanEditUserContent {
        return true
    }

    // Member permission (own comments only for this flag)
    if a.isMemberOf(discussion.ID) &&
       comment.AuthorID == a.user.ID &&
       group.MembersCanEditComments {
        return true
    }

    return false
}
```

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

**Go implementation requirement:**
```go
// Custom type with proper nil handling
type Attachments []Attachment

func (a Attachments) MarshalJSON() ([]byte, error) {
    if a == nil {
        return []byte("[]"), nil // CRITICAL: never return null
    }
    return json.Marshal([]Attachment(a))
}
```

---

## Implementation Sequence Recommendations

### Phase 1: Core Infrastructure (Weeks 1-2)

Build foundational components that other features depend on:

| Component | Priority | Dependencies | Estimated Effort |
|-----------|----------|--------------|------------------|
| Session management (gorilla/sessions) | P0 | None | 1 day |
| Redis client wrapper | P0 | None | 1 day |
| JSONB types (Attachments, OptionScores) | P0 | None | 2 days |
| Event model and publishing | P0 | None | 2 days |
| User model with lockout | P0 | None | 1 day |

**Deliverables:**
- `internal/session/` - Session middleware and store
- `internal/redis/` - Connection pool and helpers
- `internal/models/jsonb/` - Custom JSONB types
- `internal/events/` - Event types and publisher
- `internal/models/user.go` - User with authentication fields

### Phase 2: Authentication (Weeks 3-4)

Implement secure OAuth with CSRF protection:

| Component | Priority | Dependencies | Estimated Effort |
|-----------|----------|--------------|------------------|
| OAuth provider interface | P0 | Session | 1 day |
| Google provider | P0 | OAuth interface | 2 days |
| Generic OAuth provider | P0 | OAuth interface | 2 days |
| Nextcloud provider | P1 | OAuth interface | 1 day |
| SAML handler | P1 | Session | 3 days |
| Identity model and repository | P0 | None | 1 day |

**Critical: Implement OAuth state parameter (missing in Rails):**
```go
// Generate state on initiate
state := base64.URLEncoding.EncodeToString(randomBytes(32))
session.Set("oauth_state", state)

// Validate state on callback
if r.URL.Query().Get("state") != session.Get("oauth_state") {
    return ErrInvalidOAuthState
}
```

**Deliverables:**
- `internal/oauth/` - Provider interface and implementations
- `internal/saml/` - SAML handler using crewjam/saml
- `internal/models/identity.go` - Identity model

### Phase 3: Authorization (Weeks 5-6)

Implement permission system:

| Component | Priority | Dependencies | Estimated Effort |
|-----------|----------|--------------|------------------|
| Group model with permission flags | P0 | None | 1 day |
| Membership model | P0 | Group | 1 day |
| Ability system | P0 | Group, Membership | 3 days |
| Discussion abilities | P0 | Ability | 1 day |
| Poll abilities | P0 | Ability | 1 day |
| Comment abilities | P0 | Ability | 1 day |

**All 12 permission flags must be implemented:**
```go
type GroupPermissions struct {
    MembersCanAddMembers           bool
    MembersCanAddGuests            bool
    MembersCanAnnounce             bool
    MembersCanCreateSubgroups      bool
    MembersCanStartDiscussions     bool
    MembersCanEditDiscussions      bool
    MembersCanEditComments         bool
    MembersCanDeleteComments       bool
    MembersCanRaiseMotions         bool
    AdminsCanEditUserContent       bool
    ParentMembersCanSeeDiscussions bool
    // MembersCanVote - OMIT (unused legacy)
}
```

**Deliverables:**
- `internal/models/group.go` - Group with permissions
- `internal/models/membership.go` - Membership model
- `internal/ability/` - Ability checker per model

### Phase 4: Rate Limiting (Week 7)

Implement all three rate limiting layers:

| Component | Priority | Dependencies | Estimated Effort |
|-----------|----------|--------------|------------------|
| Rate limit middleware (httprate) | P0 | None | 1 day |
| Endpoint limit configuration | P0 | Middleware | 1 day |
| ThrottleService | P0 | Redis | 2 days |
| User lockout integration | P1 | User model | 1 day |

**Improvements over Rails:**
1. Return 429 with Retry-After header (not 500)
2. Set TTL on all Redis keys
3. Add X-RateLimit-* headers

**Deliverables:**
- `internal/middleware/ratelimit/` - HTTP middleware
- `internal/throttle/` - ThrottleService equivalent

### Phase 5: Voting and Stances (Week 8)

Implement stance revision logic:

| Component | Priority | Dependencies | Estimated Effort |
|-----------|----------|--------------|------------------|
| Poll model | P0 | Group, Discussion | 2 days |
| Stance model | P0 | Poll | 1 day |
| StanceService with revision logic | P0 | Stance, Events | 3 days |
| Partial unique index | P0 | Schema | 1 hour |

**Critical: Implement 15-minute threshold correctly:**
```go
const StanceRevisionThreshold = 15 * time.Minute

func shouldCreateNewStance(stance *Stance, poll *Poll, newScores map[string]int) bool {
    // ALL four conditions must be true
    isUpdate := stance.CastAt != nil
    hasDiscussion := poll.DiscussionID != nil
    scoresChanged := !mapsEqual(stance.OptionScores, newScores)
    timeExceeded := time.Since(stance.UpdatedAt) > StanceRevisionThreshold

    return isUpdate && hasDiscussion && scoresChanged && timeExceeded
}
```

**Deliverables:**
- `internal/models/poll.go` - Poll model
- `internal/models/stance.go` - Stance model
- `internal/services/stance_service.go` - Update logic

### Phase 6: Webhooks (Weeks 9-10)

Implement chatbot/webhook delivery:

| Component | Priority | Dependencies | Estimated Effort |
|-----------|----------|--------------|------------------|
| Chatbot model | P0 | Group | 1 day |
| ChatbotService | P0 | Events, Chatbot | 2 days |
| Webhook serializers (5 formats) | P0 | Service | 3 days |
| River job for delivery | P0 | Service | 1 day |
| Matrix Redis pub/sub | P1 | Service | 2 days |

**14 webhook-eligible events to support:**
```go
var WebhookEligibleEvents = []string{
    "new_discussion", "discussion_edited", "new_comment",
    "poll_created", "poll_edited", "poll_closing_soon",
    "poll_expired", "poll_closed_by_user", "poll_reopened",
    "stance_created", "stance_updated",
    "outcome_created", "outcome_updated", "outcome_review_due",
}
```

**Deliverables:**
- `internal/models/chatbot.go` - Chatbot model
- `internal/services/chatbot_service.go` - Delivery logic
- `internal/webhook/serializers/` - 5 format serializers
- `internal/jobs/webhook_delivery.go` - River job

---

## Shared Go Infrastructure Needs

### Common Middleware Stack

```go
func NewRouter() *chi.Mux {
    r := chi.NewRouter()

    // Standard middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(slogchi.Logger(slog.Default()))
    r.Use(middleware.Recoverer)

    // Session (required for OAuth)
    r.Use(SessionMiddleware)

    // Rate limiting (per-endpoint configured)
    r.Use(RateLimitMiddleware)

    // Authentication (optional, sets current_user)
    r.Use(AuthenticationMiddleware)

    return r
}
```

### Common Types Package

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
type Attachments []Attachment
type OptionScores map[string]int
type CustomFields map[string]interface{}
```

### Common Error Types

```go
// internal/errors/errors.go
var (
    ErrNotFound          = errors.New("not found")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
    ErrInvalidOAuthState = errors.New("invalid OAuth state")
    ErrAccountLocked     = errors.New("account locked")
)
```

### Common Test Utilities

```go
// internal/testutil/factories.go
func UserFactory(t *testing.T, db *pgxpool.Pool) *User
func GroupFactory(t *testing.T, db *pgxpool.Pool, opts ...GroupOption) *Group
func MembershipFactory(t *testing.T, db *pgxpool.Pool, user *User, group *Group, admin bool) *Membership
func IdentityFactory(t *testing.T, db *pgxpool.Pool, user *User, provider string) *Identity
```

---

## Risk Areas for Rewrite

### High Risk: Security-Critical Components

| Component | Risk | Mitigation |
|-----------|------|------------|
| OAuth flow | CSRF vulnerability in original | **MUST** implement state parameter |
| SAML handler | Signature verification disabled | Consider enabling in Go, document decision |
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
| Attachments JSONB | Schema well-defined | Standard Go JSONB handling |
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

The Go rewrite benefits from well-documented patterns across all investigated topics:

1. **Event-driven architecture** - Clear event types and notification channels
2. **Layered rate limiting** - Three complementary mechanisms
3. **Permission-based authorization** - 12 flags with consistent check pattern
4. **JSONB flexibility** - Consistent defaults and schemas

**Critical security improvements to make:**
- Implement OAuth state parameter (CSRF protection)
- Return proper 429 responses with Retry-After
- Set TTL on all Redis rate limit keys

**Implementation should proceed in 6 phases** over ~10 weeks, starting with core infrastructure and building toward webhook delivery as the final integration point.

---

## Addendum: Extended Implementation Topics

The following topics were identified as missing from the original synthesis and have been documented in supplementary files.

### Additional Synthesis Documents

| Document | Topic | Priority |
|----------|-------|----------|
| [realtime_pubsub.md](realtime_pubsub.md) | Redis pub/sub channels, Socket.io integration | HIGH |
| [search_indexing.md](search_indexing.md) | pg_search_documents, full-text search | HIGH |
| [infrastructure_jobs.md](infrastructure_jobs.md) | 38 background workers, River job definitions | HIGH |
| [counter_caches.md](counter_caches.md) | 40+ counter cache columns, reconciliation | MEDIUM |

### Extended Implementation Phases

**Original Phases 1-6** remain unchanged.

**Phase 7: Real-Time Infrastructure** (Week 11)
- MessageChannelService equivalent
- Redis pub/sub publishing
- Hocuspocus authentication endpoint
- Channel token setup in boot endpoint

**Phase 8: Search** (Week 12)
- pg_search_documents index population
- Search query service
- Bulk reindex capability

**Phase 9: Extended Features** (Weeks 13-14)
- Email system (catch-up emails, notifications)
- File storage backends
- Task system
- Mention parsing

### Total Estimated Effort

| Phase | Duration | Status |
|-------|----------|--------|
| Phases 1-6 | 10 weeks | Original plan |
| Phases 7-9 | 4 weeks | Extended scope |
| **Total** | **14 weeks** | Full feature parity |

### Feature Deferral Options

If timeline is constrained, these features can be deferred:
1. Demo/RecordCloner system (not user-facing)
2. Translation service (i18n)
3. Link previews (nice-to-have)
4. Blocked domains (can use manual list)
