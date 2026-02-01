# Auth Domain - Models Documentation

**Generated:** 2026-02-01
**Confidence:** 4/5 (High - based on direct source code analysis)

---

## Table of Contents

1. [User](#user)
2. [LoginToken](#logintoken)
3. [Identity](#identity)
4. [LoggedOutUser](#loggedoutuser)
5. [AnonymousUser](#anonymoususer)

---

## User

**File:** `/app/models/user.rb`
**Table:** `users`

### Purpose

The core authentication entity representing a registered user in the system. Integrates with Devise for authentication and supports multiple authentication strategies (password, OAuth, SAML, magic link).

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| id | integer | Primary key |
| email | citext | Unique email address (case-insensitive) |
| encrypted_password | string(128) | Bcrypt-hashed password (optional for passwordless users) |
| reset_password_token | string | Token for Devise password recovery |
| reset_password_sent_at | datetime | Timestamp of last password reset email |
| remember_created_at | datetime | Timestamp for "remember me" cookie creation |
| sign_in_count | integer | Number of successful sign-ins |
| current_sign_in_at | datetime | Timestamp of current session start |
| last_sign_in_at | datetime | Timestamp of previous session |
| current_sign_in_ip | inet | IP address of current session |
| last_sign_in_ip | inet | IP address of previous session |
| name | string(255) | Display name |
| username | string(255) | Unique alphanumeric identifier (auto-generated) |
| email_verified | boolean | Whether email ownership has been confirmed |
| deactivated_at | datetime | Timestamp when user was deactivated (soft delete) |
| deactivator_id | integer | User who performed the deactivation |
| is_admin | boolean | Global admin privileges |
| avatar_kind | string | Avatar display type: initials, gravatar, uploaded |
| avatar_initials | string | Two-letter initials for avatar |
| unsubscribe_token | string | Token for email unsubscribe actions |
| api_key | string | API authentication key for external integrations |
| email_api_key | string | Token for email reply integration |
| secret_token | string | Session security token (invalidated on password change) |
| key | string | Public URL identifier |
| time_zone | string | User's preferred timezone |
| selected_locale | string | Manually selected language |
| detected_locale | string | Automatically detected language |
| experiences | jsonb | Feature tracking flags (hash of experience names to boolean) |
| legal_accepted_at | datetime | Timestamp of terms of service acceptance |
| email_newsletter | boolean | Newsletter subscription preference |
| failed_attempts | integer | Count of failed login attempts (for lockout) |
| unlock_token | string | Token for unlocking locked accounts |
| locked_at | datetime | Timestamp when account was locked |
| bot | boolean | Whether this is a system/bot account |

### Associations

| Association | Type | Description |
|-------------|------|-------------|
| memberships | has_many | Group membership records |
| admin_memberships | has_many | Memberships where user is admin |
| groups | has_many through | Groups user belongs to |
| adminable_groups | has_many through | Groups where user is admin |
| identities | has_many | OAuth/SAML identity links |
| login_tokens | has_many | Magic link tokens |
| stances | has_many | Poll votes (as participant) |
| discussion_readers | has_many | Discussion read tracking |
| notifications | has_many | User notifications |
| comments | has_many | Authored comments |
| events | has_many | Events triggered by user |
| reactions | has_many | Reaction records |

### Validations

- **email**: required, unique, email format, max 200 characters
- **name**: required if `require_valid_signup` is true, max 100 characters
- **username**: unique, alphanumeric only (a-z0-9), max 30 characters, auto-generated
- **legal_accepted**: required if `require_valid_signup` and TERMS_URL is set
- **password**: minimum 8 characters when being set, confirmation must match
- **password (production only)**: checked against pwned password database

### Callbacks

- **before_validation**: `generate_username` - creates unique username from name
- **before_save**: `set_avatar_initials` - calculates initials from name
- **before_save**: `set_legal_accepted_at` - records acceptance timestamp

### Scopes

| Scope | Description |
|-------|-------------|
| active | Users without deactivated_at |
| deactivated | Users with deactivated_at set |
| verified | Users with email_verified = true |
| unverified | Users with email_verified = false |
| admins | Global admins (is_admin = true) |
| humans | Non-bot users |
| bots | Bot users |
| search_for(q) | Name/username/email search |
| visible_by(user) | Users sharing group membership |

### Concerns/Mixins

- `CustomCounterCache::Model` - membership count caching
- `ReadableUnguessableUrls` - generates key for URLs
- `MessageChannel` - real-time messaging support
- `HasExperiences` - feature flag tracking
- `HasAvatar` - avatar management
- `SelfReferencing` - self-association support
- `NoForbiddenEmails` - email validation
- `HasRichText` - rich text bio support
- `LocalesHelper` - locale handling
- `HasTokens` (extended) - token initialization

### Devise Modules

- `database_authenticatable` - password authentication
- `recoverable` - password reset via email
- `registerable` - user registration
- `rememberable` - "remember me" cookies
- `lockable` - account lockout after failed attempts
- `trackable` - sign-in tracking
- `pwned_password` (production only) - breached password check

### Key Methods

| Method | Description |
|--------|-------------|
| ability | Returns CanCanCan ability instance for authorization |
| is_logged_in? | Always returns true (vs LoggedOutUser returns false) |
| has_password | Returns true if encrypted_password is present |
| active_for_authentication? | Returns false if deactivated_at is set |
| remember_me | Always returns true (always remember) |
| find_for_database_authentication | Only finds verified users for password auth |

---

## LoginToken

**File:** `/app/models/login_token.rb`
**Table:** `login_tokens`

### Purpose

Represents a one-time magic link token for passwordless authentication. Sent via email and used to authenticate users without requiring a password.

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| id | integer | Primary key |
| user_id | integer | Associated user (required) |
| token | string | URL-safe token for magic link |
| code | integer | 6-digit numeric code for manual entry |
| used | boolean | Whether token has been consumed |
| redirect | string | URL path to redirect after authentication |
| is_reactivation | boolean | Whether token is for account reactivation |
| created_at | datetime | Creation timestamp (used for expiration) |
| updated_at | datetime | Last update timestamp |

### Associations

| Association | Type | Description |
|-------------|------|-------------|
| user | belongs_to | The user this token authenticates |

### Validations

None explicit - relies on user_id foreign key constraint.

### Callbacks

- **after_initialize**: Generates unique token if not set
- **after_initialize**: Generates 6-digit code if not set

### Scopes

| Scope | Description |
|-------|-------------|
| unused | Tokens where used = false |

### Key Methods

| Method | Description |
|--------|-------------|
| useable? | Returns true if: not used, not expired, user exists |
| expires_at | Returns created_at + EXPIRATION minutes |
| user (override) | Returns verified user with matching email, or original user |

### Configuration

- **EXPIRATION**: Configurable via `LOGIN_TOKEN_EXPIRATION_MINUTES` env var, defaults to 1440 (24 hours)
- **Code generation**: Random 6-digit number between 100000 and 999999

### Token Generation

Uses `HasTokens.initialized_with_token` which calls `User.generate_unique_secure_token` to create cryptographically secure tokens.

---

## Identity

**File:** `/app/models/identity.rb`
**Table:** `omniauth_identities`

### Purpose

Links external OAuth/SAML identity providers to user accounts. Stores the unique identifier (uid) from the external provider along with profile information.

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| id | integer | Primary key |
| user_id | integer | Associated user (nullable for pending identities) |
| identity_type | string | Provider name: oauth, saml, google, nextcloud |
| uid | string | Unique identifier from the provider |
| email | string | Email from the provider |
| name | string | Display name from the provider |
| access_token | string | OAuth access token (optional) |
| logo | string | URL to user's avatar from provider |
| custom_fields | jsonb | Additional provider-specific data |
| created_at | datetime | Creation timestamp |
| updated_at | datetime | Last update timestamp |

### Associations

| Association | Type | Description |
|-------------|------|-------------|
| user | belongs_to | The linked user account (optional) |

### Validations

- **identity_type**: required
- **uid**: required

### Scopes

| Scope | Description |
|-------|-------------|
| with_user | Identities that have a linked user |

### Key Methods

| Method | Description |
|--------|-------------|
| force_user_attrs! | Updates linked user's name and email from identity |
| assign_logo! | Downloads avatar from provider and attaches to user |

### Configuration

Provider types loaded from `/config/providers.yml`:
- oauth (generic OAuth2)
- saml
- google
- nextcloud

### Identity Linking Logic

1. **uid is source of truth**: If identity with matching uid+type exists, update and use existing user
2. **Email matching**: In standard mode, only link to verified users
3. **SSO-only mode** (`FEATURES_DISABLE_EMAIL_LOGIN`): Create user or link to any user by email
4. **Force sync** (`LOOMIO_SSO_FORCE_USER_ATTRS`): Always update user name/email from provider

---

## LoggedOutUser

**File:** `/app/models/logged_out_user.rb`
**Type:** Plain Ruby class (not ActiveRecord)

### Purpose

Null object pattern implementation representing a user who is not authenticated. Provides safe defaults for all user methods, preventing nil checks throughout the codebase.

### Attributes (attr_accessor)

| Attribute | Description |
|-----------|-------------|
| name | Display name (from pending actions or nil) |
| email | Email address (from pending actions or nil) |
| token | Token for pending authentication |
| avatar_initials | Calculated from name/email |
| locale | Preferred locale |
| legal_accepted | Whether terms are accepted |
| time_zone | Timezone (defaults to UTC) |
| date_time_pref | Date/time format preference |
| autodetect_time_zone | Timezone auto-detection flag |

### Constructor Parameters

```
initialize(
  name: nil,
  email: nil,
  token: nil,
  locale: I18n.locale,
  time_zone: 'UTC',
  date_time_pref: 'day_abbr',
  params: {},
  session: {}
)
```

### Null Behaviors

| Method Category | Behavior |
|-----------------|----------|
| nil_methods | Returns nil: id, created_at, avatar_url, persisted?, etc. |
| false_methods | Returns false: is_logged_in?, is_member_of?, has_password, etc. |
| empty_methods | Returns empty array: group_ids, attachments, etc. |
| hash_methods | Returns empty hash: experiences |
| none_methods | Returns empty relation: notifications, memberships, groups, etc. |

### Key Methods

| Method | Description |
|--------|-------------|
| is_logged_in? | Always returns false |
| ability | Returns Ability::Base instance (with limited permissions) |
| can?(action, resource) | Delegates to ability |
| create_user | Creates a real User from stored attributes |
| email_status | Looks up actual email status for stored email |
| group_token | Retrieves from params or session |
| membership_token | Retrieves from params or session |
| stance_token | Retrieves from params |
| discussion_reader_token | Retrieves from params |

### Mixins

- `Null::User` - provides null method definitions
- `AvatarInitials` - generates initials from name

---

## AnonymousUser

**File:** `/app/models/anonymous_user.rb`
**Type:** Plain Ruby class extending LoggedOutUser

### Purpose

Represents an anonymous participant in polls where anonymous voting is enabled. Used to display stance authors without revealing identity.

### Key Methods

| Method | Return Value |
|--------|--------------|
| name | Translated "Anonymous" string |
| username | Symbol :anonymous |
| avatar_kind | "initials" |
| avatar_initials | Person emoji "👤" |

### Usage

Created when displaying stances in anonymous polls. The participant's real identity is stored in the Stance record but displayed using AnonymousUser.

---

## Open Questions

1. **Secret token invalidation**: The secret_token is regenerated on sign-out but the mechanism for invalidating sessions on password change relies on Devise's built-in behavior - needs verification.

2. **Pwned password integration**: Only enabled in production - unclear if there's a development testing mechanism.

3. **Email verification flow**: The transition from email_verified=false to true happens via UserService.verify but the timing of when this is called during various auth flows could be clarified.

---

## Related Files

- `/app/models/concerns/has_tokens.rb` - Token generation
- `/app/models/concerns/null/user.rb` - Null user behavior
- `/app/models/concerns/null/object.rb` - Base null object
- `/app/helpers/pending_actions_helper.rb` - Pending token handling
- `/db/schema.rb` - Database schema definitions
