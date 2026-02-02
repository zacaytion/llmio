# Loomio Security/Bug Report

**Generated:** 2026-02-01
**Purpose:** Security analysis for Loomio rewrite contract
**Scope:** Authentication, authorization, rate limiting, webhooks, CSRF, injection vulnerabilities

---

## Executive Summary

This report documents security issues and potential vulnerabilities discovered during the Loomio codebase analysis. Issues are categorized by severity and domain, with code references and remediation guidance.

### Risk Summary

| Severity | Count | Description |
|----------|-------|-------------|
| CRITICAL | 0 | No immediate exploitation risks found |
| HIGH | 3 | Missing OAuth state, ThrottleService 500, Bot API rate limits |
| MEDIUM | 4 | Webhook security gaps, Redis TTL, CSRF skips |
| LOW | 3 | Data hygiene and tracking gaps |

### Key Findings

1. **OAuth CSRF vulnerability** - Missing state parameter enables login CSRF attacks
2. **Rate limiting returns HTTP 500** - ThrottleService raises unhandled exception instead of 429
3. **Bot APIs have no rate limiting** - `/api/b2/` and `/api/b3/` bypass all throttling
4. **Webhook payloads unsigned** - No HMAC verification for webhook recipients
5. **Multiple CSRF token skips** - Several controllers disable CSRF protection

---

## Issues by Domain

### 1. Authentication & OAuth

#### ISSUE-001: Missing OAuth State Parameter (CSRF)

| Attribute | Value |
|-----------|-------|
| **Severity** | HIGH |
| **Confidence** | HIGH |
| **Domain** | Authentication |
| **Source** | `/Users/z/Code/loomio/discovery/final/oauth_security.md` |

**Description:**
Loomio's OAuth implementation does not generate or validate the `state` parameter, which RFC 6749 recommends for CSRF protection.

**Affected Files:**

| File | Lines | Issue |
|------|-------|-------|
| `app/controllers/identities/base_controller.rb` | 93-95 | No state in `oauth_params` |
| `app/controllers/identities/google_controller.rb` | 13-15 | No state in `oauth_params` |
| `app/controllers/identities/oauth_controller.rb` | 13-16 | No state in `oauth_params` |
| `app/controllers/identities/nextcloud_controller.rb` | 17-19 | No state in `oauth_params` |
| `app/controllers/identities/base_controller.rb` | 7-58 | No state validation in callback |

**Code Snippet:**
```ruby
# app/controllers/identities/base_controller.rb:93-95
def oauth_params
  { client.client_key_name => client.key, redirect_uri: redirect_uri, scope: oauth_scope }
  # NOTE: No 'state' parameter generated or validated
end
```

**Attack Scenario:**
1. Attacker initiates OAuth flow on Loomio, captures authorization URL
2. Attacker creates page that auto-submits callback with attacker's OAuth code
3. Victim visits attacker's page while logged into Loomio
4. Victim's browser completes OAuth callback, linking attacker's identity to victim's account
5. Attacker gains access to victim's account via OAuth login

**Remediation:**
```ruby
# In oauth action:
session[:oauth_state] = SecureRandom.hex(32)
redirect_to oauth_uri + "&state=#{session[:oauth_state]}"

# In create action:
raise OAuthError unless params[:state] == session.delete(:oauth_state)
```

---

#### ISSUE-002: OAuth Access Tokens Stored But Unused

| Attribute | Value |
|-----------|-------|
| **Severity** | LOW |
| **Confidence** | HIGH |
| **Domain** | Authentication |
| **Source** | `/Users/z/Code/loomio/discovery/final/oauth_security.md` |

**Description:**
OAuth access tokens are stored in the `omniauth_identities.access_token` column but never read or used after initial profile fetch. This creates unnecessary attack surface if the database is compromised.

**Affected Files:**
- `app/controllers/identities/base_controller.rb:78` - Token stored
- `db/schema.rb:627` - Column definition

**Code Snippet:**
```ruby
# app/controllers/identities/base_controller.rb:78
def fetch_identity_params(token)
  client = "Clients::#{controller_name.classify}".constantize.new(token: token)
  client.fetch_identity_params.merge({ access_token: token, identity_type: controller_name })
  # access_token stored but never subsequently read
end
```

**Remediation:**
- Remove `access_token` storage or
- Implement token rotation/cleanup job or
- Use tokens for provider-specific features (profile sync)

---

### 2. Rate Limiting

#### ISSUE-003: ThrottleService Returns HTTP 500 Instead of 429

| Attribute | Value |
|-----------|-------|
| **Severity** | HIGH |
| **Confidence** | HIGH |
| **Domain** | Rate Limiting |
| **Source** | `/Users/z/Code/loomio/discovery/final/rate_limiting.md` |

**Description:**
When `ThrottleService::LimitReached` is raised, no controller rescue handler catches it, resulting in HTTP 500 Internal Server Error instead of the proper HTTP 429 Too Many Requests.

**Affected Files:**
- `app/services/throttle_service.rb:2-3,22` - Exception definition and raising
- `app/controllers/api/v1/snorlax_base.rb:2-7` - Missing rescue_from handler

**Code Snippet:**
```ruby
# app/services/throttle_service.rb:22
raise ThrottleService::LimitReached.new "Throttled! #{key}-#{id}"

# app/controllers/api/v1/snorlax_base.rb:2-7 - NO handler for LimitReached
rescue_from(CanCan::AccessDenied)                    { |e| respond_with_standard_error e, 403 }
rescue_from(Subscription::MaxMembersExceeded)        { |e| respond_with_standard_error e, 403 }
rescue_from(ActionController::UnpermittedParameters) { |e| respond_with_standard_error e, 400 }
rescue_from(ActionController::ParameterMissing)      { |e| respond_with_standard_error e, 400 }
rescue_from(ActiveRecord::RecordNotFound)            { |e| respond_with_standard_error e, 404 }
# Missing: rescue_from(ThrottleService::LimitReached)
```

**Impact:**
- Clients cannot distinguish rate limits from server errors
- No `Retry-After` header provided
- Monitoring alerts fire on "legitimate" throttling
- Non-compliant with HTTP standards

**Remediation:**
```ruby
# Add to app/controllers/api/v1/snorlax_base.rb
rescue_from(ThrottleService::LimitReached) do |e|
  response.headers['Retry-After'] = '3600'
  render json: {
    error: 'rate_limit_exceeded',
    message: 'Too many requests. Please try again later.',
    retry_after: 3600
  }, status: 429
end
```

---

#### ISSUE-004: Bot APIs Lack Rate Limiting

| Attribute | Value |
|-----------|-------|
| **Severity** | HIGH |
| **Confidence** | HIGH |
| **Domain** | Rate Limiting |
| **Source** | `/Users/z/Code/loomio/discovery/final/rate_limiting.md` |

**Description:**
The `/api/b2/` and `/api/b3/` bot API namespaces have no rate limiting at any level. They are excluded from Rack::Attack IP-based limits and do not use ThrottleService.

**Affected Files:**
- `config/initializers/rack_attack.rb:17-39` - Only `/api/v1/` routes listed
- `app/controllers/api/b2/base_controller.rb` - No ThrottleService calls
- `app/controllers/api/b3/users_controller.rb` - No ThrottleService calls

**Code Snippet:**
```ruby
# config/initializers/rack_attack.rb:17-39
IP_POST_LIMITS = {
  '/api/v1/trials' => 10,
  '/api/v1/announcements' => 100,
  # ... only /api/v1/ routes listed
  # NO /api/b1/, /api/b2/, /api/b3/ entries
}
```

**Impact:**
- Authenticated users can make unlimited API calls
- No protection against compromised API keys
- DoS vector via bot endpoints
- Resource exhaustion attacks possible

**Remediation:**
```ruby
# Add to config/initializers/rack_attack.rb IP_POST_LIMITS
'/api/b2/discussions' => 50,
'/api/b2/polls' => 50,
'/api/b2/memberships' => 100,
'/api/b2/comments' => 100,
'/api/b3/users' => 10,
```

---

#### ISSUE-005: Redis Throttle Counters Have No TTL

| Attribute | Value |
|-----------|-------|
| **Severity** | MEDIUM |
| **Confidence** | HIGH |
| **Domain** | Rate Limiting |
| **Source** | `/Users/z/Code/loomio/discovery/final/rate_limiting.md` |

**Description:**
ThrottleService uses `Redis::Counter` which does not set TTL on increment. Keys persist indefinitely until explicitly deleted by `rake loomio:hourly_tasks`. If the rake job fails, users remain throttled and Redis memory grows.

**Affected Files:**
- `app/services/throttle_service.rb:14-15` - No TTL set
- `lib/tasks/loomio.rake:222-241` - Manual reset via scan_each

**Code Snippet:**
```ruby
# app/services/throttle_service.rb:14-15
Redis::Counter.new(k).increment(inc)
Redis::Counter.new(k).value <= ENV.fetch('THROTTLE_MAX_'+key, max)
# No TTL set - keys persist indefinitely
```

**Impact:**
- If hourly_tasks fails, users remain throttled beyond window
- Redis memory grows without cleanup
- Dependent on external scheduler reliability

**Remediation:**
```ruby
# Replace with atomic INCRBY + EXPIRE
CACHE_REDIS_POOL.with do |client|
  client.multi do |multi|
    multi.incrby(k, inc)
    multi.expire(k, per.to_s == 'hour' ? 3600 : 86400)
  end
end
```

---

### 3. Webhooks

#### ISSUE-006: No Webhook HMAC Signing

| Attribute | Value |
|-----------|-------|
| **Severity** | MEDIUM |
| **Confidence** | HIGH |
| **Domain** | Webhooks |
| **Source** | `/Users/z/Code/loomio/discovery/final/webhook_events.md` |

**Description:**
Outgoing webhook payloads are not signed with HMAC or any cryptographic signature. Webhook receivers cannot verify payload authenticity or detect tampering.

**Affected Files:**
- `app/extras/clients/webhook.rb:1-23` - No signature generation
- `app/extras/clients/base.rb:62-68` - Payload serialized without signing
- `app/services/chatbot_service.rb:49-55` - Direct POST without signature

**Code Snippet:**
```ruby
# app/services/chatbot_service.rb:52-53
payload = serializer.new(event, root: false, scope: {template_name: template_name, recipient: recipient}).as_json
req = Clients::Webhook.new.post(chatbot.server, params: payload)
# No X-Loomio-Signature header or HMAC computation
```

**Impact:**
- Webhook receivers cannot verify payload origin
- No protection against replay attacks
- Receivers must rely on IP filtering or other mechanisms

**Remediation:**
```ruby
# Generate HMAC signature
signature = OpenSSL::HMAC.hexdigest('SHA256', chatbot.webhook_secret, payload.to_json)
headers['X-Loomio-Signature'] = "sha256=#{signature}"
```

---

#### ISSUE-007: No Webhook Circuit Breaker

| Attribute | Value |
|-----------|-------|
| **Severity** | MEDIUM |
| **Confidence** | HIGH |
| **Domain** | Webhooks |
| **Source** | `/Users/z/Code/loomio/discovery/final/webhook_events.md` |

**Description:**
There is no circuit breaker mechanism to automatically disable failing webhooks. Failed deliveries are logged to Sentry but continue indefinitely, wasting resources.

**Affected Files:**
- `db/schema.rb:156-170` - No `failure_count`, `disabled`, `last_failure_at` columns
- `app/models/chatbot.rb:1-18` - No failure tracking methods
- `app/services/chatbot_service.rb:52-55` - Errors logged but no state update

**Code Snippet:**
```ruby
# app/services/chatbot_service.rb:52-55
req = Clients::Webhook.new.post(chatbot.server, params: payload)
if req.response.code != 200
  Sentry.capture_message("chatbot id #{chatbot.id} post event id #{event.id} failed...")
  # No failure counter increment, no auto-disable
end
```

**Impact:**
- Failing webhooks continue receiving attempts indefinitely
- Resource waste on permanently broken integrations
- No admin visibility into webhook health

**Remediation:**
- Add `failure_count`, `last_failure_at`, `disabled_at` columns to chatbots table
- Increment failure count on non-200 responses
- Auto-disable after N consecutive failures
- Provide admin UI to view webhook health and re-enable

---

### 4. CSRF Protection

#### ISSUE-008: Multiple Controllers Skip CSRF Verification

| Attribute | Value |
|-----------|-------|
| **Severity** | MEDIUM |
| **Confidence** | HIGH |
| **Domain** | CSRF |
| **Source** | Codebase search |

**Description:**
Several controllers skip CSRF token verification. While some are legitimate (SAML callbacks, API endpoints), others may warrant review.

**Controllers with `skip_before_action :verify_authenticity_token`:**

| Controller | Reason | Risk |
|------------|--------|------|
| `identities/saml_controller.rb:2` | SAML callbacks (legitimate) | LOW - SAML has own verification |
| `application_controller.rb:25` | Bug tunnel endpoint | LOW - Sentry integration |
| `direct_uploads_controller.rb:3` | Active Storage uploads | **MEDIUM** - Also sets `protect_from_forgery` |
| `api/b3/users_controller.rb:2` | Bot API | LOW - API key auth |
| `api/hocuspocus_controller.rb:2` | Collaborative editing | **MEDIUM** - Uses secret_token |
| `received_emails_controller.rb:2` | Inbound email webhook | LOW - External service |
| `api/b2/base_controller.rb:2` | Bot API | LOW - API key auth |

**Code Snippet:**
```ruby
# app/controllers/direct_uploads_controller.rb:1-3
class DirectUploadsController < ActiveStorage::DirectUploadsController
  protect_from_forgery with: :exception  # Set THEN skipped - contradiction
  skip_before_action :verify_authenticity_token
```

**Concern:**
The `DirectUploadsController` sets `protect_from_forgery` then immediately skips it, which is contradictory and may indicate confusion about the intended security model.

**Remediation:**
- Review each skip to ensure API authentication is sufficient
- Remove contradictory `protect_from_forgery` from DirectUploadsController
- Document why each skip is safe

---

### 5. Permissions & Authorization

#### ISSUE-009: `members_can_add_guests` Not Tracked in Paper Trail

| Attribute | Value |
|-----------|-------|
| **Severity** | LOW |
| **Confidence** | HIGH |
| **Domain** | Permissions |
| **Source** | `/Users/z/Code/loomio/discovery/final/permission_flags.md` |

**Description:**
The `members_can_add_guests` permission flag is actively used in authorization checks but is not included in the paper_trail `only:` list, meaning changes to this setting are not audit-logged.

**Affected Files:**
- `app/models/group.rb:132-158` - Paper trail configuration
- `app/models/ability/group.rb:45-48` - Flag used in ability check

**Code Snippet:**
```ruby
# app/models/group.rb:132-158 - members_can_add_guests ABSENT from list
has_paper_trail only: [:name,
                       :members_can_add_members,
                       :members_can_edit_discussions,
                       # ... other flags listed
                       # members_can_add_guests NOT included
                       :admins_can_edit_user_content,
                       :attachments]
```

**Impact:**
- No audit trail for guest invitation permission changes
- Security-relevant setting changes not tracked

**Remediation:**
Add `members_can_add_guests` to the paper_trail `only:` list.

---

### 6. Input Handling

#### ISSUE-010: SQL Query Patterns Review

| Attribute | Value |
|-----------|-------|
| **Severity** | LOW |
| **Confidence** | MEDIUM |
| **Domain** | SQL Injection |
| **Source** | Codebase search |

**Description:**
Several files use string interpolation in SQL queries. Most use parameterized queries correctly, but the patterns warrant documentation.

**Safe Patterns Found:**
```ruby
# app/queries/attachment_query.rb - Uses parameterized queries
where("active_storage_blobs.filename ilike ?", "%#{query}%")  # Safe - uses ?

# app/models/discussion.rb:97
kept.where('discussions.title ilike ?', "%#{q}%")  # Safe - uses ?

# app/models/poll.rb:172
where("polls.title ilike :fragment", fragment: "%#{fragment}%")  # Safe - uses :named
```

**Pattern Requiring Review:**
```ruby
# app/queries/discussion_query.rb:39 - String interpolation in WHERE clause
.where("#{'(discussions.private = false) OR ' if or_public}
        (discussions.group_id IN (:user_group_ids))
        #{'OR (groups.parent_members_can_see_discussions = TRUE...' if or_subgroups}")
```

**Assessment:**
The interpolated values are boolean conditionals (not user input), so this is not a SQL injection vulnerability. However, the pattern is unusual and should be documented.

**Recommendation:**
Consider refactoring to use scopes or explicit conditional chaining for clarity.

---

### 7. HTML/XSS Protection

#### ISSUE-011: HTML Sanitization Implementation

| Attribute | Value |
|-----------|-------|
| **Severity** | INFORMATIONAL |
| **Confidence** | HIGH |
| **Domain** | XSS Prevention |
| **Source** | Codebase search |

**Description:**
Loomio implements proper HTML sanitization for rich text content via `HasRichText` concern.

**Implementation:**
```ruby
# app/models/concerns/has_rich_text.rb:16-28
define_method "sanitize_#{field}!" do
  tags = %w[strong em b i p s code pre big div small hr br span mark
            h1 h2 h3 ul ol li abbr a img video audio blockquote table
            thead th tr td iframe u]
  attributes = %w[href src alt title data-type data-iframe-container
                  data-done data-mention-id poster controls data-author-id
                  data-uid data-checked data-due-on data-color data-remind
                  width height target colspan rowspan data-text-align]

  self[field] = Rails::Html::WhiteListSanitizer.new.sanitize(
    self[field], tags: tags, attributes: attributes
  )
end

before_save :"sanitize_#{field}!"
```

**Assessment:**
- Whitelist-based sanitization (good)
- Applied before_save (good)
- `iframe` tag allowed (intentional for embedded content)
- External links get `rel="nofollow ugc noreferrer noopener"` (good)

**Potential Concern:**
The `iframe` tag is allowed. Ensure embedding sources are validated or use `sandbox` attribute.

---

## Additional Security Observations

### Secret Token Exposure

The `secret_token` is serialized to the frontend for real-time channel authentication:

```ruby
# app/serializers/current_user_serializer.rb:5
:secret_token

# app/models/boot/user.rb:18
channel_token: user.secret_token
```

**Assessment:** This is intentional for WebSocket authentication. The token is regenerated on logout (`sessions_controller.rb:19`).

### API Key Handling

B2 API uses user-specific API keys passed as query parameters:

```ruby
# app/controllers/api/b2/base_controller.rb:11
@current_user ||= User.active.find_by(api_key: params[:api_key])
```

**Concern:** API keys in query strings may be logged in server access logs and browser history.

**Recommendation:** Consider supporting `Authorization` header as alternative.

---

## Summary Table

| ID | Issue | Severity | Confidence | Domain | File Reference |
|----|-------|----------|------------|--------|----------------|
| ISSUE-001 | Missing OAuth state parameter | HIGH | HIGH | Auth | `identities/base_controller.rb` |
| ISSUE-002 | OAuth tokens stored unused | LOW | HIGH | Auth | `identities/base_controller.rb` |
| ISSUE-003 | ThrottleService returns 500 | HIGH | HIGH | Rate Limit | `snorlax_base.rb` |
| ISSUE-004 | Bot APIs lack rate limiting | HIGH | HIGH | Rate Limit | `rack_attack.rb` |
| ISSUE-005 | Redis counters no TTL | MEDIUM | HIGH | Rate Limit | `throttle_service.rb` |
| ISSUE-006 | No webhook HMAC signing | MEDIUM | HIGH | Webhooks | `chatbot_service.rb` |
| ISSUE-007 | No webhook circuit breaker | MEDIUM | HIGH | Webhooks | `chatbot.rb` |
| ISSUE-008 | CSRF verification skips | MEDIUM | HIGH | CSRF | Multiple controllers |
| ISSUE-009 | Permission flag untracked | LOW | HIGH | Permissions | `group.rb` |
| ISSUE-010 | SQL patterns review | LOW | MEDIUM | Injection | `discussion_query.rb` |
| ISSUE-011 | HTML sanitization | INFO | HIGH | XSS | `has_rich_text.rb` |

---

## Remediation Priority

### Immediate (Before Production)

1. **ISSUE-001** - Add OAuth state parameter for CSRF protection
2. **ISSUE-003** - Add rescue_from handler for ThrottleService::LimitReached

### Short Term (Next Sprint)

3. **ISSUE-004** - Add rate limiting to bot API endpoints
4. **ISSUE-005** - Add Redis TTL to throttle counters
5. **ISSUE-006** - Implement webhook HMAC signing

### Medium Term (Next Quarter)

6. **ISSUE-007** - Implement webhook circuit breaker
7. **ISSUE-008** - Review and document CSRF skip decisions
8. **ISSUE-009** - Add members_can_add_guests to paper_trail

### Low Priority (Backlog)

9. **ISSUE-002** - Clean up unused OAuth tokens
10. **ISSUE-010** - Refactor SQL query patterns for clarity

---

## Appendix: File References

| File | Purpose |
|------|---------|
| `app/controllers/identities/base_controller.rb` | OAuth callback handling |
| `app/controllers/api/v1/snorlax_base.rb` | REST controller error handlers |
| `app/services/throttle_service.rb` | Application rate limiting |
| `config/initializers/rack_attack.rb` | IP-based rate limiting |
| `app/services/chatbot_service.rb` | Webhook delivery |
| `app/models/chatbot.rb` | Webhook configuration |
| `app/models/group.rb` | Group model with paper_trail |
| `app/models/concerns/has_rich_text.rb` | HTML sanitization |
| `app/controllers/api/b2/base_controller.rb` | Bot API base controller |
| `app/controllers/api/b3/users_controller.rb` | B3 API controller |

---

*Generated: 2026-02-01*
*Analysis Sources: oauth_security.md, rate_limiting.md, webhook_events.md, permission_flags.md, codebase search*
