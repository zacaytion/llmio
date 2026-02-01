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

---

## Summary: Rails Component Mapping

| Rails Component | Purpose |
|-----------------|---------|
| `Identity` model | Identity storage |
| `Identities::BaseController` | OAuth flow orchestration |
| `Identities::GoogleController` | Google-specific handling |
| `Clients::Google` | Token exchange & user info |
| `config/providers.yml` | Provider configuration |
| `session[:pending_identity_id]` | Session state for unlinked identities |
| `ruby-saml` gem | SAML 2.0 handling |
