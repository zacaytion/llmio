# Business Logic Specification

**Document Version:** 1.0
**Generated:** 2026-02-01
**Purpose:** Define business logic behaviors for Loomio application rewrite

---

## Table of Contents

1. [Service Layer Patterns](#1-service-layer-patterns)
2. [Event System](#2-event-system)
3. [Permission System](#3-permission-system)
4. [Notification Flow](#4-notification-flow)
5. [Real-time Updates](#5-real-time-updates)
6. [Search System](#6-search-system)
7. [Rate Limiting](#7-rate-limiting)
8. [File Storage](#8-file-storage)
9. [Templates](#9-templates)
10. [Demo System](#10-demo-system)
11. [Transcription Service](#11-transcription-service)
12. [Group Export](#12-group-export)
13. [Uncertainties](#13-uncertainties)

---

## 1. Service Layer Patterns

**Confidence: HIGH**

### 1.1 Mutation Flow

All data mutations follow a consistent pattern through service classes:

```
Controller
    -> Service.action(model:, actor:)
        -> Authorization check (CanCanCan)
        -> Business logic
        -> Model.save
        -> Event.publish!
        -> PublishEventWorker.perform_async(event_id)
```

### 1.2 Core Services

| Service | Model | Key Methods |
|---------|-------|-------------|
| `DiscussionService` | Discussion | `create`, `update`, `close`, `reopen`, `move`, `discard`, `destroy` |
| `PollService` | Poll | `create`, `update`, `close`, `reopen`, `expire_lapsed_polls`, `publish_closing_soon` |
| `StanceService` | Stance | `create`, `update`, `revoke` |
| `CommentService` | Comment | `create`, `update`, `discard`, `destroy` |
| `GroupService` | Group | `create`, `update`, `archive`, `destroy` |
| `MembershipService` | Membership | `create`, `update`, `destroy`, `add_members`, `add_guests` |
| `OutcomeService` | Outcome | `create`, `update`, `publish_review_due` |
| `UserService` | User | `verify`, `deactivate`, `reactivate` |

### 1.3 Service Method Structure

```ruby
def self.action(model:, actor:, params: {})
  # 1. Authorization
  actor.ability.authorize! :action, model

  # 2. Validation
  return false unless model.valid?

  # 3. Business logic / attribute assignment
  model.assign_attributes(params)

  # 4. Persistence
  model.save!

  # 5. Side effects (events, reindexing, etc.)
  Events::ModelAction.publish!(model, user: actor)

  # 6. Return value
  model
end
```

### 1.4 Vote Revision Rule

**Stance update creates new record only when ALL conditions met:**

| Condition | Value |
|-----------|-------|
| Time since last vote | > 15 minutes |
| Choices changed | YES |
| Poll is in discussion | YES |

Otherwise, the existing stance record is updated in place.

---

## 2. Event System

**Confidence: HIGH**

### 2.1 Event Architecture

Events use Rails STI (Single Table Inheritance) with a base `Event` model and specific event type classes.

```ruby
# Event creation flow
Event.publish!(eventable, user: actor)
  -> Event.create!
  -> PublishEventWorker.perform_async(event.id)
  -> Event#trigger! (calls concern methods)
```

### 2.2 Complete Event Type Reference (42 Types)

#### Events with Real-time Broadcasts (16)

| Event Type | Triggers | Side Effects |
|------------|----------|--------------|
| `NewComment` | Comment created | LiveUpdate, Mentions |
| `CommentEdited` | Comment updated | LiveUpdate, Mentions |
| `NewDiscussion` | Discussion created | LiveUpdate, InApp, Mentions |
| `DiscussionEdited` | Discussion updated | LiveUpdate, InApp, Mentions |
| `DiscussionClosed` | Discussion closed | LiveUpdate |
| `DiscussionReopened` | Discussion reopened | LiveUpdate |
| `DiscussionMoved` | Discussion moved to group | LiveUpdate |
| `PollCreated` | Poll created | LiveUpdate, InApp, Mentions |
| `PollEdited` | Poll updated | LiveUpdate, InApp, Mentions |
| `PollClosedByUser` | Poll manually closed | LiveUpdate |
| `StanceCreated` | Vote cast | LiveUpdate, InApp |
| `StanceUpdated` | Vote updated | LiveUpdate, InApp |
| `OutcomeCreated` | Outcome published | LiveUpdate, InApp |
| `OutcomeUpdated` | Outcome edited | LiveUpdate, InApp |
| `InvitationAccepted` | Membership accepted | LiveUpdate, InApp |
| `ReactionCreated` | Reaction added | LiveUpdate, InApp |

#### Events with Notifications Only (18)

| Event Type | Triggers | Recipients |
|------------|----------|------------|
| `UserMentioned` | @user in content | Mentioned users |
| `GroupMentioned` | @group in content | Group members |
| `CommentRepliedTo` | Reply to comment | Comment author |
| `MembershipCreated` | Invitation sent | Invitee |
| `MembershipRequestApproved` | Request approved | Requester |
| `MembershipRequested` | Membership requested | Group admins |
| `NewCoordinator` | Admin added | New admin |
| `NewDelegate` | Delegate added | New delegate |
| `UserAddedToGroup` | Member added | New member |
| `DiscussionAnnounced` | Discussion announcement | Announcement recipients |
| `PollAnnounced` | Poll announcement | Announcement recipients |
| `OutcomeAnnounced` | Outcome announcement | Announcement recipients |
| `PollClosingSoon` | Poll closing within 24h | Undecided voters (configurable) |
| `PollExpired` | Poll auto-closed | Poll author |
| `PollReminder` | Manual reminder sent | Poll recipients |
| `PollOptionAdded` | Option added to poll | Voters |
| `OutcomeReviewDue` | Outcome review date reached | Outcome author |
| `UnknownSender` | Reply email from unknown | Group admins |

#### Events without User Notifications (8)

| Event Type | Purpose | Delivery |
|------------|---------|----------|
| `AnnouncementResend` | Re-send invitation email | Email only |
| `MembershipResent` | Re-send membership email | Email only |
| `PollReopened` | Poll reopened | Chatbots only |
| `DiscussionDescriptionEdited` | Legacy event | None (unused) |
| `DiscussionTitleEdited` | Legacy event | None (unused) |
| `DiscussionForked` | Discussion forked | None |
| `UserJoinedGroup` | User self-joined | None |
| `UserReactivated` | User reactivated | Email only |

### 2.3 Event Concerns

Events include behavior via concerns:

| Concern | Purpose |
|---------|---------|
| `Events::LiveUpdate` | Publish to Redis for real-time clients |
| `Events::Notify::InApp` | Create Notification records, publish to user rooms |
| `Events::Notify::ByEmail` | Queue email delivery |
| `Events::Notify::Chatbots` | Queue webhook/matrix delivery |
| `Events::Notify::Mentions` | Process @mentions, create mention events |

### 2.4 Webhook-Eligible Events (14)

Configured in `config/webhook_event_kinds.yml`:

```yaml
- new_discussion
- discussion_edited
- new_comment
- poll_created
- poll_edited
- poll_closing_soon
- poll_expired
- poll_closed_by_user
- poll_reopened
- outcome_created
- outcome_updated
- outcome_review_due
- stance_created
- stance_updated
```

---

## 3. Permission System

**Confidence: HIGH**

### 3.1 CanCanCan Ability Files

Authorization is implemented via CanCanCan abilities split by model:

| File | Authorizes |
|------|------------|
| `app/models/ability/group.rb` | Group, Membership actions |
| `app/models/ability/discussion.rb` | Discussion actions |
| `app/models/ability/poll.rb` | Poll actions |
| `app/models/ability/comment.rb` | Comment actions |
| `app/models/ability/stance.rb` | Stance/vote actions |
| `app/models/ability/outcome.rb` | Outcome actions |
| `app/models/ability/poll_template.rb` | PollTemplate actions |
| `app/models/ability/discussion_template.rb` | DiscussionTemplate actions |
| `app/models/ability/chatbot.rb` | Chatbot/webhook actions |
| `app/models/ability/tag.rb` | Tag actions |

### 3.2 Permission Flags

#### Group Permission Flags (12 total)

| Flag | Default | Paper Trail | Used | Description |
|------|---------|-------------|------|-------------|
| `members_can_add_members` | false | YES | YES | Invite new members |
| `members_can_edit_discussions` | true | YES | YES | Edit any discussion |
| `members_can_edit_comments` | true | YES | YES | Edit own comments |
| `members_can_delete_comments` | true | YES | YES | Delete own comments |
| `members_can_raise_motions` | true | YES | YES | Create polls |
| `members_can_start_discussions` | true | YES | YES | Create discussions |
| `members_can_create_subgroups` | false | YES | YES | Create child groups |
| `members_can_announce` | true | YES | YES | Send notifications |
| `members_can_add_guests` | true | **NO** | YES | Invite non-members |
| `members_can_vote` | true | NO | **NO (deprecated)** | Unused |
| `admins_can_edit_user_content` | true | YES | YES | Admins edit others' content |
| `parent_members_can_see_discussions` | false | YES | YES | Subgroup visibility |

#### Permission Check Pattern

```ruby
# In ability file
can [:action], Model do |model|
  model.group.members_can_do_thing && model.group.members.exists?(user.id)
  || model.group.admins.exists?(user.id)
end
```

### 3.3 Null::Group (Direct Discussions)

Direct discussions (no group) use a Null Object pattern with specific permission defaults:

| Flag | NullGroup Value | Reasoning |
|------|-----------------|-----------|
| `members_can_add_guests` | true | Participants can invite |
| `members_can_edit_discussions` | true | Can edit |
| `members_can_edit_comments` | true | Can edit own |
| `members_can_delete_comments` | true | Can delete own |
| `members_can_raise_motions` | true | Can create polls |
| `members_can_announce` | true | Can notify |
| `members_can_add_members` | false | No membership concept |
| `members_can_start_discussions` | false | No sub-discussions |
| `members_can_create_subgroups` | false | No subgroups |
| `admins_can_edit_user_content` | false | No admin role |

### 3.4 Role Hierarchy

| Role | Permissions |
|------|-------------|
| Guest | View content, vote (if allowed), add comments |
| Member | All guest + create content, invite guests (if flags allow) |
| Admin | All member + manage group settings, members, delete content |

---

## 4. Notification Flow

**Confidence: HIGH**

### 4.1 Notification Pipeline

```
Event.trigger!
    |
    +-> InApp notifications
    |   -> Notification.create! (for each recipient)
    |   -> MessageChannelService.publish_models(user_id:)
    |
    +-> Email notifications
    |   -> EventMailer.deliver_later
    |   -> Check: active, verified, no spam complaints
    |   -> Check: volume preferences
    |
    +-> Webhook notifications
        -> GenericWorker.perform_async('ChatbotService', 'publish_event!')
        -> POST to webhook URL
```

### 4.2 Email Delivery Rules

| Condition | Email Sent? |
|-----------|-------------|
| User has `complaints_count > 0` | NO |
| User not `email_verified` | NO |
| User volume = "quiet" | NO (except direct mentions) |
| User deactivated | NO |

### 4.3 Catch-up Email Schedule

| `email_catch_up_day` | Frequency | Description |
|----------------------|-----------|-------------|
| 7 | Daily | Every day at 6 AM user timezone |
| 8 | Every other day | Odd weekday at 6 AM |
| 0-6 | Weekly | Specific day of week at 6 AM |

### 4.4 Reply-by-Email Address Format

```
d={discussion_id}&u={user_id}&k={email_api_key}@{REPLY_HOSTNAME}
pt=c&pi={comment_id}&d={discussion_id}&u={user_id}&k={api_key}@{REPLY_HOSTNAME}
```

| Parameter | Meaning |
|-----------|---------|
| `d` | Discussion ID |
| `u` | User ID |
| `k` | User's email_api_key (auth token) |
| `pt` | Parent type (c=Comment, p=Poll, s=Stance, o=Outcome) |
| `pi` | Parent ID |

---

## 5. Real-time Updates

**Confidence: HIGH**

### 5.1 Pub/Sub Architecture

```
Event Created
    -> PublishEventWorker.perform_async(event_id)
    -> Event.trigger!
    -> LiveUpdate.notify_clients!
    -> MessageChannelService.publish_models(records, group_id: or user_id:)
    -> Redis.publish("/records", {room:, records:})
    -> Socket.io server
    -> Client WebSocket
```

### 5.2 Room Routing

| Room Pattern | Use Case | Priority |
|--------------|----------|----------|
| `group-{id}` | Group member broadcasts | Highest (wins if both specified) |
| `user-{id}` | Personal notifications, guest updates | Lower |
| `notice` | System-wide announcements | N/A |

### 5.3 Redis Channels

| Channel | Purpose |
|---------|---------|
| `/records` | Model updates (events, comments, etc.) |
| `/system_notice` | System broadcasts (maintenance, reload) |
| `chatbot/publish` | Matrix chatbot events |
| `chatbot/test` | Matrix chatbot connection test |

### 5.4 Message Payload Format

```json
{
  "room": "group-123",
  "records": {
    "events": [...],
    "discussions": [...],
    "comments": [...],
    "users": [...]
  }
}
```

### 5.5 Guest Routing

Events on discussions with guests publish to each guest individually:

```ruby
eventable.guests.find_each do |user|
  MessageChannelService.publish_models([self], user_id: user.id)
end
```

---

## 6. Search System

**Confidence: HIGH**

### 6.1 pg_search Configuration

| Setting | Value |
|---------|-------|
| Text configuration | `'simple'` (no stemming) |
| Prefix matching | Enabled |
| Results limit | 20 |
| Highlight tags | `<b>` / `</b>` |

### 6.2 Searchable Models

| Model | Indexed Content |
|-------|-----------------|
| Discussion | title, description, author name |
| Comment | body, author name |
| Poll | title, details, author name |
| Stance | reason, participant name |
| Outcome | statement, author name |

### 6.3 Reindex Triggers

| Trigger | Method Called |
|---------|---------------|
| Discussion discarded | `SearchService.reindex_by_discussion_id` |
| Discussion moved | `SearchService.reindex_by_discussion_id` |
| Poll updated | `SearchService.reindex_by_poll_id` |
| Poll closed | `SearchService.reindex_by_poll_id` |
| User reactivated | `SearchService.reindex_by_author_id` |
| User name changed | `SearchService.reindex_by_author_id` |

### 6.4 Stance Indexing Rules

Stances are excluded from search index if:
- `cast_at IS NULL` (vote not cast)
- Poll is anonymous AND still open
- Poll has `hide_results = until_closed` AND still open

### 6.5 Full Reindex

```ruby
SearchService.reindex_everything
# OR async via:
GenericWorker.perform_async('SearchService', 'reindex_everything')
```

---

## 7. Rate Limiting

**Confidence: HIGH**

### 7.1 Three-Layer Architecture

| Layer | Implementation | Response |
|-------|----------------|----------|
| 1. IP-based | Rack::Attack middleware | HTTP 429 |
| 2. User-based | ThrottleService (Redis) | HTTP 500 (bug - should be 429) |
| 3. Authentication | Devise Lockable | Account lockout |

### 7.2 Rack::Attack IP Limits

| Endpoint | Requests/Hour/IP |
|----------|------------------|
| `/api/v1/trials` | 10 |
| `/api/v1/announcements` | 100 |
| `/api/v1/memberships` | 100 |
| `/api/v1/membership_requests` | 50 |
| `/api/v1/invitations` | 50 |
| `/api/v1/webhooks` | 50 |
| `/api/v1/login_tokens` | 50 |
| `/api/v1/profile` | 50 |
| `/api/v1/email_actions` | 50 |
| `/api/v1/discussions` | 500 |
| `/api/v1/comments` | 500 |
| `/api/v1/stances` | 500 |
| `/rails/active_storage/direct_uploads` | 20 |

### 7.3 ThrottleService Usage

| Location | Key | Limit | Period |
|----------|-----|-------|--------|
| UserInviter | `UserInviterInvitations` | `user.invitations_rate_limit` | day |
| ReceivedEmailService | `bounce` | 1 | hour |

### 7.4 ThrottleService Reset

Redis counters have NO automatic TTL. Reset via scheduled task:

```ruby
# Hourly
ThrottleService.reset!('hour')

# Daily (at midnight)
ThrottleService.reset!('day')
```

### 7.5 Bot API Rate Limits

**WARNING:** Bot APIs (`/api/b1/`, `/api/b2/`, `/api/b3/`) have NO rate limiting currently.

---

## 8. File Storage

**Confidence: HIGH**

### 8.1 Storage Backends

| Backend | Service Key | Selection |
|---------|-------------|-----------|
| Local disk | `:local` | Default development |
| Amazon S3 | `:amazon` | `AWS_BUCKET` is set |
| DigitalOcean Spaces | `:digitalocean` | `ACTIVE_STORAGE_SERVICE=digitalocean` |
| S3-Compatible | `:s3_compatible` | `ACTIVE_STORAGE_SERVICE=s3_compatible` |
| Google Cloud Storage | `:google` | `ACTIVE_STORAGE_SERVICE=google` |

### 8.2 Selection Logic

```ruby
if ENV['AWS_BUCKET']
  config.active_storage.service = :amazon
else
  config.active_storage.service = ENV.fetch('ACTIVE_STORAGE_SERVICE', :local)
end
```

### 8.3 Image Processing

| Setting | Value |
|---------|-------|
| Processor | vips |
| Default resize | 1280x1280 max |
| Quality | 80-85 |
| Metadata | Stripped (EXIF removed) |

### 8.4 File Size Limits

**No application-level file size limits.** Relies on infrastructure:
- Web server `client_max_body_size`
- Cloud storage provider limits

### 8.5 Models with Attachments

| Model | Attachment | Type |
|-------|------------|------|
| User | `uploaded_avatar` | has_one_attached |
| Group | `cover_photo` | has_one_attached |
| Group | `logo` | has_one_attached |
| Document | `file` | has_one_attached |
| ReceivedEmail | `attachments` | has_many_attached |
| HasRichText models | `files`, `image_files` | has_many_attached |

---

## 9. Templates

**Confidence: MEDIUM**

### 9.1 Poll Types (9)

| Type | Vote Method | Description |
|------|-------------|-------------|
| `count` | show_thumbs | Opt-in/out tracking |
| `check` | show_thumbs | Sense check with three options |
| `question` | reason_only | Open-ended question (no options) |
| `proposal` | show_thumbs | Formal proposal with agree/disagree/block |
| `meeting` | time_poll | Date/time scheduling |
| `poll` | choose | Single or multiple choice |
| `dot_vote` | allocate | Distribute points across options |
| `score` | score | Rate options on a scale |
| `ranked_choice` | ranked_choice | Rank options by preference |

### 9.2 Poll Template Model

```ruby
# Key fields
poll_type              # proposal, poll, count, etc.
process_name           # Display name
process_subtitle       # Short description
process_introduction   # Rich text intro
details                # Default poll details
poll_options           # JSONB array of default options
default_duration_in_days
anonymous              # Hide voter names
hide_results           # 0=show, 1=until_vote, 2=until_closed
stance_reason_required # 0=disabled, 1=optional, 2=required
notify_on_closing_soon # 0=nobody, 1=author, 2=undecided, 3=voters
```

### 9.3 Poll Option Structure

```json
{
  "name": "Agree",
  "icon": "agree",
  "meaning": "I support this proposal",
  "prompt": "Why do you agree?",
  "color": "#70C9F8"
}
```

### 9.4 Discussion Template Model

```ruby
# Key fields
process_name           # Display name
process_subtitle       # Short description
process_introduction   # Rich text intro (format field)
title                  # Default discussion title
description            # Default description
tags                   # JSONB array of default tags
poll_template_keys_or_ids  # Associated poll templates
```

---

## 10. Demo System

**Confidence: MEDIUM**

### 10.1 Demo Architecture

Demos are pre-generated group clones stored in a Redis queue for instant access.

```
Demo model
    -> RecordCloner.create_clone_group(demo.group)
    -> Redis::List('demo_group_ids').push(group.id)
```

### 10.2 Demo Queue Management

| Operation | Method |
|-----------|--------|
| Fill queue | `DemoService.refill_queue` |
| Take demo | `DemoService.take_demo(actor)` |
| Ensure queue | `DemoService.ensure_queue` (hourly) |
| Generate public demos | `DemoService.generate_demo_groups` (daily) |

### 10.3 Demo Configuration

| Variable | Purpose |
|----------|---------|
| `FEATURES_DEMO_GROUPS` | Enable demo system |
| `FEATURES_DEMO_GROUPS_SIZE` | Queue size (default: 3) |

### 10.4 Demo Group Lifecycle

1. Admin creates Demo record pointing to source Group
2. `DemoService` clones group with all content (discussions, polls, comments, events)
3. Clone stored in Redis queue
4. User takes demo -> assigned ownership, subscription created
5. Expired demos deleted daily

---

## 11. Transcription Service

**Confidence: HIGH**

### 11.1 Integration

| Service | Model | Environment Variable |
|---------|-------|---------------------|
| OpenAI Whisper | whisper-1 | `OPENAI_API_KEY` |

### 11.2 Implementation

```ruby
TranscriptionService.available?
# Returns true if OPENAI_API_KEY is present

TranscriptionService.transcribe(file)
# Sends audio to OpenAI Whisper API
# Returns verbose_json response
```

### 11.3 Transcript Handling

Transcripts are appended to records via `AppendTranscriptWorker`:
- Processes audio file attachments
- Appends transcript text to record body/description
- Publishes real-time update

---

## 12. Group Export

**Confidence: HIGH**

### 12.1 Export Format (CSV)

```
Export for {group.full_name}

Groups (N)
id, key, name, description, created_at
...

Memberships (N)
group_id, user_id, user_name, user_email, admin, created_at, accepted_at
...

Discussions (N)
id, group_id, author_id, author_name, title, description, created_at
...

Comments (N)
id, group_id, discussion_id, author_id, author_name, title, author_name, body, created_at
...

Polls (N)
id, key, discussion_id, group_id, author_id, author_name, title, details, closing_at, closed_at, created_at, poll_type, custom_fields
...

Stances (N)
id, poll_id, participant_id, author_name, reason, latest, created_at, updated_at
...

Outcomes (N)
id, poll_id, author_id, statement, created_at, updated_at
...
```

### 12.2 Export Workers

| Worker | Purpose |
|--------|---------|
| `GroupExportWorker` | Full JSON/file export |
| `GroupExportCsvWorker` | CSV export |

---

## 13. Uncertainties

**Items requiring original author clarification or additional investigation:**

### HIGH Uncertainty

| Topic | Gap | Recommendation |
|-------|-----|----------------|
| Matrix chatbot protocol | No controller found, Redis pubsub only | Investigate external bot service |
| Hocuspocus collaborative editing | Configuration and protocol details unknown | Document Yjs/Hocuspocus integration |
| Socket.io server implementation | No source code in Rails app | External service documentation needed |

### MEDIUM Uncertainty

| Topic | Gap | Recommendation |
|-------|-----|----------------|
| Real-time payload format | Inferred from serializers, not captured live | Verify with actual traffic |
| Poll template JSON schema | Custom fields structure not fully documented | Create formal schema |
| Demo expiration timing | Exact TTL not found | Confirm with original authors |
| Task reminder scheduling | Logic in TaskService not fully traced | Complete investigation |

### LOW Uncertainty

| Topic | Gap | Recommendation |
|-------|-----|----------------|
| `login_emails` queue usage | Defined but no explicit worker assignment | Likely Devise integration |
| Paper Trail `members_can_add_guests` exclusion | Appears to be oversight | Add to tracking |
| `members_can_vote` flag | Deprecated but present in schema | Safe to remove |

---

## Appendix: Configuration Files

| File | Purpose |
|------|---------|
| `config/poll_types.yml` | Poll type definitions (9 types) |
| `config/poll_templates.yml` | Default poll templates |
| `config/discussion_templates.yml` | Default discussion templates |
| `config/webhook_event_kinds.yml` | Webhook-eligible events (14) |
| `config/providers.yml` | OAuth provider configuration |
| `config/locales/` | I18n translation files |

---

*Generated: 2026-02-01*
*Source: Loomio codebase reverse-engineering*
*Confidence levels: HIGH (direct code inspection), MEDIUM (inferred), LOW (uncertain)*
