# Rate Limiting - Protected Controllers and Endpoints

## Overview

This document details which controllers and endpoints are protected by rate limiting, and identifies gaps in coverage.

---

## Rack::Attack Protected Endpoints

### Authentication & Registration

| Endpoint | Controller | Methods | Hourly Limit | Notes |
|----------|------------|---------|--------------|-------|
| `/api/v1/sessions` | `Api::V1::SessionsController` | POST | 10 | Login attempts |
| `/api/v1/registrations` | `Api::V1::RegistrationsController` | POST | 10 | Account creation |
| `/api/v1/login_tokens` | `Api::V1::LoginTokensController` | POST | 10 | Magic link requests |
| `/api/v1/identities` | `Api::V1::IdentitiesController` | POST | 10 | OAuth identity linking |

### User Actions

| Endpoint | Controller | Methods | Hourly Limit | Notes |
|----------|------------|---------|--------------|-------|
| `/api/v1/profile` | `Api::V1::ProfileController` | POST/PUT/PATCH | 100 | Profile updates |
| `/api/v1/contact_messages` | `Api::V1::ContactMessagesController` | POST | 10 | Support contact |
| `/api/v1/contact_requests` | `Api::V1::ContactRequestsController` | POST | 10 | Contact form |

### Group & Membership Operations

| Endpoint | Controller | Methods | Hourly Limit | Notes |
|----------|------------|---------|--------------|-------|
| `/api/v1/groups` | `Api::V1::GroupsController` | POST | 20 | Group creation |
| `/api/v1/memberships` | `Api::V1::MembershipsController` | POST/PUT/PATCH | 100 | Membership changes |
| `/api/v1/membership_requests` | `Api::V1::MembershipRequestsController` | POST | 100 | Join requests |
| `/api/v1/announcements` | `Api::V1::AnnouncementsController` | POST | 100 | Invitations/announcements |

### Content Creation

| Endpoint | Controller | Methods | Hourly Limit | Notes |
|----------|------------|---------|--------------|-------|
| `/api/v1/discussions` | `Api::V1::DiscussionsController` | POST/PUT/PATCH | 100 | Thread CRUD |
| `/api/v1/comments` | `Api::V1::CommentsController` | POST/PUT/PATCH | 100 | Comment CRUD |
| `/api/v1/polls` | `Api::V1::PollsController` | POST/PUT/PATCH | 100 | Poll CRUD |
| `/api/v1/stances` | `Api::V1::StancesController` | POST/PUT/PATCH | 100 | Vote submission |
| `/api/v1/outcomes` | `Api::V1::OutcomesController` | POST/PUT/PATCH | 100 | Poll outcome CRUD |
| `/api/v1/reactions` | `Api::V1::ReactionsController` | POST | 100 | Emoji reactions |

### Templates

| Endpoint | Controller | Methods | Hourly Limit | Notes |
|----------|------------|---------|--------------|-------|
| `/api/v1/templates` | Unknown | POST | 10 | Template operations |

### Other

| Endpoint | Controller | Methods | Hourly Limit | Notes |
|----------|------------|---------|--------------|-------|
| `/api/v1/trials` | `Api::V1::TrialsController` | POST | 10 | Trial signups |
| `/api/v1/link_previews` | `Api::V1::LinkPreviewsController` | POST | 100 | URL metadata fetch |
| `/api/v1/discussion_readers` | `Api::V1::DiscussionReadersController` | POST/PUT/PATCH | 1000 | Read state updates |
| `/rails/active_storage/direct_uploads` | ActiveStorage | POST | 20 | File uploads |

---

## ThrottleService Protected Operations

### User Invitations

**File:** `/Users/z/Code/loomio/app/extras/user_inviter.rb`

```ruby
# Line 102-106
ThrottleService.limit!(key: 'UserInviterInvitations',
                       id: actor.id,
                       max: actor.invitations_rate_limit,
                       inc: emails.length + ids.length,
                       per: :day)
```

**Calling Controllers:**

| Controller | Action | Through |
|------------|--------|---------|
| `Api::V1::AnnouncementsController` | `create` | `GroupService.invite`, `DiscussionService.invite`, `PollService.invite`, `OutcomeService.invite` |
| `Api::V1::MembershipsController` | Various | Indirect through services |

**Response on Limit:**
- Raises `ThrottleService::LimitReached`
- **NOT CAUGHT** - results in 500 Internal Server Error

### Email Bounce Notices

**File:** `/Users/z/Code/loomio/app/services/received_email_service.rb`

```ruby
# Line 35
if ThrottleService.can?(key: 'bounce', id: email.sender_email.downcase, max: 1, per: 'hour')
  ForwardMailer.bounce(to: email.sender_name_and_email).deliver_now
end
```

**Not directly controller-triggered** - runs via background processing of received emails.

---

## Devise Lockable Protected Controller

### Sessions Controller

**File:** `/Users/z/Code/loomio/app/controllers/api/v1/sessions_controller.rb`

Inherits from `Devise::SessionsController`. Devise lockable automatically tracks:
- `failed_attempts` column on User model
- `locked_at` timestamp
- `unlock_token` for email unlock

**Lockout behavior:**
- After 20 failed attempts (configurable via `MAX_LOGIN_ATTEMPTS`)
- Account locked for 6 hours OR until email unlock link clicked

---

## Unprotected Endpoints (Potential Gaps)

### GET Requests (Not rate limited by Rack::Attack)

All GET requests are currently unprotected:

| Endpoint | Risk | Recommendation |
|----------|------|----------------|
| `/api/v1/search` | Search abuse | Consider limiting |
| `/api/v1/discussions` | Data scraping | Monitor |
| `/api/v1/groups` | Enumeration | Monitor |
| `/api/v1/polls` | Data scraping | Monitor |
| `/api/v1/events` | Timeline scraping | Monitor |
| `/api/v1/notifications` | Polling abuse | Consider limiting |
| `/api/v1/boot` | Session info | Low risk |

### Bot/Integration APIs (Not visible in Rack::Attack config)

| Endpoint | Purpose | Status |
|----------|---------|--------|
| `/api/b1/*` | Bot API v1 | **UNKNOWN** |
| `/api/b2/*` | Bot API v2 | **UNKNOWN** |
| `/api/b3/*` | Bot API v3 | **UNKNOWN** |

These may be:
1. Intentionally unprotected (bot tokens provide identification)
2. Protected by a different mechanism
3. Overlooked

### Webhook Endpoints

| Endpoint | Purpose | Status |
|----------|---------|--------|
| Various webhook receivers | External integrations | **UNKNOWN** |

---

## Controller Hierarchy

```
ActionController::Base
  |
  +-- ApplicationController
  |     |
  |     +-- Various HTML controllers
  |
  +-- Api::V1::SnorlaxBase
        |
        +-- Api::V1::RestfulController
              |
              +-- Api::V1::GroupsController
              +-- Api::V1::DiscussionsController
              +-- Api::V1::PollsController
              +-- Api::V1::CommentsController
              +-- Api::V1::AnnouncementsController
              +-- (etc.)

Devise::SessionsController
  |
  +-- Api::V1::SessionsController

Devise::RegistrationsController
  |
  +-- Api::V1::RegistrationsController
```

---

## Exception Handling Gap

**Issue:** `ThrottleService::LimitReached` is not caught by any controller.

**Current rescue_from handlers in SnorlaxBase:**

```ruby
rescue_from(CanCan::AccessDenied)                    { |e| respond_with_standard_error e, 403 }
rescue_from(Subscription::MaxMembersExceeded)        { |e| respond_with_standard_error e, 403 }
rescue_from(ActionController::UnpermittedParameters) { |e| respond_with_standard_error e, 400 }
rescue_from(ActionController::ParameterMissing)      { |e| respond_with_standard_error e, 400 }
rescue_from(ActiveRecord::RecordNotFound)            { |e| respond_with_standard_error e, 404 }
rescue_from(ActiveRecord::RecordInvalid)             { |e| respond_with_errors }
```

**Missing:**

```ruby
# Should be added:
rescue_from(ThrottleService::LimitReached) { |e| respond_with_standard_error e, 429 }
```

---

## Recommendations

### Immediate

1. **Add rescue_from for ThrottleService::LimitReached** in `Api::V1::SnorlaxBase`
2. **Audit bot API endpoints** (`/api/b1/`, `/api/b2/`, `/api/b3/`)

### Short-term

3. **Consider GET request rate limiting** for search and heavy endpoints
4. **Add Retry-After headers** to 429 responses

### Long-term

5. **Implement per-user rate limiting** in addition to per-IP
6. **Add rate limit headers** (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`)
