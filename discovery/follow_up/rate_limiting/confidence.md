# Rate Limiting - Verification Checklist

## Research Claims Verification

### Claim 1: ThrottleService exists
| Status | Evidence |
|--------|----------|
| **PASS** | File exists at `/Users/z/Code/loomio/app/services/throttle_service.rb` |

### Claim 2: ThrottleService uses Redis-based throttling
| Status | Evidence |
|--------|----------|
| **PASS** | Uses `CACHE_REDIS_POOL` and `Redis::Counter` (lines 6, 14-15) |

### Claim 3: Configurable time windows (hour/day)
| Status | Evidence |
|--------|----------|
| **PASS** | `per` parameter accepts `'hour'` or `'day'` (line 12) |

### Claim 4: API rate limit is 100/hour
| Status | Evidence |
|--------|----------|
| **FAIL** | Rack::Attack has varying limits per endpoint (10-1000/hour). ThrottleService default is 100 but only used for invitations with different limits. |

### Claim 5: Login rate limit is 5/hour
| Status | Evidence |
|--------|----------|
| **FAIL** | Login is protected by Devise lockable (20 attempts total) and Rack::Attack (10/hour for `/api/v1/sessions`). Neither is 5/hour. |

### Claim 6: Email bounce rate limit is 1/hour
| Status | Evidence |
|--------|----------|
| **PASS** | Verified at `/Users/z/Code/loomio/app/services/received_email_service.rb:35` |

### Claim 7: Key pattern is `THROTTLE-{HOUR|DAY}-{key}-{id}`
| Status | Evidence |
|--------|----------|
| **PASS** | Verified at `/Users/z/Code/loomio/app/services/throttle_service.rb:13` |

---

## Discovery Claims Verification

### Claim 1: Rate limiting is absent beyond Devise lockable
| Status | Evidence |
|--------|----------|
| **FAIL** | Rack::Attack is fully configured with extensive endpoint coverage. ThrottleService provides additional application-level limits. |

### Claim 2: No visible application-level rate limiting
| Status | Evidence |
|--------|----------|
| **FAIL** | Both Rack::Attack (middleware) and ThrottleService (application) exist. |

---

## Ground Truth Verification Matrix

| Question | Verified | Source File | Line(s) | Confidence |
|----------|----------|-------------|---------|------------|
| ThrottleService exists? | Yes | `app/services/throttle_service.rb` | 1-25 | 5/5 |
| ThrottleService actively used? | Yes | `app/extras/user_inviter.rb`, `app/services/received_email_service.rb` | 102-106, 35 | 5/5 |
| Rack::Attack configured? | Yes | `config/initializers/rack_attack.rb` | 1-63 | 5/5 |
| Rack::Attack in Gemfile? | Yes | `Gemfile` | 49 | 5/5 |
| Rack::Attack version? | 6.8.0 | `Gemfile.lock` | 529 | 5/5 |
| IP detection handles Cloudflare? | Yes | `config/initializers/rack_attack.rb` | 4-8 | 5/5 |
| POST/PUT/PATCH limited? | Yes | `config/initializers/rack_attack.rb` | 44 | 5/5 |
| GET requests limited? | No | `config/initializers/rack_attack.rb` | N/A | 5/5 |
| 429 response configured? | Default only | `config/initializers/rack_attack.rb` | N/A | 4/5 |
| Retry-After header set? | No | `config/initializers/rack_attack.rb` | N/A | 5/5 |
| Devise lockable enabled? | Yes | `app/models/user.rb`, `config/initializers/devise.rb` | 25, 134-155 | 5/5 |
| Max login attempts? | 20 (default) | `config/initializers/devise.rb` | 152 | 5/5 |
| ThrottleService::LimitReached caught? | No | `app/controllers/api/v1/snorlax_base.rb` | 2-7 | 5/5 |
| Redis key pattern correct? | Yes | `app/services/throttle_service.rb` | 13 | 5/5 |
| Tests exist? | Yes | `spec/services/throttle_service_spec.rb` | 1-57 | 5/5 |

---

## Environment Variables Verified

| Variable | Default | Location | Verified |
|----------|---------|----------|----------|
| `RACK_ATTACK_RATE_MULTPLIER` | 1 | rack_attack.rb:11 | Yes |
| `RACK_ATTACK_TIME_MULTPLIER` | 1 | rack_attack.rb:12 | Yes |
| `THROTTLE_MAX_{key}` | varies | throttle_service.rb:15 | Yes |
| `PAID_INVITATIONS_RATE_LIMIT` | 50000 | user.rb:197 | Yes |
| `TRIAL_INVITATIONS_RATE_LIMIT` | 500 | user.rb:199 | Yes |
| `MAX_LOGIN_ATTEMPTS` | 20 | devise.rb:152 | Yes |
| `REDIS_CACHE_URL` | REDIS_URL | sidekiq.rb:4 | Yes |
| `REDIS_URL` | redis://localhost:6379/0 | sidekiq.rb:4 | Yes |

---

## Security Gap Assessment

| Gap | Severity | Verified | Recommendation |
|-----|----------|----------|----------------|
| ThrottleService::LimitReached unhandled | HIGH | Yes - no rescue_from in controller chain | Add handler returning 429 |
| No Retry-After headers | MEDIUM | Yes - not in rack_attack.rb | Add custom throttled_response |
| GET requests unprotected | MEDIUM | Yes - line 44 condition | Consider adding GET limits for search |
| Bot APIs not in Rack::Attack | UNKNOWN | Yes - not in IP_POST_LIMITS | Audit /api/b1/, b2/, b3/ |
| No per-user rate limiting | LOW | N/A | Consider for abuse prevention |

---

## Documentation Corrections Needed

### Previous Discovery Documentation

**File:** `discovery/initial/synthesis/uncertainties.md`

**Incorrect claim (line ~27):**
> "No visible rate limiting on login attempts beyond Devise lockable"

**Correction:** Rack::Attack limits `/api/v1/sessions` to 10 POST requests per hour per IP.

---

### Previous Discovery Documentation

**File:** `discovery/initial/auth/controllers.md`

**Incorrect claim (line ~497):**
> "Rate limiting: No visible rate limiting on login attempts (relies on Devise lockable)"

**Correction:** Three layers of protection:
1. Rack::Attack: 10/hour IP-based
2. Devise lockable: 20 attempts then lockout
3. Devise pwned_password: Checks password breach databases (production only)

---

## Overall Confidence Assessment

| Category | Score | Justification |
|----------|-------|---------------|
| Existence of rate limiting | 5/5 | Multiple verified systems |
| Endpoint coverage accuracy | 5/5 | Direct code analysis |
| Threshold accuracy | 5/5 | All values from source |
| Redis patterns | 5/5 | Code-verified patterns |
| Gap identification | 4/5 | Bot APIs need further investigation |
| **Overall** | **4.8/5** | High confidence, minor unknowns around bot APIs |

---

## Open Questions

1. **Bot API rate limiting:** Are `/api/b1/`, `/api/b2/`, `/api/b3/` intentionally unprotected or protected elsewhere?

2. **Webhook rate limiting:** How are incoming webhook endpoints protected?

3. **Redis key expiration:** Do ThrottleService keys automatically expire, or do they persist until reset? (Need to check redis-objects default TTL behavior)

4. **Cache store for Rack::Attack:** Is it using Rails.cache (which may be Redis or memory) or a dedicated store?

---

## Verification Commands Used

```bash
# Find ThrottleService
grep -r "ThrottleService" app/

# Find Rack::Attack config
cat config/initializers/rack_attack.rb

# Check Gemfile
grep -E "rack-attack|redis-objects" Gemfile

# Find usages
grep -rn "ThrottleService" --include="*.rb" .

# Check rescue_from handlers
grep -rn "rescue_from" app/controllers/

# Verify Devise lockable
grep -rn "lockable\|maximum_attempts" config/
```
