# OAuth Security Analysis

This document provides a comprehensive security analysis of Loomio's custom OAuth implementation, focusing on the pending identity flow, access token lifecycle, and SSO-only mode.

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Pending Identity Flow](#pending-identity-flow)
3. [Access Token Lifecycle](#access-token-lifecycle)
4. [SSO-Only Mode Complete Flow](#sso-only-mode-complete-flow)
5. [CSRF Vulnerability Analysis](#csrf-vulnerability-analysis)
6. [Security Recommendations](#security-recommendations)

---

## Executive Summary

| Finding | Confidence | Severity |
|---------|------------|----------|
| Missing OAuth state parameter (CSRF) | HIGH | HIGH |
| Access tokens stored but unused | HIGH | LOW (data hygiene) |
| Pending identity flow secure | HIGH | N/A |
| SSO-only mode auto-creates users | HIGH | MEDIUM (by design) |

---

## Pending Identity Flow

### Overview

When OAuth authentication succeeds but no matching user is found, the system stores the identity ID in the session and prompts the user to either create a new account or link to an existing one.

### State Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        OAuth Callback (create action)                        │
│                  app/controllers/identities/base_controller.rb:7-58          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────┐
                    │  Fetch access_token from provider   │
                    │  (lines 13-14, 71-73)               │
                    └─────────────────────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────┐
                    │  Fetch identity params (uid, email) │
                    │  (lines 16-17, 76-78)               │
                    └─────────────────────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────┐
                    │  Find existing identity by uid?     │
                    │  (line 20)                          │
                    └─────────────────────────────────────┘
                           │                    │
                     ┌─────┘                    └─────┐
                     ▼ YES                      NO   ▼
        ┌──────────────────────┐      ┌──────────────────────────┐
        │ Update identity      │      │ Create new Identity      │
        │ (line 24)            │      │ (line 27)                │
        │ Sign in user         │      └──────────────────────────┘
        │ (line 51)            │                    │
        └──────────────────────┘                    ▼
                                   ┌─────────────────────────────────────┐
                                   │ SSO-only mode?                       │
                                   │ ENV['FEATURES_DISABLE_EMAIL_LOGIN'] │
                                   │ (line 29)                           │
                                   └─────────────────────────────────────┘
                                           │              │
                                     ┌─────┘              └─────┐
                                YES  ▼                    NO   ▼
                    ┌───────────────────────────┐  ┌───────────────────────────┐
                    │ Link to current_user OR   │  │ Link to current_user OR   │
                    │ ANY user by email OR      │  │ VERIFIED user by email    │
                    │ CREATE new user           │  │ (line 42)                 │
                    │ (lines 33-39)             │  └───────────────────────────┘
                    └───────────────────────────┘              │
                                   │                          │
                                   │                          ▼
                                   │             ┌────────────────────────────┐
                                   │             │ identity.user present?     │
                                   │             │ (line 48)                  │
                                   │             └────────────────────────────┘
                                   │                    │              │
                                   │              ┌─────┘              └─────┐
                                   │         YES  ▼                    NO   ▼
                                   ▼    ┌──────────────────┐  ┌──────────────────────────┐
                    ┌──────────────┐    │ sign_in(user)    │  │ SET pending_identity_id  │
                    │ sign_in      │    │ (line 51)        │  │ session[:pending_        │
                    │ (line 51)    │    └──────────────────┘  │ identity_id] = id        │
                    └──────────────┘                          │ (line 54)                │
                          │                      │            └──────────────────────────┘
                          │                      │                         │
                          └──────────────────────┴─────────────────────────┘
                                                 │
                                                 ▼
                                   ┌─────────────────────────────┐
                                   │ Redirect to back_to or      │
                                   │ dashboard_path (line 57)    │
                                   └─────────────────────────────┘
```

### Where `pending_identity_id` is SET

**File**: `app/controllers/identities/base_controller.rb:54`
```ruby
session[:pending_identity_id] = identity.id
```

**Conditions for setting**:
1. OAuth callback succeeds (valid code, valid profile)
2. New identity created (no existing identity with same uid)
3. `identity.user` is nil after user linking logic

**Also set in SAML controller**: `app/controllers/identities/saml_controller.rb:60`
```ruby
session[:pending_identity_id] = identity.id
```

### Where `pending_identity_id` is CONSUMED

#### 1. On Sign-In (Automatic)

**File**: `app/helpers/current_user_helper.rb:7-11`
```ruby
def sign_in(user)
  @current_user = nil
  user = UserService.verify(user: user)
  super(user) && handle_pending_actions(user) && associate_user_to_visit
end
```

**File**: `app/helpers/pending_actions_helper.rb:24-26`
```ruby
def consume_pending_identity(user)
  pending_identity.update(user: user) if pending_identity
end
```

The `handle_pending_actions` is called in two places:
1. **Custom sign_in override**: `app/helpers/current_user_helper.rb:10`
2. **Before action on all controllers**: `app/controllers/application_controller.rb:17`
3. **Before action on API controllers**: `app/controllers/api/v1/restful_controller.rb:10`

#### 2. Session Cleanup

**File**: `app/helpers/pending_actions_helper.rb:13`
```ruby
session.delete(:pending_identity_id)
```

This happens inside `handle_pending_actions` after consuming.

### Where `pending_identity` is READ (for serialization)

**File**: `app/helpers/pending_actions_helper.rb:96-98`
```ruby
def pending_identity
  Identity.find_by(id: session[:pending_identity_id]) if session[:pending_identity_id]
end
```

**File**: `app/helpers/pending_actions_helper.rb:104-111`
```ruby
def serialized_pending_identity
  Pending::TokenSerializer.new(pending_login_token, root: false).as_json ||
  Pending::IdentitySerializer.new(pending_identity, root: false).as_json ||
  ...
end
```

**Serialized to frontend via**: `app/controllers/api/v1/boot_controller.rb:22`
```ruby
identity: serialized_pending_identity,
```

### Frontend Handling

**File**: `vue/src/shared/services/session.js:26`
```javascript
AppConfig['pendingIdentity'] = data.pending_identity;
```

**File**: `vue/src/components/auth/form.vue:54-56`
```javascript
pendingIdentity() {
  return (AppConfig.pending_identity || {});
},
```

### Security Assessment

**Confidence**: HIGH

The pending identity flow is secure because:
1. Identity ID is stored server-side in session (not client-side)
2. Identity is linked to user only on subsequent sign-in
3. Session is cleared after consumption
4. No sensitive data exposed to frontend (only serialized metadata)

---

## Access Token Lifecycle

### When Stored

**File**: `app/controllers/identities/base_controller.rb:78`
```ruby
def fetch_identity_params(token)
  client = "Clients::#{controller_name.classify}".constantize.new(token: token)
  client.fetch_identity_params.merge({ access_token: token, identity_type: controller_name })
end
```

The `access_token` is stored in the identity record when:
1. OAuth callback creates a new identity (line 27, 45)
2. OAuth callback updates an existing identity (line 24)

### Database Schema

**File**: `db/schema.rb:627`
```ruby
t.string "access_token", default: ""
```

### After Initial OAuth - Token Usage

**Search performed**: `grep -rn "identity.*access_token\|\.access_token" app/`

**Result**: No code reads `identity.access_token` after OAuth callback.

The access token is:
1. Fetched from provider during OAuth callback
2. Stored in the `omniauth_identities.access_token` column
3. **NEVER subsequently read or used**

### Evidence

The only `.access_token` usages found are:
1. `app/models/chatbot.rb:14` - Different context (Matrix chatbot, unrelated to OAuth identity)
2. `app/services/chatbot_service.rb` - Chatbot integration, not OAuth
3. `app/services/transcription_service.rb` - OpenAI API key, not OAuth
4. OAuth callback code itself

### Security Assessment

**Confidence**: HIGH

**Finding**: OAuth access tokens are stored but unused after initial profile fetch.

**Implications**:
1. **Data hygiene concern**: Tokens accumulate in database with no purpose
2. **Potential attack surface**: If database is compromised, tokens could be used against OAuth providers
3. **No refresh token handling**: Tokens likely expire but are never refreshed
4. **Not a functional bug**: Application works correctly without token reuse

**Recommendation**: Consider not storing tokens, or implement token rotation/expiration cleanup.

---

## SSO-Only Mode Complete Flow

### Activation

**Environment variable**: `ENV['FEATURES_DISABLE_EMAIL_LOGIN']`

### Effect on User Linking

When `FEATURES_DISABLE_EMAIL_LOGIN` is set:

**File**: `app/controllers/identities/base_controller.rb:29-43`
```ruby
if ENV['FEATURES_DISABLE_EMAIL_LOGIN']
  # SSO-only mode: uid is the source of truth
  # Try to find existing user by email (for initial account linking)
  # or create a new verified user
  identity.user = current_user.presence || User.find_by(email: identity.email)

  if identity.user.nil?
    # No existing user found - create new verified user for SSO
    identity.user = User.new(identity_params.slice(:name, :email).merge(email_verified: true))
    identity.user.save!
  end
else
  # Standard mode: only link to verified users or current user
  identity.user = current_user.presence || User.verified.find_by(email: identity.email)
end
```

### Complete Flow Diagram (SSO-Only Mode)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           USER INITIATES SSO                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. User visits /google/oauth (or /oauth/oauth, /saml/oauth)                 │
│    File: app/controllers/identities/base_controller.rb:2-5                  │
│    - Stores session[:back_to] = params[:back_to] || request.referrer        │
│    - Redirects to OAuth provider (Google, generic OAuth, etc.)              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 2. User authenticates with OAuth provider (external)                        │
│    Provider redirects back to /google/authorize (or /oauth/authorize)       │
│    with authorization code                                                   │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 3. Callback handler (create action)                                         │
│    File: app/controllers/identities/base_controller.rb:7-58                 │
│    - Exchanges code for access_token (line 13)                              │
│    - Fetches user profile (uid, email, name) (line 16)                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 4. Identity lookup by uid + identity_type                                   │
│    File: app/controllers/identities/base_controller.rb:20                   │
│    identity = Identity.find_by(uid: uid, identity_type: 'google')           │
└─────────────────────────────────────────────────────────────────────────────┘
                           │                        │
                     ┌─────┘                        └─────┐
              FOUND  ▼                               NOT FOUND
    ┌─────────────────────────────┐       ┌─────────────────────────────────────┐
    │ 5a. Update existing identity │       │ 5b. Create new identity             │
    │     (line 24)                │       │     (lines 27-45)                   │
    │     identity.update(params)  │       │     identity = Identity.new(params) │
    └─────────────────────────────┘       └─────────────────────────────────────┘
                     │                                      │
                     │                                      ▼
                     │       ┌─────────────────────────────────────────────────┐
                     │       │ 6. SSO-only user linking (line 29-39)           │
                     │       │    - Try current_user (if logged in)            │
                     │       │    - Try User.find_by(email: identity.email)    │
                     │       │      (links to ANY user, not just verified)     │
                     │       │    - If neither: CREATE new verified user       │
                     │       └─────────────────────────────────────────────────┘
                     │                                      │
                     └──────────────────┬───────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 7. Sign in user                                                              │
│    File: app/controllers/identities/base_controller.rb:50-52                │
│    - identity.force_user_attrs! if ENV['LOOMIO_SSO_FORCE_USER_ATTRS']       │
│    - sign_in(identity.user)                                                  │
│    - flash[:notice] = "Signed in successfully"                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 8. Custom sign_in processing                                                 │
│    File: app/helpers/current_user_helper.rb:7-11                            │
│    - UserService.verify(user) - marks user email_verified: true             │
│    - handle_pending_actions(user) - processes any pending tokens/invites    │
│    - associate_user_to_visit - Ahoy analytics                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 9. Redirect to back_to or dashboard                                          │
│    File: app/controllers/identities/base_controller.rb:57                   │
│    redirect_to session.delete(:back_to) || dashboard_path                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key SSO-Only Behaviors

| Behavior | Standard Mode | SSO-Only Mode |
|----------|---------------|---------------|
| Link to unverified user | NO (creates pending_identity) | YES (links directly) |
| Create user on first SSO | NO (creates pending_identity) | YES (auto-creates verified user) |
| Email verification required | YES | NO (SSO is proof of identity) |
| User can change name/email | YES | Depends on `LOOMIO_SSO_FORCE_USER_ATTRS` |

### Attribute Syncing

When `ENV['LOOMIO_SSO_FORCE_USER_ATTRS']` is set:

**File**: `app/controllers/identities/base_controller.rb:50`
```ruby
identity.force_user_attrs! if ENV['LOOMIO_SSO_FORCE_USER_ATTRS']
```

**File**: `app/models/identity.rb:14-16`
```ruby
def force_user_attrs!
  user.update(name: name, email: email)
end
```

**File**: `app/services/user_service.rb:67-71`
```ruby
if ENV['LOOMIO_SSO_FORCE_USER_ATTRS']
  params.delete(:name)
  params.delete(:email)
  params.delete(:username)
end
```

### Security Assessment

**Confidence**: HIGH

**Finding**: SSO-only mode auto-creates users and bypasses email verification.

**Security implications**:
1. **By design**: This is intentional behavior for enterprise SSO deployments
2. **Trust assumption**: SSO provider is trusted as source of truth for identity
3. **Risk**: If SSO provider is compromised, attackers can create arbitrary users
4. **Mitigation**: Only enable in environments where SSO provider is fully trusted

---

## CSRF Vulnerability Analysis

### Vulnerability Confirmed

**Severity**: HIGH
**Confidence**: HIGH

Loomio's OAuth implementation is missing the OAuth 2.0 `state` parameter for CSRF protection.

### Evidence

#### OAuth Controllers - No State Parameter Generated

**File**: `app/controllers/identities/base_controller.rb:93-95`
```ruby
def oauth_params
  { client.client_key_name => client.key, redirect_uri: redirect_uri, scope: oauth_scope }
end
```

**File**: `app/controllers/identities/google_controller.rb:13-15`
```ruby
def oauth_params
  super.merge(response_type: :code, scope: client.scope.join('+'))
end
```

**File**: `app/controllers/identities/oauth_controller.rb:13-16`
```ruby
def oauth_params
  client = Clients::Oauth.instance
  { client.client_key_name => client.key, redirect_uri: redirect_uri, scope: ENV.fetch('OAUTH_SCOPE'),  response_type: :code }
end
```

**File**: `app/controllers/identities/nextcloud_controller.rb:17-19`
```ruby
def oauth_params
  { client.client_key_name => client.key, redirect_uri: redirect_uri, response_type: :code }
end
```

**Observation**: None of these include a `state` parameter.

#### OAuth Callback - No State Validation

**File**: `app/controllers/identities/base_controller.rb:7-58`

The `create` action processes the callback without validating any state parameter:
```ruby
def create
  if params[:error].present?
    # ... error handling
  end

  access_token = fetch_access_token  # Uses only params[:code]
  # ... no state validation
end
```

### Attack Scenario

1. Attacker initiates OAuth flow on Loomio, gets redirect URL
2. Attacker creates malicious page that auto-submits callback URL with attacker's OAuth code
3. Victim (logged into Loomio) visits attacker's page
4. Victim's browser submits OAuth callback, linking attacker's identity to victim's account
5. Attacker can now log into victim's account via OAuth

### Comparison with Standard OAuth 2.0

**RFC 6749 Section 4.1.1** specifies the `state` parameter:
> RECOMMENDED. An opaque value used by the client to maintain state between the request and callback. The authorization server includes this value when redirecting the user-agent back to the client. The parameter SHOULD be used for preventing cross-site request forgery.

### Mitigating Factors

1. **session[:back_to]**: While not a CSRF token, the session-based back_to provides some protection since attacker cannot control victim's session
2. **Identity linking logic**: Attacker would need a valid OAuth account to exploit
3. **Limited impact in SSO-only mode**: Users cannot unlink SSO identities easily

### SAML Controller - Similar Issue

**File**: `app/controllers/identities/saml_controller.rb:2**
```ruby
skip_before_action :verify_authenticity_token
```

SAML callback skips Rails CSRF protection entirely, relying on SAML response validation. However, SAML has its own relay state mechanism.

### Specific Line References

| File | Line | Issue |
|------|------|-------|
| `app/controllers/identities/base_controller.rb` | 93-95 | No state in oauth_params |
| `app/controllers/identities/google_controller.rb` | 13-15 | No state in oauth_params |
| `app/controllers/identities/oauth_controller.rb` | 13-16 | No state in oauth_params |
| `app/controllers/identities/nextcloud_controller.rb` | 17-19 | No state in oauth_params |
| `app/controllers/identities/base_controller.rb` | 7-58 | No state validation in create |

---

## Security Recommendations

### Critical Priority

1. **Add OAuth state parameter**
   - Generate cryptographically random state in `oauth` action
   - Store in session
   - Validate in `create` action before processing
   - Reject callbacks with mismatched or missing state

### High Priority

2. **Clean up unused access tokens**
   - Consider not storing `access_token` in identity records
   - Or implement token expiration/cleanup job
   - Or use tokens for provider-specific features

### Medium Priority

3. **Document SSO-only mode security assumptions**
   - Enterprise deployments should understand auto-user-creation implications
   - Consider adding admin notification for new SSO user creation

### Low Priority

4. **Add rate limiting to OAuth callbacks**
   - Prevent brute-force attempts on OAuth code exchange
   - (Note: Rack::Attack may already cover this)

---

## Appendix: Key File References

| Component | File Path | Key Lines |
|-----------|-----------|-----------|
| OAuth base controller | `app/controllers/identities/base_controller.rb` | 1-108 |
| Google OAuth | `app/controllers/identities/google_controller.rb` | 1-20 |
| Generic OAuth | `app/controllers/identities/oauth_controller.rb` | 1-17 |
| Nextcloud OAuth | `app/controllers/identities/nextcloud_controller.rb` | 1-24 |
| SAML controller | `app/controllers/identities/saml_controller.rb` | 1-107 |
| Identity model | `app/models/identity.rb` | 1-28 |
| Pending actions helper | `app/helpers/pending_actions_helper.rb` | 1-113 |
| Current user helper | `app/helpers/current_user_helper.rb` | 1-36 |
| User service | `app/services/user_service.rb` | 1-92 |
| OAuth client base | `app/extras/clients/base.rb` | 1-117 |
| Google client | `app/extras/clients/google.rb` | 1-36 |
| OAuth client | `app/extras/clients/oauth.rb` | 1-53 |
| Identity serializer | `app/serializers/pending/identity_serializer.rb` | 1-21 |
| Boot controller | `app/controllers/api/v1/boot_controller.rb` | 1-38 |
| Session service (Vue) | `vue/src/shared/services/session.js` | 1-65 |
| Auth form (Vue) | `vue/src/components/auth/form.vue` | 1-77 |

---

*Document generated: 2026-02-01*
*Analysis confidence: HIGH for all findings*
