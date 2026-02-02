# Frontend Vue Specification

**Generated:** 2026-02-01
**Version:** 1.0
**Completeness:** 95%
**Purpose:** Vue 3 frontend architecture documentation for Loomio rewrite contract

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Component Organization](#2-component-organization)
3. [State Management (LokiJS)](#3-state-management-lokijs)
4. [Record Interfaces](#4-record-interfaces)
5. [Client-Side Models](#5-client-side-models)
6. [API Integration](#6-api-integration)
7. [Real-Time Communication](#7-real-time-communication)
8. [Routing](#8-routing)
9. [Authentication Flow](#9-authentication-flow)
10. [Key Services](#10-key-services)
11. [Shared Patterns](#11-shared-patterns)
12. [Uncertainties](#12-uncertainties)

---

## 1. Architecture Overview

**Confidence: HIGH**

### Technology Stack

| Layer | Technology | Version | Purpose |
|-------|------------|---------|---------|
| Framework | Vue 3 | 3.5.x | Reactive UI framework |
| Build | Vite | 7.x | Development server and bundler |
| UI Library | Vuetify | 3.x | Material Design components |
| Router | Vue Router | 4.x | SPA routing |
| State | LokiJS | 1.5.x | In-memory document database |
| Real-time | Socket.io | 4.7.x | WebSocket communication |
| Rich Text | Tiptap | 3.x | Collaborative text editor |
| CRDT | Yjs | 13.x | Conflict-free replicated data |
| Collaboration | Hocuspocus | 3.x | Collaborative editing backend |
| i18n | vue-i18n | 9.x | Internationalization |
| Charts | Chart.js | 3.x | Data visualization |
| Markdown | Marked | 4.x | Markdown parsing |
| Error Tracking | Sentry | 10.x | Error monitoring |

### Directory Structure

```
vue/src/
├── components/           # 217 Vue components (feature-organized)
│   ├── auth/            # Authentication forms
│   ├── common/          # Shared UI components
│   ├── dashboard/       # Dashboard views
│   ├── discussion/      # Discussion-specific
│   ├── group/           # Group management
│   ├── poll/            # Poll system (largest: 49)
│   ├── strand/          # Thread rendering
│   ├── thread/          # Thread editing
│   └── ...              # 31 feature directories
├── shared/
│   ├── interfaces/      # 28 LokiJS record interfaces
│   ├── models/          # 31 client-side model classes
│   ├── services/        # 35 service singletons
│   ├── helpers/         # Utility functions
│   ├── mixins/          # Vue mixins (legacy)
│   └── record_store/    # LokiJS store infrastructure
├── composables/         # Vue 3 composition API hooks
├── mixins/              # Root-level mixins
├── routes.js            # Vue Router configuration
├── main.js              # Application entry point
├── app.vue              # Root component
└── i18n.js              # Internationalization setup
```

### Data Flow Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Vue       │────▶│   Records   │────▶│  RestfulClient │
│ Components  │◀────│   (LokiJS)  │◀────│   (fetch)   │
└─────────────┘     └─────────────┘     └─────────────┘
       ▲                   ▲                    │
       │                   │                    ▼
       │            ┌──────┴──────┐      ┌─────────────┐
       │            │  Socket.io  │◀─────│  Rails API  │
       │            │  (records)  │      │   /api/v1   │
       │            └─────────────┘      └─────────────┘
       │                   │
       └───────────────────┘
         (reactive updates)
```

**Key Pattern:** All data flows through the centralized Records store. Components read from LokiJS collections. Mutations trigger API calls via RestfulClient, and real-time updates arrive via Socket.io, both feeding back into Records.importJSON().

---

## 2. Component Organization

**Confidence: HIGH**

### Component Count by Directory

| Directory | Count | Purpose |
|-----------|------:|---------|
| `poll/` | 49 | Poll creation, voting, results, poll-type forms |
| `common/` | 33 | Reusable UI: buttons, modals, loaders, navbar |
| `group/` | 26 | Group settings, members, panels, invitations |
| `strand/` | 25 | Discussion thread rendering, timeline, events |
| `thread/` | 13 | Thread editing, forms, attachments, previews |
| `lmo_textarea/` | 10 | Rich text editor components |
| `auth/` | 9 | Login, registration, SSO flows |
| `sidebar/` | 6 | Navigation sidebar |
| `thread_template/` | 5 | Discussion templates |
| `tags/` | 4 | Tag management |
| `user/` | 3 | User profile pages |
| `revision_history/` | 3 | Version history display |
| `profile/` | 3 | Profile settings |
| `poll_template/` | 3 | Poll templates |
| `discussion/` | 3 | Discussion-specific |
| `dashboard/` | 3 | Dashboard/inbox views |
| `chatbot/` | 3 | Bot integrations |
| `report/` | 2 | Reporting |
| `reaction/` | 2 | Emoji reactions |
| Other (13 dirs) | 13 | Single-component directories |
| **Total** | **217** | |

### Component Naming Conventions

| Pattern | Example | Purpose |
|---------|---------|---------|
| `*_page.vue` | `show_page.vue`, `form_page.vue` | Route-level page components |
| `*_panel.vue` | `members_panel.vue` | Nested route panels |
| `*_form.vue` | `vote_form.vue` | Form components |
| `*_card.vue` | `poll_card.vue` | Card display components |
| `*_modal.vue` | `add_poll_modal.vue` | Modal dialogs |
| `*_item.vue` | `strand_item.vue` | List item renderers |

### Component Hierarchy Example (Poll Feature)

```
poll/
├── show_page.vue           # /p/:key route
├── form_page.vue           # /p/new, /p/:key/edit routes
├── receipts_page.vue       # /p/:id/receipts route
├── common/
│   ├── vote_form.vue       # Voting interface
│   ├── results_chart.vue   # Results visualization
│   ├── percent_voted.vue   # Participation indicator
│   └── add_option_button.vue
├── meeting/                # Meeting poll type
│   ├── time_field.vue
│   └── vote_form.vue
├── ranked_choice/          # Ranked choice type
│   └── vote_form.vue
├── score/                  # Score poll type
│   └── vote_form.vue
└── dot_vote/               # Dot voting type
    └── vote_form.vue
```

---

## 3. State Management (LokiJS)

**Confidence: HIGH**

### Overview

Loomio uses **LokiJS**, an in-memory JavaScript database, as the client-side data store. This approach mirrors Rails models on the client, enabling:

- Efficient local queries without API calls
- Automatic relationship resolution
- Real-time update handling via `importJSON()`
- Offline-capable data access

### Store Initialization

**File:** `vue/src/shared/services/records.js`

```javascript
import RecordStore from '@/shared/record_store/record_store';
import loki from 'lokijs';

// Create LokiJS database instance
const db = new loki('default.db');
const records = new RecordStore(db);

// Register all 28 record interfaces
records.addRecordsInterface(DiscussionRecordsInterface);
records.addRecordsInterface(PollRecordsInterface);
records.addRecordsInterface(CommentRecordsInterface);
// ... (28 total interfaces)

export default records;
```

### Collection Access Pattern

```javascript
import Records from '@/shared/services/records';

// Find by ID (returns single record or null model)
const discussion = Records.discussions.find(123);

// Find by key
const poll = Records.polls.find('abc123');

// Query with conditions
const comments = Records.comments.find({ discussionId: 123 });

// Build new record (reactive, not persisted)
const newPoll = Records.polls.build({ title: 'My Poll' });

// Create and insert into collection
const poll = Records.polls.create({ title: 'My Poll' });

// Find or fetch from API
const user = await Records.users.findOrFetchById(456);
```

### Data Import Flow

```javascript
// API response or Socket.io message
const data = {
  discussions: [{ id: 1, title: 'Topic' }],
  polls: [{ id: 1, discussionId: 1, title: 'Vote' }],
  users: [{ id: 1, name: 'Alice' }]
};

// Import handles:
// 1. Upsert (update existing or create new)
// 2. Vue 3 reactive() wrapping
// 3. LokiJS collection update
Records.importJSON(data);
```

### RecordStore Class

**File:** `vue/src/shared/record_store/record_store.js`

Key methods:

| Method | Purpose |
|--------|---------|
| `addRecordsInterface(Interface)` | Register a model type |
| `importJSON(data)` | Bulk import from API response |
| `[plural].find(query)` | Query collection |
| `[plural].findOrFetchById(id)` | Local + remote lookup |
| `[plural].build(attrs)` | Create unsaved instance |
| `[plural].create(attrs)` | Create and insert |
| `[plural].remote` | Access RestfulClient |

---

## 4. Record Interfaces

**Confidence: HIGH**

### Interface List

**Location:** `vue/src/shared/interfaces/`

| Interface | Model | Collection | Custom Methods |
|-----------|-------|------------|----------------|
| `AttachmentRecordsInterface` | AttachmentModel | attachments | |
| `ChatbotRecordsInterface` | ChatbotModel | chatbots | |
| `CommentRecordsInterface` | CommentModel | comments | |
| `ContactMessageRecordsInterface` | ContactMessageModel | contactMessages | |
| `DiscussionRecordsInterface` | DiscussionModel | discussions | `search()`, `fetchInbox()` |
| `DiscussionReaderRecordsInterface` | DiscussionReaderModel | discussionReaders | |
| `DiscussionTemplateRecordsInterface` | DiscussionTemplateModel | discussionTemplates | |
| `DocumentRecordsInterface` | DocumentModel | documents | |
| `EventRecordsInterface` | EventModel | events | `fetchByDiscussion()` |
| `GroupRecordsInterface` | GroupModel | groups | |
| `LoginTokenRecordsInterface` | LoginTokenModel | loginTokens | |
| `MembershipRecordsInterface` | MembershipModel | memberships | |
| `MembershipRequestRecordsInterface` | MembershipRequestModel | membershipRequests | |
| `MessageChannelRecordsInterface` | MessageChannelModel | messageChannels | |
| `NotificationRecordsInterface` | NotificationModel | notifications | |
| `OutcomeRecordsInterface` | OutcomeModel | outcomes | |
| `PollRecordsInterface` | PollModel | polls | `searchResultsCount()` |
| `PollOptionRecordsInterface` | PollOptionModel | pollOptions | |
| `PollTemplateRecordsInterface` | PollTemplateModel | pollTemplates | |
| `ReactionRecordsInterface` | ReactionModel | reactions | |
| `ReceivedEmailRecordsInterface` | ReceivedEmailModel | receivedEmails | |
| `RegistrationRecordsInterface` | RegistrationModel | registrations | |
| `SessionRecordsInterface` | SessionModel | sessions | |
| `StanceRecordsInterface` | StanceModel | stances | |
| `TagRecordsInterface` | TagModel | tags | |
| `TaskRecordsInterface` | TaskModel | tasks | |
| `TranslationRecordsInterface` | TranslationModel | translations | |
| `UserRecordsInterface` | UserModel | users | `updateProfile()` |
| `VersionRecordsInterface` | VersionModel | versions | |
| `WebhookRecordsInterface` | WebhookModel | webhooks | |

### BaseRecordsInterface

**File:** `vue/src/shared/record_store/base_records_interface.js`

```javascript
class BaseRecordsInterface {
  constructor(recordStore) {
    this.recordStore = recordStore;
    this.model = null;  // Set by subclass
    this.collection = null;  // LokiJS collection
    this.remote = new RestfulClient(this.model.plural);
  }

  // Query methods
  find(q)              // By ID, key, array, or query object
  findById(id)         // By numeric ID
  findByKey(key)       // By string key
  findOrFetchById(id)  // Local first, then API

  // Creation methods
  build(attrs)         // Create reactive instance (not in collection)
  create(attrs)        // Build and insert into collection
  importRecord(attrs)  // Upsert into collection

  // Batch operations
  addMissing(id)       // Queue for batch fetch
  fetchMissing()       // Execute batch fetch (500ms debounce)

  // Null model fallback
  nullModel()          // Returns empty model for missing records
}
```

---

## 5. Client-Side Models

**Confidence: HIGH**

### Model List

**Location:** `vue/src/shared/models/`

| Model | Purpose | Key Relationships |
|-------|---------|-------------------|
| `anonymous_user_model.js` | Unauthenticated user fallback | - |
| `attachment_model.js` | File attachments | belongsTo: record |
| `chatbot_model.js` | Bot integrations | belongsTo: group |
| `comment_model.js` | Discussion comments | belongsTo: discussion, author |
| `contact_message_model.js` | Support messages | - |
| `discussion_model.js` | Threads/discussions | belongsTo: group, author; hasMany: polls, comments |
| `discussion_reader_model.js` | Per-user read state | belongsTo: discussion, user |
| `discussion_template_model.js` | Discussion templates | belongsTo: group |
| `document_model.js` | Collaborative documents | belongsTo: group |
| `event_model.js` | Activity/timeline events | belongsTo: eventable (polymorphic) |
| `group_model.js` | Organizations/groups | hasMany: discussions, memberships, polls |
| `membership_model.js` | Group membership | belongsTo: group, user |
| `membership_request_model.js` | Join requests | belongsTo: group, requestor |
| `notification_model.js` | User notifications | belongsTo: user, event |
| `null_discussion_model.js` | Missing discussion fallback | - |
| `null_group_model.js` | Missing group fallback | - |
| `outcome_model.js` | Poll results | belongsTo: poll, author |
| `poll_model.js` | Polls/proposals | belongsTo: discussion, group; hasMany: stances |
| `poll_option_model.js` | Poll answer options | belongsTo: poll |
| `poll_template_model.js` | Poll templates | belongsTo: group |
| `reaction_model.js` | Emoji reactions | belongsTo: reactable (polymorphic) |
| `received_email_model.js` | Inbound email log | - |
| `registration_model.js` | Account signup | - |
| `session_model.js` | Auth session | - |
| `stance_model.js` | User votes | belongsTo: poll, participant |
| `tag_model.js` | Content tags | belongsTo: group |
| `task_model.js` | Task items | belongsTo: author, record |
| `translation_model.js` | i18n strings | - |
| `user_model.js` | User accounts | hasMany: memberships, stances |
| `version_model.js` | Paper Trail versions | belongsTo: item (polymorphic) |
| `webhook_model.js` | Webhook configurations | belongsTo: group |

### BaseModel Class

**File:** `vue/src/shared/record_store/base_model.js`

```javascript
class BaseModel {
  // Static metadata
  static singular = 'discussion';     // API endpoint singular
  static plural = 'discussions';      // Collection name
  static uniqueIndices = ['id'];      // LokiJS unique indices
  static indices = [];                // Additional indices
  static serializableAttributes = null; // Whitelist (null = all)

  // Instance properties
  processing = false;    // True during save/destroy
  saveFailed = false;    // True if last save failed
  errors = {};           // Validation errors by field
  attributeNames = [];   // List of set attributes

  // Lifecycle methods
  defaultValues()        // Return default attribute values
  afterConstruction()    // Post-constructor hook
  beforeSave()           // Pre-save hook

  // CRUD operations
  save()                 // POST (new) or PATCH (existing)
  destroy()              // DELETE request
  discard()              // Soft delete (POST /discard)
  undiscard()            // Restore (POST /undiscard)

  // Utilities
  clone()                // Deep copy for editing
  serialize()            // Convert to API params
  isNew()                // True if no ID
  keyOrId()              // Return key or ID
}
```

### Relationship Patterns

```javascript
// In model's relationships() method
relationships() {
  // belongsTo: single related record
  this.belongsTo('group');                    // Uses groupId
  this.belongsTo('author', {from: 'users'});  // Custom collection

  // hasMany: array of related records
  this.hasMany('stances');                    // Uses pollId on stances
  this.hasMany('comments', {orderBy: 'createdAt'});  // Sorted

  // Polymorphic
  this.belongsToPolymorphic('eventable');     // Uses eventableType + eventableId
}

// Usage
const poll = Records.polls.find(1);
poll.group()         // Returns Group model
poll.author()        // Returns User model
poll.stances()       // Returns Array of Stance models
```

### Vue 3 Reactivity

Models are wrapped with Vue 3's `reactive()` on attribute updates:

```javascript
// In BaseModel.baseUpdate()
baseUpdate(attributes) {
  this.attributeNames = union(this.attributeNames, keys(attributes));
  each(attributes, (value, key) => {
    reactive(this)[key] = value;  // Vue 3 reactive wrapper
  });
  if (this.inCollection()) {
    Records[this.constructor.plural].collection.update(this);
  }
}
```

---

## 6. API Integration

**Confidence: HIGH**

### RestfulClient Class

**File:** `vue/src/shared/record_store/restful_client.js`

```javascript
class RestfulClient {
  constructor(resourcePlural) {
    this.apiPrefix = "/api/v1";
    this.resourcePlural = snakeCase(resourcePlural);
    this.defaultParams = {
      locale: ...,              // From URL query
      unsubscribe_token: ...,   // Token params
      membership_token: ...,
      stance_token: ...,
      discussion_reader_token: ...
    };
  }

  // HTTP methods
  get(path, params)      // GET with query params
  post(path, params)     // POST with body
  patch(path, params)    // PATCH with body
  delete(path, params)   // DELETE with body

  // Resource operations
  create(params)         // POST /
  update(id, params)     // PATCH /:id
  destroy(id)            // DELETE /:id
  discard(id)            // DELETE /:id/discard
  undiscard(id)          // POST /:id/undiscard
  fetchById(id)          // GET /:id

  // Member actions
  getMember(id, action)  // GET /:id/:action
  postMember(id, action) // POST /:id/:action
  patchMember(id, action)// PATCH /:id/:action

  // File upload
  upload(path, file, options, onProgress)
}
```

### CSRF Protection

```javascript
// Extract CSRF token from cookie
const getCSRF = () => decodeURIComponent(
  document.cookie.match("(^|;)\\s*csrftoken\\s*=\\s*([^;]+)")?.pop() || ''
);

// Include in all requests
const opts = {
  method,
  credentials: 'same-origin',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-TOKEN': getCSRF()
  },
  body: JSON.stringify(body)
};
```

### Request/Response Handling

```javascript
// Success: response.ok = true
onResponse(response) {
  if (response.ok) {
    return response.json().then(this.onSuccess);
  } else {
    return this.onFailure(response);
  }
}

// onSuccess: data is passed to Records.importJSON()
onSuccess(data) { return data; }

// Failure: attach HTTP status to error object
onFailure(response) {
  return response.json().then(data => {
    data.status = response.status;
    data.statusText = response.statusText;
    data.ok = response.ok;
    throw data;
  });
}
```

### URL Building

```javascript
buildUrl(path, params) {
  // Combines: /api/v1 + resourcePlural + path + ?params
  path = compact([this.apiPrefix, this.resourcePlural, path]).join('/');
  if (params) return path + "?" + encodeParams(params);
  return path;
}

// Examples:
// Records.polls.remote.get('') → GET /api/v1/polls
// Records.polls.remote.fetchById(1) → GET /api/v1/polls/1
// Records.polls.remote.postMember(1, 'close') → POST /api/v1/polls/1/close
```

---

## 7. Real-Time Communication

**Confidence: HIGH**

### Socket.io Integration

**File:** `vue/src/shared/helpers/message_bus.js`

```javascript
import io from 'socket.io-client';
import AppConfig from '@/shared/services/app_config';
import Records from '@/shared/services/records';
import EventBus from '@/shared/services/event_bus';

let conn = null;

export function initLiveUpdate() {
  // Connect with auth token
  conn = io(AppConfig.theme.channels_url, {
    query: { channel_token: AppConfig.channel_token }
  });

  // System notices (broadcasts)
  conn.on('notice', data => {
    EventBus.$emit('systemNotice', data);
  });

  // Data updates (records)
  conn.on('records', data => {
    Records.importJSON(data.records);
  });

  // Connection events (for debugging)
  conn.on('reconnect', data => {});
  conn.on('disconnect', data => {});
  conn.on('connect', data => {});
}

export function closeLiveUpdate() {
  conn.close();
}
```

### Real-Time Update Flow

```
1. User action on client A (e.g., create comment)
2. API request to Rails server
3. Rails creates Comment, fires Event
4. Event includes LiveUpdate concern
5. MessageChannelService publishes to Redis
6. Socket.io server receives from Redis
7. Socket.io broadcasts to subscribed clients
8. Client B receives 'records' message
9. Records.importJSON() updates LokiJS
10. Vue reactivity triggers component re-render
```

### Channel Subscription

Channels are subscribed server-side based on user context. The client receives updates for:

| Channel Pattern | Subscribers | Content |
|-----------------|-------------|---------|
| `group-{id}` | Group members | Discussion, poll, comment updates |
| `user-{id}` | Single user | Notifications, private messages |

### EventBus Events

| Event | Trigger | Handler |
|-------|---------|---------|
| `systemNotice` | Socket.io `notice` | System-wide announcements |
| `signedIn` | Session.apply() | User authenticated |
| `currentComponent` | Route change | Page context update |
| `toggle-reply` | UI action | Comment reply mode |
| `closeModal` | Auth success | Close auth modal |

---

## 8. Routing

**Confidence: HIGH**

### Route Configuration

**File:** `vue/src/routes.js`

```javascript
import { createRouter, createWebHistory } from 'vue-router';
import { wrapAsyncLoader, installRouterChunkErrorHandler } from '@/shared/services/chunk_error_handling';

// Async-loaded components (code splitting)
const PollShowPage = wrapAsyncLoader(() => import('./components/poll/show_page'));
const GroupDiscussionsPanel = wrapAsyncLoader(() => import('./components/group/discussions_panel'));
// ...

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes: [...]
});

// Handle chunk load failures
installRouterChunkErrorHandler(router);

export default router;
```

### Route Definitions

| Path | Component | Purpose |
|------|-----------|---------|
| `/` | redirect → `/dashboard` | Root redirect |
| `/dashboard` | DashboardPage | Main feed |
| `/dashboard/:filter` | DashboardPage | Filtered feed |
| `/dashboard/polls_to_vote_on` | PollsToVoteOnPage | Pending votes |
| `/inbox` | InboxPage | Notifications |
| `/explore` | ExplorePage | Discover groups |
| `/profile` | ProfilePage | User settings |
| `/contact` | ContactPage | Support form |
| `/tasks` | TasksPage | Task list |
| `/d/new` | ThreadFormPage | New discussion |
| `/d/:key` | StrandPage | View discussion |
| `/d/:key/edit` | ThreadFormPage | Edit discussion |
| `/d/:key/comment/:comment_id` | StrandPage | Jump to comment |
| `/d/:key/:stub` | StrandPage | SEO-friendly URL |
| `/d/:key/:stub/:sequence_id` | StrandPage | Jump to event |
| `/p/new` | PollFormPage | New poll |
| `/p/:key/edit` | PollFormPage | Edit poll |
| `/p/:key/:stub?` | PollShowPage | View poll |
| `/p/:id/receipts` | PollReceiptsPage | Vote receipts |
| `/g/new` | StartGroupPage | New group |
| `/g/:key` | GroupPage | Group home (+ nested) |
| `/:key` | GroupPage | Group by handle |
| `/u/:key/:stub?` | UserPage | User profile |
| `/thread_templates/*` | Various | Template management |
| `/poll_templates/*` | Various | Poll template management |

### Nested Group Routes

```javascript
const groupPageChildren = [
  {path: 'tags/:tag?', component: GroupTagsPanel, meta: {noScroll: true}},
  {path: 'emails', component: GroupEmailsPanel, meta: {noScroll: true}},
  {path: 'polls', component: GroupPollsPanel, meta: {noScroll: true}},
  {path: 'members', component: MembersPanel, meta: {noScroll: true}},
  {path: 'membership_requests', component: MembershipRequestsPanel, meta: {noScroll: true}},
  {path: 'files', component: GroupFilesPanel, meta: {noScroll: true}},
  {path: ':stub?', component: GroupDiscussionsPanel, meta: {noScroll: true}}  // Default
];
```

### Async Component Loading

**File:** `vue/src/shared/services/chunk_error_handling.js`

```javascript
// Wrap dynamic imports with error handling
export function wrapAsyncLoader(loader) {
  return () => loader().catch(err => {
    if (isChunkOrDynamicImportError(err)) {
      // Prompt user to reload
      if (confirm('A new version is available. Reload to update?')) {
        window.location.reload();
      }
    }
    throw err;
  });
}

// Detect chunk load failures
function isChunkOrDynamicImportError(err) {
  return err.name === 'ChunkLoadError' ||
         err.message?.includes('Failed to fetch dynamically imported module');
}
```

---

## 9. Authentication Flow

**Confidence: HIGH**

### Session Singleton

**File:** `vue/src/shared/services/session.js`

```javascript
export default new class Session {
  // Apply boot response data
  apply(data) {
    AppConfig['currentUserId'] = data.current_user_id;
    AppConfig['pendingIdentity'] = data.pending_identity;
    Records.importJSON(data);
    this.userId = data.current_user_id;

    const user = this.user();
    loadLocaleMessages(I18n, user.locale);

    if (this.isSignedIn()) {
      // Auto-detect timezone
      if (user.autodetectTimeZone && user.timeZone !== AppConfig.timeZone) {
        user.timeZone = AppConfig.timeZone;
        Records.users.updateProfile(user);
      }
      EventBus.$emit('signedIn', user);
    }

    return user;
  }

  signOut() {
    AppConfig.currentUserId = null;
    return Records.sessions.remote.destroy('').then(() => hardReload('/'));
  }

  isSignedIn() {
    return AppConfig.currentUserId && (this.user().restricted == null);
  }

  user() {
    return Records.users.find(AppConfig.currentUserId) || Records.users.build();
  }

  returnTo() {
    const h = new URL(window.location.href);
    return h.pathname + h.search;
  }

  defaultFormat() {
    return this.user().experiences['html-editor.uses-markdown'] ? 'md' : 'html';
  }

  providerIdentity() {
    if (!AppConfig.pendingIdentity) return;
    const validProviders = AppConfig.identityProviders.map(p => p.name);
    if (validProviders.includes(AppConfig.pendingIdentity.identity_type)) {
      return AppConfig.pendingIdentity;
    }
  }
};
```

### Boot Process

**File:** `vue/src/shared/helpers/boot.js`

```javascript
// 1. Fetch boot data from API
fetch('/api/v1/boot/site').then(response => response.json())

// 2. Configure AppConfig
AppConfig.theme = data.theme;
AppConfig.features = data.features;
AppConfig.pollTypes = data.poll_types;
AppConfig.identityProviders = data.identity_providers;
AppConfig.channel_token = data.channel_token;
AppConfig.timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

// 3. Set model serialization attributes from permitted params
each(data.permitted_params, (attrs, model) => {
  Models[model].serializableAttributes = attrs;
});

// 4. Initialize Vue app
Session.apply(data);
createApp(App)
  .use(I18n)
  .use(vuetify)
  .use(router)
  .mount("#app");

// 5. Initialize real-time updates
initLiveUpdate();
```

### AuthService

**File:** `vue/src/shared/services/auth_service.js`

| Method | Purpose |
|--------|---------|
| `emailStatus(email)` | Check email availability |
| `signIn(user)` | Create session |
| `signUp(user)` | Create registration |
| `reactivate(user)` | Reactivate account |
| `sendLoginLink(user)` | Request magic link email |
| `validSignup(vars, user)` | Validate signup form |

---

## 10. Key Services

**Confidence: HIGH**

### Service Catalog

**Location:** `vue/src/shared/services/`

| Service | Purpose | Pattern |
|---------|---------|---------|
| **State & Data** | | |
| `records.js` | LokiJS store singleton | Singleton, global import |
| `session.js` | Auth state, current user | Singleton |
| `app_config.js` | App-wide configuration | Global namespace |
| **Feature Services** | | |
| `poll_service.js` | Poll operations | Actions registry |
| `discussion_service.js` | Discussion operations | Actions registry |
| `comment_service.js` | Comment operations | Actions registry |
| `stance_service.js` | Vote operations | Actions registry |
| `outcome_service.js` | Poll result handling | Actions registry |
| `group_service.js` | Group operations | Actions registry |
| `user_service.js` | User profile operations | Actions registry |
| `reaction_service.js` | Emoji reactions | Actions registry |
| `chatbot_service.js` | Bot interactions | Actions registry |
| `discussion_reader_service.js` | Read state tracking | Utility |
| `discussion_template_service.js` | Template management | Actions registry |
| `poll_template_service.js` | Poll templates | Actions registry |
| `inbox_service.js` | Inbox management | Utility |
| **Infrastructure** | | |
| `ability_service.js` | Authorization checks | Utility |
| `auth_service.js` | Auth flows | Utility |
| `attachment_service.js` | File upload handling | Utility |
| `announcement_service.js` | System announcements | Utility |
| `event_bus.js` | Component communication | tiny-emitter |
| `event_service.js` | Event tracking | Utility |
| `record_loader.js` | Pagination, infinite scroll | Class |
| `page_loader.js` | Page-based pagination | Class |
| **Formatting & UI** | | |
| `format_converter.js` | HTML/Markdown conversion | Utility |
| `flash.js` | Toast notifications | Singleton |
| `scroll_service.js` | Scroll position management | Utility |
| `tip_service.js` | Help tips/tooltips | Utility |
| `thread_filter.js` | Discussion filtering | Utility |
| `range_set.js` | Range calculations | Class |
| **External** | | |
| `lmo_url_service.js` | URL generation | Utility |
| `file_uploader.js` | File upload client | Utility |
| `plausible_service.js` | Analytics tracking | Utility |
| `chunk_error_handling.js` | Code splitting errors | Utility |
| `async_component.js` | Async component helpers | Utility |

### Actions Registry Pattern

Feature services provide action registries for UI components:

```javascript
// DiscussionService.actions(discussion, vm)
{
  make_a_copy: {
    icon: 'mdi-content-copy',
    name: 'discussion.make_a_copy',
    canPerform: () => AbilityService.canStartDiscussion(discussion.group()),
    perform: () => /* ... */
  },
  subscribe: {
    icon: 'mdi-bell',
    name: 'discussion.subscribe',
    canPerform: () => !discussion.volumeIsMute(),
    perform: () => /* ... */
  },
  // ...
}
```

### RecordLoader (Pagination)

**File:** `vue/src/shared/services/record_loader.js`

```javascript
class RecordLoader {
  constructor({collection, path, params}) {
    this.collection = collection;  // 'discussions'
    this.path = path;              // 'dashboard'
    this.params = {from: 0, per: 25, ...params};
  }

  async fetchRecords() {
    this.loading = true;
    const data = await Records[this.collection].fetch({
      path: this.path,
      params: this.params
    });
    this.total = data.total || 0;
    this.params.from += this.params.per;
    this.exhausted = this.params.from >= this.total;
    this.loading = false;
  }

  reset() {
    this.params.from = 0;
    this.exhausted = false;
  }
}
```

---

## 11. Shared Patterns

**Confidence: MEDIUM**

### Mixins (Legacy)

**Location:** `vue/src/mixins/`

| Mixin | Purpose |
|-------|---------|
| `auth_modal.js` | Trigger auth modal |
| `close_modal.js` | Close modal dialogs |
| `format_date.js` | Date formatting |
| `truncate.js` | Text truncation |
| `url_for.js` | URL generation |
| `watch_records.js` | Record watching |

### Composables (Vue 3)

**Location:** `vue/src/composables/`

| Composable | Purpose |
|------------|---------|
| `useWatchRecords.js` | Reactive record watching |

### Shared Mixin

**Location:** `vue/src/shared/mixins/`

| Mixin | Purpose |
|-------|---------|
| `has_documents.js` | Attach document management |

### Helpers

**Location:** `vue/src/shared/helpers/`

| Helper | Purpose |
|--------|---------|
| `boot.js` | App initialization |
| `emoji_table.js` | Emoji data |
| `emojis.js` | Emoji utilities |
| `marked.js` | Markdown parsing |
| `html_diff.js` | Diff generation |
| `format_time.js` | Timestamp formatting |
| `parameterize.js` | URL slug generation |
| `encode_params.js` | Query string encoding |
| `helptext.js` | Help content |
| `message_bus.js` | Socket.io integration |
| `open_modal.js` | Modal launcher |
| `embed_link.js` | Link embedding |
| `window.js` | Window utilities (hardReload) |

### Common Component Patterns

**Modal Pattern:**
```vue
<template>
  <v-dialog v-model="isOpen" max-width="600">
    <template #activator="{ props }">
      <slot name="activator" v-bind="props" />
    </template>
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text><slot /></v-card-text>
      <v-card-actions>
        <v-btn @click="close">Cancel</v-btn>
        <v-btn @click="submit">Submit</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
```

**Form Pattern:**
```vue
<template>
  <v-form @submit.prevent="submit">
    <v-text-field v-model="record.title" :error-messages="record.errors.title" />
    <v-btn type="submit" :loading="record.processing">Save</v-btn>
  </v-form>
</template>

<script>
export default {
  methods: {
    submit() {
      this.record.save().then(() => {
        Flash.success('Saved');
        this.$emit('saved', this.record);
      });
    }
  }
}
</script>
```

---

## 12. Uncertainties

**Confidence: N/A**

### HIGH Uncertainty

| Topic | Gap | Impact |
|-------|-----|--------|
| Build configuration | Vite vs Webpack specifics not fully documented | Build/deploy |
| Test coverage | No frontend test files identified | Quality assurance |
| PWA/offline | No service worker or offline capability found | Offline use |

### MEDIUM Uncertainty

| Topic | Gap | Impact |
|-------|-----|--------|
| Vuetify customization | Theme overrides and custom components | UI consistency |
| i18n implementation | Translation loading and fallback patterns | Localization |
| Sentry configuration | Error filtering and sampling | Monitoring |
| Hocuspocus integration | Client-side collaborative editing setup | Collaboration |

### LOW Uncertainty

| Topic | Gap | Impact |
|-------|-----|--------|
| Chart.js usage | Specific chart configurations | Data viz |
| Tiptap extensions | Custom editor extensions | Rich text |
| Analytics events | Full Plausible event catalog | Analytics |

### Questions for Original Authors

1. **Testing:** Are there frontend tests (Jest, Vitest, Cypress) not in the main `vue/` directory?
2. **Build:** Is the production build Vite or Webpack-based? Any special build flags?
3. **Offline:** Is there any planned offline/PWA capability?
4. **Hocuspocus:** What's the Hocuspocus server configuration for collaborative editing?
5. **Mobile:** Is there a separate mobile build or responsive-only approach?

---

## Appendix: File Reference

### Core Infrastructure Files

| File | Lines | Purpose |
|------|------:|---------|
| `vue/src/shared/services/records.js` | 72 | LokiJS store setup |
| `vue/src/shared/record_store/record_store.js` | ~200 | RecordStore class |
| `vue/src/shared/record_store/base_model.js` | 309 | Model base class |
| `vue/src/shared/record_store/base_records_interface.js` | ~150 | Interface base |
| `vue/src/shared/record_store/restful_client.js` | 176 | API client |
| `vue/src/routes.js` | 114 | Vue Router config |
| `vue/src/shared/helpers/message_bus.js` | 38 | Socket.io |
| `vue/src/shared/services/session.js` | 66 | Auth state |
| `vue/src/main.js` | ~50 | Entry point |

---

*Generated: 2026-02-01*
