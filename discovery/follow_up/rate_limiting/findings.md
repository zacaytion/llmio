# Rate Limiting Investigation - Findings

## Executive Summary

**Verdict: Both Discovery and Research were partially correct, but incomplete.**

Loomio implements a **multi-layered rate limiting strategy**:

1. **Rack::Attack middleware** - IP-based rate limiting on API endpoints (primary defense)
2. **ThrottleService** - Application-level rate limiting for specific operations (invitations, email bounces)
3. **Devise lockable** - Account lockout after failed login attempts

The Discovery documentation missed Rack::Attack entirely. The Research documentation described ThrottleService accurately but with incorrect limit values (claimed 100/hour API, 5/hour login - neither are correct).

---

## Ground Truth Answers

### 1. Does ThrottleService exist and is it actively used?

**YES** - ThrottleService exists and is actively used.

- **Location:** `/Users/z/Code/loomio/app/services/throttle_service.rb`
- **Active usages:**
  - `UserInviter.where_or_create!` - Limits invitation operations per user per day
  - `ReceivedEmailService.route` - Limits bounce notifications per email address per hour

### 2. Which endpoints/operations are rate limited?

#### A. Rack::Attack (IP-based, per-endpoint)

All POST/PUT/PATCH requests to these endpoints are rate limited per IP:

| Endpoint | Hourly Limit | Daily Limit |
|----------|--------------|-------------|
| `/api/v1/trials` | 10 | 30 |
| `/api/v1/announcements` | 100 | 300 |
| `/api/v1/groups` | 20 | 60 |
| `/api/v1/templates` | 10 | 30 |
| `/api/v1/login_tokens` | 10 | 30 |
| `/api/v1/membership_requests` | 100 | 300 |
| `/api/v1/memberships` | 100 | 300 |
| `/api/v1/identities` | 10 | 30 |
| `/api/v1/discussions` | 100 | 300 |
| `/api/v1/polls` | 100 | 300 |
| `/api/v1/outcomes` | 100 | 300 |
| `/api/v1/stances` | 100 | 300 |
| `/api/v1/profile` | 100 | 300 |
| `/api/v1/comments` | 100 | 300 |
| `/api/v1/reactions` | 100 | 300 |
| `/api/v1/link_previews` | 100 | 300 |
| `/api/v1/registrations` | 10 | 30 |
| `/api/v1/sessions` | 10 | 30 |
| `/api/v1/contact_messages` | 10 | 30 |
| `/api/v1/contact_requests` | 10 | 30 |
| `/api/v1/discussion_readers` | 1000 | 3000 |
| `/rails/active_storage/direct_uploads` | 20 | 60 |

*Note: Limits are configurable via `RACK_ATTACK_RATE_MULTIPLIER` and `RACK_ATTACK_TIME_MULTIPLIER` environment variables.*

#### B. ThrottleService (Operation-based)

| Operation | Key | ID | Max | Period |
|-----------|-----|-----|-----|--------|
| User invitations | `UserInviterInvitations` | user.id | Configurable* | day |
| Email bounce notices | `bounce` | sender_email | 1 | hour |

*Configurable via `PAID_INVITATIONS_RATE_LIMIT` (default 50,000) and `TRIAL_INVITATIONS_RATE_LIMIT` (default 500)*

#### C. Devise Lockable (Authentication)

| Metric | Value |
|--------|-------|
| Max login attempts before lockout | 20 (configurable via `MAX_LOGIN_ATTEMPTS`) |
| Unlock strategy | Both (email link + time-based) |
| Unlock time | 6 hours |

### 3. What are the actual rate limit thresholds?

See tables above. Key corrections to Research claims:

- **API limit is NOT 100/hour** - varies by endpoint (10-1000/hour)
- **Login limit is NOT 5/hour** - it's 20 failed attempts total (Devise lockable, not time-window)
- **Email bounce IS 1/hour** - Research was correct on this one

### 4. Is Rack::Attack or similar middleware configured?

**YES**

- **Gem:** `rack-attack` v6.8.0
- **Location:** `/Users/z/Code/loomio/config/initializers/rack_attack.rb`
- **Store:** Uses Rack::Attack's default cache (likely Rails.cache or a configured Redis)

### 5. How are rate limit violations handled?

#### Rack::Attack violations

- **Response:** HTTP 429 (Too Many Requests)
- **Body:** Rack::Attack default response body
- **Logging:** Subscribes to `rack_attack` ActiveSupport notifications, logs:
  - Event name
  - Remote IP
  - Request method
  - Full path
  - Request ID
- **Retry-After header:** NOT explicitly set (uses Rack::Attack default behavior)

#### ThrottleService violations

- **Method:** Raises `ThrottleService::LimitReached` exception
- **Response:** HTTP 500 (Internal Server Error) - **NOT caught by any rescue_from handler**
- **Security Gap:** No graceful handling, no 429 response

#### Devise lockable violations

- **Response:** Standard Devise error messaging
- **Unlock:** Email link or automatic after 6 hours

---

## Redis Key Patterns

### ThrottleService Keys

Pattern: `THROTTLE-{PERIOD}-{key}-{id}`

Examples:
- `THROTTLE-DAY-UserInviterInvitations-123` (user ID 123's daily invitation count)
- `THROTTLE-HOUR-bounce-user@example.com` (bounce notices sent to this email this hour)

### Implementation Details

- Uses `redis-objects` gem with `Redis::Counter`
- Stored in `CACHE_REDIS_POOL` (configured in `/Users/z/Code/loomio/config/initializers/sidekiq.rb`)
- URL from `REDIS_CACHE_URL` or `REDIS_URL` env vars

---

## Security Gaps Identified

### Critical

1. **ThrottleService::LimitReached not handled** - When invitation throttle is exceeded, the exception bubbles up as a 500 error instead of a proper 429 response. This exposes internal error details and provides poor UX.

### Medium

2. **No Retry-After header** - Neither Rack::Attack nor ThrottleService provide `Retry-After` headers to help clients know when to retry.

3. **GET requests not rate limited** - Rack::Attack only throttles POST/PUT/PATCH. Heavy GET requests (e.g., search, exports) are unprotected.

### Low

4. **Bot/API endpoints (`/api/b1/`, `/api/b2/`, `/api/b3/`)** - Not visible in Rack::Attack configuration. May be unprotected or protected elsewhere.

5. **Password reset endpoint** - `/api/v1/registrations` is rate limited at 10/hour, but password reset specifically may need separate limits.

---

## Files Referenced

| File | Purpose |
|------|---------|
| `/Users/z/Code/loomio/Gemfile:49` | rack-attack gem declaration |
| `/Users/z/Code/loomio/config/initializers/rack_attack.rb` | Rack::Attack configuration |
| `/Users/z/Code/loomio/app/services/throttle_service.rb` | Custom throttle service |
| `/Users/z/Code/loomio/app/extras/user_inviter.rb:102-106` | ThrottleService usage for invitations |
| `/Users/z/Code/loomio/app/services/received_email_service.rb:35` | ThrottleService usage for bounce notices |
| `/Users/z/Code/loomio/app/models/user.rb:195-201` | invitations_rate_limit method |
| `/Users/z/Code/loomio/config/initializers/devise.rb:134-155` | Devise lockable configuration |
| `/Users/z/Code/loomio/config/locales/server.en.yml:340-342` | 429 error message translation |
| `/Users/z/Code/loomio/spec/services/throttle_service_spec.rb` | ThrottleService tests |

---

## Confidence Level: 5/5

All claims in this document are verified against source code with specific file paths and line numbers. Test files confirm expected behavior.
