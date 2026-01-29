# Loomio Codebase Discovery Report

**Date:** 2026-01-29
**Source:**
- github.com/loomio/loomio (shallow clone at `orig/loomio`)
- github.com/loomio/loomio_channel_server (shallow clone at `orig/loomio_channel_server`)

**Purpose:** Phase 1 Discovery for Rails → Go rewrite planning

---

## Executive Summary

Loomio is a mature, actively-maintained Rails 8 application for collaborative decision-making. The codebase is **moderately complex** with well-organized architecture but significant depth in several areas:

| Metric | Value | Migration Impact |
|--------|-------|------------------|
| Ruby files | 1,546 | High volume |
| Vue components | 216 | Frontend retained |
| Database tables | 56 | Schema migration needed |
| Migrations | 941 | 10+ years of evolution |
| Background workers | 38 | Job system needed |
| API controllers | 30+ | Contract preservation critical |
| Channel server | ~150 LOC Node.js | Real-time infrastructure |

**Key Finding:** The frontend (Vue 3) is well-decoupled via API. The rewrite can preserve the frontend entirely if API contracts are maintained.

**Architecture Note:** Loomio is a **two-service system**: the Rails app + a separate Node.js channel server for real-time features. Both must be understood for a complete migration.

---

## 1. Repository Structure

```
loomio/                     # 38MB, 4,368 files
├── app/                    # Rails application code
│   ├── models/            # 53+ models (4,239 LOC total)
│   ├── controllers/       # 24 top-level + api/v1 (30+)
│   ├── services/          # 44+ service objects
│   ├── workers/           # 38 Sidekiq workers
│   ├── serializers/       # 49 API serializers
│   ├── mailers/           # 7 mailers
│   ├── queries/           # 10 query objects
│   └── views/             # HAML templates (minimal, API-first)
├── vue/                    # Vue 3 frontend (separate build)
│   └── src/components/    # 216 components
├── db/
│   ├── schema.rb          # 1,093 lines, 56 tables
│   └── migrate/           # 941 migrations
├── config/
│   ├── routes.rb          # 471 lines
│   └── locales/           # 20+ languages
└── spec/                   # 109 RSpec test files
```

**Stack Versions:**
- Ruby 3.4.7
- Rails 8.0.0
- PostgreSQL (with extensions)
- Vue 3.5 + Vuetify + Vite

---

## 2. Domain Model

### Core Entities

```
User ─┬─< Membership >─── Group
      │                     │
      │                     ├─< Discussion ─< Comment
      │                     │       │
      │                     │       └─< Event (threaded)
      │                     │
      │                     └─< Poll ─< Stance
      │                            │       │
      │                            │       └─< StanceChoice
      │                            │
      │                            └─< Outcome
      │
      └─< Notification
```

### Key Models by Complexity

| Model | Lines | Associations | Notes |
|-------|-------|--------------|-------|
| User | 377 | 25+ | Core identity, Devise auth |
| Group | ~300 | 20+ | Hierarchical (parent_id), extensive settings |
| Discussion | ~250 | 15+ | Threaded via Event, templates |
| Poll | ~350 | 15+ | Multiple poll types, voting logic |
| Event | ~200 | 10+ | Event-sourcing pattern, threading |
| Membership | ~150 | 8+ | Roles, permissions, delegation |

### PostgreSQL-Specific Features

The schema uses several PostgreSQL features that need careful Go mapping:

```ruby
# Extensions required
enable_extension "citext"           # Case-insensitive text
enable_extension "hstore"           # Key-value storage
enable_extension "pgcrypto"         # Cryptographic functions
enable_extension "pg_stat_statements"

# Column types requiring attention
t.jsonb "custom_fields"             # 15+ tables use JSONB
t.string "tags", array: true        # Array columns
t.citext "handle"                   # Case-insensitive handles
```

---

## 3. API Structure

### Versioned APIs

| Namespace | Purpose | Complexity |
|-----------|---------|------------|
| `/api/v1/*` | Primary API (30+ controllers) | High |
| `/api/b1/*` | Bot API v1 | Low |
| `/api/b2/*` | Bot API v2 (adds comments) | Low |
| `/api/b3/*` | Bot API v3 (user management) | Low |
| `/api/hocuspocus` | Real-time collaboration auth | Medium |

### Key API Patterns

- **RESTful resources** with custom actions
- **Serializers** (49) define API contracts
- **Pagination** via standard Rails patterns
- **Authentication** via Devise session + tokens

### Sample Route Complexity

```ruby
resources :groups do
  get :token, :subgroups, :export, :export_csv
  post :reset_token, :archive, :upload_document
  resources :memberships
  resources :discussion_templates
  resources :poll_templates
end
```

---

## 4. Background Job System

**Queue:** Sidekiq 7.0 with Redis

### Worker Categories (38 total)

| Category | Count | Examples |
|----------|-------|----------|
| Email/Notifications | 8 | `SendDailyCatchUpEmailWorker`, `PublishEventWorker` |
| Data Cleanup | 6 | `CleanupService`, `DestroyGroupWorker` |
| Async Operations | 10 | `GroupExportWorker`, `GeoLocationWorker` |
| Migrations/Repairs | 14 | `MigrateTagsWorker`, `RepairThreadWorker` |

**Go Equivalent Options:** Asynq, River, or Machinery (all support PostgreSQL or Redis backends)

---

## 5. Real-Time Features & Channel Server

Loomio's real-time features are provided by a **separate Node.js service**: `loomio_channel_server`. This is critical infrastructure that must be migrated or preserved.

### Channel Server Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        loomio_channel_server                        │
│                         (~150 lines Node.js)                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐  │
│  │  records.js │    │   bots.js   │    │    hocuspocus.mjs       │  │
│  │  Socket.io  │    │ Matrix Bot  │    │  Collaborative Editing  │  │
│  │  Port 5000  │    │             │    │   Port 5000/4444        │  │
│  └──────┬──────┘    └──────┬──────┘    └───────────┬─────────────┘  │
│         │                  │                       │                │
│         └────────┬─────────┘                       │                │
│                  │                                 │                │
│           ┌──────▼──────┐                   ┌──────▼──────┐         │
│           │    Redis    │                   │    Rails    │         │
│           │  Pub/Sub    │                   │  /api/hocus │         │
│           └──────┬──────┘                   │   pocus     │         │
│                  │                          │  (auth only)│         │
└──────────────────│──────────────────────────└─────────────┘─────────┘
                   │
            ┌──────▼──────┐
            │  Rails App  │
            │  (publishes │
            │   events)   │
            └─────────────┘
```

### Component 1: Live Updates (records.js)

**Purpose:** Push real-time updates when content changes (new comments, votes, etc.)

**Technology:** Socket.io server listening on port 5000

**Data Flow:**
1. Rails app publishes to Redis channel `/records`
2. Channel server receives via Redis subscription
3. Server broadcasts to appropriate Socket.io rooms (`user-{id}`, `group-{id}`)

```javascript
// Key patterns from records.js
await redisSub.subscribe('/records', (json, channel) => {
  let data = JSON.parse(json)
  io.to(data.room).emit('records', data)  // Broadcast to room
})

// User authentication via Redis lookup
let user = await redis.get("/current_users/"+channel_token)
socket.join("user-"+user.id)
user.group_ids.forEach(groupId => { socket.join("group-"+groupId) })
```

**Go Migration Path:**
- gorilla/websocket or nhooyr.io/websocket for WebSocket server
- go-redis for pub/sub subscription
- Room-based broadcasting is straightforward

### Component 2: Matrix Bot Integration (bots.js)

**Purpose:** Send notifications to Matrix chat rooms when events occur in Loomio

**Technology:** matrix-bot-sdk + Redis pub/sub

**Data Flow:**
1. Rails publishes to Redis channel `chatbot/publish` or `chatbot/test`
2. Bot receives message, sends to Matrix room via Matrix API

```javascript
// Key patterns from bots.js
await subscriber.pSubscribe('chatbot/*', (json, channel) => {
  if (channel == 'chatbot/publish') {
    bots[key].resolveRoom(params.config.channel).then((roomId) => {
      bots[key].sendHtmlText(roomId, params.payload.html);
    })
  }
})
```

**Go Migration Path:**
- mautrix/go (Matrix SDK for Go)
- Simple Redis subscription pattern

### Component 3: Collaborative Editing (hocuspocus.mjs)

**Purpose:** Real-time collaborative text editing (Google Docs-style)

**Technology:** Hocuspocus server (Y.js protocol) + SQLite for document storage

**Data Flow:**
1. Browser connects to Hocuspocus with auth token
2. Hocuspocus calls Rails `/api/hocuspocus` to verify permission
3. Y.js CRDT sync handles collaborative editing
4. Documents stored in anonymous SQLite database

```javascript
// Key patterns from hocuspocus.mjs
async onAuthenticate(data) {
  const { token, documentName } = data;
  const response = await fetch(authUrl, {
    method: 'POST',
    body: JSON.stringify({ user_secret: token, document_name: documentName }),
  })
  if (response.status != 200) throw new Error("Not authorized!");
}
```

**Supported Document Types:** comment, discussion, poll, stance, outcome, pollTemplate, discussionTemplate, group, user

**Go Migration Options:**
1. **Keep Node.js:** Hocuspocus is mature, Y.js ecosystem is JavaScript-native
2. **Go alternative:** y-crdt/y-rb has Rust bindings, could wrap for Go
3. **Replace with custom:** Build simpler OT/CRDT system if full Y.js compatibility not needed

### Redis Channels Summary

| Channel | Direction | Purpose |
|---------|-----------|---------|
| `/records` | Rails → Channel Server | Content updates |
| `/system_notice` | Rails → Channel Server | Global announcements |
| `/current_users/{token}` | Rails → Redis (GET) | User session data |
| `chatbot/publish` | Rails → Channel Server | Matrix notifications |
| `chatbot/test` | Rails → Channel Server | Bot testing |

### Frontend Integration

The Vue frontend connects to the channel server:

```javascript
// From vue/package.json dependencies
"socket.io-client": "^4.7.5",      // For records.js connection
"@hocuspocus/provider": "^3.4.0",  // For hocuspocus.mjs connection
"yjs": "^13.6.29",                 // CRDT library
```

### Migration Decision Required

**Option A: Keep Channel Server as Node.js**
- Pros: Minimal risk, Hocuspocus is mature, Y.js ecosystem is JS-native
- Cons: Two languages in production, separate deployment

**Option B: Rewrite in Go**
- Pros: Single language, unified deployment
- Cons: Y.js/CRDT implementation is complex, higher risk

**Option C: Hybrid**
- Rewrite records.js + bots.js in Go (simple, ~100 lines)
- Keep hocuspocus.mjs as Node.js (complex, well-tested)

**Recommendation:** Option C (Hybrid) balances risk and simplification

---

## 6. Authentication & Authorization

### Authentication (Devise)

- Database-backed sessions
- Password with pwned-password checking (production)
- OAuth via `Identity` model (multiple providers)
- SAML SSO support (`ruby-saml`)
- Magic link login (`LoginToken`)

### Authorization (CanCanCan)

Ability definitions likely in `app/models/ability/`. Pattern:

```ruby
user.can?(:update, discussion)
user.can?(:vote, poll)
```

**Go Equivalent:** Casbin or custom RBAC

---

## 7. External Integrations

| Integration | Gem | Purpose | Go Equivalent |
|-------------|-----|---------|---------------|
| AWS S3 | aws-sdk-s3 | File storage | aws-sdk-go-v2 |
| Google Cloud Storage | google-cloud-storage | File storage | cloud.google.com/go/storage |
| Sentry | sentry-ruby | Error tracking | sentry-go |
| OpenAI | ruby-openai | AI features | sashabaranov/go-openai |
| Google Translate | google-cloud-translate | Translation | cloud.google.com/go/translate |
| MaxMind | maxminddb | Geo-IP | oschwald/maxminddb-golang |
| iCalendar | icalendar | Calendar export | arran4/golang-ical |

### Email

- **Outbound:** ActionMailer with Premailer (inline CSS)
- **Inbound:** ActionMailbox (parses incoming emails)
- 7 mailers with templates

---

## 8. Frontend Architecture

### Technology Stack

```json
{
  "vue": "3.5.27",
  "vuetify": "latest",
  "vite": "7.3.1",
  "tiptap": "3.16.0 (25+ extensions)",
  "yjs": "13.6.29",
  "socket.io-client": "4.7.5"
}
```

### Component Organization

```
vue/src/components/
├── auth/           # Login, registration
├── discussion/     # Thread display, comments
├── poll/           # Voting UI, results
├── group/          # Group management
├── lmo_textarea/   # Rich text editor (Tiptap)
├── dashboard/      # User dashboard
└── ... (20 directories, 216 components)
```

### State Management

- LokiJS for client-side database
- Vue composables (not Vuex/Pinia)
- Real-time sync via Yjs

**Migration Impact:** Frontend is API-driven and can be preserved. Critical to maintain API contract compatibility.

---

## 9. Test Coverage

### Test Structure

```
spec/
├── models/         # Model unit tests
├── services/       # Service tests
├── controllers/    # API tests
├── workers/        # Job tests
├── factories.rb    # FactoryBot definitions
└── support/        # Test helpers
```

**Test Count:** 109 spec files

**Testing Stack:**
- RSpec 7.1
- FactoryBot
- WebMock (HTTP stubbing)
- DatabaseCleaner

**Gap Analysis Needed:** Test coverage percentage unknown. Consider running `bundle exec rspec --format documentation` to assess.

---

## 10. Complexity Hotspots

Areas requiring extra attention during migration:

### High Complexity

| Area | Why | Recommendation |
|------|-----|----------------|
| **Event threading** | Nested tree structure, position_key | Study `Event` model deeply |
| **Poll types** | Multiple voting algorithms | Document each type's logic |
| **Real-time collab** | Yjs/Hocuspocus integration | Consider keeping Node.js server |
| **Permissions** | Complex group/discussion/poll permissions | Map all ability rules |

### Medium Complexity

| Area | Why | Recommendation |
|------|-----|----------------|
| **Email parsing** | ActionMailbox complexity | Study ReceivedEmailService |
| **Translations** | 20+ locales, auto-translate | Preserve i18n keys exactly |
| **Search** | pg_search integration | Consider PostgreSQL full-text |
| **Soft deletes** | discarded_at pattern throughout | Implement consistently |

### Lower Complexity

| Area | Why |
|------|-----|
| File attachments | Standard ActiveStorage, well-abstracted |
| Group hierarchy | Simple parent_id relationship |
| Notifications | Event-based, straightforward |

---

## 11. Migration Strategy Implications

Based on this analysis:

### Recommended: Hybrid Approach

1. **Preserve:** Vue frontend (entirely)
2. **Preserve:** Hocuspocus collaborative editing (Node.js) — Y.js ecosystem is JS-native
3. **Rewrite:** Rails backend → Go
4. **Rewrite:** records.js + bots.js → Go (simple, ~100 lines each)
5. **Migrate:** PostgreSQL schema (with transformations)

### Critical Success Factors

1. **API Contract Parity** — The 49 serializers define the contract. Test every endpoint.
2. **PostgreSQL Feature Handling** — JSONB, arrays, citext need Go library support
3. **Event System** — The threading/position logic is complex. Dedicate time.
4. **Permission Mapping** — Extract all CanCanCan rules before rewriting

### Suggested Phase 1 Next Steps

1. [ ] Export all API routes with parameters (`rails routes > routes.txt`)
2. [ ] Document all serializer fields (API contract)
3. [ ] Map all Ability rules (authorization matrix)
4. [ ] Identify all PostgreSQL-specific queries (N+1, custom SQL)
5. [ ] Run test suite and measure coverage

---

## Appendix A: Key Files Reference

### Rails App (loomio/)

| Purpose | Path |
|---------|------|
| Routes | `config/routes.rb` |
| Schema | `db/schema.rb` |
| User model | `app/models/user.rb` |
| Abilities | `app/models/ability/*.rb` |
| API base | `app/controllers/api/v1/` |
| Hocuspocus auth | `app/controllers/api/hocuspocus_controller.rb` |
| Serializers | `app/serializers/` |
| Workers | `app/workers/` |
| Vue entry | `vue/src/main.js` |

### Channel Server (loomio_channel_server/)

| Purpose | Path |
|---------|------|
| Entry point | `index.js` |
| Live updates (Socket.io) | `records.js` |
| Matrix bot | `bots.js` |
| Collaborative editing | `hocuspocus.mjs` |
| Error handling | `bugs.js` |

## Appendix B: Gem → Go Mapping

| Gem | Go Equivalent | Notes |
|-----|---------------|-------|
| devise | JWT + bcrypt | Or Authelia for full-featured |
| cancancan | casbin | Policy-based |
| sidekiq | asynq/river | Redis or PostgreSQL |
| paper_trail | Custom or gorm plugin | Audit logging |
| friendly_id | Custom slugs | Simple implementation |
| pg_search | PostgreSQL FTS | Native support |
| active_model_serializers | Custom or go-json | Manual mapping likely |

## Appendix C: Node.js → Go Mapping (Channel Server)

| Node.js Package | Go Equivalent | Notes |
|-----------------|---------------|-------|
| socket.io | gorilla/websocket | Room broadcasting needs custom impl |
| redis (node-redis) | go-redis/redis | Direct equivalent |
| matrix-bot-sdk | mautrix/go | Matrix protocol SDK |
| @hocuspocus/server | Keep as Node.js | Y.js ecosystem is JS-native |
| @sentry/node | sentry-go | Direct equivalent |

---

*Report generated from Loomio + loomio_channel_server at 2026-01-29 (shallow clones)*
