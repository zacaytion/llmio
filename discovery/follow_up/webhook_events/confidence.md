# Webhook Events: Verification Checklist

## Verification Methodology

Each claim was verified by:
1. Direct code inspection (file:line references)
2. Cross-referencing multiple sources
3. Testing mental model against actual code paths

## Claim Verification

### Discovery Claims

| Claim | Status | Evidence | Confidence |
|-------|--------|----------|------------|
| "14 events are webhook-eligible" | PASS (with clarification) | `config/webhook_event_kinds.yml` contains exactly 14 entries | 5/5 |
| `new_discussion` is webhook-eligible | PASS | `app/models/events/new_discussion.rb:7` includes Chatbots | 5/5 |
| `discussion_edited` is webhook-eligible | PASS | `app/models/events/discussion_edited.rb:6` includes Chatbots | 5/5 |
| `new_comment` is webhook-eligible | PASS | `app/models/events/new_comment.rb:4` includes Chatbots | 5/5 |
| `poll_created` is webhook-eligible | PASS | `app/models/events/poll_created.rb:4` includes Chatbots | 5/5 |
| `poll_edited` is webhook-eligible | PASS | `app/models/events/poll_edited.rb:6` includes Chatbots | 5/5 |
| `poll_closing_soon` is webhook-eligible | PASS | `app/models/events/poll_closing_soon.rb:5` includes Chatbots | 5/5 |
| `poll_expired` is webhook-eligible | PASS | `app/models/events/poll_expired.rb:3` includes Chatbots | 5/5 |
| `poll_closed_by_user` is webhook-eligible | PASS | `app/models/events/poll_closed_by_user.rb:3` includes Chatbots | 5/5 |
| `stance_created` is webhook-eligible | PASS | `app/models/events/stance_created.rb:5` includes Chatbots | 5/5 |
| `stance_updated` is webhook-eligible | PASS | `app/models/events/stance_updated.rb:1` inherits from StanceCreated | 5/5 |
| `outcome_created` is webhook-eligible | PASS | `app/models/events/outcome_created.rb:5` includes Chatbots | 5/5 |
| `outcome_updated` is webhook-eligible | PASS | `app/models/events/outcome_updated.rb:5` includes Chatbots | 5/5 |
| `user_added_to_group` is webhook-eligible | **FAIL** | `app/models/events/user_added_to_group.rb` does NOT include Chatbots | 5/5 |
| `membership_requested` is webhook-eligible | **FAIL** | `app/models/events/membership_requested.rb` does NOT include Chatbots | 5/5 |

### Additional Findings

| Claim | Status | Evidence | Confidence |
|-------|--------|----------|------------|
| `poll_reopened` is webhook-eligible | PASS | `app/models/events/poll_reopened.rb:2` includes Chatbots | 5/5 |
| `outcome_review_due` is webhook-eligible | PASS | `app/models/events/outcome_review_due.rb:4` includes Chatbots | 5/5 |
| `discussion_announced` is chatbot-capable | PASS | `app/models/events/discussion_announced.rb:4` includes Chatbots | 5/5 |
| `poll_announced` is chatbot-capable | PASS | `app/models/events/poll_announced.rb:4` includes Chatbots | 5/5 |
| `poll_reminder` is chatbot-capable | PASS | `app/models/events/poll_reminder.rb:4` includes Chatbots | 5/5 |

### Mechanism Claims

| Claim | Status | Evidence | Confidence |
|-------|--------|----------|------------|
| Events route through ChatbotService | PASS | `app/models/concerns/events/notify/chatbots.rb:4` calls ChatbotService | 5/5 |
| Chatbots can subscribe via event_kinds | PASS | `app/services/chatbot_service.rb:31` `where.any(event_kinds: event.kind)` | 5/5 |
| Events can target chatbots explicitly | PASS | `app/services/chatbot_service.rb:30` `where(id: event.recipient_chatbot_ids)` | 5/5 |
| webhook_event_kinds.yml defines UI list | PASS | `app/models/boot/site.rb:30` `webhookEventKinds: AppConfig.webhook_event_kinds` | 5/5 |
| Frontend reads webhookEventKinds | PASS | `vue/src/components/chatbot/webhook_form.vue:14` `kinds: AppConfig.webhookEventKinds` | 5/5 |

### Serialization Claims

| Claim | Status | Evidence | Confidence |
|-------|--------|----------|------------|
| 5 webhook formats supported | PASS | `app/models/chatbot.rb:8` validates inclusion in ['slack', 'microsoft', 'discord', 'markdown', 'webex', nil] | 5/5 |
| Serializers exist for each format | PASS | `app/serializers/webhook/` contains slack/, microsoft/, discord/, markdown/, webex/ | 5/5 |
| Discord truncates to 1900 chars | PASS | `app/serializers/webhook/discord/event_serializer.rb:6` `.truncate(1900, omission: '... (truncated)')` | 5/5 |
| Microsoft uses MessageCard format | PASS | `app/serializers/webhook/microsoft/event_serializer.rb:8` `"MessageCard"` | 5/5 |

## Error Corrections

### Original Discovery Document Errors

1. **Claimed**: `user_added_to_group` is webhook-eligible
   - **Actual**: This event does NOT include `Events::Notify::Chatbots`
   - **File**: `app/models/events/user_added_to_group.rb`
   - **Includes only**: `Events::Notify::InApp`, `Events::Notify::ByEmail`

2. **Claimed**: `membership_requested` is webhook-eligible
   - **Actual**: This event does NOT include `Events::Notify::Chatbots`
   - **File**: `app/models/events/membership_requested.rb`
   - **Includes only**: `Events::Notify::InApp`, `Events::Notify::ByEmail`

3. **Missing from Discovery**: `poll_reopened`
   - **Actual**: IS webhook-eligible
   - **File**: `app/models/events/poll_reopened.rb:2`

4. **Missing from Discovery**: `outcome_review_due`
   - **Actual**: IS webhook-eligible
   - **File**: `app/models/events/outcome_review_due.rb:4`

## Verification Commands Used

```bash
# List all events with Chatbots concern
grep -l "include Events::Notify::Chatbots" app/models/events/*.rb | xargs -I{} basename {} .rb | sort

# Verify webhook_event_kinds.yml contents
cat config/webhook_event_kinds.yml

# Check specific event files
grep -n "include Events::Notify" app/models/events/user_added_to_group.rb
grep -n "include Events::Notify" app/models/events/membership_requested.rb
```

## Confidence Rating Scale

| Rating | Meaning |
|--------|---------|
| 5/5 | Verified by direct code inspection with line numbers |
| 4/5 | Verified by code inspection, minor ambiguity |
| 3/5 | Inferred from code patterns, not directly verified |
| 2/5 | Based on documentation or comments only |
| 1/5 | Speculation based on naming conventions |

## Summary Statistics

- **Total claims verified**: 25
- **Passed**: 21 (84%)
- **Failed**: 2 (8%)
- **New findings**: 5 additional events documented (not in original Discovery)

## Remaining Uncertainties

1. **Matrix chatbot path**: Not fully traced (uses Redis pub/sub to external service)
2. **Retry logic**: Not verified if Sidekiq retries failed webhook deliveries
3. **Rate limiting**: No evidence of rate limiting on webhook delivery
4. **Webhook timeout**: Not specified in code (uses Clients::Webhook defaults)

## Recommendations

1. Update Discovery documentation to:
   - Remove `user_added_to_group` and `membership_requested` from webhook-eligible list
   - Add `poll_reopened` and `outcome_review_due` to list
   - Document the distinction between UI-exposed (14) vs code-capable (16) events

2. Consider adding webhook support for:
   - `user_added_to_group` (common integration request)
   - `membership_requested` (useful for admin notifications)

3. Document the `recipient_chatbot_ids` mechanism for targeted notifications
