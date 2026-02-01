# OAuth Security - Follow-up Investigation Brief

## Discrepancy Summary

Discovery documentation flags a potential CSRF vulnerability in OAuth implementation - specifically noting that OAuth state parameter validation may be missing. Our research documentation does not address OAuth security at all.

## Discovery Claims

**Source**: `discovery/initial/synthesis/uncertainties.md`

> "HIGH priority: OAuth state parameter not visible (security risk)"

**Source**: `discovery/initial/auth/confidence.md`

> "OAuth state parameter: Not verified if CSRF protection is implemented via state parameter in OAuth flows"

The discovery team rated Auth domain confidence as 3.8/5, with OAuth security being a key factor in the reduced score.

## Our Research Claims

**Source**: `research/investigation/api.md`, `research/investigation/authorization.md`

Our research documents OAuth as an authentication method but does not analyze:
- State parameter handling
- CSRF protection mechanisms
- Token validation flows
- Session binding after OAuth callback

## Ground Truth Needed

1. Does the Loomio OAuth implementation use state parameters to prevent CSRF?
2. How does OmniAuth handle state validation in this codebase?
3. Are there any custom CSRF protections in the identity controllers?

## Investigation Targets

- [ ] File: `orig/loomio/app/controllers/identities/base_controller.rb` - Check for state parameter handling in callback methods
- [ ] File: `orig/loomio/config/initializers/omniauth.rb` - Check OmniAuth configuration for state handling
- [ ] File: `orig/loomio/app/controllers/identities/google_controller.rb` - Verify OAuth flow implementation
- [ ] File: `orig/loomio/app/controllers/identities/saml_controller.rb` - Check SAML-specific security
- [ ] Command: `grep -r "state" orig/loomio/app/controllers/identities/` - Find state parameter references
- [ ] Command: `grep -r "omniauth.origin" orig/loomio/` - Check origin tracking for CSRF protection

## Priority

**HIGH** - This is a security concern. OAuth CSRF attacks can allow account takeover through forced authentication to attacker-controlled accounts.

## Rails Context (from OmniAuth Documentation)

### OmniAuth CSRF Protection

OmniAuth provides built-in CSRF protection. From official documentation:

**Required Gems for Rails:**
```ruby
gem 'omniauth'
gem "omniauth-rails_csrf_protection"  # CRITICAL for CSRF protection
```

The `omniauth-rails_csrf_protection` gem is **required** for Rails applications to handle CSRF protection. Without it, the OAuth flow may be vulnerable.

**Authenticity Token Configuration:**
```ruby
# Configure OmniAuth's authenticity token protection
OmniAuth::AuthenticityTokenProtection.default_options(key: "csrf.token", authenticity_param: "_csrf")
```

### Key Investigation Points

1. **Check Gemfile**: Does Loomio include `omniauth-rails_csrf_protection`?
2. **Check Initializer**: Is authenticity token protection configured?
3. **Check Provider Config**: Is `provider_ignores_state: true` set anywhere?

### Common Vulnerable Patterns

1. **Missing gem**: Not including `omniauth-rails_csrf_protection`
2. **provider_ignores_state: true**: Explicitly disabling protection
3. **Custom OAuth flow**: Bypassing OmniAuth's built-in protection
4. **POST-only callbacks**: OmniAuth requires POST for security; GET callbacks may be vulnerable

### Expected Secure Implementation

```ruby
# config/initializers/omniauth.rb
Rails.application.config.middleware.use OmniAuth::Builder do
  provider :developer if Rails.env.development?
  provider :google_oauth2, ENV['GOOGLE_CLIENT_ID'], ENV['GOOGLE_CLIENT_SECRET']
  # OmniAuth handles state validation automatically with omniauth-rails_csrf_protection
end
```

## Impact on Go Rewrite

If Loomio's OAuth is vulnerable:
- Go implementation MUST include state parameter validation
- Need to implement secure state generation and verification
- Consider using gorilla/sessions or similar for state storage

If Loomio's OAuth is secure:
- Document the pattern for Go implementation
- Ensure Go OAuth library provides equivalent protection
