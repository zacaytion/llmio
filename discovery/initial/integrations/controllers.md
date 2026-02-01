# Integrations Domain: Controllers

**Generated:** 2026-02-01
**Confidence Rating:** 4/5

---

## Overview

Integration controllers handle API endpoints for chatbot management, inbound email processing, OAuth authentication, and external bot API access.

---

## 1. Chatbot Management API

### Api::V1::ChatbotsController

**Location:** `/app/controllers/api/v1/chatbots_controller.rb`
**Base Class:** Api::V1::RestfulController
**Routes:** `/api/v1/chatbots`

#### Actions

**index**
- GET /api/v1/chatbots?group_id=X
- Loads and authorizes group via show_chatbots permission
- Returns all chatbots for the group
- Scope includes current_user_is_admin flag for serialization

**create** (inherited)
- POST /api/v1/chatbots
- Delegates to ChatbotService.create
- Requires group admin role

**update** (inherited)
- PATCH /api/v1/chatbots/:id
- Delegates to ChatbotService.update
- Requires group admin role

**destroy** (inherited)
- DELETE /api/v1/chatbots/:id
- Delegates to ChatbotService.destroy
- Requires group admin role

**test**
- POST /api/v1/chatbots/test
- Sends test message to verify chatbot configuration
- Calls ChatbotService.publish_test! with params
- Returns HTTP 200 OK

#### Serialization

Uses ChatbotSerializer with conditional attributes:
- server and channel only included if current_user_is_admin is true
- This prevents non-admin group members from seeing webhook URLs

---

## 2. Inbound Email Processing

### ReceivedEmailsController

**Location:** `/app/controllers/received_emails_controller.rb`
**Base Class:** ApplicationController
**Routes:** POST /received_emails

#### Security

Skips CSRF verification (webhook endpoint from external email service).

#### Actions

**create**
- Receives JSON payload from email forwarding service (Mailin format)
- Parses mailinMsg JSON parameter containing headers, text, html
- Handles base64-encoded attachments
- Validates email is addressed to Loomio and not auto-response
- Saves email and routes via ReceivedEmailService
- Always returns HTTP 200 OK (prevents retries)

#### Expected Payload Format

```pseudo
{
  mailinMsg: JSON string containing {
    html: "<html>...</html>",
    text: "plain text body",
    headers: {
      from: "Sender Name <sender@example.com>",
      to: "route@reply.loomio.com",
      subject: "Email subject"
    },
    attachments: [
      {
        generatedFileName: "attachment1.pdf",
        contentType: "application/pdf"
      }
    ]
  },
  "attachment1.pdf": "base64-encoded-content"
}
```

---

## 3. Bot API v2

**Location:** `/app/controllers/api/b2/`
**Base Route:** `/api/b2/`
**Authentication:** User API key via params[:api_key]

### Api::B2::BaseController

**Purpose:** Base controller for external bot/integration API.

**Authentication:**
- Skips CSRF verification
- Authenticates via api_key parameter
- Looks up User.active.find_by(api_key: params[:api_key])
- Raises CanCan::AccessDenied if no matching user

**Parameter Handling:**
- Transforms flat params into resource-nested format for PermittedParams

### Api::B2::DiscussionsController

**Routes:**
- GET /api/b2/discussions/:id?api_key=X - show discussion
- POST /api/b2/discussions?api_key=X - create discussion

**Actions:**
- show: Load and authorize discussion, return serialized
- create: Create discussion via DiscussionService

### Api::B2::PollsController

**Routes:**
- GET /api/b2/polls/:id?api_key=X - show poll
- POST /api/b2/polls?api_key=X - create poll

**Actions:**
- show: Load and authorize poll, return serialized
- create: Create poll via PollService, then invite voters

### Api::B2::CommentsController

**Routes:**
- POST /api/b2/comments?api_key=X - create comment

**Actions:**
- create: Create comment via CommentService

### Api::B2::MembershipsController

**Routes:**
- GET /api/b2/memberships?group_id=X&api_key=X - list memberships
- POST /api/b2/memberships?group_id=X&api_key=X - sync memberships

**Actions:**

**index:**
- Returns all memberships for specified group
- Requires API user to be admin of the group

**create (sync):**
- Syncs group membership with provided email list
- Compares params[:emails] with current member emails
- Invites new emails via GroupService.invite
- Optionally removes absent members if remove_absent=1
- Returns added_emails and removed_emails arrays

**Authorization:**
- User must be admin of target group or a site admin

---

## 4. Bot API v3

**Location:** `/app/controllers/api/b3/`
**Base Route:** `/api/b3/`
**Authentication:** Environment variable B3_API_KEY

### Api::B3::UsersController

**Purpose:** Administrative user management API for trusted integrations.

**Authentication:**
- Requires B3_API_KEY environment variable of 16+ characters
- Validates params[:b3_api_key] matches ENV['B3_API_KEY']

**Routes:**
- POST /api/b3/users/deactivate?id=X&b3_api_key=X
- POST /api/b3/users/reactivate?id=X&b3_api_key=X

**Actions:**

**deactivate:**
- Finds active user by ID
- Enqueues DeactivateUserWorker
- Returns { success: :ok }

**reactivate:**
- Finds deactivated user by ID
- Calls UserService.reactivate
- Returns { success: :ok }

---

## 5. OAuth Identity Controllers

**Location:** `/app/controllers/identities/`
**Base Route:** `/[provider]/`

### Identities::BaseController

**Purpose:** Base controller for OAuth authentication flows.

**Actions:**

**oauth (GET /[provider])**
- Stores return URL in session
- Redirects to OAuth provider's authorization URL

**create (GET /[provider]/authorize)**
- Callback from OAuth provider
- Exchanges authorization code for access token
- Fetches user profile from provider
- Finds or creates Identity record
- Links identity to existing user or creates pending identity
- Signs in user if identity has linked user
- Redirects to stored return URL

**destroy (DELETE /[provider])**
- Removes identity link from current user
- Redirects to referrer

**OAuth URL Construction:**
- Uses provider-specific client class for parameters
- Includes client_id, redirect_uri, scope

### Identities::GoogleController

**Routes:**
- GET /google - initiate OAuth
- GET /google/authorize - OAuth callback
- DELETE /google - unlink

**OAuth Host:** accounts.google.com/o/oauth2/v2/auth

### Identities::OauthController

**Purpose:** Generic OAuth controller for configurable SSO.

**Configuration via Environment:**
- OAUTH_AUTH_URL - authorization endpoint
- OAUTH_SCOPE - requested scopes

### Identities::NextcloudController

**Purpose:** Nextcloud OAuth integration.

### Identities::SamlController

**Purpose:** SAML SSO integration.

---

## API Authentication Summary

| API | Authentication Method | Who Can Access |
|-----|----------------------|----------------|
| V1 Chatbots | Session cookie | Group admins |
| B2 Discussions/Polls/Comments | User api_key param | Any user with API key |
| B2 Memberships | User api_key param | Group admins |
| B3 Users | B3_API_KEY env var | Trusted system integrations |
| Inbound Email | None (webhook) | External email service |
| OAuth | OAuth tokens | OAuth providers |

---

## Route Configuration

From `/config/routes.rb`:

```pseudo
namespace :api do
  namespace :b2 do
    resources :discussions, only: [:create, :show]
    resources :polls, only: [:create, :show]
    resources :memberships, only: [:index, :create]
    resources :comments, only: [:create]
  end

  namespace :b3, only: [] do
    resources :users do
      collection do
        post :deactivate
        post :reactivate
      end
    end
  end

  namespace :v1 do
    resources :chatbots, only: [:create, :destroy, :index, :update] do
      post :test, on: :collection
    end
  end
end

post 'received_emails' => 'received_emails#create'
```
