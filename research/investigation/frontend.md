# Frontend

> Vue SPA, Socket.io, and Hocuspocus integration.

## Technology Stack

| Technology | Purpose | Source |
|------------|---------|--------|
| Vue 3 | SPA framework | `orig/loomio/vue/` |
| Vuetify | UI components | Material Design |
| Vue Router | Client routing | |
| Pinia | State management | |
| TipTap | Rich text editor | Collaborative editing |
| Socket.io-client | Real-time updates | |
| HocuspocusProvider | Yjs sync | |
| y-indexeddb | Offline persistence | |

## Boot Process

**Source:** `orig/loomio/vue/src/shared/boot.js`

1. **Fetch boot data:** `GET /api/v1/boot`
2. **Parse response:**
   ```javascript
   {
     current_user,
     memberships,
     groups,
     notifications,
     channels_uri,      // wss://channels.example.com
     hocuspocus_uri,    // wss://hocuspocus.example.com
     channel_token      // user.secret_token
   }
   ```
3. **Initialize Socket.io** with channel_token
4. **Populate stores** (users, groups, memberships)

## Socket.io Integration

**Source:** `orig/loomio/vue/src/shared/socket.coffee`

### Connection

```javascript
import { io } from 'socket.io-client'

const socket = io(channels_uri, {
  query: { channel_token },
  reconnectionDelay: 1000,
  reconnectionDelayMax: 5000
})
```

### Event Handling

```javascript
// Record updates
socket.on('records', (data) => {
  // data.records contains { discussions: [], comments: [], ... }
  RecordStore.importRecords(data.records)
})

// System notices
socket.on('notice', (data) => {
  Flash.show(data.message)
})
```

### Room Subscription

Automatic via `channel_token` - server joins user to appropriate rooms.

## Hocuspocus Integration

**Source:** `orig/loomio/vue/src/shared/hocuspocus_connector.coffee`

### Provider Setup

```javascript
import { HocuspocusProvider } from '@hocuspocus/provider'
import { IndexeddbPersistence } from 'y-indexeddb'

const provider = new HocuspocusProvider({
  url: hocuspocus_uri,
  name: documentName,  // e.g., 'comment-123-body'
  token: `${user.id},${user.secret_token}`,
  onStatus: (status) => { /* connected/disconnected */ },
  onSynced: () => { /* initial sync complete */ }
})

// Offline persistence
new IndexeddbPersistence(documentName, provider.document)
```

### Document Naming

Pattern: `{record_type}-{record_id}` or `{record_type}-{record_id}-{field}`

Examples:
- `comment-456-body`
- `discussion-789-description`
- `poll-321` (for entire poll)

### TipTap Editor

```javascript
import { Editor } from '@tiptap/core'
import Collaboration from '@tiptap/extension-collaboration'
import CollaborationCursor from '@tiptap/extension-collaboration-cursor'

const editor = new Editor({
  extensions: [
    StarterKit,
    Collaboration.configure({
      document: provider.document
    }),
    CollaborationCursor.configure({
      provider,
      user: { name: currentUser.name, color: randomColor() }
    })
  ]
})
```

## State Management

### RecordStore Pattern

```javascript
// Import records from API/Socket
RecordStore.importRecords({
  discussions: [...],
  comments: [...],
  users: [...]
})

// Access by ID
RecordStore.find('discussion', 123)

// Query
RecordStore.filter('comment', { discussion_id: 123 })
```

### Real-time Updates

1. Socket receives `records` event
2. RecordStore.importRecords() merges new data
3. Vue reactivity updates components

## Offline Support

### Yjs IndexedDB

- Documents persisted locally via `y-indexeddb`
- Reconnection syncs local changes
- CRDT ensures conflict-free merging

### Connection State

```javascript
socket.on('connect', () => {
  store.setOnline(true)
})

socket.on('disconnect', () => {
  store.setOnline(false)
})
```

## Key Components

| Component | Purpose |
|-----------|---------|
| `thread_page` | Discussion view with timeline |
| `poll_common` | Poll display/voting |
| `composer` | New discussion/comment |
| `rich_text_editor` | TipTap wrapper |
| `notifications_dropdown` | Real-time notifications |

## Go Backend Implications

1. **API responses** must match RecordStore expectations
2. **Socket.io protocol** or custom WebSocket
3. **Hocuspocus** can remain Node.js or reimplement
4. **Boot endpoint** structure is critical
5. **Record serialization** format must match

---
