# Rate Limiting - Final Synthesis

## Executive Summary

Loomio implements a **three-layer rate limiting strategy**:

1. **Rack::Attack Middleware** - IP-based rate limiting on API endpoints (primary defense)
2. **ThrottleService** - Application-level rate limiting for specific operations (invitations, email bounces)
3. **Devise Lockable** - Account lockout after failed login attempts

This document consolidates verified findings for understanding the rate limiting architecture.

---

## Confirmed Architecture

### Layer 1: Rack::Attack (Middleware Level)

**Source:** `orig/loomio/config/initializers/rack_attack.rb`

**Gem Version:** rack-attack 6.8.0

**Scope:** All POST/PUT/PATCH requests to `/api/v1/*` endpoints

#### Rate Limit Configuration

| Endpoint | Hourly Limit | Daily Limit | Category |
|----------|--------------|-------------|----------|
| `/api/v1/sessions` | 10 | 30 | Authentication |
| `/api/v1/registrations` | 10 | 30 | Authentication |
| `/api/v1/login_tokens` | 10 | 30 | Authentication |
| `/api/v1/identities` | 10 | 30 | Authentication |
| `/api/v1/trials` | 10 | 30 | Signups |
| `/api/v1/templates` | 10 | 30 | Signups |
| `/api/v1/contact_messages` | 10 | 30 | Contact |
| `/api/v1/contact_requests` | 10 | 30 | Contact |
| `/api/v1/groups` | 20 | 60 | Groups |
| `/rails/active_storage/direct_uploads` | 20 | 60 | Uploads |
| `/api/v1/announcements` | 100 | 300 | Content |
| `/api/v1/membership_requests` | 100 | 300 | Content |
| `/api/v1/memberships` | 100 | 300 | Content |
| `/api/v1/discussions` | 100 | 300 | Content |
| `/api/v1/polls` | 100 | 300 | Content |
| `/api/v1/outcomes` | 100 | 300 | Content |
| `/api/v1/stances` | 100 | 300 | Content |
| `/api/v1/profile` | 100 | 300 | Content |
| `/api/v1/comments` | 100 | 300 | Content |
| `/api/v1/reactions` | 100 | 300 | Content |
| `/api/v1/link_previews` | 100 | 300 | Content |
| `/api/v1/discussion_readers` | 1000 | 3000 | Read State |

**Limit Calculation:**
- Hourly: `limit * RATE_MULTIPLIER` per `TIME_MULTIPLIER` hours
- Daily: `limit * 3 * RATE_MULTIPLIER` per `TIME_MULTIPLIER` days

**Environment Variables:**
```bash
RACK_ATTACK_RATE_MULTPLIER=1   # Note: typo in original (missing 'I')
RACK_ATTACK_TIME_MULTPLIER=1   # Note: typo in original (missing 'I')
```

#### IP Detection

```ruby
# Priority order for client IP detection
env['HTTP_CF_CONNECTING_IP']     # Cloudflare header (highest priority)
env['action_dispatch.remote_ip']  # Rails ActionDispatch
ip                                # Rack default (fallback)
```

#### Response on Throttle

- **Status:** 429 Too Many Requests
- **Body:** `"Retry later\n"` (default Rack::Attack)
- **Retry-After:** Not set
- **Logging:** ActiveSupport notification with IP, method, path, request_id

---

### Layer 2: ThrottleService (Application Level)

**Source:** `orig/loomio/app/services/throttle_service.rb`

**Purpose:** Fine-grained rate limiting for specific operations with user/entity context.

#### Implementation

```ruby
module ThrottleService
  class LimitReached < StandardError
  end

  def self.can?(key: 'default-key', id: 1, max: 100, inc: 1, per: 'hour')
    raise "Throttle per is not hour or day: #{per}" unless ['hour', 'day'].include? per.to_s
    k = "THROTTLE-#{per.upcase}-#{key}-#{id}"
    Redis::Counter.new(k).increment(inc)
    Redis::Counter.new(k).value <= ENV.fetch('THROTTLE_MAX_'+key, max)
  end

  def self.limit!(key: 'default-key', id: 1, max: 100, inc: 1, per: 'hour')
    if can?(key: key, id: id, max: max, inc: inc, per: per)
      return true
    else
      raise ThrottleService::LimitReached.new "Throttled! #{key}-#{id}"
    end
  end

  def self.reset!(per)
    CACHE_REDIS_POOL.with do |client|
      client.scan_each(match: "THROTTLE-#{per.upcase}*") { |key| client.del(key) }
    end
  end
end
```

#### API Methods

| Method | Behavior | Returns | Throws |
|--------|----------|---------|--------|
| `can?(...)` | Non-blocking check, increments counter | `true/false` | - |
| `limit!(...)` | Blocking check, increments counter | `true` | `LimitReached` |
| `reset!(per)` | Clears all counters for period | - | - |

#### Current Usages

| Operation | Key | ID | Max | Period | Source |
|-----------|-----|-----|-----|--------|--------|
| User invitations | `UserInviterInvitations` | `actor.id` | Dynamic* | day | `app/extras/user_inviter.rb:102-106` |
| Email bounce | `bounce` | `sender_email.downcase` | 1 | hour | `app/services/received_email_service.rb:35` |

*Dynamic limit based on user type:
- Paying users: `ENV['PAID_INVITATIONS_RATE_LIMIT']` (default: 50,000/day)
- Trial users: `ENV['TRIAL_INVITATIONS_RATE_LIMIT']` (default: 500/day)

**Source:** `orig/loomio/app/models/user.rb:195-200`

---

### Layer 3: Devise Lockable (Authentication Level)

**Source:** `orig/loomio/config/initializers/devise.rb:138-155`

**Configuration:**

```ruby
config.lock_strategy = :failed_attempts
config.unlock_keys = [:email]
config.unlock_strategy = :both
config.maximum_attempts = ENV.fetch('MAX_LOGIN_ATTEMPTS', 20).to_i
config.unlock_in = 6.hours
```

| Setting | Value | Description |
|---------|-------|-------------|
| Lock strategy | `:failed_attempts` | Lock after N failed logins |
| Maximum attempts | 20 (default) | Failed logins before lockout |
| Unlock strategy | `:both` | Email link OR time-based |
| Unlock time | 6 hours | Auto-unlock after this period |

**Database Columns on User:**
- `failed_attempts` (integer)
- `locked_at` (timestamp)
- `unlock_token` (string)

---

## Redis Key Patterns

### ThrottleService Keys

**Pattern:** `THROTTLE-{PERIOD}-{key}-{id}`

**Examples:**
```
THROTTLE-DAY-UserInviterInvitations-42        # User 42's daily invitation count
THROTTLE-HOUR-bounce-user@example.com         # Bounce notices for this email
```

**Data Type:** Redis String (via redis-objects `Redis::Counter`)

### Redis Configuration

**Source:** `orig/loomio/config/initializers/sidekiq.rb`

```ruby
channels_redis_url = (ENV['REDIS_CACHE_URL'] || ENV.fetch('REDIS_URL', 'redis://localhost:6379/0'))
CACHE_REDIS_POOL = ConnectionPool.new(size: ENV.fetch('REDIS_POOL_SIZE', 30).to_i, timeout: 5) {
  Redis.new(url: channels_redis_url)
}
```

**Environment Variables:**
```bash
REDIS_URL=redis://localhost:6379/0      # Primary Redis URL
REDIS_CACHE_URL=                         # Optional separate cache Redis
REDIS_POOL_SIZE=30                       # Connection pool size
```

---

## Environment Variables Summary

| Variable | Default | Purpose |
|----------|---------|---------|
| `RACK_ATTACK_RATE_MULTIPLIER` | 1 | Scale all rate limits |
| `RACK_ATTACK_TIME_MULTIPLIER` | 1 | Scale time windows |
| `REDIS_URL` | `redis://localhost:6379/0` | Primary Redis connection |
| `REDIS_CACHE_URL` | (REDIS_URL) | Separate cache Redis |
| `REDIS_POOL_SIZE` | 30 | Redis connection pool size |
| `THROTTLE_MAX_{KEY}` | varies | Override specific throttle limits |
| `PAID_INVITATIONS_RATE_LIMIT` | 50000 | Daily invitations for paying users |
| `TRIAL_INVITATIONS_RATE_LIMIT` | 500 | Daily invitations for trial users |
| `MAX_LOGIN_ATTEMPTS` | 20 | Failed logins before lockout |

---

## Security Considerations

### Known Gaps in Loomio (to avoid)

1. **ThrottleService::LimitReached returns 500** - Should catch and return 429
2. **No Retry-After header** - Should include this header
3. **GET requests not rate limited** - Consider adding limits for search endpoints
4. **No per-user rate limiting** - Consider adding for authenticated endpoints
5. **Redis keys may not expire** - Implementation MUST set TTL

### Recommended Improvements

1. Add `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers
2. Rate limit GET requests on heavy endpoints (`/search`, `/boot`)
3. Implement per-user rate limiting in addition to per-IP
4. Add metrics/monitoring for rate limit hits
5. Consider Allow2Ban pattern for repeated offenders

---

## File References

| Component | Source File | Lines |
|-----------|-------------|-------|
| Rack::Attack config | `orig/loomio/config/initializers/rack_attack.rb` | 1-63 |
| ThrottleService | `orig/loomio/app/services/throttle_service.rb` | 1-25 |
| User invitation throttle | `orig/loomio/app/extras/user_inviter.rb` | 102-106 |
| Invitation rate limits | `orig/loomio/app/models/user.rb` | 195-200 |
| Email bounce throttle | `orig/loomio/app/services/received_email_service.rb` | 35-41 |
| Devise lockable config | `orig/loomio/config/initializers/devise.rb` | 136-155 |
| Redis pool config | `orig/loomio/config/initializers/sidekiq.rb` | 4-6 |
| rack-attack gem | `orig/loomio/Gemfile` | 49 |
| redis-objects gem | `orig/loomio/Gemfile` | 51 |
| ThrottleService tests | `orig/loomio/spec/services/throttle_service_spec.rb` | 1-57 |
