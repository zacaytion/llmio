# OAuth Providers: Identity Models

## Identity Model

**File:** `/Users/z/Code/loomio/app/models/identity.rb`

```ruby
class Identity < ApplicationRecord
  extend HasCustomFields
  self.table_name = :omniauth_identities

  validates :identity_type, presence: true
  validates :uid, presence: true

  belongs_to :user, required: false

  PROVIDERS = YAML.load_file(Rails.root.join("config", "providers.yml"))['identity']

  scope :with_user, -> { where.not(user: nil) }

  def force_user_attrs!
    user.update(name: name, email: email)
  end

  def assign_logo!
    return unless user && logo
    user.uploaded_avatar.attach(
      io: URI.open(URI.parse(logo)),
      filename: File.basename(logo)
    )
    user.update(avatar_kind: :uploaded)
  rescue OpenURI::HTTPError, TypeError
    # Can't load logo uri as attachment; do nothing
  end
end
```

## Database Schema

**File:** `/Users/z/Code/loomio/db/schema.rb` (lines 619-633)

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `user_id` | integer | Foreign key to users table (nullable) |
| `email` | string(255) | Email from SSO provider |
| `identity_type` | string(255) | Provider name: google, oauth, saml, nextcloud |
| `uid` | string(255) | Unique identifier from SSO provider |
| `name` | string(255) | Display name from SSO provider |
| `access_token` | string | OAuth access token (empty for SAML) |
| `logo` | string | URL to user's avatar from SSO |
| `custom_fields` | jsonb | Extensible JSON fields |
| `created_at` | timestamp | Record creation time |
| `updated_at` | timestamp | Record update time |

### Indexes

- `index_omniauth_identities_on_identity_type_and_uid` - Composite unique index for provider+uid lookup
- `index_personas_on_email` - Index on email (legacy name from migration)
- `index_personas_on_user_id` - Index on user association

## Provider Constants

**File:** `/Users/z/Code/loomio/config/providers.yml`

```yaml
identity:
  - oauth
  - saml
  - google
  - nextcloud
```

## Identity Type Values

| identity_type | Provider | Authentication Protocol |
|---------------|----------|------------------------|
| `google` | Google | OAuth 2.0 |
| `oauth` | Generic (configurable) | OAuth 2.0 |
| `saml` | Enterprise IdP | SAML 2.0 |
| `nextcloud` | Nextcloud | OAuth 2.0 |

## User Association

An Identity can exist in two states:

1. **Linked** (`user_id` populated): Identity is associated with a Loomio user
2. **Pending** (`user_id` is NULL): SSO succeeded but no Loomio user linked yet

When `user_id` is NULL, the identity ID is stored in `session[:pending_identity_id]` and the user is prompted to:
- Create a new account
- Link to existing account via email verification

## SSO Mode Behavior

When `ENV['FEATURES_DISABLE_EMAIL_LOGIN']` is set:

1. If user exists by email -> auto-link identity
2. If no user exists -> auto-create verified user from SSO attributes
3. UID becomes the source of truth (email can change in SSO and sync to Loomio)

When `ENV['LOOMIO_SSO_FORCE_USER_ATTRS']` is set:

1. On each SSO login, user's `name` and `email` are synced from SSO provider
2. User cannot edit these fields in Loomio settings

## Client Models

Each provider has a corresponding client class in `/Users/z/Code/loomio/app/extras/clients/`:

### Base Client

**File:** `/Users/z/Code/loomio/app/extras/clients/base.rb`

Provides:
- HTTP GET/POST methods with JSON handling
- Default parameter injection (client_id, client_secret, token)
- Environment variable based configuration via `self.instance`

### Google Client

**File:** `/Users/z/Code/loomio/app/extras/clients/google.rb`

```ruby
class Clients::Google < Clients::Base
  def fetch_access_token(code, redirect_uri)
    data = post("token", params: { code: code, redirect_uri: redirect_uri, grant_type: :authorization_code }).json
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

  def scope
    %w(email profile).freeze
  end

  def default_host
    "https://www.googleapis.com/oauth2/v4".freeze
  end
end
```

**Environment variables:** `GOOGLE_APP_KEY`, `GOOGLE_APP_SECRET`

### Generic OAuth Client

**File:** `/Users/z/Code/loomio/app/extras/clients/oauth.rb`

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

**Environment variables:**
- `OAUTH_APP_KEY` - Client ID
- `OAUTH_APP_SECRET` - Client secret
- `OAUTH_AUTH_URL` - Authorization endpoint
- `OAUTH_TOKEN_URL` - Token exchange endpoint
- `OAUTH_PROFILE_URL` - User info endpoint
- `OAUTH_SCOPE` - Requested scopes
- `OAUTH_ATTR_UID` - JSON path to user ID in profile response
- `OAUTH_ATTR_NAME` - JSON path to name in profile response
- `OAUTH_ATTR_EMAIL` - JSON path to email in profile response

### Nextcloud Client

**File:** `/Users/z/Code/loomio/app/extras/clients/nextcloud.rb`

```ruby
class Clients::Nextcloud < Clients::Base
  def fetch_access_token(code, uri)
    post(
      'index.php/apps/oauth2/api/v1/token',
      params: { code: code, redirect_uri: uri, grant_type: :authorization_code }
    ).json['access_token']
  end

  def fetch_identity_params
    data = get('ocs/v2.php/cloud/user', params: { format: :json }).json['ocs']['data']
    {
      uid: data['id'],
      email: data['email']
    }
  end

  def default_host
    ENV['NEXTCLOUD_HOST']
  end
end
```

**Environment variables:** `NEXTCLOUD_APP_KEY`, `NEXTCLOUD_APP_SECRET`, `NEXTCLOUD_HOST`

## SAML Model

SAML does not use a client class. It uses the `ruby-saml` gem directly in the controller. Identity params are extracted from the SAML response:

```ruby
# From /Users/z/Code/loomio/app/controllers/identities/saml_controller.rb:17-23
identity_params = {
  identity_type: 'saml',
  uid: saml_response.nameid,
  email: saml_response.nameid,
  name: saml_response.attributes['displayName'],
  access_token: nil
}
```

**Environment variables:**
- `SAML_IDP_METADATA` - Inline IdP metadata XML
- `SAML_IDP_METADATA_URL` - URL to fetch IdP metadata
- `SAML_ISSUER` - SP entity ID (defaults to metadata URL)
- `SAML_LOGIN_PROVIDER_NAME` - Display name for login button
