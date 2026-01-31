# Real-Time System

> Channel server, Redis pub/sub, and WebSocket protocols.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Rails App     │───▶│     Redis       │◀───│ Channel Server  │
│                 │    │   (pub/sub)     │    │   (Node.js)     │
└─────────────────┘    └─────────────────┘    └────────┬────────┘
                                                       │
                                   ┌───────────────────┴───────────────────┐
                                   │                                       │
                            ┌──────▼──────┐                         ┌──────▼──────┐
                            │  Socket.io  │                         │ Hocuspocus  │
                            │  (records)  │                         │ (editing)   │
                            └─────────────┘                         └─────────────┘
```

## Channel Server

**Source:** `orig/loomio_channel_server/`

Two separate processes:
1. **records.js** - Socket.io server for live updates (port 5000)
2. **hocuspocus.mjs** - Yjs WebSocket for collaborative editing

**Port Configuration:**
```javascript
// orig/loomio_channel_server/hocuspocus.mjs:21
const port = (process.env.RAILS_ENV == 'production') ? 5000 : 4444
```
- Production: port 5000 (in separate container)
- Development: port 4444 (to avoid conflicts)

**Hocuspocus Server Config:**
```javascript
// orig/loomio_channel_server/hocuspocus.mjs:25-30
const server = new Server({
  port: port,
  timeout: 30000,      // 30s connection timeout
  debounce: 5000,      // 5s debounce before persistence
  maxDebounce: 30000,  // 30s max before forced persistence
  quiet: true,
  // ...
});
```

## Redis Pub/Sub Channels

### /records

**Publisher:** Rails (`MessageChannelService.publish_serialized_records`)

**Subscriber:** records.js

**Payload:**
```json
{
  "room": "group-123",
  "records": {
    "discussions": [...],
    "comments": [...]
  }
}
```

**Room Types:**
- `user-{id}` - User-specific updates
- `group-{id}` - Group-wide updates

### /system_notice

**Publisher:** Rails (`MessageChannelService.publish_system_notice`)

**Subscriber:** records.js (broadcasts to ALL sockets)

**Payload:**
```json
{
  "type": "maintenance",
  "message": "System down in 5 minutes"
}
```

### chatbot/*

**Publisher:** Rails (`ChatbotService`)

**Subscriber:** bots.js

**Channels:**
- `chatbot/test` - Test bot configuration (creates new client each time)
- `chatbot/publish` - Send messages to Matrix rooms (cached clients)

**Client Caching Pattern:**
```javascript
// chatbot/publish caches clients by config key
const key = JSON.stringify(params.config)
if (!bots[key]) {
  bots[key] = new MatrixClient(params.config.server, params.config.access_token);
}
```

**Note:** `chatbot/test` creates a new client per request while `chatbot/publish` caches by config. This caching has no eviction strategy—potential memory concern if many unique configs are used.

**Source:** `orig/loomio_channel_server/bots.js:37-47`

## Redis Key: /current_users/{token}

**Writer:** Rails (`BootController.set_channel_token`)

**Reader:** records.js (WebSocket authentication)

**Value:**
```json
{
  "id": 123,
  "name": "User Name",
  "group_ids": [1, 2, 3]
}
```

**Expiration:** Set via `CACHE_REDIS_EXPIRE` env var

**Rails Code:**
```ruby
# orig/loomio/app/controllers/api/v1/boot_controller.rb:26-32
def set_channel_token
  CACHE_REDIS_POOL.with do |client|
    client.set("/current_users/#{current_user.secret_token}",
      {name: current_user.name,
       group_ids: current_user.group_ids,
       id: current_user.id}.to_json)
  end
end
```

## Socket.io Connection Flow

1. Client connects with `channel_token` query param
2. Server reads `/current_users/{token}` from Redis
3. If valid, socket joins rooms:
   - `notice` (all sockets)
   - `user-{id}`
   - `group-{id}` for each group membership
4. Server relays Redis `/records` to appropriate rooms

**Connection State Recovery:** 30-minute window for reconnection.

## Hocuspocus Flow

### Authentication

1. Client connects with token and document name
2. Server POSTs to Rails `/api/hocuspocus`:
   ```json
   {"user_secret": "token", "document_name": "comment-123-body"}
   ```
3. Rails validates and returns 200 or error

### Token Format

Assembled by **client**: `{user_id},{secret_token}`

Example: `123,abc-def-ghi-jkl`

### Document Naming

Pattern: `{record_type}-{record_id}` or `{record_type}-{record_id}-{field}`

**Record Types:**
- comment, discussion, poll, stance, outcome
- pollTemplate, discussionTemplate, group, user

**Examples:**
- `comment-456-body`
- `discussion-789-description`

### Ephemeral Storage

**Important:** Hocuspocus uses `SQLite({database: ''})` - intentionally ephemeral.

**Why:** Rails DB is the source of truth. Hocuspocus only handles real-time sync during active editing. Content is saved to Rails via API when user saves.

## Rails → Redis Publishing

**Source:** `orig/loomio/app/services/message_channel_service.rb`

```ruby
module MessageChannelService
  def self.publish_serialized_records(data, group_id: nil, user_id: nil)
    CACHE_REDIS_POOL.with do |client|
      room = "user-#{user_id}" if user_id
      room = "group-#{group_id}" if group_id
      client.publish("/records", {room: room, records: data}.to_json)
    end
  end

  def self.publish_system_notice(message)
    CACHE_REDIS_POOL.with do |client|
      client.publish("/system_notice", message.to_json)
    end
  end
end
```

## Event → Real-time Mapping

Not all events trigger real-time updates. The pattern:

1. Service creates event: `Event.create!(kind: 'new_comment', ...)`
2. EventBus broadcasts: `EventBus.broadcast('new_comment_event', event)`
3. Listener publishes to Redis (if applicable)

**Open Question:** Which events trigger Redis pub/sub? See [questions.md](./questions.md).

## Go Implementation Notes

**Socket.io Options:**
- `github.com/googollee/go-socket.io` - Protocol compatible
- Custom WebSocket + JSON (simpler, no protocol baggage)

**Hocuspocus Options:**
- Keep Node.js hocuspocus as separate service
- Implement subset of Yjs protocol
- Use Automerge (has Go bindings)

**Key Considerations:**
1. Redis pub/sub is standard - use go-redis
2. Room-based broadcasting can use sync.Map of channels
3. Connection state recovery needs session storage
4. Hocuspocus complexity may warrant keeping Node.js

---
