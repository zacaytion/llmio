# Discussions Domain - Controllers

**Generated:** 2026-02-01
**Domain:** DiscussionsController, CommentsController
**Confidence:** 5/5 (High - Based on direct controller inspection)

---

## Overview

The discussions domain exposes two main API controllers:
- **Api::V1::DiscussionsController** - RESTful endpoints for discussion management
- **Api::V1::CommentsController** - RESTful endpoints for comment management

Both inherit from `Api::V1::RestfulController` which provides standard CRUD operations.

---

## DiscussionsController

**File:** `/app/controllers/api/v1/discussions_controller.rb`

### Standard RESTful Actions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/discussions` | List discussions |
| GET | `/api/v1/discussions/:key` | Show single discussion |
| POST | `/api/v1/discussions` | Create discussion |
| PATCH | `/api/v1/discussions/:key` | Update discussion |

### Custom Actions

#### index

Lists discussions with filtering support.

**Query Parameters:**
- `group_id` - Filter by group
- `subgroups` - Include subgroups ('all', 'mine', or none)
- `tags` - Filter by tags
- `filter` - State filter ('open', 'closed', 'all')
- `xids` - Specific discussion IDs (x-separated)

**Flow:**
1. Load and authorize group (optional)
2. Get accessible records via DiscussionQuery.visible_to
3. Apply filter via DiscussionQuery.filter
4. Paginate and serialize

#### show

Displays a single discussion.

**Special Handling:**
- If closed_at exists but closer_id is nil, attempts to find closer from events
- If created_event is missing, calls EventService.repair_discussion
- Calls accept_pending_membership to handle invitation acceptance

#### create

Creates a discussion, with optional forking support.

**Special Logic:**
If `forked_event_ids` is present and non-empty:
1. Create the discussion first
2. Call EventService.move_comments to move events to new discussion

This enables the "fork to new discussion" feature.

#### dashboard

Lists discussions for current user's dashboard.

**Requirements:** User must be logged in.

**Flow:**
1. Get discussions from user's groups and guest discussions
2. Filter to open discussions
3. Order by latest activity

#### direct

Lists direct (group-less) discussions.

**Flow:**
1. Query visible discussions with `only_direct: true`
2. Excludes public and subgroup discussions
3. Orders by latest activity

#### inbox

Lists unread/undismissed discussions.

**Requirements:** User must be logged in.

**Flow:**
1. Get discussions via DiscussionQuery.inbox
2. Filter to recent (6 weeks)
3. Order by latest activity

#### search

Searches discussions within a group.

**Requirements:** Group must be specified.

**Query Parameters:**
- `q` - Search query (required)

#### move

Moves discussion to another group.

**Flow:**
1. Load discussion
2. Call DiscussionService.move with new group_id
3. Return updated discussion

#### history

Returns read history for discussion.

**Response:** Array of objects with:
- `reader_id` - DiscussionReader ID
- `last_read_at` - When user last read
- `user_name` - User's name

**Restriction:** Returns 403 if discussion has anonymous polls.

#### mark_as_seen

Marks discussion as visited.

**Flow:** Calls DiscussionService.mark_as_seen

#### mark_as_read

Marks specific event ranges as read.

**Body Parameters:**
- `ranges` - Array of [start, end] pairs or string like "1-5,8-10"

**Flow:** Calls DiscussionService.mark_as_read

#### dismiss / recall

Dismisses or recalls discussion from inbox.

#### close / reopen

Closes or reopens discussion.

#### move_comments

Moves comments between discussions (forking destination side).

**Body Parameters:**
- `forked_event_ids` - Array of event IDs to move

#### pin / unpin

Pins or unpins discussion to top of group list.

#### set_volume

Updates notification volume for current user.

**Body Parameters:**
- `volume` - One of 'mute', 'quiet', 'normal', 'loud'

#### discard

Soft-deletes discussion.

---

## CommentsController

**File:** `/app/controllers/api/v1/comments_controller.rb`

### Standard RESTful Actions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/comments` | Create comment |
| PATCH | `/api/v1/comments/:id` | Update comment |
| DELETE | `/api/v1/comments/:id` | Destroy comment |

### Custom Actions

#### discard

Soft-deletes a comment.

**Flow:**
1. Load comment
2. Call CommentService.discard
3. Return event with reduced scope (excludes discussion, group, user)

#### undiscard

Restores a discarded comment.

**Flow:**
1. Load comment
2. Call CommentService.undiscard
3. Return event

#### destroy

Permanently deletes a comment.

**Flow:**
1. Load comment
2. Get parent event (for returning updated children)
3. Call destroy_action
4. Reload parent event
5. Return serialized children events

**Note:** Destroy is only allowed if comment has no replies.

---

## Query Object: DiscussionQuery

**File:** `/app/queries/discussion_query.rb`

The controller uses DiscussionQuery for visibility and filtering.

### start

Base query with:
- `kept` scope (excludes discarded)
- LEFT JOIN to groups (for archived check)
- Includes author and group

### dashboard(chain:, user:)

Filters to:
- Discussions in user's groups
- Direct discussions where user is a guest

### inbox(chain:, user:)

Filters to:
- Dashboard discussions that are:
  - Not dismissed (or dismissed before last activity)
  - Not fully read (or last_read before last activity)

### visible_to(chain:, user:, ...)

Applies visibility rules:
1. Public discussions (if `or_public: true`)
2. Discussions in user's groups
3. Discussions where user has guest access via DiscussionReader
4. Parent group visible discussions (if `or_subgroups: true`)

**Parameters:**
- `group_ids` - Limit to specific groups
- `discussion_ids` - Limit to specific discussions
- `tags` - Filter by tags (array contains)
- `or_public` - Include public discussions
- `or_subgroups` - Include subgroup discussions
- `only_direct` - Only group-less discussions
- `only_unread` - Only unread discussions

Also supports token-based access via `user.discussion_reader_token`.

### filter(chain:, filter:)

Applies state filter:
- `show_closed` or `closed` -> is_closed scope
- `all` -> no filter
- default -> is_open scope

Orders by pinned status then latest activity.

---

## Request/Response Patterns

### Discussion Creation

**Request:**
```
POST /api/v1/discussions
{
  "discussion": {
    "title": "...",
    "description": "...",
    "description_format": "html",
    "private": true,
    "group_id": 123,
    "recipient_user_ids": [1, 2, 3],
    "recipient_emails": ["a@b.com"],
    "recipient_audience": "group"
  }
}
```

**Response:**
```
{
  "events": [{...created_event...}],
  "discussions": [{...discussion...}],
  "groups": [{...group...}],
  "users": [{...author...}]
}
```

### Mark as Read

**Request:**
```
PATCH /api/v1/discussions/:key/mark_as_read
{
  "ranges": "1-5,8-10"
}
```

**Response:** 200 OK

### Forking Comments

**Request:**
```
POST /api/v1/discussions
{
  "discussion": {
    "title": "Forked Discussion",
    "group_id": 123,
    "forked_event_ids": [456, 457, 458]
  }
}
```

This creates a new discussion and moves the specified events to it.

---

## Authorization Flow

Controllers delegate to services, which perform authorization:

```
PSEUDO-CODE:
Controller.action
  -> load_resource (finds by key or id)
  -> Service.action(resource:, actor: current_user)
     -> actor.ability.authorize! :action, resource
     -> perform operation
     -> return event
  -> respond_with_resource
```

The `load_resource` method uses ModelLocator which supports both `id` and `key` lookups.

---

## Visibility Permissions (from Ability::Discussion)

| Action | Who Can Do It |
|--------|---------------|
| show | Anyone who passes DiscussionQuery.visible_to |
| create | Email-verified user who is group admin OR member with permission OR creating direct discussion |
| update | Author OR discussion admin OR member with members_can_edit_discussions |
| move | Author OR discussion admin OR member with members_can_edit_discussions |
| pin | Same as update |
| destroy/discard | Author OR discussion admin |
| announce | Group admin OR member with members_can_announce |
| add_members | Any discussion member |
| add_guests | Group admin OR member with members_can_add_guests (requires subscription) |

---

## Comment Permissions (from Ability::Comment)

| Action | Who Can Do It |
|--------|---------------|
| create | Discussion member AND discussion not closed |
| update | (Author with members_can_edit_comments) OR (Admin with admins_can_edit_user_content) |
| discard | (Author AND member) OR discussion admin |
| destroy | (Author with members_can_delete_comments OR admin) AND no child comments |
| show | Can show discussion AND comment is kept |
