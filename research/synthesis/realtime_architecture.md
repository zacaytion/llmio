# Real-time Architecture - Confirmed Architecture

## Summary

Both Discovery and Research documentation **fully agree** on the real-time system architecture. This is a multi-service system using Redis pub/sub to connect Rails, Socket.io (Node.js), and Hocuspocus (collaborative editing).

## Key Details

### Architecture Overview

```
┌─────────────┐     Redis Pub/Sub     ┌──────────────────┐
│   Rails     │ ───────────────────▶  │  Channel Server  │
│   (app)     │                       │   (Socket.io)    │
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

Both sources confirm these channels:

| Channel | Publisher | Subscriber | Payload |
|---------|-----------|------------|---------|
| `/records` | Rails | Channel Server | `{room: "user-123", records: {...}}` |
| `/system_notice` | Rails | Channel Server | `{type: "...", message: "..."}` |
| `chatbot/test` | Rails | Bots.js | Matrix client creation per request |
| `chatbot/publish` | Rails | Bots.js | Cached Matrix client by config key |

### Redis Keys

Both sources confirm:

| Key Pattern | Purpose | Value | TTL |
|-------------|---------|-------|-----|
| `/current_users/{token}` | WebSocket auth | `{id, name, group_ids}` | `CACHE_REDIS_EXPIRE` |

### Socket.io Room Structure

Both sources confirm room-based message routing:

| Room Pattern | Members | Purpose |
|--------------|---------|---------|
| `notice` | All connected clients | System broadcasts |
| `user-{id}` | Single user's connections | Personal notifications |
| `group-{id}` | Group members online | Group activity updates |

### Socket.io Connection Flow

1. Client connects with `channel_token` query parameter
2. Server reads `/current_users/{token}` from Redis
3. Server joins client to rooms: `notice`, `user-{id}`, `group-{id}` (each group)
4. Server relays Redis `/records` messages to appropriate rooms
5. 30-minute reconnection window for token validity

### Hocuspocus Authentication

Both sources confirm (Research has more detail):

**Auth Endpoint**: `POST /api/hocuspocus`

**Request**:
```json
{
  "user_secret": "123,abc-def-ghi",
  "document_name": "comment-456-body"
}
```

**Token Format**: `{user_id},{secret_token}` (client-assembled)

**Document Naming**: `{record_type}-{record_id}-{field}`

**Supported Record Types** (9):
- `comment`, `discussion`, `poll`, `stance`, `outcome`
- `pollTemplate`, `discussionTemplate`
- `group`, `user`

### Hocuspocus Storage

Both sources confirm **intentionally ephemeral**:

```javascript
// SQLite with empty string = in-memory only
const server = new Hocuspocus({
  storage: new SQLite({ database: '' })
})
```

**Rationale**: Rails database is source of truth. Hocuspocus only handles real-time sync. Content persists when user saves via Rails API.

### Port Configuration

| Service | Production | Development |
|---------|------------|-------------|
| Socket.io (records.js) | 5000 | 5000 |
| Hocuspocus | 5000 | 4444 |

Note: Same port in production because they run as separate Docker containers.

## Source Alignment

| Aspect | Discovery | Research | Status |
|--------|-----------|----------|--------|
| Redis channels | 3 main | 3 main + chatbot/* | ✅ Research more complete |
| `/current_users/{token}` | Documented | Documented | ✅ Confirmed |
| Socket.io rooms | 3 patterns | 3 patterns | ✅ Confirmed |
| Hocuspocus auth endpoint | `/api/hocuspocus` | `/api/hocuspocus` | ✅ Confirmed |
| Token format | Not documented | `{user_id},{secret_token}` | ✅ Research more complete |
| Document naming | Not documented | `{record_type}-{record_id}-{field}` | ✅ Research more complete |
| Ephemeral storage | "unclear" | Intentional design | ✅ Research clarifies |

## Implementation Notes

### Go Real-time Options

**Option 1: Keep Node.js Channel Server**
- Minimal change, proven working
- Go publishes to same Redis channels
- Only need to implement `/api/hocuspocus` endpoint

**Option 2: Go Socket.io Server**
- Use `googollee/go-socket.io` (protocol-compatible)
- Implement room management with `sync.Map`
- Subscribe to Redis pub/sub

**Option 3: Pure WebSocket**
- Use `nhooyr.io/websocket` (approved in CLAUDE.md)
- Custom protocol, not Socket.io compatible
- Requires Vue client changes

### Go Redis Publishing

```go
// MessageChannelService equivalent
type MessageChannel struct {
    redis *redis.Client
}

func (m *MessageChannel) PublishRecords(room string, records interface{}) error {
    payload, _ := json.Marshal(map[string]interface{}{
        "room":    room,
        "records": records,
    })
    return m.redis.Publish(ctx, "/records", payload).Err()
}

func (m *MessageChannel) PublishSystemNotice(notice SystemNotice) error {
    payload, _ := json.Marshal(notice)
    return m.redis.Publish(ctx, "/system_notice", payload).Err()
}
```

### Go Hocuspocus Auth Endpoint

```go
// POST /api/hocuspocus
func (h *HocuspocusHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserSecret   string `json:"user_secret"`
        DocumentName string `json:"document_name"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Parse token: "123,abc-def-ghi" -> user_id=123, secret="abc-def-ghi"
    parts := strings.SplitN(req.UserSecret, ",", 2)
    userID, _ := strconv.ParseInt(parts[0], 10, 64)
    secretToken := parts[1]

    // Verify user
    user, err := h.userRepo.FindByIDAndSecret(userID, secretToken)
    if err != nil {
        http.Error(w, "Unauthorized", 401)
        return
    }

    // Parse document: "comment-456-body" -> type=comment, id=456, field=body
    docParts := strings.Split(req.DocumentName, "-")
    recordType := docParts[0]
    recordID, _ := strconv.ParseInt(docParts[1], 10, 64)

    // Check permission
    if !h.canEditDocument(user, recordType, recordID) {
        http.Error(w, "Forbidden", 403)
        return
    }

    w.WriteHeader(200)
}
```

### User Token Setup (Boot Endpoint)

```go
// During /api/v1/boot response
func (b *BootHandler) SetupChannelAuth(user *User) string {
    token := user.SecretToken // or generate new channel token

    // Store in Redis for Channel Server to read
    userData, _ := json.Marshal(map[string]interface{}{
        "id":        user.ID,
        "name":      user.Name,
        "group_ids": user.GroupIDs(),
    })

    key := fmt.Sprintf("/current_users/%s", token)
    b.redis.Set(ctx, key, userData, cacheExpiration)

    return token
}
```
