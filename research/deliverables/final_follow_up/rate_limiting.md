# Rate Limiting - Follow-up Items

## Overview

This document identifies discrepancies, contradictions, and areas requiring clarification between the third-party discovery findings and our research baseline.

---

## Discrepancies Between Findings

### 1. Our Research Claims vs. Third-Party Verification

| Claim (Our Research) | Third-Party Finding | Status | Priority |
|---------------------|---------------------|--------|----------|
| API limit is 100/hour | Varies by endpoint (10-1000/hour via Rack::Attack) | **INCORRECT** | HIGH |
| Login limit is 5/hour | 10/hour (Rack::Attack) + 20 total attempts (Devise lockable) | **INCORRECT** | HIGH |
| Email bounce is 1/hour | Confirmed at 1/hour | **CORRECT** | - |
| Redis key pattern `THROTTLE-{HOUR|DAY}-{key}-{id}` | Confirmed | **CORRECT** | - |
| ThrottleService exists | Confirmed | **CORRECT** | - |

**Analysis:** Our research (`research/investigation/api.md`) incorrectly stated:
- "default 100/hour" - This is the ThrottleService default, NOT the API rate limit
- "Login attempts: 5/hour" - Source of this value is unclear; actual is 10/hour IP-based + 20 attempts lockout

**Action Required:** Determine the source of the "5/hour login" claim in our research. Was this a misinterpretation of Devise lockable or an outdated configuration?

---

## Contradictions Needing Resolution

### 2. Redis Key Expiration (TTL) - UNVERIFIED

**Third-party claim (confidence.md):**
> "Redis key expiration: Do ThrottleService keys automatically expire, or do they persist until reset?"

**Issue:** Neither source confirms whether Redis counters have TTL set automatically.

**Code Analysis:**
```ruby
# orig/loomio/app/services/throttle_service.rb:13-14
k = "THROTTLE-#{per.upcase}-#{key}-#{id}"
Redis::Counter.new(k).increment(inc)
```

The `redis-objects` gem's `Redis::Counter` does NOT automatically set TTL. Keys would persist indefinitely unless:
1. `ThrottleService.reset!(period)` is called
2. Redis memory eviction kicks in
3. A separate cleanup job exists

**Priority:** HIGH

**Investigation Target:**
- [ ] Check if there's a scheduled job that calls `ThrottleService.reset!`
- [ ] Verify redis-objects configuration for default TTL
- [ ] Check if Rails cache layer adds expiration

**File to investigate:** `orig/loomio/config/schedule.rb` or `whenever` gem configuration for scheduled resets.

---

### 3. Rack::Attack Cache Store - UNCLEAR

**Third-party claim (services.md):**
> "Store: Uses Rack::Attack's default cache (likely Rails.cache or a configured Redis)"

**Issue:** The third-party finding is speculative. The actual cache store is not confirmed.

**Code Analysis:**
```ruby
# orig/loomio/config/initializers/rack_attack.rb
# No explicit cache.store configuration visible
```

Rack::Attack defaults to `Rails.cache` if not configured. Rails cache configuration would be in `config/environments/*.rb`.

**Priority:** MEDIUM

**Questions for Third Party:**
1. Was `config/environments/production.rb` checked for cache store configuration?
2. Is there a `Rack::Attack.cache.store = ...` line elsewhere?

**Investigation Target:**
- [ ] `orig/loomio/config/environments/production.rb` - Check `config.cache_store` setting
- [ ] Verify if Rack::Attack shares Redis with Sidekiq/sessions or uses separate store

---

### 4. Bot API Endpoints (`/api/b1/`, `/api/b2/`, `/api/b3/`) - UNKNOWN STATUS

**Third-party claim (controllers.md):**
> "Bot APIs not in Rack::Attack: UNKNOWN - not in IP_POST_LIMITS"

**Issue:** These endpoints are documented in our research (`research/investigation/api.md`) but their rate limiting status is unverified.

**Priority:** MEDIUM

**Questions for Third Party:**
1. Were these endpoints searched for separate rate limiting configuration?
2. Do they rely on API key authentication as implicit rate limiting?
3. Are they intended for trusted bot integrations only?

**Investigation Targets:**
- [ ] `orig/loomio/app/controllers/api/b1/` - Check for controller-level rate limiting
- [ ] `orig/loomio/config/routes.rb` - Verify bot API authentication requirements

---

## Areas Where Third-Party Findings Are Incomplete

### 5. ThrottleService::LimitReached Exception Handling

**Third-party claim (findings.md):**
> "Response: HTTP 500 (Internal Server Error) - NOT caught by any rescue_from handler"

**Confirmation:** Verified by checking `snorlax_base.rb` - no `rescue_from(ThrottleService::LimitReached)` exists.

**Incomplete Analysis:** The third-party notes this as a "Critical" security gap but doesn't investigate:
1. Whether this is intentional (fail-closed approach)
2. What the actual error response body contains (stack trace exposure?)
3. Whether frontend handles 500 errors gracefully

**Priority:** HIGH

**Investigation Target:**
- [ ] Test actual 500 response body in development mode
- [ ] Check if `ActionController::Base` has a catch-all for production
- [ ] Review `config/environments/production.rb` for error handling config

---

### 6. Typo in Environment Variables

**Observation:** Third-party documents note environment variables with typos:
```ruby
# Actual code (rack_attack.rb:11-12)
RATE_MULTIPLIER = ENV.fetch('RACK_ATTACK_RATE_MULTPLIER', 1).to_i  # Missing 'I' in MULTIPLIER
TIME_MULTIPLIER = ENV.fetch('RACK_ATTACK_TIME_MULTPLIER', 1).to_i  # Missing 'I' in MULTIPLIER
```

**Issue:** The typos are in the actual codebase. Third-party documented them as-is, but didn't flag this as a potential configuration pitfall.

**Priority:** LOW

**Note:** Document correct spellings and support both variants for backwards compatibility.

---

### 7. Missing Retry-After Header Investigation

**Third-party claim (findings.md):**
> "Retry-After header: NOT explicitly set (uses Rack::Attack default behavior)"

**Incomplete:** The default Rack::Attack behavior is:
- Returns `429 Too Many Requests`
- Body: `"Retry later\n"`
- **No Retry-After header by default**

However, the third-party didn't verify:
1. Whether clients (Vue frontend) handle 429 responses
2. What the frontend's retry behavior is
3. Whether there's a custom `throttled_response` configured elsewhere

**Priority:** LOW (for implementation, we control both sides)

---

## Specific Questions for Third Party

### Critical Questions

1. **Redis TTL:** Did you verify that `redis-objects` `Redis::Counter` sets expiration automatically? The code doesn't show explicit TTL. How do counters reset?

2. **500 Error Body:** What is the actual response body when `ThrottleService::LimitReached` is raised? Is there stack trace leakage?

3. **Bot API Rate Limits:** Did you search `app/controllers/api/b1/`, `b2/`, `b3/` for rate limiting? These are missing from the analysis.

### Clarification Questions

4. **Rack::Attack Cache Store:** Can you confirm the actual cache store used? Is it `Rails.cache` (memory/file) or Redis?

5. **Login Rate Limit Source:** Our research claimed "5/hour" - did you find any evidence of this value anywhere in the codebase, even in comments or old commits?

6. **Cloudflare Integration:** The Rack::Attack config checks `HTTP_CF_CONNECTING_IP`. Is Cloudflare rate limiting also configured externally?

---

## Files Requiring Investigation

| File | Purpose | Priority |
|------|---------|----------|
| `orig/loomio/config/schedule.rb` | Check for scheduled ThrottleService.reset! calls | HIGH |
| `orig/loomio/config/environments/production.rb` | Verify Rails.cache and error handling config | HIGH |
| `orig/loomio/app/controllers/api/b1/base_controller.rb` | Check bot API rate limiting | MEDIUM |
| `orig/loomio/lib/tasks/*.rake` | Check for throttle reset rake tasks | MEDIUM |
| `orig/loomio/config/application.rb` | Check middleware stack for custom throttling | LOW |

---

## Impact on Implementation

### Issues to Address Before Implementation

1. **Define Canonical Rate Limits:** Our research had incorrect values. The implementation needs a clear specification derived from verified Rack::Attack limits.

2. **Redis Key Expiration Strategy:** Must implement TTL on counters. Cannot rely on manual reset.

3. **Error Response Standardization:** Must return proper 429 with `Retry-After` header, not 500.

4. **Bot API Decision:** Need product decision on whether `/api/b*` endpoints need rate limiting.

---

## Priority Summary

| Item | Priority | Type |
|------|----------|------|
| Redis TTL verification | HIGH | Investigation |
| ThrottleService exception response body | HIGH | Investigation |
| Correct rate limit values in our docs | HIGH | Documentation |
| Bot API rate limiting status | MEDIUM | Investigation |
| Rack::Attack cache store verification | MEDIUM | Investigation |
| Retry-After header client handling | LOW | Investigation |
| Environment variable typos | LOW | Documentation |
