# Notification Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/notification.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Notification model represents user notifications generated from events. Notifications provide in-app alerts about activity relevant to the user, linking to the triggering event and its associated content.

---

## Attributes

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `user_id` | integer | - | YES | FK to users | Notification recipient |
| `event_id` | integer | - | YES | FK to events | Triggering event |
| `actor_id` | integer | - | YES | FK to users | User who triggered action |
| `url` | string | - | YES | - | Deep link URL |
| `viewed` | boolean | false | NO | - | Read status |
| `translation_values` | jsonb | {} | NO | - | I18n interpolation values |
| `created_at` | datetime | - | YES | Creation timestamp |
| `updated_at` | datetime | - | YES | Last update |

---

## translation_values JSONB Format

```json
{
  "name": "Jane Doe",
  "title": "Weekly Team Sync",
  "poll_type": "Proposal"
}
```

These values are interpolated into notification translation strings:
- `name` - Actor's display name
- `title` - Content title (discussion, poll, etc.)
- `poll_type` - Localized poll type name

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `user` | presence required | always |
| `event` | presence required | always |

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `user` | User | - | Notification recipient |
| `actor` | User | class_name: "User" | Action performer |
| `event` | Event | - | Source event |

---

## Scopes

```ruby
scope :dangling, -> {
  joins('left join events e on notifications.event_id = e.id')
    .joins('left join users u on u.id = notifications.user_id')
    .where('e.id is null or u.id is null')
}

scope :user_mentions, -> {
  joins(:event).where("events.kind": :user_mentioned)
}
```

---

## Delegate Methods

```ruby
delegate :eventable, to: :event, allow_nil: true
delegate :kind, to: :event, allow_nil: true
delegate :locale, to: :user
delegate :message_channel, to: :user
```

---

## Instance Methods

The Notification model is intentionally simple, with most logic handled by the Event concerns that create notifications.

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `event_id` | INDEX | |
| `user_id` | INDEX | |
| `(user_id, id)` | INDEX | For pagination |
| `id` | INDEX (desc) | For reverse chronological listing |

---

## Creation Flow

Notifications are created by `Events::Notify::InApp` concern:

```ruby
module Events::Notify::InApp
  def notify_users!
    notifications.import(built_notifications)
    built_notifications.each { |n|
      MessageChannelService.publish_models(Array(n), user_id: n.user_id)
    }
  end

  def built_notifications
    @built ||= notification_recipients.active.map { |recipient|
      notification_for(recipient)
    }
  end

  def notification_for(recipient)
    I18n.with_locale(recipient.locale) do
      notifications.build(
        user: recipient,
        actor: notification_actor,
        url: notification_url,
        translation_values: notification_translation_values
      )
    end
  end

  def notification_actor
    user.presence
  end

  def notification_url
    polymorphic_path(eventable)
  end

  def notification_translation_values
    {
      name: notification_translation_name,
      title: TranslationService.plain_text(eventable.title_model, :title, user),
      poll_type: (I18n.t(:"poll_types.#{notification_poll_type}") if notification_poll_type)
    }.compact
  end
end
```

---

## Real-time Publishing

When created, notifications are immediately pushed to the user's message channel:

```ruby
MessageChannelService.publish_models(Array(n), user_id: n.user_id)
```

The user receives notifications on their personal channel `/user-{user.key}`.

---

## Event Kinds That Generate Notifications

Most event types include `Events::Notify::InApp` and generate notifications. Key exceptions:
- Some system events
- Events where the user is the actor (filtered in `notification_recipients`)

Common notification-generating events:
- `new_discussion` - Discussion created and announced
- `new_comment` - Comment on subscribed discussion
- `poll_created` - Poll created and announced
- `stance_created` - Vote on watched poll
- `outcome_created` - Outcome published
- `user_mentioned` - @mentioned in content
- `group_mentioned` - Group @mentioned
- `poll_closing_soon` - Poll closing reminder
- `invitation_accepted` - Invitee accepted
- `membership_request_approved` - Join request approved

---

## Notification Volume Filtering

Notifications respect user volume preferences. The `notification_recipients` method filters:

```ruby
def notification_recipients
  Queries::UsersByVolumeQuery.app_notifications(eventable)
    .where(id: all_recipient_user_ids)
    .where.not(id: user.id || 0)  # Exclude actor
end
```

Volume levels:
- `mute` (0) - No notifications
- `quiet` (1) - In-app only (includes notifications)
- `normal` (2) - In-app + email
- `loud` (3) - All notifications

---

## Marking as Read

Notifications are marked as viewed via:
- User clicking/dismissing the notification
- `NotificationService.mark_as_read` service method
- Bulk read operations

---

## Retention & Cleanup

Notifications are cleaned up when:
- Associated event is destroyed (via `dependent: :destroy`)
- Associated user is destroyed (via dangling cleanup)
- Manual cleanup tasks

---

## Uncertainties

1. **Notification limit per user** - No visible cap on stored notifications
2. **Auto-dismiss behavior** - Whether viewing related content marks notifications read
3. **Email relationship** - Email notifications are separate from in-app (via Events::Notify::ByEmail)

**Confidence Level:** HIGH for model structure, MEDIUM for notification lifecycle details.
