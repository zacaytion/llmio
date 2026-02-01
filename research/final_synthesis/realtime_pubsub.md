# Real-Time Pub/Sub - Implementation Synthesis

## Executive Summary

Loomio uses Redis pub/sub to enable real-time updates via a Node.js Socket.io server. The Go implementation must publish to the same Redis channels with identical message format to maintain compatibility with the existing channel server.

---

## Confirmed Architecture

### System Overview

```
┌─────────────┐     Redis Pub/Sub     ┌──────────────────┐
│     Go      │ ───────────────────▶  │  Channel Server  │
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
| `/records` | Go | Channel Server | `{room: "user-123", records: {...}}` |
| `/system_notice` | Go | Channel Server | `{version: "x.y.z", notice: "...", reload: bool}` |
| `chatbot/*` | Go | Bots.js | Webhook delivery payloads |

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

### Go Implementation

```go
package realtime

import (
    "context"
    "encoding/json"

    "github.com/redis/go-redis/v9"
)

type MessageChannelService struct {
    redis *redis.Client
}

func NewMessageChannelService(redis *redis.Client) *MessageChannelService {
    return &MessageChannelService{redis: redis}
}

// PublishModels serializes and publishes models to the appropriate room
func (m *MessageChannelService) PublishModels(
    ctx context.Context,
    models []Serializable,
    opts PublishOptions,
) error {
    if len(models) == 0 {
        return nil
    }

    // Serialize models using appropriate serializer
    data := m.serializeModels(models, opts)

    // Determine room from options
    var room string
    if opts.UserID != 0 {
        room = fmt.Sprintf("user-%d", opts.UserID)
    } else if opts.GroupID != 0 {
        room = fmt.Sprintf("group-%d", opts.GroupID)
    }

    return m.publishRecords(ctx, room, data)
}

// PublishRecords publishes serialized data to the /records channel
func (m *MessageChannelService) publishRecords(
    ctx context.Context,
    room string,
    records interface{},
) error {
    payload, err := json.Marshal(map[string]interface{}{
        "room":    room,
        "records": records,
    })
    if err != nil {
        return fmt.Errorf("marshal records: %w", err)
    }

    return m.redis.Publish(ctx, "/records", payload).Err()
}

// PublishSystemNotice broadcasts a system notice to all connected clients
func (m *MessageChannelService) PublishSystemNotice(
    ctx context.Context,
    notice SystemNotice,
) error {
    payload, err := json.Marshal(notice)
    if err != nil {
        return fmt.Errorf("marshal notice: %w", err)
    }

    return m.redis.Publish(ctx, "/system_notice", payload).Err()
}

type PublishOptions struct {
    UserID     int64
    GroupID    int64
    Serializer string
    Root       string
}

type SystemNotice struct {
    Version string `json:"version"`
    Notice  string `json:"notice"`
    Reload  bool   `json:"reload"`
}
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

### Go Integration Points

```go
// After creating an event
func (s *EventService) Publish(ctx context.Context, event *Event) error {
    // ... create event in DB ...

    // Publish to real-time channel
    return s.messageChannel.PublishModels(ctx, []Serializable{event}, PublishOptions{
        GroupID: event.GroupID,
    })
}

// After creating a notification
func (s *NotificationService) Create(ctx context.Context, notification *Notification) error {
    // ... create notification in DB ...

    // Publish to user's personal channel
    return s.messageChannel.PublishModels(ctx, []Serializable{notification}, PublishOptions{
        UserID: notification.UserID,
    })
}
```

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

### Go Handler

```go
func (h *HocuspocusHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserSecret   string `json:"user_secret"`
        DocumentName string `json:"document_name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Parse token: "123,abc-def-ghi"
    parts := strings.SplitN(req.UserSecret, ",", 2)
    if len(parts) != 2 {
        http.Error(w, "Invalid token format", http.StatusUnauthorized)
        return
    }

    userID, err := strconv.ParseInt(parts[0], 10, 64)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusUnauthorized)
        return
    }
    secretToken := parts[1]

    // Verify user
    user, err := h.userRepo.FindByIDAndSecret(r.Context(), userID, secretToken)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Parse document: "comment-456-body"
    docParts := strings.SplitN(req.DocumentName, "-", 3)
    if len(docParts) < 2 {
        http.Error(w, "Invalid document name", http.StatusBadRequest)
        return
    }

    recordType := docParts[0]
    recordID, _ := strconv.ParseInt(docParts[1], 10, 64)

    // Check permission
    if !h.canEditDocument(r.Context(), user, recordType, recordID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

---

## Channel Token Setup (Boot Endpoint)

### Redis Key Pattern

`/current_users/{token}` → User data for Socket.io authentication

### Go Implementation

```go
func (b *BootHandler) SetupChannelAuth(ctx context.Context, user *User) (string, error) {
    token := user.SecretToken // or generate specific channel token

    userData, err := json.Marshal(map[string]interface{}{
        "id":        user.ID,
        "name":      user.Name,
        "group_ids": user.GroupIDs(),
    })
    if err != nil {
        return "", fmt.Errorf("marshal user data: %w", err)
    }

    key := fmt.Sprintf("/current_users/%s", token)
    expiration := time.Duration(b.config.CacheRedisExpire) * time.Second

    if err := b.redis.Set(ctx, key, userData, expiration).Err(); err != nil {
        return "", fmt.Errorf("set channel token: %w", err)
    }

    return token, nil
}
```

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

### Go Serializer Pattern

```go
type RecordSet struct {
    Events      []EventJSON      `json:"events,omitempty"`
    Discussions []DiscussionJSON `json:"discussions,omitempty"`
    Comments    []CommentJSON    `json:"comments,omitempty"`
    Users       []UserJSON       `json:"users,omitempty"`
    Groups      []GroupJSON      `json:"groups,omitempty"`
    // ... other record types
}

func (m *MessageChannelService) serializeModels(models []Serializable, opts PublishOptions) RecordSet {
    var result RecordSet

    for _, model := range models {
        switch v := model.(type) {
        case *Event:
            result.Events = append(result.Events, v.ToJSON())
            // Include sideloaded records
            result.Users = append(result.Users, v.Author.ToJSON())
        case *Comment:
            result.Comments = append(result.Comments, v.ToJSON())
        // ... handle other types
        }
    }

    return result
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestPublishModels(t *testing.T) {
    redis := miniredis.RunT(t)
    client := redis.NewClient()
    service := NewMessageChannelService(client)

    // Subscribe to channel
    sub := client.Subscribe(context.Background(), "/records")

    // Publish event
    event := &Event{ID: 1, Kind: "new_comment", GroupID: 123}
    err := service.PublishModels(context.Background(), []Serializable{event}, PublishOptions{
        GroupID: 123,
    })
    require.NoError(t, err)

    // Verify message
    msg, err := sub.ReceiveMessage(context.Background())
    require.NoError(t, err)

    var payload map[string]interface{}
    json.Unmarshal([]byte(msg.Payload), &payload)
    assert.Equal(t, "group-123", payload["room"])
}
```

### Integration Tests

```go
func TestChannelServerIntegration(t *testing.T) {
    // Start real Redis
    // Start channel server
    // Connect Socket.io client
    // Publish from Go
    // Verify client receives message
}
```

---

## Configuration

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `REDIS_URL` | Redis connection | `redis://localhost:6379` |
| `CACHE_REDIS_EXPIRE` | Channel token TTL | `1800` (30 min) |

### Redis Connection Pool

```go
func NewRedisClient(cfg *Config) *redis.Client {
    opt, _ := redis.ParseURL(cfg.RedisURL)
    opt.PoolSize = 20
    opt.MinIdleConns = 5
    return redis.NewClient(opt)
}
```
