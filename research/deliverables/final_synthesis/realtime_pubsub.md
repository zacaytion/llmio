# Real-Time Pub/Sub - Implementation Synthesis

## Executive Summary

Loomio uses Redis pub/sub to enable real-time updates via a Node.js Socket.io server. This document describes the Redis channels and message formats for maintaining compatibility with the existing channel server.

---

## Confirmed Architecture

### System Overview

```
┌─────────────┐     Redis Pub/Sub     ┌──────────────────┐
│    Rails    │ ───────────────────▶  │  Channel Server  │
│    (app)    │                       │   (Socket.io)    │
└─────────────┘                       └────────┬─────────┘
       │                                       │
       │ POST /api/hocuspocus                  │ Socket.io
       │                                       │
       ▼                                       ▼
┌─────────────┐                       ┌──────────────────┐
│  Hocuspocus │                       │  Vue.js Client   │
│   (Y.js)    │◀─────────────────────▶│  (Browser)       │
└─────────────┘     Y.js/WebSocket    └──────────────────┘
```

### Redis Pub/Sub Channels

| Channel | Publisher | Subscriber | Payload |
|---------|-----------|------------|---------|
| `/records` | Rails | Channel Server | `{room: "user-123", records: {...}}` |
| `/system_notice` | Rails | Channel Server | `{version: "x.y.z", notice: "...", reload: bool}` |
| `chatbot/*` | Rails | Bots.js | Webhook delivery payloads |

### Socket.io Room Structure

| Room Pattern | Members | Purpose |
|--------------|---------|---------|
| `notice` | All connected clients | System broadcasts |
| `user-{id}` | Single user's connections | Personal notifications |
| `group-{id}` | Group members online | Group activity updates |

---

## MessageChannelService Implementation

### Source Code Reference

From `app/services/message_channel_service.rb`:

```ruby
class MessageChannelService
  def self.publish_models(models, serializer: nil, scope: {}, root: nil, group_id: nil, user_id: nil)
    return if models.blank?
    cache = RecordCache.for_collection(models, user_id)
    data = serialize_models(models, serializer: serializer, scope: scope.merge(cache: cache, current_user_id: user_id), root: root)
    publish_serialized_records(data, group_id: group_id, user_id: user_id)
  end

  def self.publish_serialized_records(data, group_id: nil, user_id: nil)
    CACHE_REDIS_POOL.with do |client|
      room = "user-#{user_id}" if user_id
      room = "group-#{group_id}" if group_id
      client.publish("/records", {room: room, records: data}.to_json)
    end
  end
end
```

---

## Event Publishing Locations

### Verified Call Sites (18+)

| Location | Purpose | Room Type |
|----------|---------|-----------|
| `concerns/events/notify/in_app.rb` | Notification delivery | user |
| `concerns/events/live_update.rb` | Event broadcasts | group |
| `workers/move_comments_worker.rb` | Comment moves | group |
| `controllers/stances_controller.rb` | Stance updates | group + user |
| `services/stance_service.rb` | Vote submission | group |
| `services/discussion_service.rb` | Discussion updates | group |
| `services/poll_service.rb` | Poll creation | group |
| `services/membership_service.rb` | Member changes | group |
| `services/translation_service.rb` | Translation updates | group |
| `services/tag_service.rb` | Tag changes | group |

---

## Hocuspocus Authentication

### Endpoint

`POST /api/hocuspocus`

### Request Format

```json
{
  "user_secret": "123,abc-def-ghi",
  "document_name": "comment-456-body"
}
```

**Token Format:** `{user_id},{secret_token}` (comma-separated)

**Document Naming:** `{record_type}-{record_id}-{field}`

### Supported Record Types (9)

- `comment`, `discussion`, `poll`, `stance`, `outcome`
- `pollTemplate`, `discussionTemplate`
- `group`, `user`

---

## Channel Token Setup (Boot Endpoint)

### Redis Key Pattern

`/current_users/{token}` → User data for Socket.io authentication

### Process

1. Generate or use existing user token
2. Store user data in Redis with key `/current_users/{token}`
3. Set expiration (default: 30 minutes from `CACHE_REDIS_EXPIRE`)
4. Return token in boot payload for frontend to use

---

## Serialization Format

### Rails Output (must match)

```json
{
  "room": "group-123",
  "records": {
    "events": [
      {"id": 1, "kind": "new_comment", "eventable_id": 456, ...}
    ],
    "comments": [
      {"id": 456, "body": "...", "author_id": 789, ...}
    ],
    "users": [
      {"id": 789, "name": "Alice", ...}
    ]
  }
}
```

---

## Configuration

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `REDIS_URL` | Redis connection | `redis://localhost:6379` |
| `CACHE_REDIS_EXPIRE` | Channel token TTL | `1800` (30 min) |
