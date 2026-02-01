# Webhook Events: Service Documentation

## ChatbotService

**File**: `/Users/z/Code/loomio/app/services/chatbot_service.rb`

### Overview

ChatbotService handles the delivery of event notifications to configured chatbots (webhooks and Matrix rooms). It provides CRUD operations for chatbot configurations and the core event publishing logic.

### Methods

#### `publish_event!(event_id)`

**Location**: Lines 22-70

The main webhook delivery method. Called asynchronously via `GenericWorker`.

```ruby
def self.publish_event!(event_id)
  event = Event.find(event_id)
  event.reload
  return if event.eventable.nil?

  chatbots = event.eventable.group.chatbots

  CACHE_REDIS_POOL.with do |client|
    # Find matching chatbots via OR condition
    chatbots.where(id: event.recipient_chatbot_ids).
        or(chatbots.where.any(event_kinds: event.kind)).each do |chatbot|

      # Template resolution logic
      template_name = event.eventable_type.tableize.singularize
      template_name = 'poll' if event.eventable_type == 'Outcome'
      template_name = 'group' if event.eventable_type == 'Membership'
      template_name = 'notification' if chatbot.notification_only

      # ... locale setup ...

      if chatbot.kind == "webhook"
        # HTTP POST to webhook URL
        serializer = "Webhook::#{chatbot.webhook_kind.classify}::EventSerializer".constantize
        payload = serializer.new(event, root: false, scope: {template_name: template_name, recipient: recipient}).as_json
        req = Clients::Webhook.new.post(chatbot.server, params: payload)
        # Error logging to Sentry if response != 200
      else
        # Matrix: Publish via Redis pub/sub
        client.publish("chatbot/publish", {
          config: chatbot.config,
          payload: { html: rendered_template }
        }.to_json)
      end
    end
  end
end
```

**Key Logic**:

1. **Chatbot Selection** (line 30-31):
   ```ruby
   chatbots.where(id: event.recipient_chatbot_ids).
       or(chatbots.where.any(event_kinds: event.kind))
   ```
   - Selects chatbots explicitly targeted via `recipient_chatbot_ids` OR
   - Chatbots subscribed to this event kind via `event_kinds` array

2. **Template Resolution** (lines 33-36):
   | Eventable Type | Template Name |
   |----------------|---------------|
   | Discussion | `discussion` |
   | Comment | `comment` |
   | Poll | `poll` |
   | Outcome | `poll` (remapped) |
   | Stance | `stance` |
   | Membership | `group` (remapped) |
   | Any (notification_only) | `notification` (override) |

3. **Locale Handling** (lines 42-47):
   ```ruby
   example_user = chatbot.author || chatbot.group.creator
   recipient = LoggedOutUser.new(
     locale: example_user.locale,
     time_zone: example_user.time_zone,
     date_time_pref: example_user.date_time_pref
   )
   I18n.with_locale(recipient.locale) { ... }
   ```

4. **Webhook Serialization** (line 50-51):
   ```ruby
   serializer = "Webhook::#{chatbot.webhook_kind.classify}::EventSerializer".constantize
   payload = serializer.new(event, root: false, scope: {...}).as_json
   ```
   Dynamically loads serializer based on `webhook_kind` (slack, microsoft, discord, markdown, webex).

5. **Webhook Delivery** (line 52):
   ```ruby
   req = Clients::Webhook.new.post(chatbot.server, params: payload)
   ```

6. **Error Handling** (lines 53-55):
   ```ruby
   if req.response.code != 200
     Sentry.capture_message("chatbot id #{chatbot.id} post event id #{event.id} failed: code: #{req.response.code} body: #{req.response.body}")
   end
   ```

#### `create(chatbot:, actor:)`

**Location**: Lines 2-7

Creates a new chatbot integration.

```ruby
def self.create(chatbot:, actor:)
  actor.ability.authorize! :create, chatbot
  return false unless chatbot.valid?
  chatbot.author = actor
  chatbot.save!
end
```

#### `update(chatbot:, params:, actor:)`

**Location**: Lines 9-15

Updates an existing chatbot configuration.

```ruby
def self.update(chatbot:, params:, actor:)
  actor.ability.authorize! :update, chatbot
  params.delete(:access_token) unless params[:access_token].present?
  chatbot.assign_attributes(params)
  return false unless chatbot.valid?
  chatbot.save!
end
```

Note: Empty `access_token` values are preserved (not overwritten with empty strings).

#### `destroy(chatbot:, actor:)`

**Location**: Lines 17-20

Deletes a chatbot integration.

#### `publish_test!(params)`

**Location**: Lines 72-83

Sends a test message to verify webhook connectivity.

```ruby
def self.publish_test!(params)
  case params[:kind]
  when 'slack_webhook'
    Clients::Webhook.new.post(params[:server], params: {text: I18n.t('chatbot.connection_test_successful')})
  else
    # Matrix test via Redis pub/sub
    MAIN_REDIS_POOL.with do |client|
      data = params.slice(:server, :access_token, :channel)
      data.merge!(message: I18n.t('chatbot.connection_test_successful', group: params[:group_name]))
      client.publish("chatbot/test", data.to_json)
    end
  end
end
```

## Clients::Webhook

**Inferred from**: `app/services/chatbot_service.rb:52`

HTTP client for webhook delivery. Provides `post(url, params:)` method.

## GenericWorker

**Location**: Used at `app/models/concerns/events/notify/chatbots.rb:4`

Sidekiq worker that calls arbitrary service methods:

```ruby
GenericWorker.perform_async('ChatbotService', 'publish_event!', id)
```

This pattern allows async execution of `ChatbotService.publish_event!(id)`.

## PublishEventWorker

**Location**: `app/models/event.rb:64`

Called after event creation:

```ruby
def self.publish!(eventable, **args)
  event = build(eventable, **args)
  event.save!
  PublishEventWorker.perform_async(event.id)
  event
end
```

Worker calls `event.trigger!` which invokes the Chatbots concern.

## Event Flow Sequence

```
1. Service (e.g., PollService.create)
   |
   v
2. Events::PollCreated.publish!(poll, actor)
   |
   v
3. Event record saved to database
   |
   v
4. PublishEventWorker.perform_async(event.id)
   |
   v
5. PublishEventWorker calls event.trigger!
   |
   v
6. Events::Notify::Chatbots#trigger! (via include)
   |
   v
7. GenericWorker.perform_async('ChatbotService', 'publish_event!', event.id)
   |
   v
8. ChatbotService.publish_event!(event_id)
   |
   v
9. For each matching chatbot:
   |
   +-> Webhook: HTTP POST via Clients::Webhook
   |
   +-> Matrix: Redis pub/sub to chatbot/publish
```

## Webhook Payload Formats

### Markdown (Generic)

```json
{
  "text": "..rendered markdown text..",
  "icon_url": "https://example.com/group/logo.png",
  "username": "Loomio"
}
```

### Slack

Same as Markdown, but uses Slack-specific template rendering.

### Discord

```json
{
  "content": "..truncated to 1900 chars..",
  "text": "..full markdown text..",
  "icon_url": "https://example.com/group/logo.png",
  "username": "Loomio"
}
```

### Microsoft Teams

```json
{
  "@type": "MessageCard",
  "@context": "http://schema.org/extensions",
  "themeColor": "#658AE7",
  "text": "..rendered markdown text..",
  "sections": []
}
```

### Webex

```json
{
  "markdown": "..same as text..",
  "text": "..rendered markdown text..",
  "icon_url": "https://example.com/group/logo.png",
  "username": "Loomio"
}
```

## Template Files

**Base Directory**: `/Users/z/Code/loomio/app/views/chatbot/`

### Markdown Templates

| Template | Events Using |
|----------|-------------|
| `markdown/discussion.text.erb` | new_discussion, discussion_edited, discussion_announced |
| `markdown/comment.text.erb` | new_comment |
| `markdown/poll.text.erb` | poll_created, poll_edited, poll_closing_soon, poll_expired, poll_closed_by_user, poll_reopened, poll_announced, poll_reminder, outcome_created, outcome_updated, outcome_review_due |
| `markdown/stance.text.erb` | stance_created, stance_updated |
| `markdown/notification.text.erb` | Any (when notification_only=true) |

### Slack Templates

| Template | Notes |
|----------|-------|
| `slack/discussion.text.erb` | Slack-formatted |
| `slack/comment.text.erb` | Slack-formatted |
| `slack/poll.text.erb` | Slack-formatted |
| `slack/stance.text.erb` | Slack-formatted |
| `slack/notification.text.erb` | Slack-formatted |

### Partials

Common partials in `markdown/`:
- `_notification.text.erb` - Event header with kind-specific text
- `_title.text.erb` - Eventable title with link
- `_body.text.erb` - Description/details content
- `_results.text.erb` - Poll results summary
- `_vote.text.erb` - Vote link
- `_rules.text.erb` - Poll rules display
- `_outcome.text.erb` - Outcome statement
- `_undecided.text.erb` - Undecided voters list
- `_attachments.text.erb` - File attachments

## API Endpoints

### Chatbot CRUD

From `config/routes.rb`:

```ruby
resources :chatbots do
  post :test, on: :collection
end
```

- `GET /api/v1/chatbots` - List group chatbots
- `POST /api/v1/chatbots` - Create chatbot
- `GET /api/v1/chatbots/:id` - Show chatbot
- `PATCH /api/v1/chatbots/:id` - Update chatbot
- `DELETE /api/v1/chatbots/:id` - Delete chatbot
- `POST /api/v1/chatbots/test` - Test connection

### Permitted Parameters

From `app/models/permitted_params.rb:194`:

```ruby
[:name, :group_id, :kind, :webhook_kind, :server, :access_token, :channel, :notification_only, :event_kinds, {event_kinds: []}]
```

## Error Handling

1. **Event Not Found**: `Event.find(event_id)` raises `ActiveRecord::RecordNotFound`
2. **Missing Eventable**: Early return if `event.eventable.nil?`
3. **HTTP Errors**: Logged to Sentry with chatbot ID, event ID, response code and body
4. **Serialization Errors**: Not explicitly handled (would raise to worker retry)

## Performance Considerations

1. **Async Processing**: All webhook deliveries are async via Sidekiq
2. **Redis Pool**: Uses `CACHE_REDIS_POOL` for Redis operations
3. **No Batching**: Each chatbot receives individual HTTP request
4. **Template Rendering**: Done inline during webhook delivery

## Configuration

### Environment Variables

None directly used by ChatbotService, but:
- `CANONICAL_HOST` - Used in serializers for URLs
- Sentry configuration for error reporting

### AppConfig

- `AppConfig.webhook_event_kinds` - List of UI-exposed event kinds
- `AppConfig.theme[:site_name]` - Used as webhook username
- `AppConfig.theme[:primary_color]` - Used in Microsoft Teams themeColor
