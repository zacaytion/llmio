# Webhook/Chatbot Event Delivery Analysis

## Executive Summary

Loomio's webhook delivery system lacks several enterprise reliability features. There is **no HMAC payload signing**, **no custom retry configuration** (uses Sidekiq defaults), and **no circuit breaker** for failing webhooks. Failed deliveries are logged to Sentry but do not disable webhooks or affect future delivery attempts.

---

## 1. Payload Signing Status

### Finding: NO HMAC Signatures
**Confidence: HIGH**

Loomio does **not** sign outgoing webhook payloads with HMAC or any other cryptographic signature.

#### Evidence

**Clients::Webhook** (`app/extras/clients/webhook.rb:1-23`)
- Simple HTTP POST client with no signature generation
- Inherits from `Clients::Base` which has no signing logic

**Clients::Base** (`app/extras/clients/base.rb:62-68`)
```ruby
def params_for(params = {})
  if require_json_payload?
    default_params.merge(params).to_json
  else
    default_params.merge(params)
  end
end
```
- Payload is serialized directly to JSON with no signature header added

**Clients::Request** (`app/extras/clients/request.rb:1-21`)
- Uses HTTParty for HTTP requests
- No signature headers added to requests

**ChatbotService.publish_event!** (`app/services/chatbot_service.rb:49-55`)
```ruby
if chatbot.kind == "webhook"
  serializer = "Webhook::#{chatbot.webhook_kind.classify}::EventSerializer".constantize
  payload = serializer.new(event, root: false, scope: {template_name: template_name, recipient: recipient}).as_json
  req = Clients::Webhook.new.post(chatbot.server, params: payload)
```
- Payload is serialized and POSTed without any signature computation

**Codebase Search Results:**
- Searched for `signature|hmac|sign` in `app/extras/clients/` - No matches
- Searched for `OpenSSL|HMAC|Digest|hexdigest` in `app/` - Found only unrelated uses:
  - SAML digest method (identity/authentication)
  - User merge SHA1 hash
  - Gravatar MD5 hash
  - User email SHA256 redaction

### Implications
- Webhook receivers cannot verify payload authenticity
- No protection against replay attacks
- Receivers must rely on source IP filtering or other mechanisms

---

## 2. Retry Configuration

### Finding: Sidekiq Default Retries (25 attempts over ~21 days)
**Confidence: HIGH**

Webhook delivery uses Sidekiq's default retry configuration with no customization.

#### Evidence

**GenericWorker** (`app/workers/generic_worker.rb:1-7`)
```ruby
class GenericWorker
  include Sidekiq::Worker

  def perform(class_name, method_name, arg1 = nil, arg2 = nil, arg3 = nil, arg4 = nil, arg5 = nil)
    class_name.constantize.send(method_name, *([arg1, arg2, arg3, arg4, arg5].compact))
  end
end
```
- No `sidekiq_options` specified
- Uses Sidekiq 7.x defaults

**Sidekiq Configuration** (`config/initializers/sidekiq.rb:1`)
```ruby
Sidekiq.default_job_options = { 'backtrace' => true }
```
- Only sets `backtrace: true`
- No custom retry count configured

**Sidekiq 7.x Default Behavior:**
- 25 retry attempts with exponential backoff
- Formula: `(retry_count ** 4) + 15 + (rand(30) * (retry_count + 1))`
- Total retry window: approximately 21 days
- After exhaustion: job moves to dead queue

**Comparison with Other Workers:**
Some workers explicitly disable retries (`app/workers/*.rb`):
```ruby
sidekiq_options queue: :low, retry: false  # reset_poll_stance_data_worker.rb:3
sidekiq_options retry: false               # repair_thread_worker.rb:3
```
GenericWorker (used for chatbot delivery) has no such configuration.

### Implications
- Failed webhook deliveries will retry 25 times over ~21 days
- Aggressive retry could flood failing endpoints
- No differentiation between transient and permanent failures

---

## 3. Circuit Breaker Status

### Finding: NO Circuit Breaker
**Confidence: HIGH**

There is no circuit breaker mechanism to automatically disable failing webhooks.

#### Evidence

**Chatbot Model Schema** (`db/schema.rb:156-170`)
```ruby
create_table "chatbots", force: :cascade do |t|
  t.string "kind"
  t.string "server"
  t.string "channel"
  t.string "access_token"
  t.integer "author_id"
  t.integer "group_id"
  t.datetime "created_at", null: false
  t.datetime "updated_at", null: false
  t.string "name"
  t.boolean "notification_only", default: false, null: false
  t.string "webhook_kind"
  t.string "event_kinds", array: true
```
- No `failure_count`, `last_failure_at`, `disabled`, `active` columns
- No circuit breaker state tracking

**Chatbot Model** (`app/models/chatbot.rb:1-18`)
```ruby
class Chatbot < ApplicationRecord
  belongs_to :group
  belongs_to :author, class_name: 'User'

  validates_presence_of :server
  validates_presence_of :name
  validates_inclusion_of :kind, in: ['matrix', 'webhook']
  validates_inclusion_of :webhook_kind, in: ['slack', 'microsoft', 'discord', 'markdown', 'webex', nil]
```
- No failure tracking methods
- No deactivation logic

**ChatbotService Error Handling** (`app/services/chatbot_service.rb:52-55`)
```ruby
req = Clients::Webhook.new.post(chatbot.server, params: payload)
if req.response.code != 200
  Sentry.capture_message("chatbot id #{chatbot.id} post event id #{event.id} failed: code: #{req.response.code} body: #{req.response.body}")
end
```
- Errors logged to Sentry only
- No failure count increment
- No automatic disabling
- No database state update

### Implications
- Failing webhooks continue receiving attempts indefinitely
- No protection against repeatedly failing endpoints
- No admin visibility into webhook health
- Resource waste on permanently broken integrations

---

## 4. Delivery Reliability Assessment

### Overall Rating: BASIC

| Feature | Status | Risk Level |
|---------|--------|------------|
| Payload Signing (HMAC) | Not Implemented | MEDIUM |
| Custom Retry Config | Not Implemented (uses defaults) | LOW |
| Circuit Breaker | Not Implemented | MEDIUM |
| Error Logging | Implemented (Sentry) | - |
| Retry Mechanism | Sidekiq defaults (25 retries) | - |
| Dead Letter Queue | Sidekiq dead queue | - |

### Positive Aspects
1. **Async Delivery**: Webhook delivery is fully async via Sidekiq (`app/models/concerns/events/notify/chatbots.rb:4`)
2. **Error Visibility**: Failed deliveries are captured in Sentry
3. **Retry on Failure**: Sidekiq provides automatic retries with exponential backoff
4. **Dead Queue**: Exhausted jobs go to Sidekiq dead queue for inspection

### Gaps
1. **No Payload Authentication**: Receivers cannot verify payload origin
2. **No Failure Tracking**: No database record of delivery attempts/failures
3. **No Auto-Disable**: Broken webhooks continue consuming resources
4. **No Delivery Confirmation**: No webhook delivery status in UI
5. **No Retry Customization**: Cannot tune retry behavior per webhook importance

---

## 5. Delivery Flow Summary

```
Event Created
    |
    v
PublishEventWorker.perform_async(event_id)
    |
    v
Event.sti_find(event_id).trigger!
    |
    v
Events::Notify::Chatbots#trigger! (if included)
    |
    v
GenericWorker.perform_async('ChatbotService', 'publish_event!', event_id)
    |
    v
ChatbotService.publish_event!(event_id)
    |
    +-- For each matching chatbot:
        |
        +-- Webhook kind: Clients::Webhook.new.post(url, payload)
        |       |
        |       +-- Success (HTTP 200): Done
        |       +-- Failure: Sentry.capture_message(...), raise exception -> Sidekiq retry
        |
        +-- Matrix kind: Redis pubsub to external bot
```

---

## File References

| File | Lines | Purpose |
|------|-------|---------|
| `app/services/chatbot_service.rb` | 22-70 | Main delivery logic |
| `app/models/chatbot.rb` | 1-18 | Chatbot model (no circuit breaker fields) |
| `app/extras/clients/webhook.rb` | 1-23 | HTTP client for webhooks |
| `app/extras/clients/base.rb` | 1-117 | Base HTTP client (no signing) |
| `app/extras/clients/request.rb` | 1-21 | HTTParty wrapper |
| `app/workers/generic_worker.rb` | 1-7 | Sidekiq worker (no custom retries) |
| `app/models/concerns/events/notify/chatbots.rb` | 1-6 | Event trigger for chatbots |
| `config/initializers/sidekiq.rb` | 1-23 | Sidekiq configuration |
| `config/webhook_event_kinds.yml` | 1-14 | 14 webhook-eligible event types |
| `db/schema.rb` | 156-170 | Chatbot table schema |

---

## Confidence Levels Summary

| Finding | Confidence | Basis |
|---------|------------|-------|
| No HMAC Signing | HIGH | Code review of all HTTP client layers; codebase search for signature/HMAC patterns |
| Default Sidekiq Retries | HIGH | GenericWorker has no sidekiq_options; Sidekiq config only sets backtrace |
| No Circuit Breaker | HIGH | Database schema review; Chatbot model review; ChatbotService error handling review |
