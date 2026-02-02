# External Services Specification

**Generated:** 2026-02-01
**Purpose:** Document all external service integrations for Loomio rewrite contract
**Confidence Levels:** HIGH, MEDIUM, LOW per service

---

## Table of Contents

1. [SMTP (Email Delivery)](#1-smtp-email-delivery)
2. [Redis (Queue, Cache, Pub/Sub)](#2-redis-queue-cache-pubsub)
3. [PostgreSQL (Database, Full-Text Search)](#3-postgresql-database-full-text-search)
4. [ActiveStorage Backends](#4-activestorage-backends)
5. [OAuth Providers](#5-oauth-providers)
6. [Webhook Targets](#6-webhook-targets)
7. [Hocuspocus (Collaborative Editing)](#7-hocuspocus-collaborative-editing)
8. [Chargify (Subscription Billing)](#8-chargify-subscription-billing)
9. [Service Dependencies Summary](#9-service-dependencies-summary)

---

## 1. SMTP (Email Delivery)

**Status:** REQUIRED
**Confidence:** HIGH

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SMTP_SERVER` | Yes* | - | SMTP server hostname |
| `SMTP_PORT` | Yes* | - | SMTP server port |
| `SMTP_AUTH` | No | - | Authentication type (plain, login, cram_md5) |
| `SMTP_USERNAME` | No | - | SMTP authentication username |
| `SMTP_PASSWORD` | No | - | SMTP authentication password |
| `SMTP_DOMAIN` | No | - | Domain for HELO command |
| `SMTP_USE_SSL` | No | false | Enable SSL (presence check) |
| `SMTP_SSL_VERIFY_MODE` | No | `none` | SSL verification: none, peer, client_once, fail_if_no_peer_cert |
| `NOTIFICATIONS_EMAIL_ADDRESS` | No | `notifications@{SMTP_DOMAIN}` | From address for outbound email |
| `REPLY_HOSTNAME` | Yes | - | Domain for reply-by-email addresses |
| `OLD_REPLY_HOSTNAME` | No | - | Legacy reply domain (migration support) |
| `COMPLAINTS_ADDRESS` | No | `complaints@email-abuse.amazonses.com` | SES abuse feedback sender |
| `SUPPORT_EMAIL` | No | - | Contact form recipient |

*Without SMTP_SERVER, delivery falls back to `:test` mode (no actual sending).

### Configuration Example

```ruby
# config/application.rb:61-75
config.action_mailer.delivery_method = :smtp
config.action_mailer.smtp_settings = {
  address: ENV['SMTP_SERVER'],
  port: ENV['SMTP_PORT'],
  authentication: ENV['SMTP_AUTH'],
  user_name: ENV['SMTP_USERNAME'],
  password: ENV['SMTP_PASSWORD'],
  domain: ENV['SMTP_DOMAIN'],
  ssl: ENV['SMTP_USE_SSL'].present?,
  openssl_verify_mode: ENV.fetch('SMTP_SSL_VERIFY_MODE', 'none')
}.compact
```

### Failure Handling

| Scenario | Behavior |
|----------|----------|
| SMTP connection failure | Sidekiq retry (25 attempts over ~21 days) |
| Invalid recipient | Bounce email via `ForwardMailer.bounce` |
| Spam complaint (SES) | User `complaints_count` incremented, excluded from future emails |
| Throttle exceeded | Rate limited to 1 bounce notification/hour/sender |

### Inbound Email (Action Mailbox)

| Variable | Required | Description |
|----------|----------|-------------|
| `REPLY_HOSTNAME` | Yes | Reply-to address domain |

**Ingress type:** `:relay` (MTA forwards to Rails endpoint)

**Reply-to address format:**
```
d={discussion_id}&u={user_id}&k={email_api_key}@{REPLY_HOSTNAME}
pt=c&pi={comment_id}&d={discussion_id}&u={user_id}&k={api_key}@{REPLY_HOSTNAME}
```

### Timeouts/Retries

- Default SMTP timeout: Ruby Net::SMTP default (60 seconds)
- Sidekiq retry: 25 attempts with exponential backoff
- No explicit connection pooling at application level

---

## 2. Redis (Queue, Cache, Pub/Sub)

**Status:** REQUIRED
**Confidence:** HIGH

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_URL` | Yes | `redis://localhost:6379/0` | Primary Redis URL (fallback for all) |
| `REDIS_QUEUE_URL` | No | `{REDIS_URL}` | Sidekiq job queue connection |
| `REDIS_CACHE_URL` | No | `{REDIS_URL}` | Rails cache and pub/sub connection |
| `REDIS_POOL_SIZE` | No | `30` | Connection pool size |

### Configuration Example

```ruby
# config/initializers/sidekiq.rb
sidekiq_redis_url = (ENV['REDIS_QUEUE_URL'] || ENV.fetch('REDIS_URL', 'redis://localhost:6379/0'))
channels_redis_url = (ENV['REDIS_CACHE_URL'] || ENV.fetch('REDIS_URL', 'redis://localhost:6379/0'))

CACHE_REDIS_POOL = ConnectionPool.new(size: ENV.fetch('REDIS_POOL_SIZE', 30).to_i, timeout: 5) {
  Redis.new(url: channels_redis_url)
}

# config/application.rb:92
config.cache_store = :redis_cache_store, { url: (ENV['REDIS_CACHE_URL'] || ENV.fetch('REDIS_URL', 'redis://localhost:6379')) }
```

### Redis Usage Patterns

| Purpose | Connection Pool | Channel/Key Pattern |
|---------|----------------|---------------------|
| Sidekiq jobs | Sidekiq-managed | `sidekiq:*` |
| Rails caching | `redis_cache_store` | Rails default keys |
| Pub/sub (records) | `CACHE_REDIS_POOL` | `/records` |
| Pub/sub (system) | `CACHE_REDIS_POOL` | `/system_notice` |
| Throttle counters | `CACHE_REDIS_POOL` | `THROTTLE-{HOUR|DAY}-{key}-{id}` |
| Demo group queue | `CACHE_REDIS_POOL` | `demo_group_ids` (Redis::List) |
| Redis::Objects | `CACHE_REDIS_POOL` | Various counters |

### Pub/Sub Message Format

**Channel: `/records`**
```json
{
  "room": "group-123",
  "records": {
    "events": [...],
    "discussions": [...],
    "comments": [...],
    "users": [...]
  }
}
```

**Channel: `/system_notice`**
```json
{
  "version": "2.15.3",
  "notice": "Maintenance in 15 minutes",
  "reload": false
}
```

### Room Routing

| Room Pattern | Use Case |
|--------------|----------|
| `group-{id}` | Group member broadcasts |
| `user-{id}` | Personal notifications, guest updates |
| `notice` | System-wide announcements |

### Failure Handling

| Scenario | Behavior |
|----------|----------|
| Redis unavailable | Application fails to start |
| Connection timeout | Pool timeout: 5 seconds, then error |
| Pub/sub failure | Message lost (fire-and-forget) |

### Throttle Service Counter TTL

**WARNING:** Throttle counters have NO TTL set on Redis keys. Reset relies on scheduled task:
- `rake loomio:hourly_tasks` calls `ThrottleService.reset!('hour')`
- At midnight: `ThrottleService.reset!('day')`

**Recommendation:** Set EXPIRE on counters for safety if scheduled tasks fail.

---

## 3. PostgreSQL (Database, Full-Text Search)

**Status:** REQUIRED
**Confidence:** HIGH

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes* | - | PostgreSQL connection URL (production) |
| `BLAZER_DATABASE_URL` | No | `{DATABASE_URL}` | Read replica for Blazer analytics |

*Development uses `config/database.yml` without DATABASE_URL.

### Configuration Example

```yaml
# config/database.yml (development)
development:
  adapter: postgresql
  database: loomio_development

# Production uses DATABASE_URL
```

### Full-Text Search (pg_search)

**Configuration:**
```ruby
# config/initializers/pg_search.rb
PgSearch.multisearch_options = {
  using: {
    tsearch: {
      prefix: true,
      negation: true,
      tsvector_column: 'ts_content',
      highlight: {
        StartSel: '<b>',
        StopSel: '</b>'
      }
    }
  }
}
```

### Searchable Models

| Model | Indexed Fields |
|-------|---------------|
| Discussion | title, description |
| Comment | body |
| Poll | title, details |
| Stance | reason |
| Outcome | statement |

### Search Schema

```sql
CREATE TABLE pg_search_documents (
  id BIGINT PRIMARY KEY,
  content TEXT,
  ts_content TSVECTOR,
  searchable_type VARCHAR,
  searchable_id BIGINT,
  group_id BIGINT,
  discussion_id BIGINT,
  poll_id BIGINT,
  author_id BIGINT,
  authored_at TIMESTAMP
);

CREATE INDEX pg_search_documents_searchable_index ON pg_search_documents USING gin(ts_content);
```

### Search Access Control

Queries filter by:
- `group_id IN (:user_group_ids)` for group content
- `discussion_id IN (:guest_discussion_ids)` for guest access
- Anonymous/hidden stances excluded until poll closes

### Reindexing

```ruby
SearchService.reindex_everything          # Full rebuild
SearchService.reindex_by_discussion_id(id) # Single discussion
SearchService.reindex_by_poll_id(id)       # Single poll
SearchService.reindex_by_author_id(id)     # Author's content
```

### Failure Handling

| Scenario | Behavior |
|----------|----------|
| Database unavailable | Application fails to start |
| Search index stale | Manual reindex required |
| Connection pool exhausted | ActiveRecord timeout error |

---

## 4. ActiveStorage Backends

**Status:** REQUIRED (at least one backend)
**Confidence:** HIGH

### Backend Selection Logic

```ruby
# config/application.rb:50-54
if ENV['AWS_BUCKET']
  config.active_storage.service = :amazon
else
  config.active_storage.service = ENV.fetch('ACTIVE_STORAGE_SERVICE', :local)
end
```

**Priority:** `AWS_BUCKET` presence forces `:amazon`, otherwise uses `ACTIVE_STORAGE_SERVICE`.

### Backend: Local (Disk)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ACTIVE_STORAGE_SERVICE` | No | `local` | Set to `local` for disk storage |

**Storage path:** `{Rails.root}/storage/`

### Backend: Amazon S3

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AWS_ACCESS_KEY_ID` | Yes | - | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | Yes | - | AWS secret key |
| `AWS_BUCKET` | Yes | - | S3 bucket name |
| `AWS_REGION` | Yes | - | AWS region (e.g., us-east-1) |

### Backend: DigitalOcean Spaces

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DO_ENDPOINT` | Yes | - | Spaces endpoint URL |
| `DO_ACCESS_KEY_ID` | Yes | - | Spaces access key |
| `DO_SECRET_ACCESS_KEY` | Yes | - | Spaces secret key |
| `DO_BUCKET` | Yes | - | Spaces bucket name |
| `ACTIVE_STORAGE_SERVICE` | Yes | - | Set to `digitalocean` |

### Backend: S3-Compatible (MinIO, Backblaze, Wasabi)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STORAGE_ENDPOINT` | Yes | - | Service endpoint URL |
| `STORAGE_ACCESS_KEY_ID` | Yes | - | Access key |
| `STORAGE_SECRET_ACCESS_KEY` | Yes | - | Secret key |
| `STORAGE_REGION` | Yes | - | Region identifier |
| `STORAGE_BUCKET_NAME` | Yes | - | Bucket name |
| `STORAGE_FORCE_PATH_STYLE` | No | false | Use path-style URLs |
| `ACTIVE_STORAGE_SERVICE` | Yes | - | Set to `s3_compatible` |

### Backend: Google Cloud Storage

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GCS_CREDENTIALS` | Yes | - | JSON credentials (string or path) |
| `GCS_PROJECT` | Yes | - | GCP project ID |
| `GCS_BUCKET` | Yes | - | GCS bucket name |
| `ACTIVE_STORAGE_SERVICE` | Yes | - | Set to `google` |

### Image Processing

| Setting | Value |
|---------|-------|
| Processor | vips |
| Max dimensions | 1280x1280 |
| Quality | 80-85 |
| EXIF stripping | Enabled |

### Rate Limiting

| Endpoint | Limit |
|----------|-------|
| `/rails/active_storage/direct_uploads` | 20/hour/IP |

### File Size Limits

**No application-level file size limits.** Relies on:
- Web server (`client_max_body_size`)
- Cloud provider limits (typically 5GB for single PUT)
- ActiveStorage Direct Upload streams to storage

---

## 5. OAuth Providers

**Status:** OPTIONAL
**Confidence:** HIGH

### Provider: Google OAuth

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_APP_KEY` | Yes | - | Google OAuth client ID |
| `GOOGLE_APP_SECRET` | Yes | - | Google OAuth client secret |

**Endpoints:**
- Authorization: `https://accounts.google.com/o/oauth2/auth`
- Token: `https://www.googleapis.com/oauth2/v4/token`
- Profile: `https://www.googleapis.com/oauth2/v2/userinfo`

**Scopes:** `email`, `profile`

### Provider: Generic OAuth

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OAUTH_APP_KEY` | Yes | - | OAuth client ID |
| `OAUTH_APP_SECRET` | Yes | - | OAuth client secret |
| `OAUTH_AUTHORIZE_URL` | Yes | - | Authorization endpoint URL |
| `OAUTH_TOKEN_URL` | Yes | - | Token exchange endpoint URL |
| `OAUTH_PROFILE_URL` | Yes | - | User profile endpoint URL |
| `OAUTH_SCOPE` | Yes | - | OAuth scopes (space-separated) |
| `OAUTH_ATTR_UID` | Yes | - | JSON path for user ID |
| `OAUTH_ATTR_NAME` | Yes | - | JSON path for user name |
| `OAUTH_ATTR_EMAIL` | Yes | - | JSON path for user email |
| `OAUTH_LOGIN_PROVIDER_NAME` | No | `OAUTH` | Display name in UI |

### Provider: Nextcloud

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXTCLOUD_HOST` | Yes | - | Nextcloud server URL (e.g., https://cloud.example.com) |
| `NEXTCLOUD_APP_KEY` | Yes | - | Nextcloud OAuth client ID |
| `NEXTCLOUD_APP_SECRET` | Yes | - | Nextcloud OAuth client secret |

**Endpoints (relative to NEXTCLOUD_HOST):**
- Authorization: `/index.php/apps/oauth2/authorize`
- Token: `/index.php/apps/oauth2/api/v1/token`
- Profile: `/ocs/v2.php/cloud/user`

### Provider: SAML

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SAML_IDP_METADATA` | Yes* | - | SAML IdP metadata XML (inline) |
| `SAML_IDP_METADATA_URL` | Yes* | - | SAML IdP metadata URL (fetched) |
| `SAML_ISSUER` | No | SP metadata URL | SAML issuer/entity ID |
| `SAML_LOGIN_PROVIDER_NAME` | No | `SAML` | Display name in UI |

*One of `SAML_IDP_METADATA` or `SAML_IDP_METADATA_URL` required.

**SAML Settings:**
```ruby
settings.name_identifier_format = 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress'
settings.security[:authn_requests_signed] = false
settings.security[:logout_requests_signed] = false
settings.security[:metadata_signed] = false
```

### SSO-Only Mode

| Variable | Required | Description |
|----------|----------|-------------|
| `FEATURES_DISABLE_EMAIL_LOGIN` | No | Disable password login, SSO only |
| `LOOMIO_SSO_FORCE_USER_ATTRS` | No | Overwrite user name/email from SSO each login |

**SSO-only behavior:**
- Auto-creates users on first SSO login
- Links to ANY user by email (not just verified)
- Bypasses email verification

### Security Vulnerability

**CRITICAL:** OAuth implementation is missing `state` parameter for CSRF protection.

**Recommendation:** Add state parameter generation, session storage, and validation.

### Failure Handling

| Scenario | Behavior |
|----------|----------|
| OAuth provider unavailable | Redirect with error flash message |
| Invalid OAuth code | Redirect with error flash message |
| SAML response invalid | HTTP 500 with error message |

---

## 6. Webhook Targets

**Status:** OPTIONAL
**Confidence:** HIGH

### Supported Webhook Formats

| Format | Target Platform |
|--------|-----------------|
| `slack` | Slack Incoming Webhooks |
| `microsoft` | Microsoft Teams Connectors |
| `discord` | Discord Webhooks |
| `markdown` | Generic Markdown POST |
| `webex` | Webex Incoming Webhooks |

### Webhook-Eligible Events (14)

From `config/webhook_event_kinds.yml`:
- `new_discussion`
- `discussion_edited`
- `new_comment`
- `poll_created`
- `poll_edited`
- `poll_closing_soon`
- `poll_expired`
- `poll_closed_by_user`
- `poll_reopened`
- `outcome_created`
- `outcome_updated`
- `outcome_review_due`
- `stance_created`
- `stance_updated`

### Delivery Configuration

| Setting | Value |
|---------|-------|
| Delivery | Async via Sidekiq `GenericWorker` |
| Retries | 25 (Sidekiq default) |
| Retry period | ~21 days |
| Error logging | Sentry |

### Payload Signing

**NOT IMPLEMENTED.** No HMAC or signature headers.

**Recommendation:** Add webhook secret per chatbot, sign payload with HMAC-SHA256, include `X-Loomio-Signature` header.

### Circuit Breaker

**NOT IMPLEMENTED.** Failing webhooks continue receiving attempts indefinitely.

**Recommendation:** Track failure count, disable after N consecutive failures.

### Failure Handling

| Scenario | Behavior |
|----------|----------|
| HTTP non-200 response | Sentry log, Sidekiq retry |
| Network timeout | Sidekiq retry |
| Endpoint permanently down | Retries for ~21 days, then dead queue |

---

## 7. Hocuspocus (Collaborative Editing)

**Status:** OPTIONAL
**Confidence:** MEDIUM

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HOCUSPOCUS_URL` | No | `wss://hocuspocus.{CANONICAL_HOST}` (prod) / `ws://localhost:4444` (dev) | Hocuspocus WebSocket URL |

### Architecture

```
Vue.js Client
    |
    +-- @hocuspocus/provider (WebSocket)
    |
Hocuspocus Server (WebSocket, port 5000 internal / 4444 dev)
    |
    +-- POST /api/hocuspocus (auth callback to Rails)
    |
Rails Backend
```

### Authentication Endpoint

**Route:** `POST /api/hocuspocus`

**Authentication Flow:**
1. Client sends `user_secret` (format: `{user_id},{secret_token}`)
2. Client sends `document_name` (format: `{record_type}-{record_id}` or `{record_type}-new-{user_id}`)
3. Rails validates user secret token
4. Rails checks CanCanCan `:update` permission on record
5. Returns HTTP 200 (authorized) or 401 (unauthorized)

### Supported Record Types

```ruby
RECORD_TYPES = %w[comment discussion poll stance outcome pollTemplate discussionTemplate group user]
```

### SSL Exclusion

Hocuspocus endpoint excluded from SSL redirect for internal Docker communication:
```ruby
config.ssl_options = { redirect: { exclude: -> request { request.path =~ /(hocuspocus)/ } } }
```

### Failure Handling

| Scenario | Behavior |
|----------|----------|
| Hocuspocus unavailable | Editor loads from Rails model (fallback) |
| Auth failure | HTTP 401, client disconnects |
| User not found | HTTP 401 |

### Uncertainties

- Exact Hocuspocus server configuration not documented
- Y.js document persistence strategy unclear (Rails is source of truth)
- No explicit connection management documentation

---

## 8. Chargify (Subscription Billing)

**Status:** OPTIONAL
**Confidence:** MEDIUM

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CHARGIFY_API_KEY` | Yes* | - | Chargify API key |
| `CHARGIFY_APP_NAME` | No | - | Chargify subdomain for admin links |

*Presence of `CHARGIFY_API_KEY` enables subscription features.

### Feature Flag

```ruby
# app/extras/app_config.rb:124
subscriptions: !!ENV.fetch('CHARGIFY_API_KEY', false)
```

### Subscription Model

```ruby
class Subscription < ApplicationRecord
  PAYMENT_METHODS = ["chargify", "manual", "barter", "paypal"]
  ACTIVE_STATES = %w[active on_hold pending]

  belongs_to :owner, class_name: 'User'
  has_many :groups
end
```

### Schema Fields

| Column | Type | Description |
|--------|------|-------------|
| `chargify_subscription_id` | integer | Chargify subscription ID |
| `plan` | string | Plan identifier |
| `state` | string | active, on_hold, pending, canceled |
| `payment_method` | string | chargify, manual, barter, paypal |
| `max_members` | integer | Member limit |
| `max_orgs` | integer | Subgroup limit |
| `max_threads` | integer | Thread limit |
| `expires_at` | datetime | Subscription expiration |
| `renews_at` | datetime | Next renewal date |
| `info` | jsonb | Management links, metadata |

### Admin Integration

Admin panel links to Chargify dashboard:
```ruby
"http://#{ENV['CHARGIFY_APP_NAME']}.chargify.com/subscriptions/#{subscription.chargify_subscription_id}"
```

### Uncertainties

- No `SubscriptionService` found in codebase (may be in separate concern)
- Chargify webhook handling not documented
- Subscription lifecycle management unclear

---

## 9. Service Dependencies Summary

### Required Services

| Service | Purpose | Fallback |
|---------|---------|----------|
| PostgreSQL | Primary database | None (required) |
| Redis | Queue, cache, pub/sub | None (required) |

### Required for Production

| Service | Purpose | Fallback |
|---------|---------|----------|
| SMTP | Email delivery | Test mode (no delivery) |
| File Storage | Attachments | Local disk |

### Optional Services

| Service | Purpose | Enable Via |
|---------|---------|------------|
| S3/GCS | Cloud file storage | `AWS_BUCKET` or `ACTIVE_STORAGE_SERVICE` |
| Google OAuth | SSO | `GOOGLE_APP_KEY` |
| Generic OAuth | SSO | `OAUTH_APP_KEY` |
| Nextcloud | SSO | `NEXTCLOUD_HOST` |
| SAML | Enterprise SSO | `SAML_IDP_METADATA*` |
| Hocuspocus | Collaborative editing | `HOCUSPOCUS_URL` |
| Chargify | Subscription billing | `CHARGIFY_API_KEY` |
| Webhooks | Chatbot integrations | Per-group configuration |

### Service Interactions

```
User Request
    |
    +-- Rails API
          |
          +-- PostgreSQL (data persistence)
          +-- Redis (cache, session)
          +-- ActiveStorage (file upload)
          +-- SMTP (email notifications)
          |
          +-- Sidekiq Worker (async)
                |
                +-- Redis (job queue)
                +-- PostgreSQL (data)
                +-- SMTP (email delivery)
                +-- Webhook HTTP clients
                +-- Redis pub/sub (real-time)
                      |
                      +-- Socket.io server (external)
                      +-- Hocuspocus server (external)
```

### Real-Time Service Flow

```
Event Created (Rails)
    |
    +-- PublishEventWorker (Sidekiq)
          |
          +-- Event.trigger!
                |
                +-- LiveUpdate.notify_clients!
                      |
                      +-- MessageChannelService.publish_models
                            |
                            +-- CACHE_REDIS_POOL.publish("/records", {...})
                                  |
                                  +-- Socket.io Server (subscribes to Redis)
                                        |
                                        +-- WebSocket to Vue.js Client
```

---

## Confidence Summary

| Service | Confidence | Notes |
|---------|------------|-------|
| SMTP | HIGH | Well-documented in code, standard Rails |
| Redis | HIGH | Clear configuration, multiple usage patterns |
| PostgreSQL | HIGH | Standard Rails, pg_search well-defined |
| ActiveStorage | HIGH | All backends documented in storage.yml |
| OAuth Providers | HIGH | Custom implementation, security issues noted |
| Webhook Targets | HIGH | Simple HTTP POST, no signing |
| Hocuspocus | MEDIUM | Auth endpoint clear, server config unknown |
| Chargify | MEDIUM | Feature flag present, full integration unclear |

---

## Uncertainties and Gaps

### High Priority

1. **OAuth CSRF vulnerability** - Missing state parameter in all OAuth flows
2. **ThrottleService HTTP response** - Returns 500 instead of 429 for rate limits
3. **Webhook reliability** - No signing, no circuit breaker

### Medium Priority

4. **Hocuspocus server configuration** - Expected setup not documented
5. **Chargify webhook handling** - Subscription lifecycle triggers unknown
6. **Redis TTL for throttle counters** - No automatic expiration

### Low Priority

7. **OAuth token storage** - Tokens stored but never used after initial fetch
8. **File size limits** - No application-level validation

---

*Document generated: 2026-02-01*
*Source: Loomio codebase analysis and discovery reports*
