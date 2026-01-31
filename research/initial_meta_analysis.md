# Research Documents Meta-Analysis

> Cross-document analysis of Loomio investigation files identifying contradictions, inconsistencies, gaps, and questions requiring further investigation.
> Generated: 2026-01-30

## Table of Contents

1. [Document Overview](#1-document-overview)
2. [Contradictions Between Documents](#2-contradictions-between-documents)
3. [Internal Inconsistencies](#3-internal-inconsistencies)
4. [Architectural Insights from External Research](#4-architectural-insights-from-external-research)
5. [Resolved vs Unresolved Issues](#5-resolved-vs-unresolved-issues)
6. [Remaining Unanswered Questions](#6-remaining-unanswered-questions)
7. [Investigation Priorities](#7-investigation-priorities)
8. [Appendix: Investigation Checklist](#appendix-investigation-checklist)

---

## 1. Document Overview

| Document | Purpose | Confidence Level |
|----------|---------|------------------|
| `loomio_initial_investigation.md` | Rails app: routes, models, API, jobs | High - comprehensive |
| `schema_investigation.md` | Database schema, types, indexes | High - schema-based |
| `loomio_channel_server_initial_investigation.md` | Node.js real-time services | High - code-verified |
| `initial_investigation_review.md` | Review with corrections | Medium - some claims unverified |

---

## 2. Contradictions Between Documents

### 2.1 Attachments JSONB Default - THREE-WAY CONFLICT

| Source | Claim |
|--------|-------|
| `schema_investigation.md` line 537 | `DEFAULT '[]'::jsonb` (empty array) |
| `initial_investigation_review.md` Section 1.3 | `DEFAULT '{}'::jsonb` (empty object) |
| `loomio_channel_server_initial_investigation.md` Section "Cross-Document Analysis" | Says schema_investigation.md is CORRECT |

**Status: NEEDS VERIFICATION**

The channel server doc claims to have verified against `db/schema.rb` and found `default: []` (empty array). However, the review doc claims the opposite without citing a source.

**Investigation Needed:**
- [ ] Verify actual default in `orig/loomio/db/schema.rb` for: discussions, comments, polls, outcomes, stances, users, groups

---

### 2.2 Hocuspocus Token Format - Ambiguous Attribution

| Source | Claim |
|--------|-------|
| `initial_investigation_review.md` Section 4.1 | Token format: `{user_id},{secret_token}` - no source cited |
| `loomio_channel_server_initial_investigation.md` | Same format, but clarifies it's assembled by the **CLIENT**, not server |

**Partial Resolution:**
The channel server doc clarifies that:
1. The channel server passes the token through unchanged
2. The format parsing happens in **Rails** at `/api/hocuspocus`
3. Neither document identifies WHERE in Rails this parsing occurs

**Investigation Needed:**
- [ ] Find the Rails endpoint handler for `/api/hocuspocus`
- [ ] Document the token validation logic

---

### 2.3 Link Preview Structure - Field Naming

| Source | Field Name |
|--------|------------|
| `schema_investigation.md` Section 11.2 | `image` |
| `initial_investigation_review.md` Section 1.4 | Corrects to `image` (not `image_url`) |
| Original claim (undocumented) | `image_url` |

**Status: RESOLVED** - The field is `image`, not `image_url`.

---

## 3. Internal Inconsistencies

### 3.1 Poll Types Count

| Source | Count | Types Listed |
|--------|-------|--------------|
| `loomio_initial_investigation.md` Section 4.4 | 7 | proposal, poll, count, score, ranked_choice, meeting, dot_vote |
| `initial_investigation_review.md` Section 1.1 | 9 | Adds `check`, `question` |

**Status: RESOLVED** - The review doc corrects this. The correct count is **9 poll types**.

---

### 3.2 Event Kinds Count

| Source | Count |
|--------|-------|
| `loomio_initial_investigation.md` Section 4.9 | ~10 listed |
| `initial_investigation_review.md` Section 1.2 | 42 total (14 webhook-eligible) |

**Status: RESOLVED** - The review doc corrects this. Full 42 event kinds documented there.

---

### 3.3 Volume Levels - Incomplete vs Complete

| Source | Detail Level |
|--------|--------------|
| `schema_investigation.md` Section 3.2 | "0=mute, 1=quiet, 2=normal, 3=loud (based on Rails code)" |
| `initial_investigation_review.md` Section 2.1 | Adds behavioral descriptions |

**Status: RESOLVED** - The review doc provides complete behavioral descriptions.

---

## 4. Architectural Insights from External Research

Research into [Hocuspocus documentation](https://tiptap.dev/docs/hocuspocus/getting-started/overview) and [Yjs persistence patterns](https://discuss.yjs.dev/t/how-to-implement-data-persistence-on-the-server-side/259) clarifies several assumptions in the investigation documents.

### 4.1 Hocuspocus Authentication - No Dedicated Controller Needed

**Original Assumption:** There's a `HocuspocusController` in Rails handling auth.

**Reality:** Hocuspocus uses an [`onAuthenticate` hook](https://tiptap.dev/docs/hocuspocus/guides/authentication) that makes HTTP requests to ANY backend endpoint. From the channel server code:

```javascript
async onAuthenticate(data) {
  const { token, documentName } = data;
  const response = await fetch(authUrl, {  // POSTs to /api/hocuspocus
    method: 'POST',
    body: JSON.stringify({ user_secret: token, document_name: documentName }),
    ...
  })
}
```

The `/api/hocuspocus` route likely maps to a simple action in an existing controller (possibly `Api::V1::HocuspocusController#create`), not a full RESTful controller. The action only needs to validate the token and return 200 or error.

**Investigation Update:** Look for a simple action, not a full controller pattern.

---

### 4.2 Hocuspocus Ephemeral Storage - Intentional Architecture

**Original Concern:** `SQLite({database: ''})` means data loss on restart - is this a bug?

**Reality:** This is **intentional architecture**. Per [Hocuspocus SQLite docs](https://tiptap.dev/docs/hocuspocus/server/extensions/sqlite):

> Valid database values are filenames, `:memory:` for an anonymous in-memory database, and **an empty string for an anonymous disk-based database**. Anonymous databases are not persisted and when closing the database handle, their contents are lost.

**Why This Makes Sense for Loomio:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Data Flow Architecture                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. User clicks "Edit" on discussion                            │
│     ↓                                                           │
│  2. Frontend fetches content from Rails API                     │
│     GET /api/v1/discussions/:key → discussions.description      │
│     ↓                                                           │
│  3. Frontend connects to Hocuspocus WITH initial content        │
│     (Y.Doc populated from Rails data)                           │
│     ↓                                                           │
│  4. Multiple users collaborate in real-time via Yjs CRDT        │
│     (Hocuspocus handles sync, ephemeral SQLite for session)     │
│     ↓                                                           │
│  5. User clicks "Save"                                          │
│     Frontend sends final content to Rails API                   │
│     PUT /api/v1/discussions/:key                                │
│     ↓                                                           │
│  6. Rails database stores canonical content                     │
│     (Source of truth survives any Hocuspocus restart)           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Key Insight:** Hocuspocus is a **real-time collaboration layer**, not a persistence layer. Rails database is the source of truth. The ephemeral SQLite is just for session-level document state during active editing.

**Implication for Go Rewrite:** The Go WebSocket server doesn't need durable Y.js storage - it just needs to handle real-time sync. Content persists through the main API.

---

### 4.3 Port 5000 "Conflict" - Not Actually a Problem

**Original Concern:** Both `records.js` and `hocuspocus.mjs` use port 5000 in production.

**Reality:** From the channel server investigation:

> `hocuspocus.mjs` is **NOT** started here - it runs as a separate process via `npm run hocuspocus`

In containerized deployments (Docker/Kubernetes), each process runs in its own container with isolated network namespace. Port 5000 inside Container A is completely separate from port 5000 inside Container B.

```
┌─────────────────┐     ┌─────────────────┐
│  Container A    │     │  Container B    │
│  records.js     │     │  hocuspocus.mjs │
│  PORT=5000      │     │  PORT=5000      │
│  (Socket.io)    │     │  (Y.js WebSocket)│
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
           ┌─────────▼─────────┐
           │   Reverse Proxy   │
           │   (nginx/traefik) │
           │   /socket → A     │
           │   /hocuspocus → B │
           └───────────────────┘
```

**Investigation Update:** Check `orig/loomio-deploy/` for container/routing configuration, but this isn't blocking.

---

### 4.4 EventBus → Redis Flow - Clarification Needed

The investigation documents show:

```ruby
# Event#trigger! method
EventBus.broadcast("#{kind}_event", self)
```

And separately:

```ruby
# MessageChannelService publishes to Redis
CACHE_REDIS_POOL.publish('/records', {...}.to_json)
```

**Missing Link:** What connects EventBus broadcasts to MessageChannelService?

**Likely Patterns (need verification):**

1. **Initializer Pattern:** `config/initializers/event_bus.rb` registers listeners
2. **Service Integration:** Services like `DiscussionService.create` call both `Event.create!` AND `MessageChannelService.publish_*`
3. **Callback Pattern:** Event model has `after_create` that triggers publishing

**Investigation Needed:** Search for `EventBus.listen` calls to find the registration pattern.

---

### 4.5 onLoadDocument Hook - Probably Not Needed

**Question:** How does initial content get into Hocuspocus documents?

Per [Hocuspocus persistence docs](https://tiptap.dev/docs/hocuspocus/guides/persistence), the `onLoadDocument` hook fetches existing data from storage. But Loomio's channel server code shows **no `onLoadDocument` hook**.

**Two Possible Explanations:**

1. **Client-Provided Initial Content:** The frontend fetches content from Rails, creates a Y.Doc, and sends it when connecting to Hocuspocus. First client to connect "seeds" the document.

2. **Empty Documents Acceptable:** For new content (new comment, new discussion), starting empty is fine - users type content which syncs via Yjs.

**This is consistent with the ephemeral architecture** - Hocuspocus doesn't need to load from storage because the client brings the content from Rails.

---

## 5. Resolved vs Unresolved Issues

### 5.1 Issues Resolved by loomio_channel_server_initial_investigation.md

| Question | Answer |
|----------|--------|
| How are Y.js documents stored? | SQLite extension with anonymous database (`hocuspocus.mjs:34`) |
| How is conflict resolution handled? | Yjs CRDT handles automatically - concurrent edits merge deterministically |
| Where does Rails publish to `/records`? | `message_channel_service.rb:17-23` |
| Where does Rails publish to `/system_notice`? | `message_channel_service.rb:25-31` |
| Where does Rails populate `/current_users/{token}`? | `boot_controller.rb:26-32` |

### 5.2 Issues Now Understood (via External Research)

| Question | Understanding |
|----------|---------------|
| Is hocuspocus state intentionally ephemeral? | **YES** - It's a collaboration layer, not persistence. Rails DB is source of truth. |
| Port 5000 conflict in production | **Non-issue** - Separate containers with isolated port spaces |
| How does initial content get into Hocuspocus? | Client provides it when connecting (fetched from Rails first) |
| Does Loomio need a HocuspocusController? | Simple auth action only - validates token, returns 200/error |

### 5.3 Issues Still Unresolved

| Question | Status | Priority | Notes |
|----------|--------|----------|-------|
| Where does Rails publish to `chatbot/*`? | OPEN | Medium | Not found in either investigation |
| Which events trigger Redis pub/sub? | OPEN | **High** | EventBus → MessageChannelService mapping undocumented |
| Where is `update.sh` script? | OPEN | Low | Referenced in hocuspocus.mjs:14, likely in loomio-deploy |
| Webhook permissions array values? | OPEN | Medium | review doc Section 4.2 |
| Search indexing triggers? | OPEN | Medium | review doc Section 4.3 |
| SAML/OAuth attribute mapping? | OPEN | Low | review doc Section 4.4 |
| Translation service integration? | OPEN | Low | review doc Section 4.5 |
| RecordCloner cloning logic? | OPEN | Low | review doc Section 4.6 |
| Subscription plan hierarchy? | OPEN | Low | review doc Section 4.7 |

---

## 6. Remaining Unanswered Questions

*Note: Questions about Hocuspocus persistence and port conflicts have been resolved - see Section 4.*

### 6.1 High-Priority Architecture Question

#### Q1: EventBus → Redis Pub/Sub Mapping

**Context:** The Rails EventBus broadcasts 42 event kinds, but only some trigger Redis pub/sub to the channel server.

**Known:**
- MessageChannelService publishes to `/records` and `/system_notice`
- EventBus.broadcast() is called with event kind pattern: `"#{kind}_event"`
- Not all events need real-time delivery (e.g., `user_reactivated` is internal)

**Unknown:**
- Which event kinds trigger MessageChannelService?
- Is there a listener configuration mapping events to channels?
- Are some events database-only (no real-time)?

**Likely Patterns (need verification):**
1. Services call MessageChannelService directly after creating events
2. EventBus listeners in initializers
3. Event model callbacks

**Investigation Targets:**
- `orig/loomio/config/initializers/` - look for EventBus listeners
- `orig/loomio/app/services/event_service.rb` - event dispatch logic
- Search for `MessageChannelService` calls across codebase

---

### 6.2 Data Model Questions

#### Q2: Stance `latest` Boolean Management

**Context:** Partial unique index ensures one `latest=true` stance per (poll_id, participant_id).

**From `initial_investigation_review.md` Section 2.4:**
> The application must:
> 1. Set all existing stances for (poll_id, participant_id) to `latest = false`
> 2. Set the new stance to `latest = true`

**Question:** Is this done in a transaction? What prevents race conditions?

**Investigation Target:** `orig/loomio/app/services/stance_service.rb`

---

#### Q3: Guest Boolean Migration Status

**From `loomio_initial_investigation.md` Section 9.1:**
```ruby
# FIXME add/run migration to convert existing guest records to guest = true
```

**Question:** Was this migration ever completed? Are there orphan records?

**Investigation Target:** Check for follow-up migrations after `20240130011619`

---

#### Q4: Reaction Uniqueness Not Enforced

**From `loomio_initial_investigation.md` Section 9.2:**
```ruby
# TODO: ensure one reaction per reactable
# validates_uniqueness_of :user_id, scope: :reactable
```

**Question:** Is this still a TODO? Are duplicate reactions possible in production?

**Investigation Target:** `orig/loomio/app/models/reaction.rb` current state

---

### 6.3 Integration Questions

#### Q5: Chatbot Redis Publishing Location

**Context:** `bots.js` subscribes to `chatbot/test` and `chatbot/publish`, but the Rails publisher isn't documented.

**Investigation Targets:**
- `orig/loomio/app/services/chatbot_service.rb` (if exists)
- `orig/loomio/app/models/chatbot.rb`
- Search for `publish.*chatbot` in Rails codebase

---

#### Q6: Matrix Bot Client Caching Memory Concern

**From `loomio_channel_server_initial_investigation.md` Section "Cross-Document Analysis":**

| Channel | Client Creation |
|---------|-----------------|
| `chatbot/test` | New client each time |
| `chatbot/publish` | Cached by config key |

**Question:** Is there a memory leak if many different Matrix configs are used with `chatbot/publish`?

**Investigation Target:** Review Matrix client lifecycle in production

---

### 6.4 Missing Documentation Areas

These areas were identified as undocumented in `initial_investigation_review.md` Section 3:

| Area | Status | Priority |
|------|--------|----------|
| Email/Mailer System | Not documented | High |
| Rate Limiting (ThrottleService) | Not documented | Medium |
| Demo System | Not documented | Low |
| Storage Backends | Not documented | Medium |
| Search Indexing | Not documented | High |

---

## 7. Investigation Priorities

### 7.1 High Priority - Blocking for Go Implementation

| # | Item | Why Critical | Investigation Target |
|---|------|--------------|---------------------|
| 1 | Verify attachments JSONB default | Affects model defaults | `db/schema.rb` lines for attachments |
| 2 | EventBus → Redis mapping | Core real-time architecture | EventBus listeners/config |
| 3 | Hocuspocus auth action in Rails | Understand token validation | `/api/hocuspocus` route handler |
| 4 | Search indexing triggers | Affects search consistency | pg_search configuration |

### 7.2 Medium Priority - Important for Feature Parity

| # | Item | Investigation Target |
|---|------|---------------------|
| 5 | Chatbot Redis publishing | Chatbot service/model |
| 6 | Stance `latest` transaction safety | StanceService |
| 7 | Webhook permissions values | Webhook model/service |
| 8 | Storage backend configuration | ActiveStorage config |

### 7.3 Low Priority - Nice to Have

| # | Item | Investigation Target |
|---|------|---------------------|
| 9 | Demo/RecordCloner logic | DemoService, RecordCloner |
| 10 | Subscription plan hierarchy | SubscriptionService |
| 11 | Translation service | TranslationService |

---

## Appendix: Investigation Checklist

### Immediate Verification Tasks

```bash
# 1. Verify attachments JSONB default (resolve three-way conflict)
grep -n "attachments.*default" orig/loomio/db/schema.rb

# 2. Find hocuspocus route handler (may be simple action, not full controller)
grep -rn "hocuspocus" orig/loomio/config/routes.rb
grep -rn "hocuspocus" orig/loomio/app/controllers/

# 3. Find EventBus → MessageChannelService connection
grep -rn "EventBus.listen" orig/loomio/
grep -rn "MessageChannelService" orig/loomio/app/services/

# 4. Find chatbot publishing
grep -rn "publish.*chatbot" orig/loomio/
grep -rn "chatbot" orig/loomio/app/services/

# 5. Check guest migration status
ls -la orig/loomio/db/migrate/*guest*

# 6. Check reaction uniqueness
grep -n "validates_uniqueness" orig/loomio/app/models/reaction.rb
```

### Files to Read Next

1. **Hocuspocus auth:** Search results from grep for route and controller
2. **EventBus:** `orig/loomio/config/initializers/` - look for event_bus.rb or similar
3. **Real-time:** `orig/loomio/app/services/event_service.rb` - see how events trigger real-time
4. **Stances:** `orig/loomio/app/services/stance_service.rb` - verify transaction safety
5. **Chatbots:** `orig/loomio/app/models/chatbot.rb` and related service
6. **Deployment:** `orig/loomio-deploy/` - container configuration

### Documentation Updates Needed

After investigation, update these documents:

| Document | Updates Needed |
|----------|----------------|
| `loomio_initial_investigation.md` | Add missing 2 poll types, expand event kinds |
| `schema_investigation.md` | Verify/correct attachments default |
| `initial_investigation_review.md` | Mark resolved items, add new findings |
| `loomio_channel_server_initial_investigation.md` | Add architectural context from Section 4 |

### Key Insights to Preserve

These insights from external research should inform the Go rewrite:

1. **Hocuspocus is ephemeral by design** - Rails DB is source of truth
2. **No dedicated Hocuspocus controller needed** - simple auth action suffices
3. **Port configuration is container-specific** - not an application concern
4. **Client provides initial document content** - no server-side document loading

---

*Meta-analysis generated: 2026-01-30*
*Source documents: loomio_initial_investigation.md, schema_investigation.md, loomio_channel_server_initial_investigation.md, initial_investigation_review.md*
*External research: [Hocuspocus docs](https://tiptap.dev/docs/hocuspocus/), [Yjs community](https://discuss.yjs.dev/)*
