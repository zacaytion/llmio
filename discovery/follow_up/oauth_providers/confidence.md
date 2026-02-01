# OAuth Providers: Verification Checklist

## Verification Summary

| Claim | Status | Confidence | Evidence |
|-------|--------|------------|----------|
| Exactly 4 identity providers exist | PASS | 5/5 | config/providers.yml, controller files |
| Google OAuth implemented | PASS | 5/5 | google_controller.rb, clients/google.rb |
| Generic OAuth implemented | PASS | 5/5 | oauth_controller.rb, clients/oauth.rb, spec |
| SAML implemented | PASS | 5/5 | saml_controller.rb, spec |
| Nextcloud OAuth implemented | PASS | 5/5 | nextcloud_controller.rb, clients/nextcloud.rb |
| Facebook OAuth exists | FAIL | 5/5 | No controller, no client, no routes |
| Slack OAuth exists | FAIL | 5/5 | No controller, no client, no routes |
| Microsoft OAuth exists | FAIL | 5/5 | No controller, no client, no routes |
| Providers enabled via env vars | PASS | 5/5 | boot/site.rb:31-33 |
| Identity table uses identity_type | PASS | 5/5 | db/schema.rb:624 |

## Detailed Verification

### 1. Provider Configuration Source

**Claim:** Providers defined in `/config/providers.yml`

**Evidence:**
```yaml
# /Users/z/Code/loomio/config/providers.yml (lines 1-6)
identity:
  - oauth
  - saml
  - google
  - nextcloud
```

**Status:** PASS (5/5)

---

### 2. Controller Files Exist

**Claim:** Each provider has a controller

**Evidence:**
```bash
$ ls /Users/z/Code/loomio/app/controllers/identities/
base_controller.rb
google_controller.rb
nextcloud_controller.rb
oauth_controller.rb
saml_controller.rb
```

**Status:** PASS (5/5)

---

### 3. Client Files Exist

**Claim:** OAuth providers have client classes

**Evidence:**
```bash
$ ls /Users/z/Code/loomio/app/extras/clients/
base.rb
google.rb
nextcloud.rb
oauth.rb
request.rb
webhook.rb  # Not auth-related
```

**Status:** PASS (5/5)

---

### 4. Routes Generated Dynamically

**Claim:** Routes generated from Identity::PROVIDERS

**Evidence:**
```ruby
# /Users/z/Code/loomio/config/routes.rb (lines 455-461)
Identity::PROVIDERS.each do |provider|
  scope provider do
    get :oauth,     to: "identities/#{provider}#oauth"
    get :authorize, to: "identities/#{provider}#create"
    get '/',        to: "identities/#{provider}#destroy"
  end
end
```

**Status:** PASS (5/5)

---

### 5. Provider Enable/Disable Mechanism

**Claim:** Providers enabled by `{PROVIDER}_APP_KEY` env var

**Evidence:**
```ruby
# /Users/z/Code/loomio/app/models/boot/site.rb (lines 31-33)
identityProviders: AppConfig.providers.fetch('identity', []).map do |provider|
  ({ name: provider, href: send("#{provider}_oauth_path") } if ENV["#{provider.upcase}_APP_KEY"])
end.compact
```

**Status:** PASS (5/5)

---

### 6. Facebook Does NOT Exist

**Claim:** Facebook OAuth is not implemented

**Evidence:**
- No file: `/app/controllers/identities/facebook_controller.rb`
- No file: `/app/extras/clients/facebook.rb`
- Not in `/config/providers.yml`
- Grep for "facebook" in controllers: No matches
- Frontend color code is vestigial (never rendered)

**Status:** PASS - CONFIRMED NON-EXISTENT (5/5)

---

### 7. Slack Does NOT Exist

**Claim:** Slack OAuth is not implemented

**Evidence:**
- No file: `/app/controllers/identities/slack_controller.rb`
- No file: `/app/extras/clients/slack.rb`
- Not in `/config/providers.yml`
- Frontend explicitly filters out 'slack': `filter(provider => provider.name !== 'slack')`
- Slack serializers are for webhooks, not auth

**Status:** PASS - CONFIRMED NON-EXISTENT (5/5)

---

### 8. Microsoft Does NOT Exist

**Claim:** Microsoft OAuth is not implemented

**Evidence:**
- No file: `/app/controllers/identities/microsoft_controller.rb`
- No file: `/app/extras/clients/microsoft.rb`
- Not in `/config/providers.yml`
- Microsoft serializers are for Teams webhooks, not auth

**Status:** PASS - CONFIRMED NON-EXISTENT (5/5)

---

### 9. Database Schema Correct

**Claim:** `omniauth_identities` table has `identity_type` column

**Evidence:**
```ruby
# /Users/z/Code/loomio/db/schema.rb (lines 619-633)
create_table "omniauth_identities" do |t|
  t.string "identity_type", limit: 255
  t.string "uid", limit: 255
  t.index ["identity_type", "uid"], name: "index_omniauth_identities_on_identity_type_and_uid"
end
```

**Status:** PASS (5/5)

---

### 10. Test Coverage Exists

**Claim:** OAuth and SAML have controller specs

**Evidence:**
```bash
$ ls /Users/z/Code/loomio/spec/controllers/identities/
oauth_controller_spec.rb   # 417 lines
saml_controller_spec.rb    # ~400 lines
```

**Status:** PASS (5/5)

---

## Confidence Ratings Scale

| Rating | Meaning |
|--------|---------|
| 5/5 | Verified via direct code inspection |
| 4/5 | High confidence from multiple indirect sources |
| 3/5 | Moderate confidence, some ambiguity |
| 2/5 | Low confidence, needs further investigation |
| 1/5 | Speculation only |

## Open Questions Requiring Further Investigation

| Question | Priority | Method to Resolve |
|----------|----------|-------------------|
| Why does frontend have Facebook color definitions? | Low | Git history |
| What is `slack_community_id` on users? | Low | Git history, check if column still used |
| Are there any orphaned identity_type values in prod DB? | Medium | Production DB query |

## Conclusion

**All critical claims verified.** The ground truth is:

1. Loomio has **exactly 4 identity providers**: google, oauth, saml, nextcloud
2. Facebook, Slack, and Microsoft OAuth **do not exist** in the codebase
3. Slack/Microsoft serializers are for **outbound webhooks**, not authentication
4. Research documentation was incorrect - likely confused integrations with SSO
