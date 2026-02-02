# Event Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/event.rb`
- `/app/models/events/*.rb` (42 subclasses)
- `/app/models/concerns/events/*.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Event model is the core of Loomio's activity tracking and notification system. Events use Single Table Inheritance (STI) to represent 42 different event types. Events drive the discussion timeline, notifications, emails, and real-time updates.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `kind` | string(255) | - | YES | presence required | Event type (STI discriminator) |
| `eventable_id` | integer | - | YES | Polymorphic FK | Source record ID |
| `eventable_type` | string(255) | - | YES | Polymorphic type | Source record class |
| `eventable_version_id` | integer | - | YES | FK to versions | Paper Trail version |

### Actor & Context

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `user_id` | integer | - | YES | FK to users | User who triggered event |
| `discussion_id` | integer | - | YES | FK to discussions | Parent discussion (for timeline) |
| `parent_id` | integer | - | YES | FK to events | Parent event (for threading) |

### Thread Position

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `sequence_id` | integer | - | YES | Sequence within discussion |
| `position` | integer | 0 | NO | Position within parent |
| `position_key` | string | - | YES | Hierarchical position key |
| `depth` | integer | 0 | NO | Nesting depth |

### Counters

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `child_count` | integer | 0 | Direct child events |
| `descendant_count` | integer | 0 | Total descendant events |

### Flags

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `announcement` | boolean | false | Was announced to users |
| `pinned` | boolean | false | Pinned to top of thread |

### Custom Data

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `custom_fields` | jsonb | {} | Event-specific data |

**custom_fields structure:**
```json
{
  "pinned_title": "Important announcement",
  "recipient_user_ids": [1, 2, 3],
  "recipient_chatbot_ids": [10, 20],
  "recipient_message": "Custom notification message",
  "recipient_audience": "group",
  "stance_ids": [100, 101, 102]
}
```

### Timestamps

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `created_at` | datetime | - | Creation timestamp |
| `updated_at` | datetime | - | Last update |

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `kind` | presence required | always |
| `eventable` | presence required | always |

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `eventable` | Polymorphic | polymorphic: true | Source record |
| `discussion` | Discussion | required: false | Parent discussion |
| `user` | User | required: false | Triggering user |
| `parent` | Event | required: false | Parent event |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `notifications` | Notification | dependent: :destroy | Generated notifications |
| `children` | Event | `-> { where('discussion_id is not null') }`, foreign_key: :parent_id | Child events |

---

## Scopes

```ruby
scope :dangling, -> {
  joins('left join discussions d on events.discussion_id = d.id')
    .where('d.id is null and discussion_id is not null')
}

scope :unreadable, -> { where.not(kind: 'discussion_closed') }

scope :invitations_in_period, ->(since, till) {
  where(kind: :announcement_created, eventable_type: 'Group')
    .within(since.beginning_of_hour, till.beginning_of_hour)
}
```

---

## Callbacks

### Before Create (if discussion_id)
- `set_parent_and_depth` - Determines parent event and depth
- `set_sequences` - Assigns sequence_id, position, position_key

### After Rollback / Before Destroy (if discussion_id)
- `reset_sequences` - Cleans up PostgreSQL sequences

### After Create / After Destroy (if discussion_id)
- `update_sequence_info!` - Updates discussion's sequence tracking

---

## Instance Methods

### Event Publishing

```ruby
def self.publish!(eventable, **args)
  event = build(eventable, **args)
  event.save!
  PublishEventWorker.perform_async(event.id)
  event
end

def self.build(eventable, **args)
  new({
    kind: name.demodulize.underscore,
    eventable: eventable,
    eventable_version_id: ((eventable.respond_to?(:versions) && eventable.versions.last&.id) || nil)
  }.merge(args))
end

def trigger!
  # Called by PublishEventWorker after save
  # Broadcasts to EventBus for concern handlers
  EventBus.broadcast("#{kind}_event", self)
end
```

### Actor Methods

```ruby
def user
  super || AnonymousUser.new
end

def real_user
  user
end

def actor
  user
end

def actor_id
  user_id
end
```

### Messaging

```ruby
def message_channel
  eventable.group.message_channel
end
```

### STI Methods

```ruby
def self.sti_find(id)
  e = self.find(id)
  e.kind_class.find(id)
end

def kind_class
  ("Events::" + kind.classify).constantize
end

def active_model_serializer
  "Events::#{eventable.class.to_s.split('::').last}Serializer".constantize
rescue NameError
  EventSerializer
end
```

### Parent Event Resolution

```ruby
def find_parent_event
  case kind
  when 'discussion_closed'   then eventable.created_event
  when 'discussion_forked'   then eventable.created_event
  when 'discussion_moved'    then discussion.created_event
  when 'discussion_edited'   then (eventable || discussion)&.created_event
  when 'discussion_reopened' then eventable.created_event
  when 'outcome_created'     then eventable.parent_event
  when 'new_comment'         then eventable.parent_event
  when 'poll_closed_by_user' then eventable.created_event
  when 'poll_closing_soon'   then eventable.created_event
  when 'poll_created'        then eventable.parent_event
  when 'poll_edited'         then eventable.created_event
  when 'poll_expired'        then eventable.created_event
  when 'poll_option_added'   then eventable.created_event
  when 'poll_reopened'       then eventable.created_event
  when 'stance_created'      then eventable.parent_event
  when 'stance_updated'      then eventable.parent_event
  else
    nil
  end
end

def max_depth_adjusted_parent
  original_parent = find_parent_event
  return nil unless original_parent
  if discussion && discussion.max_depth == original_parent.depth
    original_parent.parent  # Move up one level to respect max_depth
  else
    original_parent
  end
end
```

### Sequence Management

```ruby
def set_parent_and_depth
  self.parent = max_depth_adjusted_parent
  self.depth = parent ? parent.depth + 1 : 0
end

def set_sequences
  self.sequence_id = next_sequence_id!
  self.position = next_position!
  self.position_key = [parent&.position_key, Event.zero_fill(position)].compact.join('-')
end

def next_sequence_id!
  # Uses PostgreSQL sequences via SequenceService
  unless SequenceService.seq_present?('discussions_sequence_id', discussion_id)
    val = Event.where(discussion_id: discussion_id)
               .where("sequence_id is not null")
               .order(sequence_id: :desc)
               .limit(1).pluck(:sequence_id).last || 0
    SequenceService.create_seq!('discussions_sequence_id', discussion_id, val)
  end
  SequenceService.next_seq!('discussions_sequence_id', discussion_id)
end

def next_position!
  return 0 unless (discussion_id && parent_id)
  unless SequenceService.seq_present?('events_position', parent_id)
    val = Event.where(parent_id: parent_id, discussion_id: discussion_id)
               .order(position: :desc)
               .limit(1).pluck(:position).last || 0
    SequenceService.create_seq!('events_position', parent_id, val)
  end
  SequenceService.next_seq!('events_position', parent_id)
end

def self.zero_fill(num)
  "0" * (5 - num.to_s.length) + num.to_s
end
```

### Recipient Methods

```ruby
def email_recipients
  Queries::UsersByVolumeQuery.email_notifications(eventable)
    .where(id: all_recipient_user_ids)
end

def notification_recipients
  Queries::UsersByVolumeQuery.app_notifications(eventable)
    .where(id: all_recipient_user_ids)
    .where.not(id: user.id || 0)
end

def all_recipients
  User.active.where(id: all_recipient_user_ids)
end

def all_recipient_user_ids
  (recipient_user_ids || []).uniq.compact
end
```

### Delegation

```ruby
delegate :group, to: :eventable, allow_nil: true
delegate :poll, to: :eventable, allow_nil: true
delegate :groups, to: :eventable, allow_nil: true
delegate :update_sequence_info!, to: :discussion, allow_nil: true
```

---

## Counter Cache Definitions

```ruby
define_counter_cache(:child_count) { |e| e.children.count }
define_counter_cache(:descendant_count) { |e|
  if e.kind == "new_discussion"
    Event.where(discussion_id: e.eventable_id).count
  elsif e.position_key && e.discussion_id
    Event.where(discussion_id: e.discussion_id)
         .where("id != ?", e.id)
         .where('position_key like ?', e.position_key + "%").count
  else
    0
  end
}

update_counter_cache :parent, :child_count
update_counter_cache :parent, :descendant_count
```

---

## Event STI Hierarchy (42 Subclasses)

### Discussion Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `new_discussion` | Discussion | Discussion created |
| `discussion_edited` | Discussion | Discussion edited |
| `discussion_title_edited` | Discussion | Title changed |
| `discussion_description_edited` | Discussion | Description changed |
| `discussion_closed` | Discussion | Discussion closed |
| `discussion_reopened` | Discussion | Discussion reopened |
| `discussion_moved` | Discussion | Moved to different group |
| `discussion_forked` | Discussion | Forked from another discussion |
| `discussion_announced` | Discussion | Announcement sent |

### Comment Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `new_comment` | Comment | Comment created |
| `comment_edited` | Comment | Comment edited |
| `comment_replied_to` | Comment | Reply notification |

### Poll Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `poll_created` | Poll | Poll created |
| `poll_edited` | Poll | Poll edited |
| `poll_announced` | Poll | Poll announcement sent |
| `poll_closing_soon` | Poll | Closing reminder |
| `poll_expired` | Poll | Poll closed automatically |
| `poll_closed_by_user` | Poll | Poll closed manually |
| `poll_reopened` | Poll | Poll reopened |
| `poll_option_added` | Poll | Voter added option |
| `poll_reminder` | Poll | Manual reminder sent |

### Stance Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `stance_created` | Stance | Vote submitted |
| `stance_updated` | Stance | Vote revised |

### Outcome Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `outcome_created` | Outcome | Outcome published |
| `outcome_updated` | Outcome | Outcome edited |
| `outcome_announced` | Outcome | Outcome announced |
| `outcome_review_due` | Outcome | Review reminder |

### Membership Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `membership_created` | Membership | Membership created |
| `membership_requested` | MembershipRequest | Join request |
| `membership_request_approved` | Membership | Request approved |
| `membership_resent` | Membership | Invitation resent |
| `invitation_accepted` | Membership | Invitation accepted |
| `user_added_to_group` | Membership | Admin added user |
| `user_joined_group` | Membership | User joined |

### User Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `user_mentioned` | Various | @mention notification |
| `group_mentioned` | Various | @group mention |
| `user_reactivated` | User | Account reactivated |
| `new_coordinator` | Membership | Made admin |
| `new_delegate` | Membership | Made delegate |

### Reaction Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `reaction_created` | Reaction | Emoji reaction added |

### Other Events

| Kind | Eventable | Description |
|------|-----------|-------------|
| `announcement_resend` | Various | Announcement resent |
| `unknown_sender` | ReceivedEmail | Email from unknown sender |

---

## Event Concerns

### Events::LiveUpdate

Sends real-time updates to connected clients.

```ruby
module Events::LiveUpdate
  def trigger!
    super
    notify_clients!
  end

  def notify_clients!
    return unless eventable
    if eventable.group_id
      MessageChannelService.publish_models([self], group_id: eventable.group.id)
    end
    if eventable.respond_to?(:guests)
      eventable.guests.find_each do |user|
        MessageChannelService.publish_models([self], user_id: user.id)
      end
    end
  end
end
```

### Events::Notify::InApp

Generates in-app notifications.

```ruby
module Events::Notify::InApp
  def trigger!
    super
    notify_users!
  end

  def notify_users!
    notifications.import(built_notifications)
    built_notifications.each { |n|
      MessageChannelService.publish_models(Array(n), user_id: n.user_id)
    }
  end
end
```

### Events::Notify::ByEmail

Sends email notifications.

```ruby
module Events::Notify::ByEmail
  def trigger!
    super
    email_users!
  end

  def email_users!
    email_recipients.active.no_spam_complaints.uniq.pluck(:id).each do |recipient_id|
      EventMailer.event(recipient_id, self.id).deliver_later
    end
  end
end
```

### Events::Notify::Mentions

Handles @mention notifications.

```ruby
module Events::Notify::Mentions
  def trigger!
    super
    return if silence_mentions?
    notify_mentioned_groups!
    notify_mentioned_users!
  end

  def notify_mentioned_users!
    return if eventable.newly_mentioned_users.empty?
    Events::UserMentioned.publish!(eventable, user, eventable.newly_mentioned_users.pluck(:id))
  end

  def notify_mentioned_groups!
    return if eventable.newly_mentioned_groups.empty?
    Events::GroupMentioned.publish!(eventable, user, eventable.newly_mentioned_groups.pluck(:id), id)
  end
end
```

### Events::Notify::Subscribers

Notifies subscribed users.

### Events::Notify::Chatbots

Sends webhook/chatbot notifications.

### Events::Notify::Author

Notifies the content author.

---

## Example Event Subclass

```ruby
class Events::NewDiscussion < Event
  include Events::LiveUpdate
  include Events::Notify::InApp
  include Events::Notify::ByEmail
  include Events::Notify::Mentions
  include Events::Notify::Subscribers
  include Events::Notify::Chatbots

  def self.publish!(
    discussion:,
    recipient_user_ids: [],
    recipient_chatbot_ids: [],
    recipient_audience: nil)

    super(discussion,
          user: discussion.author,
          recipient_user_ids: recipient_user_ids,
          recipient_chatbot_ids: recipient_chatbot_ids,
          recipient_audience: recipient_audience.presence)
  end

  def discussion
    eventable
  end
end
```

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `(discussion_id, sequence_id)` | UNIQUE | |
| `(eventable_type, eventable_id)` | INDEX | |
| `(eventable_id, kind)` | INDEX | |
| `(parent_id, discussion_id)` | PARTIAL | where discussion_id NOT NULL |
| `parent_id` | INDEX | |
| `position_key` | INDEX | |
| `user_id` | INDEX | |
| `created_at` | INDEX | |

---

## Position Key Format

Position keys are hierarchical strings that enable efficient tree queries:

```
"00001"                    # Top-level event, position 1
"00001-00001"              # First child of first event
"00001-00001-00003"        # Third grandchild
"00002-00015-00001-00007"  # Deeply nested event
```

Each segment is zero-padded to 5 digits, allowing up to 99,999 items per level.

---

## Uncertainties

1. **EventBus.broadcast** - External event handling mechanisms not fully documented
2. **SequenceService** - PostgreSQL sequence management implementation
3. **Queries::UsersByVolumeQuery** - User query helper for notification filtering
4. **PublishEventWorker** - Sidekiq worker that calls `trigger!`

**Confidence Level:** HIGH for core event model, MEDIUM for notification concern implementations.
