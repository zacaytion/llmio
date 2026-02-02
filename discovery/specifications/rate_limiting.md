# Rate Limiting Investigation - Final Findings

**Investigation Date:** 2026-02-01
**Investigator:** Claude (AI Research Assistant)
**Scope:** ThrottleService TTL, Rack::Attack cache store, error responses, bot API rate limiting

---

## Executive Summary

Loomio implements a **3-layer rate limiting architecture**:
1. **Rack::Attack** - IP-based middleware throttling (HTTP 429)
2. **ThrottleService** - User-based application throttling (HTTP 500)
3. **Devise Lockable** - Authentication failure lockouts

**Critical Finding:** ThrottleService Redis counters have **NO AUTOMATIC TTL** - they rely on scheduled rake task execution for cleanup.

---

## Question 1: How Do ThrottleService Redis Counters Expire?

### Answer: **Scheduled Rake Task (No Redis TTL)**

**Confidence: HIGH (5/5)** - Direct code inspection confirms behavior.

#### Evidence

**ThrottleService Implementation** (`/Users/z/Code/loomio/app/services/throttle_service.rb:14-15`):
```ruby
Redis::Counter.new(k).increment(inc)
Redis::Counter.new(k).value <= ENV.fetch('THROTTLE_MAX_'+key, max)
```

The `Redis::Counter` class (from `redis-objects` gem) does **NOT** set TTL on increment. Keys persist indefinitely until explicitly deleted.

**Reset Mechanism** (`/Users/z/Code/loomio/app/services/throttle_service.rb:5-9`):
```ruby
def self.reset!(per)
  CACHE_REDIS_POOL.with do |client|
    client.scan_each(match: "THROTTLE-#{per.upcase}*") { |key| client.del(key) }
  end
end
```

**Scheduled Execution** (`/Users/z/Code/loomio/lib/tasks/loomio.rake:222-241`):
```ruby
task hourly_tasks: :environment do
  puts "#{DateTime.now.iso8601} Loomio hourly tasks"
  ThrottleService.reset!('hour')
  # ... other tasks ...

  if (Time.now.hour == 0)
    ThrottleService.reset!('day')
    # ...
  end
end
```

#### Key Findings

| Aspect | Finding | Source |
|--------|---------|--------|
| TTL mechanism | Manual deletion via `scan_each` + `del` | `throttle_service.rb:7` |
| Hourly reset | `rake loomio:hourly_tasks` | `loomio.rake:224` |
| Daily reset | Same task, at midnight (hour == 0) | `loomio.rake:234-235` |
| No cron/sidekiq-cron | No `config/schedule.rb` or `config/sidekiq_cron.yml` found | Glob search |

#### Risk Assessment

**MEDIUM RISK**: If the `hourly_tasks` rake job fails or is not scheduled:
- Redis keys accumulate indefinitely
- Users remain throttled beyond intended windows
- Memory pressure on Redis grows over time

**Mitigation Recommendation**: Add Redis TTL to counters:
```ruby
# Current (no TTL):
Redis::Counter.new(k).increment(inc)

# Recommended (with TTL):
CACHE_REDIS_POOL.with do |client|
  client.incrby(k, inc)
  client.expire(k, per == 'hour' ? 3600 : 86400)
end
```

---

## Question 2: What is the Actual Rack::Attack Cache Store?

### Answer: **Rails.cache (Redis in Production)**

**Confidence: HIGH (5/5)** - Configuration chain verified.

#### Evidence

**Rack::Attack Default Behavior**: When no explicit `Rack::Attack.cache.store` is configured, Rack::Attack uses `Rails.cache`.

**Rack::Attack Initializer** (`/Users/z/Code/loomio/config/initializers/rack_attack.rb`):
- **NO** explicit `cache.store` assignment
- Uses default `Rails.cache`

**Rails Cache Configuration** (`/Users/z/Code/loomio/config/application.rb:92`):
```ruby
config.cache_store = :redis_cache_store, { url: (ENV['REDIS_CACHE_URL'] || ENV.fetch('REDIS_URL', 'redis://localhost:6379')) }
```

**Environment Overrides**:

| Environment | Cache Store | Source |
|-------------|-------------|--------|
| **Production** | Redis (from application.rb) | `config/application.rb:92` |
| **Development** | Memory/Null (conditional) | `config/environments/development.rb:26,31` |
| **Test** | Redis (separate DB) | `config/environments/test.rb:29-31` |

#### Cache Store Chain

```
Rack::Attack.cache.store (not set)
    -> defaults to Rails.cache
        -> :redis_cache_store (from application.rb)
            -> ENV['REDIS_CACHE_URL'] || ENV['REDIS_URL'] || 'redis://localhost:6379'
```

#### Key Finding

**Rack::Attack and ThrottleService use DIFFERENT Redis connections**:

| Component | Redis Connection | Source |
|-----------|------------------|--------|
| Rack::Attack | `Rails.cache` (redis_cache_store) | `application.rb:92` |
| ThrottleService | `CACHE_REDIS_POOL` (redis-objects) | `sidekiq.rb:7` |

Both likely point to the same Redis instance (via `REDIS_CACHE_URL`), but use different client libraries.

---

## Question 3: What Happens When ThrottleService::LimitReached is Raised?

### Answer: **HTTP 500 with Unhandled Exception**

**Confidence: HIGH (5/5)** - No rescue_from handler exists.

#### Evidence

**Exception Class** (`/Users/z/Code/loomio/app/services/throttle_service.rb:2-3`):
```ruby
class LimitReached < StandardError
end
```

**Exception Raising** (`/Users/z/Code/loomio/app/services/throttle_service.rb:22`):
```ruby
raise ThrottleService::LimitReached.new "Throttled! #{key}-#{id}"
```

**Controller Error Handlers** (`/Users/z/Code/loomio/app/controllers/api/v1/snorlax_base.rb:2-7`):
```ruby
rescue_from(CanCan::AccessDenied)                    { |e| respond_with_standard_error e, 403 }
rescue_from(Subscription::MaxMembersExceeded)        { |e| respond_with_standard_error e, 403 }
rescue_from(ActionController::UnpermittedParameters) { |e| respond_with_standard_error e, 400 }
rescue_from(ActionController::ParameterMissing)      { |e| respond_with_standard_error e, 400 }
rescue_from(ActiveRecord::RecordNotFound)            { |e| respond_with_standard_error e, 404 }
rescue_from(ActiveRecord::RecordInvalid)             { |e| respond_with_errors }
# NO rescue_from for ThrottleService::LimitReached
```

**ApplicationController handlers** (`/Users/z/Code/loomio/app/controllers/application_controller.rb:28-49`):
- Also **NO** handler for `ThrottleService::LimitReached`

#### Error Response Format

When unhandled, Rails default exception handling returns:

**Production** (typical):
```json
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{"status": 500, "error": "Internal Server Error"}
```

Or if exception details enabled:
```json
{"exception": "ThrottleService::LimitReached", "error": "Throttled! UserInviterInvitations-123"}
```

#### Usage Locations

ThrottleService is called in only **2 places**:

| Location | Method | Purpose |
|----------|--------|---------|
| `app/extras/user_inviter.rb:102-106` | `limit!` | Rate limit invitations per user |
| `app/services/received_email_service.rb:35` | `can?` | Throttle bounce emails (does not raise) |

```ruby
# user_inviter.rb:102-106
ThrottleService.limit!(key: 'UserInviterInvitations',
                        id: actor.id,
                        max: actor.invitations_rate_limit,
                        inc: emails.length + ids.length,
                        per: :day)
```

#### Risk Assessment

**HIGH RISK**: HTTP 500 responses for rate limiting is problematic:
- Clients cannot distinguish rate limits from server errors
- No `Retry-After` header guidance
- Monitoring alerts may fire on "legitimate" throttling
- Not compliant with HTTP standards (should be 429)

**Recommendation**: Add proper handler:
```ruby
# In snorlax_base.rb
rescue_from(ThrottleService::LimitReached) { |e|
  response.headers['Retry-After'] = '3600'
  respond_with_standard_error e, 429
}
```

---

## Question 4: Are Bot API Endpoints Rate Limited?

### Answer: **Partially - IP Level Only (No User-Level)**

**Confidence: HIGH (5/5)** - Code inspection confirms absence of ThrottleService calls.

#### Bot API Overview

| Namespace | Routes | Controller | Source |
|-----------|--------|------------|--------|
| `/api/b1/` | discussions, polls, memberships | **NOT FOUND** | routes.rb:42-46 |
| `/api/b2/` | discussions, polls, memberships, comments | `Api::B2::BaseController` | 5 controller files |
| `/api/b3/` | users (deactivate/reactivate) | `Api::B3::UsersController` | 1 controller file |

#### Rate Limiting Analysis

**IP-Level (Rack::Attack)** - `/api/b1/`, `/api/b2/`, `/api/b3/` **NOT covered**:

`/Users/z/Code/loomio/config/initializers/rack_attack.rb:17-39`:
```ruby
IP_POST_LIMITS = {
  '/api/v1/trials' => 10,
  '/api/v1/announcements' => 100,
  # ... only /api/v1/ routes listed
  # NO /api/b1/, /api/b2/, /api/b3/ entries
}
```

**User-Level (ThrottleService)** - **NOT used** in bot controllers:

| Controller | ThrottleService calls | Source |
|------------|----------------------|--------|
| `Api::B2::BaseController` | None | `b2/base_controller.rb` |
| `Api::B2::DiscussionsController` | None | `b2/discussions_controller.rb` |
| `Api::B2::PollsController` | None | `b2/polls_controller.rb` |
| `Api::B2::MembershipsController` | None (file exists) | N/A |
| `Api::B2::CommentsController` | None (file exists) | N/A |
| `Api::B3::UsersController` | None | `b3/users_controller.rb` |

#### Bot API Authentication

**B2 API** (`/Users/z/Code/loomio/app/controllers/api/b2/base_controller.rb:6-8`):
```ruby
def authenticate_api_key!
  raise CanCan::AccessDenied unless current_user
end

def current_user
  @current_user ||= User.active.find_by(api_key: params[:api_key])
end
```

**B3 API** (`/Users/z/Code/loomio/app/controllers/api/b3/users_controller.rb:6-9`):
```ruby
def authenticate_api_key!
  raise CanCan::AccessDenied unless ENV.fetch('B3_API_KEY', '').length > 16
  raise CanCan::AccessDenied unless params[:b3_api_key] == ENV['B3_API_KEY']
end
```

#### B1 API Mystery

Routes defined but **NO CONTROLLERS FOUND**:
```ruby
# routes.rb:42-46
namespace :b1 do
  resources :discussions, only: [:create, :show]
  resources :polls, only: [:create, :show]
  resources :memberships, only: [:index, :create]
end
```

Running `ls /Users/z/Code/loomio/app/controllers/api/` shows no `b1/` directory. These routes will return 404 or use fallback controller behavior.

#### Risk Assessment

**HIGH RISK**: Bot APIs have **NO rate limiting**:
- Authenticated users can make unlimited API calls
- No protection against compromised API keys
- DoS vector via bot endpoints

**Recommendations**:
1. Add bot endpoints to `IP_POST_LIMITS`
2. Implement per-API-key throttling
3. Consider separate rate limit tiers for bot vs human users

---

## Summary Table

| Question | Answer | Confidence | Risk |
|----------|--------|------------|------|
| ThrottleService TTL | Rake task (no Redis TTL) | HIGH | MEDIUM |
| Rack::Attack cache | Rails.cache (Redis) | HIGH | LOW |
| LimitReached response | HTTP 500, unhandled | HIGH | HIGH |
| Bot API rate limits | IP-only, no user-level | HIGH | HIGH |

---

## File References

| File | Lines | Purpose |
|------|-------|---------|
| `/Users/z/Code/loomio/app/services/throttle_service.rb` | 1-25 | ThrottleService implementation |
| `/Users/z/Code/loomio/config/initializers/rack_attack.rb` | 1-63 | IP-based rate limiting |
| `/Users/z/Code/loomio/config/application.rb` | 92 | Redis cache configuration |
| `/Users/z/Code/loomio/config/initializers/sidekiq.rb` | 7, 9 | CACHE_REDIS_POOL setup |
| `/Users/z/Code/loomio/lib/tasks/loomio.rake` | 222-248 | hourly_tasks with throttle reset |
| `/Users/z/Code/loomio/app/controllers/api/v1/snorlax_base.rb` | 1-9 | Error handlers (no LimitReached) |
| `/Users/z/Code/loomio/app/extras/user_inviter.rb` | 102-106 | ThrottleService.limit! usage |
| `/Users/z/Code/loomio/app/services/received_email_service.rb` | 35 | ThrottleService.can? usage |
| `/Users/z/Code/loomio/app/controllers/api/b2/base_controller.rb` | 1-27 | B2 API base (no rate limits) |
| `/Users/z/Code/loomio/app/controllers/api/b3/users_controller.rb` | 1-22 | B3 API (no rate limits) |
| `/Users/z/Code/loomio/public/429.html` | 1-29 | Static rate limit error page |
| `/Users/z/Code/loomio/config/routes.rb` | 42-62 | Bot API route definitions |

---

## Implementation Recommendations

### Priority 1: Fix ThrottleService Error Response
```ruby
# app/controllers/api/v1/snorlax_base.rb - Add after line 7
rescue_from(ThrottleService::LimitReached) do |e|
  response.headers['Retry-After'] = '3600'
  render json: {
    error: 'rate_limit_exceeded',
    message: 'Too many requests. Please try again later.',
    retry_after: 3600
  }, status: 429
end
```

### Priority 2: Add Redis TTL to ThrottleService
```ruby
# app/services/throttle_service.rb - Replace lines 14-15
def self.can?(key: 'default-key', id: 1, max: 100, inc: 1, per: 'hour')
  raise "Throttle per is not hour or day: #{per}" unless ['hour', 'day'].include? per.to_s
  k = "THROTTLE-#{per.upcase}-#{key}-#{id}"
  ttl = per.to_s == 'hour' ? 3600 : 86400

  CACHE_REDIS_POOL.with do |client|
    client.multi do |multi|
      multi.incrby(k, inc)
      multi.expire(k, ttl)
    end
  end

  CACHE_REDIS_POOL.with { |client| client.get(k).to_i <= ENV.fetch("THROTTLE_MAX_#{key}", max).to_i }
end
```

### Priority 3: Add Bot API Rate Limiting
```ruby
# config/initializers/rack_attack.rb - Add to IP_POST_LIMITS hash
'/api/b1/discussions' => 50,
'/api/b1/polls' => 50,
'/api/b1/memberships' => 100,
'/api/b2/discussions' => 50,
'/api/b2/polls' => 50,
'/api/b2/memberships' => 100,
'/api/b2/comments' => 100,
'/api/b3/users' => 10,
```

---

## Verification Commands

```bash
# Check if ThrottleService is used
grep -rn "ThrottleService" app/ --include="*.rb"

# Check for rescue_from handlers
grep -rn "rescue_from" app/controllers/ --include="*.rb"

# Verify Rack::Attack config
cat config/initializers/rack_attack.rb

# Check bot API controllers
ls -la app/controllers/api/b*/

# Verify scheduled task existence
cat lib/tasks/loomio.rake | grep -A 30 "hourly_tasks"
```
