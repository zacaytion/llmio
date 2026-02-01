# Expert Guide: Rails Patterns in Loomio

**Generated:** 2026-02-01
**Purpose:** Essential context for all swarm agents investigating the Loomio codebase

---

## Table of Contents

1. [Rails Conventions in This Codebase](#1-rails-conventions-in-this-codebase)
2. [Service Layer Patterns](#2-service-layer-patterns)
3. [Event Sourcing Implementation](#3-event-sourcing-implementation)
4. [Authorization Patterns](#4-authorization-patterns)
5. [Serializer Conventions](#5-serializer-conventions)
6. [Concern/Mixin Patterns](#6-concernmixin-patterns)
7. [Query Object Patterns](#7-query-object-patterns)
8. [Worker/Job Patterns](#8-workerjob-patterns)
9. [Frontend Integration Points](#9-frontend-integration-points)
10. [Key Cross-Cutting Concerns](#10-key-cross-cutting-concerns)

---

## 1. Rails Conventions in This Codebase

### Model Organization

Models are located in `/app/models/` and follow a layered structure:

- **Core models** (top-level): `user.rb`, `group.rb`, `discussion.rb`, `poll.rb`, `comment.rb`, `event.rb`, `stance.rb`, `membership.rb`
- **Null objects**: `null_discussion.rb`, `null_poll.rb`, `null_group.rb` - Used when a real record does not exist to avoid nil checks
- **STI subclasses**: Located in subdirectories like `events/` for Event subclasses and `ability/` for authorization modules
- **Concerns**: Located in `/app/models/concerns/` - shared behaviors mixed into models

Models heavily use concerns for shared functionality. A typical model like Poll includes many concerns:
- HasRichText (for body/details content)
- HasMentions (for @mention parsing)
- HasTags (for tagging)
- HasEvents (for activity tracking)
- HasCustomFields (for JSONB custom_fields column)
- Searchable (for full-text search)
- ReadableUnguessableUrls (for key-based URLs)

### Controller Organization

Controllers follow a namespace hierarchy:

- `/app/controllers/api/v1/` - Primary internal API for Vue SPA (39 controllers)
- `/app/controllers/api/b2/` - Bot/integration API v2
- `/app/controllers/api/b3/` - Bot/integration API v3 (users only)
- `/app/controllers/identities/` - OAuth provider controllers
- `/app/controllers/dev/` - Development/testing controllers
- Root controllers - Non-API controllers for pages and email actions

API controllers inherit from `Api::V1::SnorlaxBase` which provides:
- Standard CRUD actions (show, index, create, update, destroy)
- Automatic service delegation via naming convention
- Pagination, filtering, and timeframe scoping
- Serialization with RecordCache
- Error handling with standard responses

### Route to Controller Mapping

Routes are defined in `/config/routes.rb`. The primary pattern:

- Routes like `POST /api/v1/polls` map to `Api::V1::PollsController#create`
- The controller automatically delegates to `PollService.create`
- Custom actions are defined on collection or member routes

### Naming Conventions

- **Services**: `{Model}Service` (e.g., `PollService`, `DiscussionService`)
- **Serializers**: `{Model}Serializer` (e.g., `PollSerializer`)
- **Abilities**: `Ability::{Model}` modules in `app/models/ability/`
- **Queries**: `{Model}Query` (e.g., `PollQuery`, `DiscussionQuery`)
- **Workers**: `{Action}Worker` (e.g., `PublishEventWorker`, `GroupExportWorker`)
- **Events**: `Events::{ActionInPastTense}` (e.g., `Events::PollCreated`, `Events::NewComment`)

---

## 2. Service Layer Patterns

### Overview

All mutations in Loomio go through service classes in `/app/services/`. This is a strict architectural pattern - controllers never directly modify models.

### Standard Service Structure

Services are classes with class methods only (no instance state). The standard pattern:

1. **Method signature**: `def self.action(model:, actor:, params: {})` with keyword arguments
2. **Authorization first**: `actor.ability.authorize! :action, model`
3. **Validation check**: `return false unless model.valid?`
4. **Transaction wrapper**: `Model.transaction do ... end` for multi-step operations
5. **Model save**: `model.save!`
6. **EventBus broadcast**: `EventBus.broadcast('model_action', model, actor)`
7. **Event publish**: `Events::ActionEvent.publish!(model, actor, ...)`

### Example Pattern from PollService

The create action follows this sequence:
1. Authorize the actor can create the poll
2. Assign the author
3. Validate the poll
4. Save the poll within a transaction
5. Create stances for voters if applicable
6. Broadcast via EventBus for side effects
7. Publish a PollCreated event for notifications/activity

### Key Services and Responsibilities

| Service | Responsibility |
|---------|---------------|
| `PollService` | Poll lifecycle: create, update, close, reopen, remind, invite voters |
| `DiscussionService` | Discussion lifecycle: create, update, close, move, pin, mark as read |
| `GroupService` | Group lifecycle: create, update, destroy, invite members, merge |
| `MembershipService` | Membership management: redeem invitations, revoke, set volume, make admin |
| `CommentService` | Comment lifecycle: create, update, discard |
| `StanceService` | Voting: create stance (vote), update stance |
| `OutcomeService` | Poll outcomes: create and announce outcomes |
| `EventService` | Event manipulation: repair threads, reset positions |
| `NotificationService` | Notification management: mark as viewed |
| `AnnouncementService` | Notification sending: audience calculation, sending announcements |

### Service-to-Service Calls

Services call other services for cross-cutting operations. For example:
- `DiscussionService.add_users` calls `DiscussionReader.import` to add discussion readers
- `PollService.invite` calls `DiscussionService.add_users` when the poll is in a discussion
- `MembershipService.redeem` calls `PollService.group_members_added` to add voters to polls

### Return Values

Services typically return:
- An Event object for successful operations that create events
- `false` for validation failures
- Nothing explicit (nil) for operations that do not create events

---

## 3. Event Sourcing Implementation

### The Event Model

Located at `/app/models/event.rb`, the Event model is the backbone of activity tracking. Key characteristics:

- **STI (Single Table Inheritance)**: The `kind` column determines the event type
- **Polymorphic association**: `eventable` can be any model (Poll, Discussion, Comment, etc.)
- **Hierarchical structure**: Events have parent_id for threading (events within discussions)
- **Sequence tracking**: `sequence_id` and `position_key` for ordering within discussions

### Event Subclasses

Located in `/app/models/events/`, there are 42 event types. Examples:

**Discussion events:**
- `Events::NewDiscussion`
- `Events::DiscussionEdited`
- `Events::DiscussionClosed`
- `Events::DiscussionMoved`

**Poll events:**
- `Events::PollCreated`
- `Events::PollEdited`
- `Events::PollClosedByUser`
- `Events::PollExpired`
- `Events::PollAnnounced`
- `Events::PollReminder`

**Stance events:**
- `Events::StanceCreated`
- `Events::StanceUpdated`

**Membership events:**
- `Events::MembershipCreated`
- `Events::InvitationAccepted`
- `Events::UserJoinedGroup`

### Event Publishing

Events are published using the `Event.publish!` class method:

1. The event is built with attributes (eventable, user, discussion, etc.)
2. The event is saved to the database
3. `PublishEventWorker.perform_async(event.id)` enqueues background processing
4. The worker calls `event.trigger!` to execute side effects

### Event Concerns for Behavior

Event behavior is composed through concerns in `/app/models/concerns/events/`:

- **`Events::Notify::InApp`**: Creates in-app notifications for recipients
- **`Events::Notify::ByEmail`**: Sends email notifications via EventMailer
- **`Events::Notify::Mentions`**: Notifies mentioned users
- **`Events::Notify::Chatbots`**: Sends notifications to configured chatbots
- **`Events::Notify::Subscribers`**: Notifies users based on volume preferences
- **`Events::LiveUpdate`**: Publishes real-time updates via MessageChannelService

Each event class includes the appropriate concerns. For example, `Events::PollCreated` includes:
- `Events::LiveUpdate`
- `Events::Notify::Mentions`
- `Events::Notify::Chatbots`
- `Events::Notify::ByEmail`
- `Events::Notify::InApp`
- `Events::Notify::Subscribers`

### The Trigger Chain

When `event.trigger!` is called:
1. Each included concern's `trigger!` method is called via `super`
2. Concerns perform their side effects (create notifications, send emails, etc.)
3. `EventBus.broadcast("#{kind}_event", self)` is called for additional listeners

### The EventBus

Located at `/lib/event_bus.rb`, EventBus is a simple pub/sub system:

- **`EventBus.broadcast(event_name, *params)`**: Broadcasts an event to all listeners
- **`EventBus.listen(*events, &block)`**: Registers a listener for events
- **`EventBus.configure { |config| ... }`**: Configuration block for registering listeners

EventBus listeners are configured in `/config/initializers/event_bus.rb`. They handle:
- Updating DiscussionReader state when users comment or vote
- Publishing real-time updates when users mark discussions as read

### Recipient Calculation

Events store recipient information in custom_fields:
- `recipient_user_ids`: Explicitly specified recipients
- `recipient_chatbot_ids`: Chatbots to notify
- `recipient_audience`: Audience type (e.g., "group", "voters")
- `recipient_message`: Custom message for notifications

The `email_recipients` and `notification_recipients` methods use `Queries::UsersByVolumeQuery` to filter recipients based on their notification volume preferences.

---

## 4. Authorization Patterns

### CanCanCan Integration

Loomio uses CanCanCan for authorization with a modular ability system.

### The Ability Architecture

Located in `/app/models/ability/`:

- **`Ability::Base`**: The main ability class that includes all modules
- **Per-model modules**: `Ability::Poll`, `Ability::Group`, `Ability::Discussion`, etc.

The Base class uses `prepend` to include all ability modules:

```
module Ability
  class Base
    include CanCan::Ability
    prepend Ability::Comment
    prepend Ability::Discussion
    prepend Ability::Group
    prepend Ability::Poll
    # ... and more
  end
end
```

### User Ability Access

Users access their ability object through the `ability` method defined on User:

- `user.ability` returns an `Ability::Base` instance
- `user.can?(:action, resource)` checks permission
- `user.ability.authorize!(:action, resource)` raises an exception if not permitted

### Permission Check Patterns

Ability modules define permissions using CanCanCan's `can` blocks. Example from `Ability::Poll`:

- `:show` - Can view the poll (via PollQuery.visible_to check)
- `:create` - Can create a poll (group admin, or member if allowed, or standalone)
- `:update` - Can modify the poll (poll admin and poll not closed)
- `:destroy` - Can delete the poll (poll admin)
- `:close` - Can close the poll (poll admin and poll active)
- `:reopen` - Can reopen the poll (poll admin and poll closed)
- `:vote_in` - Can vote in the poll (logged in, poll active, is voter)
- `:announce` - Can invite/notify about the poll (group admin or poll admin with permission)

### Permission Factors

Permissions typically consider:
- User's role (admin, member, guest)
- Group settings (members_can_raise_motions, members_can_announce, etc.)
- Resource state (closed, archived, discarded)
- Relationship to resource (author, participant, inviter)

### Authorization in Controllers

Controllers authorize via services, not directly. The flow:

1. Controller receives request
2. Controller calls service: `PollService.create(poll: poll, actor: current_user)`
3. Service authorizes: `actor.ability.authorize! :create, poll`
4. If unauthorized, CanCan::AccessDenied is raised
5. Controller rescues the exception and returns 403

SnorlaxBase has: `rescue_from(CanCan::AccessDenied) { |e| respond_with_standard_error e, 403 }`

### Group Permission Settings

Groups have boolean settings that control member abilities:
- `members_can_add_members`
- `members_can_add_guests`
- `members_can_announce`
- `members_can_edit_discussions`
- `members_can_edit_comments`
- `members_can_delete_comments`
- `members_can_raise_motions`
- `members_can_start_discussions`
- `members_can_create_subgroups`

These settings are checked in ability definitions to determine if non-admin members can perform actions.

---

## 5. Serializer Conventions

### Serializer Framework

Loomio uses ActiveModelSerializers 0.8 (a legacy version). Serializers are located in `/app/serializers/`.

### Base Serializer

`ApplicationSerializer` provides common functionality:

- **Cache integration**: `cache_fetch(keys, id) { fallback }` method for RecordCache
- **Conditional includes**: `include_type?(type)` checks for excluded types
- **Common associations**: Pre-defined methods for author, group, discussion, poll
- **Discarded handling**: `hide_when_discarded` class method to nil out attributes on discarded records

### Serializer Structure

Serializers define:
- **`attributes`**: List of model attributes to serialize
- **`has_one`/`has_many`**: Associated records with their serializer and root
- **Custom methods**: Computed attributes that may use cache

Example from PollSerializer:
- Declares many attributes (id, title, closed_at, etc.)
- Declares associations (discussion, created_event, group, author, poll_options, my_stance)
- Has conditional includes (`include_results?`, `include_my_stance?`)
- Computes `results` by calling `PollService.calculate_results`

### RecordCache Integration

RecordCache (in `/app/services/record_cache.rb`) pre-loads associated records to avoid N+1 queries:

1. Controller builds collection
2. `RecordCache.for_collection(collection, user_id)` creates a cache
3. Cache pre-loads groups, users, memberships, polls, stances, etc.
4. Serializers use `cache_fetch` to retrieve pre-loaded records
5. If not in cache, the block is evaluated as fallback

### Serialization Scope

Serializers receive a scope hash containing:
- `cache`: RecordCache instance
- `current_user_id`: The requesting user's ID
- `exclude_types`: Types to exclude from serialization

The scope is built in SnorlaxBase's `default_scope` method.

### Response Structure

API responses follow a pattern where:
- The root key matches the resource type (e.g., "polls", "events")
- Associated records are embedded under their own root keys
- Meta information includes total count and root name

---

## 6. Concern/Mixin Patterns

### Overview

Concerns in `/app/models/concerns/` provide reusable behavior. They follow ActiveSupport::Concern conventions with `included` blocks and `ClassMethods` modules.

### Key Concerns

#### HasRichText

Location: `/app/models/concerns/has_rich_text.rb`

Provides:
- HTML sanitization with whitelist of allowed tags/attributes
- Format validation (html or md)
- File attachments via Active Storage (`has_many_attached :files`)
- Link preview handling
- Content locale detection
- Task parsing from rich text content
- Heading ID generation for anchor links

Models declare: `is_rich_text on: [:body]` or `is_rich_text on: [:details]`

#### HasEvents

Location: `/app/models/concerns/has_events.rb`

Provides:
- `has_many :events` association
- `has_many :notifications, through: :events`
- `has_many :users_notified, through: :notifications`

This is included in any model that generates events (Discussion, Poll, Comment, etc.)

#### HasMentions

Location: `/app/models/concerns/has_mentions.rb`

Provides:
- Username extraction from text (via twitter-text gem)
- User ID extraction from HTML spans with data-mention-id
- `mentioned_users` and `mentioned_groups` methods
- `newly_mentioned_users` to avoid re-notifying on edits

Models declare: `is_mentionable on: [:body]`

#### Searchable

Location: `/app/models/concerns/searchable.rb`

Integrates pg_search for full-text search:
- Calls `multisearchable` to enable pg_search
- Models must implement `pg_search_insert_statement` class method
- Provides `rebuild_pg_search_documents` for reindexing

#### HasVolume

Used on Membership, DiscussionReader, and Stance to track notification volume preferences. Volumes are: mute, quiet, normal, loud.

#### HasCustomFields

Location: `/app/models/concerns/has_custom_fields.rb`

Provides dynamic accessor methods for JSONB custom_fields column. Used on Event and other models to store flexible data.

#### ReadableUnguessableUrls

Generates secure, unguessable keys for URL identification. Used on Discussion, Poll, Group, etc.

### Event Concerns

Event behavior concerns in `/app/models/concerns/events/`:

- **`Events::LiveUpdate`**: Publishes to MessageChannelService for real-time updates
- **`Events::Notify::InApp`**: Creates Notification records and publishes to user channels
- **`Events::Notify::ByEmail`**: Enqueues EventMailer jobs for email notifications
- **`Events::Notify::Mentions`**: Handles @mention notifications
- **`Events::Notify::Chatbots`**: Sends to configured webhook/chatbot integrations
- **`Events::Notify::Subscribers`**: Notifies based on volume preferences

These concerns override `trigger!` and call `super` to chain behavior.

---

## 7. Query Object Patterns

### Overview

Query objects in `/app/queries/` encapsulate complex query logic. They are classes with class methods that return ActiveRecord relations.

### Standard Structure

Query objects follow a pattern:
- `start` method: Returns base relation with includes and default scope
- `visible_to(user:, ...)` method: Applies visibility/authorization scoping
- `filter(chain:, params:)` method: Applies filter parameters

### PollQuery Example

Location: `/app/queries/poll_query.rb`

- `start`: Returns `Poll.distinct.kept.includes(:poll_options, :group, :author)`
- `visible_to`: Joins memberships, discussion_readers, and stances to check access
- `filter`: Applies filters like group_key, discussion_key, tags, status, author_id

The visibility check uses LEFT OUTER JOINs to check multiple access paths:
1. User is the poll author
2. Group allows public access
3. User has an active membership in the group
4. User has guest access via DiscussionReader
5. User has guest access via Stance

### DiscussionQuery Example

Location: `/app/queries/discussion_query.rb`

- `start`: Returns discussions with group join and author include
- `dashboard`: Filters to user's group discussions and guest discussions
- `inbox`: Filters to unread/undismissed discussions
- `visible_to`: Complex visibility with public/private and subgroup considerations
- `filter`: Applies open/closed filtering and ordering

### UsersByVolumeQuery

Location: `/app/extras/queries/users_by_volume_query.rb`

Specialized query for notification recipients:
- Joins discussion_readers, memberships, and stances
- Filters by volume level (mute, quiet, normal, loud)
- Used by events to determine who should receive notifications

### Controller Usage

Controllers call query objects via `visible_records` methods:

```ruby
def visible_records
  PollQuery.visible_to(user: current_user)
end
```

The base controller's `instantiate_collection` method chains:
1. `accessible_records` (calls visible_records or public_records)
2. `timeframe_collection` (applies since/until filtering)
3. Pagination (offset/limit)
4. Ordering

---

## 8. Worker/Job Patterns

### Overview

Workers in `/app/workers/` are Sidekiq jobs for background processing. They include `Sidekiq::Worker`.

### PublishEventWorker

Location: `/app/workers/publish_event_worker.rb`

The most important worker - called after every event is created:
- Receives event_id
- Finds the event using STI-aware lookup: `Event.sti_find(event_id)`
- Calls `event.trigger!` to execute all side effects

### GenericWorker

A general-purpose worker that can call any class method:
- Called as: `GenericWorker.perform_async('ClassName', 'method_name', arg1, arg2, ...)`
- Used for deferred service calls like `GenericWorker.perform_async('PollService', 'group_members_added', group_id)`

### Export Workers

- `GroupExportWorker`: Generates group data exports as JSON
- `GroupExportCsvWorker`: Generates CSV exports

### Scheduled Workers

Some workers run on schedules (configured via sidekiq-cron or similar):
- `CloseExpiredPollWorker`: Closes polls past their closing_at time
- Poll closing soon notifications

### Worker Patterns

Workers typically:
1. Accept IDs rather than objects (for serialization)
2. Load records fresh from database
3. Handle errors gracefully (Sidekiq will retry)
4. Perform single, focused operations

### Background Job Triggers

Jobs are triggered by:
- Service methods calling `Worker.perform_async(...)` or `Worker.perform_in(...)`
- EventBus listeners
- Scheduled tasks
- Email processing

---

## 9. Frontend Integration Points

### API Request/Response Contract

The Vue frontend communicates with Rails via JSON API:

**Request:**
- HTTP method determines action (GET=index/show, POST=create, PATCH=update, DELETE=destroy)
- Parameters sent as JSON body or query params
- Permitted params defined in `/app/models/permitted_params.rb`

**Response:**
- JSON with root key matching resource type
- Associated records embedded under their own roots
- Meta object with count and pagination info
- Events returned for create/update operations (contains the modified record)

### Boot Endpoint

`/api/v1/boot/user` returns:
- Current user data
- User's memberships
- User's groups
- Notifications
- Configuration data

This bootstraps the frontend state on initial load.

### Real-time Updates

Real-time updates use Redis pub/sub:

1. Server publishes to Redis via `MessageChannelService.publish_models(...)`
2. External channels service (configured via CHANNELS_URL) subscribes to Redis
3. Channels service pushes to clients via WebSocket/SSE
4. Client updates LokiJS store

Channels can be:
- `user-{id}`: Personal notifications
- `group-{id}`: Group activity

### LokiJS Record Store

The frontend uses LokiJS as an in-memory document store:
- Located in `/vue/src/shared/services/records.js`
- Interfaces in `/vue/src/shared/interfaces/` mirror Rails models
- Records are imported from API responses
- Relationships computed via ID references

### Authentication

Authentication uses Devise with session cookies:
- Session created on sign in
- CSRF token required for state-changing requests
- Token-based access for email links and API keys

### Token-Based Access

Some access uses tokens instead of sessions:
- `unsubscribe_token`: For email unsubscribe links
- `membership_token`: For invitation acceptance
- `discussion_reader_token`: For guest access to discussions
- `stance_token`: For guest access to polls

---

## 10. Key Cross-Cutting Concerns

### Paper Trail Versioning

Models with audit requirements include Paper Trail:
- `Discussion`, `Poll`, `Comment`, `Outcome`, `User`
- Configuration: `has_paper_trail only: [attributes]`
- Versions stored in `versions` table with JSONB object_changes
- Events link to versions via `eventable_version_id`
- Used for edit history display

### Soft Delete (Discard)

Many models use soft delete via the `discard` gem:
- Column: `discarded_at` timestamp
- Column: `discarded_by` user ID
- Scope: `.kept` excludes discarded records
- Services call `model.update(discarded_at: Time.now, discarded_by: actor.id)`

Discarded records:
- Are excluded from queries by default
- Have their events hidden from timelines
- Have sensitive attributes hidden in serialization

### Timestamps and Ordering

Important timestamp patterns:
- `created_at`, `updated_at`: Standard Rails
- `last_activity_at`: Discussion's latest activity time
- `closed_at`: When poll/discussion was closed
- `accepted_at`: When membership was accepted
- `revoked_at`: When access was revoked
- `discarded_at`: When record was soft deleted

Ordering patterns:
- Discussions: by pinned status, then last_activity_at
- Events: by position_key for threaded display
- Polls: by created_at or closing_at

### Position and Sequence Handling

Events have complex positioning:
- `sequence_id`: Global sequence within a discussion (for read tracking)
- `position`: Position among siblings
- `position_key`: Hierarchical position string for sorting (e.g., "00001-00003-00002")

`SequenceService` manages atomic sequence generation using a partition_sequences table to avoid race conditions.

### Volume Preferences

Notification volume is a key concept:
- Levels: mute (0), quiet (1), normal (2), loud (3)
- Stored on: Membership, DiscussionReader, Stance
- Cascades: Stance > DiscussionReader > Membership > User default
- Controls: What notifications/emails a user receives

### Guest Access

Guests are users with access to specific content without group membership:
- DiscussionReader with `guest: true` for discussion guests
- Stance with `guest: true` for poll guests
- Guests can be invited by email
- Guests have limited permissions compared to members

### Announcement Audiences

When inviting/notifying, several audiences are available:
- `group`: All group members
- `voters`: All poll voters
- `undecided`: Voters who haven't voted
- `decided`: Voters who have voted
- `non_voters`: Group members not yet invited to vote

---

## Summary for Swarm Agents

When investigating any domain:

1. **Start with the service**: Find the `{Domain}Service` in `/app/services/`
2. **Check the model**: Look at the model in `/app/models/` for associations and concerns
3. **Examine abilities**: Look at `/app/models/ability/{domain}.rb` for permissions
4. **Review events**: Check `/app/models/events/` for domain-specific events
5. **Check queries**: Look at `/app/queries/{domain}_query.rb` for visibility logic
6. **Review serializers**: Check `/app/serializers/{domain}_serializer.rb` for API output
7. **Check workers**: Look for domain-specific workers in `/app/workers/`
8. **Review specs**: Check `/spec/services/` and `/spec/models/` for behavior documentation

All domains share common patterns:
- Services for mutations
- Ability modules for authorization
- Events for activity tracking
- Query objects for visibility
- Serializers for API output
- RecordCache for performance
