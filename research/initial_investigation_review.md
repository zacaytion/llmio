# Loomio Investigation Documents Review

> Review of `loomio_initial_investigation.md` and `schema_investigation.md` for consistency, clarity, and completeness.
> Generated: 2026-01-30

## Table of Contents

1. [Contradictions & Inaccuracies](#1-contradictions--inaccuracies)
2. [Areas Lacking Clarity](#2-areas-lacking-clarity)
3. [Missing Documentation](#3-missing-documentation)
4. [Unanswered Questions](#4-unanswered-questions)
5. [Other Findings](#5-other-findings)
6. [Recommendations](#6-recommendations)

---

## 1. Contradictions & Inaccuracies

### 1.1 Poll Types - Incomplete List

**In Documents:**
> `loomio_initial_investigation.md` Section 4.4 lists 7 poll types:
> `proposal`, `poll`, `count`, `score`, `ranked_choice`, `meeting`, `dot_vote`

**Actual (9 types):**

| Poll Type | Config Lines | Missing from Docs |
|-----------|--------------|-------------------|
| `count` | 1-41 | |
| `check` | 43-89 | **YES** |
| `question` | 91-112 | **YES** |
| `proposal` | 114-207 | |
| `meeting` | 209-232 | |
| `poll` | 234-261 | |
| `dot_vote` | 263-284 | |
| `score` | 286-307 | |
| `ranked_choice` | 309-331 | |

**Source:** `orig/loomio/config/poll_types.yml`

**Impact:** Go implementation would be missing 2 poll types. The `check` type appears to be a simple checkbox/attendance poll, while `question` appears to be for open-ended questions.

---

### 1.2 Event Kinds - Severely Incomplete

**In Documents:**
> `loomio_initial_investigation.md` Section 4.9 lists ~10 event kinds

**Actual: 42 event kinds**

<details>
<summary>Complete Event Kinds List (click to expand)</summary>

**Discussion Events:**
1. `new_discussion`
2. `discussion_edited`
3. `discussion_title_edited`
4. `discussion_description_edited`
5. `discussion_closed`
6. `discussion_reopened`
7. `discussion_forked`
8. `discussion_moved`
9. `discussion_announced`

**Comment Events:**
10. `new_comment`
11. `comment_edited`
12. `comment_replied_to`

**Poll Events:**
13. `poll_created`
14. `poll_edited`
15. `poll_closing_soon`
16. `poll_closed_by_user`
17. `poll_expired`
18. `poll_reopened`
19. `poll_option_added`
20. `poll_announced`
21. `poll_reminder`

**Stance Events:**
22. `stance_created`
23. `stance_updated`

**Outcome Events:**
24. `outcome_created`
25. `outcome_updated`
26. `outcome_announced`
27. `outcome_review_due`

**Membership Events:**
28. `membership_created`
29. `membership_requested`
30. `membership_request_approved`
31. `membership_resent`
32. `invitation_accepted`
33. `user_added_to_group`
34. `user_joined_group`
35. `new_coordinator`
36. `new_delegate`

**Mention Events:**
37. `user_mentioned`
38. `group_mentioned`

**Other Events:**
39. `reaction_created`
40. `announcement_resend`
41. `user_reactivated`
42. `unknown_sender`

</details>

**Source:** `orig/loomio/app/models/events/` (42 files)

**Additional Note:** Only 14 of these are webhook-eligible:
- Source: `orig/loomio/config/webhook_event_kinds.yml`

---

### 1.3 Attachments JSONB - Incorrect Default & Structure

**In Documents:**
> `schema_investigation.md` line 537: `DEFAULT '[]'::jsonb`
>
> Structure shown as: `[{signed_id, filename, content_type, byte_size}]`

**Actual:**

Default: `'{}'::jsonb` (empty object, not empty array)

Full structure:
```typescript
{
  id: number,           // blob_id
  filename: string,
  content_type: string,
  byte_size: number,
  preview_url: string,  // Missing from docs
  download_url: string, // Missing from docs
  icon: string,         // Missing from docs
  signed_id: string
}
```

**Source:** `orig/loomio/app/models/concerns/has_rich_text.rb:92-103`

---

### 1.4 Link Previews - Missing Fields

**In Documents:**
> Structure: `{url, title, description, image_url}`

**Actual:**
```typescript
{
  title: string,       // max 240 chars
  description: string, // max 240 chars
  image: string,       // Note: 'image' not 'image_url'
  url: string,
  fit: string,         // Missing - 'contain'
  align: string,       // Missing - 'center'
  hostname: string     // Missing
}
```

**Source:** `orig/loomio/app/services/link_preview_service.rb:26-32`

---

### 1.5 Internal Inconsistency - schema_investigation.md

Line 537 shows attachments default as `'[]'::jsonb` but the actual SQL schema shows `'{}'::jsonb`. This appears to be a transcription error.

---

## 2. Areas Lacking Clarity

### 2.1 Volume Levels - Incomplete Explanation

**In Documents:**
> "Values appear to be: 0=mute, 1=quiet, 2=normal, 3=loud (based on Rails code)"

**Clarification Needed:**

| Level | Value | Behavior |
|-------|-------|----------|
| `mute` | 0 | No notifications at all |
| `quiet` | 1 | App notifications only (no email) |
| `normal` | 2 | Both email and app notifications |
| `loud` | 3 | Maximum engagement - includes all notifications plus extras |

**Source:** `orig/loomio/app/models/concerns/has_volume.rb:5`

Helper methods available:
- `volume_is_mute?`
- `volume_is_quiet?`
- `volume_is_normal?`
- `volume_is_loud?`
- `volume_is_normal_or_loud?`

---

### 2.2 Custom Fields Macro - Unexplained Pattern

Documents mention `custom_fields` JSONB but don't explain the `set_custom_fields` macro pattern.

**How It Works:**
```ruby
# orig/loomio/app/models/concerns/has_custom_fields.rb:2-7
set_custom_fields(*fields) # creates getter/setter for each field
# Accesses self[:custom_fields][field_name]
```

**Usage by Model:**

| Model | Custom Fields | Source |
|-------|---------------|--------|
| Poll | `meeting_duration`, `time_zone`, `can_respond_maybe` | `poll.rb:56-58` |
| Event | `pinned_title`, `recipient_user_ids`, `recipient_chatbot_ids`, `recipient_message`, `recipient_audience`, `stance_ids` | `event.rb:12` |
| Outcome | `event_summary`, `event_description`, `event_location` | `outcome.rb:51` |

---

### 2.3 Webhook-Eligible Events Not Clarified

Documents mention webhooks but don't distinguish which event kinds can trigger webhooks.

**Only 14 of 42 events are webhook-eligible.**

Source: `orig/loomio/config/webhook_event_kinds.yml`

---

### 2.4 Stance latest Boolean - Enforcement Mechanism Unclear

Documents note the unique partial index:
```sql
CREATE UNIQUE INDEX ON stances (poll_id, participant_id, latest) WHERE latest = true;
```

But don't explain how `latest` is managed. The application must:
1. Set all existing stances for (poll_id, participant_id) to `latest = false`
2. Set the new stance to `latest = true`

This is done in `StanceService` during vote updates.

---

## 3. Missing Documentation

### 3.1 Email/Mailer System (Not Covered)

**7 Mailers Found:**

| Mailer | Emails Sent |
|--------|-------------|
| BaseMailer | Base class with UTM tracking, spam checking |
| UserMailer | `redacted`, `accounts_merged`, `merge_verification`, `catch_up`, `membership_request_approved`, `user_added_to_group`, `group_export_ready`, `login`, `contact_request` |
| EventMailer | `event` - generic notification for discussions, comments, polls, stances, outcomes |
| GroupMailer | `destroy_warning` |
| ContactMailer | `contact_message` |
| TaskMailer | `task_due_reminder` |
| ForwardMailer | `forward_message`, `bounce` |

**Source:** `orig/loomio/app/mailers/`

**Catch-up Email System:**
- Worker: `send_daily_catch_up_email_worker.rb`
- Runs at 6 AM in user's timezone
- Supports daily, weekly, bi-weekly frequencies
- Controlled by `email_catch_up_day` user field

---

### 3.2 Rate Limiting System (Not Covered)

**ThrottleService:**

```ruby
# orig/loomio/app/services/throttle_service.rb

ThrottleService.can?(key:, id:, max:, inc:, per:)
# Returns true if under limit, increments counter

ThrottleService.limit!(...)
# Same as can? but raises ThrottleService::LimitReached

ThrottleService.reset!(period)
# Clears all throttle keys for period ('hour' or 'day')
```

**Configuration:**
- Redis-backed counters
- Key pattern: `THROTTLE-{HOUR|DAY}-{key}-{id}`
- Default limit: 100 (overridable via `ENV['THROTTLE_MAX_{KEY}']`)

**Used For:**
- Email bounce throttling (1 per hour)
- API rate limiting

---

### 3.3 Demo System (Not Covered)

**DemoService Methods:**

| Method | Purpose |
|--------|---------|
| `refill_queue` | Maintains Redis queue of pre-cloned demo groups |
| `take_demo(actor)` | Assigns demo group to user, adds membership |
| `ensure_queue` | Validates queue, removes deleted groups |
| `generate_demo_groups` | Creates public demo templates |

**Demo Model:** `orig/loomio/app/models/demo.rb`
- Fields: `name`, `description`, `demo_handle`, `recorded_at`, `author_id`, `group_id`, `priority`

**Redis Queue:** `demo_group_ids`
**Queue Size:** `ENV['FEATURES_DEMO_GROUPS_SIZE']` (default: 3)

---

### 3.4 Subscription Plans (Incomplete)

**Plan Types:**
- `free` - Default plan
- `demo` - Demo/trial plan
- Legacy: `was-gift`, `was-paid` (consolidated to free)

**Subscription States:**
- `active`, `on_hold`, `pending`, `past_due`, `trialing`, `canceled`

**Payment Methods:**
- `chargify`, `manual`, `barter`, `paypal`, `none`

**Feature Flags:**
- `max_members` - Member limit
- `max_threads` - Discussion limit
- `max_orgs` - Organization limit
- `allow_subgroups` - Enable subgroup creation
- `allow_guests` - Allow guest users

**Source:** `orig/loomio/app/models/subscription.rb`

---

### 3.5 File Storage Backends (Not Covered)

**5 Storage Options:**

| Backend | Service | Environment Variables |
|---------|---------|----------------------|
| Disk (test) | Disk | Root: `tmp/storage` |
| Disk (local) | Disk | Root: `storage/` |
| Amazon S3 | S3 | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_BUCKET`, `AWS_REGION` |
| DigitalOcean Spaces | S3 | `DO_ENDPOINT`, `DO_ACCESS_KEY_ID`, `DO_SECRET_ACCESS_KEY`, `DO_BUCKET` |
| S3-Compatible | S3 | `STORAGE_ENDPOINT`, `STORAGE_ACCESS_KEY_ID`, etc. |
| Google Cloud Storage | GCS | `GCS_CREDENTIALS`, `GCS_PROJECT`, `GCS_BUCKET` |

**Source:** `orig/loomio/config/storage.yml`

---

### 3.6 Services Not Documented

The documents list 46+ services but don't describe many important ones:

| Service | Purpose |
|---------|---------|
| `LinkPreviewService` | Fetches URL metadata for previews |
| `RecordCloner` | Duplicates records with associations |
| `CleanupService` | Deletes orphan records |
| `DemoService` | Demo group management |
| `ThrottleService` | Rate limiting |
| `TranslationService` | Content translation |
| `GroupExportService` | Data export generation |

---

## 4. Unanswered Questions

### 4.1 Hocuspocus Collaborative Editing

**Question:** How does collaborative editing sync between Rails and the Node.js channel server?

**Partial Answer Found:**
- Token format: `{user_id},{secret_token}`
- Document names: `{record_type}-{record_id}-{user_id_if_new}`
- Rails validates via `/api/hocuspocus` endpoint

**Still Unknown:**
- How are Y.js documents stored/persisted?
- How is conflict resolution handled?
- What happens when documents are edited offline?

---

### 4.2 Webhook Permissions Array

**Question:** What permissions can be set in `webhooks.permissions`?

**Found:**
- Column: `character varying[] DEFAULT '{}'`
- Used for filtering which records are included in webhook payloads

**Still Unknown:**
- Exact permission values
- How permissions map to record types

---

### 4.3 Search Indexing Triggers

**Question:** When/how is `pg_search_documents` populated?

**Partial Answer:**
- Uses pg_search gem
- Indexes discussions, comments, polls
- Has `ts_content` tsvector column

**Still Unknown:**
- Are updates synchronous or via background job?
- What triggers reindexing?
- How is the `content` field composed?

---

### 4.4 SAML/OAuth Detailed Flows

**Question:** What's the complete authentication flow for SSO?

**Found:**
- Routes documented
- Two modes: SSO-only vs standard
- Controllers in `app/controllers/identities/`

**Still Unknown:**
- Attribute mapping
- Group provisioning
- Session management details

---

### 4.5 Translation System

**Question:** How are translations managed beyond the `translations` table?

**Partial Answer:**
- `translations` table stores content translations
- `translatable_type`, `translatable_id` for polymorphic association
- `fields` hstore for translated field key-values

**Still Unknown:**
- How are translations requested/created?
- What translation service is used?
- Is auto-translation supported?

---

### 4.6 RecordCloner Details

**Question:** How exactly does group cloning work for demos?

**Found:**
- Service: `app/services/record_cloner.rb`
- Used by DemoService

**Still Unknown:**
- Which associations are cloned?
- How are IDs remapped?
- How is user data anonymized?

---

### 4.7 Subscription Plan Hierarchy

**Question:** What's the full plan level hierarchy in SubscriptionService::PLANS?

**Found:**
- Reference in `subscription.rb` to `SubscriptionService::PLANS`

**Still Unknown:**
- Exact plan levels/tiers
- Feature differences between plans
- Pricing structure

---

## 5. Other Findings

### 5.1 Experiences JSONB - Used Inconsistently

Documents show `experiences` on users and memberships, but it was **removed from groups**:
- Migration: `20230615234611_remove_experiences_from_groups.rb`

**Known Keys Used:**
- `html-editor.uses-markdown` - Editor format preference
- `betaFeatures` - Beta feature access
- `happiness` - (test usage)

---

### 5.2 Counter Caches - Extensive Use

Both documents mention counter caches but neither provides a complete inventory. The group model alone has **17 counter cache columns**:

- `memberships_count`
- `admin_memberships_count`
- `pending_memberships_count`
- `discussions_count`
- `public_discussions_count`
- `open_discussions_count`
- `closed_discussions_count`
- `polls_count`
- `closed_polls_count`
- `closed_motions_count`
- `proposal_outcomes_count`
- `subgroups_count`
- `invitations_count`
- `recent_activity_count`
- `discussion_templates_count`
- `poll_templates_count`
- `delegates_count`

---

### 5.3 Position Key Format

Documents mention `position_key` but don't explain the format:

```
Format: "parent_key-position"
Example: "00001-00002-00003"
```

Zero-padded to allow string sorting while maintaining tree order.

**Source:** Referenced in event positioning logic

---

### 5.4 Email-to-Thread Formats

`schema_investigation.md` doesn't document the special email address formats:

```
d=100&k=key&u=999@mail.loomio.com  # Email to specific thread
group+u=99&k=key@mail.loomio.com   # Email to group (new thread)
```

**Source:** `orig/loomio/app/services/received_email_service.rb:62-76`

---

## 6. Recommendations

### 6.1 Immediate Updates to Investigation Docs

1. **Add missing poll types:** `check`, `question`
2. **Expand event kinds section** with complete 42-item list
3. **Fix attachments JSONB default** from `[]` to `{}`
4. **Add missing link preview fields:** `fit`, `align`, `hostname`
5. **Document volume level behaviors** beyond just the enum values

### 6.2 New Documentation Sections Needed

1. **Email System** - Mailers, notification delivery, catch-up emails
2. **Rate Limiting** - ThrottleService usage and configuration
3. **Demo System** - DemoService, RecordCloner, queue management
4. **Storage Backends** - ActiveStorage configuration options
5. **Search Indexing** - pg_search configuration and triggers

### 6.3 Follow-up Investigation Areas

1. **Hocuspocus Integration** - Deep dive into collaborative editing
2. **RecordCloner** - Document cloning logic for Go implementation
3. **SubscriptionService** - Full plan hierarchy and feature matrix
4. **Translation Service** - How translations are requested/stored
5. **Webhook Permissions** - Complete permission value inventory

---

*Document generated: 2026-01-30*
*Review of: loomio_initial_investigation.md, schema_investigation.md*
