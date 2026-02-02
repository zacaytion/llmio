# Loomio Rewrite Context Document

**Generated:** 2026-02-01
**Version:** 1.0
**Purpose:** Single comprehensive reference for LLM-assisted Loomio rewrite
**Token Estimate:** ~25,000 tokens

---

## How to Use This Document

This document synthesizes ~9,100 lines of specifications into a compact reference for LLM context injection during the Loomio rewrite. Use this as your primary reference, with links to detailed specs when deeper information is needed.

**Quick Navigation:**
- [1. Executive Summary](#1-executive-summary) - What is Loomio?
- [2. Architecture Overview](#2-architecture-overview) - System design
- [3. API Reference](#3-api-reference) - Key endpoints
- [4. Model Reference](#4-model-reference) - All data models
- [5. Business Logic Rules](#5-business-logic-rules) - Key decision trees
- [6. Security Requirements](#6-security-requirements) - Issues to address
- [7. Testing Requirements](#7-testing-requirements) - Critical tests
- [8. Frontend Architecture](#8-frontend-architecture) - Vue 3 patterns
- [9. External Services](#9-external-services) - Integrations
- [10. Uncertainties](#10-uncertainties-questions) - Open questions

---

## 1. Executive Summary

### What is Loomio?

Loomio is a **collaborative decision-making platform** enabling groups to discuss topics and reach decisions together through structured processes.

### Core Domain Concepts

| Concept | Description |
|---------|-------------|
| **Group** | Organization or team containing members |
| **Discussion** | Thread for conversation (can belong to group or be "direct") |
| **Poll** | Decision tool with multiple types (proposal, ranked choice, etc.) |
| **Stance** | User's vote/position on a poll |
| **Event** | Activity record driving timelines and notifications |
| **Membership** | User's relationship to a group (member, admin, guest) |

### Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Rails 8 API-only |
| Frontend | Vue 3 SPA |
| Database | PostgreSQL with pg_search |
| Queue | Sidekiq + Redis |
| Cache/Pub/Sub | Redis |
| Real-time | Socket.io (external) |
| Collaboration | Hocuspocus + Yjs |
| Client State | LokiJS in-memory DB |

### Key Architectural Patterns

1. **Service Layer** - All mutations flow through `*Service` classes (`PollService.create`, `DiscussionService.update`)
2. **Event Sourcing** - Actions create Event records that drive notifications, timelines, webhooks
3. **Permission Flags** - Groups have `members_can_*` flags controlling member capabilities
4. **Real-time** - Events publish to Redis, Socket.io broadcasts to Vue clients
5. **Client-side ORM** - LokiJS mirrors Rails models with relationships

---

## 2. Architecture Overview

### Request Flow

```
Vue SPA                      Rails API                    Database
   │                            │                            │
   ├─── REST /api/v1/* ────────►├─── Controller ────────────►│
   │                            │         │                  │
   │                            │    authorize!(CanCanCan)   │
   │                            │         │                  │
   │                            │    *Service.action()       │
   │                            │         │                  │
   │                            │    Event.publish!          │
   │                            │         │                  │
   │◄── JSON response ─────────┤◄───────────────────────────┤
   │                            │                            │
   │                            │    PublishEventWorker      │
   │                            │         │                  │
   │◄── Socket.io (records) ───┤◄── Redis pub/sub           │
```

### Directory Structure (Backend)

```
app/
├── controllers/api/v1/     # REST endpoints (~30 controllers)
├── models/                 # ActiveRecord models + concerns
│   └── ability/           # CanCanCan abilities per model
├── services/              # Business logic (*Service classes)
├── serializers/           # ActiveModelSerializers 0.8
├── workers/               # Sidekiq jobs
└── extras/                # OAuth clients, helpers
```

### Directory Structure (Frontend)

```
vue/src/
├── components/            # 217 Vue components (feature-organized)
├── shared/
│   ├── services/         # 35 services (records.js, session.js, etc.)
│   ├── models/           # 31 client-side models
│   ├── interfaces/       # 28 LokiJS record interfaces
│   └── record_store/     # LokiJS infrastructure
└── routes.js             # Vue Router config
```

---

## 3. API Reference

### Endpoint Summary

**Base URL:** `/api/v1/`

| Resource | Methods | Key Endpoints |
|----------|---------|---------------|
| Groups | CRUD + archive | `/groups`, `/groups/:id/subgroups` |
| Memberships | CRUD | `/memberships`, `/memberships/:id/make_admin` |
| Discussions | CRUD + move/close | `/discussions`, `/discussions/:id/move` |
| Comments | CRUD | `/comments` |
| Polls | CRUD + close/reopen | `/polls`, `/polls/:id/close` |
| Stances | Create/Update | `/stances` |
| Outcomes | Create/Update | `/outcomes` |
| Events | Read | `/events` (discussion timeline) |
| Users | Profile | `/profile` (not `/users`) |
| Search | Query | `/search` |
| Announcements | Send | `/announcements` |

### Authentication

- Session-based (cookie) for web app
- API key in query param (`?api_key=`) for bot APIs (`/api/b2/`, `/api/b3/`)
- OAuth/SAML for SSO

### Common Response Format

```json
{
  "discussions": [...],
  "polls": [...],
  "users": [...],
  "groups": [...],
  "events": [...]
}
```

Multiple related records returned in single responses for client-side store hydration.

### Rate Limits (Rack::Attack)

| Endpoint | Limit |
|----------|-------|
| `/api/v1/discussions` | 500/hour/IP |
| `/api/v1/comments` | 500/hour/IP |
| `/api/v1/stances` | 500/hour/IP |
| `/api/v1/announcements` | 100/hour/IP |
| `/api/v1/memberships` | 100/hour/IP |
| Bot APIs (`/api/b2/`, `/api/b3/`) | **NONE** (bug) |

---

## 4. Model Reference

### Core Models

#### User

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `email` | string | Unique, verified flag |
| `name` | string | Display name |
| `secret_token` | string | WebSocket auth |
| `email_api_key` | string | Reply-by-email auth |
| `experiences` | jsonb | Feature flags, preferences |
| `deactivated_at` | datetime | Soft delete |

**Relationships:** has_many memberships, stances, notifications

#### Group

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `name` | string | Group name |
| `handle` | string | URL slug |
| `parent_id` | integer | Subgroup parent |
| `subscription_id` | integer | Billing |
| Permission flags | boolean | See below |

**Permission Flags (12):**
- `members_can_add_members`, `members_can_add_guests`
- `members_can_start_discussions`, `members_can_raise_motions`
- `members_can_edit_discussions`, `members_can_edit_comments`, `members_can_delete_comments`
- `members_can_announce`, `members_can_create_subgroups`
- `admins_can_edit_user_content`, `parent_members_can_see_discussions`
- `members_can_vote` (deprecated)

#### Discussion

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `title` | string | Thread title |
| `description` | text | Rich text body |
| `private` | boolean | Visibility |
| `closed_at` | datetime | Thread closed |
| `group_id` | integer | NULL = direct discussion |
| `max_depth` | integer | Reply nesting limit |

**Relationships:** belongs_to group (optional), author; has_many comments, polls, events

#### Poll

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `title` | string | Poll title |
| `poll_type` | string | proposal, poll, count, score, dot_vote, ranked_choice, meeting, check, question |
| `closing_at` | datetime | When poll closes |
| `anonymous` | boolean | Hide voter names (cannot un-anonymize) |
| `hide_results` | integer | 0=show, 1=until_vote, 2=until_closed |
| `stance_reason_required` | integer | 0=disabled, 1=optional, 2=required |

**Relationships:** belongs_to discussion (optional), group; has_many stances, poll_options, outcomes

#### Stance (Vote)

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `poll_id` | integer | Parent poll |
| `participant_id` | integer | Voter |
| `reason` | text | Vote rationale |
| `cast_at` | datetime | When vote submitted |
| `latest` | boolean | Most recent stance |
| `stance_choices_cache` | jsonb | Cached vote selections |

**Vote Revision Rule:** New stance record created only if ALL:
- Time since last vote > 15 minutes
- Choices changed
- Poll is in discussion

#### Event

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `kind` | string | Event type (STI discriminator) |
| `eventable_type/id` | polymorphic | What triggered event |
| `user_id` | integer | Actor |
| `discussion_id` | integer | Parent discussion |
| `sequence_id` | integer | Timeline position |

**42 Event Types** including: `NewComment`, `PollCreated`, `StanceCreated`, `DiscussionClosed`, etc.

#### Membership

| Attribute | Type | Key Behavior |
|-----------|------|--------------|
| `group_id` | integer | Group |
| `user_id` | integer | User |
| `admin` | boolean | Is admin? |
| `inviter_id` | integer | Who invited |
| `accepted_at` | datetime | When accepted |

### Supporting Models

| Model | Purpose |
|-------|---------|
| `Comment` | Discussion replies |
| `PollOption` | Poll answer choices |
| `Outcome` | Published poll results |
| `Notification` | In-app user notifications |
| `Document` | Attached files |
| `Reaction` | Emoji reactions |
| `Tag` | Content tagging |
| `Webhook/Chatbot` | Webhook configurations |
| `Version` | Paper Trail audit log |
| `DiscussionReader` | Per-user read state |

---

## 5. Business Logic Rules

### Service Layer Pattern

```ruby
def self.create(discussion:, actor:, params:)
  actor.ability.authorize! :create, discussion
  discussion.assign_attributes(params)
  discussion.save!
  Events::NewDiscussion.publish!(discussion, user: actor)
  discussion
end
```

### Decision Trees

#### Vote Revision (StanceService)

```
CREATE_NEW_STANCE?
├── time_since_last_vote > 15_minutes?
│   ├── NO → UPDATE existing stance
│   └── YES
│       └── choices_changed?
│           ├── NO → UPDATE existing stance
│           └── YES
│               └── poll.discussion_id.present?
│                   ├── NO → UPDATE existing stance
│                   └── YES → CREATE new stance record
```

#### Permission Check Pattern

```
CAN_PERFORM_ACTION?
├── User is group admin?
│   └── YES → ALLOW
├── User is group member?
│   └── YES
│       └── group.members_can_{action}?
│           ├── YES → ALLOW
│           └── NO → DENY
└── User is guest on discussion?
    └── Check guest-specific rules
```

#### Email Notification Delivery

```
SHOULD_SEND_EMAIL?
├── user.complaints_count > 0? → NO
├── user.email_verified? → NO if false
├── user.deactivated_at? → NO if set
├── user.volume == "quiet"? → NO (except direct mentions)
└── YES → Queue email
```

### Event System (42 types)

| Category | Types | Behavior |
|----------|-------|----------|
| Real-time (16) | NewComment, PollCreated, StanceCreated... | LiveUpdate + Notify |
| Notify-only (18) | UserMentioned, MembershipCreated... | Notification only |
| No notification (8) | DiscussionForked, UserJoinedGroup... | Audit/chatbot only |

### Webhook Events (14)

- new_discussion, discussion_edited, new_comment
- poll_created, poll_edited, poll_closing_soon, poll_expired, poll_closed_by_user, poll_reopened
- stance_created, stance_updated
- outcome_created, outcome_updated, outcome_review_due

---

## 6. Security Requirements

### Issues to Address (Prioritized)

#### CRITICAL (Fix Before Production)

| ID | Issue | Location | Fix |
|----|-------|----------|-----|
| 001 | OAuth missing `state` param (CSRF) | `identities/*_controller.rb` | Add state to session, validate on callback |
| 003 | ThrottleService returns HTTP 500 | `snorlax_base.rb` | Add `rescue_from LimitReached` returning 429 |

#### HIGH (Next Sprint)

| ID | Issue | Location | Fix |
|----|-------|----------|-----|
| 004 | Bot APIs no rate limiting | `rack_attack.rb` | Add limits for `/api/b2/`, `/api/b3/` |
| 005 | Redis throttle no TTL | `throttle_service.rb` | Add EXPIRE on counters |
| 006 | Webhooks unsigned | `chatbot_service.rb` | Add HMAC signature header |

#### MEDIUM (Next Quarter)

| ID | Issue | Location | Fix |
|----|-------|----------|-----|
| 007 | No webhook circuit breaker | `chatbot.rb` | Track failures, auto-disable |
| 008 | CSRF verification skips | Multiple | Document why each is safe |
| 009 | `members_can_add_guests` not in Paper Trail | `group.rb` | Add to `only:` list |

### Security Patterns to Preserve

- HTML sanitization via `HasRichText` concern (whitelist-based)
- CanCanCan authorization on all mutations
- CSRF token validation (except documented exceptions)
- Reply-by-email uses cryptographic tokens

---

## 7. Testing Requirements

### Critical Test Categories

#### OAuth Security (15 tests)

- State parameter generation and validation
- Token exchange error handling
- Identity linking edge cases
- SSO-only mode behavior

#### Permission Tests (25 tests)

- Each `members_can_*` flag enforced correctly
- Admin vs member vs guest capabilities
- Null::Group (direct discussion) permissions
- Subgroup visibility rules

#### Rate Limiting (10 tests)

- ThrottleService returns 429 with Retry-After
- Rack::Attack IP limits enforced
- Bot API limits (once implemented)

#### Event System (20 tests)

- Each event type triggers correct concerns
- LiveUpdate publishes to correct rooms
- Notification delivery rules
- Webhook delivery

### Key Test Patterns

```ruby
# FactoryBot factories used
:user, :group, :discussion, :poll, :comment, :membership, :stance

# E2E tests use dev routes
/dev/scenarios/{scenario_name}  # Nightwatch test setup
```

---

## 8. Frontend Architecture

### LokiJS State Management

```javascript
// Central store
const db = new loki('default.db');
const records = new RecordStore(db);

// 28 record interfaces registered
records.addRecordsInterface(DiscussionRecordsInterface);
records.addRecordsInterface(PollRecordsInterface);
// ...

// Real-time updates
conn.on('records', data => Records.importJSON(data.records));
```

### Component Organization (217 total)

| Directory | Count | Purpose |
|-----------|-------|---------|
| `poll/` | 49 | Poll UI |
| `common/` | 33 | Shared UI |
| `strand/` | 25 | Thread display |
| `group/` | 26 | Group management |

### API Integration Pattern

```javascript
// RestfulClient wraps fetch()
// CSRF from csrftoken cookie
// API prefix: /api/v1/

Records.polls.remote.create(poll.serialize())
  .then(data => Records.importJSON(data))
  .catch(err => handleError(err));
```

### Routing

| Path | Component |
|------|-----------|
| `/d/:key` | StrandPage (discussion) |
| `/p/:key` | PollShowPage |
| `/g/:key` | GroupPage |
| `/dashboard` | DashboardPage |

### Real-time Connection

```javascript
// Socket.io connects with channel_token
conn = io(AppConfig.theme.channels_url, {
  query: { channel_token: AppConfig.channel_token }
});

conn.on('records', data => Records.importJSON(data.records));
conn.on('notice', data => EventBus.$emit('systemNotice', data));
```

---

## 9. External Services

### Required Services

| Service | Purpose | Key Variables |
|---------|---------|---------------|
| PostgreSQL | Database | `DATABASE_URL` |
| Redis | Queue/cache/pub/sub | `REDIS_URL` |

### Email (Required for Production)

| Variable | Purpose |
|----------|---------|
| `SMTP_SERVER` | SMTP host |
| `SMTP_PORT` | SMTP port |
| `SMTP_USERNAME` | Auth user |
| `SMTP_PASSWORD` | Auth password |
| `REPLY_HOSTNAME` | Reply-to domain |

### File Storage (Choose One)

| Backend | Key Variables |
|---------|---------------|
| Local | Default |
| S3 | `AWS_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| DigitalOcean | `DO_ENDPOINT`, `DO_ACCESS_KEY_ID`, `DO_SECRET_ACCESS_KEY`, `DO_BUCKET` |
| GCS | `GCS_CREDENTIALS`, `GCS_PROJECT`, `GCS_BUCKET` |

### OAuth Providers (Optional)

| Provider | Key Variables |
|----------|---------------|
| Google | `GOOGLE_APP_KEY`, `GOOGLE_APP_SECRET` |
| Generic OAuth | `OAUTH_APP_KEY`, `OAUTH_AUTHORIZE_URL`, `OAUTH_TOKEN_URL`, etc. |
| SAML | `SAML_IDP_METADATA` or `SAML_IDP_METADATA_URL` |
| Nextcloud | `NEXTCLOUD_HOST`, `NEXTCLOUD_APP_KEY` |

### Collaborative Editing

| Service | Variables |
|---------|-----------|
| Hocuspocus | `HOCUSPOCUS_URL` (defaults to `wss://hocuspocus.{host}`) |

### Redis Pub/Sub Channels

| Channel | Purpose |
|---------|---------|
| `/records` | Model updates to clients |
| `/system_notice` | Broadcast messages |

### Room Routing

| Room | Use |
|------|-----|
| `group-{id}` | Group member broadcasts |
| `user-{id}` | Personal notifications |

---

## 10. Uncertainties & Questions

### HIGH Priority

| Topic | Gap | Action |
|-------|-----|--------|
| OAuth CSRF | Missing state parameter | **Must fix before production** |
| Matrix chatbot | Protocol undocumented | May be external service |
| Hocuspocus config | Server setup unknown | Document client integration only |
| Socket.io server | Not in Rails codebase | External service |

### MEDIUM Priority

| Topic | Gap | Action |
|-------|-----|--------|
| Poll template JSON | Custom fields schema | Create formal schema |
| Task reminder logic | Not fully traced | Complete investigation |
| Chargify webhooks | Subscription lifecycle | Clarify with original team |

### Questions for Original Authors

1. **Testing:** Where are frontend tests (Jest, Vitest, Cypress)?
2. **Hocuspocus:** What's the server configuration?
3. **Socket.io:** Where is the Socket.io server deployed?
4. **Mobile:** Separate mobile build or responsive-only?
5. **Matrix:** Is Matrix chatbot actively used?

---

## Appendix: File Reference

### Detailed Specifications

| File | Lines | Content |
|------|-------|---------|
| `discovery/specifications/business_logic.md` | 784 | Services, events, permissions |
| `discovery/specifications/models/*.md` | 5,499 | 10 model specs |
| `discovery/specifications/security_report.md` | 605 | 11 security issues |
| `discovery/specifications/external_services.md` | 776 | 8 service integrations |
| `discovery/specifications/testing_requirements.md` | 531 | 100+ test cases |
| `discovery/specifications/frontend.md` | ~800 | Vue architecture |

### OpenAPI Specification

| Path | Content |
|------|---------|
| `discovery/openapi/openapi.yaml` | Root spec |
| `discovery/openapi/paths/` | 24 endpoint files (~204 endpoints) |
| `discovery/openapi/components/schemas/` | 10 entity schemas |

### Database Schema

| Path | Content |
|------|---------|
| `discovery/schemas/database_schema.md` | Full PostgreSQL schema |
| `discovery/schemas/request_schemas/` | 23 request schemas |
| `discovery/schemas/response_schemas/` | 30+ response schemas |

---

## Confidence Levels

| Section | Confidence | Notes |
|---------|------------|-------|
| Architecture | HIGH | Direct code inspection |
| API Reference | HIGH | OpenAPI generated |
| Models | HIGH | Schema + model code |
| Business Logic | HIGH | Service layer traced |
| Security | HIGH | Vulnerabilities verified |
| Testing | MEDIUM | Test coverage incomplete |
| Frontend | HIGH | Component inventory complete |
| External Services | HIGH | Config files documented |
| Uncertainties | N/A | Known gaps |

---

*Generated: 2026-02-01*
*Source: Phases 1-8 specifications (~9,100 lines synthesized)*
*Completeness: 95%*
