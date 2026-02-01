# Webhook-Eligible Events: Complete Enumeration

## Executive Summary

The Loomio codebase has **16 event classes** that include webhook/chatbot notification capability, but only **14 events** are exposed to users in the webhook configuration UI. This discrepancy exists by design - the 2 additional events (`discussion_announced`, `poll_announced`) are "announcement" events that trigger chatbots via the `recipient_chatbot_ids` mechanism rather than event-kind filtering.

## Ground Truth Answers

### 1. What constant or configuration defines webhook-eligible events?

**Answer**: There are TWO mechanisms:

**A. UI-Exposed Event Kinds (User-Selectable)**
- **File**: `/Users/z/Code/loomio/config/webhook_event_kinds.yml`
- **Access**: `AppConfig.webhook_event_kinds`
- **Consumed by**:
  - `app/models/boot/site.rb:30` - Sent to frontend as `webhookEventKinds`
  - `vue/src/components/chatbot/webhook_form.vue:14` - Populates checkbox list

**B. Chatbot-Capable Events (Code-Level)**
- **Pattern**: Events that `include Events::Notify::Chatbots` concern
- **File**: `/Users/z/Code/loomio/app/models/concerns/events/notify/chatbots.rb`
- **Behavior**: When triggered, schedules `ChatbotService.publish_event!` via GenericWorker

### 2. Is the list of 14 from Discovery complete and accurate?

**Partially Accurate**. Discovery's list had 2 errors:

| Discovery Claimed | Actual Status | Notes |
|-------------------|---------------|-------|
| `user_added_to_group` | NOT webhook-eligible | Does NOT include Chatbots concern |
| `membership_requested` | NOT webhook-eligible | Does NOT include Chatbots concern |
| `poll_reopened` | Missing from Discovery | IS webhook-eligible |
| `outcome_review_due` | Missing from Discovery | IS webhook-eligible |

### 3. What determines webhook eligibility (code pattern vs. configuration)?

**Answer**: Both, in a layered system:

1. **Code Pattern (Required)**: Event class MUST `include Events::Notify::Chatbots` to have any chatbot capability
2. **Configuration (User Selection)**: Events in `webhook_event_kinds.yml` appear in the UI for user selection
3. **Runtime Filtering**: `ChatbotService.publish_event!` filters by either:
   - Event kind matching chatbot's `event_kinds` array, OR
   - Event having chatbot in `recipient_chatbot_ids` custom field

### 4. Are there any conditionally eligible events?

**Yes**, two categories:

**A. UI-Hidden but Chatbot-Capable (via `recipient_chatbot_ids` only):**
- `discussion_announced` - Only triggered when users are invited to discussion
- `poll_announced` - Only triggered when users are invited to poll
- `poll_reminder` - Only triggered when poll author sends reminder

**B. Conditionally Triggered:**
- `outcome_review_due` - Only fires when outcome has `review_on` date set

## Complete Event Classification

### Tier 1: UI-Exposed Webhook Events (14 events)

These appear in the webhook configuration UI and can be selected by users:

| Event Kind | Category | Eventable Type | File |
|------------|----------|----------------|------|
| `new_discussion` | Discussion | Discussion | `app/models/events/new_discussion.rb` |
| `discussion_edited` | Discussion | Discussion | `app/models/events/discussion_edited.rb` |
| `new_comment` | Comment | Comment | `app/models/events/new_comment.rb` |
| `poll_created` | Poll | Poll | `app/models/events/poll_created.rb` |
| `poll_edited` | Poll | Poll | `app/models/events/poll_edited.rb` |
| `poll_closing_soon` | Poll Lifecycle | Poll | `app/models/events/poll_closing_soon.rb` |
| `poll_expired` | Poll Lifecycle | Poll | `app/models/events/poll_expired.rb` |
| `poll_closed_by_user` | Poll Lifecycle | Poll | `app/models/events/poll_closed_by_user.rb` |
| `poll_reopened` | Poll Lifecycle | Poll | `app/models/events/poll_reopened.rb` |
| `stance_created` | Vote | Stance | `app/models/events/stance_created.rb` |
| `stance_updated` | Vote | Stance | `app/models/events/stance_updated.rb` |
| `outcome_created` | Outcome | Outcome | `app/models/events/outcome_created.rb` |
| `outcome_updated` | Outcome | Outcome | `app/models/events/outcome_updated.rb` |
| `outcome_review_due` | Outcome | Outcome | `app/models/events/outcome_review_due.rb` |

### Tier 2: Chatbot-Capable but UI-Hidden (2 events)

These can trigger chatbots only via `recipient_chatbot_ids` mechanism:

| Event Kind | Category | Trigger Condition |
|------------|----------|-------------------|
| `discussion_announced` | Announcement | When users explicitly invited to discussion |
| `poll_announced` | Announcement | When users explicitly invited to poll |

### Tier 3: Chatbot-Capable, Test/Admin Only (1 event)

| Event Kind | Category | Notes |
|------------|----------|-------|
| `poll_reminder` | Poll | Used when author sends manual reminder |

### Tier 4: NOT Webhook-Eligible (Discovery Errors)

These events do NOT include the Chatbots concern:

| Event Kind | Reason |
|------------|--------|
| `user_added_to_group` | Only notifies via InApp/ByEmail |
| `membership_requested` | Only notifies via InApp/ByEmail |

## Event-to-Chatbot Routing Logic

From `app/services/chatbot_service.rb:29-31`:

```ruby
chatbots.where(id: event.recipient_chatbot_ids).
    or(chatbots.where.any(event_kinds: event.kind)).each do |chatbot|
```

This means chatbots receive events when EITHER:
1. Chatbot ID is in event's `recipient_chatbot_ids` (explicit targeting), OR
2. Event's kind matches one of chatbot's configured `event_kinds` (subscription model)

## Template Resolution

From `app/services/chatbot_service.rb:33-36`:

```ruby
template_name = event.eventable_type.tableize.singularize
template_name = 'poll' if event.eventable_type == 'Outcome'
template_name = 'group' if event.eventable_type == 'Membership'
template_name = 'notification' if chatbot.notification_only
```

Template mapping:
- Discussion events -> `chatbot/{format}/discussion.text.erb`
- Comment events -> `chatbot/{format}/comment.text.erb`
- Poll/Outcome events -> `chatbot/{format}/poll.text.erb`
- Stance events -> `chatbot/{format}/stance.text.erb`
- Membership events -> `chatbot/{format}/group.text.erb`
- Notification-only mode -> `chatbot/{format}/notification.text.erb`

## Authoritative Source Locations

| Item | Location | Line |
|------|----------|------|
| UI Event Kinds Config | `config/webhook_event_kinds.yml` | 1-14 |
| Frontend Event Kinds | `app/models/boot/site.rb` | 30 |
| Chatbots Concern | `app/models/concerns/events/notify/chatbots.rb` | 1-6 |
| Chatbot Service | `app/services/chatbot_service.rb` | 22-70 |
| Chatbot Model | `app/models/chatbot.rb` | 1-18 |
| Webhook Serializers | `app/serializers/webhook/*/event_serializer.rb` | varies |

## Conclusion

The "14 webhook-eligible events" claim from Discovery is **correct for user-facing configuration**, but there are actually **16-17 event classes** with chatbot capability at the code level. The distinction is:
- 14 = Events users can select in webhook UI
- 16 = Events that `include Events::Notify::Chatbots` (excludes `stance_updated` which inherits)
- 17 = Including `stance_updated` via inheritance from `stance_created`
