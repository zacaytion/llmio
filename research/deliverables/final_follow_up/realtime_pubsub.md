# Real-Time Pub/Sub - Follow-up Analysis

## Executive Summary

The real-time architecture was documented in `research/synthesis/realtime_architecture.md` but **was not included in the third-party follow-up investigation**. This document captures open questions that need clarification before implementation.

---

## Source Code Verification

Verified `MessageChannelService` at `app/services/message_channel_service.rb`:

```ruby
def self.publish_serialized_records(data, group_id: nil, user_id: nil)
  CACHE_REDIS_POOL.with do |client|
    room = "user-#{user_id}" if user_id
    room = "group-#{group_id}" if group_id
    client.publish("/records", {room: room, records: data}.to_json)
  end
end
```

**18+ locations** call `MessageChannelService.publish_models`:
- `concerns/events/notify/in_app.rb` - Notification delivery
- `concerns/events/live_update.rb` - Event publishing
- `workers/move_comments_worker.rb` - Comment moves
- `controllers/api/v1/stances_controller.rb` - Stance updates
- `services/stance_service.rb`, `discussion_service.rb`, `poll_service.rb`, etc.

---

## Open Questions for Third Party

### HIGH Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 1 | **Which event kinds trigger pub/sub publishing?** | Real-time feature parity | Need complete mapping of EventKind → pub/sub |
| 2 | **Is there room-based filtering beyond user/group?** | Performance, privacy | Could `group_id` nil trigger broadcast? |
| 3 | **What determines whether an event uses user_id vs group_id routing?** | Message delivery logic | `concerns/events/live_update.rb` logic |

### MEDIUM Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 4 | What is the expected latency for pub/sub delivery? | Performance expectations | Any SLA or monitoring? |
| 5 | Are there backpressure mechanisms if Redis pub/sub backs up? | Reliability | Connection pool behavior |
| 6 | How does `chatbot/*` pub/sub differ from `/records`? | Bot integration | `chatbot_service.rb` publishing |

### LOW Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 7 | Is there monitoring for pub/sub message delivery failures? | Observability | Logging/metrics in channel server |
| 8 | What happens if Channel Server is down? | Degraded mode | Client reconnection behavior |

---

## Discrepancies Identified

### Room Routing Logic

**Observed in source** (`live_update.rb`):
```ruby
MessageChannelService.publish_models([self], group_id: eventable.group.id)
# vs
MessageChannelService.publish_models([self], user_id: user.id)
```

**Question:** When is `user_id` used vs `group_id`? The logic appears to be:
- Notifications → `user_id` routing (personal)
- Events → `group_id` routing (group members)

But this needs verification - can both be set simultaneously?

### Event → Pub/Sub Mapping

**Not documented:** Which of the 42 event kinds trigger real-time updates?

From `concerns/events/live_update.rb`:
```ruby
module Events
  module LiveUpdate
    def trigger_live_update
      MessageChannelService.publish_models([self], group_id: eventable.group.id)
    end
  end
end
```

**Question:** Which events include `LiveUpdate` concern? Is it all 42 or a subset?

---

## Files Requiring Investigation

| File | Purpose | Priority |
|------|---------|----------|
| `app/models/concerns/events/live_update.rb` | Event pub/sub trigger | HIGH |
| `app/models/concerns/events/notify/in_app.rb` | Notification pub/sub | HIGH |
| `app/services/message_channel_service.rb` | Core pub/sub service | VERIFIED |
| `loomio_channel_server/src/records.js` | Redis subscriber | MEDIUM |
| `app/services/chatbot_service.rb` | Bot pub/sub | LOW |

---

## Implementation Requirements

### Must Implement

1. **MessageChannelService equivalent** with Redis pub/sub
2. **Room routing logic** matching Rails behavior
3. **RecordCache serialization** for efficient payloads

### Compatibility with Node.js Channel Server

The Node.js channel server expects:
```json
{
  "room": "group-123",
  "records": {
    "events": [...],
    "discussions": [...],
    "users": [...]
  }
}
```

Implementation must produce identical JSON structure for channel server compatibility.

---

## Priority Assessment

| Area | Priority | Blocking? |
|------|----------|-----------|
| Event → pub/sub mapping | HIGH | Yes - affects what clients receive |
| Room routing logic | HIGH | Yes - affects message delivery |
| Chatbot pub/sub | LOW | No - separate integration |
