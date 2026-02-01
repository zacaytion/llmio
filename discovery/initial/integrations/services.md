# Integrations Domain: Services

**Generated:** 2026-02-01
**Confidence Rating:** 4/5

---

## Overview

The integrations domain has two primary services: ChatbotService for outbound notifications to external platforms, and ReceivedEmailService for processing inbound emails.

---

## 1. ChatbotService

**Location:** `/app/services/chatbot_service.rb`

### Purpose

Manages chatbot CRUD operations and publishes events to configured chatbots/webhooks.

### Methods

#### create(chatbot:, actor:)

**Purpose:** Create a new chatbot integration for a group.

**Flow:**
1. Authorize actor can create chatbot (must be group admin)
2. Validate chatbot attributes
3. Set actor as author
4. Save chatbot record

**Authorization:** Requires group admin role via Ability::Chatbot

#### update(chatbot:, params:, actor:)

**Purpose:** Update an existing chatbot configuration.

**Flow:**
1. Authorize actor can update chatbot
2. Strip empty access_token from params (preserve existing if not provided)
3. Assign new attributes
4. Validate and save

#### destroy(chatbot:, actor:)

**Purpose:** Delete a chatbot integration.

**Flow:**
1. Authorize actor can destroy
2. Delete the chatbot record

#### publish_event!(event_id)

**Purpose:** Send event notifications to all matching chatbots for a group.

**Flow:**
1. Find and reload the event
2. Return early if eventable is nil
3. Get all chatbots for the eventable's group
4. Filter to chatbots that either:
   - Are explicitly in recipient_chatbot_ids for this event, OR
   - Have the event's kind in their event_kinds array
5. For each matching chatbot:
   a. Determine template based on eventable type (discussion, poll, stance, etc.)
   b. Create a LoggedOutUser with appropriate locale/timezone for rendering
   c. Render notification in the chatbot's locale
   d. For webhook kind:
      - Serialize using format-specific serializer (Slack, Discord, Microsoft, etc.)
      - POST to webhook URL via Clients::Webhook
      - Log errors to Sentry on non-200 responses
   e. For matrix kind:
      - Render HTML template
      - Publish to Redis for external Matrix client service

#### publish_test!(params)

**Purpose:** Send a test message to verify chatbot configuration.

**Flow:**
1. For Slack webhooks:
   - POST a simple text message to the webhook URL
2. For Matrix:
   - Publish test message to Redis for external Matrix client

---

## 2. ReceivedEmailService

**Location:** `/app/services/received_email_service.rb`

### Purpose

Routes inbound emails to appropriate handlers - creating discussions, comments, or forwarding to external addresses.

### Key Methods

#### route_all

**Purpose:** Process all unreleased emails without a group assignment.

**Flow:** Iterate through unreleased emails with nil group_id and call route() on each.

#### route(email)

**Purpose:** Determine how to handle a single inbound email.

**Flow:**
1. Return early if email is already released
2. Destroy email if no sender_email present
3. Handle notifications address replies:
   - If sent to notifications address, send bounce notice (throttled to 1/hour per sender)
   - Destroy the email
4. Handle abuse complaints:
   - Increment complaints_count on complainer's user record
   - Mark email as released
5. Block banned sender hostnames
6. Require valid route_address from here
7. Parse route_path to determine routing type:

**Personal Email-to-Thread Route:**
Pattern: `d=<discussion_id>&u=<user_id>&k=<key>@reply.host`
- Authenticate user via email_api_key
- Create comment on the discussion
- Mark email as released

**Personal Email-to-Group Route:**
Pattern: `<handle>+u=<user_id>&k=<key>@reply.host`
- Authenticate user via email_api_key
- Create new discussion in the group
- Only if thread_from_mail feature enabled

**Forwarding Rule Route:**
Pattern: `<handle>@reply.host` matching ForwardEmailRule
- Forward email to the rule's target email
- Destroy original email

**Group Handle Route:**
Pattern: `<group_handle>@reply.host`
- Check if sender email is blocked for this group
- Associate email with group
- Authenticate sender as group member (or via MemberEmailAlias)
- If authorized: create discussion, mark as released
- If not authorized: publish UnknownSender event for admin notification

8. Destroy if no suitable route found

#### extract_reply_body(text, author_name)

**Purpose:** Strip quoted content and signatures from email replies.

**Flow:**
1. Normalize line endings
2. Repeatedly apply regex patterns to strip:
   - "Original Message" headers
   - Signature delimiters (-- and __)
   - "On ... wrote:" patterns (multiple languages)
   - Author name as signature start
   - Hidden reply delimiter characters
   - Date/sender patterns in various formats

#### delete_old_emails

**Purpose:** Clean up old received emails.

**Flow:** Delete all ReceivedEmail records older than 60 days.

#### refresh_forward_email_rules

**Purpose:** Reload forwarding rules from default configuration file.

**Flow:** Read handles from db/default_forward_email_rules.txt and populate ForwardEmailRule table.

### Private Helper Methods

**parse_route_params(route_path):**
Parses route parameters from email addresses like `handle+u=1&k=abc` into a hash.

**actor_from_email(email):**
Authenticates user from embedded route parameters (u=user_id, k=email_api_key).

**actor_from_email_and_group(email, group):**
Finds user by sender email if they're a group member, or via MemberEmailAlias.

**address_is_blocked(email, group):**
Checks MemberEmailAlias.blocked for sender email.

**discussion_params(email):**
Builds discussion attributes from email (title from subject, body from content, attachments).

**comment_params(email):**
Builds comment attributes from email (body from reply_body, parent from route params).

---

## 3. OAuth Client Services

**Location:** `/app/extras/clients/`

### Purpose

API clients for OAuth authentication and external service communication.

### Clients::Base

**Base class providing:**
- HTTP GET/POST methods via HTTParty
- Request/response handling with success/failure callbacks
- Default parameter and header management
- Token authentication

### Clients::Webhook

**Purpose:** POST webhook payloads to external services.

**Key Methods:**
- post(url, params:) - POST JSON payload to webhook URL
- serialized_event(event, format, webhook) - Select and apply format-specific serializer

**Serializer Selection:** Tries in order:
1. Event-kind-specific serializer (e.g., Webhook::Slack::NewCommentSerializer)
2. Eventable-type serializer (e.g., Webhook::Slack::PollSerializer)
3. Base serializer (e.g., Webhook::Slack::BaseSerializer)

### Clients::Oauth

**Purpose:** Generic OAuth 2.0 client for configurable SSO.

**Configuration via Environment:**
- OAUTH_TOKEN_URL - token exchange endpoint
- OAUTH_PROFILE_URL - user profile endpoint
- OAUTH_ATTR_UID/NAME/EMAIL - JSON paths for user attributes

**Key Methods:**
- fetch_access_token(code, uri) - exchange authorization code for access token
- fetch_identity_params() - retrieve user profile from provider

### Clients::Google

**Purpose:** Google OAuth 2.0 client.

**Endpoints:**
- Token: googleapis.com/oauth2/v4/token
- Profile: googleapis.com/oauth2/v2/userinfo

**Scopes:** email, profile

---

## Event Flow for Chatbot Notifications

The event trigger chain for chatbot notifications:

1. Action occurs (e.g., new discussion created)
2. Service publishes event via Event.publish!
3. PublishEventWorker enqueues asynchronously
4. Worker calls event.trigger!
5. Events::Notify::Chatbots concern's trigger! method executes
6. GenericWorker enqueued with ChatbotService.publish_event!(event_id)
7. ChatbotService finds matching chatbots and sends notifications

**Concern Location:** `/app/models/concerns/events/notify/chatbots.rb`

```pseudo
module Events::Notify::Chatbots
  def trigger!
    call parent trigger! method
    enqueue GenericWorker for ChatbotService.publish_event! with event id
  end
end
```

---

## Email Processing Flow

```pseudo
1. External email service POSTs to /received_emails
2. ReceivedEmailsController builds ReceivedEmail from JSON params
3. If addressed to Loomio and not auto-response:
   a. Save email record
   b. Call ReceivedEmailService.route(email)
4. Route method parses address and authenticates sender
5. Creates appropriate content (discussion or comment)
6. Marks email as released
7. Cron job cleans up old emails after 60 days
```
