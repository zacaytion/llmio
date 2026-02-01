# OAuth Providers: Controller Documentation

## Controller Hierarchy

```
Identities::BaseController < ApplicationController
├── Identities::GoogleController
├── Identities::NextcloudController
└── Identities::OauthController

Identities::SamlController < ApplicationController  (standalone)
```

## Base Controller

**File:** `/Users/z/Code/loomio/app/controllers/identities/base_controller.rb`

### Actions

#### `oauth` (GET)
Initiates OAuth flow by redirecting to provider's authorization endpoint.

```ruby
def oauth
  session[:back_to] = params[:back_to] || request.referrer
  redirect_to oauth_url
end
```

**Query params:**
- `back_to` (optional): URL to return to after authentication

#### `create` (GET)
Handles OAuth callback from provider.

**Flow:**
1. Check for error param (user cancelled)
2. Exchange code for access token
3. Fetch user profile from provider
4. Find or create Identity record by uid+identity_type
5. Link to existing user or set pending identity
6. Sign in user (if linked) and redirect

**Query params:**
- `code`: Authorization code from provider
- `error`: Error code if user cancelled

**Session effects:**
- Clears `session[:back_to]`
- May set `session[:pending_identity_id]` if no user linked

#### `destroy` (GET)
Removes identity connection for current user.

```ruby
def destroy
  if i = current_user.identities.find_by(identity_type: controller_name)
    i.destroy
    redirect_to request.referrer || root_path
  else
    respond_with_error 500, "Not connected to #{controller_name}!"
  end
end
```

### Private Methods

```ruby
def fetch_access_token
  client = "Clients::#{controller_name.classify}".constantize.instance
  client.fetch_access_token(params[:code], redirect_uri)
end

def fetch_identity_params(token)
  client = "Clients::#{controller_name.classify}".constantize.new(token: token)
  client.fetch_identity_params.merge({ access_token: token, identity_type: controller_name })
end

def redirect_uri
  send :"#{controller_name}_authorize_url"
end

def oauth_url
  "#{oauth_host}?#{oauth_params.to_query}"
end
```

## Google Controller

**File:** `/Users/z/Code/loomio/app/controllers/identities/google_controller.rb`

Inherits from BaseController with Google-specific overrides:

```ruby
class Identities::GoogleController < Identities::BaseController
  private

  def oauth_url
    super.gsub("%2B", "+")  # Fix scope URL encoding
  end

  def oauth_host
    "https://accounts.google.com/o/oauth2/v2/auth"
  end

  def oauth_params
    super.merge(response_type: :code, scope: client.scope.join('+'))
  end

  def client
    Clients::Google.instance
  end
end
```

### Routes

| Method | Path | Action | Named Route |
|--------|------|--------|-------------|
| GET | `/google/oauth` | `oauth` | `google_oauth_path` |
| GET | `/google/authorize` | `create` | `google_authorize_path` |
| GET | `/google/` | `destroy` | `google_unauthorize_path` |

## Nextcloud Controller

**File:** `/Users/z/Code/loomio/app/controllers/identities/nextcloud_controller.rb`

```ruby
class Identities::NextcloudController < Identities::BaseController
  private

  def oauth_host
    ENV['NEXTCLOUD_HOST']
  end

  def oauth_url
    "#{oauth_host}#{oauth_authorize_path}?#{oauth_params.to_query}"
  end

  def oauth_authorize_path
    '/index.php/apps/oauth2/authorize'.freeze
  end

  def oauth_params
    { client.client_key_name => client.key, redirect_uri: redirect_uri, response_type: :code }
  end

  def client
    Clients::Nextcloud.instance
  end
end
```

### Routes

| Method | Path | Action | Named Route |
|--------|------|--------|-------------|
| GET | `/nextcloud/oauth` | `oauth` | `nextcloud_oauth_path` |
| GET | `/nextcloud/authorize` | `create` | `nextcloud_authorize_path` |
| GET | `/nextcloud/` | `destroy` | `nextcloud_unauthorize_path` |

## Generic OAuth Controller

**File:** `/Users/z/Code/loomio/app/controllers/identities/oauth_controller.rb`

```ruby
class Identities::OauthController < Identities::BaseController
  private

  def oauth_url
    "#{oauth_auth_url}?#{oauth_params.to_query}"
  end

  def oauth_auth_url
    ENV.fetch('OAUTH_AUTH_URL')
  end

  def oauth_params
    client = Clients::Oauth.instance
    {
      client.client_key_name => client.key,
      redirect_uri: redirect_uri,
      scope: ENV.fetch('OAUTH_SCOPE'),
      response_type: :code
    }
  end
end
```

### Routes

| Method | Path | Action | Named Route |
|--------|------|--------|-------------|
| GET | `/oauth/oauth` | `oauth` | `oauth_oauth_path` |
| GET | `/oauth/authorize` | `create` | `oauth_authorize_path` |
| GET | `/oauth/` | `destroy` | `oauth_unauthorize_path` |

## SAML Controller

**File:** `/Users/z/Code/loomio/app/controllers/identities/saml_controller.rb`

SAML controller does NOT inherit from BaseController. It uses `ruby-saml` gem directly.

### Actions

#### `oauth` (GET)
Initiates SAML authentication request.

```ruby
def oauth
  session[:back_to] = params[:back_to] || request.referrer
  auth_request = OneLogin::RubySaml::Authrequest.new
  redirect_to auth_request.create(saml_settings)
end
```

#### `create` (POST)
Handles SAML response callback.

```ruby
def create
  saml_response = OneLogin::RubySaml::Response.new(params[:SAMLResponse], skip_recipient_check: true)
  saml_response.settings = saml_settings

  return respond_with_error(500, "SAML response is not valid") unless saml_response.is_valid?

  identity_params = {
    identity_type: 'saml',
    uid: saml_response.nameid,
    email: saml_response.nameid,
    name: saml_response.attributes['displayName'],
    access_token: nil
  }
  # ... identity linking logic same as OAuth
end
```

#### `metadata` (GET)
Serves SAML SP metadata XML for IdP configuration.

```ruby
def metadata
  meta = OneLogin::RubySaml::Metadata.new
  render xml: meta.generate(saml_settings), content_type: "application/samlmetadata+xml"
end
```

#### `destroy` (GET)
Removes SAML identity connection.

### Routes

| Method | Path | Action | Named Route |
|--------|------|--------|-------------|
| GET | `/saml/oauth` | `oauth` | `saml_oauth_path` |
| POST | `/saml/oauth` | `create` | `saml_oauth_callback_path` |
| GET | `/saml/metadata` | `metadata` | `saml_metadata_path` |
| GET | `/saml/` | `destroy` | `saml_unauthorize_path` |

### SAML Settings

```ruby
def saml_settings
  if ENV['SAML_IDP_METADATA']
    settings = OneLogin::RubySaml::IdpMetadataParser.new.parse(ENV['SAML_IDP_METADATA'])
  else
    settings = OneLogin::RubySaml::IdpMetadataParser.new.parse_remote(ENV.fetch('SAML_IDP_METADATA_URL'))
  end

  settings.assertion_consumer_service_url = saml_oauth_callback_url
  settings.issuer = ENV.fetch('SAML_ISSUER', saml_metadata_url)
  settings.assertion_consumer_logout_service_url = saml_unauthorize_url
  settings.name_identifier_format = 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress'

  # Security settings
  settings.soft = true
  settings.security[:authn_requests_signed] = false
  settings.security[:logout_requests_signed] = false
  settings.security[:logout_responses_signed] = false
  settings.security[:metadata_signed] = false
  settings.security[:digest_method] = XMLSecurity::Document::SHA1
  settings.security[:signature_method] = XMLSecurity::Document::RSA_SHA1

  settings
end
```

## API Endpoint for Identity Commands

**File:** `/Users/z/Code/loomio/config/routes.rb:334`

```ruby
get "identities/:id/:command", to: "api/v1/identities#command"
```

This endpoint is under the API namespace and handles commands on identity records (likely for admin operations).

## Error Handling

All controllers inherit error handling from ApplicationController:

- `respond_with_error(status, message)` - Returns JSON error response
- OAuth cancellation sets `flash[:error]` and redirects
- Token/profile fetch failures return 401 JSON response

## Test Coverage

**File:** `/Users/z/Code/loomio/spec/controllers/identities/oauth_controller_spec.rb`

Comprehensive tests covering:
- OAuth redirect with correct parameters
- Successful authentication flows
- User creation in SSO-only mode
- Identity linking to existing users
- SSO attribute syncing
- Error handling (cancellation, token failure, profile failure)
- Identity destruction

**File:** `/Users/z/Code/loomio/spec/controllers/identities/saml_controller_spec.rb`

Similar comprehensive coverage for SAML flows.

Note: No dedicated specs for Google or Nextcloud controllers (covered by generic OAuth tests).
