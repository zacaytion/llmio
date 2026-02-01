# OAuth Providers: Complete Findings

## Executive Summary

The initial discovery documentation listed **4 providers** (oauth, saml, google, nextcloud), while external research claimed **5 providers** (Google, Facebook, Slack, Microsoft, SAML). After thorough code investigation:

**Ground Truth: Loomio has exactly 4 OAuth/SSO identity providers**
1. Google OAuth 2.0
2. Generic OAuth 2.0 (configurable)
3. SAML 2.0
4. Nextcloud OAuth 2.0

Facebook, Slack, and Microsoft **do not exist as identity providers** - they were never implemented or have been completely removed. The webhook serializers for Slack/Microsoft Teams are for *outbound notifications*, not authentication.

## Provider Configuration

### Source of Truth: `/config/providers.yml`

```yaml
# for providers which provide OAuth login capability
identity:
  - oauth
  - saml
  - google
  - nextcloud
```

**File path:** `/Users/z/Code/loomio/config/providers.yml`

### Identity Model Constant

```ruby
# /Users/z/Code/loomio/app/models/identity.rb:10
PROVIDERS = YAML.load_file(Rails.root.join("config", "providers.yml"))['identity']
```

### Route Generation

```ruby
# /Users/z/Code/loomio/config/routes.rb:455-461
Identity::PROVIDERS.each do |provider|
  scope provider do
    get :oauth,     to: "identities/#{provider}#oauth",   as: :"#{provider}_oauth"
    get :authorize, to: "identities/#{provider}#create",  as: :"#{provider}_authorize"
    get '/',        to: "identities/#{provider}#destroy", as: :"#{provider}_unauthorize"
  end
end
```

**Generated routes per provider:**
- `GET /google/oauth` -> `Identities::GoogleController#oauth`
- `GET /google/authorize` -> `Identities::GoogleController#create`
- `GET /google/` -> `Identities::GoogleController#destroy`
- (same pattern for oauth, nextcloud)

**SAML has special routing:**
```ruby
# /Users/z/Code/loomio/config/routes.rb:463-466
scope :saml do
  post :oauth,    to: 'identities/saml#create',   as: :saml_oauth_callback
  get :metadata,  to: 'identities/saml#metadata', as: :saml_metadata
end
```

## Provider Enable/Disable Mechanism

Providers are enabled dynamically via environment variables. The boot payload generation:

```ruby
# /Users/z/Code/loomio/app/models/boot/site.rb:31-33
identityProviders: AppConfig.providers.fetch('identity', []).map do |provider|
  ({ name: provider, href: send("#{provider}_oauth_path") } if ENV["#{provider.upcase}_APP_KEY"])
end.compact
```

**Required environment variables per provider:**

| Provider | Enable Variable | Additional Required |
|----------|-----------------|---------------------|
| google | `GOOGLE_APP_KEY` | `GOOGLE_APP_SECRET` |
| oauth | `OAUTH_APP_KEY` | `OAUTH_APP_SECRET`, `OAUTH_AUTH_URL`, `OAUTH_TOKEN_URL`, `OAUTH_PROFILE_URL`, `OAUTH_SCOPE`, `OAUTH_ATTR_UID`, `OAUTH_ATTR_NAME`, `OAUTH_ATTR_EMAIL` |
| nextcloud | `NEXTCLOUD_APP_KEY` | `NEXTCLOUD_APP_SECRET`, `NEXTCLOUD_HOST` |
| saml | `SAML_APP_KEY` | `SAML_IDP_METADATA` or `SAML_IDP_METADATA_URL`, optionally `SAML_ISSUER` |

## Clarifying the Research Discrepancy

### Why Research Documentation Listed Facebook/Slack/Microsoft

1. **Webhook integrations confused with OAuth**: Loomio has Slack and Microsoft Teams *webhook serializers* for sending outbound notifications:
   - `/Users/z/Code/loomio/app/serializers/webhook/slack/event_serializer.rb`
   - `/Users/z/Code/loomio/app/serializers/webhook/microsoft/event_serializer.rb`

   These are for **chatbot webhooks**, not SSO authentication.

2. **Frontend code has color/icon definitions for providers that don't exist**:
   ```javascript
   // /Users/z/Code/loomio/vue/src/components/auth/provider_form.vue:22-28
   providerColor(provider) {
     switch (provider) {
       case 'facebook': return '#3b5998';  // Never instantiated
       case 'slack': return '#e9a820';     // Never instantiated
       // ...
     }
   }
   ```

3. **Historical provider removal**: The frontend explicitly filters out 'slack':
   ```javascript
   // /Users/z/Code/loomio/vue/src/components/auth/provider_form.vue:43
   providers() { return AppConfig.identityProviders.filter(provider => provider.name !== 'slack'); }
   ```
   This suggests Slack was once planned or partially implemented but removed.

4. **Legacy migration exists**: There's a migration adding `slack_community_id` to users:
   - `/Users/z/Code/loomio/db/migrate/20170310101359_add_slack_community_to_user.rb`

   This was for a different Slack integration (workspace communities), not SSO.

## Controller File Inventory

```
/Users/z/Code/loomio/app/controllers/identities/
├── base_controller.rb      # Shared OAuth flow logic
├── google_controller.rb    # Google-specific OAuth
├── nextcloud_controller.rb # Nextcloud-specific OAuth
├── oauth_controller.rb     # Generic OAuth (configurable)
└── saml_controller.rb      # SAML 2.0 (separate from base)
```

**No Facebook, Slack, or Microsoft controller files exist.**

## Client Service Inventory

```
/Users/z/Code/loomio/app/extras/clients/
├── base.rb       # Base HTTP client
├── google.rb     # Google API client
├── nextcloud.rb  # Nextcloud API client
├── oauth.rb      # Generic OAuth client
├── request.rb    # HTTP request wrapper
└── webhook.rb    # Outbound webhook client (not auth)
```

**No Facebook, Slack, or Microsoft client files exist.**

## Test Coverage

```
/Users/z/Code/loomio/spec/controllers/identities/
├── oauth_controller_spec.rb  # 417 lines, comprehensive
└── saml_controller_spec.rb   # 15160 bytes, comprehensive
```

**No Google or Nextcloud controller specs exist** (likely covered by generic OAuth tests).

## Database Schema

```ruby
# /Users/z/Code/loomio/db/schema.rb:619-633
create_table "omniauth_identities", id: :serial, force: :cascade do |t|
  t.integer "user_id"
  t.string "email", limit: 255
  t.datetime "created_at", precision: nil, null: false
  t.datetime "updated_at", precision: nil, null: false
  t.string "identity_type", limit: 255   # <- provider name stored here
  t.string "uid", limit: 255             # <- unique ID from SSO provider
  t.string "name", limit: 255
  t.string "access_token", default: ""
  t.string "logo"
  t.jsonb "custom_fields", default: {}, null: false
  t.index ["identity_type", "uid"], name: "index_omniauth_identities_on_identity_type_and_uid"
end
```

Valid `identity_type` values are: `google`, `oauth`, `saml`, `nextcloud`

## Open Questions

1. **Why does the frontend have Facebook color definitions?**
   - Likely vestigial code from planned feature or copy-paste from another project
   - Safe to remove

2. **What is `slack_community_id` on users table?**
   - Appears to be for a different integration (Slack workspace linking), not SSO
   - Would need git history investigation to confirm

3. **Are there any database records with other identity_type values?**
   - Would require production database query: `SELECT DISTINCT identity_type FROM omniauth_identities;`
