# OAuth Security - Follow-up Items

## Executive Summary

The third-party discovery confirms a **HIGH severity security vulnerability** in Loomio's OAuth implementation. The custom OAuth flow lacks state parameter validation, making it vulnerable to CSRF attacks that could enable account hijacking.

This document catalogs discrepancies, contradictions, unclear areas, and specific questions requiring resolution before Go implementation.

---

## Discrepancies Between Discovery and Our Research

### 1. OAuth Analysis Scope Gap

| Aspect | Discovery Finding | Our Research | Discrepancy |
|--------|-------------------|--------------|-------------|
| OAuth CSRF protection | Confirmed missing state parameter | Not analyzed | **MAJOR GAP** - Our research did not investigate OAuth security at all |
| OmniAuth usage | Confirmed NOT used | Assumed OmniAuth | Our research brief incorrectly assumed OmniAuth was used |
| Custom OAuth implementation | Fully documented | Not documented | Discovery provides critical implementation details we lacked |

**Priority: HIGH**

**Action Required**: Update `research/investigation/authorization.md` to include OAuth authentication flow security analysis separate from CanCanCan authorization.

### 2. SAML Security Configuration

| Aspect | Discovery Finding | Our Research | Discrepancy |
|--------|-------------------|--------------|-------------|
| SAML signature verification | Disabled by default | Not analyzed | **SECURITY GAP** |
| SAML CSRF handling | Correctly skipped (POST binding) | Not analyzed | Acceptable pattern |

**Priority: MEDIUM**

**File Reference**: `orig/loomio/app/controllers/identities/saml_controller.rb:97-100`

```ruby
settings.security[:authn_requests_signed] = false
settings.security[:logout_requests_signed] = false
settings.security[:logout_responses_signed] = false
settings.security[:metadata_signed] = false
```

**Question for Third Party**: Is the SAML signature configuration intentional for compatibility, or should we recommend enabling signatures in the Go implementation?

---

## Contradictions Requiring Resolution

### 1. OmniAuth Assumption

**Our Research Brief** (`research/follow_up/oauth_security.md`) states:

> "How does OmniAuth handle state validation in this codebase?"
> "Check for `omniauth-rails_csrf_protection` gem"

**Discovery Finding**: OmniAuth is NOT used at all. Loomio has a completely custom OAuth implementation.

**Evidence** (`orig/loomio/Gemfile`):
- No `gem 'omniauth'` present
- No OmniAuth provider gems
- No `config/initializers/omniauth.rb` exists

**Resolution**: Our research brief was based on incorrect assumptions. The entire OmniAuth-related investigation path is N/A.

**Priority: HIGH** - This fundamentally changes the security analysis approach.

---

## Areas Where Discovery Findings Are Unclear or Incomplete

### 1. Token Refresh Mechanism

**Discovery Statement** (confidence.md):
> "No refresh token handling visible"

**Unclear**:
- Does Loomio actually need refresh tokens for any functionality?
- Are the stored access tokens used post-authentication?
- What is the token expiry behavior?

**Investigation Target**: `orig/loomio/app/models/identity.rb` and any code that uses `Identity#access_token`

**Priority: LOW** - May not affect Go implementation if tokens are only used during OAuth callback.

### 2. Access Token Storage Security

**Discovery Finding**: Access tokens stored as plaintext in `omniauth_identities.access_token` column.

**Unclear**:
- Are these access tokens ever used after initial authentication?
- What is the threat model - is database access assumed trusted?
- Should Go implementation encrypt at rest?

**File Reference**: `orig/loomio/db/schema.rb:627`
```ruby
t.string "access_token", default: ""
```

**Priority: MEDIUM** - Encryption at rest is a security best practice but may not be required for MVP.

### 3. SAML RelayState Handling

**Discovery** documents OAuth CSRF but doesn't explicitly analyze SAML RelayState parameter handling.

**Question**: Does the SAML flow use RelayState for redirect preservation, and is it validated?

**Investigation Target**: `orig/loomio/app/controllers/identities/saml_controller.rb:7-8`

```ruby
def oauth
  session[:back_to] = params[:back_to] || request.referrer
  auth_request = OneLogin::RubySaml::Authrequest.new
  redirect_to auth_request.create(saml_settings)
end
```

**Priority: LOW** - SAML has its own security model via signed assertions.

---

## Specific Questions for Third Party

### Security Severity

1. **Has this OAuth CSRF vulnerability been exploited in production Loomio instances?**
   - No evidence provided in discovery
   - Important for risk assessment

2. **Is the SAML signature disabled configuration intentional?**
   - Appears to be for IdP compatibility
   - Should Go implementation default to requiring signatures?

3. **What is the actual attack surface?**
   - How many Loomio instances have OAuth enabled?
   - Is Google OAuth the primary provider?

### Implementation Details

4. **Is `access_token` used after initial OAuth callback?**
   - File: `orig/loomio/app/models/identity.rb`
   - Need grep for `identity.access_token` usage

5. **What happens to pending identities?**
   - `session[:pending_identity_id]` is set when no user match found
   - Where is this consumed?

### Go Implementation Guidance

6. **Should the Go implementation fix the CSRF vulnerability?**
   - Assumed YES but need explicit confirmation
   - Should we match Rails behavior for backwards compatibility?

7. **What OAuth libraries are recommended for Go?**
   - Discovery doesn't provide Go-specific guidance
   - Options: `golang.org/x/oauth2`, custom implementation

---

## Files Requiring Further Investigation

| File | Line(s) | Investigation Reason | Priority |
|------|---------|---------------------|----------|
| `orig/loomio/app/controllers/identities/base_controller.rb` | 2-4, 7-58 | Verify OAuth flow matches discovery docs | HIGH |
| `orig/loomio/app/controllers/identities/oauth_controller.rb` | Full | Generic OAuth provider implementation | HIGH |
| `orig/loomio/app/controllers/identities/nextcloud_controller.rb` | Full | Fourth OAuth provider | MEDIUM |
| `orig/loomio/app/extras/clients/oauth.rb` | Full | Generic OAuth client | MEDIUM |
| `orig/loomio/app/controllers/sessions_controller.rb` | Full | How pending_identity_id is consumed | LOW |
| `orig/loomio/app/helpers/protected_from_forgery.rb` | 9-10 | CSRF bypass in development mode | LOW |

---

## Priority Summary

| Category | Item | Priority | Status |
|----------|------|----------|--------|
| Vulnerability | OAuth CSRF (missing state parameter) | **HIGH** | Confirmed by discovery |
| Gap | Our research missing OAuth analysis | **HIGH** | Requires update |
| Contradiction | OmniAuth assumption incorrect | **HIGH** | Resolved |
| Security | SAML signatures disabled | **MEDIUM** | Needs clarification |
| Security | Access token plaintext storage | **MEDIUM** | Best practice concern |
| Question | Token refresh behavior | **LOW** | May not affect Go |
| Question | Pending identity flow | **LOW** | Edge case handling |

---

## Recommended Actions

### Immediate (Before Go Implementation)

1. **Confirm OAuth CSRF fix is in scope** for Go rewrite
2. **Research Go OAuth libraries** with built-in state validation
3. **Document secure OAuth flow** as implementation target

### Short-term

4. **Investigate pending identity flow** - understand full user journey
5. **Decide on SAML signature policy** - secure by default vs. compatibility

### Long-term

6. **Consider encrypted token storage** in Go implementation
7. **Add security audit checklist** to Go OAuth implementation requirements
