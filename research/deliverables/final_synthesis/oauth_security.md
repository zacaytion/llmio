# OAuth Security - Synthesized Findings

## Summary

This document synthesizes confirmed findings from both third-party discovery and our research, providing implementation-ready details for understanding Loomio's authentication system.

**Critical Security Note**: The existing Loomio OAuth implementation has a confirmed CSRF vulnerability. Any reimplementation MUST NOT reproduce this vulnerability.

---

## Confirmed Findings (Both Sources Agree)

### 1. No OmniAuth - Custom OAuth Implementation

**Confirmed**: Loomio implements OAuth 2.0 Authorization Code flow without using the OmniAuth gem.

**Evidence**:
- `orig/loomio/Gemfile` - No omniauth gems present
- `orig/loomio/app/controllers/identities/` - Custom controller hierarchy
- `orig/loomio/app/extras/clients/` - Custom HTTP clients for token exchange

**Implementation Components**:

| Component | File | Purpose |
|-----------|------|---------|
| Base Controller | `app/controllers/identities/base_controller.rb` | OAuth flow orchestration |
| Google Controller | `app/controllers/identities/google_controller.rb` | Google-specific params |
| Generic OAuth Controller | `app/controllers/identities/oauth_controller.rb` | Configurable OAuth provider |
| Nextcloud Controller | `app/controllers/identities/nextcloud_controller.rb` | Nextcloud-specific |
| SAML Controller | `app/controllers/identities/saml_controller.rb` | SAML 2.0 (separate flow) |
| Google Client | `app/extras/clients/google.rb` | Token exchange & user info |
| Base Client | `app/extras/clients/base.rb` | HTTP request infrastructure |

### 2. Four Identity Providers Supported

**Confirmed** (`orig/loomio/config/providers.yml`):

```yaml
identity:
  - oauth      # Generic OAuth2 (ENV configurable)
  - saml       # SAML 2.0
  - google     # Google OAuth2
  - nextcloud  # Nextcloud OAuth2
```

### 3. Identity Storage Model

**Confirmed** (`orig/loomio/app/models/identity.rb`):

```ruby
class Identity < ApplicationRecord
  self.table_name = :omniauth_identities  # Legacy naming

  validates :identity_type, presence: true
  validates :uid, presence: true
  belongs_to :user, required: false
end
```

**Database Schema** (`orig/loomio/db/schema.rb:619-633`):

| Column | Type | Purpose |
|--------|------|---------|
| `id` | serial | Primary key |
| `user_id` | integer | Foreign key to users (nullable) |
| `identity_type` | string | Provider type (google, oauth, saml, nextcloud) |
| `uid` | string | Unique identifier from OAuth provider |
| `email` | string | Email from OAuth profile |
| `name` | string | Display name from OAuth profile |
| `access_token` | string | OAuth access token (plaintext) |
| `logo` | string | Avatar URL |
| `custom_fields` | jsonb | Provider-specific data |

**Unique Index**: `(identity_type, uid)` - Ensures one identity per provider/user combo.

### 4. OAuth Flow (Current Implementation)

**Confirmed Flow** (from `base_controller.rb`):

```
1. User clicks "Login with Google"
   GET /google/oauth

2. Controller stores back_to URL in session
   session[:back_to] = params[:back_to] || request.referrer

3. Redirect to OAuth provider
   redirect_to "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=...&scope=..."

4. User authenticates with provider

5. Provider redirects back with authorization code
   GET /google/authorize?code=AUTH_CODE

6. Controller exchanges code for access token
   POST https://googleapis.com/oauth2/v4/token
   {code, client_secret, redirect_uri, grant_type: authorization_code}

7. Controller fetches user profile
   GET https://googleapis.com/oauth2/v2/userinfo
   Authorization: Bearer ACCESS_TOKEN

8. Find or create Identity record by uid

9. Link to user (current_user, verified email match, or pending)

10. Sign in and redirect to back_to URL
```

### 5. User Linking Logic

**Confirmed** (`base_controller.rb:29-46`):

```ruby
# SSO-only mode (FEATURES_DISABLE_EMAIL_LOGIN=true)
identity.user = current_user.presence || User.find_by(email: identity.email)
if identity.user.nil?
  identity.user = User.new(name:, email:, email_verified: true)
  identity.user.save!
end

# Standard mode
identity.user = current_user.presence || User.verified.find_by(email: identity.email)
```

**Linking Priority**:
1. Current user (if logged in during OAuth)
2. Existing user with matching verified email
3. New user (SSO-only mode)
4. Pending identity (standard mode, stored in session)

### 6. Environment Variables

**Confirmed OAuth Configuration**:

| Variable | Provider | Purpose |
|----------|----------|---------|
| `GOOGLE_APP_KEY` | Google | OAuth client ID |
| `GOOGLE_APP_SECRET` | Google | OAuth client secret |
| `OAUTH_APP_KEY` | Generic | OAuth client ID |
| `OAUTH_APP_SECRET` | Generic | OAuth client secret |
| `OAUTH_AUTH_URL` | Generic | Authorization endpoint |
| `OAUTH_TOKEN_URL` | Generic | Token endpoint |
| `OAUTH_PROFILE_URL` | Generic | User info endpoint |
| `OAUTH_SCOPE` | Generic | OAuth scopes |
| `OAUTH_ATTR_UID` | Generic | JSON path for user ID |
| `OAUTH_ATTR_NAME` | Generic | JSON path for name |
| `OAUTH_ATTR_EMAIL` | Generic | JSON path for email |
| `NEXTCLOUD_HOST` | Nextcloud | Server URL |
| `NEXTCLOUD_APP_KEY` | Nextcloud | OAuth client ID |
| `NEXTCLOUD_APP_SECRET` | Nextcloud | OAuth client secret |
| `SAML_IDP_METADATA` | SAML | Inline IdP metadata XML |
| `SAML_IDP_METADATA_URL` | SAML | IdP metadata URL |
| `SAML_ISSUER` | SAML | Service provider issuer |
| `FEATURES_DISABLE_EMAIL_LOGIN` | All | Require SSO for all logins |
| `LOOMIO_SSO_FORCE_USER_ATTRS` | All | Sync user attrs on each login |

---

## New Information from Discovery

### 1. Confirmed OAuth CSRF Vulnerability

**CRITICAL**: The OAuth implementation lacks state parameter validation.

**Evidence** (`base_controller.rb:93-95`):

```ruby
def oauth_params
  { client.client_key_name => client.key, redirect_uri: redirect_uri, scope: oauth_scope }
  # NO STATE PARAMETER
end
```

**Attack Vector**:
1. Attacker initiates OAuth flow with their own credentials
2. Captures authorization code from callback URL
3. Crafts malicious URL: `https://loomio.example/google/authorize?code=ATTACKERS_CODE`
4. Victim clicks link while logged into Loomio
5. Victim's account gets linked to attacker's OAuth identity

**CVSS Assessment**: 6.5 (Medium-High)
- Attack Complexity: Low
- Privileges Required: None
- User Interaction: Required
- Impact: Account linking hijack

### 2. SAML Security Configuration

**Finding**: SAML signature verification is disabled.

**Evidence** (`saml_controller.rb:97-100`):

```ruby
settings.security[:authn_requests_signed] = false
settings.security[:logout_requests_signed] = false
settings.security[:logout_responses_signed] = false
settings.security[:metadata_signed] = false
```

**Risk**: Medium - Relies on IdP metadata integrity.

### 3. Rails CSRF Does Not Protect OAuth

**Finding**: Standard Rails CSRF protection exists but doesn't cover OAuth CSRF.

**Evidence** (`protected_from_forgery.rb`):

```ruby
def verified_request?
  super || Rails.env.development? || cookies['csrftoken'] == request.headers['X-CSRF-TOKEN']
end
```

**Why OAuth Is Vulnerable**:
- OAuth initiation is via GET (`get :oauth` in routes)
- OAuth callback is via GET (`get :authorize` in routes)
- GET requests bypass CSRF verification
- No state parameter ties request to session

---

## Security-Critical Patterns to Preserve

### 1. UID as Immutable Identity Key

**Pattern**: Always identify users by provider's `uid`, never by email alone.

**Rationale**: Emails can change. UID is the stable identifier across all OAuth providers.

```
// CORRECT: Find by UID
identity = repo.FindByTypeAndUID("google", "118234567890123456789")

// WRONG: Find by email (email can change)
identity = repo.FindByEmail("user@gmail.com")
```

### 2. Only Link to Verified Users (Standard Mode)

**Pattern**: When not in SSO-only mode, only auto-link to users with verified emails.

**Rationale**: Prevents attackers from claiming accounts by using unverified email addresses.

```
// Standard mode - only verified users
user = userRepo.FindVerifiedByEmail(identity.Email)

// SSO-only mode - any matching email is trusted
user = userRepo.FindByEmail(identity.Email)
```

### 3. Session-Bound State Parameter

**Pattern**: State must be cryptographically random and stored in server-side session.

**Requirements**:
- 32+ bytes of cryptographic randomness
- Stored server-side (session), not in URL parameters
- Single-use (delete after validation)
- Short-lived (session timeout)

### 4. Attribute Sync Flag

**Pattern**: Only overwrite user attributes if explicitly configured.

```ruby
if ENV['LOOMIO_SSO_FORCE_USER_ATTRS'] == "true"
  user.name = identity.name
  user.email = identity.email
  user.save
end
```

### 5. Pending Identity State

**Pattern**: When no user match is found, store identity ID in session for later linking.

```ruby
if identity.user_id.nil?
  session[:pending_identity_id] = identity.id
  # Redirect to registration/linking page
end
```

---

## Migration Notes

### Database Compatibility

The `omniauth_identities` table schema should be preserved for API compatibility:

```sql
CREATE TABLE omniauth_identities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    email VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    identity_type VARCHAR(255),
    uid VARCHAR(255),
    name VARCHAR(255),
    access_token VARCHAR(255) DEFAULT '',
    logo VARCHAR(255),
    custom_fields JSONB DEFAULT '{}' NOT NULL
);

CREATE UNIQUE INDEX index_omniauth_identities_on_identity_type_and_uid
    ON omniauth_identities(identity_type, uid);
```

### Route Compatibility

Maintain existing URL patterns:

| Rails Route | Handler |
|-------------|---------|
| `GET /google/oauth` | OAuthHandler.Initiate |
| `GET /google/authorize` | OAuthHandler.Callback |
| `GET /oauth/oauth` | GenericOAuthHandler.Initiate |
| `GET /oauth/authorize` | GenericOAuthHandler.Callback |
| `POST /saml/oauth` | SAMLHandler.Callback |
| `GET /saml/metadata` | SAMLHandler.Metadata |
