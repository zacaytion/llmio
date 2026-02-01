# Auth Domain - Controllers Documentation

**Generated:** 2026-02-01
**Confidence:** 4/5 (High - based on direct source code analysis)

---

## Table of Contents

1. [Sessions API](#sessions-api)
2. [Registrations API](#registrations-api)
3. [Login Tokens API](#login-tokens-api)
4. [Login Tokens (Non-API)](#login-tokens-non-api)
5. [Identity Controllers](#identity-controllers)
6. [Password Recovery](#password-recovery)
7. [Profile API](#profile-api-auth-related)

---

## Sessions API

**File:** `/app/controllers/api/v1/sessions_controller.rb`
**Base:** `Devise::SessionsController`

### POST /api/v1/sessions (create)

Authenticates a user and creates a session.

**Authentication:** None required

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| user[email] | string | Yes* | User's email address |
| user[password] | string | No | User's password |
| user[code] | integer | No | 6-digit login code from email |
| user[name] | string | No | Name to update on successful login |

*Email is required for password auth and code auth; not required for token auth.

**Authentication Strategies:**

1. **Pending Login Token**: If session contains pending_login_token and it's useable, authenticate that token's user
2. **Code Authentication**: If code parameter present, find unused token matching email+code
3. **Password Authentication**: Falls back to Devise warden authentication

**Success Response (200):**

```
{
  "current_user_id": 123,
  "users": [...],
  "memberships": [...],
  "groups": [...],
  "notifications": [...],
  // ... full boot payload
}
```

**Error Response (401):**

```
// Account locked
{ "errors": { "password": ["auth_form.account_locked"] } }

// Invalid password
{ "errors": { "password": ["auth_form.invalid_password"] } }
```

**Side Effects:**
- Creates Devise session cookie
- Sets signed_in cookie (via Warden callback)
- Clears pending_login_token from session
- Broadcasts 'session_create' event
- May update user name if provided

---

### DELETE /api/v1/sessions (destroy)

Signs out the current user.

**Authentication:** Required

**Response (200):**

```
{ "success": "ok" }
```

**Side Effects:**
- Regenerates user's secret_token (invalidates any sessions using old token)
- Destroys Devise session
- Clears signed_in cookie

---

## Registrations API

**File:** `/app/controllers/api/v1/registrations_controller.rb`
**Base:** `Devise::RegistrationsController`

### POST /api/v1/registrations (create)

Creates a new user account or updates an existing unverified account.

**Authentication:** None required

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| user[name] | string | Yes | Display name |
| user[email] | string | Yes | Email address |
| user[legal_accepted] | boolean | Conditional | Required if TERMS_URL is configured |
| user[email_newsletter] | boolean | No | Newsletter subscription preference |

**Permission Check:**

Registration is allowed if:
- `AppConfig.app_features[:create_user]` is true, OR
- User has a pending invitation (membership, discussion_reader, stance), OR
- User has a pending group token

**Email Verification Logic:**

Email can be auto-verified if pending token's user email matches the registration email:
- pending_membership.user.email matches
- pending_login_token.user.email matches
- pending_discussion_reader.user.email matches
- pending_stance.user.email matches
- pending_identity.email matches

**Success Response (200) - Auto-verified:**

```
{
  "success": "ok",
  "signed_in": true,
  "current_user_id": 123,
  // ... full boot payload
}
```

**Success Response (200) - Requires email verification:**

```
{
  "success": "ok",
  "signed_in": false
}
```

**Error Response (422) - Email taken:**

```
{ "errors": { "email": ["Email address is already registered"] } }
```

**Error Response (422) - Invitation required:**

```
{ "errors": {
    "email": ["auth_form.invitation_required"],
    "name": ["auth_form.invitation_required"]
  }
}
```

**Error Response (422) - Validation errors:**

```
{ "errors": { "legal_accepted": [...], "name": [...] } }
```

**Side Effects:**
- Creates or updates User record
- If auto-verified: signs in user, handles pending actions
- If not auto-verified: creates and sends LoginToken

---

## Login Tokens API

**File:** `/app/controllers/api/v1/login_tokens_controller.rb`
**Base:** `Api::V1::RestfulController`

### POST /api/v1/login_tokens (create)

Requests a magic link login email.

**Authentication:** None required

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| email | string | Yes | Email address to send link to |

**Success Response (200):**

```
{ "success": "ok" }
```

**Error Response (404):**

User not found (returns generic 404)

**Error Response (400):**

Email parameter missing

**Side Effects:**
- Creates LoginToken record
- Sends email immediately with magic link and code
- Updates user's detected_locale from Accept-Language header

---

## Login Tokens (Non-API)

**File:** `/app/controllers/login_tokens_controller.rb`
**Base:** `ApplicationController`

### GET /login_tokens/:token (show)

Consumes a magic link token from email.

**Authentication:** None required

**URL Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| token | string | The login token |

**Behavior:**
1. Finds token by token value (raises 404 if not found)
2. Stores token in session as pending_login_token
3. Redirects to token's redirect path, or dashboard if none

**Note:** Token is NOT validated here - validation happens in SessionsController#create when user completes sign-in.

---

## Identity Controllers

### Base Controller

**File:** `/app/controllers/identities/base_controller.rb`

Provides common OAuth flow implementation for all providers.

### GET /:provider/oauth

Initiates OAuth flow by redirecting to provider.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| back_to | string | URL to return to after auth (stored in session) |

**Behavior:**
1. Store back_to (or referrer) in session
2. Build OAuth URL with client_id, redirect_uri, scope
3. Redirect to provider

---

### GET /:provider/authorize (create)

Handles OAuth callback from provider.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| code | string | Authorization code from provider |
| error | string | Error if user cancelled |

**Success Flow:**

1. IF error parameter present, redirect to back_to with flash error
2. Exchange code for access_token
3. Fetch user profile (uid, email, name) from provider
4. Find existing identity by uid+type
5. IF identity exists: update attributes, sign in user
6. IF new identity:
   - **SSO-only mode** (FEATURES_DISABLE_EMAIL_LOGIN):
     - Link to current_user OR find user by email OR create new user
   - **Standard mode**:
     - Link to current_user OR find VERIFIED user by email
     - If no user found, store identity ID in session as pending
7. IF LOOMIO_SSO_FORCE_USER_ATTRS, sync name/email to user
8. Sign in user (or keep logged out with pending identity)
9. Redirect to back_to or dashboard

**Error Responses:**

| Status | Condition |
|--------|-----------|
| 401 | Access token not returned |
| 401 | User profile fetch failed |

---

### GET /:provider (destroy)

Disconnects identity from user account.

**Authentication:** Required

**Behavior:**
1. Find identity for current user by type
2. Destroy identity
3. Redirect to referrer or root

---

### OAuth Controller

**File:** `/app/controllers/identities/oauth_controller.rb`

Generic OAuth2 implementation using environment variables:
- OAUTH_AUTH_URL
- OAUTH_TOKEN_URL
- OAUTH_PROFILE_URL
- OAUTH_SCOPE
- OAUTH_APP_KEY
- OAUTH_APP_SECRET
- OAUTH_ATTR_UID, OAUTH_ATTR_NAME, OAUTH_ATTR_EMAIL

---

### Google Controller

**File:** `/app/controllers/identities/google_controller.rb`

Google OAuth2 with hardcoded authorize URL:
- https://accounts.google.com/o/oauth2/v2/auth

---

### Nextcloud Controller

**File:** `/app/controllers/identities/nextcloud_controller.rb`

Nextcloud OAuth2 using:
- NEXTCLOUD_HOST environment variable
- /index.php/apps/oauth2/authorize path

---

### SAML Controller

**File:** `/app/controllers/identities/saml_controller.rb`

SAML authentication using ruby-saml gem.

#### GET /saml/oauth

Initiates SAML flow.

**Behavior:**
1. Store back_to in session
2. Build SAML auth request
3. Redirect to IdP

---

#### POST /saml/oauth (create)

Handles SAML response.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| SAMLResponse | string | Base64-encoded SAML response |

**Behavior:**
1. Parse SAML response
2. Validate response signature
3. Extract identity: nameid as uid/email, displayName attribute
4. Follow same linking logic as OAuth
5. Sign in or store pending identity
6. Redirect to back_to

**Error Response (500):**

SAML response validation failed

---

#### GET /saml/metadata

Returns SAML service provider metadata.

**Response:** XML with content-type application/samlmetadata+xml

---

## Password Recovery

**File:** `/app/controllers/users/passwords_controller.rb`
**Base:** `Devise::PasswordsController`

Standard Devise password recovery with custom update action that signs in user after successful reset.

### POST /users/password (create)

Sends password reset email (standard Devise).

### PUT /users/password (update)

Resets password using token.

**Behavior:**
1. Reset password via Devise
2. If successful, sign in user
3. Redirect to after_sign_in_path

---

## Profile API (Auth-Related)

**File:** `/app/controllers/api/v1/profile_controller.rb`

### GET /api/v1/profile/email_status

Returns email lookup information for auth form.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| email | string | Email to check |

**Response:**

```
{
  "name": "User Name",
  "email": "user@example.com",
  "has_password": true,
  "has_token": true,
  "email_status": "active",
  "email_verified": true,
  "avatar_kind": "initials",
  // ... other user fields
}
```

Uses Pending::UserSerializer which includes:
- has_password: whether user has set a password
- has_token: whether pending_login_token exists
- auth_form: suggested auth form to show

---

## Routes Summary

```
# API routes
POST   /api/v1/sessions          -> SessionsController#create
DELETE /api/v1/sessions          -> SessionsController#destroy
POST   /api/v1/registrations     -> RegistrationsController#create
POST   /api/v1/login_tokens      -> LoginTokensController#create

# Non-API routes
GET    /login_tokens/:token      -> LoginTokensController#show

# OAuth routes (for each provider: oauth, google, nextcloud)
GET    /:provider/oauth          -> Identities::ProviderController#oauth
GET    /:provider/authorize      -> Identities::ProviderController#create
GET    /:provider/               -> Identities::ProviderController#destroy

# SAML routes
GET    /saml/oauth               -> Identities::SamlController#oauth
POST   /saml/oauth               -> Identities::SamlController#create
GET    /saml/metadata            -> Identities::SamlController#metadata

# Devise routes
POST   /users/password           -> Users::PasswordsController#create
PUT    /users/password           -> Users::PasswordsController#update
```

---

## Open Questions

1. **CSRF protection on SAML**: Controller skips verify_authenticity_token - is this standard SAML practice?

2. **Rate limiting**: No visible rate limiting on login attempts (relies on Devise lockable)

3. **OAuth state parameter**: Not visible in the code - potential CSRF vulnerability?

---

## Related Files

- `/config/routes.rb` - Route definitions
- `/config/initializers/devise.rb` - Devise configuration
- `/config/providers.yml` - Identity provider list
- `/app/clients/` - OAuth client implementations
