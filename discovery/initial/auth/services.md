# Auth Domain - Services Documentation

**Generated:** 2026-02-01
**Confidence:** 4/5 (High - based on direct source code analysis)

---

## Table of Contents

1. [UserService](#userservice)
2. [LoginTokenService](#logintokenservice)
3. [CurrentUserHelper](#currentuserhelper)
4. [PendingActionsHelper](#pendingactionshelper)

---

## UserService

**File:** `/app/services/user_service.rb`

### Purpose

Handles user lifecycle operations including creation, verification, profile updates, deactivation, and data redaction. All user mutations should flow through this service.

### Public Methods

#### create(params:)

Creates or updates an unverified user account.

**Signature:**
```
def self.create(params:)
  # params: Hash with :name, :email, :legal_accepted, :email_newsletter
  # returns: User object (may have validation errors)
  # raises: UserService::EmailTakenError if verified user exists
```

**Logic:**
1. IF a verified user exists with this email, raise EmailTakenError
2. Find or create an unverified user with this email
3. Set attributes from params
4. Mark as require_valid_signup = true (triggers full validation)
5. Save the user
6. Return user (check errors for validation failures)
7. On unique constraint violation, retry (handles race conditions)

**Triggered by:** RegistrationsController#create

**Side effects:**
- Creates User record if new
- Updates existing unverified User record

---

#### verify(user:)

Marks a user's email as verified.

**Signature:**
```
def self.verify(user:)
  # user: User object to verify
  # returns: User object (verified)
```

**Logic:**
1. IF already verified, return user unchanged
2. IF another verified user exists with same email, return that user instead
3. ELSE set email_verified = true on this user
4. IF email_newsletter is true, enqueue NewsletterService.subscribe via GenericWorker

**Triggered by:** CurrentUserHelper#sign_in (wraps Devise sign_in)

**Side effects:**
- Updates email_verified to true
- May enqueue newsletter subscription worker

---

#### deactivate(user:, actor:)

Soft-deletes a user account.

**Signature:**
```
def self.deactivate(user:, actor:)
  # user: User to deactivate
  # actor: User performing the action
  # raises: CanCan::AccessDenied if unauthorized
```

**Logic:**
1. Authorize actor can :deactivate user
2. Enqueue DeactivateUserWorker with user_id and actor_id

**Triggered by:** ProfileController#deactivate

**Side effects:**
- Enqueues background job (actual deactivation happens async)
- Worker sets deactivated_at, revokes memberships, etc.

---

#### redact(user:, actor:)

Permanently removes all personally identifying information (GDPR deletion).

**Signature:**
```
def self.redact(user:, actor:)
  # user: User to redact
  # actor: User performing the action
  # raises: CanCan::AccessDenied if unauthorized
```

**Logic:**
1. Authorize actor can :redact user
2. Enqueue RedactUserWorker with user_id and actor_id

**Triggered by:** ProfileController#destroy

**Side effects:**
- Worker nullifies: email, name, username, avatar_initials, geographic data
- Worker nullifies: IP addresses, password, reset tokens
- Worker stores email_sha256 hash (for future identification)
- Worker deletes Paper Trail versions for this user

---

#### reactivate(user_id)

Restores a deactivated user account.

**Signature:**
```
def self.reactivate(user_id)
  # user_id: Integer ID of user to reactivate
```

**Logic:**
1. Find user by ID
2. Restore memberships that were revoked at deactivation time
3. Update group membership counts
4. Clear deactivated_at timestamp
5. Reindex user's content in search

**Triggered by:** Admin actions or merge operations

---

#### set_volume(user:, actor:, params:)

Updates a user's default notification volume.

**Signature:**
```
def self.set_volume(user:, actor:, params:)
  # params: { volume: "loud"|"normal"|"quiet"|"mute", apply_to_all: boolean }
  # raises: CanCan::AccessDenied if unauthorized
```

**Logic:**
1. Authorize actor can :update user
2. Update default_membership_volume
3. IF apply_to_all is true, propagate to all memberships, discussion_readers, and stances
4. Broadcast EventBus event

---

#### update(user:, actor:, params:)

Updates user profile attributes.

**Signature:**
```
def self.update(user:, actor:, params:)
  # params: User attributes to update
  # returns: false if validation fails
  # raises: CanCan::AccessDenied if unauthorized
```

**Logic:**
1. Authorize actor can :update user
2. IF LOOMIO_SSO_FORCE_USER_ATTRS is set, remove name/email/username from params
3. Assign attributes and files
4. Return false if invalid
5. Save user
6. Broadcast EventBus event
7. IF name changed, reindex authored content for search

---

#### save_experience(user:, actor:, params:)

Records a user experience flag.

**Signature:**
```
def self.save_experience(user:, actor:, params:)
  # params: { experience: "name", value: true|false, remove_experience: boolean }
  # raises: CanCan::AccessDenied if unauthorized
```

**Logic:**
1. Authorize actor can :update user
2. Extract experience name and value (default true, or nil if remove_experience)
3. Update experiences hash
4. Save user
5. Broadcast EventBus event

---

### Exceptions

- `UserService::EmailTakenError` - raised when creating user with email already verified

---

## LoginTokenService

**File:** `/app/services/login_token_service.rb`

### Purpose

Creates and sends magic link login tokens for passwordless authentication.

### Public Methods

#### create(actor:, uri:)

Creates a login token and emails it to the user.

**Signature:**
```
def self.create(actor:, uri:)
  # actor: User to create token for
  # uri: URI of the page the user came from (for redirect)
  # returns: nil
```

**Logic:**
1. IF actor is not present (LoggedOutUser), return early
2. Create LoginToken with redirect path (only if same host as CANONICAL_HOST)
3. Send login email immediately via UserMailer.login
4. Broadcast EventBus event

**Triggered by:**
- LoginTokensController#create (requesting magic link)
- RegistrationsController#create (when email can't be auto-verified)

**Side effects:**
- Creates LoginToken record
- Sends email immediately (not queued)

**Security notes:**
- Redirect URL only stored if host matches CANONICAL_HOST (prevents open redirect)

---

## CurrentUserHelper

**File:** `/app/helpers/current_user_helper.rb`

### Purpose

Provides the `current_user` method used throughout the application and handles sign-in with verification.

### Key Methods

#### sign_in(user)

Wraps Devise sign_in with additional processing.

**Logic:**
1. Clear cached current_user
2. Call UserService.verify(user:) to mark email as verified
3. Call parent sign_in (Devise)
4. Call handle_pending_actions to process pending tokens
5. Associate user to analytics visit

---

#### current_user

Returns the authenticated user or a LoggedOutUser.

**Logic:**
1. Return cached @current_user if present
2. Return Devise's current_user if authenticated
3. Return new LoggedOutUser with current locale, params, and session

---

#### require_current_user

Guard method for authenticated endpoints.

**Logic:**
1. IF not logged in, respond with 401 error

---

#### deny_spam_users

Prevents known spam patterns from accessing the system.

**Logic:**
1. IF current user's email matches spam regex, raise SpamUserDeniedError

---

#### restricted_user (private)

Provides limited user access via unsubscribe_token.

**Logic:**
1. IF params contains unsubscribe_token, find user and mark as restricted
2. Restricted users have limited permissions (handled by serializers)

---

## PendingActionsHelper

**File:** `/app/helpers/pending_actions_helper.rb`

### Purpose

Manages pending authentication tokens stored in session during sign-up/sign-in flows. Consumes these tokens after successful authentication.

### Overview

When users follow invitation links or magic links, tokens are stored in session. After authentication, these pending actions are consumed to grant access.

### Pending Token Types

| Token Type | Session Key | Description |
|------------|-------------|-------------|
| login_token | pending_login_token | Magic link authentication |
| identity | pending_identity_id | OAuth/SAML identity to link |
| group | pending_group_token | Join group invitation |
| membership | pending_membership_token | Direct membership invitation |
| discussion_reader | pending_discussion_reader_token | Discussion guest access |
| stance | pending_stance_token | Poll guest voting access |

### Key Methods

#### handle_pending_actions(user)

Main entry point called after sign-in.

**Logic:**
1. IF user is logged in:
2. Delete pending_user_id from session
3. Consume pending login token (mark as used)
4. Consume pending identity (link to user)
5. Consume pending group (create membership)
6. Consume pending membership (redeem invitation)
7. Consume pending discussion reader (grant guest access)
8. Consume pending stance (grant voting access)
9. Clear all pending session keys

---

#### consume_pending_login_token

Marks the pending login token as used.

---

#### consume_pending_identity(user)

Links the pending OAuth/SAML identity to the user.

---

#### consume_pending_group(user)

Creates membership in the pending group.

**Logic:**
1. IF pending group exists and user is not already a member
2. Create membership
3. Redeem via MembershipService

---

#### consume_pending_membership(user)

Redeems the pending membership invitation.

---

#### consume_pending_discussion_reader(user)

Grants guest access to discussion via DiscussionReaderService.redeem

---

#### consume_pending_stance(user)

Grants guest poll access via StanceService.redeem

---

### Lookup Methods

Each pending type has a lookup method:

| Method | Returns |
|--------|---------|
| pending_login_token | LoginToken from session[:pending_login_token] |
| pending_identity | Identity from session[:pending_identity_id] |
| pending_group | Group from session[:pending_group_token] |
| pending_membership | Membership.pending from token |
| pending_discussion_reader | DiscussionReader.redeemable from token |
| pending_stance | Stance from token |
| pending_invitation | First of: membership, discussion_reader, or stance |
| serialized_pending_identity | Serialized pending data for frontend |

---

## Related Background Workers

### DeactivateUserWorker

Performs async user deactivation:
- Sets deactivated_at and deactivator_id
- Revokes all memberships
- Updates group membership counts
- Reindexes content

### RedactUserWorker

Performs GDPR data deletion:
- Nullifies all PII fields
- Stores email_sha256 for identification
- Deletes Paper Trail versions
- Clears sensitive tokens

---

## EventBus Events

| Event | Triggered By | Payload |
|-------|--------------|---------|
| session_create | SessionsController#create | user |
| registration_create | RegistrationsController#create | user |
| login_token_create | LoginTokenService#create | token, actor |
| user_update | UserService#update | user, actor, params |
| user_set_volume | UserService#set_volume | user, actor, params |
| user_save_experience | UserService#save_experience | user, actor, params |

---

## Open Questions

1. **Rate limiting on login tokens**: How many tokens can be created per user/timeframe?

2. **Token cleanup**: Are expired/used tokens ever purged from the database?

3. **Session invalidation timing**: When does secret_token regeneration actually invalidate existing sessions?

---

## Related Files

- `/app/controllers/api/v1/sessions_controller.rb` - Sign-in endpoint
- `/app/controllers/api/v1/registrations_controller.rb` - Registration endpoint
- `/app/controllers/api/v1/login_tokens_controller.rb` - Magic link request
- `/app/controllers/login_tokens_controller.rb` - Magic link consumption
- `/app/mailers/user_mailer.rb` - Login email sending
- `/app/workers/deactivate_user_worker.rb` - Deactivation background job
- `/app/workers/redact_user_worker.rb` - GDPR deletion background job
