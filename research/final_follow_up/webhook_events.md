# Webhook Events - Follow-up Items

## Executive Summary

The third-party discovery documents provide accurate and well-verified information about Loomio's webhook system. Their investigation correctly identified errors in earlier discovery claims and provided file-level evidence. The remaining follow-up items are minor clarifications and implementation details needed for the Go rewrite.

---

## Discrepancies Between Sources

### 1. Event List Corrections (RESOLVED)

**Third-party Discovery Status**: CORRECT

The third-party correctly identified that our original research brief (`research/follow_up/webhook_eligible_events.md`) contained errors inherited from earlier discovery documents:

| Event | Our Research Brief | Third-party Finding | Actual (Verified) |
|-------|-------------------|---------------------|-------------------|
| `user_added_to_group` | Listed as webhook-eligible | NOT webhook-eligible | **Third-party CORRECT** - File lacks `Chatbots` concern |
| `membership_requested` | Listed as webhook-eligible | NOT webhook-eligible | **Third-party CORRECT** - File lacks `Chatbots` concern |
| `poll_reopened` | Not listed | IS webhook-eligible | **Third-party CORRECT** - In `webhook_event_kinds.yml` |
| `outcome_review_due` | Not listed | IS webhook-eligible | **Third-party CORRECT** - In `webhook_event_kinds.yml` |

**Evidence verified at**:
- `orig/loomio/app/models/events/user_added_to_group.rb:1-19` - Only includes `InApp`, `ByEmail`
- `orig/loomio/app/models/events/membership_requested.rb:1-28` - Only includes `InApp`, `ByEmail`
- `orig/loomio/app/models/events/poll_reopened.rb:2` - Includes `Chatbots`
- `orig/loomio/app/models/events/outcome_review_due.rb:4` - Includes `Chatbots`
- `orig/loomio/config/webhook_event_kinds.yml` - Authoritative list of 14 events

**Priority**: LOW - No action needed, third-party is correct

---

## Areas Requiring Clarification

### 1. Retry Logic for Failed Webhooks

**Priority**: MEDIUM

**Third-party Claim**: "No retry mechanism visible in code" (confidence.md, line 118)

**Issue**: The third-party documents state there is no retry mechanism, but this needs verification of Sidekiq defaults.

**Questions for Third Party**:
1. Does `GenericWorker` inherit Sidekiq's default retry behavior (25 retries with exponential backoff)?
2. Are there any Sidekiq configuration overrides in `config/sidekiq.yml` that affect webhook delivery?

**Investigation Targets**:
- `orig/loomio/app/workers/generic_worker.rb` - No explicit `sidekiq_options retry: false`
- `orig/loomio/config/sidekiq.yml` - Check default retry configuration

**Current Finding**:
`GenericWorker` does NOT set `retry: false`, which means it uses Sidekiq's default retry policy (25 retries). This is actually **good** for reliability but should be confirmed.

```ruby
# orig/loomio/app/workers/generic_worker.rb
class GenericWorker
  include Sidekiq::Worker
  # No sidekiq_options - uses defaults
  def perform(class_name, method_name, arg1 = nil, ...)
    class_name.constantize.send(method_name, *([arg1, ...].compact))
  end
end
```

---

### 2. HTTP Client Timeout Configuration

**Priority**: LOW

**Third-party Claim**: "Webhook timeout: Not specified in code (uses Clients::Webhook defaults)" (confidence.md, line 119)

**Issue**: The actual HTTP timeout for webhook delivery is not documented.

**Questions for Third Party**:
1. What timeout does `Clients::Request` use for HTTP connections?
2. Is there any circuit breaker or rate limiting on failed webhooks?

**Investigation Targets**:
- `orig/loomio/app/extras/clients/request.rb` - HTTP client implementation
- `orig/loomio/app/extras/clients/base.rb` - Base class with `perform` method

---

### 3. Matrix Chatbot Client Caching

**Priority**: LOW

**Third-party Claim**: Documents Matrix client caching but notes "no eviction strategy" (services.md)

**Noted in realtime.md**: "This caching has no eviction strategy - potential memory concern if many unique configs are used."

**Questions for Third Party**:
1. Is this a known issue in production?
2. Should we implement LRU eviction in the Go version?

**Investigation Target**:
- `orig/loomio_channel_server/bots.js:37-47`

---

### 4. Serializer Fallback Chain

**Priority**: LOW

**Observation**: `Clients::Webhook.serialized_event` uses a different serializer resolution pattern than `ChatbotService.publish_event!`

**File**: `orig/loomio/app/extras/clients/webhook.rb:15-21`

```ruby
def serialized_event(event, format, webhook)
  serializer = [
    "Webhook::#{format.classify}::#{event.kind.classify}Serializer",
    "Webhook::#{format.classify}::#{event.eventable.class}Serializer",
    "Webhook::#{format.classify}::BaseSerializer"
  ].detect { |str| str.constantize rescue nil }.constantize
  # ...
end
```

**But** `ChatbotService.publish_event!` at line 50 uses:
```ruby
serializer = "Webhook::#{chatbot.webhook_kind.classify}::EventSerializer".constantize
```

**Questions for Third Party**:
1. Is `Clients::Webhook.serialized_event` actually called in production?
2. Is the `post_content!` method dead code, or is it used elsewhere?
3. The fallback chain suggests event-kind-specific serializers may exist - are there any?

---

## Contradictions Needing Resolution

### None Identified

The third-party documents are internally consistent and align with source code verification.

---

## Incomplete Areas in Third-party Documentation

### 1. HMAC/Signature Verification

**Priority**: HIGH

**Missing**: Neither source documents whether Loomio supports webhook signatures (HMAC) for payload verification.

**Common webhook pattern** that may or may not be implemented:
```
X-Loomio-Signature: sha256=<hmac_of_payload>
```

**Investigation Targets**:
- `orig/loomio/app/extras/clients/webhook.rb` - Check for signature generation
- `orig/loomio/app/extras/clients/base.rb` - Check `headers_for` method
- `orig/loomio/app/services/chatbot_service.rb:52` - Check if additional headers sent

**Question for Third Party**:
1. Does Loomio sign outgoing webhook payloads?
2. If so, what algorithm and header format?

---

### 2. Rate Limiting on Webhook Delivery

**Priority**: MEDIUM

**Missing**: No documentation on rate limiting for webhook delivery.

**Questions**:
1. Is there any per-chatbot rate limiting?
2. What happens if a webhook endpoint is slow/unresponsive?
3. Are there circuit breakers for consistently failing webhooks?

---

### 3. Webhook Payload Size Limits

**Priority**: LOW

**Discord Documented**: 1900 character truncation (verified in code)

**Missing**: Are there size limits for other webhook formats?

**Questions**:
1. What is the maximum payload size for Slack/Microsoft/Webex?
2. How does the system handle very large attachments lists?

---

## Specific File References for Investigation

| File | Line(s) | Investigation Purpose |
|------|---------|----------------------|
| `orig/loomio/app/extras/clients/request.rb` | All | HTTP timeout configuration |
| `orig/loomio/config/sidekiq.yml` | All | Default retry configuration |
| `orig/loomio/app/extras/clients/webhook.rb` | 3-5 | Dead code investigation (`post_content!`) |
| `orig/loomio/app/extras/clients/base.rb` | 100-103 | Headers for signature verification |
| `orig/loomio_channel_server/bots.js` | 37-47 | Matrix client caching strategy |

---

## Priority Summary

| Priority | Count | Items |
|----------|-------|-------|
| HIGH | 1 | HMAC/Signature verification |
| MEDIUM | 2 | Retry logic, Rate limiting |
| LOW | 4 | Timeouts, Matrix caching, Serializer chain, Payload sizes |

---

## Action Items for Go Implementation

1. **Confirm Sidekiq retry defaults** apply to webhook delivery before implementing River job configuration
2. **Investigate HMAC signatures** before implementing webhook delivery
3. **Consider implementing rate limiting** even if original doesn't have it (defense in depth)
4. **Document payload size limits** for each webhook format

---

## Conclusion

The third-party discovery documents are **highly accurate** with 5/5 confidence ratings justified by file:line evidence. The two errors they corrected in the earlier discovery documents were verified against source code.

The remaining follow-up items are implementation details (timeouts, signatures, rate limits) rather than architectural discrepancies. These should be investigated during implementation rather than blocking synthesis.
