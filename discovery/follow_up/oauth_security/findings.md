# OAuth Security Investigation Findings

## Executive Summary

**CRITICAL FINDING**: Loomio does NOT use OmniAuth. It implements a custom OAuth flow without state parameter validation, making it **VULNERABLE** to OAuth CSRF attacks.

## Discrepancy Resolution

| Original Claim | Finding | Status |
|----------------|---------|--------|
| OAuth state parameter may be missing | **CONFIRMED** - No state parameter in OAuth flows | VULNERABILITY CONFIRMED |
| Check for `omniauth-rails_csrf_protection` gem | **NOT APPLICABLE** - OmniAuth is not used | N/A |
| Check `provider_ignores_state` setting | **NOT APPLICABLE** - OmniAuth is not used | N/A |

## Ground Truth Answers

### 1. Does the Loomio OAuth implementation use state parameters to prevent CSRF?

**NO.** The OAuth implementation does not use state parameters.

**Evidence** (`app/controllers/identities/base_controller.rb:85-95`):
```ruby
def oauth_url
  "#{oauth_host}?#{oauth_params.to_query}"
end

def oauth_params
  { client.client_key_name => client.key, redirect_uri: redirect_uri, scope: oauth_scope }
end
```

**Evidence** (`app/controllers/identities/google_controller.rb:13-15`):
```ruby
def oauth_params
  super.merge(response_type: :code, scope: client.scope.join('+'))
end
```

**Evidence** (`app/controllers/identities/oauth_controller.rb:13-16`):
```ruby
def oauth_params
  client = Clients::Oauth.instance
  { client.client_key_name => client.key, redirect_uri: redirect_uri, scope: ENV.fetch('OAUTH_SCOPE'),  response_type: :code }
end
```

**Analysis**: None of the OAuth params methods include a `state` parameter. The `state` parameter is the standard OAuth 2.0 mechanism for CSRF protection.

### 2. How does OmniAuth handle state validation in this codebase?

**NOT APPLICABLE.** Loomio does NOT use OmniAuth at all.

**Evidence** (`Gemfile`):
- No `gem 'omniauth'` present
- No `gem 'omniauth-rails_csrf_protection'` present
- No `gem 'omniauth-google-oauth2'` or similar provider gems

**Evidence** (`config/initializers/`):
- No `omniauth.rb` initializer exists
- OmniAuth middleware is not configured anywhere

**Implementation**: Loomio uses a custom OAuth implementation in:
- `app/controllers/identities/base_controller.rb` - Base OAuth flow
- `app/controllers/identities/google_controller.rb` - Google-specific
- `app/controllers/identities/oauth_controller.rb` - Generic OAuth
- `app/controllers/identities/nextcloud_controller.rb` - Nextcloud-specific
- `app/extras/clients/*.rb` - Token exchange clients

### 3. Are there any custom CSRF protections in the identity controllers?

**PARTIAL** - Rails CSRF protection exists but does not cover the OAuth CSRF attack vector.

**Evidence** (`app/helpers/protected_from_forgery.rb:1-24`):
```ruby
module ProtectedFromForgery
  def self.included(base)
    base.after_action :set_xsrf_token
  end

  protected
  def verified_request?
    super || Rails.env.development? || cookies['csrftoken'] == request.headers['X-CSRF-TOKEN']
  end
end
```

**Analysis**: This provides standard Rails CSRF protection for form submissions, but:
1. OAuth initiation uses GET requests (`get :oauth` in routes)
2. OAuth callbacks use GET requests (`get :authorize` in routes)
3. No state parameter is generated, stored, or validated

## Attack Vector Analysis

### OAuth CSRF Attack Scenario

1. Attacker initiates OAuth flow with their own OAuth account
2. Attacker captures the authorization code from the callback URL
3. Attacker crafts malicious link: `https://loomio.example/google/authorize?code=ATTACKERS_CODE`
4. Victim clicks the malicious link while logged into Loomio
5. Loomio exchanges the code for attacker's identity
6. Attacker's identity gets linked to victim's account OR victim gets signed in as attacker

### Specific Vulnerable Code Path

**File**: `app/controllers/identities/base_controller.rb:7-58`

```ruby
def create
  if params[:error].present?
    flash[:error] = t(:'auth.oauth_cancelled')
    return redirect_to session.delete(:back_to) || dashboard_path
  end

  access_token = fetch_access_token  # No state validation before this
  # ... processes OAuth callback without verifying request origin
end
```

**Attack**: An attacker can force a victim to hit this endpoint with an authorization code from the attacker's OAuth account.

## Routes Analysis

**File**: `config/routes.rb:455-466`

```ruby
Identity::PROVIDERS.each do |provider|
  scope provider do
    get :oauth,      to: "identities/#{provider}#oauth",   as: :"#{provider}_oauth"
    get :authorize,  to: "identities/#{provider}#create",  as: :"#{provider}_authorize"
    get '/',         to: "identities/#{provider}#destroy", as: :"#{provider}_unauthorize"
  end
end
```

**Issues**:
1. OAuth initiation is via GET (should be POST with CSRF token)
2. OAuth callback is via GET (acceptable but requires state validation)
3. No state parameter validation anywhere in the flow

## Provider Configuration

**File**: `config/providers.yml`

```yaml
identity:
  - oauth
  - saml
  - google
  - nextcloud
```

Four OAuth/SSO providers are configured, all sharing the vulnerable base controller.

## SAML Analysis

**File**: `app/controllers/identities/saml_controller.rb:2`

```ruby
skip_before_action :verify_authenticity_token
```

SAML callback skips CSRF protection, which is intentional for SAML (POST binding), but the SAML response itself contains protections via XML signatures and response validation.

## Severity Assessment

| Factor | Assessment |
|--------|------------|
| **CVSS Base Score** | ~6.5 (Medium-High) |
| **Attack Complexity** | Low - Simple crafted URL |
| **Privileges Required** | None |
| **User Interaction** | Required - Victim must click link |
| **Impact** | Account linking hijack possible |
| **Exploitability** | High - Well-known attack |

## Recommended Mitigations

### Immediate (High Priority)

1. **Add state parameter to OAuth flow**:
   - Generate cryptographically random state in `oauth` action
   - Store state in session
   - Validate state in `create` action before processing code

2. **Change OAuth initiation to POST**:
   - Update routes to use POST for `/oauth`
   - Use forms with CSRF tokens for OAuth buttons

### Example Secure Implementation

```ruby
# In oauth action
def oauth
  session[:back_to] = params[:back_to] || request.referrer
  session[:oauth_state] = SecureRandom.hex(32)
  redirect_to oauth_url
end

def oauth_params
  super.merge(state: session[:oauth_state])
end

# In create action
def create
  if params[:state].blank? || params[:state] != session.delete(:oauth_state)
    return respond_with_error(403, "Invalid OAuth state - possible CSRF attack")
  end
  # ... existing flow
end
```

## References

- [RFC 9700: OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/rfc9700/) (January 2025)
- [OWASP OAuth2 Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/OAuth2_Cheat_Sheet.html)
- [Auth0: Prevent CSRF Attacks in OAuth 2.0](https://auth0.com/blog/prevent-csrf-attacks-in-oauth-2-implementations/)
- [Auth0: State Parameters](https://auth0.com/docs/secure/attack-protection/state-parameters)
