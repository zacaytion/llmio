# Auth Domain - Frontend Documentation

**Generated:** 2026-02-01
**Confidence:** 3/5 (Medium - based on component analysis, some flow details inferred)

---

## Table of Contents

1. [Overview](#overview)
2. [Auth Components](#auth-components)
3. [Auth Service](#auth-service)
4. [Authentication Flow](#authentication-flow)

---

## Overview

The authentication frontend is built with Vue 3 components using Vuetify for UI. Components are located in `/vue/src/components/auth/`. The auth flow is modal-based, with different forms shown based on the user's state (new user, existing user with password, existing user without password, etc.).

---

## Auth Components

### auth/form.vue

**Purpose:** Main auth container that displays appropriate sub-form based on user state.

**Props:**
- `user: Object` - User model object tracking auth state
- `preventClose: Boolean` - Disables dismiss button for required auth

**Data:**
- emailLogin - Whether email login is enabled (from AppConfig)
- siteName - Site name for display
- pendingGroup - Group from pending invitation

**Key Logic:**
- Watches for pending group to display invitation context
- Determines which auth form to show based on pending identity type

**Child Components:**
- `auth-provider-form` - OAuth/SAML provider buttons
- `auth-email-form` - Email input form (if email login enabled)

---

### auth/email_form.vue

**Purpose:** Email input form for initiating email-based authentication.

**Props:**
- `user: Object` - User model object

**Data:**
- email - Local email value for input
- loading - Form submission state

**Methods:**

| Method | Description |
|--------|-------------|
| submit() | Validates email, calls AuthService.emailStatus() |
| validateEmail() | Client-side email format validation |

**API Calls:**
- `AuthService.emailStatus(user)` - checks email status to determine next step

**User Interactions:**
1. User enters email
2. Click "Continue with email" or press Enter
3. System looks up email status
4. Redirects to signin_form or signup_form based on result

---

### auth/provider_form.vue

**Purpose:** Displays OAuth/SAML provider login buttons.

**Props:**
- `user: Object` - User model object

**Computed:**
- providers - List of configured identity providers from AppConfig
- emailLogin - Whether to show "or enter email" text

**Methods:**

| Method | Description |
|--------|-------------|
| select(provider) | Redirects to provider OAuth URL with back_to parameter |
| iconClass(provider) | Returns appropriate icon (mdi-google, mdi-key-variant, etc.) |
| providerColor(provider) | Returns brand color for provider |
| providerName(name) | Returns display name (customizable via theme settings) |

**User Interactions:**
1. User clicks provider button
2. Browser redirects to `/[provider]/oauth?back_to=[current_url]`
3. OAuth flow happens server-side
4. User returns to application authenticated (or with pending identity)

---

### auth/signin_form.vue

**Purpose:** Sign-in form for existing users (password or magic link).

**Props:**
- `user: Object` - User model with email and auth info populated

**Data:**
- vars.name - Name field for updating user
- loading - Form submission state

**Methods:**

| Method | Description |
|--------|-------------|
| signIn() | Submits credentials via AuthService.signIn() |
| signInAndSetPassword() | Signs in then opens password change modal |
| sendLoginLink() | Requests magic link via AuthService.sendLoginLink() |
| submit() | Dispatches to signIn() or sendLoginLink() based on state |

**Conditional Display:**

| Condition | Display |
|-----------|---------|
| user.hasToken | "Sign in as [name]" button (token-based) |
| user.hasPassword | Password field + "Sign in" button |
| !user.hasPassword | "Send login link" button only |

**API Calls:**
- `AuthService.signIn(user)` - POST to sessions
- `AuthService.sendLoginLink(user)` - POST to login_tokens

---

### auth/signup_form.vue

**Purpose:** Registration form for new users.

**Props:**
- `user: Object` - User model with email populated

**Data:**
- vars.name - Name input value
- vars.legalAccepted - Terms checkbox state
- vars.emailNewsletter - Newsletter checkbox state
- loading - Form submission state

**Computed:**
- termsUrl - URL to terms of service (from AppConfig)
- privacyUrl - URL to privacy policy (from AppConfig)
- newsletterEnabled - Whether newsletter signup is offered
- allow - Whether registration is allowed (app feature or pending identity)

**Methods:**

| Method | Description |
|--------|-------------|
| submit() | Validates and calls AuthService.signUp() |

**Validation:**
- Name is required
- Legal acceptance required if terms URL is configured

**API Calls:**
- `AuthService.signUp(user)` - POST to registrations

---

### auth/identity_form.vue

**Purpose:** Form shown when OAuth identity exists but no user linked (account linking flow).

**Props:**
- `user: Object` - User model from pending identity
- `identity: Object` - Pending identity object

**Data:**
- email - Email input for linking to existing account
- loading - Form submission state

**Methods:**

| Method | Description |
|--------|-------------|
| submit() | Sends login link to entered email for account linking |
| createAccount() | Switches to signup form to create new account |

**User Interactions:**
1. OAuth returned with no linked user
2. User can either:
   - Enter existing account email to link accounts
   - Click "Create account" to register new account

---

### auth/complete.vue

**Purpose:** Final step after magic link sent - shows code entry form.

**Props:**
- `user: Object` - User model with sentLoginLink state

**Data:**
- attempts - Number of sign-in attempts (rate limiting display)
- loading - Form submission state

**Methods:**

| Method | Description |
|--------|-------------|
| submit() | Calls AuthService.signIn() with code |
| submitAndSetPassword() | Signs in and opens password change modal |

**Display Logic:**
- Shows "Check your email" message
- 6-digit code input field
- "Set password" link for users who want to add password
- After 3 attempts, shows "too many attempts" message

---

### auth/modal.vue

**Purpose:** Modal container for auth forms.

### auth/inactive.vue

**Purpose:** Displays message for deactivated accounts.

---

## Auth Service

**File:** `/vue/src/shared/services/auth_service.js`

Singleton service handling authentication API calls.

### Methods

#### emailStatus(user)

Checks email status to determine authentication path.

**Logic:**
1. Calls `Records.users.emailStatus(email, pendingToken)`
2. Applies response data to user object
3. Sets: name, avatar info, has_password, email_status, email_verified, auth_form

**Response Handling:**
- If user exists: populates user object with profile data
- Sets `user.hasToken` if pending login token exists
- Sets `user.authForm` to suggested form ('signIn', 'signUp', 'identity')

---

#### signIn(user)

Submits sign-in credentials.

**Logic:**
1. Builds session with email, name, password, code
2. POSTs to `/api/v1/sessions`
3. On success: calls authSuccess()
4. On failure: sets appropriate error message

---

#### signUp(user)

Submits registration.

**Logic:**
1. Builds registration with email, name, legalAccepted, emailNewsletter
2. POSTs to `/api/v1/registrations`
3. If auto-signed-in or has token: calls authSuccess()
4. Otherwise: sets authForm to 'complete', sentLoginLink to true

---

#### authSuccess(data)

Handles successful authentication.

**Logic:**
1. Applies boot data to Session
2. Emits 'closeModal' event
3. Shows success flash message

---

#### sendLoginLink(user)

Requests a magic link.

**Logic:**
1. Calls `Records.loginTokens.fetchToken(email)`
2. Updates user: authForm = 'complete', sentLoginLink = true

---

#### validSignup(vars, user)

Client-side signup validation.

**Validates:**
- Name is present
- Legal accepted (if terms URL configured)

---

#### reactivate(user)

Reactivates a deactivated account.

**Logic:**
1. Calls `Records.users.reactivate(user)`
2. Sets sentLoginLink = true

---

## Authentication Flow

### Flow 1: Email + Password

```
1. User enters email in email_form
2. AuthService.emailStatus() returns has_password: true
3. signin_form shown with password field
4. User enters password, clicks "Sign in"
5. AuthService.signIn() POSTs to /api/v1/sessions
6. On success: Session.apply(), modal closes, user redirected
```

### Flow 2: Magic Link (Existing User)

```
1. User enters email in email_form
2. AuthService.emailStatus() returns has_password: false
3. signin_form shown with "Send login link" button
4. User clicks button
5. AuthService.sendLoginLink() POSTs to /api/v1/login_tokens
6. complete.vue shown with code input
7. User enters 6-digit code from email
8. AuthService.signIn() with code
9. On success: user authenticated
```

### Flow 3: New User Registration

```
1. User enters email in email_form
2. AuthService.emailStatus() returns email_status: "unused"
3. signup_form shown
4. User enters name, accepts terms
5. AuthService.signUp() POSTs to /api/v1/registrations
6. If email can be auto-verified: user signed in immediately
7. Otherwise: complete.vue shown, magic link sent
```

### Flow 4: OAuth/SAML

```
1. User clicks provider button in provider_form
2. Browser redirects to /[provider]/oauth?back_to=[url]
3. User authenticates with provider
4. Server handles callback, creates/links identity
5. If user linked: signed in, redirected to back_to
6. If no user linked: pending_identity_id in session
   - identity_form shown to link or create account
```

### Flow 5: Invitation

```
1. User clicks invitation link
2. Membership token stored in session
3. Auth modal shown
4. User completes any auth flow (email, OAuth, etc.)
5. PendingActionsHelper consumes membership
6. User now member of group
```

---

## AppConfig Settings

Auth-related settings from `AppConfig`:

| Setting | Description |
|---------|-------------|
| features.app.email_login | Enable email authentication |
| features.app.create_user | Allow open registration |
| theme.site_name | Site name for display |
| theme.terms_url | URL to terms of service |
| theme.privacy_url | URL to privacy policy |
| theme.saml_login_provider_name | Custom SAML provider name |
| theme.oauth_login_provider_name | Custom OAuth provider name |
| identityProviders | Array of configured providers |
| pending_identity | Current pending identity data |
| newsletterEnabled | Show newsletter signup |

---

## State Management

Auth state is managed through:

1. **User object** - Passed between components, tracks:
   - email, name, password
   - hasPassword, hasToken
   - authForm (current form to display)
   - sentLoginLink, sentPasswordLink
   - errors (validation errors)

2. **Session service** - Global session state
   - current_user after authentication
   - Applied via Session.apply(bootData)

3. **AppConfig.pending_identity** - Server-provided pending identity data

---

## Open Questions

1. **Session refresh**: How does the frontend handle session expiration?

2. **Error recovery**: What happens if OAuth returns error after several steps?

3. **Multi-tab behavior**: How are multiple auth tabs handled?

---

## Related Files

- `/vue/src/shared/services/session.js` - Session management
- `/vue/src/shared/services/records.js` - Data store
- `/vue/src/shared/interfaces/user_model.js` - User model definition
- `/vue/src/mixins/auth_modal.js` - Auth modal mixin
