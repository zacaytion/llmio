# Domain Models

> Core entities, associations, and patterns.

## Core Entities

### User
**Source:** `orig/loomio/app/models/user.rb` (377 lines)

| Column | Type | Purpose |
|--------|------|---------|
| `email` | citext | Unique, case-insensitive |
| `username` | string | Unique handle |
| `secret_token` | UUID | WebSocket authentication |
| `deactivated_at` | timestamp | Soft deactivation |
| `experiences` | JSONB | Feature flags, preferences |

**Key Methods:** `is_member_of?(group)`, `is_admin_of?(group)`, `email_api_key`

### Group
**Source:** `orig/loomio/app/models/group.rb` (476 lines)

Hierarchical organizations with `parent_id` self-reference for subgroups.

| Column | Type | Purpose |
|--------|------|---------|
| `handle` | citext | Unique URL slug |
| `parent_id` | integer | Self-referential (subgroups) |
| `members_can_*` | boolean | 11 permission flags |

**Counter Caches (17):** `memberships_count`, `admin_memberships_count`, `pending_memberships_count`, `discussions_count`, `public_discussions_count`, `open_discussions_count`, `closed_discussions_count`, `polls_count`, `closed_polls_count`, `closed_motions_count`, `proposal_outcomes_count`, `subgroups_count`, `invitations_count`, `recent_activity_count`, `discussion_templates_count`, `poll_templates_count`, `delegates_count`

### Discussion
**Source:** `orig/loomio/app/models/discussion.rb` (287 lines)

| Column | Type | Purpose |
|--------|------|---------|
| `key` | string | Unique URL key |
| `description_format` | string | `'md'` or `'html'` |
| `private` | boolean | Visibility |
| `tags` | string[] | PostgreSQL array |
| `attachments` | JSONB | File references |

### Poll
**Source:** `orig/loomio/app/models/poll.rb` (546 lines)

**9 Poll Types** (from `config/poll_types.yml`):

| Type | Purpose |
|------|---------|
| `proposal` | Yes/No/Abstain consensus |
| `poll` | Multiple choice |
| `count` | Simple headcount |
| `score` | Numeric rating |
| `ranked_choice` | Preference ordering |
| `meeting` | Date/time scheduling |
| `dot_vote` | Budget allocation |
| `check` | Checkbox/attendance |
| `question` | Open-ended questions |

**Custom Fields:** `meeting_duration`, `time_zone`, `can_respond_maybe`

### Stance
**Source:** `orig/loomio/app/models/stance.rb` (321 lines)

Individual votes with `option_scores` JSONB storing `{poll_option_id: score}`.

**Key Pattern:** `latest` boolean with partial unique index ensures one active stance per (poll_id, participant_id).

**Anonymous Handling:** `participant` returns nil for anonymous polls; use `real_participant` to bypass.

### Comment
**Source:** `orig/loomio/app/models/comment.rb` (162 lines)

Threaded with polymorphic `parent` (Discussion, Comment, or Stance).

**Reparenting Logic:** If replying to deleted comment (via email), reparents to discussion.

### Event
**Source:** `orig/loomio/app/models/event.rb` (100 lines)

Activity timeline with polymorphic `eventable`.

**42 Event Kinds** (14 webhook-eligible):

<details>
<summary>Complete List</summary>

**Discussion:** new_discussion, discussion_edited, discussion_title_edited, discussion_description_edited, discussion_closed, discussion_reopened, discussion_forked, discussion_moved, discussion_announced

**Comment:** new_comment, comment_edited, comment_replied_to

**Poll:** poll_created, poll_edited, poll_closing_soon, poll_closed_by_user, poll_expired, poll_reopened, poll_option_added, poll_announced, poll_reminder

**Stance:** stance_created, stance_updated

**Outcome:** outcome_created, outcome_updated, outcome_announced, outcome_review_due

**Membership:** membership_created, membership_requested, membership_request_approved, membership_resent, invitation_accepted, user_added_to_group, user_joined_group, new_coordinator, new_delegate

**Other:** user_mentioned, group_mentioned, reaction_created, announcement_resend, user_reactivated, unknown_sender
</details>

**Webhook-Eligible (14):** Defined in `config/webhook_event_kinds.yml`

**Position Key Format:**
```
Format: "{parent_position_key}-{zero_padded_position}"
Example: "00001-00002-00003"
```

Zero-padded to 5 digits (`Event.zero_fill`) enabling string sorting while maintaining tree hierarchy. Built recursively from parent's position_key + current position.

**Source:** `orig/loomio/app/models/event.rb:121-122`

## Model Patterns

### Soft Delete (Discard)
```ruby
include Discard::Model
default_scope { kept }  # WHERE discarded_at IS NULL
```

**Tables with `discarded_at`:** discussions, comments, polls, tasks, discussion_templates, poll_templates

**Tables with `archived_at`:** groups (archive, not delete)

**Tables with `deactivated_at`:** users

### Access Revocation
```ruby
scope :active, -> { where(revoked_at: nil) }
```

**Tables:** memberships, discussion_readers, stances

### Volume Levels

| Level | Value | Behavior |
|-------|-------|----------|
| mute | 0 | No notifications |
| quiet | 1 | App notifications only (no email) |
| normal | 2 | Both email and app |
| loud | 3 | Maximum engagement + extras |

**Source:** `orig/loomio/app/models/concerns/has_volume.rb`

### Rich Text Pattern
```ruby
# Paired columns for content and format
description / description_format  # 'md' or 'html'
body / body_format
details / details_format
```

### Custom Fields Pattern
**Source:** `orig/loomio/app/models/concerns/has_custom_fields.rb`

```ruby
set_custom_fields :meeting_duration, :time_zone, :can_respond_maybe
# Creates getter/setters accessing custom_fields JSONB
```

**Usage by Model:**
- Poll: `meeting_duration`, `time_zone`, `can_respond_maybe`
- Event: `pinned_title`, `recipient_user_ids`, `recipient_chatbot_ids`, `recipient_message`, `recipient_audience`, `stance_ids`
- Outcome: `event_summary`, `event_description`, `event_location`

### Polymorphic Associations

| Model | Type Column | ID Column | Types |
|-------|-------------|-----------|-------|
| Event | eventable_type | eventable_id | Discussion, Comment, Poll, Stance, Outcome, Membership |
| Comment | parent_type | parent_id | Discussion, Comment, Stance |
| Reaction | reactable_type | reactable_id | Discussion, Comment, Poll, Stance, Outcome |
| Document | model_type | model_id | Discussion, Group, Poll, Comment |

### Experiences JSONB
User preferences and feature flags stored in `users.experiences` and `memberships.experiences`.

**Known Keys:**
- `html-editor.uses-markdown` - Editor format preference
- `betaFeatures` - Beta feature access

**Note:** Removed from groups in migration `20230615234611`.

## Key Relationships

```
User ──┬── Membership ──┬── Group
       │ admin, delegate│ parent_id (self)
       │ volume         │
       │ revoked_at     │
       │                │
       ├── Discussion ──┼── Comment (threaded)
       │   author_id    │   parent_type: Discussion|Comment|Stance
       │                │
       ├── Poll ────────┼── PollOption
       │   poll_type    │   └── voter_scores JSONB
       │                │
       └── Stance ──────┴── StanceChoice
           option_scores    └── poll_option_id, score
           latest (bool)
```

---
