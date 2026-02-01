# Webhook Eligible Events - Follow-up Investigation Brief

## Discrepancy Summary

Both Discovery and Research agree that **14 events are webhook-eligible**, but **neither document set enumerates all 14 events**. This is a gap in both documents that needs codebase verification.

## Discovery Claims

**Source**: `discovery/initial/integrations/models.md`

Documents the Chatbot/Webhook model includes:
- `event_kinds: string[]` - array of event kinds to trigger on

**Source**: `discovery/initial/events/models.md`

States "14 webhook-eligible events" but only lists examples:
- `new_discussion`, `discussion_edited`
- `new_comment`
- `poll_created`, `poll_edited`, `poll_closing_soon`, `poll_expired`, `poll_closed_by_user`
- `stance_created`, `stance_updated`
- `outcome_created`, `outcome_updated`
- `user_added_to_group`, `membership_requested`

This list has 14 items, but is described as examples rather than an exhaustive list.

## Our Research Claims

**Source**: `research/investigation/api.md`

States "14 webhook-eligible events" referencing the CLAUDE.md project overview.

**Source**: `research/investigation/models.md`

Lists all 42 event kinds but does not flag which are webhook-eligible.

## Ground Truth Needed

1. What constant or configuration defines webhook-eligible events?
2. Is the list of 14 from Discovery complete and accurate?
3. What determines webhook eligibility (code pattern vs. configuration)?
4. Are there any conditionally eligible events?

## Investigation Targets

- [ ] File: `orig/loomio/app/models/event.rb` - Look for `WEBHOOK_KINDS` or similar constant
- [ ] File: `orig/loomio/app/models/concerns/events/notify/chatbots.rb` - Check notification logic
- [ ] Command: `grep -rn "webhook\|chatbot" orig/loomio/app/models/events/` - Find webhook-related code in event classes
- [ ] Command: `grep -rn "event_kinds\|WEBHOOK" orig/loomio/config/` - Check for configuration
- [ ] File: `orig/loomio/app/workers/webhook_worker.rb` - Check which events trigger webhooks

## Priority

**MEDIUM** - Webhook integrations are important for:
- Slack/Discord/Teams notifications
- Third-party automation
- Go rewrite must support exact same events

## Rails Context

### Event Notification Pattern

Events typically use concerns to handle different notification channels:

```ruby
# app/models/event.rb
class Event < ApplicationRecord
  include Events::Notify::InApp
  include Events::Notify::Email
  include Events::Notify::Chatbots  # Webhook notifications

  WEBHOOK_ELIGIBLE_KINDS = %w[
    new_discussion discussion_edited
    new_comment
    poll_created poll_edited poll_closing_soon poll_expired poll_closed_by_user
    stance_created stance_updated
    outcome_created outcome_updated
    user_added_to_group membership_requested
  ].freeze
end
```

### Chatbot Concern Pattern

```ruby
# app/models/concerns/events/notify/chatbots.rb
module Events::Notify::Chatbots
  extend ActiveSupport::Concern

  included do
    after_commit :notify_chatbots, if: :webhook_eligible?
  end

  def webhook_eligible?
    Event::WEBHOOK_ELIGIBLE_KINDS.include?(kind)
  end

  def notify_chatbots
    chatbots_for_event.each do |chatbot|
      WebhookWorker.perform_async(chatbot.id, id)
    end
  end
end
```

### Expected 14 Webhook Events

Based on Discovery's list:

| Event Kind | Category | Description |
|------------|----------|-------------|
| `new_discussion` | Discussion | New thread created |
| `discussion_edited` | Discussion | Thread content changed |
| `new_comment` | Comment | Reply posted |
| `poll_created` | Poll | New poll/proposal |
| `poll_edited` | Poll | Poll content changed |
| `poll_closing_soon` | Poll | Approaching deadline |
| `poll_expired` | Poll | Deadline passed |
| `poll_closed_by_user` | Poll | Manually closed |
| `stance_created` | Stance | New vote cast |
| `stance_updated` | Stance | Vote changed |
| `outcome_created` | Outcome | Decision recorded |
| `outcome_updated` | Outcome | Decision modified |
| `user_added_to_group` | Membership | New member |
| `membership_requested` | Membership | Join request |

## Verification Checklist

When investigating, verify:
- [ ] Exact constant/config defining eligible events
- [ ] Whether list is hardcoded or configurable per chatbot
- [ ] Any events that are conditionally eligible
- [ ] Payload format for each event type

## Impact on Go Rewrite

For Go implementation:
- Define webhook-eligible events as a constant/enum
- Ensure WebhookWorker/service handles exactly these 14 events
- Document payload format for each event type
- Consider making list configurable for future extensibility

```go
var WebhookEligibleEvents = map[string]bool{
    "new_discussion":       true,
    "discussion_edited":    true,
    "new_comment":          true,
    "poll_created":         true,
    "poll_edited":          true,
    "poll_closing_soon":    true,
    "poll_expired":         true,
    "poll_closed_by_user":  true,
    "stance_created":       true,
    "stance_updated":       true,
    "outcome_created":      true,
    "outcome_updated":      true,
    "user_added_to_group":  true,
    "membership_requested": true,
}
```
