# OAuth Security - Models Documentation

## Identity Model

**File**: `app/models/identity.rb`

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

### Table Schema

**Table**: `omniauth_identities` (legacy naming, predates custom implementation)

| Column | Type | Notes |
|--------|------|-------|
| id | serial | Primary key |
| identity_type | string | Provider type (google, oauth, saml, nextcloud) |
| uid | string | Unique identifier from OAuth provider |
| user_id | integer | Foreign key to users table |
| email | string | Email from OAuth profile |
| name | string | Display name from OAuth profile |
| logo | string | Avatar URL from OAuth profile |
| access_token | string | OAuth access token (stored for API access) |
| custom_fields | jsonb | Additional provider-specific data |
| created_at | timestamp | |
| updated_at | timestamp | |

**Index**: `index_omniauth_identities_on_identity_type_and_uid` (unique constraint)

### Providers Configuration

**File**: `config/providers.yml`

```yaml
identity:
  - oauth      # Generic OAuth2 provider (configurable via ENV)
  - saml       # SAML 2.0 provider
  - google     # Google OAuth2
  - nextcloud  # Nextcloud OAuth2
```

### Relationships

```
User (1) ----< Identity (many)
```

A user can have multiple identity records (one per OAuth provider). An identity can exist without a user (pending state during OAuth flow).

### Identity Lifecycle

1. **Creation during OAuth callback**:
   - User authenticates with OAuth provider
   - Identity created with `uid`, `identity_type`, `email`, `name`
   - If user logged in: linked to `current_user`
   - If user has matching verified email: linked to existing user
   - Otherwise: stored in session as `pending_identity_id`

2. **SSO-only mode** (`ENV['FEATURES_DISABLE_EMAIL_LOGIN']`):
   - User is automatically created if not found
   - User is marked as `email_verified: true`

3. **Attribute syncing** (`ENV['LOOMIO_SSO_FORCE_USER_ATTRS']`):
   - User's name and email are overwritten from SSO on each login

### Security Considerations

1. **Access Token Storage**: OAuth access tokens are stored in plaintext in the database. This is a security concern if database is compromised.

2. **No Token Refresh**: There is no refresh token handling visible in the codebase. Long-lived access tokens may expire.

3. **UID as Immutable Key**: The `uid` field is treated as the source of truth for identity matching, which is correct. Email changes in the OAuth provider won't break identity linking.

## User Model (OAuth-relevant portions)

**File**: `app/models/user.rb` (excerpted)

### OAuth-Related Fields

| Field | Type | Purpose |
|-------|------|---------|
| email | string | Primary identifier, may be updated from OAuth |
| email_verified | boolean | True if email confirmed or SSO-created |
| name | string | Display name, may be updated from OAuth |

### OAuth-Related Associations

```ruby
has_many :identities
```

### OAuth-Related Scopes

```ruby
scope :verified, -> { where(email_verified: true) }
```

Used in identity linking to only auto-link to verified accounts.

## Session Model (OAuth-relevant)

OAuth uses Rails session for:

| Session Key | Purpose |
|-------------|---------|
| `back_to` | Redirect URL after OAuth completion |
| `pending_identity_id` | Identity ID when no user match found |
| `_csrf_token` | Standard Rails CSRF token (not used in OAuth flow) |

**MISSING**: No `oauth_state` session key for CSRF protection.
