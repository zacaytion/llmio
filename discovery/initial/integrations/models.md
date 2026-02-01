# Integrations Domain: Models

**Generated:** 2026-02-01
**Confidence Rating:** 4/5

---

## Overview

The integrations domain handles outbound notifications to external services (webhooks/chatbots), inbound email processing, and OAuth identity management. The primary models are Chatbot, ReceivedEmail, and Identity.

---

## 1. Chatbot Model

**Location:** `/app/models/chatbot.rb`

### Purpose

Represents a configured integration for sending Loomio events to external messaging platforms. Supports both webhook-based integrations (Slack, Discord, Microsoft Teams, Mattermost, Webex) and Matrix protocol.

### Associations

- BELONGS TO group - the group this chatbot is configured for
- BELONGS TO author (User) - who created this chatbot configuration

### Key Attributes

| Column | Type | Description |
|--------|------|-------------|
| kind | string | Integration type: "matrix" or "webhook" |
| webhook_kind | string | Webhook format: "slack", "microsoft", "discord", "markdown", "webex", or nil |
| server | string | Webhook URL or Matrix homeserver URL |
| access_token | string | Authentication token (for Matrix) |
| channel | string | Target channel identifier (for Matrix) |
| event_kinds | string[] | Array of event types to notify about |
| notification_only | boolean | When true, send minimal notification; when false, include full content |
| name | string | User-friendly name for this integration |
| author_id | integer | User who created this integration |
| group_id | integer | Group this integration belongs to |

### Validations

- server is required
- name is required
- kind must be "matrix" or "webhook"
- webhook_kind must be one of the supported formats or nil

### Config Method

Returns a hash with server, access_token, and channel for Matrix integration publishing.

---

## 2. ReceivedEmail Model

**Location:** `/app/models/received_email.rb`

### Purpose

Stores inbound emails for processing. Emails can create discussions, add comments to threads, or be forwarded according to routing rules.

### Associations

- BELONGS TO group - associated group (set during routing)
- HAS MANY ATTACHED attachments - email attachments via Active Storage

### Key Attributes

| Column | Type | Description |
|--------|------|-------------|
| headers | hstore | Email headers as key-value pairs |
| body_text | string | Plain text email body |
| body_html | string | HTML email body |
| spf_valid | boolean | SPF validation result |
| dkim_valid | boolean | DKIM validation result |
| released | boolean | Whether email has been processed |
| group_id | integer | Group determined during routing |

### Scopes

- unreleased - emails not yet processed
- released - emails that have been processed

### Key Methods

**Header Parsing:**
- header(name) - retrieve a specific email header
- sender_email - extract sender email from "from" header
- sender_name - extract sender display name
- recipient_emails - all recipient email addresses
- subject - cleaned subject line (strips Re:, Fwd:, etc.)

**Route Parsing:**
- route_address - find the Loomio-addressable recipient
- route_path - local part of the route address (before @)

**Content Processing:**
- full_body - HTML body if present, otherwise plain text
- reply_body - extracted reply content with quoted text stripped

**Classification:**
- is_addressed_to_loomio? - checks if sent to a recognized Loomio address
- is_auto_response? - detects auto-reply messages by subject patterns
- is_complaint? - detects abuse complaint notifications
- sent_to_notifications_address? - checks if sent to the notifications address

---

## 3. Identity Model

**Location:** `/app/models/identity.rb`

### Purpose

Links external OAuth identity providers to Loomio users. Stores SSO credentials and profile information.

### Table

Uses table name `omniauth_identities` (legacy naming).

### Associations

- BELONGS TO user (optional) - linked Loomio user

### Key Attributes

| Column | Type | Description |
|--------|------|-------------|
| identity_type | string | Provider name (google, oauth, nextcloud, saml) |
| uid | string | Unique identifier from the provider |
| email | string | Email from the provider |
| name | string | Display name from the provider |
| access_token | string | OAuth access token |
| logo | string | Avatar URL from the provider |

### Validations

- identity_type is required
- uid is required

### Key Methods

- force_user_attrs! - updates the linked user's name and email to match the identity
- assign_logo! - downloads and attaches the provider's avatar to the user

### Scopes

- with_user - identities that have a linked user

---

## 4. ForwardEmailRule Model

**Location:** Defined via database table, used in ReceivedEmailService

### Purpose

Stores routing rules for forwarding emails to specific addresses based on handle matching.

### Key Attributes

| Column | Type | Description |
|--------|------|-------------|
| handle | string | Local part to match (before @) |
| email | string | Destination email address for forwarding |

---

## 5. MemberEmailAlias Model

**Location:** Mentioned in routing code, allows alternative email addresses for members

### Purpose

Maps alternative email addresses to group members for inbound email routing. Can be used to allow or block specific email addresses from starting discussions.

### Key Attributes

| Column | Type | Description |
|--------|------|-------------|
| user_id | integer | Linked user (nil means blocked) |
| email | string | The alternative email address |
| group_id | integer | Group this alias applies to |
| author_id | integer | Admin who created this alias |

### Scopes

- allowed - aliases with a user_id (can post)
- blocked - aliases without a user_id (cannot post)

---

## 6. OAuth Application Tables (Doorkeeper)

**Location:** Migration `/db/migrate/20151211015455_create_doorkeeper_tables.rb`

### Purpose

Standard Doorkeeper OAuth tables for external applications to authenticate with Loomio.

### Tables

**oauth_applications:**
- name, uid, secret - application credentials
- redirect_uri - OAuth callback URL
- scopes - allowed scopes

**oauth_access_grants:**
- Temporary authorization codes during OAuth flow

**oauth_access_tokens:**
- Issued access tokens for API access
- Supports refresh tokens

---

## 7. User Token Fields

**Location:** `/app/models/user.rb`

### Purpose

Users have two auto-generated tokens for integration purposes.

### Token Fields

| Field | Purpose |
|-------|---------|
| api_key | Authentication for B2 bot API access |
| email_api_key | Authentication for email-to-thread routing |

These tokens are generated automatically on user creation using the HasTokens concern with secure random generation.

---

## Key Patterns

### Event-Driven Notifications

Chatbots subscribe to specific event kinds. When events occur, the event's trigger chain includes the `Events::Notify::Chatbots` concern which enqueues chatbot notification via GenericWorker.

### Email Routing

Inbound emails use a sophisticated routing system:
1. Parse recipient address to determine route type
2. Authenticate sender via email_api_key or group membership
3. Create appropriate content (discussion or comment)
4. Mark email as released after processing

### Credential Storage

Integration credentials are stored in plain text in the database:
- Chatbot access tokens in the access_token column
- User API keys in api_key and email_api_key columns
- OAuth tokens in the Identity model

Security relies on access control rather than encryption at rest.
