# Loomio External Services Integration

**Generated:** 2026-02-01
**Purpose:** Document all external service integrations, configuration, and API interactions

---

## Table of Contents

1. [OAuth Providers](#1-oauth-providers)
2. [Cloud Storage](#2-cloud-storage)
3. [Email Services](#3-email-services)
4. [Real-time Services](#4-real-time-services)
5. [Error Tracking](#5-error-tracking)
6. [Translation Services](#6-translation-services)
7. [Billing Integration](#7-billing-integration)
8. [Webhooks and Chatbots](#8-webhooks-and-chatbots)
9. [AI/ML Services](#9-aiml-services)
10. [Analytics](#10-analytics)
11. [GeoIP](#11-geoip)

---

## 1. OAuth Providers

### 1.1 Google OAuth

**Purpose:** Single sign-on authentication via Google accounts

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `GOOGLE_APP_KEY` | OAuth client ID |
| `GOOGLE_APP_SECRET` | OAuth client secret |

**API Interactions:**

```
Authorization URL: https://accounts.google.com/o/oauth2/v2/auth
Token URL: https://oauth2.googleapis.com/token
Profile URL: https://www.googleapis.com/oauth2/v1/userinfo

Scopes requested:
  - email
  - profile

User attributes fetched:
  - uid (id)
  - email
  - name
```

**Controller:** `/app/controllers/identities/google_controller.rb`

**Flow:**
1. User clicks "Sign in with Google"
2. Redirect to Google authorization URL
3. Google redirects back with authorization code
4. Exchange code for access token
5. Fetch user profile
6. Find/create Identity and link to User

---

### 1.2 Generic OAuth 2.0

**Purpose:** Support for custom OAuth 2.0 providers (enterprise SSO)

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `OAUTH_APP_KEY` | OAuth client ID |
| `OAUTH_APP_SECRET` | OAuth client secret |
| `OAUTH_AUTH_URL` | Authorization endpoint |
| `OAUTH_TOKEN_URL` | Token endpoint |
| `OAUTH_PROFILE_URL` | User info endpoint |
| `OAUTH_SCOPE` | OAuth scopes to request |
| `OAUTH_ATTR_UID` | JSON path to user ID |
| `OAUTH_ATTR_NAME` | JSON path to user name |
| `OAUTH_ATTR_EMAIL` | JSON path to user email |

**Controller:** `/app/controllers/identities/oauth_controller.rb`

**Use Case:** Integrating with corporate identity providers that support standard OAuth 2.0

---

### 1.3 Nextcloud OAuth

**Purpose:** Single sign-on for Nextcloud installations

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `NEXTCLOUD_HOST` | Nextcloud server URL |
| `NEXTCLOUD_APP_KEY` | OAuth client ID |
| `NEXTCLOUD_APP_SECRET` | OAuth client secret |

**API Interactions:**

```
Authorization URL: {NEXTCLOUD_HOST}/index.php/apps/oauth2/authorize
Token URL: {NEXTCLOUD_HOST}/index.php/apps/oauth2/api/v1/token
Profile URL: {NEXTCLOUD_HOST}/ocs/v1.php/cloud/users/{uid}
```

**Controller:** `/app/controllers/identities/nextcloud_controller.rb`

---

### 1.4 SAML 2.0

**Purpose:** Enterprise single sign-on via SAML identity providers

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `SAML_IDP_METADATA_URL` | IdP metadata endpoint |
| `SAML_IDP_SSO_TARGET_URL` | IdP SSO endpoint |
| `SAML_IDP_SLO_TARGET_URL` | IdP logout endpoint (optional) |
| `SAML_IDP_CERT` | IdP certificate for signature validation |
| `SAML_SP_ENTITY_ID` | Service provider entity ID |
| `SAML_NAME_IDENTIFIER_FORMAT` | NameID format |
| `SAML_ATTR_EMAIL` | Attribute name for email |
| `SAML_ATTR_NAME` | Attribute name for display name |

**Controller:** `/app/controllers/identities/saml_controller.rb`

**Endpoints:**

```
GET  /saml/oauth     - Initiate SAML auth request
POST /saml/oauth     - Handle SAML response (ACS)
GET  /saml/metadata  - SP metadata (XML)
```

**Library:** `ruby-saml` gem

---

### 1.5 SSO Behavioral Configuration

| Environment Variable | Description |
|---------------------|-------------|
| `FEATURES_DISABLE_EMAIL_LOGIN` | Force SSO-only authentication |
| `LOOMIO_SSO_FORCE_USER_ATTRS` | Sync name/email from IdP on each login |

When `FEATURES_DISABLE_EMAIL_LOGIN` is true:
- Password login is disabled
- Registration creates users immediately from SSO
- Users cannot be created without SSO identity

---

## 2. Cloud Storage

### 2.1 AWS S3

**Purpose:** File storage for attachments, exports, and user uploads

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |
| `AWS_REGION` | S3 bucket region |
| `AWS_BUCKET` | S3 bucket name |

**Configuration File:** `/config/storage.yml`

```yaml
amazon:
  service: S3
  access_key_id: <%= ENV['AWS_ACCESS_KEY_ID'] %>
  secret_access_key: <%= ENV['AWS_SECRET_ACCESS_KEY'] %>
  region: <%= ENV['AWS_REGION'] %>
  bucket: <%= ENV['AWS_BUCKET'] %>
```

**Usage:**
- Active Storage attachments (logos, cover photos, user avatars)
- File attachments in rich text content
- Export file storage (temporary, auto-deleted after 1 week)

**Library:** `aws-sdk-s3` gem

---

### 2.2 Google Cloud Storage

**Purpose:** Alternative to S3 for GCP deployments

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `GCS_PROJECT` | GCP project ID |
| `GCS_BUCKET` | GCS bucket name |
| `GCS_CREDENTIALS` | Service account JSON (base64 or path) |

**Configuration File:** `/config/storage.yml`

```yaml
google:
  service: GCS
  project: <%= ENV['GCS_PROJECT'] %>
  credentials: <%= ENV['GCS_CREDENTIALS'] %>
  bucket: <%= ENV['GCS_BUCKET'] %>
```

**Library:** `google-cloud-storage` gem

---

### 2.3 Local Storage (Development)

**Purpose:** File storage on local filesystem

**Configuration:**

```yaml
local:
  service: Disk
  root: <%= Rails.root.join("storage") %>
```

Default for development environment.

---

## 3. Email Services

### 3.1 Outbound Email (SMTP)

**Purpose:** Sending notification emails, invitations, and system messages

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `SMTP_SERVER` | SMTP server hostname |
| `SMTP_PORT` | SMTP server port |
| `SMTP_USERNAME` | SMTP authentication username |
| `SMTP_PASSWORD` | SMTP authentication password |
| `SMTP_DOMAIN` | HELO domain |
| `SMTP_AUTH` | Authentication type (plain, login, cram_md5) |
| `SMTP_TLS` | Enable TLS (true/false) |

**Configuration File:** `/config/environments/production.rb`

```ruby
config.action_mailer.smtp_settings = {
  address: ENV['SMTP_SERVER'],
  port: ENV['SMTP_PORT'],
  user_name: ENV['SMTP_USERNAME'],
  password: ENV['SMTP_PASSWORD'],
  domain: ENV['SMTP_DOMAIN'],
  authentication: ENV['SMTP_AUTH'],
  enable_starttls_auto: ENV['SMTP_TLS'] == 'true'
}
```

**Mailer Classes:**
- `UserMailer` - Password reset, email verification
- `EventMailer` - Event notifications
- `GroupMailer` - Group announcements
- `ForwardMailer` - Email bounce notices

---

### 3.2 Inbound Email (Action Mailbox)

**Purpose:** Processing replies sent to discussion/poll notification emails

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `REPLY_HOSTNAME` | Hostname for reply-to addresses |
| `ACTION_MAILBOX_INGRESS_PASSWORD` | Webhook authentication |

**Reply Address Format:**

```
d={discussion_key}+{token}@{REPLY_HOSTNAME}    (discussion reply)
p={poll_key}+{token}@{REPLY_HOSTNAME}          (poll comment)
n={notification_id}+{token}@{REPLY_HOSTNAME}   (notification reply)
```

**Processing Flow:**

1. User replies to email notification
2. Email server forwards to `/rails/action_mailbox/relay/inbound_emails`
3. Action Mailbox parses email
4. `ReceivedEmailService.route` processes based on recipient address
5. Creates Comment or Stance in appropriate context

---

### 3.3 External Email Webhook (Mailin Format)

**Purpose:** Alternative email ingestion via webhook

**Endpoint:** `POST /received_emails`

**Expected Payload:**

```json
{
  "mailinMsg": {
    "html": "<html>...</html>",
    "text": "plain text body",
    "headers": {
      "from": "Sender Name <sender@example.com>",
      "to": "route@reply.loomio.com",
      "subject": "Re: Discussion Title"
    },
    "attachments": [
      {
        "generatedFileName": "attachment.pdf",
        "contentType": "application/pdf"
      }
    ]
  },
  "attachment.pdf": "base64-encoded-content"
}
```

**Controller:** `/app/controllers/received_emails_controller.rb`

---

### 3.4 Email CSS Processing

**Purpose:** Inline CSS for email HTML rendering

**Library:** `premailer-rails` gem

**Function:** Automatically converts stylesheet links and style tags to inline styles in email HTML, ensuring consistent rendering across email clients.

---

## 4. Real-time Services

### 4.1 Hocuspocus (Collaborative Editing)

**Purpose:** Real-time collaborative rich text editing via Yjs CRDT

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `HOCUSPOCUS_URL` | Hocuspocus WebSocket server URL |

**Architecture:**

```
Vue Client (Tiptap Editor)
    |
    +-- @hocuspocus/provider
    |
    v
Hocuspocus Server (WebSocket)
    |
    +-- Y.js Document Sync
    |
    v
Rails Backend
    +-- /api/hocuspocus/authenticate (JWT auth)
    +-- /api/hocuspocus/webhook (document persistence)
```

**Frontend Libraries:**
- `@hocuspocus/provider` - WebSocket client
- `yjs` - CRDT implementation
- `y-indexeddb` - Local persistence
- `@tiptap/extension-collaboration` - Tiptap integration

**Backend Endpoint:** `/app/controllers/api/hocuspocus_controller.rb`

**Authentication:** JWT token issued by Rails, verified by Hocuspocus

---

### 4.2 Redis Channels Service

**Purpose:** Real-time updates for UI (notifications, model changes)

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `CHANNELS_URL` | WebSocket/SSE channels server URL |
| `REDIS_URL` | Redis connection for pub/sub |

**Publishing Flow:**

```ruby
MessageChannelService.publish_models(
  models: [poll],
  serializer: PollSerializer,
  channel: "user-#{user_id}"
)
```

**Channel Types:**
- `user-{id}` - Personal notifications, model updates
- `group-{id}` - Group activity (deprecated in favor of user channels)

**Redis Pub/Sub:**

```ruby
# Publisher (Rails)
REDIS_POOL.with do |redis|
  redis.publish("loomio:#{channel}", payload.to_json)
end

# Subscriber (Channels Service - external)
redis.subscribe("loomio:*") do |on|
  on.message { |channel, payload| broadcast_to_clients(channel, payload) }
end
```

---

### 4.3 Redis Configuration

**Purpose:** Caching, sessions, background jobs, real-time pub/sub

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `REDIS_URL` | Primary Redis URL |
| `REDIS_CACHE_URL` | Cache-specific Redis (optional) |

**Connection Pools:**
- `MAIN_REDIS_POOL` - General purpose
- `CACHE_REDIS_POOL` - Caching operations

**Library:** `redis` gem, `redis-objects` gem

---

## 5. Error Tracking

### 5.1 Sentry

**Purpose:** Error tracking, performance monitoring, and alerting

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `SENTRY_DSN` | Sentry Data Source Name |
| `SENTRY_SAMPLE_RATE` | Transaction sample rate (0.0-1.0) |

**Configuration File:** `/config/initializers/sentry.rb`

```ruby
Sentry.init do |config|
  config.dsn = ENV['SENTRY_DSN']
  config.traces_sample_rate = ENV['SENTRY_SAMPLE_RATE']&.to_f || 0.1
  config.breadcrumbs_logger = [:active_support_logger, :http_logger]
end
```

**Libraries:**
- `sentry-ruby` - Core SDK
- `sentry-rails` - Rails integration
- `sentry-sidekiq` - Background job error tracking

**Frontend:**
- `@sentry/vue` - Vue.js integration
- `@sentry/browser` - Browser error capture

---

## 6. Translation Services

### 6.1 Google Cloud Translate

**Purpose:** Automatic translation of user-generated content

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `GOOGLE_TRANSLATE_PROJECT_ID` | GCP project ID |
| `GOOGLE_TRANSLATE_CREDENTIALS` | Service account JSON |

**Service:** `/app/services/translation_service.rb`

**Usage:**

```ruby
TranslationService.translate(
  text: content,
  source_locale: 'en',
  target_locale: 'es'
)
```

**Features:**
- Translates discussion descriptions, poll details, comments
- Caches translations in `translations` table
- Detects source language using `cld` gem if not specified

**Library:** `google-cloud-translate` gem

---

## 7. Billing Integration

### 7.1 Chargify

**Purpose:** Subscription management and billing

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `CHARGIFY_API_KEY` | Chargify API key |
| `CHARGIFY_SUBDOMAIN` | Chargify site subdomain |
| `CHARGIFY_PRODUCT_HANDLE` | Default product handle |

**Integration Points:**
- Subscription creation on group upgrade
- Plan changes and upgrades
- Cancellation handling
- Webhook processing for billing events

**Engine:** `engines/loomio_subs` (optional, loaded conditionally)

**Note:** Billing integration is optional. When not configured, subscription features are disabled and all groups operate in "free" mode.

---

## 8. Webhooks and Chatbots

### 8.1 Webhook System

**Purpose:** Push notifications to external systems on Loomio events

**Model:** `/app/models/chatbot.rb`

**Supported Platforms:**

| Platform | Kind | Webhook Format |
|----------|------|----------------|
| Slack | slack | Slack Block Kit |
| Discord | discord | Discord embeds |
| Microsoft Teams | teams | Adaptive Cards |
| Mattermost | mattermost | Mattermost attachments |
| Matrix | matrix | Matrix events |
| Webex | webex | Webex cards |
| Generic | webhook | Markdown or JSON |

**Configuration per Chatbot:**

```ruby
chatbot.config
# => {
#      server: "https://hooks.slack.com/...",
#      access_token: "xoxb-...",
#      channel: "#general"
#    }
```

**Event Trigger Flow:**

```
Event published
    |
    v
Events::Notify::Chatbots concern
    |
    +-- Find chatbots for group
    +-- Filter by event kind (notification_only vs all)
    |
    v
GenericWorker.perform_async('ChatbotService', 'publish_event!', event_id, chatbot_id)
    |
    v
ChatbotService.publish_event!
    |
    +-- Select serializer based on webhook_kind
    +-- Build payload
    +-- POST to webhook URL
    +-- Log errors to Sentry on failure
```

**Serializers:** `/app/serializers/webhook/`
- `slack_serializer.rb`
- `discord_serializer.rb`
- `microsoft_serializer.rb`
- `markdown_serializer.rb`

---

### 8.2 Matrix Integration

**Purpose:** Post Loomio events to Matrix rooms

**Special Handling:**

Matrix uses a different flow than HTTP webhooks:

```ruby
ChatbotService.publish_event!
    |
    +-- IF chatbot.kind == 'matrix':
    |     Publish to Redis: CACHE_REDIS_POOL.rpush('chatbot/publish', payload)
    |     External Matrix service consumes from Redis
    |
    +-- ELSE:
    |     HTTP POST to webhook URL
```

---

### 8.3 Bot API (B2)

**Purpose:** External systems creating content in Loomio

**Authentication:** User API key via `api_key` parameter

**Endpoints:**

| Method | Path | Action |
|--------|------|--------|
| GET | /api/b2/discussions/:id | Fetch discussion |
| POST | /api/b2/discussions | Create discussion |
| GET | /api/b2/polls/:id | Fetch poll |
| POST | /api/b2/polls | Create poll |
| POST | /api/b2/comments | Create comment |
| GET | /api/b2/memberships | List members |
| POST | /api/b2/memberships | Sync memberships |

**Use Case:** Automation scripts, CI/CD integration, external tools creating Loomio content.

---

### 8.4 Admin API (B3)

**Purpose:** Administrative user management for trusted systems

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `B3_API_KEY` | Server-side API key (must be 17+ characters) |

**Endpoints:**

| Method | Path | Action |
|--------|------|--------|
| POST | /api/b3/users/deactivate | Deactivate user account |
| POST | /api/b3/users/reactivate | Reactivate user account |

**Use Case:** Integration with HR systems for automatic account provisioning/deprovisioning.

---

## 9. AI/ML Services

### 9.1 OpenAI

**Purpose:** Transcription and content generation

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |

**Service:** `/app/services/transcription_service.rb`

**Usage:**
- Audio/video transcription
- Content summarization (experimental)

**Library:** `ruby-openai` gem

---

## 10. Analytics

### 10.1 Blazer (Internal)

**Purpose:** SQL-based analytics dashboard for administrators

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `BLAZER_DATABASE_URL` | Read replica connection (optional) |

**Access:** `/admin/blazer` (requires admin authentication)

**Features:**
- Custom SQL queries
- Scheduled query execution
- Dashboard creation
- CSV export

**Library:** `blazer` gem

---

### 10.2 Plausible (External)

**Purpose:** Privacy-focused web analytics

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `PLAUSIBLE_DOMAIN` | Tracked domain |
| `PLAUSIBLE_API_HOST` | Plausible server URL |

**Frontend Library:** `plausible-tracker`

**Features:**
- Page view tracking
- Custom events
- Privacy-compliant (no cookies)

---

## 11. GeoIP

### 11.1 MaxMind

**Purpose:** Geographic location lookup from IP addresses

**Configuration:**

| Environment Variable | Description |
|---------------------|-------------|
| `MAXMIND_DB_PATH` | Path to MaxMind GeoIP database file |

**Worker:** `/app/workers/geo_location_worker.rb`

**Usage:**
- User location for time zone detection
- Regional analytics
- Security monitoring

**Library:** `maxminddb` gem

---

## Configuration Summary

### Required for Basic Operation

| Service | Variables |
|---------|-----------|
| Database | `DATABASE_URL` |
| Redis | `REDIS_URL` |
| Email | `SMTP_SERVER`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD` |

### Recommended for Production

| Service | Variables |
|---------|-----------|
| Storage | `AWS_*` or `GCS_*` |
| Error Tracking | `SENTRY_DSN` |
| Real-time | `CHANNELS_URL`, `HOCUSPOCUS_URL` |

### Optional Integrations

| Service | Variables |
|---------|-----------|
| OAuth | `GOOGLE_APP_KEY/SECRET`, `OAUTH_*`, `SAML_*` |
| Translation | `GOOGLE_TRANSLATE_*` |
| Billing | `CHARGIFY_*` |
| AI | `OPENAI_API_KEY` |
| Analytics | `PLAUSIBLE_*` |
| GeoIP | `MAXMIND_DB_PATH` |

---

*End of External Services Documentation*
