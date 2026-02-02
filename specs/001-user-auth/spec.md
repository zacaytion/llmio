# Feature Specification: User Authentication

**Feature Branch**: `001-user-auth`
**Created**: 2026-02-01
**Status**: Draft
**Input**: User description: "User authentication and session management - email/password registration, login, logout with in-memory sessions and Argon2id password hashing"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - User Registration (Priority: P1)

A new user wants to create an account so they can participate in group discussions and decision-making.

**Why this priority**: Without registration, no users can exist in the system. This is the foundational entry point for all other features.

**Independent Test**: Can be fully tested by submitting registration form with valid data and verifying user record is created. Delivers immediate value: users can create accounts.

**Acceptance Scenarios**:

1. **Given** a visitor on the registration page, **When** they submit valid email, name, and matching passwords (8+ characters), **Then** a new account is created and they see a confirmation message.

2. **Given** a visitor on the registration page, **When** they submit an email that already exists, **Then** they see an error message indicating the email is taken.

3. **Given** a visitor on the registration page, **When** they submit passwords that don't match, **Then** they see an error message about password mismatch.

4. **Given** a visitor on the registration page, **When** they submit a password shorter than 8 characters, **Then** they see an error message about minimum password length.

5. **Given** a visitor on the registration page, **When** they submit without a name, **Then** they see an error message that name is required.

---

### User Story 2 - User Login (Priority: P1)

A registered user wants to log in to access their groups, discussions, and polls.

**Why this priority**: Equal to registration - users must be able to access their accounts. Login enables all authenticated features.

**Independent Test**: Can be fully tested by logging in with valid credentials and verifying session is active. Delivers value: returning users can access their data.

**Acceptance Scenarios**:

1. **Given** a registered user with verified email, **When** they submit correct email and password, **Then** they are logged in and redirected to the application.

2. **Given** a registered user, **When** they submit incorrect password, **Then** they see an error message about invalid credentials (without revealing which field was wrong).

3. **Given** an email that doesn't exist, **When** someone attempts to log in, **Then** they see the same generic invalid credentials error (no account enumeration).

4. **Given** a user whose email is not verified, **When** they attempt to log in, **Then** they see a message indicating email verification is required.

5. **Given** a deactivated user account, **When** they attempt to log in, **Then** they see a message that their account is not active.

---

### User Story 3 - User Logout (Priority: P2)

A logged-in user wants to log out to secure their session, especially on shared devices.

**Why this priority**: Important for security but depends on login being implemented first.

**Independent Test**: Can be fully tested by clicking logout and verifying session is terminated. Delivers value: users can secure their accounts.

**Acceptance Scenarios**:

1. **Given** a logged-in user, **When** they click logout, **Then** their session ends and they are redirected to the login page.

2. **Given** a logged-in user who has logged out, **When** they try to access a protected page, **Then** they are redirected to login.

3. **Given** a user with an expired session (7 days), **When** they try to access the application, **Then** they are redirected to login.

---

### User Story 4 - Session Persistence (Priority: P2)

A user wants their session to persist across browser tabs and page refreshes within the session duration.

**Why this priority**: Essential for usability but secondary to core login/logout flow.

**Independent Test**: Can be tested by logging in, opening new tabs, and verifying session remains active. Delivers value: seamless multi-tab experience.

**Acceptance Scenarios**:

1. **Given** a logged-in user, **When** they open the application in a new browser tab, **Then** they remain logged in.

2. **Given** a logged-in user, **When** they refresh the page, **Then** they remain logged in.

3. **Given** a logged-in user, **When** they close and reopen the browser within 7 days, **Then** they remain logged in.

---

### Edge Cases

- What happens when a user registers with mixed-case email (e.g., "User@Example.COM")? System treats email as case-insensitive.
- What happens when a user's session expires mid-action? User is redirected to login with option to continue after re-authentication.
- What happens when the same user logs in from multiple devices? All sessions remain valid independently.
- What happens when server restarts? All sessions are cleared; users must log in again.
- What happens when a user submits login form with empty fields? Validation errors shown for required fields.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow visitors to create accounts with email, name, and password.
- **FR-002**: System MUST validate email format and uniqueness (case-insensitive).
- **FR-003**: System MUST require passwords of at least 8 characters.
- **FR-004**: System MUST require password confirmation during registration.
- **FR-005**: System MUST require a name during registration.
- **FR-006**: System MUST securely hash passwords before storage (passwords never stored in plaintext).
- **FR-007**: System MUST allow registered users with verified emails to log in.
- **FR-008**: System MUST create a session upon successful login.
- **FR-009**: System MUST set session duration to 7 days from login time.
- **FR-010**: System MUST allow logged-in users to log out, terminating their session.
- **FR-011**: System MUST prevent account enumeration (same error for wrong email vs wrong password).
- **FR-012**: System MUST reject login attempts for unverified email addresses.
- **FR-013**: System MUST reject login attempts for deactivated accounts.
- **FR-014**: System MUST generate a unique username from name/email during registration.
- **FR-015**: System MUST generate a unique public URL key for each user.
- **FR-016**: System MUST persist session across browser tabs and page refreshes.

### Key Entities

- **User**: A person who has registered an account. Key attributes: email (unique, case-insensitive), name, username (unique, auto-generated), email verification status, account status (active/deactivated), registration date.

- **Session**: An authenticated user's active login state. Key attributes: associated user, creation time, expiration time (7 days from creation), originating device info (user agent, IP address).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete registration in under 60 seconds.
- **SC-002**: Users can log in within 5 seconds of submitting credentials.
- **SC-003**: 95% of registration attempts with valid data succeed on first try.
- **SC-004**: Zero passwords stored in plaintext or reversible encryption.
- **SC-005**: Session remains active for the full 7-day duration without requiring re-login.
- **SC-006**: Invalid login attempts receive response within 3 seconds (prevents timing-based enumeration).
- **SC-007**: System handles 100 concurrent login attempts without degradation.

## Assumptions

- Email verification flow is out of scope for this feature; users will be manually marked as verified for MVP.
- Password reset functionality is out of scope; will be a separate feature.
- OAuth/SSO authentication is out of scope; will be added in a future feature.
- "Remember me" functionality is out of scope; fixed 7-day session duration only.
- Account lockout after failed attempts is out of scope for MVP.
- Sessions are stored in memory and will be lost on server restart (acceptable for MVP).

## Out of Scope

- Email verification sending/flow
- Password reset/recovery
- OAuth/SAML/SSO authentication
- Multi-factor authentication
- Account lockout policies
- "Remember me" / extended sessions
- Session management UI (view/revoke other sessions)
- Password strength meter
- Breach detection (HaveIBeenPwned integration)

## Dependencies

- User database table must exist with required fields
- Frontend login and registration forms (separate feature)
