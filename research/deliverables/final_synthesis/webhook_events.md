# Webhook Events - Final Synthesis

## Executive Summary

This document provides implementation-ready specifications for the webhook/chatbot notification system in Loomio. The system delivers event notifications to configured webhook endpoints (Slack, Discord, Microsoft Teams, Webex) and Matrix chat rooms.

---

## Complete Webhook-Eligible Events (14 Events)

### Authoritative Source

**File**: `orig/loomio/config/webhook_event_kinds.yml`

These 14 events appear in the webhook configuration UI and can be selected by users:

| # | Event Kind | Category | Eventable Type | Template |
|---|------------|----------|----------------|----------|
| 1 | `new_discussion` | Discussion | Discussion | `discussion` |
| 2 | `discussion_edited` | Discussion | Discussion | `discussion` |
| 3 | `new_comment` | Comment | Comment | `comment` |
| 4 | `poll_created` | Poll | Poll | `poll` |
| 5 | `poll_edited` | Poll | Poll | `poll` |
| 6 | `poll_closing_soon` | Poll Lifecycle | Poll | `poll` |
| 7 | `poll_expired` | Poll Lifecycle | Poll | `poll` |
| 8 | `poll_closed_by_user` | Poll Lifecycle | Poll | `poll` |
| 9 | `poll_reopened` | Poll Lifecycle | Poll | `poll` |
| 10 | `stance_created` | Vote | Stance | `stance` |
| 11 | `stance_updated` | Vote | Stance | `stance` |
| 12 | `outcome_created` | Outcome | Outcome | `poll` |
| 13 | `outcome_updated` | Outcome | Outcome | `poll` |
| 14 | `outcome_review_due` | Outcome | Outcome | `poll` |

### Additional Chatbot-Capable Events (Not in UI)

These events include the `Chatbots` concern but are not exposed in the webhook configuration UI. They can only be triggered via `recipient_chatbot_ids`:

| Event Kind | Trigger Condition |
|------------|-------------------|
| `discussion_announced` | When users explicitly invited to discussion |
| `poll_announced` | When users explicitly invited to poll |
| `poll_reminder` | When poll author sends manual reminder |

---

## Delivery Mechanism

### Architecture Flow

```
1. Service creates event
   │
   ▼
2. Events::XYZ.publish!(...)
   │
   ▼
3. Event record saved to database
   │
   ▼
4. PublishEventWorker.perform_async(event.id)
   │
   ▼
5. Worker calls event.trigger!
   │
   ▼
6. If event includes Events::Notify::Chatbots:
   GenericWorker.perform_async('ChatbotService', 'publish_event!', event.id)
   │
   ▼
7. ChatbotService.publish_event!(event_id)
   │
   ├─► Webhook: HTTP POST via Clients::Webhook
   │
   └─► Matrix: Redis pub/sub to chatbot/publish channel
```

### Chatbot Selection Logic

**File**: `orig/loomio/app/services/chatbot_service.rb:30-31`

```ruby
chatbots.where(id: event.recipient_chatbot_ids).
    or(chatbots.where.any(event_kinds: event.kind))
```

Chatbots receive events when EITHER:
1. Chatbot ID is in event's `recipient_chatbot_ids` custom field (explicit targeting), OR
2. Event's kind matches one of chatbot's configured `event_kinds` array (subscription model)

### Template Resolution

**File**: `orig/loomio/app/services/chatbot_service.rb:33-36`

```ruby
template_name = event.eventable_type.tableize.singularize
template_name = 'poll' if event.eventable_type == 'Outcome'
template_name = 'group' if event.eventable_type == 'Membership'
template_name = 'notification' if chatbot.notification_only
```

| Eventable Type | Template Name |
|----------------|---------------|
| Discussion | `discussion` |
| Comment | `comment` |
| Poll | `poll` |
| Outcome | `poll` (remapped) |
| Stance | `stance` |
| Membership | `group` (remapped) |
| Any (notification_only=true) | `notification` (override) |

---

## Payload Structures

### Webhook Kinds Supported

**File**: `orig/loomio/app/models/chatbot.rb:8`

```ruby
validates_inclusion_of :webhook_kind, in: ['slack', 'microsoft', 'discord', 'markdown', 'webex', nil]
```

### Base Payload (Markdown Format)

**Serializer**: `orig/loomio/app/serializers/webhook/markdown/event_serializer.rb`

```json
{
  "text": "...rendered markdown content...",
  "icon_url": "https://example.com/group-logo.png",
  "username": "Loomio"
}
```

| Field | Source | Description |
|-------|--------|-------------|
| `text` | Template rendering | Main message content in Markdown |
| `icon_url` | `group.self_or_parent_logo_url(128)` | Group logo URL (128px) |
| `username` | `AppConfig.theme[:site_name]` | Site name (default: "Loomio") |

### Slack Format

**Serializer**: `orig/loomio/app/serializers/webhook/slack/event_serializer.rb`

```json
{
  "text": "...slack-formatted text..."
}
```

Note: `icon_url` and `username` are excluded - Slack handles these via webhook configuration.

### Discord Format

**Serializer**: `orig/loomio/app/serializers/webhook/discord/event_serializer.rb`

```json
{
  "content": "...truncated to 1900 chars...",
  "text": "...full markdown text...",
  "icon_url": "https://example.com/group-logo.png",
  "username": "Loomio"
}
```

Note: `content` field is truncated to 1900 characters (Discord's 2000 char limit minus buffer).

### Microsoft Teams Format

**Serializer**: `orig/loomio/app/serializers/webhook/microsoft/event_serializer.rb`

```json
{
  "@type": "MessageCard",
  "@context": "http://schema.org/extensions",
  "themeColor": "#658AE7",
  "text": "...rendered markdown...",
  "sections": []
}
```

| Field | Value | Description |
|-------|-------|-------------|
| `@type` | `"MessageCard"` | Microsoft card type |
| `@context` | `"http://schema.org/extensions"` | Schema context |
| `themeColor` | `AppConfig.theme[:primary_color]` | Accent color for card |
| `sections` | `[]` | Empty array (reserved for future use) |

### Webex Format

**Serializer**: `orig/loomio/app/serializers/webhook/webex/event_serializer.rb`

```json
{
  "markdown": "...same as text...",
  "text": "...rendered markdown...",
  "icon_url": "https://example.com/group-logo.png",
  "username": "Loomio"
}
```

Note: Webex uses `markdown` field for message content.

---

## Chatbot Model Schema

**File**: `orig/loomio/app/models/chatbot.rb`

| Column | Type | Description | Validations |
|--------|------|-------------|-------------|
| `id` | integer | Primary key | |
| `group_id` | integer | FK to groups | required |
| `author_id` | integer | FK to users | required |
| `name` | string | Display name | required |
| `kind` | string | `'matrix'` or `'webhook'` | required, enum |
| `server` | string | Webhook URL or Matrix server | required |
| `channel` | string | Matrix room ID | nullable |
| `access_token` | string | Matrix access token | nullable |
| `webhook_kind` | string | `'slack'`, `'microsoft'`, `'discord'`, `'markdown'`, `'webex'` | nullable, enum |
| `notification_only` | boolean | Use minimal notification template | default: false |
| `event_kinds` | string[] | Array of subscribed event kinds | PostgreSQL array |

### Chatbot Associations

```ruby
belongs_to :group
belongs_to :author, class_name: 'User'
```

---

## Retry Logic

### Current Behavior

`GenericWorker` uses Sidekiq's default retry policy:
- **Default retries**: 25
- **Backoff**: Exponential (sidekiq_retry_in formula)
- **Dead queue**: After 25 failures

The worker does NOT set `retry: false`, unlike some other workers:

```ruby
# GenericWorker - uses defaults
class GenericWorker
  include Sidekiq::Worker
  # No sidekiq_options override
end

# Compare to workers that disable retry:
# send_daily_catch_up_email_worker.rb:3:  sidekiq_options retry: false
```

---

## Error Handling

**File**: `orig/loomio/app/services/chatbot_service.rb:53-55`

```ruby
if req.response.code != 200
  Sentry.capture_message("chatbot id #{chatbot.id} post event id #{event.id} failed: code: #{req.response.code} body: #{req.response.body}")
end
```

### Error Scenarios

| Scenario | Handling |
|----------|----------|
| Event not found | `ActiveRecord::RecordNotFound` raised, job fails/retries |
| Eventable is nil | Early return, no notification sent |
| HTTP non-200 response | Logged to Sentry, job considered successful |
| Network error | Exception raised, job fails/retries |
| Serialization error | Exception raised, job fails/retries |

---

## API Endpoints

### Chatbot CRUD

**File**: `orig/loomio/config/routes.rb`

```ruby
resources :chatbots do
  post :test, on: :collection
end
```

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/chatbots` | List group chatbots |
| POST | `/api/v1/chatbots` | Create chatbot |
| GET | `/api/v1/chatbots/:id` | Show chatbot |
| PATCH | `/api/v1/chatbots/:id` | Update chatbot |
| DELETE | `/api/v1/chatbots/:id` | Delete chatbot |
| POST | `/api/v1/chatbots/test` | Test connection |

### Permitted Parameters

**File**: `orig/loomio/app/models/permitted_params.rb:194`

```ruby
[:name, :group_id, :kind, :webhook_kind, :server, :access_token, :channel, :notification_only, :event_kinds, {event_kinds: []}]
```

---

## Configuration Sources

### Environment Variables

| Variable | Usage |
|----------|-------|
| `CANONICAL_HOST` | Base URL for links in webhook payloads |

### AppConfig

| Key | Usage |
|-----|-------|
| `AppConfig.webhook_event_kinds` | List of UI-exposed event kinds |
| `AppConfig.theme[:site_name]` | Webhook username field |
| `AppConfig.theme[:primary_color]` | Microsoft Teams themeColor |

---

## Matrix Integration

### Redis Pub/Sub Channels

| Channel | Purpose |
|---------|---------|
| `chatbot/test` | Test bot configuration (new client each time) |
| `chatbot/publish` | Send messages to Matrix rooms (cached clients) |

### Publish Payload

```json
{
  "config": {
    "server": "https://matrix.org",
    "access_token": "...",
    "channel": "!roomid:matrix.org"
  },
  "payload": {
    "html": "<p>Message content</p>"
  }
}
```

---

## Source File References

| Component | File | Purpose |
|-----------|------|---------|
| Event kinds config | `orig/loomio/config/webhook_event_kinds.yml` | Authoritative list |
| Chatbots concern | `orig/loomio/app/models/concerns/events/notify/chatbots.rb` | Event trigger hook |
| ChatbotService | `orig/loomio/app/services/chatbot_service.rb` | Delivery logic |
| Chatbot model | `orig/loomio/app/models/chatbot.rb` | Schema and validations |
| Markdown serializer | `orig/loomio/app/serializers/webhook/markdown/event_serializer.rb` | Base payload |
| Slack serializer | `orig/loomio/app/serializers/webhook/slack/event_serializer.rb` | Slack-specific |
| Discord serializer | `orig/loomio/app/serializers/webhook/discord/event_serializer.rb` | Discord truncation |
| Microsoft serializer | `orig/loomio/app/serializers/webhook/microsoft/event_serializer.rb` | MessageCard |
| Webex serializer | `orig/loomio/app/serializers/webhook/webex/event_serializer.rb` | Webex format |
| HTTP client | `orig/loomio/app/extras/clients/webhook.rb` | HTTP delivery |
| Generic worker | `orig/loomio/app/workers/generic_worker.rb` | Async execution |

---

## Conclusion

This synthesis provides complete implementation details for the webhook/chatbot system:

- **14 webhook-eligible events** confirmed from authoritative source
- **5 webhook formats** (Slack, Discord, Microsoft, Webex, Markdown) with payload structures
- **2 delivery mechanisms** (HTTP POST for webhooks, Redis pub/sub for Matrix)
- **Retry logic** via Sidekiq defaults (25 retries with exponential backoff)

The implementation should use a job queue for processing with similar retry semantics, and implement each webhook format serializer to match the documented payload structures.
