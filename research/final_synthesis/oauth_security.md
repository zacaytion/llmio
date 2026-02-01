# OAuth Security - Synthesized Findings

## Summary

This document synthesizes confirmed findings from both third-party discovery and our research, providing implementation-ready details for the Go rewrite of Loomio's authentication system.

**Critical Security Note**: The existing Loomio OAuth implementation has a confirmed CSRF vulnerability. The Go implementation MUST NOT reproduce this vulnerability.

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

## Implementation-Ready Details for Go Rewrite

### 1. Secure OAuth Flow (MUST Implement)

```go
// OAuth initiation handler
func (h *OAuthHandler) Initiate(w http.ResponseWriter, r *http.Request) {
    // Generate cryptographically secure state
    state := generateSecureRandom(32)

    // Store state in session
    session := h.sessionStore.Get(r)
    session.Set("oauth_state", state)
    session.Set("oauth_back_to", r.URL.Query().Get("back_to"))
    session.Save(w)

    // Build authorization URL with state
    authURL := fmt.Sprintf("%s?%s",
        h.config.AuthURL,
        url.Values{
            "client_id":     {h.config.ClientID},
            "redirect_uri":  {h.config.RedirectURI},
            "scope":         {h.config.Scope},
            "response_type": {"code"},
            "state":         {state},  // CRITICAL: Include state
        }.Encode(),
    )

    http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuth callback handler
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
    // CRITICAL: Validate state FIRST
    session := h.sessionStore.Get(r)
    expectedState := session.Get("oauth_state")
    actualState := r.URL.Query().Get("state")

    if expectedState == "" || actualState != expectedState {
        http.Error(w, "Invalid OAuth state - possible CSRF attack", http.StatusForbidden)
        return
    }

    // Clear state from session (one-time use)
    session.Delete("oauth_state")
    session.Save(w)

    // Proceed with token exchange...
    code := r.URL.Query().Get("code")
    accessToken, err := h.exchangeCodeForToken(code)
    // ...
}
```

### 2. Identity Model for Go

```go
type Identity struct {
    ID           int64     `db:"id"`
    UserID       *int64    `db:"user_id"`       // Nullable
    IdentityType string    `db:"identity_type"` // google, oauth, saml, nextcloud
    UID          string    `db:"uid"`           // Provider's unique ID
    Email        string    `db:"email"`
    Name         string    `db:"name"`
    AccessToken  string    `db:"access_token"`  // Consider encrypting
    Logo         string    `db:"logo"`
    CustomFields JSONB     `db:"custom_fields"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

// FindOrCreateByUID finds existing identity or creates new one
func (r *IdentityRepository) FindOrCreateByUID(ctx context.Context, params IdentityParams) (*Identity, error) {
    // First try to find existing
    identity, err := r.FindByTypeAndUID(ctx, params.IdentityType, params.UID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return nil, err
    }

    if identity != nil {
        // Update mutable fields (email, name may change in SSO)
        identity.Email = params.Email
        identity.Name = params.Name
        identity.AccessToken = params.AccessToken
        identity.Logo = params.Logo
        return r.Update(ctx, identity)
    }

    // Create new identity
    return r.Create(ctx, params)
}
```

### 3. User Linking Logic for Go

```go
func (s *OAuthService) LinkIdentityToUser(ctx context.Context, identity *Identity, currentUser *User) error {
    // Priority 1: Current user (if logged in during OAuth)
    if currentUser != nil {
        identity.UserID = &currentUser.ID
        return s.identityRepo.Update(ctx, identity)
    }

    // Priority 2: Find by verified email
    if s.config.SSOOnlyMode {
        // SSO-only: any matching email
        user, err := s.userRepo.FindByEmail(ctx, identity.Email)
        if err != nil && !errors.Is(err, sql.ErrNoRows) {
            return err
        }
        if user != nil {
            identity.UserID = &user.ID
            return s.identityRepo.Update(ctx, identity)
        }

        // Create new verified user
        newUser := &User{
            Name:          identity.Name,
            Email:         identity.Email,
            EmailVerified: true,
        }
        if err := s.userRepo.Create(ctx, newUser); err != nil {
            return err
        }
        identity.UserID = &newUser.ID
        return s.identityRepo.Update(ctx, identity)
    }

    // Standard mode: only verified users
    user, err := s.userRepo.FindVerifiedByEmail(ctx, identity.Email)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return err
    }
    if user != nil {
        identity.UserID = &user.ID
        return s.identityRepo.Update(ctx, identity)
    }

    // No match - identity remains unlinked (pending state)
    return nil
}
```

### 4. sqlc Queries

```sql
-- name: FindIdentityByTypeAndUID :one
SELECT * FROM omniauth_identities
WHERE identity_type = $1 AND uid = $2;

-- name: CreateIdentity :one
INSERT INTO omniauth_identities (
    user_id, identity_type, uid, email, name, access_token, logo, custom_fields
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateIdentity :one
UPDATE omniauth_identities SET
    email = $2,
    name = $3,
    access_token = $4,
    logo = $5,
    custom_fields = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: FindUserIdentities :many
SELECT * FROM omniauth_identities
WHERE user_id = $1
ORDER BY created_at;
```

### 5. Recommended Go OAuth Library

Use `golang.org/x/oauth2` for standards-compliant OAuth 2.0:

```go
import "golang.org/x/oauth2"
import "golang.org/x/oauth2/google"

func NewGoogleOAuthConfig() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_APP_KEY"),
        ClientSecret: os.Getenv("GOOGLE_APP_SECRET"),
        RedirectURL:  "https://example.com/google/authorize",
        Scopes:       []string{"email", "profile"},
        Endpoint:     google.Endpoint,
    }
}

// Generate auth URL with state
func (c *oauth2.Config) AuthCodeURL(state string) string
// Exchange code for token
func (c *oauth2.Config) Exchange(ctx, code) (*oauth2.Token, error)
```

**Note**: `golang.org/x/oauth2` handles state parameter generation and validation when used correctly.

---

## Security-Critical Patterns to Preserve

### 1. UID as Immutable Identity Key

**Pattern**: Always identify users by provider's `uid`, never by email alone.

**Rationale**: Emails can change. UID is the stable identifier across all OAuth providers.

```go
// CORRECT: Find by UID
identity, _ := repo.FindByTypeAndUID("google", "118234567890123456789")

// WRONG: Find by email (email can change)
identity, _ := repo.FindByEmail("user@gmail.com")
```

### 2. Only Link to Verified Users (Standard Mode)

**Pattern**: When not in SSO-only mode, only auto-link to users with verified emails.

**Rationale**: Prevents attackers from claiming accounts by using unverified email addresses.

```go
// Standard mode - only verified users
user, _ := userRepo.FindVerifiedByEmail(ctx, identity.Email)

// SSO-only mode - any matching email is trusted
user, _ := userRepo.FindByEmail(ctx, identity.Email)
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

```go
if os.Getenv("LOOMIO_SSO_FORCE_USER_ATTRS") == "true" {
    user.Name = identity.Name
    user.Email = identity.Email
    userRepo.Update(ctx, user)
}
```

### 5. Pending Identity State

**Pattern**: When no user match is found, store identity ID in session for later linking.

```go
if identity.UserID == nil {
    session.Set("pending_identity_id", identity.ID)
    // Redirect to registration/linking page
}
```

---

## Test Cases for Go Implementation

### Security Tests

```go
func TestOAuthCSRFProtection(t *testing.T) {
    // Test 1: Callback without state should fail
    req := httptest.NewRequest("GET", "/google/authorize?code=valid_code", nil)
    // No state in request or session
    resp := executeCallback(req)
    assert.Equal(t, http.StatusForbidden, resp.StatusCode)

    // Test 2: Callback with mismatched state should fail
    session.Set("oauth_state", "expected_state")
    req = httptest.NewRequest("GET", "/google/authorize?code=valid_code&state=wrong_state", nil)
    resp = executeCallback(req)
    assert.Equal(t, http.StatusForbidden, resp.StatusCode)

    // Test 3: Callback with correct state should succeed
    state := "random_state_value"
    session.Set("oauth_state", state)
    req = httptest.NewRequest("GET", "/google/authorize?code=valid_code&state="+state, nil)
    resp = executeCallback(req)
    assert.Equal(t, http.StatusFound, resp.StatusCode)

    // Test 4: State should be single-use
    req = httptest.NewRequest("GET", "/google/authorize?code=another_code&state="+state, nil)
    resp = executeCallback(req)
    assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
```

### Identity Linking Tests

```go
func TestIdentityLinking(t *testing.T) {
    // Test 1: Logged-in user - link to current user
    // Test 2: Verified email match - link to existing user
    // Test 3: Unverified email match - do NOT link (standard mode)
    // Test 4: No match - create pending identity
    // Test 5: SSO-only mode - create new user
    // Test 6: Existing identity - update attributes
}
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

| Rails Route | Go Route | Handler |
|-------------|----------|---------|
| `GET /google/oauth` | `GET /google/oauth` | `OAuthHandler.Initiate` |
| `GET /google/authorize` | `GET /google/authorize` | `OAuthHandler.Callback` |
| `GET /oauth/oauth` | `GET /oauth/oauth` | `GenericOAuthHandler.Initiate` |
| `GET /oauth/authorize` | `GET /oauth/authorize` | `GenericOAuthHandler.Callback` |
| `POST /saml/oauth` | `POST /saml/oauth` | `SAMLHandler.Callback` |
| `GET /saml/metadata` | `GET /saml/metadata` | `SAMLHandler.Metadata` |
