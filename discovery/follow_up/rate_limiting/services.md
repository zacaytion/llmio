# Rate Limiting Services Documentation

## Overview

Loomio implements rate limiting through two distinct services:

1. **Rack::Attack** - Middleware-level IP-based throttling
2. **ThrottleService** - Application-level operation-based throttling

---

## 1. Rack::Attack Middleware

### Location
`/Users/z/Code/loomio/config/initializers/rack_attack.rb`

### Purpose
Protect API endpoints from abuse by limiting requests per IP address.

### Configuration

```ruby
# Environment variable multipliers for tuning
RATE_MULTIPLIER = ENV.fetch('RACK_ATTACK_RATE_MULTPLIER', 1).to_i
TIME_MULTIPLIER = ENV.fetch('RACK_ATTACK_TIME_MULTPLIER', 1).to_i
```

### IP Detection

```ruby
class Rack::Attack::Request < ::Rack::Request
  def remote_ip
    # Priority: Cloudflare header > ActionDispatch > Rack
    @remote_ip ||= (env['HTTP_CF_CONNECTING_IP'] ||
                    env['action_dispatch.remote_ip'] ||
                    ip).to_s
  end
end
```

### Throttle Rules

Two rules per endpoint:
1. **Hourly limit**: `limit * RATE_MULTIPLIER` requests per `1 * TIME_MULTIPLIER` hour
2. **Daily limit**: `limit * 3 * RATE_MULTIPLIER` requests per `1 * TIME_MULTIPLIER` day

### Protected Endpoints

```ruby
IP_POST_LIMITS = {
  '/api/v1/trials' => 10,
  '/api/v1/announcements' => 100,
  '/api/v1/groups' => 20,
  '/api/v1/templates' => 10,
  '/api/v1/login_tokens' => 10,
  '/api/v1/membership_requests' => 100,
  '/api/v1/memberships' => 100,
  '/api/v1/identities' => 10,
  '/api/v1/discussions' => 100,
  '/api/v1/polls' => 100,
  '/api/v1/outcomes' => 100,
  '/api/v1/stances' => 100,
  '/api/v1/profile' => 100,
  '/api/v1/comments' => 100,
  '/api/v1/reactions' => 100,
  '/api/v1/link_previews' => 100,
  '/api/v1/registrations' => 10,
  '/api/v1/sessions' => 10,
  '/api/v1/contact_messages' => 10,
  '/api/v1/contact_requests' => 10,
  '/api/v1/discussion_readers' => 1000,
  '/rails/active_storage/direct_uploads' => 20
}
```

### Request Matching

Only applies to:
- POST requests
- PUT requests
- PATCH requests
- Paths starting with the configured routes

```ruby
req.remote_ip if (req.post? || req.put? || req.patch?) && req.path.starts_with?(route)
```

### Logging

Subscribes to ActiveSupport notifications:

```ruby
ActiveSupport::Notifications.subscribe(/rack_attack/) do |name, start, finish, request_id, req_h|
  req = req_h[:request]
  Rails.logger.warn [name,
                     req.remote_ip,
                     req.request_method,
                     req.fullpath,
                     request_id].join(' ')
end
```

Log format: `{event_name} {ip} {method} {path} {request_id}`

### Response on Throttle

Default Rack::Attack behavior:
- Status: `429 Too Many Requests`
- Body: `"Retry later\n"` (default)
- No custom `Retry-After` header

---

## 2. ThrottleService

### Location
`/Users/z/Code/loomio/app/services/throttle_service.rb`

### Purpose
Fine-grained application-level rate limiting for specific operations with user/entity context.

### Implementation

```ruby
module ThrottleService
  class LimitReached < StandardError
  end

  def self.reset!(per)
    CACHE_REDIS_POOL.with do |client|
      client.scan_each(match: "THROTTLE-#{per.upcase}*") { |key| client.del(key) }
    end
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
end
```

### API Methods

#### `ThrottleService.can?(options)`

Non-blocking check. Increments counter and returns boolean.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `key` | String | `'default-key'` | Operation identifier |
| `id` | Any | `1` | Entity identifier (user_id, email, etc.) |
| `max` | Integer | `100` | Maximum allowed count |
| `inc` | Integer | `1` | Increment amount |
| `per` | String | `'hour'` | Time window: `'hour'` or `'day'` |

Returns: `true` if under limit, `false` if exceeded.

#### `ThrottleService.limit!(options)`

Blocking version. Raises exception on limit exceeded.

Same parameters as `can?`.

Returns: `true` if under limit.
Raises: `ThrottleService::LimitReached` if exceeded.

#### `ThrottleService.reset!(per)`

Clears all throttle counters for a time period.

| Parameter | Type | Description |
|-----------|------|-------------|
| `per` | String | `'hour'` or `'day'` |

### Redis Key Pattern

```
THROTTLE-{PERIOD}-{key}-{id}
```

Examples:
- `THROTTLE-DAY-UserInviterInvitations-42`
- `THROTTLE-HOUR-bounce-user@example.com`

### Environment Variable Override

Limits can be overridden per-key via environment variables:

```ruby
ENV.fetch('THROTTLE_MAX_'+key, max)
```

Example: `THROTTLE_MAX_UserInviterInvitations=100000`

### Redis Configuration

Uses `CACHE_REDIS_POOL` from `/Users/z/Code/loomio/config/initializers/sidekiq.rb`:

```ruby
channels_redis_url = (ENV['REDIS_CACHE_URL'] || ENV.fetch('REDIS_URL', 'redis://localhost:6379/0'))
CACHE_REDIS_POOL = ConnectionPool.new(size: ENV.fetch('REDIS_POOL_SIZE', 30).to_i, timeout: 5) {
  Redis.new(url: channels_redis_url)
}
```

### Current Usages

#### 1. User Invitations

**Location:** `/Users/z/Code/loomio/app/extras/user_inviter.rb:102-106`

```ruby
ThrottleService.limit!(key: 'UserInviterInvitations',
                       id: actor.id,
                       max: actor.invitations_rate_limit,
                       inc: emails.length + ids.length,
                       per: :day)
```

**Limits** (from `/Users/z/Code/loomio/app/models/user.rb:195-201`):

```ruby
def invitations_rate_limit
  if user.is_paying?
    ENV.fetch('PAID_INVITATIONS_RATE_LIMIT', 50000)
  else
    ENV.fetch('TRIAL_INVITATIONS_RATE_LIMIT', 500)
  end.to_i
end
```

| User Type | Default Limit | Env Variable |
|-----------|---------------|--------------|
| Paying users | 50,000/day | `PAID_INVITATIONS_RATE_LIMIT` |
| Trial users | 500/day | `TRIAL_INVITATIONS_RATE_LIMIT` |

#### 2. Email Bounce Notices

**Location:** `/Users/z/Code/loomio/app/services/received_email_service.rb:35`

```ruby
if ThrottleService.can?(key: 'bounce', id: email.sender_email.downcase, max: 1, per: 'hour')
  ForwardMailer.bounce(to: email.sender_name_and_email).deliver_now
else
  Rails.logger.info("bounce throttled for #{email.sender_email}")
end
```

| Limit | Period | Purpose |
|-------|--------|---------|
| 1 | hour | Prevent email flooding when users reply to noreply addresses |

---

## 3. Devise Lockable

### Location
`/Users/z/Code/loomio/config/initializers/devise.rb:134-155`

### Purpose
Lock accounts after repeated failed login attempts.

### Configuration

```ruby
config.lock_strategy = :failed_attempts
config.unlock_keys = [:email]
config.unlock_strategy = :both  # email + time
config.maximum_attempts = ENV.fetch('MAX_LOGIN_ATTEMPTS', 20).to_i
config.unlock_in = 6.hours
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MAX_LOGIN_ATTEMPTS` | 20 | Failed attempts before lockout |

---

## Testing

### ThrottleService Tests

**Location:** `/Users/z/Code/loomio/spec/services/throttle_service_spec.rb`

```ruby
describe 'ThrottleService' do
  it 'limits the number of times i can do something'
  it 'limits the number of times i can do something, with inc'
  it 'correctly resets a throttle'
  it 'does not reset all throttles'
  it 'raises exception for limit!'
end
```

### ReceivedEmailService Tests

**Location:** `/Users/z/Code/loomio/spec/services/received_email_service_spec.rb:96-111`

```ruby
it 'sends a delivery failure notice and destroys the email when sent to notifications address and not throttled' do
  expect(ThrottleService).to receive(:can?).with(key: 'bounce', id: 'user@example.com', max: 1, per: 'hour').and_return(true)
  # ...
end

it 'does not send a delivery failure notice when throttled and still destroys the email' do
  expect(ThrottleService).to receive(:can?).with(key: 'bounce', id: 'user@example.com', max: 1, per: 'hour').and_return(false)
  # ...
end
```
