# Unanswered Questions

> Questions requiring further investigation with source references.

## High Priority (Blocking Go Implementation)

| Question | Investigation Target |
|----------|---------------------|
| **How is `pg_search_documents` populated?** | `orig/loomio/app/models/concerns/` pg_search config, callbacks |
| **What triggers search reindexing?** | Search for `update_pg_search_document` or similar |
| **What are webhook permission values?** | `orig/loomio/app/models/webhook.rb`, grep for `permissions` |
| **Which events trigger Redis pub/sub?** | `orig/loomio/config/initializers/`, EventBus listeners |
| **How does Stance `latest` transaction work?** | `orig/loomio/app/services/stance_service.rb` |
| **Guest migration incomplete** | Data integrity: existing guest records have `guest = false` instead of `true` |

### Investigation Commands

```bash
# Find pg_search configuration
grep -rn "pg_search" orig/loomio/app/models/

# Find EventBus listeners
grep -rn "EventBus.listen" orig/loomio/

# Find MessageChannelService callers
grep -rn "MessageChannelService" orig/loomio/app/

# Find webhook permissions usage
grep -rn "permissions" orig/loomio/app/models/webhook.rb
grep -rn "webhook.*permissions" orig/loomio/
```

## Medium Priority (Feature Parity)

| Question | Investigation Target |
|----------|---------------------|
| **How does RecordCloner work?** | `orig/loomio/app/services/record_cloner.rb` |
| **Which associations are cloned?** | RecordCloner methods |
| **How are IDs remapped during cloning?** | RecordCloner internals |
| **What translation service is used?** | `orig/loomio/app/services/translation_service.rb` |
| **Is auto-translation supported?** | Translation configuration |
| **What are exact subscription plan tiers?** | `orig/loomio/app/services/subscription_service.rb` |
| **Where does Rails publish to `chatbot/*`?** | `orig/loomio/app/services/chatbot_service.rb` |
| **Complete counter cache inventory?** | All models with `_count` columns |

### Investigation Commands

```bash
# Find RecordCloner
cat orig/loomio/app/services/record_cloner.rb

# Find translation service
grep -rn "translate" orig/loomio/app/services/

# Find subscription plans
grep -rn "PLANS" orig/loomio/app/services/subscription_service.rb

# Find chatbot publishing
grep -rn "publish.*chatbot" orig/loomio/
grep -rn "chatbot" orig/loomio/app/services/
```

## Low Priority (Nice to Have)

| Question | Investigation Target |
|----------|---------------------|
| **SAML attribute mapping details?** | `orig/loomio/app/controllers/identities/` |
| **Group provisioning via SSO?** | SAML/OAuth controllers |
| **ActionCable vs Socket.io history?** | Git history, ActionCable remnants |
| **Stimulus controller usage?** | `orig/loomio/app/javascript/controllers/` |
| **Turbo/Hotwire usage?** | Search for Turbo imports |
| **API v2 planning?** | routes.rb, deprecation notices |

### Investigation Commands

```bash
# Find SAML/OAuth attribute mapping
grep -rn "saml" orig/loomio/app/controllers/identities/
grep -rn "omniauth" orig/loomio/app/controllers/identities/

# Check for Stimulus
ls orig/loomio/app/javascript/controllers/

# Check for Turbo
grep -rn "turbo" orig/loomio/
```

## Resolved Questions

| Question | Resolution |
|----------|------------|
| **How are Y.js documents persisted?** | SQLite with empty string = ephemeral. Intentional - Rails DB is source of truth. |
| **Is hocuspocus state intentionally ephemeral?** | Yes. Client provides initial content fetched from Rails. |
| **Port 5000 conflict in production?** | Non-issue. Separate containers with isolated port namespaces. |
| **Where does Rails publish to `/records`?** | `MessageChannelService.publish_serialized_records` in `app/services/message_channel_service.rb:17-23` |
| **Where does Rails publish to `/system_notice`?** | `MessageChannelService.publish_system_notice` in `app/services/message_channel_service.rb:25-31` |
| **Where does Rails populate `/current_users/{token}`?** | `BootController.set_channel_token` in `app/controllers/api/v1/boot_controller.rb:25-32` |
| **Poll types count?** | 9 types (not 7). Missing `check` and `question`. |
| **Event kinds count?** | 42 total, 14 webhook-eligible. |
| **Link preview field name?** | `image` not `image_url`. |
| **Volume level behaviors?** | 0=mute (none), 1=quiet (app only), 2=normal (email+app), 3=loud (all+extras) |
| **Hocuspocus auth endpoint?** | POST `/api/hocuspocus` with `{user_secret, document_name}` |
| **Where is `update.sh` script?** | `orig/loomio-deploy/update.sh` |
| **Y.js offline editing?** | IndexedDB fallback via `y-indexeddb` package. |
| **Attachments JSONB default?** | `[]` (empty array). Verified: all 8 occurrences in schema.rb use `default: []`. |
| **Matrix client caching?** | `chatbot/test` creates new client; `chatbot/publish` caches by config key (no eviction). See realtime.md. |

## Outstanding Contradictions

None. All contradictions resolved.

### Guest Migration Details

**FIXME in source code:** `orig/loomio/db/migrate/20240130011619_add_guest_boolean_to_discussion_readers_and_stances.rb`

```ruby
# FIXME add/run migration to convert existing guest records to guest = true
```

**Impact:** Existing guest records in `discussion_readers` and `stances` tables have `guest = false` when they should be `true`. Go implementation should:
1. Be aware of this data inconsistency
2. Consider running a data migration to fix existing records
3. Ensure new guest records are correctly flagged

## Next Investigation Areas

Priority-ordered list for future investigation:

1. **EventBus → Redis mapping** - Critical for real-time feature parity
2. **pg_search configuration** - Required for search functionality
3. **RecordCloner** - Needed for demo/template features
4. **Webhook permissions** - Complete webhook implementation
5. **Subscription tiers** - Billing/feature gating

---
