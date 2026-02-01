# Rate Limiting - Follow-up Investigation Brief

## Discrepancy Summary

Discovery and Research documentation **contradict each other** on rate limiting implementation:
- Discovery claims rate limiting is absent beyond Devise lockable
- Research claims ThrottleService implements 100/hour API limit and 5/hour login

This is a significant discrepancy that affects security posture understanding.

## Discovery Claims

**Source**: `discovery/initial/synthesis/uncertainties.md`

> "HIGH priority: Rate limiting unclear - relies on Devise lockable, no visible broader rate limiting"

**Source**: `discovery/initial/auth/services.md`

> "No visible application-level rate limiting beyond Devise's lockable module for login attempts"

## Our Research Claims

**Source**: `research/investigation/api.md`

> "Rate Limiting: Redis-based, key pattern `THROTTLE-{HOUR|DAY}-{key}-{id}`, default 100/hour"
> "Email bounces: 1/hour, Login attempts: 5/hour"

Research specifically documents a ThrottleService with:
- Redis-based throttling
- Configurable time windows (hour/day)
- Different limits for different operations
- Key pattern for Redis storage

## Ground Truth Needed

1. Does ThrottleService exist and is it actively used?
2. Which endpoints/operations are rate limited?
3. What are the actual rate limit thresholds?
4. Is Rack::Attack or similar middleware configured?
5. How are rate limit violations handled (429 response, retry-after header)?

## Investigation Targets

- [ ] File: `orig/loomio/app/services/throttle_service.rb` - Verify service exists and implementation
- [ ] File: `orig/loomio/config/initializers/rack_attack.rb` - Check for Rack::Attack configuration
- [ ] Command: `grep -r "ThrottleService" orig/loomio/app/` - Find all usages
- [ ] Command: `grep -r "THROTTLE" orig/loomio/` - Find Redis key pattern usage
- [ ] Command: `grep -r "rate_limit\|throttle" orig/loomio/app/controllers/` - Find controller-level rate limiting
- [ ] File: `orig/loomio/Gemfile` - Check for rack-attack or throttle gems

## Priority

**HIGH** - Rate limiting is critical for:
- Preventing brute force attacks on authentication
- Protecting against API abuse
- Ensuring service availability

## Rails Context (from Rack::Attack Documentation)

### Rack::Attack Pattern

Rack::Attack is the standard Rails middleware for blocking and throttling abusive requests. From official documentation:

**Login Throttling by IP:**
```ruby
# Limits POST requests to '/login' to 5 per 20 seconds per IP
Rack::Attack.throttle('logins/ip', limit: 5, period: 20.seconds) do |req|
  if req.path == '/login' && req.post?
    req.ip
  end
end
```

**Login Throttling by Email:**
```ruby
# Limits POST requests to '/login' to 6 per 60 seconds per email
Rack::Attack.throttle('limit logins per email', limit: 6, period: 60) do |req|
  if req.path == '/login' && req.post?
    # Normalize email to prevent bypass attacks
    req.params['email'].to_s.downcase.gsub(/\s+/, "")
  end
end
```

**Blocking Login Scrapers (Allow2Ban):**
```ruby
# After 20 requests in 1 minute, block IP for 1 hour
Rack::Attack.blocklist('allow2ban login scrapers') do |req|
  Rack::Attack::Allow2Ban.filter(req.ip, maxretry: 20, findtime: 1.minute, bantime: 1.hour) do
    req.path == '/login' and req.post?
  end
end
```

**Redis Cache Store (required for production):**
```ruby
# Use separate Redis database for throttling
Rack::Attack.cache.store = ActiveSupport::Cache::RedisCacheStore.new(url: "...")
```

### Custom Service Pattern

Research describes a custom ThrottleService pattern:

```ruby
# app/services/throttle_service.rb
class ThrottleService
  def self.throttle!(key, limit: 100, period: :hour)
    redis_key = "THROTTLE-#{period.upcase}-#{key}"
    count = Redis.current.incr(redis_key)
    Redis.current.expire(redis_key, period_seconds(period))
    raise RateLimitExceeded if count > limit
  end
end
```

### Devise Lockable

Devise lockable only handles:
- Failed login attempts (configurable max)
- Account locking after N failures
- Unlock strategies (email, time, both)

This does NOT protect against:
- API endpoint abuse
- Password reset flooding
- Email enumeration
- General DoS

### Key Investigation Points

1. **Check Gemfile**: Does Loomio include `rack-attack` gem?
2. **Check Initializer**: Is `config/initializers/rack_attack.rb` present?
3. **Check ThrottleService**: Is there a custom service in `app/services/`?
4. **Check Cache Config**: Is Redis configured for rate limiting?

## Reconciliation Hypothesis

Both documents may be partially correct:
- **Discovery** looked at the authentication controllers and found only Devise lockable
- **Research** found ThrottleService which may be used for specific operations (email, API)

The investigation should determine:
1. ThrottleService scope (which operations use it)
2. Whether broader API rate limiting exists (Rack::Attack)
3. Coverage gaps in the current implementation

## Impact on Go Rewrite

Go implementation needs:
- Rate limiting middleware (e.g., go-chi/httprate, ulule/limiter)
- Redis-based distributed rate limiting for multi-instance deployment
- Different limits for different operation types:
  - Login attempts: strict (5/hour)
  - API requests: moderate (100/hour)
  - Email operations: strict (1/hour for bounces)
