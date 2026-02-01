# Rate Limiting - Final Synthesis

## Executive Summary

Loomio implements a **three-layer rate limiting strategy**:

1. **Rack::Attack Middleware** - IP-based rate limiting on API endpoints (primary defense)
2. **ThrottleService** - Application-level rate limiting for specific operations (invitations, email bounces)
3. **Devise Lockable** - Account lockout after failed login attempts

This document consolidates verified findings for implementation in the Go rewrite.

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

## Go Implementation Specification

### Middleware Configuration

```go
package ratelimit

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/httprate"
    "github.com/redis/go-redis/v9"
)

// EndpointLimit defines rate limits for an API endpoint
type EndpointLimit struct {
    Path       string
    HourLimit  int
    DayLimit   int
}

// DefaultLimits matches Loomio's Rack::Attack configuration
var DefaultLimits = []EndpointLimit{
    // Authentication (strict)
    {"/api/v1/sessions", 10, 30},
    {"/api/v1/registrations", 10, 30},
    {"/api/v1/login_tokens", 10, 30},
    {"/api/v1/identities", 10, 30},

    // Signups/Contact (strict)
    {"/api/v1/trials", 10, 30},
    {"/api/v1/templates", 10, 30},
    {"/api/v1/contact_messages", 10, 30},
    {"/api/v1/contact_requests", 10, 30},

    // Groups/Uploads (moderate)
    {"/api/v1/groups", 20, 60},

    // Content (standard)
    {"/api/v1/announcements", 100, 300},
    {"/api/v1/membership_requests", 100, 300},
    {"/api/v1/memberships", 100, 300},
    {"/api/v1/discussions", 100, 300},
    {"/api/v1/polls", 100, 300},
    {"/api/v1/outcomes", 100, 300},
    {"/api/v1/stances", 100, 300},
    {"/api/v1/profile", 100, 300},
    {"/api/v1/comments", 100, 300},
    {"/api/v1/reactions", 100, 300},
    {"/api/v1/link_previews", 100, 300},

    // High-frequency (read state)
    {"/api/v1/discussion_readers", 1000, 3000},
}

// Config for rate limiting
type Config struct {
    RateMultiplier int
    TimeMultiplier int
    RedisURL       string
}

func DefaultConfig() Config {
    return Config{
        RateMultiplier: envInt("RACK_ATTACK_RATE_MULTIPLIER", 1),
        TimeMultiplier: envInt("RACK_ATTACK_TIME_MULTIPLIER", 1),
        RedisURL:       env("REDIS_URL", "redis://localhost:6379/0"),
    }
}
```

### IP Detection

```go
// GetClientIP extracts client IP with Cloudflare support
func GetClientIP(r *http.Request) string {
    // Priority 1: Cloudflare header
    if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
        return ip
    }

    // Priority 2: X-Forwarded-For (first IP)
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        if idx := strings.Index(xff, ","); idx > 0 {
            return strings.TrimSpace(xff[:idx])
        }
        return strings.TrimSpace(xff)
    }

    // Priority 3: X-Real-IP
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }

    // Fallback: Remote address
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
```

### ThrottleService Equivalent

```go
package throttle

import (
    "context"
    "errors"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
)

var ErrLimitReached = errors.New("rate limit exceeded")

type Service struct {
    redis *redis.Client
}

func NewService(redisClient *redis.Client) *Service {
    return &Service{redis: redisClient}
}

// Can checks and increments the counter, returns true if under limit
func (s *Service) Can(ctx context.Context, key string, id string, max int, inc int, per string) (bool, error) {
    if per != "hour" && per != "day" {
        return false, fmt.Errorf("invalid period: %s (must be 'hour' or 'day')", per)
    }

    redisKey := fmt.Sprintf("THROTTLE-%s-%s-%s", strings.ToUpper(per), key, id)

    // Increment counter
    count, err := s.redis.IncrBy(ctx, redisKey, int64(inc)).Result()
    if err != nil {
        return false, fmt.Errorf("redis incr failed: %w", err)
    }

    // Set TTL on first increment
    if count == int64(inc) {
        ttl := time.Hour
        if per == "day" {
            ttl = 24 * time.Hour
        }
        s.redis.Expire(ctx, redisKey, ttl)
    }

    // Check environment override
    envMax := getEnvInt("THROTTLE_MAX_"+key, max)

    return count <= int64(envMax), nil
}

// Limit checks and increments, returns error if limit exceeded
func (s *Service) Limit(ctx context.Context, key string, id string, max int, inc int, per string) error {
    allowed, err := s.Can(ctx, key, id, max, inc, per)
    if err != nil {
        return err
    }
    if !allowed {
        return fmt.Errorf("%w: %s-%s", ErrLimitReached, key, id)
    }
    return nil
}

// Reset clears all counters for a period
func (s *Service) Reset(ctx context.Context, per string) error {
    pattern := fmt.Sprintf("THROTTLE-%s-*", strings.ToUpper(per))
    iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()
    for iter.Next(ctx) {
        s.redis.Del(ctx, iter.Val())
    }
    return iter.Err()
}

func getEnvInt(key string, defaultVal int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return defaultVal
}
```

### Invitation Rate Limit

```go
// InvitationsRateLimit returns the daily invitation limit for a user
func InvitationsRateLimit(isPaying bool) int {
    if isPaying {
        return getEnvInt("PAID_INVITATIONS_RATE_LIMIT", 50000)
    }
    return getEnvInt("TRIAL_INVITATIONS_RATE_LIMIT", 500)
}
```

### HTTP Response Handler

```go
// RateLimitExceeded returns a proper 429 response
func RateLimitExceeded(w http.ResponseWriter, retryAfter time.Duration) {
    w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusTooManyRequests)
    w.Write([]byte(`{"errors":{"base":["Rate limit exceeded. Please try again later."]}}`))
}

// ThrottleErrorHandler middleware catches throttle.ErrLimitReached
func ThrottleErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                if errors.Is(err.(error), throttle.ErrLimitReached) {
                    RateLimitExceeded(w, time.Hour)
                    return
                }
                panic(err) // re-throw non-throttle panics
            }
        }()
        next.ServeHTTP(w, r)
    })
}
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

### Known Gaps in Loomio (to avoid in Go)

1. **ThrottleService::LimitReached returns 500** - Go must catch and return 429
2. **No Retry-After header** - Go must include this header
3. **GET requests not rate limited** - Consider adding limits for search endpoints
4. **No per-user rate limiting** - Consider adding for authenticated endpoints
5. **Redis keys may not expire** - Go implementation MUST set TTL

### Recommended Improvements for Go

1. Add `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers
2. Rate limit GET requests on heavy endpoints (`/search`, `/boot`)
3. Implement per-user rate limiting in addition to per-IP
4. Add metrics/monitoring for rate limit hits
5. Consider Allow2Ban pattern for repeated offenders

---

## Testing Strategy

### Unit Tests

```go
func TestThrottleService_Can(t *testing.T) {
    // Test basic limiting
    for i := 0; i < 5; i++ {
        ok, _ := svc.Can(ctx, "test", "1", 5, 1, "hour")
        assert.True(t, ok)
    }
    ok, _ := svc.Can(ctx, "test", "1", 5, 1, "hour")
    assert.False(t, ok) // 6th request should fail
}

func TestThrottleService_Limit(t *testing.T) {
    err := svc.Limit(ctx, "test", "1", 1, 1, "hour")
    assert.NoError(t, err)

    err = svc.Limit(ctx, "test", "1", 1, 1, "hour")
    assert.ErrorIs(t, err, throttle.ErrLimitReached)
}

func TestThrottleService_Reset(t *testing.T) {
    svc.Can(ctx, "test", "1", 1, 1, "hour")
    svc.Reset(ctx, "hour")
    ok, _ := svc.Can(ctx, "test", "1", 1, 1, "hour")
    assert.True(t, ok) // Counter should be reset
}
```

### Integration Tests

```go
func TestRateLimitMiddleware(t *testing.T) {
    // Test 10 requests to /api/v1/sessions - should all succeed
    for i := 0; i < 10; i++ {
        resp := POST("/api/v1/sessions", validCredentials)
        assert.NotEqual(t, 429, resp.StatusCode)
    }

    // 11th request should be rate limited
    resp := POST("/api/v1/sessions", validCredentials)
    assert.Equal(t, 429, resp.StatusCode)
    assert.NotEmpty(t, resp.Header.Get("Retry-After"))
}
```

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
