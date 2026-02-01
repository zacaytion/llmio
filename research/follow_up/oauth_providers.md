# OAuth Providers - Follow-up Investigation Brief

## Discrepancy Summary

Discovery and Research documentation list **different OAuth providers** as available in Loomio:
- Discovery: 4 providers (oauth, saml, google, nextcloud)
- Research: 5 providers (Google, Facebook, Slack, Microsoft, SAML)

These lists have minimal overlap, suggesting one or both are incomplete or outdated.

## Discovery Claims

**Source**: `discovery/initial/auth/controllers.md`

Lists identity controllers:
- `Identities::GoogleController` - Google OAuth 2.0
- `Identities::SamlController` - SAML 2.0
- `Identities::OauthController` - Generic OAuth 2.0
- `Identities::NextcloudController` - Nextcloud-specific OAuth

**Source**: `discovery/initial/synthesis/external_services.md`

Documents configuration for:
- Google OAuth (`GOOGLE_APP_ID`, `GOOGLE_APP_SECRET`)
- Generic OAuth (`OAUTH_*` environment variables)
- Nextcloud (`NEXTCLOUD_*` environment variables)
- SAML (`SAML_*` environment variables)

## Our Research Claims

**Source**: `research/investigation/api.md` (inferred from routes)

Lists providers found in routes:
- Google
- Facebook
- Slack
- Microsoft
- SAML

**Source**: `research/loomio_initial_investigation.md`

Mentions OAuth as authentication method without enumerating specific providers.

## Ground Truth Needed

1. What identity controller files actually exist?
2. What providers are configured in routes.rb?
3. What providers are mentioned in environment template?
4. Are some providers deprecated/removed?

## Investigation Targets

- [ ] Command: `ls -la orig/loomio/app/controllers/identities/` - List all identity controllers
- [ ] File: `orig/loomio/config/routes.rb` - Check OAuth/identity routes (lines 387-420)
- [ ] File: `orig/loomio/config/providers.yml` or similar - Check provider configuration
- [ ] Command: `grep -r "provider" orig/loomio/config/initializers/omniauth.rb` - Check OmniAuth setup
- [ ] File: `orig/loomio-deploy/env_template` - Check which providers have env vars documented

## Priority

**HIGH** - OAuth providers directly affect:
- User authentication options in Go rewrite
- Third-party integration requirements
- Environment configuration documentation

## Rails Context

### OmniAuth Provider Registration

Providers are typically registered in an initializer:

```ruby
# config/initializers/omniauth.rb
Rails.application.config.middleware.use OmniAuth::Builder do
  provider :google_oauth2, ENV['GOOGLE_CLIENT_ID'], ENV['GOOGLE_CLIENT_SECRET']
  provider :facebook, ENV['FACEBOOK_APP_ID'], ENV['FACEBOOK_APP_SECRET']
  provider :slack, ENV['SLACK_CLIENT_ID'], ENV['SLACK_CLIENT_SECRET']
  # etc.
end
```

### Route Patterns

OAuth routes follow this pattern:
```ruby
# config/routes.rb
namespace :identities do
  get ':provider/callback', to: 'omniauth#callback'
  # or individual controllers
  get 'google/callback', to: 'google#callback'
end
```

### Provider Controllers

Each provider may have its own controller or use a shared OmniAuth callback:

```ruby
# Individual controller pattern (Discovery documents this)
class Identities::GoogleController < Identities::BaseController
  def callback
    # Google-specific handling
  end
end

# Shared callback pattern (Research may have assumed this)
class Identities::OmniauthCallbacksController < ApplicationController
  def callback
    provider = request.env['omniauth.auth']['provider']
    # Handle all providers uniformly
  end
end
```

## Reconciliation Hypothesis

The discrepancy likely arises from:

1. **Discovery examined actual controller files**: Found 4 specific controllers
2. **Research examined routes or documentation**: Found 5 provider names (possibly including deprecated ones)

Possible scenarios:
- Facebook/Slack/Microsoft routes exist but controllers use generic `OauthController`
- Facebook/Slack/Microsoft were removed and routes are stale
- Nextcloud is new and Research documentation is outdated

## Impact on Go Rewrite

Need to implement:
- **Confirmed**: Google OAuth 2.0, SAML 2.0
- **Verify**: Generic OAuth 2.0 support for arbitrary providers
- **Clarify**: Nextcloud-specific requirements (if different from generic OAuth)
- **Deprecation check**: Facebook, Slack, Microsoft support status

For Go implementation, consider:
- `golang.org/x/oauth2` for OAuth 2.0
- `crewjam/saml` for SAML 2.0
- Configurable provider list via environment variables
