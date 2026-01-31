# loomio_channel_server Investigation

This document provides a comprehensive analysis of the `loomio_channel_server` Node.js application, which serves as the real-time services layer for Loomio.

## Related Documents

- `research/loomio_initial_investigation.md` - Main Rails application investigation
- `research/schema_investigation.md` - Database schema details
- `research/initial_investigation_review.md` - Review with gaps and corrections

## Overview

The channel server is a lightweight Node.js application (~200 lines of code) that handles three critical real-time features for Loomio:

1. **Live Updates** - Real-time comment/record updates via Socket.io
2. **Collaborative Editing** - Real-time document editing via Hocuspocus (Yjs)
3. **Matrix Bot Integration** - Chat notifications via Matrix protocol

All components communicate with the main Rails application through **Redis pub/sub channels**.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           loomio_channel_server                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────────┐  │
│  │   index.js      │───▶│   records.js    │    │    hocuspocus.mjs       │  │
│  │   (entry point) │    │   (Socket.io)   │    │    (separate process)   │  │
│  │                 │───▶│                 │    │                         │  │
│  │                 │    └────────┬────────┘    └───────────┬─────────────┘  │
│  │                 │             │                         │                │
│  │                 │───▶┌────────┴────────┐                │                │
│  │                 │    │    bots.js      │                │                │
│  │                 │    │   (Matrix Bot)  │                │                │
│  └────────┬────────┘    └────────┬────────┘                │                │
│           │                      │                         │                │
│           ▼                      ▼                         ▼                │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────────┐  │
│  │    bugs.js      │    │                 │    │                         │  │
│  │  (Sentry/logs)  │    │      Redis      │    │     Rails App API       │  │
│  └─────────────────┘    │    (pub/sub)    │    │   POST /api/hocuspocus  │  │
│                         └────────┬────────┘    └─────────────────────────┘  │
│                                  │                                          │
└──────────────────────────────────┼──────────────────────────────────────────┘
                                   │
                                   ▼
                         ┌─────────────────┐
                         │   Rails App     │
                         │   (Loomio)      │
                         └─────────────────┘
```

## File-by-File Analysis

### 1. `index.js` - Entry Point

**Location**: `orig/loomio_channel_server/index.js`
**Lines**: 12

```javascript
// index.js:1-11
"use strict";
const bugs = require('./bugs.js')
const records = require('./records.js')
const bots = require('./bots.js')

try {
  records();
  bots();
} catch (e) {
  bugs.log(e);
}
```

**Analysis**:
- Boots both the Socket.io server (`records`) and Matrix bot listener (`bots`)
- Wraps startup in try/catch with Sentry error reporting
- Note: `hocuspocus.mjs` is **NOT** started here - it runs as a separate process via `npm run hocuspocus`

---

### 2. `records.js` - Socket.io Server

**Location**: `orig/loomio_channel_server/records.js`
**Lines**: 61

#### Configuration (lines 1-14)

```javascript
// records.js:1-14
"use strict";

const bugs = require('./bugs.js')
const { Server } = require("socket.io");
const redis = require('redis').createClient({
  url: (process.env.REDIS_URL || 'redis://localhost:6379/0')
});

const publicAppUrl = process.env.PUBLIC_APP_URL ||
  "https://" + (process.env.CANONICAL_HOST ||
  (process.env.VIRTUAL_HOST || '').replace('channels.',''))

const config = {
  port: (process.env.PORT || 5000),
  allowedOrigin: publicAppUrl,
}
```

**Key Points**:
- Port defaults to 5000
- CORS origin derived from multiple env vars with fallback chain
- Redis URL defaults to localhost

#### Server Setup (lines 16-27)

```javascript
// records.js:16-27
module.exports = async () => {
  try {
    const io = new Server(config.port, {
      connectionStateRecovery: {
        maxDisconnectionDuration: 30 * 60 * 1000,  // 30 minutes
        skipMiddlewares: true,
      },
      cors: {
        origin: config.allowedOrigin,
        credentials: true
      }
    })
```

**Key Points**:
- Uses Socket.io v4's **Connection State Recovery** feature
- 30-minute disconnection window allows clients to reconnect and receive missed events
- `skipMiddlewares: true` on reconnection for performance

#### Redis Subscriptions (lines 29-42)

```javascript
// records.js:29-42
    redis.on('error', (err) => bugs.log(err) );
    await redis.connect();

    const redisSub = redis.duplicate();
    await redisSub.connect();

    await redisSub.subscribe('/records', (json, channel) => {
      let data = JSON.parse(json)
      io.to(data.room).emit('records', data)
    })

    await redisSub.subscribe('/system_notice', (json, channel) => {
      io.emit('notice', JSON.parse(json))
    })
```

**Key Points**:
- Creates duplicate Redis connection for pub/sub (required by node-redis v4)
- **Channel `/records`**: Broadcasts to specific room (`io.to(data.room)`)
- **Channel `/system_notice`**: Broadcasts to ALL connected sockets (`io.emit`)

#### Connection Handler (lines 44-56)

```javascript
// records.js:44-56
    io.on("connection", async (socket) => {
      socket.join("notice")

      let channel_token = socket.handshake.query.channel_token
      let user = await redis.get("/current_users/"+channel_token)

      if (user) {
        user = JSON.parse(user)
        socket.join("user-"+user.id)
        user.group_ids.forEach(groupId => { socket.join("group-"+groupId) })
        console.log("have current user!", user.name, user.group_ids)
      }
    })
```

**Key Points**:
- All sockets join `notice` room for system-wide broadcasts
- User authentication via `channel_token` query parameter
- **The `channel_token` = `user.secret_token`** from Rails (auto-generated UUID, see `orig/loomio/app/models/user.rb`)
- User data fetched from Redis key `/current_users/{token}`
- Sockets join user-specific room: `user-{id}`
- Sockets join group-specific rooms: `group-{groupId}` for each group membership

**Rails Side**: The Rails app populates `/current_users/{secret_token}` in Redis when users log in. See `loomio_initial_investigation.md` Section 4.1 for User model details.

**Redis Key Pattern**: `/current_users/{channel_token}` stores JSON:
```json
{
  "id": 123,
  "name": "User Name",
  "group_ids": [1, 2, 3]
}
```

---

### 3. `bots.js` - Matrix Bot Integration

**Location**: `orig/loomio_channel_server/bots.js`
**Lines**: 50

#### Setup (lines 1-22)

```javascript
// bots.js:1-22
"use strict";

const bugs = require('./bugs.js')
const MatrixSDK = require("matrix-bot-sdk");
const MatrixClient = MatrixSDK.MatrixClient;

const redis = require('redis').createClient({
  url: (process.env.REDIS_URL || 'redis://localhost:6379/0')
});

const bots = {};  // Cache for Matrix clients

console.log("booting bots!");
module.exports = async () => {
  try {
    redis.on('error', (err) => bugs.log('bots redis client error', err));
    await redis.connect();
    let bots = {}

    const subscriber = redis.duplicate();
    await subscriber.connect();
    console.log("bot redis connected");
```

**Key Points**:
- Uses pattern subscription to listen to all `chatbot/*` channels
- Maintains a cache of Matrix clients (`bots` object) to reuse connections

#### Channel Handlers (lines 23-45)

```javascript
// bots.js:23-45
    await subscriber.pSubscribe('chatbot/*', (json, channel) => {
      console.log(`bot message: channel: ${channel}, json: ${json}`);

      const params = JSON.parse(json);

      if (channel == 'chatbot/test') {
        const client = new MatrixClient(params['server'], params['access_token']);
        client.resolveRoom(params['channel']).then((roomId) => {
          client.sendMessage(roomId, {"msgtype": "m.notice", "body": params['message']});
        })
      }

      if (channel == 'chatbot/publish') {
        const key = JSON.stringify(params.config)
        if (!bots[key]) {
          bots[key] = new MatrixClient(params.config.server, params.config.access_token);
        }

        bots[key].resolveRoom(params.config.channel).then((roomId) => {
          bots[key].sendHtmlText(roomId, params.payload.html);
        })
      }
    });
```

**Channel `chatbot/test`** - Test Messages:
```json
{
  "server": "https://matrix.org",
  "access_token": "syt_...",
  "channel": "#room:matrix.org",
  "message": "Test message"
}
```

**Channel `chatbot/publish`** - Production Messages:
```json
{
  "config": {
    "server": "https://matrix.org",
    "access_token": "syt_...",
    "channel": "#room:matrix.org"
  },
  "payload": {
    "html": "<p>HTML formatted message</p>"
  }
}
```

**Key Differences**:
- `chatbot/test`: Creates new client each time, sends plain text (`m.notice`)
- `chatbot/publish`: Caches clients by config, sends HTML (`sendHtmlText`)

---

### 4. `hocuspocus.mjs` - Collaborative Editing Server

**Location**: `orig/loomio_channel_server/hocuspocus.mjs`
**Lines**: 62

#### Sentry Setup (lines 1-8)

```javascript
// hocuspocus.mjs:1-8
"use strict";
import * as Sentry from "@sentry/node";
const dsn = process.env.SENTRY_PUBLIC_DSN || process.env.SENTRY_DSN

if (dsn) {
  console.log("sentry dsn: ", dsn);
  Sentry.init({dsn});
}
```

#### Auth URL Configuration (lines 14-19)

```javascript
// hocuspocus.mjs:14-19
// trying make things backwards compativle for people doing ./update.sh
// hocuspocus calling back to rails server to auth the connecting browser
const authUrl = (process.env.PRIVATE_APP_URL ||
                 process.env.APP_URL ||
                 process.env.PUBLIC_APP_URL ||
                `https://${process.env.CANONICAL_HOST}`) + '/api/hocuspocus'
```

**Important Comment** (line 14): This comment indicates backwards compatibility concerns for users upgrading via `./update.sh`. The auth URL has a fallback chain for different deployment configurations.

#### Port Configuration (line 21)

```javascript
// hocuspocus.mjs:21
const port = (process.env.RAILS_ENV == 'production') ? 5000 : 4444
```

**Key Point**: Different ports for production vs development to avoid conflicts with the main Socket.io server.

#### Server Configuration (lines 25-51)

```javascript
// hocuspocus.mjs:25-51
const server = new Server({
  port: port,
  timeout: 30000,         // 30 second connection timeout
  debounce: 5000,         // 5 second debounce for persistence
  maxDebounce: 30000,     // 30 second max debounce
  quiet: true,
  name: "hocuspocus",
  extensions: [
    new Logger(),
    new SQLite({database: ''}), // anonymous database on disk
  ],
  async onAuthenticate(data) {
    const { token, documentName } = data;
    const response = await fetch(authUrl, {
        method: 'POST',
        body: JSON.stringify({ user_secret: token, document_name: documentName }),
        headers: { 'Content-type': 'application/json; charset=UTF-8' },
    })
    console.debug(`hocuspocus debug post: ${token} ${documentName} ${response.status}`);

    if (response.status != 200) {
      throw new Error("Not authorized!");
    } else {
      return true;
    }
  },
});
```

**Configuration Options**:
| Option | Value | Purpose |
|--------|-------|---------|
| `timeout` | 30000ms | Connection timeout |
| `debounce` | 5000ms | Delay before persisting changes |
| `maxDebounce` | 30000ms | Maximum delay before forced persistence |
| `quiet` | true | Suppress verbose logging |

**Extensions**:
- `Logger`: Logs server events
- `SQLite({database: ''})`: Uses anonymous SQLite database for persistence

**Authentication Flow**:
1. Client connects with `token` (user_secret) and `documentName`
2. Server POSTs to Rails app at `/api/hocuspocus`
3. Request body: `{ user_secret: token, document_name: documentName }`
4. If Rails returns 200: access granted
5. Otherwise: throws "Not authorized!" error

**Token Format** (from Rails): `{user_id},{secret_token}` (e.g., `123,abc-def-ghi`)

**Document Name Format**: `{record_type}-{record_id}-{user_id_if_new}`

**Supported Record Types** (9 total):
- `comment`, `discussion`, `poll`, `stance`, `outcome`
- `pollTemplate`, `discussionTemplate`, `group`, `user`

**Persistence**: The SQLite extension stores Yjs documents as binary blobs. This answers the question from `initial_investigation_review.md` Section 4.1 about how Y.js documents are persisted.

**Conflict Resolution**: Handled automatically by Yjs CRDT - concurrent edits merge deterministically without conflicts.

#### Startup (lines 53-61)

```javascript
// hocuspocus.mjs:53-61
if (dsn) {
  try {
    server.listen();
  } catch (e) {
    Sentry.captureException(e);
  }
} else {
  server.listen();
}
```

**Note**: Error handling only wraps Sentry capture when DSN is configured.

---

### 5. `bugs.js` - Error Logging

**Location**: `orig/loomio_channel_server/bugs.js`
**Lines**: 21

```javascript
// bugs.js:1-20
"use strict";
const dsn = process.env.SENTRY_PUBLIC_DSN || process.env.SENTRY_DSN
const Sentry = require("@sentry/node");
const SentryTracing = require("@sentry/tracing");

if (dsn) {
	console.log("have DSN:", dsn)
  Sentry.init({ dsn: dsn, tracesSampleRate: 0.1 });
}

module.exports = {
	log: (e) => {
		if (dsn) {
			Sentry.captureException(e);
		}else{
			console.log("error:", e);
		}
	}
}
```

**Key Points**:
- `tracesSampleRate: 0.1` - Only 10% of transactions are traced
- Falls back to `console.log` when no DSN configured
- Exports single `log(e)` function for error capture

---

## Dependencies

### Runtime Dependencies

| Package | Version | Purpose | Used In | npm | Docs |
|---------|---------|---------|---------|-----|------|
| `socket.io` | ^4.7.1 | Real-time WebSocket communication | `records.js` | [npm](https://www.npmjs.com/package/socket.io) | [Docs](https://socket.io/docs/v4/) |
| `redis` | ^4.6.15 | Redis client for pub/sub | `records.js`, `bots.js` | [npm](https://www.npmjs.com/package/redis) | [Docs](https://github.com/redis/node-redis) |
| `@hocuspocus/server` | ^3.4.0 | Yjs WebSocket backend for collaborative editing | `hocuspocus.mjs` | [npm](https://www.npmjs.com/package/@hocuspocus/server) | [Docs](https://tiptap.dev/docs/hocuspocus/introduction) |
| `@hocuspocus/extension-sqlite` | ^3.4.0 | SQLite persistence for Hocuspocus | `hocuspocus.mjs` | [npm](https://www.npmjs.com/package/@hocuspocus/extension-sqlite) | [Docs](https://tiptap.dev/docs/hocuspocus/server/extensions) |
| `@hocuspocus/extension-logger` | ^3.4.0 | Logging extension for Hocuspocus | `hocuspocus.mjs` | [npm](https://www.npmjs.com/package/@hocuspocus/extension-logger) | [Docs](https://tiptap.dev/docs/hocuspocus/server/extensions) |
| `@hocuspocus/extension-database` | ^3.4.0 | Database abstraction for Hocuspocus | Not used (declared) | [npm](https://www.npmjs.com/package/@hocuspocus/extension-database) | - |
| `matrix-bot-sdk` | ^0.5.19 | Matrix protocol bot SDK | `bots.js` | [npm](https://www.npmjs.com/package/matrix-bot-sdk) | [Docs](https://github.com/turt2live/matrix-bot-sdk) |
| `@sentry/node` | ^7.118.0 | Error tracking and monitoring | `bugs.js`, `hocuspocus.mjs` | [npm](https://www.npmjs.com/package/@sentry/node) | [Docs](https://docs.sentry.io/platforms/javascript/guides/node/) |
| `@sentry/tracing` | ^7.114.0 | Performance tracing for Sentry | `bugs.js` | [npm](https://www.npmjs.com/package/@sentry/tracing) | [Docs](https://docs.sentry.io/platforms/javascript/guides/node/tracing/) |
| `dotenv` | ^16.0.1 | Environment variable loading | Not explicitly used | [npm](https://www.npmjs.com/package/dotenv) | [Docs](https://github.com/motdotla/dotenv) |

### Peer Dependencies (Implicit)

| Package | Purpose | Notes |
|---------|---------|-------|
| `yjs` | CRDT implementation | Required by Hocuspocus for collaborative editing |

---

## Dependency Deep Dives

### Socket.io v4

**Purpose**: Bidirectional real-time communication between browser clients and the server.

**Key Features Used**:
- **Connection State Recovery** (`records.js:19-22`): Allows clients to reconnect within 30 minutes and receive missed events
- **Rooms** (`records.js:45,52,53`): Organize sockets into channels (user rooms, group rooms, notice room)
- **CORS** (`records.js:23-26`): Cross-origin support for web clients
- **Emit to rooms** (`records.js:37`): `io.to(room).emit('event', data)`
- **Broadcast to all** (`records.js:41`): `io.emit('event', data)`

**API Summary**:
```javascript
// Server creation
const io = new Server(port, options)

// Room management
socket.join("room-name")

// Emit to specific room
io.to("room-name").emit("event", data)

// Broadcast to all
io.emit("event", data)

// Connection event
io.on("connection", (socket) => { ... })
```

### Redis (node-redis v4)

**Purpose**: Pub/sub messaging between Rails app and channel server.

**Key Features Used**:
- **Pub/Sub** (`records.js:35-42`, `bots.js:23`): Subscribe to channels for real-time message relay
- **Pattern Subscribe** (`bots.js:23`): `pSubscribe('chatbot/*')` for wildcard matching
- **Duplicate connection** (`records.js:32`, `bots.js:20`): Required for pub/sub in RESP2 mode
- **Key/Value** (`records.js:48`): Store and retrieve user session data

**API Summary**:
```javascript
// Create client
const client = createClient({ url: 'redis://...' })
await client.connect()

// Duplicate for pub/sub
const subscriber = client.duplicate()
await subscriber.connect()

// Subscribe to channel
await subscriber.subscribe('/channel', (message, channel) => { ... })

// Pattern subscribe
await subscriber.pSubscribe('pattern/*', (message, channel) => { ... })

// Get value
const value = await client.get('/key')
```

### Hocuspocus

**Purpose**: Real-time collaborative editing backend using Yjs CRDT.

**Key Features Used**:
- **Server** (`hocuspocus.mjs:25`): WebSocket server for Yjs documents
- **onAuthenticate hook** (`hocuspocus.mjs:36-49`): Custom authentication via Rails API
- **SQLite extension** (`hocuspocus.mjs:34`): Document persistence
- **Debouncing** (`hocuspocus.mjs:28-29`): Batch persistence operations

**API Summary**:
```javascript
import { Server } from '@hocuspocus/server'
import { SQLite } from '@hocuspocus/extension-sqlite'

const server = new Server({
  port: 4444,
  timeout: 30000,
  debounce: 5000,
  maxDebounce: 30000,
  extensions: [new SQLite({ database: 'path.sqlite' })],
  async onAuthenticate({ token, documentName }) {
    // Return true to allow, throw to deny
  }
})

server.listen()
```

### Matrix Bot SDK

**Purpose**: Send notifications to Matrix chat rooms.

**Key Features Used**:
- **MatrixClient** (`bots.js:29,38`): Client for Matrix protocol
- **resolveRoom** (`bots.js:30,41`): Convert room alias to room ID
- **sendMessage** (`bots.js:31`): Send plain text message
- **sendHtmlText** (`bots.js:42`): Send HTML formatted message

**API Summary**:
```javascript
import { MatrixClient } from 'matrix-bot-sdk'

const client = new MatrixClient(homeserverUrl, accessToken)

// Resolve room alias to ID
const roomId = await client.resolveRoom('#room:server.org')

// Send plain text notice
await client.sendMessage(roomId, { msgtype: 'm.notice', body: 'text' })

// Send HTML message
await client.sendHtmlText(roomId, '<p>HTML content</p>')
```

### Sentry

**Purpose**: Error tracking and performance monitoring.

**Key Features Used**:
- **init** (`bugs.js:8`, `hocuspocus.mjs:7`): Initialize SDK with DSN
- **captureException** (`bugs.js:15`, `hocuspocus.mjs:57`): Report errors
- **tracesSampleRate** (`bugs.js:8`): Control performance tracing sample rate

**API Summary**:
```javascript
const Sentry = require('@sentry/node')

Sentry.init({
  dsn: 'https://...@sentry.io/...',
  tracesSampleRate: 0.1  // 10% of transactions
})

// Capture error
Sentry.captureException(error)
```

---

## Configuration Reference

### Environment Variables

| Variable | Default | Used In | Purpose |
|----------|---------|---------|---------|
| `PORT` | `5000` | `records.js:12` | Socket.io server port |
| `REDIS_URL` | `redis://localhost:6379/0` | `records.js:6`, `bots.js:8` | Redis connection URL |
| `PUBLIC_APP_URL` | (computed) | `records.js:9`, `hocuspocus.mjs:18` | CORS origin, auth URL fallback |
| `CANONICAL_HOST` | - | `records.js:9`, `hocuspocus.mjs:19` | Primary hostname |
| `VIRTUAL_HOST` | - | `records.js:9` | Docker/proxy hostname (strips `channels.` prefix) |
| `PRIVATE_APP_URL` | - | `hocuspocus.mjs:16` | Internal Rails app URL for auth |
| `APP_URL` | - | `hocuspocus.mjs:17` | Alternative Rails app URL |
| `SENTRY_PUBLIC_DSN` | - | `bugs.js:2`, `hocuspocus.mjs:3` | Sentry DSN (primary) |
| `SENTRY_DSN` | - | `bugs.js:2`, `hocuspocus.mjs:3` | Sentry DSN (fallback) |
| `RAILS_ENV` | - | `hocuspocus.mjs:21` | Controls Hocuspocus port (5000 prod, 4444 dev) |

### URL Resolution Logic

**CORS Origin** (`records.js:9`):
```
PUBLIC_APP_URL || "https://" + (CANONICAL_HOST || VIRTUAL_HOST.replace('channels.',''))
```

**Hocuspocus Auth URL** (`hocuspocus.mjs:16-19`):
```
(PRIVATE_APP_URL || APP_URL || PUBLIC_APP_URL || "https://" + CANONICAL_HOST) + '/api/hocuspocus'
```

---

## Redis Channel Protocol

> **Note**: The Rails side of this protocol is documented in `loomio_initial_investigation.md` Section 8 (Real-time System). The channel server is the subscriber/consumer side.

### Channel: `/records`

**Publisher**: Rails app
**Subscriber**: `records.js`
**Purpose**: Broadcast record updates to specific rooms

**Payload**:
```json
{
  "room": "user-123",
  "type": "comment",
  "data": { ... }
}
```

**Room Types**:
- `user-{id}` - Messages for specific user
- `group-{id}` - Messages for group members

### Channel: `/system_notice`

**Publisher**: Rails app
**Subscriber**: `records.js`
**Purpose**: System-wide announcements to all connected clients

**Payload**:
```json
{
  "type": "maintenance",
  "message": "System will be down for maintenance"
}
```

### Channel: `chatbot/test`

**Publisher**: Rails app
**Subscriber**: `bots.js`
**Purpose**: Test Matrix bot configuration

**Payload**:
```json
{
  "server": "https://matrix.org",
  "access_token": "syt_...",
  "channel": "#room:matrix.org",
  "message": "Test message"
}
```

### Channel: `chatbot/publish`

**Publisher**: Rails app
**Subscriber**: `bots.js`
**Purpose**: Send notifications to Matrix rooms

**Payload**:
```json
{
  "config": {
    "server": "https://matrix.org",
    "access_token": "syt_...",
    "channel": "#room:matrix.org"
  },
  "payload": {
    "html": "<p>Notification content</p>"
  }
}
```

### Redis Key: `/current_users/{channel_token}`

**Writer**: Rails app
**Reader**: `records.js`
**Purpose**: User session data for WebSocket authentication

**Value**:
```json
{
  "id": 123,
  "name": "User Name",
  "group_ids": [1, 2, 3]
}
```

---

## Recommended Test Cases

Since no tests currently exist, here are the recommended test cases for comprehensive coverage:

### Unit Tests

#### `bugs.js`

| Test Case | Description |
|-----------|-------------|
| `log() with DSN` | Should call `Sentry.captureException()` when DSN is configured |
| `log() without DSN` | Should call `console.log()` when no DSN |
| `init with DSN` | Should initialize Sentry with correct tracesSampleRate (0.1) |

#### `records.js`

| Test Case | Description |
|-----------|-------------|
| Server starts | Should start Socket.io server on configured port |
| CORS configuration | Should configure CORS with correct origin |
| Redis connection | Should connect to Redis with correct URL |
| Redis error handling | Should log errors via bugs.log() |
| `/records` subscription | Should emit to correct room when message received |
| `/system_notice` subscription | Should broadcast to all sockets |
| Connection - join notice room | New socket should join "notice" room |
| Connection - valid token | Should join user and group rooms when valid token |
| Connection - invalid token | Should only join notice room when token invalid |
| Connection State Recovery | Should reconnect within 30 minute window |

#### `bots.js`

| Test Case | Description |
|-----------|-------------|
| Redis connection | Should connect and duplicate for pub/sub |
| Pattern subscription | Should subscribe to `chatbot/*` pattern |
| `chatbot/test` handler | Should create new client and send notice message |
| `chatbot/publish` handler | Should cache client and send HTML message |
| `chatbot/publish` client reuse | Should reuse cached client for same config |
| Error handling | Should log errors via bugs.log() |

#### `hocuspocus.mjs`

| Test Case | Description |
|-----------|-------------|
| Server starts (dev) | Should start on port 4444 when RAILS_ENV != production |
| Server starts (prod) | Should start on port 5000 when RAILS_ENV == production |
| Auth URL resolution | Should build correct URL from env vars |
| onAuthenticate - success | Should return true when Rails returns 200 |
| onAuthenticate - failure | Should throw when Rails returns non-200 |
| Sentry integration | Should capture exceptions when DSN configured |
| SQLite extension | Should initialize with anonymous database |

### Integration Tests

| Test Case | Description |
|-----------|-------------|
| End-to-end records flow | Redis publish → Socket.io emit to room |
| End-to-end notice flow | Redis publish → Socket.io broadcast |
| Matrix bot test message | Redis publish → Matrix message sent |
| Matrix bot publish message | Redis publish → HTML message sent |
| Hocuspocus auth flow | Client connect → Rails API call → access granted/denied |

### Mock Strategies

| Component | Mock Library | Purpose |
|-----------|--------------|---------|
| Redis | `redis-mock` or `ioredis-mock` | Mock pub/sub and key/value |
| Socket.io | `socket.io-client` + `socket.io-mock` | Test client connections |
| Matrix SDK | Jest mock | Mock MatrixClient methods |
| Sentry | Jest mock | Verify captureException calls |
| fetch (Hocuspocus) | `nock` or `msw` | Mock Rails API responses |

---

## Important Comments & Notes

### Backwards Compatibility Comment

**Location**: `hocuspocus.mjs:14`
```javascript
// trying make things backwards compativle for people doing ./update.sh
```

**Context**: This comment indicates that the auth URL resolution logic has multiple fallbacks to support users who upgrade via the `./update.sh` script. The fallback chain ensures the server works with different deployment configurations.

### Logging Statements

**Location**: `records.js:54`
```javascript
console.log("have current user!", user.name, user.group_ids)
```
**Note**: Debug logging that could be noisy in production.

**Location**: `bots.js:13`
```javascript
console.log("booting bots!");
```
**Note**: Startup confirmation logging.

**Location**: `bots.js:24`
```javascript
console.log(`bot message: channel: ${channel}, json: ${json}`);
```
**Note**: Debug logging of all bot messages - could be verbose.

---

## Cross-Document Analysis: Issues Found

### Contradictions & Errors

#### 1. Attachments JSONB Default - Review Document Has Error

**`schema_investigation.md` line 537:**
> `discussions.attachments jsonb DEFAULT '[]'::jsonb`

**`initial_investigation_review.md` Section 1.3 claims:**
> Actual: Default: `'{}'::jsonb` (empty object, not empty array)

**VERIFIED against `db/schema.rb` lines 189, 259, 298, etc.:**
```ruby
t.jsonb "attachments", default: [], null: false
```

**Conclusion**: `schema_investigation.md` is **CORRECT**. The `initial_investigation_review.md` Section 1.3 has an error - the default IS `[]` (empty array), not `{}` (empty object).

#### 2. Hocuspocus Token Format - Clarification Needed

**`initial_investigation_review.md` Section 4.1 states:**
> Token format: `{user_id},{secret_token}`

**Actual (`hocuspocus.mjs:40`):**
```javascript
body: JSON.stringify({ user_secret: token, document_name: documentName })
```

**Clarification**: The channel server passes `token` through unchanged. The token format `{user_id},{secret_token}` is assembled by the **client**, not the channel server. The format parsing happens in **Rails** at `/api/hocuspocus`. Neither the main investigation nor this one clarifies where in Rails this parsing occurs.

#### 3. Port 5000 Conflict in Production

**My investigation:**
- `records.js` (Socket.io): PORT env var (default 5000)
- `hocuspocus.mjs`: port 5000 (production) or 4444 (development)

**Issue**: Both services use port 5000 in production. This only works because:
1. They run as **separate processes/containers** (see `Procfile` vs `npm run hocuspocus`)
2. In Docker deployments, they're likely on different containers

**Missing from main investigation**: This deployment architecture isn't documented. The `loomio-deploy` submodule should clarify this.

### Gaps in Main Investigation Documents

#### 4. Redis Publishing Locations - NOW DOCUMENTED

**What channel server reads:**
- `/current_users/{token}` - user session cache
- `/records` - record updates
- `/system_notice` - system notices
- `chatbot/*` - bot messages

**FOUND - Rails locations that WRITE to Redis:**

| Redis Channel/Key | Rails File | Method | Line |
|-------------------|------------|--------|------|
| `/current_users/{token}` | `app/controllers/api/v1/boot_controller.rb` | `set_channel_token` | 26-32 |
| `/records` | `app/services/message_channel_service.rb` | `publish_serialized_records` | 17-23 |
| `/system_notice` | `app/services/message_channel_service.rb` | `publish_system_notice` | 25-31 |

**`boot_controller.rb:26-32`** - Populates user session for WebSocket auth:
```ruby
def set_channel_token
  CACHE_REDIS_POOL.with do |client|
    client.set("/current_users/#{current_user.secret_token}",
      {name: current_user.name,
       group_ids: current_user.group_ids,
       id: current_user.id}.to_json)
  end
end
```

**`message_channel_service.rb:17-23`** - Publishes record updates:
```ruby
def self.publish_serialized_records(data, group_id: nil, user_id: nil)
  CACHE_REDIS_POOL.with do |client|
    room = "user-#{user_id}" if user_id
    room = "group-#{group_id}" if group_id
    client.publish("/records", {room: room, records: data}.to_json)
  end
end
```

**Note**: `chatbot/*` channels are still undocumented - need to search for chatbot publishing code.

#### 5. Event Kinds → Redis Channel Mapping Missing

**`initial_investigation_review.md`** notes 42 event kinds but only 14 are webhook-eligible.

**Question**: Which event kinds trigger Redis pub/sub to the channel server? This mapping isn't documented. The EventBus broadcasts events, but where does it decide to publish to Redis?

#### 6. SQLite Anonymous Database - Operational Concern

**`hocuspocus.mjs:34`:**
```javascript
new SQLite({database: ''})  // anonymous database on disk
```

**Issue**: Empty string for SQLite database means a temporary file. If the hocuspocus process restarts, **all collaborative editing state may be lost**.

**Questions**:
1. Is this intentional (ephemeral editing state)?
2. How does production handle hocuspocus restarts?
3. Should the Go rewrite use persistent storage?

This isn't addressed in any document.

#### 7. update.sh Script Referenced but Undocumented

**`hocuspocus.mjs:14`:**
```javascript
// trying make things backwards compativle for people doing ./update.sh
```

**Issue**: This script is referenced but not documented. Where is it? What does it do? The `loomio-deploy` submodule may contain it.

### Inconsistencies Within My Investigation

#### 8. Matrix Bot Behavioral Difference

I documented but didn't highlight this significant difference:

| Channel | Client Creation | Message Type |
|---------|-----------------|--------------|
| `chatbot/test` | **New client each time** | Plain text (`m.notice`) |
| `chatbot/publish` | **Cached by config** | HTML (`sendHtmlText`) |

This inconsistency in the original code means:
- Test messages could fail silently if Matrix server is down
- Production messages reuse connections (better performance)
- Potential memory leak if many different configs are used

### New Questions Raised

| Question | Impact | Status |
|----------|--------|--------|
| Where does Rails populate `/current_users/{token}`? | Go rewrite needs this | **RESOLVED** - `boot_controller.rb:26-32` |
| Where does Rails publish to `/records`? | Go rewrite needs this | **RESOLVED** - `message_channel_service.rb:17-23` |
| Where does Rails publish to `chatbot/*`? | Go rewrite needs this | **OPEN** - Search for chatbot publish code |
| Is hocuspocus state intentionally ephemeral? | Architecture decision | **OPEN** - Check loomio-deploy |
| Where is `update.sh`? | Deployment understanding | **OPEN** - Search loomio-deploy submodule |
| How do multiple channel server instances coordinate? | Scalability | **LIKELY OK** - Redis pub/sub naturally broadcasts to all subscribers |
| Which events trigger Redis pub/sub? | Event system mapping | **OPEN** - Trace EventBus → MessageChannelService calls |

---

## Answers to Review Document Questions

This investigation resolves several unanswered questions from `initial_investigation_review.md` Section 4:

| Question | Answer |
|----------|--------|
| How are Y.js documents stored/persisted? | SQLite extension stores Yjs documents as binary blobs in an anonymous database (`hocuspocus.mjs:34`) |
| How is conflict resolution handled? | Yjs CRDT handles this automatically - concurrent edits merge deterministically |
| What happens when documents are edited offline? | Yjs syncs changes when reconnected; the 30-second debounce (`hocuspocus.mjs:28-29`) batches persistence |

---

## Recommendations for Go Rewrite

### Library Equivalents

| Node.js | Go Equivalent | Notes |
|---------|---------------|-------|
| `socket.io` | `github.com/googollee/go-socket.io` or `github.com/olahol/melody` + custom protocol | go-socket.io has protocol compatibility; melody is simpler but needs custom events |
| `redis` | `github.com/redis/go-redis/v9` | Official Redis client with pub/sub support |
| `@hocuspocus/server` | `github.com/yjs/y-websocket` (reference) + custom Go implementation | No direct Go equivalent; would need Yjs CRDT implementation |
| `matrix-bot-sdk` | `maunium.net/go/mautrix` | Full-featured Matrix client library |
| `@sentry/node` | `github.com/getsentry/sentry-go` | Official Sentry SDK for Go |

### Architecture Considerations

1. **Process Model**:
   - Node.js runs two processes (main + hocuspocus)
   - Go could run both in single process with goroutines, or keep separate for isolation

2. **WebSocket Handling**:
   - Consider using `gorilla/websocket` as the base
   - Socket.io protocol compatibility may require `go-socket.io`
   - For simpler needs, custom WebSocket + JSON protocol could suffice

3. **Hocuspocus/Yjs Complexity**:
   - No production-ready Go implementation of Yjs CRDT
   - Options:
     - Keep Node.js hocuspocus as separate service
     - Implement subset of Yjs protocol
     - Use alternative CRDT library (e.g., Automerge has Go bindings)

4. **Concurrency**:
   - Go's goroutines map well to this use case
   - Use channels for pub/sub relay internally
   - Consider connection pooling for Redis

5. **Configuration**:
   - Use `github.com/kelseyhightower/envconfig` or `viper` for env vars
   - Maintain same environment variable names for backwards compatibility

6. **Error Handling**:
   - Sentry Go SDK integrates similarly
   - Consider structured logging with `zerolog` or `zap`

### Migration Path

1. **Phase 1**: Rewrite `records.js` (Socket.io server) - highest impact, most straightforward
2. **Phase 2**: Rewrite `bots.js` (Matrix bot) - isolated, clear boundaries
3. **Phase 3**: Evaluate Hocuspocus options - may keep Node.js or implement subset
4. **Phase 4**: Consolidate into single Go service with graceful degradation
