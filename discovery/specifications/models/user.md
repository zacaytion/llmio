# User Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/user.rb`
- `/app/models/identity.rb`
- `/app/models/login_token.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The User model represents authenticated accounts in Loomio. Users can be human members or bot accounts, and may authenticate via password, OAuth/SSO, or passwordless login tokens.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `email` | citext | - | YES | UNIQUE, max 200 | Case-insensitive email address |
| `name` | string(255) | - | YES | max 100 | Display name |
| `username` | string(255) | - | YES | UNIQUE, max 30, alphanumeric only | URL-safe identifier |
| `key` | string(255) | - | YES | UNIQUE | Public URL key (8 chars) |
| `short_bio` | string | "" | NO | max 5000 | User biography |
| `short_bio_format` | string(10) | "md" | NO | "md" or "html" | Biography format |
| `location` | string | "" | NO | - | User-entered location |
| `bot` | boolean | false | NO | - | Bot account flag |

### Authentication (Devise)

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `encrypted_password` | string(128) | "" | YES | min 8 chars when set | Devise password hash |
| `reset_password_token` | string | - | YES | UNIQUE | Password reset token |
| `reset_password_sent_at` | datetime | - | YES | - | When reset token was sent |
| `remember_created_at` | datetime | - | YES | - | Remember me timestamp |
| `remember_token` | string | - | YES | indexed | Session persistence |
| `sign_in_count` | integer | 0 | NO | - | Total sign-in count |
| `current_sign_in_at` | datetime | - | YES | - | Current session start |
| `last_sign_in_at` | datetime | - | YES | - | Previous session start |
| `current_sign_in_ip` | inet | - | YES | - | Current IP address |
| `last_sign_in_ip` | inet | - | YES | - | Previous IP address |
| `failed_attempts` | integer | 0 | NO | - | Devise lockable counter |
| `unlock_token` | string | - | YES | UNIQUE | Devise unlock token |
| `locked_at` | datetime | - | YES | - | Account lock timestamp |
| `email_verified` | boolean | false | NO | indexed | Email verification status |

### Tokens

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `secret_token` | string | gen_random_uuid() | NO | - | Internal secret |
| `unsubscribe_token` | string(255) | auto-generated | YES | UNIQUE | Email unsubscribe token |
| `email_api_key` | string(255) | auto-generated | YES | - | Reply-by-email authentication |
| `api_key` | string | auto-generated | YES | indexed | Bot API authentication |
| `authentication_token` | string(255) | - | YES | - | Legacy auth token |

### Avatar

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `avatar_kind` | string(255) | "initials" | NO | - | Avatar type: initials, uploaded, gravatar |
| `avatar_initials` | string(255) | - | YES | max 3 chars | Computed initials |
| `uploaded_avatar_file_name` | string(255) | - | YES | - | Paperclip filename (legacy) |
| `uploaded_avatar_content_type` | string(255) | - | YES | - | Paperclip content type |
| `uploaded_avatar_file_size` | integer | - | YES | - | Paperclip file size |
| `uploaded_avatar_updated_at` | datetime | - | YES | - | Paperclip update time |

**Note:** ActiveStorage attachment `uploaded_avatar` replaces Paperclip.

### Locale & Timezone

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `selected_locale` | string(255) | - | YES | - | User's preferred locale |
| `detected_locale` | string(255) | - | YES | - | Auto-detected locale |
| `content_locale` | string | - | YES | - | Content creation locale |
| `time_zone` | string(255) | - | YES | defaults to 'UTC' | User timezone |
| `autodetect_time_zone` | boolean | true | NO | - | Auto-detect timezone |
| `date_time_pref` | string | - | YES | defaults to 'day_abbr' | Date/time format preference |

### Email Preferences

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `email_catch_up` | boolean | true | NO | - | Receive digest emails |
| `email_catch_up_day` | integer | - | YES | 0-8 | Digest frequency |
| `email_when_mentioned` | boolean | true | NO | - | Email on @mention |
| `email_on_participation` | boolean | false | NO | - | Email on participation |
| `email_when_proposal_closing_soon` | boolean | false | NO | - | Email on poll closing |
| `email_newsletter` | boolean | false | NO | - | Marketing emails |
| `default_membership_volume` | integer | 2 | NO | enum: 0-3 | Default notification volume |

**Volume Enum Values:**
- 0: mute
- 1: quiet
- 2: normal (default)
- 3: loud

### GeoIP Location

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `country` | string | - | YES | GeoIP country |
| `region` | string | - | YES | GeoIP region |
| `city` | string | - | YES | GeoIP city |

### Status & Counters

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `created_at` | datetime | - | YES | Account creation time |
| `updated_at` | datetime | - | YES | Last update time |
| `deactivated_at` | datetime | - | YES | Soft delete timestamp |
| `deactivator_id` | integer | - | YES | User who deactivated |
| `is_admin` | boolean | false | YES | System administrator flag |
| `last_seen_at` | datetime | - | YES | Last activity timestamp |
| `legal_accepted_at` | datetime | - | YES | Terms acceptance time |
| `memberships_count` | integer | 0 | NO | Counter cache |
| `complaints_count` | integer | 0 | NO | Spam complaint counter |
| `auto_translate` | boolean | false | NO | Auto-translate content |

### JSONB Fields

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `experiences` | jsonb | {} | Feature flags/tutorials seen |
| `attachments` | jsonb | [] | Profile attachments |
| `link_previews` | jsonb | [] | Cached link previews |

**experiences structure:**
```json
{
  "welcomeModal": true,
  "announcementHelpCard": true,
  "pollTypes": ["proposal", "count"],
  "html-editor.uses-markdown": true
}
```

### Legacy Fields

| Column | Type | Description |
|--------|------|-------------|
| `email_sha256` | string | Hashed email for matching |
| `facebook_community_id` | integer | Legacy FB integration |
| `slack_community_id` | integer | Legacy Slack integration |

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `email` | presence, email format, max 200, uniqueness | always |
| `email` | exclusion from forbidden addresses | NoForbiddenEmails concern |
| `name` | presence | if `require_valid_signup` |
| `name` | max length 100 | always |
| `username` | uniqueness | if email present |
| `username` | max length 30 | always |
| `username` | format: alphanumeric only (`/\A[a-z0-9]*\z/`) | always |
| `short_bio` | max length 5000 | always |
| `password` | min length 8 | when password provided |
| `password` | confirmation match | when password/confirmation provided |
| `legal_accepted` | presence | if `require_valid_signup` AND `ENV['TERMS_URL']` set |
| `name`, `email` | no spam regex | NoSpam concern |

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Has Many

| Association | Class | Foreign Key | Options | Description |
|-------------|-------|-------------|---------|-------------|
| `memberships` | Membership | user_id | `-> { active }`, dependent: :destroy | Active group memberships |
| `all_memberships` | Membership | user_id | dependent: :destroy | All memberships including revoked |
| `admin_memberships` | Membership | user_id | `-> { where(admin: true, revoked_at: nil) }` | Admin memberships |
| `groups` | Group | through: :memberships | `-> { where(archived_at: nil) }` | Active groups |
| `adminable_groups` | Group | through: :admin_memberships | `-> { where(archived_at: nil) }` | Groups user can admin |
| `discussions` | Discussion | through: :groups | - | Discussions in user's groups |
| `authored_discussions` | Discussion | author_id | dependent: :destroy | Discussions created by user |
| `authored_polls` | Poll | author_id | dependent: :destroy | Polls created by user |
| `created_groups` | Group | creator_id | dependent: :destroy | Groups created by user |
| `identities` | Identity | user_id | dependent: :destroy | OAuth/SSO identities |
| `reactions` | Reaction | user_id | dependent: :destroy | Emoji reactions |
| `stances` | Stance | participant_id | dependent: :destroy | Poll votes |
| `participated_polls` | Poll | through: :stances | - | Polls user voted in |
| `group_polls` | Poll | through: :groups | source: :polls | All polls in user's groups |
| `discussion_readers` | DiscussionReader | user_id | dependent: :destroy | Read tracking records |
| `guest_discussion_readers` | DiscussionReader | user_id | `-> { active.guests }` | Guest access to discussions |
| `guest_discussions` | Discussion | through: :guest_discussion_readers | - | Discussions user is guest in |
| `guest_stances` | Stance | participant_id | `-> { latest.guests }` | Guest voting records |
| `guest_polls` | Poll | through: :guest_stances | - | Polls user is guest voter in |
| `notifications` | Notification | user_id | dependent: :destroy | User notifications |
| `comments` | Comment | user_id | dependent: :destroy | Comments authored |
| `documents` | Document | author_id | dependent: :destroy | Documents uploaded |
| `login_tokens` | LoginToken | user_id | dependent: :destroy | Passwordless login tokens |
| `events` | Event | user_id | dependent: :destroy | Events triggered by user |
| `membership_requests` | MembershipRequest | requestor_id | dependent: :destroy | Pending join requests |
| `tags` | Tag | through: :groups | - | Tags in user's groups |

---

## Scopes

```ruby
scope :active, -> { where(deactivated_at: nil) }
scope :deactivated, -> { where("deactivated_at IS NOT NULL") }
scope :no_spam_complaints, -> { where(complaints_count: 0) }
scope :has_spam_complaints, -> { where("complaints_count > 0") }
scope :sorted_by_name, -> { order("lower(name)") }
scope :admins, -> { where(is_admin: true) }
scope :coordinators, -> {
  joins(:memberships).where('memberships.admin = ?', true).group('users.id')
}
scope :verified, -> { where(email_verified: true) }
scope :unverified, -> { where(email_verified: false) }
scope :humans, -> { where(bot: false) }
scope :bots, -> { where(bot: true) }

scope :search_for, lambda { |q|
  where("users.name ilike :first OR
         users.name ilike :other OR
         users.username ilike :first OR
         users.email ilike :first",
        first: "#{q}%", other: "% #{q}%")
}

scope :visible_by, ->(user) {
  distinct.active.verified.joins(:memberships)
    .where("memberships.group_id": user.group_ids)
    .where.not(id: user.id)
}

scope :mention_search, lambda { |model, query|
  # Complex scope for finding mentionable users in context
  # Returns users who are members of model's group, discussion guests, or poll guests
}

scope :email_when_proposal_closing_soon, -> {
  active.where(email_when_proposal_closing_soon: true)
}

scope :email_proposal_closing_soon_for, ->(group) {
  email_when_proposal_closing_soon
    .joins(:memberships)
    .where('memberships.group_id': group.id)
}
```

---

## Callbacks

### Before Validation
- `generate_username` - Auto-generates username from name/email if blank

### Before Save
- `set_legal_accepted_at` - Sets timestamp when legal_accepted is true
- `set_avatar_initials` - Computes initials from name (HasAvatar concern)

### Before Create
- `set_default_avatar_kind` - Sets avatar_kind based on uploaded/gravatar/initials

### After Initialize
- `initialized_with_token :unsubscribe_token` - Generates secure token
- `initialized_with_token :email_api_key` - Generates secure token
- `initialized_with_token :api_key` - Generates secure token
- `initialized_with_token :secret_token` - Generates secure token

---

## Instance Methods

### Authentication & Status

```ruby
def active_for_authentication?
  # Returns false if deactivated (for Devise)
  super && !deactivated_at
end

def has_password
  encrypted_password.present?
end

def is_logged_in?
  true  # Always true for real users (vs LoggedOutUser)
end

def email_status
  deactivated_at.present? ? :inactive : :active
end
```

### Name & Identity

```ruby
def name
  # Returns "Deleted account" translation if deactivated and name nil
  if deactivated_at && self[:name].nil?
    I18n.t('profile_page.deleted_account')
  else
    self[:name]
  end
end

def first_name
  name.to_s.split(' ').first
end

def last_name
  name.split(' ').drop(1).join(' ')
end

def name_and_email
  "\"#{name}\" <#{email}>"
end

def generate_username
  self.username ||= ::UsernameGenerator.new(self).generate
end
```

### Locale & Timezone

```ruby
def locale
  first_supported_locale([selected_locale, detected_locale].compact).to_s
end

def time_zone
  return 'UTC' if self[:time_zone] == "Etc/Unknown"
  self[:time_zone] || 'UTC'
end

def date_time_pref
  self[:date_time_pref] || 'day_abbr'
end

def default_format
  experiences['html-editor.uses-markdown'] ? 'md' : 'html'
end

def update_detected_locale(locale)
  update_attribute(:detected_locale, locale) if detected_locale&.to_s != locale.to_s
end
```

### Membership & Permissions

```ruby
def is_member_of?(group)
  !!memberships.find_by(group_id: group&.id)
end

def is_admin_of?(group)
  !!memberships.find_by(group_id: group&.id, admin: true)
end

def ability
  @ability ||= ::Ability::Base.new(self)
end

delegate :can?, :cannot?, to: :ability

def browseable_group_ids
  Group.where(
    "id in (:group_ids) OR
    (parent_id in (:group_ids) AND is_visible_to_parent_members = TRUE)",
    group_ids: self.group_ids
  ).pluck(:id)
end
```

### Billing

```ruby
def is_paying?
  group_ids = self.group_ids.concat(self.groups.pluck(:parent_id).compact).uniq
  Group.where(id: group_ids)
       .where(parent_id: nil)
       .joins(:subscription)
       .where.not('subscriptions.plan': 'trial')
       .exists?
end

def invitations_rate_limit
  if is_paying?
    ENV.fetch('PAID_INVITATIONS_RATE_LIMIT', 50000)
  else
    ENV.fetch('TRIAL_INVITATIONS_RATE_LIMIT', 500)
  end.to_i
end
```

### Avatar (HasAvatar concern)

```ruby
def avatar_kind
  return 'mdi-duck' if deactivated_at?
  return 'mdi-email-outline' if !name
  read_attribute(:avatar_kind)
end

def avatar_url(size = 512)
  case avatar_kind
  when 'gravatar'
    gravatar_url(size: size, secure: true, default: 'retro')
  when 'uploaded'
    uploaded_avatar_url(size)
  else
    nil
  end
end

def thumb_url
  avatar_url(128)
end
```

---

## Class Methods

```ruby
def self.email_status_for(email)
  find_by(email: email)&.email_status || :unused
end

def self.find_for_database_authentication(warden_conditions)
  # Only finds verified email users for password auth
  super(warden_conditions.merge(email_verified: true))
end

def self.helper_bot
  # Finds or creates the system helper bot user
  verified.find_by(email: BaseMailer::NOTIFICATIONS_EMAIL_ADDRESS) ||
  create!(
    email: BaseMailer::NOTIFICATIONS_EMAIL_ADDRESS,
    name: 'Loomio Helper Bot',
    password: SecureRandom.hex(20),
    email_verified: true,
    bot: true,
    avatar_kind: :gravatar
  )
end
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `CustomCounterCache::Model` | Counter cache updates |
| `ReadableUnguessableUrls` | Generates 8-char `key` field |
| `MessageChannel` | Real-time pub/sub channel |
| `HasExperiences` | Feature flag management |
| `HasAvatar` | Avatar handling (gravatar, uploaded, initials) |
| `SelfReferencing` | `user` and `user_id` methods |
| `NoForbiddenEmails` | Excludes system email addresses |
| `HasRichText` | Rich text bio with sanitization |
| `HasTokens` | Token initialization |
| `HasDefaults` | Default value initialization |
| `NoSpam` | Spam regex validation |

---

## Paper Trail Tracking

Tracked fields:
- `name`
- `username`
- `email`
- `email_newsletter`
- `deactivated_at`
- `deactivator_id`

---

## Related Models

### Identity (omniauth_identities)

OAuth/SSO provider identities linked to users.

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `id` | serial | auto | Primary key |
| `user_id` | integer | - | FK to users |
| `email` | string(255) | - | Provider email |
| `name` | string(255) | - | Provider name |
| `uid` | string(255) | - | Provider user ID |
| `identity_type` | string(255) | - | Provider: google, oauth, saml, nextcloud |
| `access_token` | string | "" | OAuth access token (stored but unused) |
| `logo` | string | - | Provider logo URL |
| `custom_fields` | jsonb | {} | Provider-specific data |

**Validations:**
- `identity_type`: presence required
- `uid`: presence required

**Methods:**
```ruby
def force_user_attrs!
  user.update(name: name, email: email)
end

def assign_logo!
  # Attaches provider logo to user's uploaded_avatar
end
```

### LoginToken

One-time login tokens for passwordless authentication.

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `id` | serial | auto | Primary key |
| `user_id` | integer | - | FK to users |
| `token` | string | - | URL token |
| `code` | integer | - | Numeric code for verification |
| `used` | boolean | false | Already consumed |
| `redirect` | string | - | Post-login redirect URL |
| `is_reactivation` | boolean | false | For reactivating accounts |

**Note:** Tokens expire after 1 hour (cleaned via hourly rake task).

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `email` | UNIQUE | Case-insensitive (citext) |
| `username` | UNIQUE | |
| `key` | UNIQUE | |
| `reset_password_token` | UNIQUE | |
| `unlock_token` | UNIQUE | |
| `unsubscribe_token` | UNIQUE | |
| `api_key` | INDEX | |
| `email_verified` | INDEX | |
| `remember_token` | INDEX | |

---

## Uncertainties

1. **MAX_AVATAR_IMAGE_SIZE_CONST** - Set to 100MB but actual enforcement unclear
2. **attr_accessor fields** - `restricted`, `token`, `membership_token`, `group_token`, `discussion_reader_token`, `stance_token` used for temporary data during operations
3. **require_valid_signup** flag - Exact conditions for when this is set unclear
4. **Pwned password check** - Only enabled in production via `devise :pwned_password`

**Confidence Level:** HIGH for attributes/validations, MEDIUM for some business logic flows.
