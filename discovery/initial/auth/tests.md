# Auth Domain - Test Documentation

**Generated:** 2026-02-01
**Confidence:** 4/5 (High - based on direct test file analysis)

---

## Table of Contents

1. [Overview](#overview)
2. [Model Tests](#model-tests)
3. [Service Tests](#service-tests)
4. [Controller Tests](#controller-tests)
5. [Test Gaps and Recommendations](#test-gaps-and-recommendations)

---

## Overview

Auth-related tests are distributed across model specs, service specs, and controller specs. The test suite uses RSpec with FactoryBot for test data.

### Key Test Files

| File | Purpose |
|------|---------|
| `/spec/models/user_spec.rb` | User model validations and behavior |
| `/spec/services/user_service_spec.rb` | User lifecycle operations |
| `/spec/services/login_token_service_spec.rb` | Magic link creation |
| `/spec/controllers/api/v1/sessions_controller_spec.rb` | Sign-in flows |
| `/spec/controllers/api/v1/registrations_controller_spec.rb` | Registration flows |
| `/spec/controllers/api/v1/login_tokens_controller_spec.rb` | Magic link API |
| `/spec/controllers/login_tokens_controller_spec.rb` | Magic link consumption |
| `/spec/controllers/identities/oauth_controller_spec.rb` | OAuth flows |
| `/spec/controllers/identities/saml_controller_spec.rb` | SAML flows |
| `/spec/controllers/memberships_controller_spec.rb` | Invitation redemption |

---

## Model Tests

### User Model (`/spec/models/user_spec.rb`)

**Password Validation:**
- Accepts valid password with matching confirmation
- Rejects mismatched password confirmation
- Requires minimum 8 character password
- Only validates password when being updated (not on load)

**Username Validation:**
- Rejects whitespace in username
- Rejects special characters (?, /, _, -)
- Requires lowercase only
- Generates unique usernames from duplicate names
- Preserves existing valid usernames
- Strips email domain from email-based names
- Converts non-ASCII characters to ASCII
- Limits username radical to 18 characters

**Avatar:**
- Sets avatar_kind to gravatar if user has one
- Sets avatar_kind to initials if no gravatar

**Email:**
- Accepts apostrophes in email addresses

**Associations:**
- User has many groups through memberships
- User has many adminable_groups through admin_memberships
- User has many admin_memberships
- User has authored discussions

**Experience Tracking:**
- Can store user experiences
- Does not affect unset experiences
- Can forget (unset) experiences

**Secret Token:**
- Ensures new users have secret_token
- Generates token for new users

---

## Service Tests

### UserService (`/spec/services/user_service_spec.rb`)

**Deactivation:**
- Deactivates user (sets deactivated_at)
- Does not change email address on deactivation

**Redaction:**
- Deactivates user
- Removes all personally identifying information:
  - Nullifies: email, name, username, avatar_initials
  - Nullifies: country, region, city
  - Nullifies: unlock_token, current_sign_in_ip, last_sign_in_ip
  - Nullifies: encrypted_password, reset_password_token, reset_password_sent_at
  - Nullifies: detected_locale, legal_accepted_at
  - Empties: short_bio, location
  - Sets false: email_newsletter, email_verified
- Deletes Paper Trail versions for user
- Stores email_sha256 hash

**Verification:**
- Sets email_verified true if email is unique
- Returns user unchanged if already verified

**Spam User Deletion:**
- Destroys groups created by spam user
- Destroys the spam user
- Destroys discussions in spam groups
- Destroys spam discussions in innocent groups
- Destroys spam comments in innocent groups

---

### LoginTokenService (`/spec/services/login_token_service_spec.rb`)

**Token Creation:**
- Creates new login token for user
- Sends email to user immediately
- Does nothing if actor is LoggedOutUser
- Stores redirect URI from referrer
- Does not store redirect if host is different (security)

---

## Controller Tests

### Sessions Controller (`/spec/controllers/api/v1/sessions_controller_spec.rb`)

**Password Authentication:**
- Signs in with valid password
- Does not sign in with blank password
- Does not sign in with nil password

**Token Authentication:**
- Signs in user with valid pending token
- Does not sign in with used token
- Does not sign in with expired token (25 hours old)
- Does not sign in with invalid token ID
- Finds verified user to sign in (prefers verified over unverified)
- Signs in unverified user (when only unverified exists)

---

### Registrations Controller (`/spec/controllers/api/v1/registrations_controller_spec.rb`)

**New User Registration:**
- Creates new user
- Sets name and email from params
- Sets legal_accepted_at
- Returns signed_in: false (requires email verification)

**Existing Unverified User:**
- Does not create duplicate user
- Updates existing user with new attributes
- Returns signed_in: false

**Existing Verified User:**
- Returns 422 error
- Returns "Email address is already registered" error

**Signup via Membership Token:**
- Does not create new user (uses invited user)
- Signs in immediately (signed_in: true)
- Sets name and legal_accepted_at

**Signup via Membership with Different Email:**
- Creates new user with different email
- Returns signed_in: false

**Signup via Login Token:**
- Does not create new user
- Signs in immediately (signed_in: true)

**Validation:**
- Requires acceptance of legal terms (422 if missing)

---

### Login Tokens API Controller (`/spec/controllers/api/v1/login_tokens_controller_spec.rb`)

**Token Creation:**
- Creates new login token for existing user
- Returns 200 success
- Updates user's detected_locale from Accept-Language header

**Error Cases:**
- Does not create token if no email provided (returns nothing)
- Does not create token for unknown email (returns 404)

---

### Login Tokens Controller (`/spec/controllers/login_tokens_controller_spec.rb`)

**Token Consumption:**
- Sets session variable (pending_login_token)
- Redirects to dashboard by default
- Redirects to token's redirect URL if set

---

### OAuth Controller (`/spec/controllers/identities/oauth_controller_spec.rb`)

**OAuth Initiation:**
- Redirects to OAuth provider with correct parameters
- Stores back_to in session from params
- Stores referrer as back_to when no param provided

**Standard Mode (Email Login Enabled):**
- Sets pending_identity when user doesn't exist (does not create user)
- Attaches identity to verified user with same email
- Does not attach to unverified user (sets pending_identity)
- Does not overwrite user name by default

**SSO-Only Mode (Email Login Disabled):**
- Creates new verified user when user doesn't exist
- Attaches identity to unverified user
- Links existing user to new identity without creating duplicate

**Existing Identity:**
- Updates identity attributes on subsequent logins
- Signs in existing user

**Already Signed In:**
- Attaches new identity to current user

**Force User Attrs Mode:**
- Overwrites user name and email from SSO
- Syncs attributes on every login

**UID as Source of Truth:**
- Updates identity email if changed in SSO
- Syncs user email if LOOMIO_SSO_FORCE_USER_ATTRS

**Error Cases:**
- Redirects with flash error on user cancel
- Returns 401 when access token not returned
- Returns 401 when profile fetch fails
- Redirects to dashboard when no back_to set

**Identity Destruction:**
- Destroys identity connection
- Redirects to referrer or root
- Returns error when not connected

---

### SAML Controller (`/spec/controllers/identities/saml_controller_spec.rb`)

Mirrors OAuth controller tests with SAML-specific setup:

**SAML Initiation:**
- Redirects to SAML IdP
- Stores back_to in session

**All OAuth scenarios replicated for SAML:**
- User creation in SSO-only mode
- Identity linking to verified users
- Pending identity for unverified users
- Force user attrs mode
- Existing identity updates
- Current user identity attachment
- Error cases (invalid SAML response returns 500)

**Metadata Endpoint:**
- Returns SAML metadata XML
- Sets correct content-type

---

### Memberships Controller (`/spec/controllers/memberships_controller_spec.rb`)

**Invitation Redemption:**
- Redeems pending membership
- Updates membership with accepting user
- Creates invitation accepted event
- Handles multiple pending memberships in same org
- Ignores when user already belongs to group
- Does not redeem already accepted membership

**Group Join:**
- Stores pending_group_token in session
- Redirects to group

**Membership Show:**
- Returns 404 for invalid token
- Redirects to group if membership used
- Sets pending_membership_token in session
- Does not auto-accept membership for logged-out user

---

## Factories Used

| Factory | File | Description |
|---------|------|-------------|
| :user | `/spec/factories.rb` | Verified user with password |
| :unverified_user | `/spec/factories.rb` | User without email verification |
| :login_token | `/spec/factories.rb` | Magic link token |
| :identity | `/spec/factories.rb` | OAuth identity |
| :membership | `/spec/factories.rb` | Accepted group membership |
| :pending_membership | `/spec/factories.rb` | Unaccepted membership |

---

## Test Gaps and Recommendations

### Missing Coverage

1. **Account Lockout:**
   - No tests for failed_attempts increment
   - No tests for account locking after max attempts
   - No tests for unlock flow

2. **Password Recovery:**
   - No specs for Users::PasswordsController
   - No tests for reset_password_token generation
   - No tests for password change flow

3. **Pwned Password Check:**
   - No tests for pwned password validation (production-only feature)
   - Would require test mode configuration

4. **Session Invalidation:**
   - No tests for secret_token regeneration on sign-out
   - No tests for session invalidation on password change

5. **Remember Me:**
   - No explicit tests for remember_me functionality
   - User model always returns true for remember_me

6. **Rate Limiting:**
   - No tests for login attempt rate limiting
   - No tests for login token request rate limiting

7. **Token Expiration:**
   - No test for 24-hour login token expiration handling
   - 25-hour test exists but edge case at 24 hours untested

8. **Email Verification Edge Cases:**
   - No test for email change during verification
   - No test for concurrent verification attempts

9. **OAuth State Parameter:**
   - No CSRF protection tests for OAuth flows
   - OAuth state parameter handling not visible in tests

10. **Identity Cleanup:**
    - No tests for identity destruction when user is deleted
    - No tests for orphaned identity handling

### Recommended Additional Tests

```
# Account lockout
it "increments failed_attempts on failed login"
it "locks account after max_attempts"
it "unlocks account after unlock_in time"
it "sends unlock email"

# Password recovery
it "generates reset_password_token"
it "sends reset password email"
it "resets password with valid token"
it "rejects expired reset token"
it "invalidates token after use"

# Session security
it "invalidates session on password change"
it "regenerates secret_token on sign_out"
it "rejects requests with old secret_token"

# Rate limiting
it "rate limits login attempts"
it "rate limits login token requests"

# Edge cases
it "handles concurrent email verification"
it "handles email change during auth flow"
it "cleans up expired login tokens"
```

---

## Test Environment Notes

1. **Devise mapping**: Tests require `request.env["devise.mapping"] = Devise.mappings[:user]`

2. **Session management**: Tests use `session[:pending_*]` to simulate pending tokens

3. **OAuth stubbing**: OAuth tests use WebMock to stub external HTTP requests

4. **SAML mocking**: SAML tests mock OneLogin::RubySaml classes

5. **Environment variables**: SSO mode tests use `stub_const('ENV', ...)` to simulate configuration

---

## Related Files

- `/spec/spec_helper.rb` - RSpec configuration
- `/spec/rails_helper.rb` - Rails-specific test setup
- `/spec/factories.rb` - FactoryBot definitions
- `/spec/support/` - Test helpers and shared examples
