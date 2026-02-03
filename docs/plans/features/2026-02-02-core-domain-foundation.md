# Core Domain Foundation: Next 3 Features

**Created**: 2026-02-02

## Summary

Build the foundational domain models for Loomio rewrite in dependency order:
1. **004-groups-memberships**: Organizational containers with hierarchy and permissions
2. **005-discussions**: Threaded conversations with comments
3. **006-events-timeline**: Activity log powering timelines

## Feature 004: Groups & Memberships

### Models

**Group**
| Column | Type | Notes |
|--------|------|-------|
| `id` | bigint | PK |
| `name` | varchar(255) | Required |
| `handle` | varchar(100) | Unique URL slug |
| `description` | text | Optional |
| `parent_id` | bigint | FK to groups (nullable) |
| `archived_at` | timestamptz | Soft delete |
| `created_at` | timestamptz | |

**Permission Flags** (all boolean, on Group):
- `members_can_add_members`, `members_can_add_guests`
- `members_can_start_discussions`, `members_can_raise_motions`
- `members_can_edit_discussions`, `members_can_edit_comments`, `members_can_delete_comments`
- `members_can_announce`, `members_can_create_subgroups`
- `admins_can_edit_user_content`, `parent_members_can_see_discussions`

**Membership**
| Column | Type | Notes |
|--------|------|-------|
| `id` | bigint | PK |
| `group_id` | bigint | FK to groups |
| `user_id` | bigint | FK to users |
| `admin` | boolean | Default false |
| `inviter_id` | bigint | FK to users (nullable) |
| `accepted_at` | timestamptz | Nullable until accepted |
| `created_at` | timestamptz | |

### Key Behaviors

- Creating a group auto-creates admin membership for creator
- `handle` must be unique, lowercase, URL-safe (alphanumeric + hyphens)
- Subgroups can inherit settings from parent (optional)
- Soft delete via `archived_at`

### API Endpoints

```
POST   /groups                      Create group
GET    /groups/:id                  Get group
PATCH  /groups/:id                  Update group
DELETE /groups/:id                  Archive group
GET    /groups/:id/subgroups        List subgroups

POST   /memberships                 Create/invite
GET    /memberships/:id             Get membership
PATCH  /memberships/:id             Update (accept invite)
DELETE /memberships/:id             Remove member
POST   /memberships/:id/make_admin  Promote to admin
POST   /memberships/:id/remove_admin Demote from admin
```

---

## Feature 005: Discussions

### Models

**Discussion**
| Column | Type | Notes |
|--------|------|-------|
| `id` | bigint | PK |
| `title` | varchar(255) | Required |
| `description` | text | Rich text body |
| `group_id` | bigint | FK (nullable = direct discussion) |
| `author_id` | bigint | FK to users |
| `private` | boolean | Default true |
| `closed_at` | timestamptz | Nullable |
| `max_depth` | integer | Default 3 |
| `created_at` | timestamptz | |

**Comment**
| Column | Type | Notes |
|--------|------|-------|
| `id` | bigint | PK |
| `discussion_id` | bigint | FK to discussions |
| `author_id` | bigint | FK to users |
| `parent_id` | bigint | FK to comments (nullable) |
| `body` | text | Rich text |
| `edited_at` | timestamptz | Nullable |
| `discarded_at` | timestamptz | Soft delete |
| `created_at` | timestamptz | |

**DiscussionReader**
| Column | Type | Notes |
|--------|------|-------|
| `id` | bigint | PK |
| `discussion_id` | bigint | FK |
| `user_id` | bigint | FK |
| `last_read_at` | timestamptz | |
| `volume` | varchar(20) | mute/normal/loud |

### Key Behaviors

- `group_id = NULL` → "direct discussion" (invited participants only)
- Permission checks via group's `members_can_*` flags
- Comments nest up to `max_depth`, then flatten
- Soft delete comments shows "[deleted]" placeholder

### API Endpoints

```
POST   /discussions                 Create discussion
GET    /discussions/:id             Get discussion
PATCH  /discussions/:id             Update discussion
DELETE /discussions/:id             Delete discussion
POST   /discussions/:id/close       Close discussion
POST   /discussions/:id/reopen      Reopen discussion

POST   /comments                    Create comment
PATCH  /comments/:id                Update comment
DELETE /comments/:id                Soft delete comment
```

---

## Feature 006: Events & Timeline

### Model

**Event**
| Column | Type | Notes |
|--------|------|-------|
| `id` | bigint | PK |
| `kind` | varchar(50) | Event type |
| `eventable_type` | varchar(100) | Polymorphic |
| `eventable_id` | bigint | Polymorphic |
| `user_id` | bigint | Actor (nullable for system events) |
| `discussion_id` | bigint | FK (nullable) |
| `sequence_id` | integer | Per-discussion sequence |
| `metadata` | jsonb | Extra context |
| `created_at` | timestamptz | |

### Initial Event Types (10)

| Kind | Eventable | Description |
|------|-----------|-------------|
| `discussion_created` | Discussion | New discussion |
| `discussion_edited` | Discussion | Title/description changed |
| `discussion_closed` | Discussion | Discussion closed |
| `discussion_reopened` | Discussion | Discussion reopened |
| `comment_created` | Comment | New comment |
| `comment_edited` | Comment | Comment edited |
| `comment_deleted` | Comment | Comment soft-deleted |
| `membership_created` | Membership | User joined/invited |
| `membership_removed` | Membership | User left/removed |
| `user_mentioned` | Comment | @mention in comment |

### Key Behaviors

- Events are **append-only** (immutable)
- `sequence_id` auto-increments per discussion (atomic via PostgreSQL sequence or advisory lock)
- Service layer creates events (not controllers)
- Foundation for notifications, webhooks, real-time (future features)

### API Endpoints

```
GET /discussions/:id/events         Paginated timeline
GET /events/:id                     Get single event
```

---

## Implementation Order

1. **004-groups-memberships** (no dependencies beyond users)
2. **005-discussions** (depends on groups for permissions)
3. **006-events-timeline** (depends on discussions, comments, memberships)

## Testing Strategy

Each feature follows constitution's TDD mandate:
- **pgTap**: Schema constraints, permission flag defaults, cascade behavior
- **Go tests**: Service layer logic, API endpoints, permission checks
- **Table-driven tests**: Cover permission matrix (admin/member/guest × each flag)

## Not In Scope

- Notifications delivery (future feature)
- Real-time broadcasting (future feature)
- Webhooks dispatch (future feature)
- Polls and voting (later domain feature)
