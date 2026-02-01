# Webhook Events Investigation

## Purpose

This investigation was conducted to verify the accuracy of Discovery claims about webhook-eligible events in Loomio. Both Discovery and Research agreed that 14 events were webhook-eligible, but neither document fully enumerated all events.

## Key Findings

### Summary

| Category | Count |
|----------|-------|
| UI-exposed webhook events | 14 |
| Code-level chatbot-capable events | 16 |
| Discovery claims verified correct | 12 of 14 |
| Discovery claims verified incorrect | 2 |
| Missing events found | 2 |

### Corrections to Discovery

**Incorrectly claimed as webhook-eligible (should be removed):**
- `user_added_to_group` - Does NOT include Chatbots concern
- `membership_requested` - Does NOT include Chatbots concern

**Missing from Discovery (should be added):**
- `poll_reopened` - IS webhook-eligible
- `outcome_review_due` - IS webhook-eligible

### Authoritative Source

The single source of truth for user-selectable webhook events is:
```
/Users/z/Code/loomio/config/webhook_event_kinds.yml
```

## Documents in This Directory

| File | Description |
|------|-------------|
| [findings.md](./findings.md) | Complete enumeration of all webhook-eligible events with evidence |
| [models.md](./models.md) | Event model documentation with webhook flags |
| [services.md](./services.md) | ChatbotService and delivery mechanism documentation |
| [payloads.md](./payloads.md) | Webhook payload format for each event type |
| [confidence.md](./confidence.md) | Verification checklist with PASS/FAIL per claim |

## Quick Reference

### Complete Webhook-Eligible Event List (14 events)

From `config/webhook_event_kinds.yml`:

1. `new_discussion`
2. `discussion_edited`
3. `new_comment`
4. `poll_created`
5. `poll_edited`
6. `poll_closing_soon`
7. `poll_expired`
8. `poll_closed_by_user`
9. `poll_reopened`
10. `stance_created`
11. `stance_updated`
12. `outcome_created`
13. `outcome_updated`
14. `outcome_review_due`

### Additional Chatbot-Capable Events (not in UI)

These events include the Chatbots concern but are not exposed in the webhook configuration UI. They can only be triggered via `recipient_chatbot_ids`:

- `discussion_announced`
- `poll_announced`
- `poll_reminder`

## Code Entry Points

| Component | File | Line |
|-----------|------|------|
| Config YAML | `config/webhook_event_kinds.yml` | 1-14 |
| AppConfig accessor | `app/extras/app_config.rb` | 3 |
| Frontend boot data | `app/models/boot/site.rb` | 30 |
| Chatbots concern | `app/models/concerns/events/notify/chatbots.rb` | 1-6 |
| ChatbotService | `app/services/chatbot_service.rb` | 22-70 |
| Chatbot model | `app/models/chatbot.rb` | 1-18 |

## Investigation Date

February 1, 2026

## Confidence Level

**Overall confidence: 5/5**

All claims verified by direct code inspection with file:line references.
