# Webhook Events: Model Documentation

## Core Models

### Event (Base Class)

**File**: `/Users/z/Code/loomio/app/models/event.rb`

```ruby
class Event < ApplicationRecord
  belongs_to :eventable, polymorphic: true
  belongs_to :discussion, required: false
  belongs_to :user, required: false

  set_custom_fields :pinned_title,
                    :recipient_user_ids,
                    :recipient_chatbot_ids,  # <-- Webhook targeting
                    :recipient_message,
                    :recipient_audience,
                    :stance_ids

  # Called after create via PublishEventWorker
  def trigger!
    EventBus.broadcast("#{kind}_event", self)
  end
end
```

Key webhook-related fields:
- `kind` (string) - Event type identifier (e.g., "new_discussion", "poll_created")
- `recipient_chatbot_ids` (custom field, array) - Explicit chatbot targeting
- `eventable_type` (string) - Polymorphic type for template resolution

### Chatbot Model

**File**: `/Users/z/Code/loomio/app/models/chatbot.rb`

```ruby
class Chatbot < ApplicationRecord
  belongs_to :group
  belongs_to :author, class_name: 'User'

  validates_presence_of :server
  validates_presence_of :name
  validates_inclusion_of :kind, in: ['matrix', 'webhook']
  validates_inclusion_of :webhook_kind, in: ['slack', 'microsoft', 'discord', 'markdown', 'webex', nil]
end
```

**Database Schema** (from `db/schema.rb:156-170`):

| Column | Type | Description |
|--------|------|-------------|
| `id` | integer | Primary key |
| `kind` | string | 'matrix' or 'webhook' |
| `server` | string | Webhook URL or Matrix server |
| `channel` | string | Matrix room ID (nullable) |
| `access_token` | string | Matrix access token (nullable) |
| `author_id` | integer | User who created the integration |
| `group_id` | integer | Group this chatbot belongs to |
| `name` | string | Display name for the integration |
| `notification_only` | boolean | Use minimal notification template |
| `webhook_kind` | string | 'slack', 'microsoft', 'discord', 'markdown', 'webex' |
| `event_kinds` | string[] | Array of subscribed event kinds |

### Events::Notify::Chatbots Concern

**File**: `/Users/z/Code/loomio/app/models/concerns/events/notify/chatbots.rb`

```ruby
module Events::Notify::Chatbots
  def trigger!
    super
    GenericWorker.perform_async('ChatbotService', 'publish_event!', id)
  end
end
```

This concern:
1. Calls `super` to trigger normal event processing
2. Schedules async chatbot notification via Sidekiq

## Event Subclasses with Webhook Capability

### Discussion Events

#### Events::NewDiscussion

**File**: `/Users/z/Code/loomio/app/models/events/new_discussion.rb`

```ruby
class Events::NewDiscussion < Event
  include Events::LiveUpdate
  include Events::Notify::InApp
  include Events::Notify::ByEmail
  include Events::Notify::Mentions
  include Events::Notify::Subscribers
  include Events::Notify::Chatbots  # <-- Webhook capable

  def self.publish!(
    discussion:,
    recipient_user_ids: [],
    recipient_chatbot_ids: [],  # <-- Explicit chatbot targeting
    recipient_audience: nil)
    # ...
  end
end
```

#### Events::DiscussionEdited

**File**: `/Users/z/Code/loomio/app/models/events/discussion_edited.rb`

```ruby
class Events::DiscussionEdited < Event
  include Events::LiveUpdate
  include Events::Notify::InApp
  include Events::Notify::ByEmail
  include Events::Notify::Mentions
  include Events::Notify::Chatbots

  def self.publish!(
    discussion:,
    actor:,
    recipient_user_ids: [],
    recipient_chatbot_ids: [],
    recipient_audience: nil,
    recipient_message: nil)
    # ...
  end
end
```

#### Events::DiscussionAnnounced (UI-Hidden)

**File**: `/Users/z/Code/loomio/app/models/events/discussion_announced.rb`

```ruby
class Events::DiscussionAnnounced < Event
  include Events::Notify::InApp
  include Events::Notify::ByEmail
  include Events::Notify::Chatbots

  def self.publish!(
    discussion:,
    actor:,
    recipient_user_ids:,
    recipient_chatbot_ids:,  # <-- Only way to trigger chatbots
    recipient_audience: nil,
    recipient_message: nil)
    # ...
  end
end
```

### Comment Events

#### Events::NewComment

**File**: `/Users/z/Code/loomio/app/models/events/new_comment.rb`

```ruby
class Events::NewComment < Event
  include Events::Notify::ByEmail
  include Events::Notify::Mentions
  include Events::Notify::Chatbots
  include Events::Notify::Subscribers
  include Events::LiveUpdate

  def self.publish!(comment)
    # No explicit chatbot targeting - relies on event_kinds subscription
  end
end
```

### Poll Events

#### Events::PollCreated

**File**: `/Users/z/Code/loomio/app/models/events/poll_created.rb`

```ruby
class Events::PollCreated < Event
  include Events::LiveUpdate
  include Events::Notify::Mentions
  include Events::Notify::Chatbots
  include Events::Notify::ByEmail
  include Events::Notify::InApp
  include Events::Notify::Subscribers

  def self.publish!(poll, actor, recipient_user_ids: [])
    # No explicit chatbot targeting - relies on event_kinds subscription
  end
end
```

#### Events::PollEdited

**File**: `/Users/z/Code/loomio/app/models/events/poll_edited.rb`

Includes `recipient_chatbot_ids` parameter for explicit targeting.

#### Events::PollClosingSoon

**File**: `/Users/z/Code/loomio/app/models/events/poll_closing_soon.rb`

No explicit chatbot targeting - relies on event_kinds subscription.

#### Events::PollExpired

**File**: `/Users/z/Code/loomio/app/models/events/poll_expired.rb`

No explicit chatbot targeting - relies on event_kinds subscription.

#### Events::PollClosedByUser

**File**: `/Users/z/Code/loomio/app/models/events/poll_closed_by_user.rb`

No explicit chatbot targeting - relies on event_kinds subscription.

#### Events::PollReopened

**File**: `/Users/z/Code/loomio/app/models/events/poll_reopened.rb`

```ruby
class Events::PollReopened < Event
  include Events::Notify::Chatbots
  # Note: Only includes Chatbots concern - no InApp, ByEmail, etc.
end
```

#### Events::PollAnnounced (UI-Hidden)

**File**: `/Users/z/Code/loomio/app/models/events/poll_announced.rb`

Includes `recipient_chatbot_ids` parameter for explicit targeting only.

#### Events::PollReminder (UI-Hidden)

**File**: `/Users/z/Code/loomio/app/models/events/poll_reminder.rb`

Includes `recipient_chatbot_ids` parameter for explicit targeting.

### Stance Events

#### Events::StanceCreated

**File**: `/Users/z/Code/loomio/app/models/events/stance_created.rb`

```ruby
class Events::StanceCreated < Event
  include Events::LiveUpdate
  include Events::Notify::InApp
  include Events::Notify::Mentions
  include Events::Notify::Chatbots
  include Events::Notify::Subscribers

  def self.publish!(stance)
    # No explicit chatbot targeting - relies on event_kinds subscription
  end
end
```

#### Events::StanceUpdated

**File**: `/Users/z/Code/loomio/app/models/events/stance_updated.rb`

```ruby
class Events::StanceUpdated < Events::StanceCreated
  # Inherits all behavior including Chatbots concern
end
```

### Outcome Events

#### Events::OutcomeCreated

**File**: `/Users/z/Code/loomio/app/models/events/outcome_created.rb`

Includes `recipient_chatbot_ids` parameter for explicit targeting.

#### Events::OutcomeUpdated

**File**: `/Users/z/Code/loomio/app/models/events/outcome_updated.rb`

Includes `recipient_chatbot_ids` parameter for explicit targeting.

#### Events::OutcomeReviewDue

**File**: `/Users/z/Code/loomio/app/models/events/outcome_review_due.rb`

No explicit chatbot targeting - relies on event_kinds subscription.

## Events WITHOUT Webhook Capability

These events do NOT include `Events::Notify::Chatbots`:

| Event | File | Notification Methods |
|-------|------|---------------------|
| `user_added_to_group` | `app/models/events/user_added_to_group.rb` | InApp, ByEmail |
| `membership_requested` | `app/models/events/membership_requested.rb` | InApp, ByEmail |
| `user_mentioned` | `app/models/events/user_mentioned.rb` | InApp, ByEmail |
| `comment_replied_to` | `app/models/events/comment_replied_to.rb` | InApp, ByEmail |
| `reaction_created` | `app/models/events/reaction_created.rb` | InApp |
| `invitation_accepted` | `app/models/events/invitation_accepted.rb` | InApp |
| `new_coordinator` | `app/models/events/new_coordinator.rb` | InApp, ByEmail |
| `new_delegate` | `app/models/events/new_delegate.rb` | InApp, ByEmail |

## Webhook Serializers

### Base: Webhook::Markdown::EventSerializer

**File**: `/Users/z/Code/loomio/app/serializers/webhook/markdown/event_serializer.rb`

```ruby
class Webhook::Markdown::EventSerializer < ActiveModel::Serializer
  attributes :text, :icon_url, :username

  def icon_url
    # Group logo URL
  end

  def username
    AppConfig.theme[:site_name]
  end

  def text
    # Renders chatbot/markdown/{template_name}.text.erb
  end
end
```

### Webhook::Slack::EventSerializer

**File**: `/Users/z/Code/loomio/app/serializers/webhook/slack/event_serializer.rb`

Extends Markdown, uses Slack-specific templates.

### Webhook::Discord::EventSerializer

**File**: `/Users/z/Code/loomio/app/serializers/webhook/discord/event_serializer.rb`

Extends Markdown, adds `content` attribute truncated to 1900 chars.

### Webhook::Microsoft::EventSerializer

**File**: `/Users/z/Code/loomio/app/serializers/webhook/microsoft/event_serializer.rb`

Uses Microsoft MessageCard format with `@type`, `@context`, `themeColor`.

### Webhook::Webex::EventSerializer

**File**: `/Users/z/Code/loomio/app/serializers/webhook/webex/event_serializer.rb`

Extends Markdown, adds `markdown` attribute.

## Model Relationships Diagram

```
Group
  |-- has_many :chatbots
  |-- has_many :discussions
  |-- has_many :polls

Chatbot
  |-- belongs_to :group
  |-- belongs_to :author (User)
  |-- event_kinds: string[] (subscribed event types)

Event
  |-- belongs_to :eventable (polymorphic)
  |-- recipient_chatbot_ids (custom field, for explicit targeting)

Discussion/Poll/Comment/Stance/Outcome
  |-- has_one :created_event (via eventable)
  |-- group (delegated for chatbot lookup)
```

## Webhook Delivery Flow

1. Event published via `Events::XYZ.publish!`
2. Event saved to database
3. `PublishEventWorker.perform_async(event.id)` scheduled
4. Worker calls `event.trigger!`
5. If event includes Chatbots concern, schedules `ChatbotService.publish_event!`
6. ChatbotService finds matching chatbots via:
   - `event.recipient_chatbot_ids`, OR
   - `chatbot.event_kinds.include?(event.kind)`
7. For each matching chatbot, serializes and POSTs payload
