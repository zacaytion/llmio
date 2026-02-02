# Loomio Database Schema Reference

**Generated:** 2026-02-01
**Source:** `db/schema.rb` (version 2025_12_03_031449)
**Database:** PostgreSQL with extensions: citext, hstore, pg_stat_statements, pgcrypto, plpgsql

---

## Table of Contents

1. [Core Domain Models](#core-domain-models)
2. [User & Authentication](#user--authentication)
3. [Groups & Memberships](#groups--memberships)
4. [Discussions & Comments](#discussions--comments)
5. [Polls & Voting](#polls--voting)
6. [Events & Notifications](#events--notifications)
7. [Files & Documents](#files--documents)
8. [Search & Indexing](#search--indexing)
9. [External Integrations](#external-integrations)
10. [System & Infrastructure](#system--infrastructure)

---

## Core Domain Models

### Model Relationship Overview

```
User ─┬── Membership ── Group
      │       │
      │       └── Subgroups (parent_id)
      │
      ├── Discussion ── DiscussionReader
      │       │
      │       ├── Comment (threaded via parent_id)
      │       │
      │       └── Poll ── Stance ── StanceChoice
      │               │       │
      │               │       └── PollOption
      │               │
      │               └── Outcome
      │
      └── Event ── Notification
```

---

## User & Authentication

### users

Primary user accounts for the application.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `email` | citext | - | YES | Unique, case-insensitive email |
| `encrypted_password` | string(128) | "" | YES | Devise password hash |
| `reset_password_token` | string | - | YES | Password reset token (unique) |
| `reset_password_sent_at` | datetime | - | YES | When reset token was sent |
| `remember_created_at` | datetime | - | YES | Remember me timestamp |
| `sign_in_count` | integer | 0 | NO | Total sign-in count |
| `current_sign_in_at` | datetime | - | YES | Current session start |
| `last_sign_in_at` | datetime | - | YES | Previous session start |
| `current_sign_in_ip` | inet | - | YES | Current IP address |
| `last_sign_in_ip` | inet | - | YES | Previous IP address |
| `created_at` | datetime | - | YES | Account creation time |
| `updated_at` | datetime | - | YES | Last update time |
| `name` | string(255) | - | YES | Display name |
| `username` | string(255) | - | YES | Unique username |
| `deactivated_at` | datetime | - | YES | Soft delete timestamp |
| `deactivator_id` | integer | - | YES | User who deactivated |
| `is_admin` | boolean | false | YES | System administrator flag |
| `email_verified` | boolean | false | NO | Email verification status |
| `avatar_kind` | string(255) | "initials" | NO | Avatar type: initials, uploaded, gravatar |
| `uploaded_avatar_file_name` | string(255) | - | YES | Paperclip filename |
| `uploaded_avatar_content_type` | string(255) | - | YES | Paperclip content type |
| `uploaded_avatar_file_size` | integer | - | YES | Paperclip file size |
| `uploaded_avatar_updated_at` | datetime | - | YES | Paperclip update time |
| `avatar_initials` | string(255) | - | YES | Computed initials |
| `key` | string(255) | - | YES | Public URL key (unique) |
| `unsubscribe_token` | string(255) | - | YES | Email unsubscribe token (unique) |
| `email_api_key` | string(255) | - | YES | Reply-by-email authentication |
| `api_key` | string | - | YES | Bot API authentication |
| `authentication_token` | string(255) | - | YES | Legacy auth token |
| `secret_token` | string | gen_random_uuid() | NO | Internal secret |
| `remember_token` | string | - | YES | Session persistence |
| `selected_locale` | string(255) | - | YES | User's preferred locale |
| `detected_locale` | string(255) | - | YES | Auto-detected locale |
| `content_locale` | string | - | YES | Content creation locale |
| `time_zone` | string(255) | - | YES | User timezone |
| `autodetect_time_zone` | boolean | true | NO | Auto-detect timezone |
| `date_time_pref` | string | - | YES | Date/time format preference |
| `country` | string | - | YES | GeoIP country |
| `region` | string | - | YES | GeoIP region |
| `city` | string | - | YES | GeoIP city |
| `short_bio` | string | "" | NO | User biography |
| `short_bio_format` | string(10) | "md" | NO | Biography format |
| `location` | string | "" | NO | User-entered location |
| `email_catch_up` | boolean | true | NO | Receive digest emails |
| `email_catch_up_day` | integer | - | YES | Digest frequency (0-8) |
| `email_when_mentioned` | boolean | true | NO | Email on @mention |
| `email_on_participation` | boolean | false | NO | Email on participation |
| `email_when_proposal_closing_soon` | boolean | false | NO | Email on poll closing |
| `email_newsletter` | boolean | false | NO | Marketing emails |
| `default_membership_volume` | integer | 2 | NO | Default notification volume |
| `memberships_count` | integer | 0 | NO | Counter cache |
| `experiences` | jsonb | {} | NO | Feature flags/tutorials seen |
| `attachments` | jsonb | [] | NO | Profile attachments |
| `link_previews` | jsonb | [] | NO | Cached link previews |
| `last_seen_at` | datetime | - | YES | Last activity timestamp |
| `legal_accepted_at` | datetime | - | YES | Terms acceptance time |
| `failed_attempts` | integer | 0 | NO | Devise lockable counter |
| `unlock_token` | string | - | YES | Devise unlock token (unique) |
| `locked_at` | datetime | - | YES | Account lock timestamp |
| `complaints_count` | integer | 0 | NO | Spam complaint counter |
| `bot` | boolean | false | NO | Bot account flag |
| `auto_translate` | boolean | false | NO | Auto-translate content |
| `email_sha256` | string | - | YES | Hashed email for matching |
| `facebook_community_id` | integer | - | YES | Legacy FB integration |
| `slack_community_id` | integer | - | YES | Legacy Slack integration |

**Indexes:**
- `email` (unique)
- `username` (unique)
- `key` (unique)
- `reset_password_token` (unique)
- `unlock_token` (unique)
- `unsubscribe_token` (unique)
- `api_key`
- `email_verified`
- `remember_token`

**JSONB Structures:**

`experiences`:
```json
{
  "welcomeModal": true,
  "announcementHelpCard": true,
  "pollTypes": ["proposal", "count"]
}
```

---

### omniauth_identities

OAuth/SSO provider identities linked to users.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `user_id` | integer | - | YES | FK to users |
| `email` | string(255) | - | YES | Provider email |
| `name` | string(255) | - | YES | Provider name |
| `uid` | string(255) | - | YES | Provider user ID |
| `identity_type` | string(255) | - | YES | Provider name (google, oauth, saml, nextcloud) |
| `access_token` | string | "" | YES | OAuth access token (stored but unused) |
| `logo` | string | - | YES | Provider logo URL |
| `custom_fields` | jsonb | {} | NO | Provider-specific data |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

**Indexes:**
- `(identity_type, uid)`
- `user_id`
- `email`

---

### login_tokens

One-time login tokens for passwordless authentication.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `user_id` | integer | - | YES | FK to users |
| `token` | string | - | YES | URL token |
| `code` | integer | - | NO | Numeric code for verification |
| `used` | boolean | false | NO | Already consumed |
| `redirect` | string | - | YES | Post-login redirect URL |
| `is_reactivation` | boolean | false | NO | For reactivating accounts |
| `created_at` | datetime | - | YES | Token creation time |
| `updated_at` | datetime | - | YES | |

**Note:** Tokens expire after 1 hour (cleaned via hourly rake task).

---

## Groups & Memberships

### groups

Organization containers with hierarchical structure.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `name` | string(255) | - | YES | Group name |
| `full_name` | string(255) | - | YES | Computed: "Parent / Child" |
| `description` | text | - | YES | Rich text description |
| `description_format` | string(10) | "md" | NO | "md" or "html" |
| `handle` | citext | - | YES | URL slug (unique) |
| `key` | string(255) | - | YES | Public URL key (unique) |
| `token` | string | - | YES | Secret API token (unique) |
| `parent_id` | integer | - | YES | FK to groups (self-referential) |
| `creator_id` | integer | - | YES | FK to users |
| `subscription_id` | integer | - | YES | FK to subscriptions |
| `archived_at` | datetime | - | YES | Soft delete timestamp |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Permission Flags:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `members_can_add_members` | boolean | false | Members invite new members |
| `members_can_add_guests` | boolean | true | Members invite guests |
| `members_can_edit_discussions` | boolean | true | Members edit any discussion |
| `members_can_edit_comments` | boolean | true | Members edit own comments |
| `members_can_delete_comments` | boolean | true | Members delete own comments |
| `members_can_raise_motions` | boolean | true | Members create polls |
| `members_can_start_discussions` | boolean | true | Members create discussions |
| `members_can_create_subgroups` | boolean | false | Members create child groups |
| `members_can_announce` | boolean | true | Members send notifications |
| `members_can_vote` | boolean | true | **DEPRECATED** - unused |
| `admins_can_edit_user_content` | boolean | true | Admins edit others' content |
| `parent_members_can_see_discussions` | boolean | false | Parent group visibility |

**Visibility Settings:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `is_visible_to_public` | boolean | true | Listed in explore |
| `is_visible_to_parent_members` | boolean | false | Visible to parent |
| `discussion_privacy_options` | string | "private_only" | "public_only", "private_only", "public_or_private" |
| `membership_granted_upon` | string | "approval" | "approval", "request", "invitation" |
| `listed_in_explore` | boolean | false | Show in public explore |

**Thread Defaults:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `new_threads_max_depth` | integer | 3 | Default reply nesting depth |
| `new_threads_newest_first` | boolean | false | Default sort order |
| `can_start_polls_without_discussion` | boolean | false | Standalone polls allowed |

**Counter Caches:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `memberships_count` | integer | 0 | Total memberships |
| `admin_memberships_count` | integer | 0 | Admin count |
| `pending_memberships_count` | integer | 0 | Pending invitations |
| `delegates_count` | integer | 0 | Delegate count |
| `discussions_count` | integer | 0 | Total discussions |
| `open_discussions_count` | integer | 0 | Open discussions |
| `closed_discussions_count` | integer | 0 | Closed discussions |
| `public_discussions_count` | integer | 0 | Public discussions |
| `polls_count` | integer | 0 | Total polls |
| `closed_polls_count` | integer | 0 | Closed polls |
| `closed_motions_count` | integer | 0 | Legacy counter |
| `proposal_outcomes_count` | integer | 0 | Outcomes count |
| `invitations_count` | integer | 0 | Sent invitations |
| `subgroups_count` | integer | 0 | Child groups |
| `recent_activity_count` | integer | 0 | Activity metric |
| `discussion_templates_count` | integer | 0 | Templates |
| `poll_templates_count` | integer | 0 | Poll templates |

**Other Fields:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `cover_photo_*` | - | - | Paperclip cover photo |
| `logo_*` | - | - | Paperclip logo |
| `country`, `region`, `city` | string | - | GeoIP location |
| `category` | string | - | Group category |
| `category_id` | integer | - | Legacy category FK |
| `theme_id` | integer | - | Legacy theme FK |
| `cohort_id` | integer | - | Analytics cohort |
| `default_group_cover_id` | integer | - | Default cover FK |
| `admin_tags` | string | - | Internal admin tags |
| `content_locale` | string | - | Content locale |
| `is_referral` | boolean | false | Referral tracking |
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |
| `info` | jsonb | {} | Extensible metadata |
| `request_to_join_prompt` | string | - | Join request prompt |

**Indexes:**
- `handle` (unique)
- `key` (unique)
- `token` (unique)
- `parent_id`
- `subscription_id`
- `name`
- `full_name`
- `created_at`
- `archived_at` (partial: where IS NULL)

---

### memberships

User membership in groups with roles and settings.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups |
| `user_id` | integer | - | YES | FK to users |
| `inviter_id` | integer | - | YES | FK to users (who invited) |
| `revoker_id` | integer | - | YES | FK to users (who revoked) |
| `invitation_id` | integer | - | YES | FK to legacy invitations |
| `admin` | boolean | false | NO | Admin role |
| `delegate` | boolean | false | NO | Delegate role |
| `volume` | integer | - | YES | Notification volume (0-3) |
| `inbox_position` | integer | 0 | YES | Dashboard ordering |
| `title` | string | - | YES | Custom member title |
| `token` | string | - | YES | Invitation token (unique) |
| `experiences` | jsonb | {} | NO | Feature tutorials seen |
| `accepted_at` | datetime | - | YES | Invitation accepted |
| `revoked_at` | datetime | - | YES | Membership revoked |
| `saml_session_expires_at` | datetime | - | YES | SAML session timeout |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `(group_id, user_id)` (unique)
- `token` (unique)
- `inviter_id`
- `(user_id, volume)`
- `volume`
- `created_at`

**Volume Values:**
- 0: Mute
- 1: Quiet (no emails)
- 2: Normal (default)
- 3: Loud (all emails)

---

### membership_requests

Pending requests to join groups.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups |
| `requestor_id` | integer | - | YES | FK to users (requester) |
| `responder_id` | integer | - | YES | FK to users (admin) |
| `name` | string(255) | - | YES | Requester name |
| `email` | string(255) | - | YES | Requester email |
| `introduction` | text | - | YES | Request message |
| `response` | string(255) | - | YES | "approved" or "ignored" |
| `responded_at` | datetime | - | YES | Response timestamp |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

---

## Discussions & Comments

### discussions

Threaded conversation containers.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups (null = direct discussion) |
| `author_id` | integer | - | YES | FK to users |
| `closer_id` | integer | - | YES | FK to users (who closed) |
| `title` | string(255) | - | YES | Discussion title |
| `description` | text | - | YES | Rich text body |
| `description_format` | string(10) | "md" | NO | "md" or "html" |
| `key` | string(255) | - | YES | Public URL key (unique) |
| `private` | boolean | true | NO | Visibility flag |
| `template` | boolean | false | NO | Is a template |
| `discussion_template_id` | integer | - | YES | FK to templates |
| `discussion_template_key` | string | - | YES | Template key |
| `closed_at` | datetime | - | YES | When closed |
| `discarded_at` | datetime | - | YES | Soft delete |
| `discarded_by` | integer | - | YES | FK to users |
| `pinned_at` | datetime | - | YES | Pinned timestamp |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Thread Settings:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `max_depth` | integer | 2 | Reply nesting depth |
| `newest_first` | boolean | false | Sort order |

**Activity Tracking:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_activity_at` | datetime | - | Last event time |
| `last_comment_at` | datetime | - | Last comment time |
| `first_sequence_id` | integer | 0 | First event sequence |
| `last_sequence_id` | integer | 0 | Latest event sequence |
| `items_count` | integer | 0 | Event count |
| `seen_by_count` | integer | 0 | Unique viewers |
| `members_count` | integer | - | Participant count |

**Poll Counters:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `closed_polls_count` | integer | 0 | Closed polls |
| `anonymous_polls_count` | integer | 0 | Anonymous polls |

**Other:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `importance` | integer | 0 | Priority/importance |
| `versions_count` | integer | 0 | Paper trail versions |
| `content_locale` | string | - | Content locale |
| `iframe_src` | string(255) | - | Embedded iframe |
| `ranges_string` | string | - | Read ranges encoding |
| `tags` | string[] | [] | Tag array |
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |
| `info` | jsonb | {} | Extensible metadata |

**Indexes:**
- `key` (unique)
- `group_id`
- `author_id`
- `last_activity_at` (desc)
- `created_at`
- `private`
- `tags` (GIN)
- `discarded_at` (partial: where IS NULL)
- `template` (partial: where IS TRUE)

---

### discussion_readers

Per-user read state and participation for discussions.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `discussion_id` | integer | - | NO | FK to discussions |
| `user_id` | integer | - | NO | FK to users |
| `inviter_id` | integer | - | YES | FK to users (who invited) |
| `revoker_id` | integer | - | YES | FK to users (who revoked) |
| `last_read_at` | datetime | - | YES | Last read time |
| `last_read_sequence_id` | integer | 0 | NO | Last read event |
| `read_ranges_string` | string | - | YES | Compact read state |
| `volume` | integer | 2 | NO | Notification volume |
| `participating` | boolean | false | NO | Active participant |
| `admin` | boolean | false | NO | Discussion admin |
| `guest` | boolean | false | NO | Guest (not group member) |
| `token` | string | - | YES | Invitation token (unique) |
| `accepted_at` | datetime | - | YES | Invitation accepted |
| `revoked_at` | datetime | - | YES | Access revoked |
| `dismissed_at` | datetime | - | YES | Dismissed from inbox |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `(user_id, discussion_id)` (unique)
- `token` (unique)
- `discussion_id`
- `guest` (partial: where = true)
- `inviter_id` (partial: where NOT NULL)

---

### comments

Threaded comments within discussions.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `discussion_id` | integer | 0 | YES | FK to discussions |
| `user_id` | integer | 0 | YES | FK to users (author) |
| `parent_id` | integer | - | YES | FK to comments (reply-to) |
| `parent_type` | string | - | NO | Polymorphic parent type |
| `body` | text | "" | YES | Rich text content |
| `body_format` | string(10) | "md" | NO | "md" or "html" |
| `edited_at` | datetime | - | YES | Last edit time |
| `discarded_at` | datetime | - | YES | Soft delete |
| `discarded_by` | integer | - | YES | FK to users |
| `content_locale` | string | - | YES | Content locale |
| `versions_count` | integer | 0 | YES | Paper trail versions |
| `comment_votes_count` | integer | 0 | NO | Legacy likes count |
| `attachments_count` | integer | 0 | NO | Legacy attachments count |
| `attachments` | jsonb | [] | NO | Rich text attachments |
| `link_previews` | jsonb | [] | NO | Cached link previews |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `discussion_id`
- `(parent_type, parent_id)`

---

### reactions

Emoji reactions on comments and other records.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `reactable_id` | integer | - | YES | Polymorphic FK |
| `reactable_type` | string | "Comment" | NO | Polymorphic type |
| `user_id` | integer | - | YES | FK to users |
| `reaction` | string | "+1" | NO | Emoji name |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `(reactable_id, reactable_type)`
- `user_id`
- `created_at`

---

## Polls & Voting

### polls

Decision-making tools with multiple poll types.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `author_id` | integer | - | NO | FK to users |
| `discussion_id` | integer | - | YES | FK to discussions |
| `group_id` | integer | - | YES | FK to groups |
| `title` | string | - | NO | Poll title |
| `details` | text | - | YES | Rich text description |
| `details_format` | string(10) | "md" | NO | "md" or "html" |
| `key` | string | - | NO | Public URL key (unique) |
| `poll_type` | string | - | NO | Type: proposal, poll, count, etc. |
| `closing_at` | datetime | - | YES | Scheduled close time |
| `closed_at` | datetime | - | YES | Actual close time |
| `discarded_at` | datetime | - | YES | Soft delete |
| `discarded_by` | integer | - | YES | FK to users |
| `template` | boolean | false | NO | Is a template |
| `poll_template_id` | integer | - | YES | FK to templates |
| `poll_template_key` | string | - | YES | Template key |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Voting Configuration:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `anonymous` | boolean | false | Hide voter identities |
| `hide_results` | integer | 0 | 0=show, 1=until_vote, 2=until_closed |
| `voter_can_add_options` | boolean | false | Voters add options |
| `specified_voters_only` | boolean | false | Limit to invited voters |
| `shuffle_options` | boolean | false | Randomize option order |
| `multiple_choice` | boolean | false | Multiple selections allowed |
| `show_none_of_the_above` | boolean | false | NOTA option |

**Scoring Configuration:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `min_score` | integer | - | Minimum score value |
| `max_score` | integer | - | Maximum score value |
| `minimum_stance_choices` | integer | - | Min selections required |
| `maximum_stance_choices` | integer | - | Max selections allowed |
| `dots_per_person` | integer | - | Dot voting allocation |
| `agree_target` | integer | - | Target agree percentage |
| `quorum_pct` | integer | - | Quorum percentage |

**Notification Settings:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `notify_on_closing_soon` | integer | 0 | Hours before closing to notify |
| `stance_reason_required` | integer | 1 | 0=hidden, 1=optional, 2=required |
| `limit_reason_length` | boolean | true | Cap reason length |

**Aggregated Data:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `voters_count` | integer | 0 | Total voters |
| `undecided_voters_count` | integer | 0 | Invited but not voted |
| `none_of_the_above_count` | integer | 0 | NOTA votes |
| `stance_counts` | jsonb | [] | Per-option vote counts |
| `stance_data` | jsonb | {} | Aggregated stance data |
| `matrix_counts` | jsonb | [] | Matrix poll data |

**Display Settings:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `chart_type` | string | - | Visualization type |
| `poll_option_name_format` | string | - | Option display format |
| `reason_prompt` | string | - | Custom reason prompt |
| `process_name` | string | - | Process type name |
| `process_subtitle` | string | - | Process subtitle |
| `process_url` | string | - | Process documentation URL |

**Other:**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `versions_count` | integer | 0 | Paper trail versions |
| `content_locale` | string | - | Content locale |
| `default_duration_in_days` | integer | - | Template default |
| `tags` | string[] | [] | Tag array |
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |
| `custom_fields` | jsonb | {} | Extensible metadata |

**Indexes:**
- `key` (unique)
- `author_id`
- `discussion_id`
- `group_id`
- `(closed_at, closing_at)`
- `(closed_at, discussion_id)`
- `tags` (GIN)

---

### poll_options

Options available for selection in polls.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `poll_id` | integer | - | YES | FK to polls |
| `name` | string | - | NO | Option name/value |
| `priority` | integer | 0 | NO | Display order |
| `icon` | string | - | YES | Emoji icon |
| `meaning` | string | - | YES | Semantic meaning |
| `prompt` | string | - | YES | Selection prompt |
| `voter_count` | integer | 0 | NO | Voters selecting this |
| `total_score` | integer | 0 | NO | Sum of all scores |
| `score_counts` | jsonb | {} | NO | Score distribution |
| `voter_scores` | jsonb | {} | NO | Per-voter scores |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Test Conditions (for proposal outcomes):**

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `test_operator` | string | - | Comparison operator |
| `test_percent` | integer | - | Target percentage |
| `test_against` | string | - | Comparison target |

---

### stances

Individual votes/responses to polls.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `poll_id` | integer | - | NO | FK to polls |
| `participant_id` | integer | - | YES | FK to users |
| `inviter_id` | integer | - | YES | FK to users (who invited) |
| `revoker_id` | integer | - | YES | FK to users (who revoked) |
| `reason` | string | - | YES | Vote explanation |
| `reason_format` | string(10) | "md" | NO | "md" or "html" |
| `latest` | boolean | true | NO | Most recent stance |
| `cast_at` | datetime | - | YES | When vote was submitted |
| `revoked_at` | datetime | - | YES | If unvoted |
| `accepted_at` | datetime | - | YES | Invitation accepted |
| `admin` | boolean | false | NO | Poll admin |
| `guest` | boolean | false | NO | Guest voter |
| `volume` | integer | 2 | NO | Notification volume |
| `token` | string | - | YES | Invitation token (unique) |
| `none_of_the_above` | boolean | false | NO | NOTA selection |
| `option_scores` | jsonb | {} | NO | Option -> score mapping |
| `versions_count` | integer | 0 | YES | Paper trail versions |
| `content_locale` | string | - | YES | Content locale |
| `attachments` | jsonb | [] | NO | Rich text attachments |
| `link_previews` | jsonb | [] | NO | Cached link previews |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `poll_id`
- `participant_id`
- `token` (unique)
- `(poll_id, participant_id, latest)` (unique, partial: where latest = true)
- `(poll_id, cast_at)` (NULLS FIRST)
- `guest` (partial: where = true)

**Note:** Stance revision creates new record only if >15 min elapsed AND choices changed AND poll is in discussion.

---

### stance_choices

Individual option selections within a stance.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `stance_id` | integer | - | YES | FK to stances |
| `poll_option_id` | integer | - | YES | FK to poll_options |
| `score` | integer | 1 | NO | Score/weight for this choice |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

---

### outcomes

Published results/decisions from polls.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `poll_id` | integer | - | YES | FK to polls |
| `poll_option_id` | integer | - | YES | FK to poll_options (winning option) |
| `author_id` | integer | - | NO | FK to users |
| `statement` | text | - | NO | Outcome statement |
| `statement_format` | string(10) | "md" | NO | "md" or "html" |
| `latest` | boolean | true | NO | Most recent outcome |
| `review_on` | date | - | YES | Review reminder date |
| `versions_count` | integer | 0 | NO | Paper trail versions |
| `content_locale` | string | - | YES | Content locale |
| `attachments` | jsonb | [] | NO | Rich text attachments |
| `link_previews` | jsonb | [] | NO | Cached link previews |
| `custom_fields` | jsonb | {} | NO | Extensible metadata |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

---

## Events & Notifications

### events

Activity log entries that drive notifications and timelines.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `kind` | string(255) | - | YES | Event type (42 types) |
| `eventable_id` | integer | - | YES | Polymorphic FK |
| `eventable_type` | string(255) | - | YES | Polymorphic type |
| `eventable_version_id` | integer | - | YES | Paper trail version |
| `user_id` | integer | - | YES | FK to users (actor) |
| `discussion_id` | integer | - | YES | FK to discussions |
| `parent_id` | integer | - | YES | FK to events (parent event) |
| `sequence_id` | integer | - | YES | Sequence within discussion |
| `position` | integer | 0 | NO | Position in thread |
| `position_key` | string | - | YES | Hierarchical position key |
| `depth` | integer | 0 | NO | Nesting depth |
| `child_count` | integer | 0 | NO | Direct child count |
| `descendant_count` | integer | 0 | NO | Total descendant count |
| `announcement` | boolean | false | NO | Was announced |
| `pinned` | boolean | false | NO | Pinned to top |
| `custom_fields` | jsonb | {} | NO | Event-specific data |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `(discussion_id, sequence_id)` (unique)
- `(eventable_type, eventable_id)`
- `(eventable_id, kind)`
- `(parent_id, discussion_id)` (partial: where discussion_id NOT NULL)
- `parent_id`
- `position_key`
- `user_id`
- `created_at`

**Event Types (42):**
See `discovery/final/realtime_pubsub.md` for complete list with triggers.

---

### notifications

User notifications generated from events.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `user_id` | integer | - | YES | FK to users (recipient) |
| `event_id` | integer | - | YES | FK to events |
| `actor_id` | integer | - | YES | FK to users (who triggered) |
| `url` | string | - | YES | Deep link URL |
| `viewed` | boolean | false | NO | Read status |
| `translation_values` | jsonb | {} | NO | I18n interpolation |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `event_id`
- `user_id`
- `(user_id, id)`
- `id` (desc)

---

## Files & Documents

### documents

Attached files and links.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `model_id` | integer | - | YES | Polymorphic FK |
| `model_type` | string | - | YES | Polymorphic type |
| `group_id` | integer | - | YES | FK to groups |
| `author_id` | integer | - | NO | FK to users |
| `title` | string | - | YES | Display title |
| `url` | string | - | YES | File URL |
| `web_url` | string | - | YES | Public web URL |
| `thumb_url` | string | - | YES | Thumbnail URL |
| `doctype` | string | - | NO | Document type |
| `icon` | string | - | YES | Icon name |
| `color` | string | - | NO | Icon color |
| `file_file_name` | string | - | YES | Legacy filename |
| `file_content_type` | string | - | YES | Legacy content type |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

---

### active_storage_attachments / active_storage_blobs

Rails ActiveStorage tables for file management.

(Standard Rails schema - see Rails documentation)

---

## Search & Indexing

### pg_search_documents

Full-text search index table.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | bigserial | auto | NO | Primary key |
| `content` | text | - | YES | Plain text content |
| `ts_content` | tsvector | - | YES | Search vector |
| `searchable_type` | string | - | YES | Polymorphic type |
| `searchable_id` | bigint | - | YES | Polymorphic FK |
| `author_id` | bigint | - | YES | FK to users |
| `group_id` | bigint | - | YES | FK to groups |
| `discussion_id` | bigint | - | YES | FK to discussions |
| `poll_id` | bigint | - | YES | FK to polls |
| `authored_at` | datetime | - | YES | Content creation time |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

**Indexes:**
- `ts_content` (GIN)
- `(searchable_type, searchable_id)`
- `author_id`
- `group_id`
- `discussion_id`
- `poll_id`
- `authored_at` (asc and desc)

**Searchable Models:** Discussion, Comment, Poll, Stance, Outcome

---

## External Integrations

### chatbots

Webhook and Matrix bot configurations.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | bigserial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups |
| `author_id` | integer | - | YES | FK to users |
| `name` | string | - | YES | Display name |
| `kind` | string | - | YES | "matrix" or "webhook" |
| `server` | string | - | YES | Server URL |
| `channel` | string | - | YES | Channel/room ID |
| `access_token` | string | - | YES | Auth token |
| `webhook_kind` | string | - | YES | slack, microsoft, discord, markdown, webex |
| `event_kinds` | string[] | - | YES | Subscribed event types |
| `notification_only` | boolean | false | NO | No payload, just ping |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

---

### webhooks

Outbound webhook configurations (newer model).

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | bigserial | auto | NO | Primary key |
| `group_id` | integer | - | NO | FK to groups |
| `author_id` | integer | - | YES | FK to users |
| `actor_id` | integer | - | YES | FK to users (last actor) |
| `name` | string | - | NO | Display name |
| `url` | string | - | YES | Webhook URL |
| `token` | string | - | YES | Auth token |
| `format` | string | "markdown" | YES | Payload format |
| `event_kinds` | jsonb | [] | NO | Subscribed events |
| `permissions` | string[] | [] | NO | Granted permissions |
| `include_body` | boolean | false | YES | Include full content |
| `include_subgroups` | boolean | false | NO | Include subgroup events |
| `is_broken` | boolean | false | NO | Delivery failing |
| `last_used_at` | datetime | - | YES | Last delivery time |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

---

### subscriptions

Billing/subscription information.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `owner_id` | integer | - | YES | FK to users |
| `chargify_subscription_id` | integer | - | YES | External billing ID |
| `plan` | string | "free" | YES | Plan name |
| `state` | string | "active" | NO | active, canceled, etc. |
| `payment_method` | string | "none" | NO | Payment type |
| `expires_at` | datetime | - | YES | Expiration time |
| `canceled_at` | datetime | - | YES | Cancellation time |
| `activated_at` | datetime | - | YES | Activation time |
| `renews_at` | datetime | - | YES | Next renewal |
| `renewed_at` | datetime | - | YES | Last renewal |
| `max_threads` | integer | - | YES | Thread limit |
| `max_members` | integer | - | YES | Member limit |
| `max_orgs` | integer | - | YES | Organization limit |
| `members_count` | integer | - | YES | Current member count |
| `allow_subgroups` | boolean | true | NO | Subgroups enabled |
| `allow_guests` | boolean | true | NO | Guests enabled |
| `lead_status` | string | - | YES | Sales lead status |
| `info` | jsonb | - | YES | Additional metadata |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

---

## System & Infrastructure

### received_emails

Inbound email queue for Action Mailbox.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | bigserial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups |
| `headers` | hstore | {} | NO | Email headers |
| `body_text` | string | - | YES | Plain text body |
| `body_html` | string | - | YES | HTML body |
| `spf_valid` | boolean | false | NO | SPF check passed |
| `dkim_valid` | boolean | false | NO | DKIM check passed |
| `released` | boolean | false | NO | Processed/released |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

---

### tags

Tag definitions scoped to groups.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups |
| `name` | citext | - | NO | Tag name (case-insensitive) |
| `color` | string | - | YES | Display color |
| `priority` | integer | 0 | NO | Sort order |
| `taggings_count` | integer | 0 | YES | Direct usage count |
| `org_taggings_count` | integer | 0 | NO | Org-wide usage count |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

**Indexes:**
- `(group_id, name)` (unique)
- `group_id`
- `name`

---

### tasks

Task/to-do items within discussions.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | bigserial | auto | NO | Primary key |
| `record_type` | string | - | YES | Polymorphic type |
| `record_id` | bigint | - | YES | Polymorphic FK |
| `author_id` | bigint | - | NO | FK to users |
| `doer_id` | integer | - | YES | FK to users (assigned) |
| `uid` | integer | - | NO | Unique ID within record |
| `name` | string | - | NO | Task description |
| `done` | boolean | - | NO | Completion status |
| `done_at` | datetime | - | YES | Completion time |
| `due_on` | date | - | YES | Due date |
| `remind` | integer | - | YES | Reminder setting |
| `remind_at` | datetime | - | YES | Reminder time |
| `discarded_at` | datetime | - | YES | Soft delete |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

---

### versions

Paper Trail audit log.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `item_type` | string(255) | - | NO | Model class name |
| `item_id` | integer | - | NO | Model record ID |
| `event` | string(255) | - | NO | create, update, destroy |
| `whodunnit` | integer | - | YES | User ID who made change |
| `object_changes` | jsonb | - | YES | Changed attributes |
| `created_at` | datetime | - | YES | |

**Tracked Models:** Discussion, Poll, Comment, Group

---

### translations

Cached translations for content.

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `translatable_id` | integer | - | YES | Polymorphic FK |
| `translatable_type` | string(255) | - | YES | Polymorphic type |
| `language` | string(255) | - | YES | Target language |
| `fields` | hstore | - | YES | Field translations |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

---

## JSONB Column Schemas

### attachments (common pattern)

Used on: discussions, comments, polls, stances, outcomes, users, groups, discussion_templates, poll_templates

```json
[
  {
    "id": "abc123",
    "type": "image",
    "filename": "photo.jpg",
    "content_type": "image/jpeg",
    "byte_size": 12345,
    "checksum": "xyz789",
    "signed_id": "eyJ...",
    "preview_url": "https://..."
  }
]
```

### link_previews (common pattern)

```json
[
  {
    "url": "https://example.com/page",
    "title": "Page Title",
    "description": "Page description",
    "image": "https://example.com/og-image.jpg",
    "hostname": "example.com"
  }
]
```

### experiences (users, memberships)

Feature flags and tutorial completion tracking:

```json
{
  "welcomeModal": true,
  "announcementHelpCard": true,
  "pollTypes": ["proposal", "count", "poll"]
}
```

### custom_fields (events, polls, outcomes, identities)

Extensible key-value storage for type-specific data:

```json
{
  "user_ids": [1, 2, 3],
  "group_ids": [10, 20],
  "recipient_user_ids": [5, 6]
}
```

### stance_data (polls)

Aggregated voting data:

```json
{
  "agree": 10,
  "disagree": 3,
  "abstain": 2,
  "block": 0
}
```

### option_scores (stances)

Per-option scores for a stance:

```json
{
  "123": 2,
  "124": 1,
  "125": 0
}
```

---

## PostgreSQL Extensions Used

| Extension | Purpose |
|-----------|---------|
| `citext` | Case-insensitive text (emails, handles) |
| `hstore` | Key-value storage (email headers, translations) |
| `pgcrypto` | UUID generation |
| `pg_stat_statements` | Query performance monitoring |
| `plpgsql` | Procedural language |

---

*Generated: 2026-02-01*
*Source: db/schema.rb version 2025_12_03_031449*
