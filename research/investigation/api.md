# API Reference

> Routes, serializers, webhooks, and rate limiting.

## API Versions

| Prefix | Purpose | Status |
|--------|---------|--------|
| `/api/v1` | Primary API (SPA, mobile) | Active |
| `/api/b1` | User auth, profile, sessions | Active |
| `/api/b2` | Discussion-centric (mobile) | Active |
| `/api/b3` | Extended API | Active |

## Route Summary

### /api/v1 Resources (28)

| Resource | Controller | Key Actions |
|----------|-----------|-------------|
| attachments | AttachmentsController | create, destroy |
| boot | BootController | index, site |
| chatbots | ChatbotsController | CRUD, test |
| comments | CommentsController | CRUD, discard, undiscard |
| contact_messages | ContactMessagesController | create |
| demos | DemosController | index, take |
| discussion_readers | DiscussionReadersController | update, update_volume |
| discussion_templates | DiscussionTemplatesController | CRUD |
| discussions | DiscussionsController | CRUD, search, history, move, discard/undiscard |
| document | DocumentController | index, for_group, for_discussion |
| events | EventsController | index, mark_as_read |
| forward_email_rules | ForwardEmailRulesController | CRUD |
| group_identities | GroupIdentitiesController | CRUD |
| groups | GroupsController | CRUD, archive, unarchive, export, join |
| login_tokens | LoginTokensController | create |
| member_email_aliases | MemberEmailAliasesController | CRUD |
| membership_requests | MembershipRequestsController | CRUD, approve, ignore |
| memberships | MembershipsController | index, join, leave, invite, resend, destroy |
| notifications | NotificationsController | index, viewed |
| outcomes | OutcomesController | CRUD |
| poll_options | PollOptionsController | CRUD |
| poll_templates | PollTemplatesController | CRUD |
| polls | PollsController | CRUD, close, reopen, remind, search |
| reactions | ReactionsController | CRUD |
| registrations | RegistrationsController | create, oauth |
| search | SearchController | index |
| stances | StancesController | CRUD |
| webhooks | WebhooksController | CRUD |

### /api/b1 Routes

| Path | Controller | Purpose |
|------|-----------|---------|
| POST /email_exists | Devise | Check email availability |
| POST /verify_email | EmailVerificationController | Verify email token |
| POST /sessions | SessionsController | Login |
| DELETE /sessions | SessionsController | Logout |
| GET /profile | ProfileController | Current user |

### /api/b2 Routes

| Path | Purpose |
|------|---------|
| GET /inbox | Inbox items |
| GET /discussions/:key | Discussion with context |
| GET /polls/:key | Poll with stances |

### /api/b3 Routes

| Path | Purpose |
|------|---------|
| GET /user | Extended user data |
| GET /notifications | Notification feed |

## Boot Endpoint

**GET /api/v1/boot**

Returns all data needed to start the SPA:

```json
{
  "current_user": {...},
  "memberships": [...],
  "groups": [...],
  "notifications": [...],
  "site": {...},
  "channels_uri": "wss://...",
  "hocuspocus_uri": "wss://...",
  "channel_token": "uuid-here"
}
```

**Key Behavior:** Sets `/current_users/{secret_token}` in Redis for WebSocket auth.

**Source:** `orig/loomio/app/controllers/api/v1/boot_controller.rb:26-32`

## Serializer Pattern

All responses wrapped with `records` key:

```json
{
  "users": [...],
  "groups": [...],
  "discussions": [...],
  "...": [...]
}
```

**Sideloading:** Related records included automatically (e.g., discussion includes author, group).

**Source:** `orig/loomio/app/serializers/`

## Authentication

### Session-Based
- Cookie: `_loomio_session`
- CSRF: `X-CSRF-Token` header

### API Key
- Header: `Loomio-API-Key: {user.email_api_key}`
- Used by bots/integrations

### OAuth
- Standard OAuth2 flow
- Tokens in `oauth_access_tokens`

## Webhooks

### Eligible Events (14)

| Event Kind | Trigger |
|------------|---------|
| new_discussion | Discussion created |
| discussion_edited | Title/description changed |
| new_comment | Comment created |
| poll_created | Poll created |
| poll_edited | Poll updated |
| poll_closing_soon | Approaching close time |
| poll_expired | Poll closed automatically |
| poll_closed_by_user | Poll closed manually |
| stance_created | Vote cast |
| stance_updated | Vote changed |
| outcome_created | Outcome announced |
| outcome_updated | Outcome edited |
| user_added_to_group | Member joined |
| membership_requested | Join request |

**Source:** `orig/loomio/config/webhook_event_kinds.yml`

### Webhook Model

```go
type Webhook struct {
    ID           int64
    GroupID      int64
    ActorID      int64
    URL          string
    Name         string
    Format       string    // 'markdown', 'microsoft', 'slack', 'discord'
    Permissions  []string  // PostgreSQL array
    EventKinds   []string  // Events to fire on
    IsDiscovered bool      // Via bot discovery?
    Token        string
}
```

**Note:** `permissions` array filtering is not fully documented. See [questions.md](./questions.md).

### Payload Format

```json
{
  "event": "new_discussion",
  "webhook": {...},
  "discussion": {...},
  "group": {...},
  "user": {...}
}
```

## Rate Limiting

**Source:** `orig/loomio/app/services/throttle_service.rb`

### API

```ruby
ThrottleService.can?(key:, id:, max:, inc:, per:)  # Check and increment
ThrottleService.limit!(...)  # Same but raises if exceeded
ThrottleService.reset!(period)  # Clear counters
```

### Configuration

- **Backend:** Redis counters
- **Key Pattern:** `THROTTLE-{HOUR|DAY}-{key}-{id}`
- **Default Limit:** 100 (override via `ENV['THROTTLE_MAX_{KEY}']`)

### Usage

| Endpoint | Limit |
|----------|-------|
| Email bounces | 1/hour |
| Login attempts | 5/hour |
| API general | 100/hour |

## Email-to-Thread

Special email addresses for reply-by-email:

```
d=100&k=key&u=999@mail.loomio.com    # Reply to discussion
group+u=99&k=key@mail.loomio.com     # New thread in group
```

**Source:** `orig/loomio/app/services/received_email_service.rb:62-76`

## Error Responses

```json
{
  "errors": {
    "field_name": ["error message"]
  }
}
```

HTTP Status Codes:
- 401: Unauthenticated
- 403: Unauthorized (CanCan)
- 404: Not found
- 422: Validation failed
- 429: Rate limited

---
