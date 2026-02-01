# OAuth Providers: Follow-up Items

## Executive Summary

Third-party discovery has **resolved** the major discrepancy identified in our research. Our research incorrectly listed Facebook, Slack, and Microsoft as OAuth providers when they do not exist in the Loomio codebase. The third-party investigation provides definitive evidence from source code analysis.

---

## Discrepancy Resolution

### Original Discrepancy

| Source | Claimed Providers |
|--------|-------------------|
| Our Research | Google, Facebook, Slack, Microsoft, SAML (5) |
| Third-party Discovery | Google, OAuth (generic), SAML, Nextcloud (4) |

### Resolution Status: RESOLVED

Third-party discovery provides definitive evidence:

1. **Source of truth**: `/config/providers.yml` explicitly lists only 4 providers
2. **Controller files**: Only 4 identity controllers exist in `app/controllers/identities/`
3. **Client files**: Only 4 client classes exist in `app/extras/clients/`
4. **Route generation**: Routes dynamically generated from `Identity::PROVIDERS` constant

**Our research was incorrect** - Facebook, Slack, and Microsoft OAuth were never implemented.

---

## Areas Requiring Clarification

### HIGH Priority

#### 1. SSO-Only Mode Behavior

**Issue**: Third-party documents mention `ENV['FEATURES_DISABLE_EMAIL_LOGIN']` and `ENV['LOOMIO_SSO_FORCE_USER_ATTRS']` but don't fully document the interaction between these flags.

**Questions**:
- When both flags are set, what is the user creation/linking flow?
- Is email or UID the canonical identifier when `FORCE_USER_ATTRS` is enabled?
- How are duplicate emails handled across multiple SSO providers?

**Files to investigate**:
- `orig/loomio/app/controllers/identities/base_controller.rb` - Lines handling SSO mode
- `orig/loomio/app/models/boot/site.rb` - Boot payload generation
- `orig/loomio/app/controllers/api/v1/sessions_controller.rb` - Session handling with SSO

**Priority**: HIGH (Affects implementation of SSO-only deployments)

---

#### 2. Pending Identity Flow

**Issue**: Third-party mentions `session[:pending_identity_id]` for unlinked identities but doesn't document the complete flow.

**Questions**:
- What controller handles the "link to existing account" flow?
- Is there a timeout on pending identities?
- How is email verification triggered for existing account linking?

**Files to investigate**:
- `orig/loomio/app/controllers/api/v1/sessions_controller.rb`
- `orig/loomio/app/controllers/registrations_controller.rb` (if exists)
- `orig/loomio/app/services/` - Any identity-related services

**Priority**: HIGH (Core user onboarding flow)

---

### MEDIUM Priority

#### 3. Generic OAuth Configuration Validation

**Issue**: Third-party documents 8 `OAUTH_*` environment variables for generic OAuth but doesn't specify:

**Questions**:
- Is there validation that all required vars are present at boot?
- What happens if `OAUTH_ATTR_*` path doesn't exist in provider response?
- Are there defaults or is it a hard failure?

**Files to investigate**:
- `orig/loomio/app/extras/clients/oauth.rb` - Error handling in `fetch_identity_params`
- `orig/loomio/config/initializers/` - Any OAuth validation

**Priority**: MEDIUM (Configuration validation)

---

#### 4. SAML Security Settings

**Issue**: Third-party shows SAML settings with security disabled:
```ruby
settings.security[:authn_requests_signed] = false
settings.security[:logout_requests_signed] = false
settings.security[:metadata_signed] = false
settings.security[:digest_method] = XMLSecurity::Document::SHA1
settings.security[:signature_method] = XMLSecurity::Document::RSA_SHA1
```

**Questions**:
- Is this intentional for maximum IdP compatibility?
- Are there environment variables to enable signing?
- Should implementation default to more secure settings?

**Files to investigate**:
- `orig/loomio/app/controllers/identities/saml_controller.rb` - Full `saml_settings` method
- `orig/loomio-deploy/env_template` - Any SAML security configuration

**Priority**: MEDIUM (Security implications for enterprise deployments)

---

#### 5. Identity Destruction Authorization

**Issue**: The `destroy` action in controllers checks `current_user.identities.find_by(identity_type:)` but doesn't document:

**Questions**:
- What prevents removal of last identity if email login is disabled?
- Is there admin override capability?

**Files to investigate**:
- `orig/loomio/app/controllers/identities/base_controller.rb` - `destroy` method
- `orig/loomio/app/models/user.rb` - Any `before_destroy` callbacks on identities

**Priority**: MEDIUM (Edge case but important for SSO-only mode)

---

### LOW Priority

#### 6. Frontend Vestigial Code

**Issue**: Third-party correctly identifies vestigial Facebook color definitions and Slack filtering:
```javascript
case 'facebook': return '#3b5998';  // Never instantiated
providers.filter(provider => provider.name !== 'slack')
```

**Questions**:
- Should we document this for frontend cleanup?
- Is there historical context in git for why Slack was filtered?

**Files to investigate**:
- `orig/loomio/vue/src/components/auth/provider_form.vue`
- Git history of provider_form.vue

**Priority**: LOW (Cosmetic frontend issue)

---

#### 7. `slack_community_id` Column

**Issue**: Third-party mentions migration adding `slack_community_id` to users table but doesn't clarify if it's still in use.

**Questions**:
- Is this column actively used or deprecated?
- Was it for Slack workspace linking (not SSO)?
- Should schema include this column?

**Files to investigate**:
- `orig/loomio/db/migrate/20170310101359_add_slack_community_to_user.rb`
- `orig/loomio/app/models/user.rb` - Any `slack_community_id` references
- `orig/loomio/db/schema.rb` - Confirm column still exists

**Priority**: LOW (May be deprecated code)

---

## Contradictions Found

### None Identified

Third-party discovery aligns with or supersedes our research findings. The only "contradiction" was our incorrect listing of Facebook/Slack/Microsoft, which third-party has definitively corrected with source evidence.

---

## Incomplete Areas in Third-Party Documentation

### 1. Access Token Refresh

Third-party documents token storage in `omniauth_identities.access_token` but doesn't address:
- Token refresh mechanism (if any)
- Token expiration handling
- Whether tokens are used post-authentication

**Needed**: Clarify if OAuth tokens are stored just for initial auth or for ongoing API calls.

### 2. Group Identity (SSO for Groups)

Third-party mentions `Ability::GroupIdentity` in authorization research but OAuth documentation doesn't cover:
- What is a "group identity"?
- How does it differ from user identity?
- Is this for SAML/OAuth configured per-group?

**Files to investigate**:
- `orig/loomio/app/models/group_identity.rb`
- `orig/loomio/app/models/ability/group_identity.rb`

### 3. Test Fixtures/Factories

Third-party mentions test files exist but doesn't document:
- Test factory definitions for Identity
- Mock OAuth/SAML providers for testing
- How to run identity-related tests

**Needed**: TDD approach requires understanding test patterns.

---

## Summary Table

| Item | Priority | Status | Action Required |
|------|----------|--------|-----------------|
| SSO-Only mode behavior | HIGH | Incomplete | Investigate controller flow |
| Pending identity flow | HIGH | Incomplete | Document complete user flow |
| OAuth config validation | MEDIUM | Unclear | Check error handling |
| SAML security settings | MEDIUM | Documented but needs context | Verify intentional |
| Identity destruction | MEDIUM | Incomplete | Check edge cases |
| Frontend vestigial code | LOW | Documented | None (frontend issue) |
| `slack_community_id` | LOW | Unknown | Verify if deprecated |
| Access token refresh | LOW | Not documented | Clarify token lifecycle |
| Group identity | LOW | Not documented | Investigate model |
| Test patterns | LOW | Not documented | Review test files |
