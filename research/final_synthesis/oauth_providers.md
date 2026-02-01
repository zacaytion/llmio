# OAuth Providers: Implementation Synthesis

## Executive Summary

Loomio implements **exactly 4 identity providers** for SSO/OAuth authentication:

| Provider | Protocol | Status |
|----------|----------|--------|
| Google | OAuth 2.0 | Active |
| Generic OAuth | OAuth 2.0 | Active (configurable) |
| SAML | SAML 2.0 | Active |
| Nextcloud | OAuth 2.0 | Active |

Facebook, Slack, and Microsoft OAuth **do not exist** and never did. Slack/Microsoft webhook serializers are for outbound notifications, not authentication.

---

## Confirmed Architecture

### Provider Configuration

**Source of truth**: `config/providers.yml`

```yaml
identity:
  - oauth
  - saml
  - google
  - nextcloud
```

This YAML is loaded into `Identity::PROVIDERS` constant and used for:
1. Dynamic route generation
2. Controller resolution
3. Frontend provider list

### Enable/Disable Mechanism

Providers are conditionally exposed to the frontend based on environment variables:

```ruby
# app/models/boot/site.rb
identityProviders: AppConfig.providers.fetch('identity', []).map do |provider|
  ({ name: provider, href: send("#{provider}_oauth_path") } if ENV["#{provider.upcase}_APP_KEY"])
end.compact
```

**Required environment variables per provider**:

| Provider | Required Variables |
|----------|-------------------|
| `google` | `GOOGLE_APP_KEY`, `GOOGLE_APP_SECRET` |
| `oauth` | `OAUTH_APP_KEY`, `OAUTH_APP_SECRET`, `OAUTH_AUTH_URL`, `OAUTH_TOKEN_URL`, `OAUTH_PROFILE_URL`, `OAUTH_SCOPE`, `OAUTH_ATTR_UID`, `OAUTH_ATTR_NAME`, `OAUTH_ATTR_EMAIL` |
| `nextcloud` | `NEXTCLOUD_APP_KEY`, `NEXTCLOUD_APP_SECRET`, `NEXTCLOUD_HOST` |
| `saml` | `SAML_APP_KEY`, `SAML_IDP_METADATA` or `SAML_IDP_METADATA_URL` |

---

## Database Schema

### Table: `omniauth_identities`

```sql
CREATE TABLE omniauth_identities (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  email VARCHAR(255),
  identity_type VARCHAR(255) NOT NULL,  -- 'google', 'oauth', 'saml', 'nextcloud'
  uid VARCHAR(255) NOT NULL,            -- Provider's unique user ID
  name VARCHAR(255),
  access_token VARCHAR DEFAULT '',
  logo VARCHAR,                         -- Avatar URL from provider
  custom_fields JSONB DEFAULT '{}' NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_identities_type_uid ON omniauth_identities(identity_type, uid);
CREATE INDEX idx_identities_user_id ON omniauth_identities(user_id);
CREATE INDEX idx_identities_email ON omniauth_identities(email);
```

### Go Schema (sqlc)

```sql
-- name: GetIdentityByProviderUID :one
SELECT * FROM omniauth_identities
WHERE identity_type = $1 AND uid = $2 LIMIT 1;

-- name: CreateIdentity :one
INSERT INTO omniauth_identities (
  user_id, email, identity_type, uid, name, access_token, logo, custom_fields
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: LinkIdentityToUser :exec
UPDATE omniauth_identities SET user_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUserIdentity :exec
DELETE FROM omniauth_identities
WHERE user_id = $1 AND identity_type = $2;
```

---

## Controller Architecture

### Rails Pattern

```
Identities::BaseController < ApplicationController
  +-- Identities::GoogleController
  +-- Identities::NextcloudController
  +-- Identities::OauthController

Identities::SamlController < ApplicationController  (standalone)
```

### Routes (Dynamic Generation)

```ruby
Identity::PROVIDERS.each do |provider|
  scope provider do
    get :oauth,     to: "identities/#{provider}#oauth"     # Initiate flow
    get :authorize, to: "identities/#{provider}#create"    # Callback
    get '/',        to: "identities/#{provider}#destroy"   # Unlink
  end
end

# SAML has special POST callback
scope :saml do
  post :oauth,    to: 'identities/saml#create'   # SAML response
  get :metadata,  to: 'identities/saml#metadata' # SP metadata
end
```

### Go Routes (chi)

```go
r := chi.NewRouter()

// Dynamic OAuth providers
for _, provider := range []string{"google", "oauth", "nextcloud"} {
    r.Route("/"+provider, func(r chi.Router) {
        r.Get("/oauth", handlers.OAuthInitiate(provider))
        r.Get("/authorize", handlers.OAuthCallback(provider))
        r.Get("/", handlers.OAuthUnlink(provider))
    })
}

// SAML routes (special handling)
r.Route("/saml", func(r chi.Router) {
    r.Get("/oauth", handlers.SAMLInitiate)
    r.Post("/oauth", handlers.SAMLCallback)
    r.Get("/metadata", handlers.SAMLMetadata)
    r.Get("/", handlers.SAMLUnlink)
})
```

---

## Authentication Flow

### OAuth 2.0 Flow (Google, Generic, Nextcloud)

```
1. User clicks "Sign in with [Provider]"
   GET /{provider}/oauth?back_to=/dashboard

2. Controller saves return URL in session, redirects to provider
   session[:back_to] = params[:back_to]
   redirect_to provider_authorization_url

3. User authenticates with provider, provider redirects back
   GET /{provider}/authorize?code=AUTH_CODE

4. Controller exchanges code for token
   token = client.fetch_access_token(code, redirect_uri)

5. Controller fetches user profile
   identity_params = client.fetch_identity_params

6. Find or create Identity by (identity_type, uid)
   identity = Identity.find_or_initialize_by(
     identity_type: provider,
     uid: identity_params[:uid]
   )

7. Link to user or set pending
   if identity.user_id.present?
     sign_in(identity.user)
   elsif user = User.find_by(email: identity_params[:email])
     # Optionally auto-link or require verification
   else
     session[:pending_identity_id] = identity.id
     # Redirect to registration/link page
   end

8. Redirect to original destination
   redirect_to session.delete(:back_to) || root_path
```

### SAML Flow

```
1. User clicks "Sign in with SAML"
   GET /saml/oauth

2. Controller builds AuthnRequest, redirects to IdP
   auth_request = OneLogin::RubySaml::Authrequest.new
   redirect_to auth_request.create(saml_settings)

3. User authenticates with IdP, IdP POSTs back
   POST /saml/oauth
   params[:SAMLResponse] = BASE64_ENCODED_RESPONSE

4. Controller validates SAML response
   response = OneLogin::RubySaml::Response.new(params[:SAMLResponse])
   return error unless response.is_valid?

5. Extract identity params from SAML attributes
   identity_params = {
     identity_type: 'saml',
     uid: response.nameid,
     email: response.nameid,
     name: response.attributes['displayName']
   }

6. Continue with same linking logic as OAuth
```

---

## Go Implementation

### Provider Interface

```go
package oauth

import "context"

type IdentityParams struct {
    UID          string
    Name         string
    Email        string
    Logo         string
    AccessToken  string
    IdentityType string
}

type Provider interface {
    // AuthorizationURL returns the OAuth authorization URL
    AuthorizationURL(redirectURI, state string) string

    // ExchangeCode exchanges auth code for access token
    ExchangeCode(ctx context.Context, code, redirectURI string) (string, error)

    // FetchIdentity fetches user identity using access token
    FetchIdentity(ctx context.Context, token string) (*IdentityParams, error)

    // Name returns the provider identifier
    Name() string
}
```

### Google Provider

```go
package oauth

import (
    "context"
    "encoding/json"
    "net/http"
    "net/url"
    "os"
)

type GoogleProvider struct {
    clientID     string
    clientSecret string
    httpClient   *http.Client
}

func NewGoogleProvider() *GoogleProvider {
    return &GoogleProvider{
        clientID:     os.Getenv("GOOGLE_APP_KEY"),
        clientSecret: os.Getenv("GOOGLE_APP_SECRET"),
        httpClient:   &http.Client{},
    }
}

func (p *GoogleProvider) Name() string {
    return "google"
}

func (p *GoogleProvider) AuthorizationURL(redirectURI, state string) string {
    params := url.Values{
        "client_id":     {p.clientID},
        "redirect_uri":  {redirectURI},
        "response_type": {"code"},
        "scope":         {"email profile"},
        "state":         {state},
    }
    return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (string, error) {
    params := url.Values{
        "client_id":     {p.clientID},
        "client_secret": {p.clientSecret},
        "code":          {code},
        "redirect_uri":  {redirectURI},
        "grant_type":    {"authorization_code"},
    }

    resp, err := p.httpClient.PostForm("https://www.googleapis.com/oauth2/v4/token", params)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        AccessToken string `json:"access_token"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }
    return result.AccessToken, nil
}

func (p *GoogleProvider) FetchIdentity(ctx context.Context, token string) (*IdentityParams, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET",
        "https://www.googleapis.com/oauth2/v2/userinfo", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var profile struct {
        ID      string `json:"id"`
        Name    string `json:"name"`
        Email   string `json:"email"`
        Picture string `json:"picture"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
        return nil, err
    }

    return &IdentityParams{
        UID:          profile.ID,
        Name:         profile.Name,
        Email:        profile.Email,
        Logo:         profile.Picture,
        AccessToken:  token,
        IdentityType: "google",
    }, nil
}
```

### Generic OAuth Provider

```go
package oauth

import (
    "context"
    "encoding/json"
    "net/http"
    "net/url"
    "os"
    "strings"
)

type GenericOAuthProvider struct {
    clientID     string
    clientSecret string
    authURL      string
    tokenURL     string
    profileURL   string
    scope        string
    attrUID      string
    attrName     string
    attrEmail    string
    httpClient   *http.Client
}

func NewGenericOAuthProvider() *GenericOAuthProvider {
    return &GenericOAuthProvider{
        clientID:     os.Getenv("OAUTH_APP_KEY"),
        clientSecret: os.Getenv("OAUTH_APP_SECRET"),
        authURL:      os.Getenv("OAUTH_AUTH_URL"),
        tokenURL:     os.Getenv("OAUTH_TOKEN_URL"),
        profileURL:   os.Getenv("OAUTH_PROFILE_URL"),
        scope:        os.Getenv("OAUTH_SCOPE"),
        attrUID:      os.Getenv("OAUTH_ATTR_UID"),
        attrName:     os.Getenv("OAUTH_ATTR_NAME"),
        attrEmail:    os.Getenv("OAUTH_ATTR_EMAIL"),
        httpClient:   &http.Client{},
    }
}

func (p *GenericOAuthProvider) Name() string {
    return "oauth"
}

func (p *GenericOAuthProvider) FetchIdentity(ctx context.Context, token string) (*IdentityParams, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", p.profileURL, nil)
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var profile map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
        return nil, err
    }

    return &IdentityParams{
        UID:          extractPath(profile, p.attrUID),
        Name:         extractPath(profile, p.attrName),
        Email:        extractPath(profile, p.attrEmail),
        AccessToken:  token,
        IdentityType: "oauth",
    }, nil
}

// extractPath navigates nested maps using dot notation
func extractPath(data map[string]interface{}, path string) string {
    parts := strings.Split(path, ".")
    current := data
    for i, part := range parts {
        if i == len(parts)-1 {
            if val, ok := current[part].(string); ok {
                return val
            }
            return ""
        }
        if nested, ok := current[part].(map[string]interface{}); ok {
            current = nested
        } else {
            return ""
        }
    }
    return ""
}
```

### SAML Handler

For SAML, use the `crewjam/saml` package:

```go
package saml

import (
    "net/http"
    "os"

    "github.com/crewjam/saml/samlsp"
)

type SAMLHandler struct {
    sp *samlsp.Middleware
}

func NewSAMLHandler() (*SAMLHandler, error) {
    idpMetadata := os.Getenv("SAML_IDP_METADATA")
    idpMetadataURL := os.Getenv("SAML_IDP_METADATA_URL")

    var metadata *samlsp.EntityDescriptor
    var err error

    if idpMetadata != "" {
        metadata, err = samlsp.ParseMetadata([]byte(idpMetadata))
    } else {
        metadata, err = samlsp.FetchMetadata(
            http.DefaultClient,
            idpMetadataURL,
        )
    }
    if err != nil {
        return nil, err
    }

    // Configure SP
    sp, err := samlsp.New(samlsp.Options{
        EntityID:    os.Getenv("SAML_ISSUER"),
        IDPMetadata: metadata,
        // ... additional config
    })
    if err != nil {
        return nil, err
    }

    return &SAMLHandler{sp: sp}, nil
}
```

---

## SSO Mode Configuration

### Environment Variables

| Variable | Effect |
|----------|--------|
| `FEATURES_DISABLE_EMAIL_LOGIN` | Hides email/password login, forces SSO |
| `LOOMIO_SSO_FORCE_USER_ATTRS` | Syncs name/email from IdP on each login |

### SSO-Only Behavior

When `FEATURES_DISABLE_EMAIL_LOGIN` is set:

1. **Existing user by email**: Auto-link identity, sign in
2. **No existing user**: Auto-create verified user from SSO attributes
3. **UID becomes canonical**: Email can change in IdP and sync to Loomio

---

## Identity Linking States

```
                              +---------------+
                              | SSO Provider  |
                              +-------+-------+
                                      |
                                      v
                        +-------------+-------------+
                        | Identity.find_or_create   |
                        | by (identity_type, uid)   |
                        +-------------+-------------+
                                      |
              +-----------------------+-----------------------+
              |                       |                       |
              v                       v                       v
    +------------------+    +------------------+    +------------------+
    | Identity has     |    | Email matches    |    | No match         |
    | user_id          |    | existing user    |    |                  |
    +--------+---------+    +--------+---------+    +--------+---------+
             |                       |                       |
             v                       v                       v
    +------------------+    +------------------+    +------------------+
    | Sign in user     |    | Link identity    |    | Set pending      |
    | Redirect to app  |    | (or verify)      |    | Prompt create    |
    +------------------+    +------------------+    +------------------+
```

---

## API Compatibility

### Frontend Boot Payload

The Vue frontend expects identity providers in the boot payload:

```json
{
  "identityProviders": [
    { "name": "google", "href": "/google/oauth" },
    { "name": "saml", "href": "/saml/oauth" }
  ]
}
```

Only providers with `{PROVIDER}_APP_KEY` set are included.

### Go Boot Handler

```go
type BootPayload struct {
    IdentityProviders []IdentityProvider `json:"identityProviders"`
}

type IdentityProvider struct {
    Name string `json:"name"`
    Href string `json:"href"`
}

func (h *BootHandler) GetProviders() []IdentityProvider {
    providers := []IdentityProvider{}

    providerConfigs := map[string]string{
        "google":    "GOOGLE_APP_KEY",
        "oauth":     "OAUTH_APP_KEY",
        "saml":      "SAML_APP_KEY",
        "nextcloud": "NEXTCLOUD_APP_KEY",
    }

    for name, envVar := range providerConfigs {
        if os.Getenv(envVar) != "" {
            providers = append(providers, IdentityProvider{
                Name: name,
                Href: "/" + name + "/oauth",
            })
        }
    }
    return providers
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestGoogleProvider_FetchIdentity(t *testing.T) {
    // Mock HTTP server returning Google profile
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{
            "id":      "12345",
            "name":    "Test User",
            "email":   "test@example.com",
            "picture": "https://example.com/avatar.jpg",
        })
    }))
    defer server.Close()

    provider := &GoogleProvider{httpClient: server.Client()}
    identity, err := provider.FetchIdentity(context.Background(), "fake-token")

    require.NoError(t, err)
    assert.Equal(t, "12345", identity.UID)
    assert.Equal(t, "test@example.com", identity.Email)
}
```

### Integration Tests

```go
func TestOAuthCallback_NewUser(t *testing.T) {
    // Setup: No existing user or identity
    db := testutil.SetupTestDB(t)
    defer db.Close()

    // Mock OAuth provider
    mockProvider := &mocks.Provider{}
    mockProvider.On("ExchangeCode", mock.Anything, "auth-code", mock.Anything).
        Return("access-token", nil)
    mockProvider.On("FetchIdentity", mock.Anything, "access-token").
        Return(&oauth.IdentityParams{
            UID:   "new-user-123",
            Email: "newuser@example.com",
            Name:  "New User",
        }, nil)

    // Execute callback
    req := httptest.NewRequest("GET", "/google/authorize?code=auth-code", nil)
    rec := httptest.NewRecorder()
    handler.OAuthCallback(mockProvider).ServeHTTP(rec, req)

    // Assert: Identity created, pending state set
    assert.Equal(t, http.StatusFound, rec.Code)
    // Verify session has pending_identity_id
}
```

---

## Summary: Rails to Go Mapping

| Rails Component | Go Equivalent |
|-----------------|---------------|
| `Identity` model | `Identity` struct + sqlc queries |
| `Identities::BaseController` | `oauth.Handler` struct |
| `Identities::GoogleController` | `oauth.GoogleProvider` |
| `Clients::Google` | Methods on `GoogleProvider` |
| `config/providers.yml` | Hardcoded provider list or config file |
| `session[:pending_identity_id]` | Session store (gorilla/sessions or similar) |
| `ruby-saml` gem | `crewjam/saml` package |
| `ENV['GOOGLE_APP_KEY']` | `os.Getenv("GOOGLE_APP_KEY")` |
