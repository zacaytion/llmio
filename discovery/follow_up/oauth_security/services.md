# OAuth Security - Service Flows Documentation

## OAuth Flow Overview

Loomio implements a custom OAuth 2.0 Authorization Code flow without using OmniAuth.

### Flow Diagram

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Browser   │     │  Loomio Server   │     │  OAuth Provider │
└─────────────┘     └──────────────────┘     └─────────────────┘
       │                    │                        │
       │  GET /google/oauth │                        │
       │───────────────────>│                        │
       │                    │                        │
       │  302 Redirect to   │                        │
       │  OAuth Provider    │                        │
       │<───────────────────│                        │
       │                    │                        │
       │  GET /authorize?client_id=...&redirect_uri=...&scope=...
       │────────────────────────────────────────────>│
       │                    │                        │
       │  User Authenticates│                        │
       │  & Grants Access   │                        │
       │                    │                        │
       │  302 Redirect with │                        │
       │  ?code=AUTH_CODE   │                        │
       │<────────────────────────────────────────────│
       │                    │                        │
       │  GET /google/authorize?code=AUTH_CODE       │
       │───────────────────>│                        │
       │                    │                        │
       │                    │  POST /token           │
       │                    │  {code, client_secret} │
       │                    │───────────────────────>│
       │                    │                        │
       │                    │  {access_token}        │
       │                    │<───────────────────────│
       │                    │                        │
       │                    │  GET /userinfo         │
       │                    │  Authorization: Bearer │
       │                    │───────────────────────>│
       │                    │                        │
       │                    │  {uid, email, name}    │
       │                    │<───────────────────────│
       │                    │                        │
       │  Sign in user OR   │                        │
       │  create identity   │                        │
       │<───────────────────│                        │
       │                    │                        │
```

## Controller Hierarchy

```
ApplicationController
    └── Identities::BaseController
            ├── Identities::GoogleController
            ├── Identities::OauthController
            └── Identities::NextcloudController

ApplicationController
    └── Identities::SamlController (separate hierarchy)
```

## Service Components

### 1. Identities::BaseController

**File**: `app/controllers/identities/base_controller.rb`

#### Action: `oauth` (Initiate OAuth flow)

```ruby
def oauth
  session[:back_to] = params[:back_to] || request.referrer
  redirect_to oauth_url
end
```

**Security Issue**: No state parameter generated or stored.

#### Action: `create` (OAuth callback handler)

```ruby
def create
  if params[:error].present?
    flash[:error] = t(:'auth.oauth_cancelled')
    return redirect_to session.delete(:back_to) || dashboard_path
  end

  access_token = fetch_access_token
  return respond_with_error(401, "OAuth authorization failed") unless access_token.present?

  identity_params = fetch_identity_params(access_token)
  return respond_with_error(401, "...") unless identity_params[:uid].present? && identity_params[:email].present?

  # Find or create identity by uid
  identity = Identity.find_by(identity_params.slice(:uid, :identity_type))

  if identity
    identity.update(identity_params)
  else
    identity = Identity.new(identity_params)
    # Link to current_user or find by email
    if ENV['FEATURES_DISABLE_EMAIL_LOGIN']
      identity.user = current_user.presence || User.find_by(email: identity.email)
      if identity.user.nil?
        identity.user = User.new(identity_params.slice(:name, :email).merge(email_verified: true))
        identity.user.save!
      end
    else
      identity.user = current_user.presence || User.verified.find_by(email: identity.email)
    end
    identity.save
  end

  if identity.user
    identity.force_user_attrs! if ENV['LOOMIO_SSO_FORCE_USER_ATTRS']
    sign_in(identity.user)
    flash[:notice] = t(:'devise.sessions.signed_in')
  else
    session[:pending_identity_id] = identity.id
  end

  redirect_to session.delete(:back_to) || dashboard_path
end
```

**Security Issues**:
1. No state parameter validation before processing `code`
2. `code` parameter directly passed to token exchange
3. Attacker can craft callback URL with their authorization code

### 2. OAuth Clients

**File**: `app/extras/clients/base.rb`

Base class for HTTP communication with OAuth providers.

**File**: `app/extras/clients/google.rb`

```ruby
class Clients::Google < Clients::Base
  def fetch_access_token(code, redirect_uri)
    data = post("token", params: {
      code: code,
      redirect_uri: redirect_uri,
      grant_type: :authorization_code
    }).json
    data['access_token']
  end

  def fetch_identity_params
    data = get("userinfo", options: { host: :"https://www.googleapis.com/oauth2/v2" }).json
    {
      uid: data['id'],
      name: data['name'],
      email: data['email'],
      logo: data['picture'],
      identity_type: 'google'
    }
  end
end
```

**File**: `app/extras/clients/oauth.rb`

```ruby
class Clients::Oauth < Clients::Base
  def fetch_access_token(code, uri)
    post(
      ENV.fetch('OAUTH_TOKEN_URL'),
      params: { code: code, redirect_uri: uri, grant_type: :authorization_code }
    ).json['access_token']
  end

  def fetch_identity_params
    data = get(ENV.fetch('OAUTH_PROFILE_URL')).json
    {
      uid: data.dig(ENV.fetch('OAUTH_ATTR_UID')),
      name: data.dig(ENV.fetch('OAUTH_ATTR_NAME')),
      email: data.dig(ENV.fetch('OAUTH_ATTR_EMAIL'))
    }
  end
end
```

**File**: `app/extras/clients/nextcloud.rb`

Similar pattern to Google, with Nextcloud-specific endpoints.

## SAML Flow (Separate)

**File**: `app/controllers/identities/saml_controller.rb`

SAML uses POST binding with XML signatures for security. The controller:

1. Skips CSRF verification (required for SAML POST binding)
2. Uses `ruby-saml` gem for response validation
3. Does NOT validate SAML signatures by default (security concern)

```ruby
skip_before_action :verify_authenticity_token

def create
  saml_response = OneLogin::RubySaml::Response.new(params[:SAMLResponse], skip_recipient_check: true)
  saml_response.settings = saml_settings

  return respond_with_error(500, "SAML response is not valid") unless saml_response.is_valid?
  # ...
end
```

**SAML Security Settings** (Line 95-102):
```ruby
settings.security[:authn_requests_signed] = false
settings.security[:logout_requests_signed] = false
settings.security[:logout_responses_signed] = false
settings.security[:metadata_signed] = false
```

These settings disable signature requirements, which may be intentional for compatibility but reduces security.

## Route Definitions

**File**: `config/routes.rb:455-466`

```ruby
Identity::PROVIDERS.each do |provider|
  scope provider do
    get :oauth,      to: "identities/#{provider}#oauth",   as: :"#{provider}_oauth"
    get :authorize,  to: "identities/#{provider}#create",  as: :"#{provider}_authorize"
    get '/',         to: "identities/#{provider}#destroy", as: :"#{provider}_unauthorize"
  end
end

scope :saml do
  post :oauth,    to: 'identities/saml#create',   as: :saml_oauth_callback
  get :metadata,  to: 'identities/saml#metadata', as: :saml_metadata
end
```

**Generated Routes**:
- `GET /google/oauth` - Initiate Google OAuth
- `GET /google/authorize` - Google OAuth callback
- `GET /oauth/oauth` - Initiate generic OAuth
- `GET /oauth/authorize` - Generic OAuth callback
- `GET /nextcloud/oauth` - Initiate Nextcloud OAuth
- `GET /nextcloud/authorize` - Nextcloud OAuth callback
- `POST /saml/oauth` - SAML callback (POST only)
- `GET /saml/metadata` - SAML metadata endpoint

## Environment Variables

### Google OAuth
- `GOOGLE_APP_KEY` - Google OAuth client ID
- `GOOGLE_APP_SECRET` - Google OAuth client secret

### Generic OAuth
- `OAUTH_APP_KEY` - OAuth client ID
- `OAUTH_APP_SECRET` - OAuth client secret
- `OAUTH_AUTH_URL` - Authorization endpoint
- `OAUTH_TOKEN_URL` - Token endpoint
- `OAUTH_PROFILE_URL` - User info endpoint
- `OAUTH_SCOPE` - OAuth scopes
- `OAUTH_ATTR_UID` - JSON path for user ID
- `OAUTH_ATTR_NAME` - JSON path for user name
- `OAUTH_ATTR_EMAIL` - JSON path for user email

### Nextcloud OAuth
- `NEXTCLOUD_HOST` - Nextcloud server URL
- `NEXTCLOUD_APP_KEY` - OAuth client ID
- `NEXTCLOUD_APP_SECRET` - OAuth client secret

### SAML
- `SAML_IDP_METADATA` - Inline IdP metadata XML
- `SAML_IDP_METADATA_URL` - IdP metadata URL
- `SAML_ISSUER` - Service provider issuer

### SSO Behavior
- `FEATURES_DISABLE_EMAIL_LOGIN` - Require SSO for all logins
- `LOOMIO_SSO_FORCE_USER_ATTRS` - Sync user attributes from SSO on each login

## Security Analysis Summary

| Component | Risk | Issue |
|-----------|------|-------|
| OAuth initiation | HIGH | GET request without CSRF token |
| OAuth callback | HIGH | No state parameter validation |
| Token storage | MEDIUM | Plaintext in database |
| SAML signatures | MEDIUM | Disabled by default |
| Session binding | LOW | Uses standard Rails sessions |
