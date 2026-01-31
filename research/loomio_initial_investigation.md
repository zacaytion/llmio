# Loomio Application Investigation

> Initial investigation of the Loomio Ruby on Rails codebase to support a Go rewrite.
> Generated: 2026-01-30

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Application Routes](#2-application-routes)
3. [Database Schema](#3-database-schema)
4. [Application Models](#4-application-models)
5. [API Contracts (Serializers)](#5-api-contracts-serializers)
6. [Test Suite](#6-test-suite)
7. [Background Jobs & Scheduled Tasks](#7-background-jobs--scheduled-tasks)
8. [Real-time System (Channel Server)](#8-real-time-system-channel-server)
9. [Important Comments & Known Issues](#9-important-comments--known-issues)
10. [Key Patterns for Go Rewrite](#10-key-patterns-for-go-rewrite)

---

## 1. Executive Summary

### 1.1 Purpose

Loomio is a collaborative decision-making platform that enables groups to discuss topics, create proposals, and reach consensus through structured voting. It supports threaded discussions, multiple poll types, and real-time collaboration.

### 1.2 Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Framework | Ruby on Rails | 7.2 |
| Database | PostgreSQL | with citext, hstore, pgcrypto extensions |
| Cache/Queue | Redis | |
| Background Jobs | Sidekiq | ~> 7.0 |
| Real-time | Socket.io (Node.js) | Channel server |
| Collaborative Editing | Hocuspocus (Y.js) | |
| Authentication | Devise | with OAuth, SAML support |
| Authorization | CanCan | |
| Search | pg_search | PostgreSQL full-text |
| File Storage | ActiveStorage | |
| Admin | ActiveAdmin | |

### 1.3 Key Architectural Patterns

1. **Service Layer** - Business logic extracted to 46+ service classes (`app/services/`)
2. **Query Objects** - Complex queries encapsulated in dedicated classes (`app/queries/`)
3. **Event-Driven Architecture** - EventBus broadcasts events for async processing
4. **Polymorphic Associations** - Events, reactions, documents are polymorphic
5. **Soft Deletes** - Uses `discard` gem with `discarded_at` timestamp
6. **Counter Caches** - Extensive use for performance (memberships_count, polls_count, etc.)
7. **Multi-version API** - Supports v1 (primary), b1, b2 (legacy), b3 (admin) APIs

### 1.4 Core Domain Concepts

```
User
  └── Membership ──► Group (with subgroups)
                        └── Discussion
                              ├── Comment (threaded)
                              └── Poll
                                    ├── PollOption
                                    ├── Stance (vote)
                                    │     └── StanceChoice
                                    └── Outcome (decision)
```

---

## 2. Application Routes

**Source:** `orig/loomio/config/routes.rb` (471 lines)

### 2.1 Admin Routes

| Path | Controller | Purpose |
|------|------------|---------|
| `/admin/sidekiq` | Sidekiq::Web | Job monitoring (requires admin) |
| `/admin/blazer` | Blazer::Engine | Analytics queries (requires admin) |

**Reference:** `config/routes.rb:16-19`

### 2.2 API Routes - Backend

#### 2.2.1 API v1 (Primary API)

**Base path:** `/api/v1`
**Format:** JSON (default)
**Authentication:** Session-based (Devise)

**Reference:** `config/routes.rb:64-335`

| Resource | Endpoints | Controller |
|----------|-----------|------------|
| **Groups** | CRUD + `suggest_handle`, `count_explore` | `api/v1/groups` |
| **Memberships** | CRUD + `join_group`, `add_to_subgroup`, `resend`, `set_volume`, `make_admin`, `remove_admin` | `api/v1/memberships` |
| **Membership Requests** | CRUD + `approve`, `ignore` | `api/v1/membership_requests` |
| **Discussions** | CRUD + `dashboard`, `inbox`, `search`, `history`, `mark_as_seen`, `mark_as_read`, `close`, `reopen`, `pin`, `move`, `move_comments`, `discard`, `undiscard` | `api/v1/discussions` |
| **Comments** | CRUD + `discard`, `undiscard` | `api/v1/comments` |
| **Polls** | CRUD + `close`, `reopen`, `add_options`, `add_to_thread`, `remind`, `export`, `discard`, `undiscard` | `api/v1/polls` |
| **Stances** | CRUD + `my_stances`, `users`, `revoke`, `make_admin`, `remove_admin` | `api/v1/stances` |
| **Outcomes** | CRUD | `api/v1/outcomes` |
| **Events** | `index`, `remove_from_thread`, `pin` | `api/v1/events` |
| **Notifications** | `index`, `viewed` | `api/v1/notifications` |
| **Reactions** | `index`, `create`, `update`, `destroy` | `api/v1/reactions` |
| **Tags** | CRUD | `api/v1/tags` |
| **Documents** | CRUD + `for_group`, `for_discussion` | `api/v1/documents` |
| **Discussion Templates** | CRUD | `api/v1/discussion_templates` |
| **Poll Templates** | CRUD | `api/v1/poll_templates` |
| **Webhooks** | CRUD | `api/v1/webhooks` |
| **Chatbots** | CRUD + `test`, `event_kinds` | `api/v1/chatbots` |
| **Tasks** | CRUD + `mark_as_done`, `mark_as_not_done`, `update_done` | `api/v1/tasks` |
| **Reports** | `create` | `api/v1/reports` |
| **Profile** | `show`, `update`, `upload_avatar`, `deactivate`, `destroy`, `save_experience`, `email_status`, `set_volume` | `api/v1/profile` |
| **Users** | `index`, `show`, `remind` | `api/v1/users` |
| **Registrations** | `create` | `api/v1/registrations` |
| **Sessions** | `create`, `destroy` | `api/v1/sessions` |
| **Login Tokens** | `create` | `api/v1/login_tokens` |
| **Announcements** | `audience`, `search`, `notify`, `history` | `api/v1/announcements` |
| **Received Emails** | `create`, `release`, `destroy` | `api/v1/received_emails` |
| **Boot** | `site`, `user` | `api/v1/boot` |
| **Links** | `create` | `api/v1/links` |
| **Trials** | `create` | `api/v1/trials` |
| **Attachments** | `create`, `destroy` | `api/v1/attachments` |
| **Hocuspocus** | `create` | `api/v1/hocuspocus` |
| **Search** | `index` | `api/v1/search` |
| **Identities** | `index`, `destroy` | `api/v1/identities` |
| **OAuth Applications** | CRUD | `api/v1/oauth_applications` |

#### 2.2.2 API b1/b2 (Legacy Public APIs)

**Base path:** `/api/b1`, `/api/b2`
**Authentication:** `api_key` query parameter

**Reference:** `config/routes.rb:42-53`

| API | Resources | Notes |
|-----|-----------|-------|
| b1 | discussions, polls, memberships | Legacy public API |
| b2 | discussions, polls, memberships, comments | Adds comments support |

#### 2.2.3 API b3 (Admin User Management)

**Base path:** `/api/b3`
**Authentication:** `B3_API_KEY` environment variable

**Reference:** `config/routes.rb:55-62`

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/b3/users/:id/deactivate` | POST | Deactivate user |
| `/api/b3/users/:id/reactivate` | POST | Reactivate user |

### 2.3 Web Routes - Frontend

**Reference:** `config/routes.rb:338-471`

Most web routes delegate to `application#index` for the Vue.js SPA:

```ruby
# SPA routes (all render the same Vue app shell)
/dashboard, /inbox, /threads, /polls, /explore, /profile
/g/:handle/*, /d/:key/*, /p/:key/*
```

**Special routes:**

| Path | Purpose |
|------|---------|
| `/users/sign_in`, `/users/sign_up` | Devise authentication |
| `/oauth/*`, `/saml/*` | SSO integration |
| `/email_actions/*` | Unsubscribe, mark as read |
| `/dev/*` | Development/testing endpoints |
| `/apple-app-site-association` | iOS app deep linking |
| `/login_token/:token` | Passwordless login |

### 2.4 OAuth Provider Routes

| Provider | Path |
|----------|------|
| Google | `/google/authorize`, `/google/callback` |
| Facebook | `/facebook/authorize`, `/facebook/callback` |
| Slack | `/slack/authorize`, `/slack/callback` |
| Microsoft | `/microsoft/authorize`, `/microsoft/callback` |
| SAML | `/saml/metadata`, `/saml/authorize`, `/saml/callback` |

**Reference:** `config/routes.rb:387-420`

---

## 3. Database Schema

**Source:** `orig/loomio/db/schema.rb` (1093 lines)
**Schema Version:** `2025_12_03_031449`

### 3.1 PostgreSQL Extensions

```ruby
enable_extension "citext"           # Case-insensitive text
enable_extension "hstore"           # Key-value storage
enable_extension "pg_stat_statements"  # Query performance
enable_extension "pgcrypto"         # UUID generation
enable_extension "plpgsql"          # PL/pgSQL
```

### 3.2 Core Tables

#### users (lines 985-1060)

Primary user account table.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `email` | citext | Unique, case-insensitive |
| `encrypted_password` | string(128) | Devise |
| `name` | string(255) | Display name |
| `username` | string(255) | Unique handle |
| `avatar_kind` | string | `initials`, `uploaded`, etc. |
| `time_zone` | string | IANA timezone |
| `selected_locale` | string | Language preference |
| `deactivated_at` | datetime | Soft delete |
| `is_admin` | boolean | System admin |
| `email_verified` | boolean | Email confirmed |
| `secret_token` | string | WebSocket auth (auto-generated UUID) |
| `memberships_count` | integer | Counter cache |
| `experiences` | jsonb | User experience flags |
| `email_when_*` | boolean | Notification preferences |

**Key Indexes:** `email` (unique), `username` (unique), `key` (unique)

#### groups (lines 409-492)

Organizations/workspaces with hierarchical support.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `name` | string(255) | Group name |
| `handle` | citext | Unique URL slug |
| `key` | string(255) | Unique random key |
| `parent_id` | integer | FK to self (subgroups) |
| `subscription_id` | integer | FK to subscriptions |
| `description` | text | Rich text |
| `archived_at` | datetime | Soft archive |
| `is_visible_to_public` | boolean | Public group |
| `members_can_*` | boolean | Permission flags (many) |
| `*_count` | integer | Many counter caches |

**Key Indexes:** `handle` (unique), `key` (unique), `parent_id`

#### discussions (lines 276-323)

Threaded conversation containers.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `group_id` | integer | FK to groups |
| `author_id` | integer | FK to users |
| `title` | string(255) | Discussion title |
| `description` | text | Rich text content |
| `description_format` | string | `md` or `html` |
| `key` | string(255) | Unique URL key |
| `private` | boolean | Visibility |
| `closed_at` | datetime | Closed state |
| `pinned_at` | datetime | Pinned state |
| `discarded_at` | datetime | Soft delete |
| `max_depth` | integer | Comment nesting depth |
| `newest_first` | boolean | Sort order |
| `tags` | string[] | Array of tag names |
| `attachments` | jsonb | Attached files |
| `link_previews` | jsonb | URL previews |

**Key Indexes:** `key` (unique), `group_id`, `author_id`, `tags` (GIN)

#### comments (lines 177-197)

Threaded messages within discussions.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `discussion_id` | integer | FK to discussions |
| `user_id` | integer | Author FK |
| `parent_id` | integer | Parent comment (threading) |
| `parent_type` | string | Polymorphic parent |
| `body` | text | Rich text content |
| `body_format` | string | `md` or `html` |
| `discarded_at` | datetime | Soft delete |
| `edited_at` | datetime | Last edit time |
| `attachments` | jsonb | Attached files |

#### polls (lines 750-811)

Decision-making polls with multiple types.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `author_id` | integer | FK to users |
| `discussion_id` | integer | FK (optional) |
| `group_id` | integer | FK to groups |
| `title` | string | Poll question |
| `details` | text | Rich text description |
| `poll_type` | string | `proposal`, `poll`, `count`, `score`, `ranked_choice`, `meeting` |
| `key` | string | Unique URL key |
| `closing_at` | datetime | When poll closes |
| `closed_at` | datetime | Actually closed time |
| `anonymous` | boolean | Hide voter identity |
| `specified_voters_only` | boolean | Invite-only voting |
| `hide_results` | integer | enum: 0=off, 1=until_vote, 2=until_closed |
| `voters_count` | integer | Counter cache |
| `undecided_voters_count` | integer | Counter cache |
| `stance_counts` | jsonb | Vote tallies |

**Key Indexes:** `key` (unique), `discussion_id`, `group_id`

#### poll_options (lines 681-698)

Individual choices within a poll.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `poll_id` | integer | FK to polls |
| `name` | string | Option text |
| `priority` | integer | Display order |
| `icon` | string | Icon name |
| `meaning` | string | Semantic meaning |
| `total_score` | integer | Sum of all votes |
| `voter_count` | integer | Count of voters |
| `voter_scores` | jsonb | `{user_id: score}` |

#### stances (lines 857-886)

Individual votes/positions on polls.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `poll_id` | integer | FK to polls |
| `participant_id` | integer | FK to users |
| `reason` | string | Vote reason |
| `latest` | boolean | Most recent vote |
| `cast_at` | datetime | When vote was cast |
| `token` | string | Guest access token |
| `admin` | boolean | Poll admin |
| `guest` | boolean | Guest voter |
| `option_scores` | jsonb | `{poll_option_id: score}` |
| `revoked_at` | datetime | Access revoked |

**Key Indexes:** `(poll_id, participant_id, latest)` unique where latest=true

#### stance_choices (lines 837-845)

Links stances to selected poll options (many-to-many).

| Column | Type | Notes |
|--------|------|-------|
| `stance_id` | integer | FK to stances |
| `poll_option_id` | integer | FK to poll_options |
| `score` | integer | Vote weight/score |

#### outcomes (lines 635-651)

Decision statements after poll closes.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `poll_id` | integer | FK to polls |
| `author_id` | integer | FK to users |
| `statement` | text | Decision text |
| `latest` | boolean | Most recent outcome |
| `review_on` | date | Review reminder date |
| `poll_option_id` | integer | Selected option (for meetings) |

#### memberships (lines 534-558)

User roles in groups.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `group_id` | integer | FK to groups |
| `user_id` | integer | FK to users |
| `inviter_id` | integer | Who invited |
| `admin` | boolean | Group admin |
| `delegate` | boolean | Delegation rights |
| `volume` | integer | Notification level |
| `token` | string | Invite token |
| `accepted_at` | datetime | When accepted |
| `revoked_at` | datetime | When removed |

**Key Indexes:** `(group_id, user_id)` unique

#### discussion_readers (lines 221-244)

Tracks who has access to and read discussions.

| Column | Type | Notes |
|--------|------|-------|
| `user_id` | integer | FK to users |
| `discussion_id` | integer | FK to discussions |
| `last_read_at` | datetime | Last viewed |
| `read_ranges_string` | string | Optimized read tracking |
| `volume` | integer | Notification level |
| `guest` | boolean | Guest reader |
| `admin` | boolean | Discussion admin |
| `inviter_id` | integer | Who invited |
| `token` | string | Guest access token |
| `revoked_at` | datetime | Access revoked |

**Key Indexes:** `(user_id, discussion_id)` unique, `token` unique

#### events (lines 346-373)

Activity timeline and notification source.

| Column | Type | Notes |
|--------|------|-------|
| `id` | serial | Primary key |
| `kind` | string | Event type (`new_discussion`, `new_comment`, etc.) |
| `eventable_id` | integer | Polymorphic FK |
| `eventable_type` | string | Polymorphic type |
| `user_id` | integer | Actor FK |
| `discussion_id` | integer | FK (for threading) |
| `sequence_id` | integer | Order in discussion |
| `parent_id` | integer | Parent event |
| `position_key` | string | Hierarchical position |
| `depth` | integer | Nesting depth |
| `pinned` | boolean | Pinned event |
| `custom_fields` | jsonb | Event metadata |

**Key Indexes:** `(discussion_id, sequence_id)` unique, `position_key`

### 3.3 Supporting Tables

| Table | Purpose | Lines |
|-------|---------|-------|
| `notifications` | User notifications | 560-573 |
| `reactions` | Emoji reactions (polymorphic) | 813-823 |
| `tags` | Group tags | 923-935 |
| `taggings` | Tag assignments (polymorphic) | 913-921 |
| `documents` | File attachments (legacy) | 325-344 |
| `translations` | Content translations | 969-978 |
| `tasks` | Todo items within content | 937-958 |
| `webhooks` | Outgoing webhooks | 1072-1089 |
| `chatbots` | Chat integrations | 156-170 |
| `subscriptions` | Billing plans | 888-911 |
| `login_tokens` | Passwordless auth | 494-504 |
| `omniauth_identities` | OAuth connections | 619-633 |
| `received_emails` | Inbound email handling | 825-835 |
| `versions` | Paper Trail audit log | 1062-1070 |

### 3.4 Template Tables

| Table | Purpose | Lines |
|-------|---------|-------|
| `discussion_templates` | Reusable discussion formats | 246-274 |
| `poll_templates` | Reusable poll configurations | 700-748 |

### 3.5 ActiveStorage Tables

| Table | Purpose |
|-------|---------|
| `active_storage_blobs` | File metadata |
| `active_storage_attachments` | File associations |
| `active_storage_variant_records` | Image variants |

---

## 4. Application Models

**Source:** `orig/loomio/app/models/`

### 4.1 User Model

**File:** `orig/loomio/app/models/user.rb` (377 lines)

#### Concerns Included
- `CustomCounterCache::Model` - Efficient counter updates
- `ReadableUnguessableUrls` - URL key generation
- `MessageChannel` - WebSocket channel names
- `HasExperiences` - Feature flags per user
- `HasAvatar` - Avatar management
- `HasRichText` - Rich text fields
- `SelfReferencing` - Related users

#### Key Relationships

```ruby
# Memberships
has_many :memberships                           # line 63
has_many :admin_memberships                     # lines 59-61
has_many :groups, through: :memberships         # lines 76-78

# Content authored
has_many :authored_discussions, class_name: 'Discussion', foreign_key: :author_id  # line 82
has_many :authored_polls, class_name: 'Poll', foreign_key: :author_id              # line 83
has_many :stances                               # line 89
has_many :comments                              # line 99

# Reading state
has_many :discussion_readers                    # line 93
```

#### Key Methods

```ruby
def is_member_of?(group)     # line 263 - Check group membership
def is_admin_of?(group)      # line 267 - Check admin status
def is_paying?               # line 186 - Check subscription status
def email_api_key            # line 320 - For email-to-thread feature
def secret_token             # Auto-generated UUID for WebSocket auth
```

### 4.2 Group Model

**File:** `orig/loomio/app/models/group.rb` (476 lines)

#### Concerns Included
- `HasRichText`
- `HasEvents`
- `SelfReferencing` - Subgroup hierarchy
- `GroupPrivacy` - Privacy logic

#### Key Relationships

```ruby
# Hierarchy
belongs_to :parent, class_name: 'Group'         # line 22
has_many :subgroups, class_name: 'Group', foreign_key: :parent_id  # lines 70-73

# Members
has_many :memberships                           # line 38
has_many :members, through: :memberships        # line 39
has_many :admin_memberships                     # line 47
has_many :admins, through: :admin_memberships   # line 48

# Content
has_many :discussions                           # line 30
has_many :polls                                 # line 53
has_many :tags                                  # line 56
```

#### Key Methods

```ruby
def add_member!(user, inviter:)    # lines 251-267
def add_admin!(user)               # lines 277-282
def is_subgroup?                   # line 328
def archive!                       # line 288
def unarchive!                     # line 296
```

### 4.3 Discussion Model

**File:** `orig/loomio/app/models/discussion.rb` (287 lines)

#### Concerns Included
- `Discard::Model` - Soft delete
- `HasRichText`
- `HasEvents`
- `Reactable`
- `HasMentions`
- `Searchable`
- `HasCreatedEvent`

#### Key Relationships

```ruby
belongs_to :group                               # line 76
belongs_to :author, class_name: 'User'          # line 77
belongs_to :closer, class_name: 'User'          # line 79

has_many :polls                                 # line 80
has_many :comments                              # line 83
has_many :discussion_readers                    # line 91
has_many :items, class_name: 'Event'            # line 89 - Timeline
```

#### Key Methods

```ruby
def members         # lines 148-154 - Group members + guests
def admins          # lines 156-162 - Discussion admins
def guests          # lines 164-169 - Non-member guests
def add_guest!(user, inviter)  # lines 175-181
def public?         # line 224 - Opposite of private
```

### 4.4 Poll Model

**File:** `orig/loomio/app/models/poll.rb` (546 lines)

#### Poll Types

| Type | Description |
|------|-------------|
| `proposal` | Yes/No/Abstain consensus |
| `poll` | Multiple choice |
| `count` | Simple headcount |
| `score` | Numeric rating |
| `ranked_choice` | Preference ordering |
| `meeting` | Date/time scheduling |
| `dot_vote` | Budget allocation |

#### Key Relationships

```ruby
belongs_to :author, class_name: 'User'          # line 142
belongs_to :discussion, optional: true          # line 146
belongs_to :group                               # line 147

has_many :poll_options                          # line 161
has_many :stances                               # line 153
has_many :voters, through: :stances             # line 155
has_many :outcomes                              # line 143
```

#### Key Methods

```ruby
def active?                     # line 447 - Poll is open
def closed?                     # line 455
def results                     # line 290 - Calculate results
def show_results?(voted:)       # line 372 - Check visibility
def quorum_reached?             # line 359
```

### 4.5 Comment Model

**File:** `orig/loomio/app/models/comment.rb` (162 lines)

#### Key Relationships

```ruby
belongs_to :discussion                          # line 54
belongs_to :user                                # line 55 (author)
belongs_to :parent, polymorphic: true           # line 56 (Discussion/Comment/Stance)
has_many :documents                             # line 58
```

#### Validation: Parent Reparenting

```ruby
# lines 136-143: If someone replies to a deleted comment via email,
# reparent to the discussion
def parent_comment_belongs_to_same_discussion
  self.parent = self.discussion if parent.nil? && discussion.present?
  unless discussion_id == parent.discussion_id
    errors.add(:parent, "Needs to have same discussion id")
  end
end
```

### 4.6 Stance Model

**File:** `orig/loomio/app/models/stance.rb` (321 lines)

#### Key Relationships

```ruby
belongs_to :poll                                # line 61
belongs_to :participant, class_name: 'User'     # line 71
belongs_to :inviter, class_name: 'User'         # line 62
has_many :stance_choices                        # line 64
has_many :poll_options, through: :stance_choices # line 65
```

#### Scopes

```ruby
scope :latest, -> { where(latest: true) }       # line 77
scope :guests, -> { where(guest: true) }        # line 78
scope :undecided, -> { where(cast_at: nil) }    # line 88
scope :decided, -> { where.not(cast_at: nil) }  # line 87
```

#### Anonymous Poll Handling

```ruby
def participant              # line 211
  poll.anonymous ? nil : real_participant
end

def real_participant         # line 215
  @real_participant ||= cache_fetch(:users_by_id, participant_id) { User.find(participant_id) }
end
```

### 4.7 Membership Model

**File:** `orig/loomio/app/models/membership.rb` (101 lines)

#### Key Relationships

```ruby
belongs_to :user                                # line 24
belongs_to :group                               # line 23
belongs_to :inviter, class_name: 'User'         # line 25
belongs_to :revoker, class_name: 'User'         # line 26
```

#### Scopes

```ruby
scope :active, -> { where(revoked_at: nil) }    # line 30
scope :pending, -> { where(accepted_at: nil) }  # line 31
scope :admin, -> { where(admin: true) }         # line 41
```

### 4.8 DiscussionReader Model

**File:** `orig/loomio/app/models/discussion_reader.rb` (138 lines)

Tracks who has access to and has read discussions.

#### Key Methods

```ruby
def viewed!(ranges)          # line 56 - Mark as read
def has_read?(ranges)        # line 62
def unread_ranges            # line 113
```

### 4.9 Event Model

**File:** `orig/loomio/app/models/event.rb` (100 lines)

#### Key Relationships

```ruby
belongs_to :eventable, polymorphic: true        # line 7
belongs_to :discussion                          # line 8
belongs_to :user                                # line 9 (actor)
belongs_to :parent, class_name: 'Event'         # line 10
has_many :notifications                         # line 6
```

#### Event Types (kinds)

| Kind | Eventable Type | Description |
|------|----------------|-------------|
| `new_discussion` | Discussion | Discussion created |
| `discussion_edited` | Discussion | Discussion modified |
| `discussion_closed` | Discussion | Discussion closed |
| `new_comment` | Comment | Comment posted |
| `poll_created` | Poll | Poll started |
| `poll_closed` | Poll | Poll ended |
| `stance_created` | Stance | Vote cast |
| `outcome_created` | Outcome | Decision published |
| `membership_created` | Membership | User joined |
| `user_mentioned` | Comment/Discussion | @mention |

#### Trigger Pattern

```ruby
# line 96-98
def trigger!
  EventBus.broadcast("#{kind}_event", self)
end
```

### 4.10 Ability (Authorization)

**File:** `orig/loomio/app/models/ability/base.rb` (56 lines)

Uses CanCan with modular ability files:

```ruby
# 25 ability modules prepended:
prepend Ability::Comment
prepend Ability::Discussion
prepend Ability::Document
prepend Ability::Group
prepend Ability::Membership
prepend Ability::Poll
prepend Ability::Stance
prepend Ability::User
# ... etc
```

**Directory:** `orig/loomio/app/models/ability/` (25 files)

---

## 5. API Contracts (Serializers)

**Source:** `orig/loomio/app/serializers/` (49 files)

### 5.1 Base Serializer Pattern

**File:** `orig/loomio/app/serializers/application_serializer.rb` (180 lines)

```ruby
class ApplicationSerializer < ActiveModel::Serializer
  embed :ids, include: true  # JSON-API style compound documents

  # Cache helper for N+1 prevention
  def cache_fetch(key_or_keys, id, &block)
    # Fetches from scope[:cache] or yields block
  end

  # Conditional field inclusion
  def include_type?(type)
    !exclude_type?(type)
  end

  # Soft delete field hiding
  def self.hide_when_discarded(names)
    # Nullifies fields when discarded_at is set
  end
end
```

### 5.2 User Serializers

#### AuthorSerializer
**File:** `orig/loomio/app/serializers/author_serializer.rb` (51 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `name` | string | |
| `username` | string | |
| `email` | string | Conditional: self, admin, or include_email |
| `avatar_initials` | string | |
| `avatar_kind` | string | Custom: 'mdi-email-outline' if unverified |
| `thumb_url` | string | Avatar thumbnail |
| `time_zone` | string | |
| `locale` | string | |
| `email_verified` | boolean | |
| `bot` | boolean | |
| `titles` | object | From experiences |
| `delegates` | object | From experiences |

#### CurrentUserSerializer
**File:** `orig/loomio/app/serializers/current_user_serializer.rb` (24 lines)

Extends UserSerializer with sensitive fields:
- `email` (always)
- `secret_token`
- `email_*` preferences
- `memberships_count`
- `is_admin`

### 5.3 Group Serializer

**File:** `orig/loomio/app/serializers/group_serializer.rb` (135 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `key` | string | Unique URL key |
| `handle` | string | URL slug |
| `name` | string | |
| `full_name` | string | With parent prefix |
| `description` | text | Rich text |
| `logo_url` | string | From self or parent |
| `cover_url` | string | From self or parent |
| `is_visible_to_public` | boolean | |
| `members_can_*` | boolean | Permission flags (many) |
| `*_count` | integer | Counter caches |
| `subscription` | object | Plan details |

**Nested:** `parent`, `current_user_membership`, `tags`

### 5.4 Discussion Serializer

**File:** `orig/loomio/app/serializers/discussion_serializer.rb` (99 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `key` | string | URL key |
| `group_id` | integer | |
| `title` | string | Hidden when discarded |
| `description` | text | Hidden when discarded |
| `private` | boolean | |
| `closed_at` | datetime | |
| `pinned_at` | datetime | |
| `tags` | array | |
| `items_count` | integer | |
| `members_count` | integer | |
| **Reader fields:** | | Via `attributes_from_reader` macro |
| `discussion_reader_id` | integer | |
| `last_read_at` | datetime | Null if anonymous polls |
| `read_ranges` | string | Empty if anonymous polls |
| `guest` | boolean | |
| `admin` | boolean | |

**Nested:** `author`, `group`, `active_polls`, `created_event`

### 5.5 Poll Serializer

**File:** `orig/loomio/app/serializers/poll_serializer.rb` (144 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `key` | string | URL key |
| `title` | string | |
| `details` | text | |
| `poll_type` | string | Type of poll |
| `anonymous` | boolean | |
| `closing_at` | datetime | |
| `closed_at` | datetime | |
| `voters_count` | integer | |
| `results` | object | Conditional: `show_results?(voted: true)` |
| `stance_counts` | array | Conditional: `show_results?(voted: true)` |
| `my_stance` | object | Current user's vote |
| `poll_option_names` | array | |

**Nested:** `discussion`, `group`, `author`, `current_outcome`, `poll_options`

### 5.6 Stance Serializer

**File:** `orig/loomio/app/serializers/stance_serializer.rb` (76 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `poll_id` | integer | |
| `participant_id` | integer | Null if anonymous |
| `reason` | text | Conditional: owner or results visible |
| `option_scores` | object | `{poll_option_id: score}` |
| `cast_at` | datetime | |
| `latest` | boolean | |
| `admin` | boolean | |
| `guest` | boolean | |

**Nested:** `poll`, `participant` (null if anonymous)

### 5.7 Comment Serializer

**File:** `orig/loomio/app/serializers/comment_serializer.rb` (27 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `discussion_id` | integer | |
| `author_id` | integer | |
| `body` | text | Hidden when discarded |
| `body_format` | string | `md` or `html` |
| `parent_id` | integer | |
| `parent_type` | string | |
| `created_at` | datetime | |
| `discarded_at` | datetime | |

**Nested:** `author`, `discussion`

### 5.8 Event Serializer

**File:** `orig/loomio/app/serializers/event_serializer.rb` (63 lines)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer | |
| `kind` | string | Event type |
| `sequence_id` | integer | Order in discussion |
| `position_key` | string | Hierarchical position |
| `depth` | integer | |
| `parent_id` | integer | |
| `discussion_id` | integer | |
| `eventable_id` | integer | |
| `eventable_type` | string | |
| `created_at` | datetime | |
| `pinned` | boolean | |
| `custom_fields` | object | For specific event kinds |

**Nested:** `actor`, `eventable` (polymorphic), `discussion`, `parent`

### 5.9 API Response Format

Standard response structure:

```json
{
  "events": [...],           // Root key from serializer
  "discussions": [...],      // Sideloaded associations
  "users": [...],
  "groups": [...],
  "meta": {
    "root": "events",
    "total": 100
  }
}
```

---

## 6. Test Suite

**Source:** `orig/loomio/spec/`

### 6.1 Test Framework

| Component | Technology |
|-----------|------------|
| Framework | RSpec |
| Factories | FactoryBot |
| Database Cleaning | DatabaseCleaner |
| HTTP Mocking | WebMock |
| Job Testing | Sidekiq (inline mode) |

**Configuration:** `orig/loomio/spec/rails_helper.rb` (103 lines)

### 6.2 Test Statistics

| Category | Files | Directory |
|----------|-------|-----------|
| Controllers | 49 | `spec/controllers/` |
| Models | 26 | `spec/models/` |
| Services | 17 | `spec/services/` |
| Queries | 4 | `spec/queries/` |
| Workers | 1 | `spec/workers/` |
| Mailers | 1 | `spec/mailers/` |
| Mailboxes | 2 | `spec/mailboxes/` |
| Helpers | 3 | `spec/helpers/` |
| Extras | 6 | `spec/extras/` |
| **Total** | **116** | |

### 6.3 Comprehensive Test Listing

#### Controllers - API v1 (18 files)

| File | Path |
|------|------|
| `announcements_controller_spec.rb` | `spec/controllers/api/v1/` |
| `attachments_controller_spec.rb` | `spec/controllers/api/v1/` |
| `comments_controller_spec.rb` | `spec/controllers/api/v1/` |
| `discussions_controller_spec.rb` | `spec/controllers/api/v1/` |
| `documents_controller_spec.rb` | `spec/controllers/api/v1/` |
| `events_controller_spec.rb` | `spec/controllers/api/v1/` |
| `groups_controller_spec.rb` | `spec/controllers/api/v1/` |
| `login_tokens_controller_spec.rb` | `spec/controllers/api/v1/` |
| `membership_requests_controller_spec.rb` | `spec/controllers/api/v1/` |
| `memberships_controller_spec.rb` | `spec/controllers/api/v1/` |
| `mentions_controller_spec.rb` | `spec/controllers/api/v1/` |
| `outcomes_controller_spec.rb` | `spec/controllers/api/v1/` |
| `polls_controller_spec.rb` | `spec/controllers/api/v1/` |
| `profile_controller_spec.rb` | `spec/controllers/api/v1/` |
| `registrations_controller_spec.rb` | `spec/controllers/api/v1/` |
| `search_controller_spec.rb` | `spec/controllers/api/v1/` |
| `sessions_controller_spec.rb` | `spec/controllers/api/v1/` |
| `stances_controller_spec.rb` | `spec/controllers/api/v1/` |

#### Controllers - API b2 (4 files)

| File | Path |
|------|------|
| `comments_controller_spec.rb` | `spec/controllers/api/b2/` |
| `discussions_controller_spec.rb` | `spec/controllers/api/b2/` |
| `memberships_controller_spec.rb` | `spec/controllers/api/b2/` |
| `polls_controller_spec.rb` | `spec/controllers/api/b2/` |

#### Controllers - API b3 (1 file)

| File | Path |
|------|------|
| `users_controller_spec.rb` | `spec/controllers/api/b3/` |

#### Controllers - Web (26 files)

| File | Path |
|------|------|
| `discussions_controller_spec.rb` | `spec/controllers/` |
| `groups_controller_spec.rb` | `spec/controllers/` |
| `polls_controller_spec.rb` | `spec/controllers/` |
| `users_controller_spec.rb` | `spec/controllers/` |
| `email_actions_controller_spec.rb` | `spec/controllers/` |
| `manifest_controller_spec.rb` | `spec/controllers/` |
| `redirect_controller_spec.rb` | `spec/controllers/` |
| `start_controller_spec.rb` | `spec/controllers/` |
| `robots_controller_spec.rb` | `spec/controllers/` |
| `oauth_controller_spec.rb` | `spec/controllers/identities/` |
| `saml_controller_spec.rb` | `spec/controllers/identities/` |
| ... | |

#### Models (26 files)

| File | Key Test Coverage |
|------|-------------------|
| `user_spec.rb` | Authentication, memberships, deactivation |
| `group_spec.rb` | Hierarchy, permissions, archiving |
| `discussion_spec.rb` | Privacy, guest access, closing |
| `poll_spec.rb` | Voting, closing, results calculation |
| `comment_spec.rb` | Threading, soft delete |
| `stance_spec.rb` | Voting, anonymous handling |
| `membership_spec.rb` | Roles, invitations |
| `event_spec.rb` | Event creation, sequencing |
| `discussion_reader_spec.rb` | Read tracking |
| `login_token_spec.rb` | Passwordless auth |
| `outcome_spec.rb` | Decision statements |
| `poll_option_spec.rb` | Vote options |
| `stance_choice_spec.rb` | Vote selections |
| `ability_spec.rb` | Authorization rules |
| `ability/discussion_spec.rb` | Discussion permissions |
| `ability/poll_spec.rb` | Poll permissions |
| `concerns/has_avatar_spec.rb` | Avatar handling |
| `concerns/events/position_spec.rb` | Event positioning |
| `discussion_event_integration_spec.rb` | Event/discussion integration |
| `group_privacy_spec.rb` | Privacy rules |
| `events/` | Event-specific tests (3 files) |

#### Services (17 files)

| File | Key Test Coverage |
|------|-------------------|
| `comment_service_spec.rb` | Comment CRUD, notifications |
| `discussion_service_spec.rb` | Discussion lifecycle |
| `group_service_spec.rb` | Group management |
| `poll_service_spec.rb` | Poll lifecycle, voting |
| `user_service_spec.rb` | User management |
| `membership_service_spec.rb` | Member operations |
| `stance_service_spec.rb` | Vote handling |
| `reaction_service_spec.rb` | Reaction handling |
| `event_service_spec.rb` | Event creation |
| `outcome_service_spec.rb` | Outcome management |
| `discussion_reader_service_spec.rb` | Read tracking |
| `login_token_service_spec.rb` | Passwordless auth |
| `task_service_spec.rb` | Task management |
| `received_email_service_spec.rb` | Email routing |
| `record_cloner_spec.rb` | Record duplication |
| `group_export_service_spec.rb` | Data export |
| `throttle_service_spec.rb` | Rate limiting |
| `translation_service_spec.rb` | Translation |
| `retry_on_error_spec.rb` | Error retry logic |
| `group_service/privacy_change_spec.rb` | Privacy changes |

#### Queries (4 files)

| File | Path |
|------|------|
| `discussion_query_spec.rb` | `spec/queries/` |
| `group_query_spec.rb` | `spec/queries/` |
| `poll_query_spec.rb` | `spec/queries/` |
| `user_query_spec.rb` | `spec/queries/` |

#### Workers (1 file)

| File | Test Cases |
|------|------------|
| `migrate_user_worker_spec.rb` | User migration (2 test cases) |

#### Mailers (1 file)

| File | Path |
|------|------|
| `user_mailer_spec.rb` | `spec/mailers/` |

#### Mailboxes (2 files)

| File | Path |
|------|------|
| `received_email_mailbox_spec.rb` | `spec/mailboxes/` |
| `received_email_mailbox_mixed_spec.rb` | `spec/mailboxes/` |

#### Helpers (3 files)

| File | Path |
|------|------|
| `email_helper_spec.rb` | `spec/helpers/` |
| `locales_helper_spec.rb` | `spec/helpers/` |
| `pretty_url_helper_spec.rb` | `spec/helpers/` |

#### Extras (6 files)

| File | Path |
|------|------|
| `event_bus_spec.rb` | `spec/extras/` |
| `model_locator_spec.rb` | `spec/extras/` |
| `range_set_spec.rb` | `spec/extras/` |
| `time_zone_to_city_spec.rb` | `spec/extras/` |
| `username_generator_spec.rb` | `spec/extras/` |
| Query files (3) | `spec/extras/queries/` |

### 6.4 Test Factories

**File:** `orig/loomio/spec/factories.rb` (8662 bytes)

Key factories defined:

```ruby
factory :user
factory :admin_user
factory :unverified_user
factory :group
factory :membership
factory :pending_membership
factory :tag
factory :identity
factory :login_token
factory :discussion
factory :poll
factory :comment
factory :reaction
factory :stance
factory :outcome
factory :event
factory :version
```

### 6.5 Test Support

| File | Purpose |
|------|---------|
| `spec/support/database_cleaner.rb` | Transaction/deletion strategies |
| `spec/support/devise.rb` | Authentication helpers |
| `spec/support/mailer_macros.rb` | Email testing utilities |

### 6.6 Test Helpers (rails_helper.rb)

```ruby
def fixture_for(path, filetype)    # lines 84-86
def described_model_name           # lines 88-90
def emails_sent_to(address)        # lines 92-94
def last_email                     # lines 96-98
def last_email_html_body           # lines 100-102
```

---

## 7. Background Jobs & Scheduled Tasks

### 7.1 Job Framework

**Framework:** Sidekiq ~> 7.0
**Configuration:** `orig/loomio/config/sidekiq.yml`
**Initializer:** `orig/loomio/config/initializers/sidekiq.rb`

**Queues (in priority order):**

| Queue | Priority | Purpose |
|-------|----------|---------|
| `critical` | Highest | Critical operations |
| `login_emails` | High | Authentication emails |
| `mailers` | High | General emails |
| `notification_emails` | High | Notification delivery |
| `default` | Normal | Default queue |
| `action_mailbox_routing` | Normal | Inbound email |
| `active_storage_analysis` | Low | File processing |
| `active_storage_purge` | Low | File cleanup |
| `low` | Low | Non-urgent tasks |
| `low_priority` | Lowest | Background tasks |

### 7.2 All Job Classes (38 total)

**Source:** `orig/loomio/app/workers/`

| Worker | Purpose | Queue | Retry |
|--------|---------|-------|-------|
| `AcceptMembershipWorker` | Accept pending memberships | default | yes |
| `AddGroupIdToDocumentsWorker` | Add group_id to documents | default | yes |
| `AddHeadingIdsWorker` | Add IDs to headings | default | yes |
| `AnnounceDiscussionWorker` | Announce new discussions | default | yes |
| `AppendTranscriptWorker` | Append transcripts | default | yes |
| `AttachDocumentWorker` | Handle document attachments | default | yes |
| `CloseExpiredPollWorker` | Close expired polls | default | yes |
| `ConvertDiscussionTemplatesWorker` | Migrate discussion templates | default | yes |
| `ConvertPollStancesInDiscussionWorker` | Convert poll stances | default | yes |
| `DeactivateUserWorker` | Deactivate users | default | yes |
| `DestroyDiscussionWorker` | Delete discussions | default | yes |
| `DestroyGroupWorker` | Delete groups | default | yes |
| `DestroyRecordWorker` | Generic record deletion | default | yes |
| `DestroyTagWorker` | Delete tags | default | yes |
| `DestroyUserWorker` | Delete user accounts | default | yes |
| `DownloadAttachmentWorker` | Download attachments | default | yes |
| `FixStancesMissingFromThreadsWorker` | Repair missing stances | default | yes |
| `GenericWorker` | Call arbitrary service methods | default | yes |
| `GeoLocationWorker` | Process geolocation | default | yes |
| `GroupExportCsvWorker` | Export group to CSV | default | yes |
| `GroupExportWorker` | Export group data | default | yes |
| `MigrateDiscussionReadersForDeactivatedMembersWorker` | Migrate readers | default | yes |
| `MigrateGuestOnDiscussionReadersAndStances` | Migrate guest data | default | yes |
| `MigratePollTemplatesWorker` | Convert poll templates | default | yes |
| `MigrateTagsWorker` | Migrate tags | default | yes |
| `MigrateUserWorker` | Merge user accounts | default | yes |
| `MoveCommentsWorker` | Move comments between discussions | default | yes |
| `PublishEventWorker` | Trigger event publishing | default | yes |
| `RedactUserWorker` | Anonymize user data | default | yes |
| `RemovePollExpiredFromThreadsWorker` | Remove expired poll indicators | default | yes |
| `RepairThreadWorker` | Repair thread integrity | default | yes |
| `ResetPollStanceDataWorker` | Reset stance data | low | no |
| `RevokeMembershipsOfDeactivatedUsersWorker` | Batch revoke memberships | default | yes |
| `SendDailyCatchUpEmailWorker` | Daily digest emails | default | no |
| `UndeleteBlobWorker` | Recover deleted files | default | yes |
| `UpdateBlockedDomainsWorker` | Update domain blocklist | default | yes |
| `UpdatePollCountsWorker` | Update vote counters | low | no |
| `UpdateTagWorker` | Update tag metadata | default | yes |

### 7.3 Key Worker Details

#### DeactivateUserWorker
**File:** `orig/loomio/app/workers/deactivate_user_worker.rb` (24 lines)

```ruby
def perform(user_id, deactivated_by_id)
  user = User.find(user_id)
  user.update!(deactivated_at: Time.now, deactivator_id: deactivated_by_id)
  user.memberships.update_all(revoked_at: Time.now, revoker_id: deactivated_by_id)
  # Reindex search data
end
```

#### MigrateUserWorker
**File:** `orig/loomio/app/workers/migrate_user_worker.rb` (85 lines)

Complex job for merging two user accounts. Updates all references across:
- comments, discussions, polls, stances, outcomes
- memberships, discussion_readers, events
- notifications, reactions, documents

#### SendDailyCatchUpEmailWorker
**File:** `orig/loomio/app/workers/send_daily_catch_up_email_worker.rb` (24 lines)

```ruby
def perform
  # Runs hourly, sends at 6 AM in user's timezone
  # Supports daily, weekly, bi-weekly frequencies
  # Respects user.email_catch_up_day preference
end
```

#### GenericWorker
**File:** `orig/loomio/app/workers/generic_worker.rb` (6 lines)

Meta-worker for calling arbitrary service methods:

```ruby
def perform(class_name, method_name, *args)
  class_name.constantize.send(method_name, *args)
end
```

### 7.4 Scheduled Tasks

**Source:** `orig/loomio/lib/tasks/loomio.rake` (276 lines)

Scheduled via external cron calling rake tasks:

#### Hourly (`rake loomio:hourly_tasks`)
**Lines 222-248**

| Job | Purpose |
|-----|---------|
| `ThrottleService.reset!('hour')` | Reset hourly rate limits |
| `PollService.expire_lapsed_polls` | Close overdue polls |
| `PollService.publish_closing_soon` | Notify closing soon |
| `TaskService.send_task_reminders` | Task due reminders |
| `ReceivedEmailService.route_all` | Process inbound emails |
| `LoginToken.delete` (> 1 hour) | Clean old tokens |
| `GeoLocationWorker` | Geo processing |
| `SendDailyCatchUpEmailWorker` | Digest emails (at 6 AM per user TZ) |

#### Daily (at hour 0)
**Lines 250-262**

| Job | Purpose |
|-----|---------|
| `ThrottleService.reset!('day')` | Reset daily rate limits |
| `Group.expired_demo.delete_all` | Remove demo groups |
| `DemoService.generate_demo_groups` | Create new demos |
| `CleanupService.delete_orphan_records` | Clean orphan records |
| `OutcomeService.publish_review_due` | Outcome review reminders |
| `ReceivedEmailService.delete_old_emails` | Clean old emails |
| `DemoService.ensure_queue` | Ensure demo queue |

#### Monthly (1st of month)
**Line 249**

| Job | Purpose |
|-----|---------|
| `UpdateBlockedDomainsWorker` | Update email domain blocklist |

#### Weekly (Sundays)
**Lines 264-270**

| Job | Purpose |
|-----|---------|
| `refresh_expiring_chargify_management_links` | Refresh billing links |
| `populate_chargify_management_links` | Populate billing links |

---

## 8. Real-time System (Channel Server)

**Source:** `orig/loomio_channel_server/`

### 8.1 Architecture Overview

The channel server is a Node.js application providing:

1. **Live Updates (Socket.io)** - Push record changes to connected clients
2. **Collaborative Editing (Hocuspocus)** - Y.js-based real-time editing
3. **Matrix Bot Integration** - Chat platform integration

### 8.2 Technology Stack

| Component | Version |
|-----------|---------|
| socket.io | 4.7.1 |
| @hocuspocus/server | 3.4.0 |
| redis | 4.6.15 |
| matrix-bot-sdk | 0.5.19 |
| @sentry/node | Error tracking |

### 8.3 Communication with Rails

#### Rails → Channel Server (Redis Pub/Sub)

| Channel | Purpose |
|---------|---------|
| `/records` | Live record updates |
| `/system_notice` | System-wide notices |
| `chatbot/*` | Bot event publishing |

**Publishing from Rails:**
```ruby
CACHE_REDIS_POOL.publish('/records', {
  room: "user-#{user_id}",
  records: { discussions: [...], comments: [...] }
}.to_json)
```

#### Message Format

```json
{
  "room": "user-123",
  "records": {
    "comments": [...],
    "discussions": [...],
    "polls": [...],
    "stances": [...],
    "events": [...]
  }
}
```

### 8.4 WebSocket Authentication

#### Socket.io Authentication

1. Client boots and receives `channel_token` (= `user.secret_token`)
2. Socket connects with token in query params
3. Server validates token against Redis key `/current_users/{token}`
4. User joins rooms: `user-{id}`, `group-{id}` for each group

**Redis user cache structure:**
```json
{
  "name": "User Name",
  "group_ids": [1, 2, 3],
  "id": 123
}
```

#### Hocuspocus Authentication

1. Token format: `{user_id},{secret_token}`
2. Document name format: `{record_type}-{record_id}-{user_id_if_new}`
3. Rails `/api/hocuspocus` validates user and permissions

**Supported record types:**
- comment, discussion, poll, stance, outcome
- pollTemplate, discussionTemplate, group, user

### 8.5 Configuration

**Environment Variables:**

| Variable | Purpose |
|----------|---------|
| `REDIS_URL` | Redis connection |
| `PUBLIC_APP_URL` | CORS origin |
| `PRIVATE_APP_URL` | Rails API base |
| `PORT` | Socket.io port (5000) |
| `RAILS_ENV` | Hocuspocus port selection |

### 8.6 Key Files

| File | Purpose |
|------|---------|
| `server.js` | Main entry point |
| `sockets.js` | Socket.io configuration |
| `hocuspocus.js` | Collaborative editing |
| `matrix_bot.js` | Matrix integration |

---

## 9. Important Comments & Known Issues

**Source:** Various files across codebase

### 9.1 FIXME - Critical Issues

#### Guest Boolean Migration Incomplete

**File:** `orig/loomio/db/migrate/20240130011619_add_guest_boolean_to_discussion_readers_and_stances.rb` (line 8)

```ruby
# FIXME add/run migration to convert existing guest records to guest = true
```

**Impact:** Existing guest records have `guest = false` when they should be `true`. This affects:
- `discussion_readers` table
- `stances` table

### 9.2 TODO - Incomplete Features

#### Reaction Uniqueness Not Enforced

**File:** `orig/loomio/app/models/reaction.rb` (line 5)

```ruby
# TODO: ensure one reaction per reactable
# validates_uniqueness_of :user_id, scope: :reactable
```

**Impact:** Multiple reactions from same user on same item are allowed.

#### Task Deletion Notifications

**File:** `orig/loomio/app/services/task_service.rb` (line 91)

```ruby
# TODO maybe notify people if a task is deleted. or mark it as discarded
model.tasks.where(uid: removed_uids).destroy_all
```

**Impact:** Tasks are hard-deleted without notification or soft-delete option.

#### Controllers Marked for Removal

**Files:**
- `orig/loomio/app/controllers/poll_templates_controller.rb` (line 2)
- `orig/loomio/app/controllers/thread_templates_controller.rb` (line 2)

```ruby
# TODO remove this file
```

#### Email Permission Handling

**File:** `orig/loomio/app/services/received_email_service.rb` (line 115)

```ruby
# TODO handle when user is not allowed to comment or create discussion
rescue CanCan::AccessDenied, ActiveRecord::RecordNotFound
```

**Impact:** Email routing silently fails when permission denied.

#### Anonymous Poll Email Display

**File:** `orig/loomio/spec/controllers/poll_mailer_spec.rb` (lines 76-99)

```ruby
# TODO expect to see user, but not their position or reason
# TODO expect anonymous but see results
# TODO expect no results panel, just a summary of how many people have voted
```

**Impact:** Anonymous poll handling in emails may be incomplete.

#### Mark-as-read Logic

**File:** `orig/loomio/spec/controllers/email_actions_controller_spec.rb` (line 125)

```ruby
# TODO: this function works but we need to revise the test and/or the mark_as_read method itself,
# to be based on discussion and sequence ids instead of time
```

**Impact:** Read tracking uses timestamps instead of sequence IDs.

### 9.3 Important Implementation Notes

#### Comment Reparenting on Email Reply to Deleted Comment

**File:** `orig/loomio/app/models/comment.rb`

```ruby
# If someone replies to a deleted comment (in practice, by email), reparent to the discussion
self.parent = self.discussion if parent.nil? && discussion.present?
```

#### RecordCache Nil Pattern

**File:** `orig/loomio/app/services/record_cache.rb` (lines 14-20)

```ruby
# if we've already queried for a record and it does not exist, then we stil add a key into the hash, with nil
# so you can safely provide a query to check, without it being run redundandly.
# this is most important for discussion_readers
```

**Impact:** Cache uses nil to represent "queried but not found" - important for N+1 prevention.

#### SSO Two-Mode Operation

**File:** `orig/loomio/app/controllers/identities/base_controller.rb`

Two authentication modes:
1. **SSO-only mode:** `uid` is source of truth, no email login
2. **Standard mode:** Email login allowed, only link to verified users

#### Email Bounce Throttling

**File:** `orig/loomio/app/services/received_email_service.rb` (lines 33-42)

```ruby
# Bounce emails sent to notifications address are throttled (1 per hour)
if ThrottleService.can?(key: 'bounce', id: email.sender_email.downcase, max: 1, per: 'hour')
```

#### Email-to-Thread Routing

**File:** `orig/loomio/app/services/received_email_service.rb` (lines 62-76)

Two special formats:
- `d=100&k=key&u=999@mail.loomio.com` - Email to thread
- `group+u=99&k=key@mail.loomio.com` - Email to group

---

## 10. Key Patterns for Go Rewrite

### 10.1 Service Layer Pattern

**Location:** `orig/loomio/app/services/` (46+ files)

Each service follows this pattern:

```ruby
class DiscussionService
  def self.create(discussion:, actor:, params: {})
    # 1. Authorize
    actor.ability.authorize!(:create, discussion)

    # 2. Business logic
    discussion.assign_attributes(params)
    discussion.save!

    # 3. Publish event
    EventBus.broadcast('discussion_create', discussion, actor)

    discussion
  end
end
```

**Go equivalent:**

```go
type DiscussionService struct {
    repo *DiscussionRepository
    auth *Authorizer
    bus  *EventBus
}

func (s *DiscussionService) Create(ctx context.Context, discussion *Discussion, actor *User) error {
    if err := s.auth.Authorize(actor, "create", discussion); err != nil {
        return err
    }
    if err := s.repo.Create(ctx, discussion); err != nil {
        return err
    }
    s.bus.Publish("discussion_create", discussion, actor)
    return nil
}
```

### 10.2 Authorization (CanCan-style)

**Pattern:** Declarative ability rules per resource type

```ruby
# ability/discussion.rb
def initialize(user)
  can :show, Discussion do |discussion|
    discussion.public? || user.is_member_of?(discussion.group)
  end

  can :create, Discussion do |discussion|
    user.is_member_of?(discussion.group) &&
      discussion.group.members_can_start_discussions
  end
end
```

**Go equivalent using Casbin or custom:**

```go
type Ability struct {
    user *User
}

func (a *Ability) Can(action string, resource interface{}) bool {
    switch r := resource.(type) {
    case *Discussion:
        return a.canDiscussion(action, r)
    case *Poll:
        return a.canPoll(action, r)
    }
    return false
}

func (a *Ability) canDiscussion(action string, d *Discussion) bool {
    switch action {
    case "show":
        return !d.Private || a.user.IsMemberOf(d.GroupID)
    case "create":
        return a.user.IsMemberOf(d.GroupID) && d.Group.MembersCanStartDiscussions
    }
    return false
}
```

### 10.3 Event Publishing

**Pattern:** EventBus broadcasts events for async processing

```ruby
# lib/event_bus.rb
EventBus.broadcast('discussion_create', discussion, actor)

# Listeners registered elsewhere
EventBus.listen('discussion_create') do |discussion, actor|
  NotificationService.notify_members(discussion)
end
```

**Go equivalent:**

```go
type EventBus struct {
    listeners map[string][]func(interface{}, *User)
    mu        sync.RWMutex
}

func (b *EventBus) Publish(event string, payload interface{}, actor *User) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for _, fn := range b.listeners[event] {
        go fn(payload, actor)
    }
}
```

### 10.4 Counter Cache Management

**Pattern:** Denormalized counts kept in sync

```ruby
# Many models use counter caches
update_counter_cache :group, :memberships_count
update_counter_cache :discussion, :comments_count
```

**Go equivalent:**

```go
// Use database triggers or explicit updates
func (r *MembershipRepository) Create(ctx context.Context, m *Membership) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(m).Error; err != nil {
            return err
        }
        return tx.Model(&Group{}).
            Where("id = ?", m.GroupID).
            UpdateColumn("memberships_count", gorm.Expr("memberships_count + 1")).
            Error
    })
}
```

### 10.5 Soft Delete Pattern

**Pattern:** Uses `discard` gem with `discarded_at` timestamp

```ruby
include Discard::Model
default_scope -> { kept }  # Excludes discarded by default
```

**Go equivalent with GORM:**

```go
type Discussion struct {
    ID          int64
    DiscardedAt *time.Time `gorm:"index"`
    // ...
}

// Scopes
func Kept(db *gorm.DB) *gorm.DB {
    return db.Where("discarded_at IS NULL")
}

func (d *Discussion) Discard() {
    now := time.Now()
    d.DiscardedAt = &now
}
```

### 10.6 Polymorphic Associations

**Pattern:** Events, reactions, documents use polymorphic foreign keys

```ruby
belongs_to :eventable, polymorphic: true  # eventable_type, eventable_id
```

**Go equivalent:**

```go
type Event struct {
    ID            int64
    EventableType string  // "Discussion", "Poll", "Comment"
    EventableID   int64
}

func (e *Event) GetEventable(db *gorm.DB) (interface{}, error) {
    switch e.EventableType {
    case "Discussion":
        var d Discussion
        return &d, db.First(&d, e.EventableID).Error
    case "Poll":
        var p Poll
        return &p, db.First(&p, e.EventableID).Error
    }
    return nil, errors.New("unknown eventable type")
}
```

### 10.7 Query Objects

**Pattern:** Complex queries encapsulated in dedicated classes

```ruby
# app/queries/discussion_query.rb
class DiscussionQuery
  def self.visible_to(user:, group_ids:, tags:, since:)
    Discussion.where(group_id: group_ids)
              .where("private = false OR id IN (?)", reader_discussion_ids(user))
              .where("last_activity_at > ?", since)
  end
end
```

**Go equivalent:**

```go
type DiscussionQuery struct {
    db *gorm.DB
}

func (q *DiscussionQuery) VisibleTo(user *User, groupIDs []int64, since time.Time) *gorm.DB {
    readerIDs := q.readerDiscussionIDs(user)
    return q.db.Where("group_id IN ?", groupIDs).
        Where("private = false OR id IN ?", readerIDs).
        Where("last_activity_at > ?", since)
}
```

### 10.8 API Serialization

**Pattern:** ActiveModel::Serializer with conditional fields

```ruby
class DiscussionSerializer < ApplicationSerializer
  attributes :id, :title, :description

  def include_description?
    !object.discarded_at
  end

  has_one :author
  has_many :active_polls
end
```

**Go equivalent (custom or using libraries like jsonapi):**

```go
type DiscussionResponse struct {
    ID          int64              `json:"id"`
    Title       string             `json:"title"`
    Description *string            `json:"description,omitempty"`
    Author      *AuthorResponse    `json:"author,omitempty"`
    ActivePolls []*PollResponse    `json:"active_polls,omitempty"`
}

func (d *Discussion) ToResponse(includeDescription bool) *DiscussionResponse {
    resp := &DiscussionResponse{
        ID:    d.ID,
        Title: d.Title,
    }
    if includeDescription && d.DiscardedAt == nil {
        resp.Description = &d.Description
    }
    return resp
}
```

### 10.9 Real-time Updates

**Pattern:** Redis pub/sub for live updates

```ruby
CACHE_REDIS_POOL.publish('/records', {
  room: "user-#{user.id}",
  records: { discussions: [serialized_discussion] }
}.to_json)
```

**Go equivalent:**

```go
type LiveUpdater struct {
    redis *redis.Client
}

func (u *LiveUpdater) PublishToUser(userID int64, records map[string]interface{}) error {
    payload, _ := json.Marshal(map[string]interface{}{
        "room":    fmt.Sprintf("user-%d", userID),
        "records": records,
    })
    return u.redis.Publish(context.Background(), "/records", payload).Err()
}
```

### 10.10 Background Job Patterns

**Pattern:** Sidekiq workers with queue selection

```ruby
class SendDailyCatchUpEmailWorker
  include Sidekiq::Worker
  sidekiq_options queue: :mailers, retry: false

  def perform
    # Job logic
  end
end
```

**Go equivalent (using asynq, machinery, or similar):**

```go
type SendDailyCatchUpTask struct{}

func (t *SendDailyCatchUpTask) ProcessTask(ctx context.Context, task *asynq.Task) error {
    // Job logic
    return nil
}

// Registration
mux.HandleFunc("send_daily_catch_up", handler)
```

---

## Appendix A: File Reference Index

### Core Application Files

| File | Lines | Purpose |
|------|-------|---------|
| `config/routes.rb` | 471 | All routes |
| `db/schema.rb` | 1093 | Database schema |
| `app/models/user.rb` | 377 | User model |
| `app/models/group.rb` | 476 | Group model |
| `app/models/discussion.rb` | 287 | Discussion model |
| `app/models/poll.rb` | 546 | Poll model |
| `app/models/comment.rb` | 162 | Comment model |
| `app/models/stance.rb` | 321 | Stance model |
| `app/models/membership.rb` | 101 | Membership model |
| `app/models/event.rb` | 100 | Event model |
| `app/models/ability/base.rb` | 56 | Authorization base |

### Controller Files

| File | Lines | Purpose |
|------|-------|---------|
| `app/controllers/api/v1/snorlax_base.rb` | 281 | API base controller |
| `app/controllers/api/v1/restful_controller.rb` | 23 | RESTful patterns |
| `app/controllers/api/v1/discussions_controller.rb` | 187 | Discussions API |
| `app/controllers/api/v1/polls_controller.rb` | ~150 | Polls API |
| `app/controllers/api/v1/groups_controller.rb` | ~100 | Groups API |

### Serializer Files

| File | Lines | Purpose |
|------|-------|---------|
| `app/serializers/application_serializer.rb` | 180 | Base serializer |
| `app/serializers/discussion_serializer.rb` | 99 | Discussion API |
| `app/serializers/poll_serializer.rb` | 144 | Poll API |
| `app/serializers/group_serializer.rb` | 135 | Group API |
| `app/serializers/stance_serializer.rb` | 76 | Stance API |
| `app/serializers/event_serializer.rb` | 63 | Event API |

### Service Files

| File | Lines | Purpose |
|------|-------|---------|
| `app/services/discussion_service.rb` | ~300 | Discussion operations |
| `app/services/poll_service.rb` | ~400 | Poll operations |
| `app/services/membership_service.rb` | ~200 | Membership operations |
| `app/services/group_service.rb` | ~150 | Group operations |
| `app/services/event_service.rb` | ~100 | Event creation |

### Worker Files

| File | Lines | Purpose |
|------|-------|---------|
| `app/workers/deactivate_user_worker.rb` | 24 | User deactivation |
| `app/workers/migrate_user_worker.rb` | 85 | User merging |
| `app/workers/send_daily_catch_up_email_worker.rb` | 24 | Digest emails |
| `app/workers/generic_worker.rb` | 6 | Generic service calls |
| `lib/tasks/loomio.rake` | 276 | Scheduled tasks |

---

*Document generated: 2026-01-30*
*Loomio schema version: 2025_12_03_031449*
