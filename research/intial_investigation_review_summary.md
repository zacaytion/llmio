# Loomio Investigation Review Summary

> Cross-document analysis of all research files for contradictions, inconsistencies, and gaps.
> Generated: 2026-01-31

## Documents Reviewed

| Document | Purpose | Lines |
|----------|---------|-------|
| `loomio_initial_investigation.md` | Rails app: routes, models, API, jobs | ~800 |
| `schema_investigation.md` | Database schema, JSONB structures, indexes | ~600 |
| `initial_investigation_review.md` | Corrections to initial investigations | ~590 |
| `initial_meta_analysis.md` | Cross-document analysis, resolutions | ~400 |
| `loomio_channel_server_initial_investigation.md` | Node.js WebSocket/Hocuspocus server | ~500 |
| `loomio_deploy_initial_investigation.md` | Docker Compose deployment architecture | ~1100 |

---

## 1. Contradictions Between Documents

### 1.1 Attachments JSONB Default - THREE-WAY CONFLICT

| Source | Claim | Line Reference |
|--------|-------|----------------|
| `schema_investigation.md` | `DEFAULT '[]'::jsonb` (empty array) | Line 537 |
| `initial_investigation_review.md` | `DEFAULT '{}'::jsonb` (empty object) | Section 1.3 |
| `initial_meta_analysis.md` | "Verified from schema.rb" but doesn't specify which | Section 1.2 |

**Status:** UNRESOLVED - Need to check actual `schema.rb` or run migration to confirm.

**Impact:** Go struct initialization differs for array vs object default.

---

### 1.2 Poll Types Count

| Source | Count | Types Listed |
|--------|-------|--------------|
| `loomio_initial_investigation.md` | 7 | proposal, poll, count, score, ranked_choice, meeting, dot_vote |
| `initial_investigation_review.md` | 9 | Adds `check`, `question` |

**Status:** RESOLVED - 9 types is correct per `config/poll_types.yml`

**Impact:** Go implementation must support all 9 types.

---

### 1.3 Event Kinds Count

| Source | Count |
|--------|-------|
| `loomio_initial_investigation.md` | ~10 |
| `initial_investigation_review.md` | 42 (complete list provided) |
| Webhook-eligible | 14 |

**Status:** RESOLVED - 42 total, 14 webhook-eligible

**Impact:** Event handler system needs all 42 types.

---

### 1.4 Hocuspocus Token Format

| Source | Format |
|--------|--------|
| `initial_investigation_review.md` | `{user_id},{secret_token}` |
| `loomio_channel_server_initial_investigation.md` | `{user_id},{secret_token}` |
| `loomio_deploy_initial_investigation.md` | Confirms same format |

**Status:** RESOLVED - Consistent across all documents.

---

### 1.5 Hocuspocus SQLite Configuration

| Source | Interpretation |
|--------|----------------|
| `loomio_channel_server_initial_investigation.md` | Notes `database: ''` (empty string) |
| `initial_meta_analysis.md` | Explains this is **intentional** - ephemeral by design |
| `loomio_deploy_initial_investigation.md` | Confirms Rails DB is source of truth |

**Status:** RESOLVED - Not a bug; documents now align on intentional ephemeral storage.

---

### 1.6 Redis Channel Names

| Source | Channel | Notes |
|--------|---------|-------|
| `loomio_channel_server_initial_investigation.md` | `/records`, `/system_notice`, `chatbot/*` | |
| `loomio_deploy_initial_investigation.md` | Same channels | Adds detail on MessageChannelService |

**Status:** CONSISTENT - No contradiction.

---

## 2. Internal Inconsistencies

### 2.1 Volume Level Descriptions

`loomio_initial_investigation.md` says:
> "Values appear to be: 0=mute, 1=quiet, 2=normal, 3=loud (based on Rails code)"

But doesn't explain behavioral differences. `initial_investigation_review.md` Section 2.1 adds:
- `mute` (0): No notifications
- `quiet` (1): App notifications only (no email)
- `normal` (2): Both email and app
- `loud` (3): Maximum engagement + extras

**Status:** CLARIFIED in review doc but not backported to original.

---

### 2.2 Link Preview Structure

| Document | Fields Listed |
|----------|---------------|
| `schema_investigation.md` | `url`, `title`, `description`, `image_url` |
| `initial_investigation_review.md` | Adds `fit`, `align`, `hostname`; notes `image` not `image_url` |

**Status:** CORRECTED in review but inconsistent with original doc.

---

### 2.3 Counter Cache Inventory

Both `loomio_initial_investigation.md` and `schema_investigation.md` mention counter caches but neither provides complete inventory.

`initial_investigation_review.md` Section 5.2 lists 17 counter cache columns on `groups` table alone.

**Status:** INCOMPLETE - Need full inventory across all models.

---

## 3. Unanswered Questions

### 3.1 High Priority (Affects Go Implementation)

| Question | Context | Source |
|----------|---------|--------|
| **How is `pg_search_documents` populated?** | Sync vs async? What triggers reindex? | `initial_investigation_review.md` 4.3 |
| **What are webhook permission values?** | `webhooks.permissions` array contents unknown | `initial_investigation_review.md` 4.2 |
| **How does RecordCloner work?** | Which associations cloned? ID remapping? | `initial_investigation_review.md` 4.6 |
| **What are the exact subscription plan tiers?** | `SubscriptionService::PLANS` not documented | `initial_investigation_review.md` 4.7 |

### 3.2 Medium Priority

| Question | Context |
|----------|---------|
| **How are translations requested/created?** | TranslationService integration unknown |
| **What translation service is used?** | Auto-translation supported? |
| **SAML attribute mapping details?** | How are user fields mapped from SAML assertions? |
| **Group provisioning via SSO?** | Auto-create groups from SSO attributes? |

### 3.3 Resolved Questions

| Question | Resolution | Source |
|----------|------------|--------|
| Hocuspocus document persistence | Ephemeral by design; Rails is source of truth | `initial_meta_analysis.md` |
| Y.js offline editing | IndexedDB fallback via `y-indexeddb` | `loomio_deploy_initial_investigation.md` |
| Boot token expiration | Uses Redis EXPIRE; default from env | `loomio_deploy_initial_investigation.md` |

---

## 4. Missing Documentation

### 4.1 Systems Not Covered

| System | Why It Matters | Partial Info In |
|--------|----------------|-----------------|
| **Email/Mailer System** | 7 mailers, catch-up emails, notification delivery | `initial_investigation_review.md` 3.1 |
| **Rate Limiting (ThrottleService)** | Redis-backed, configurable limits | `initial_investigation_review.md` 3.2 |
| **Demo System** | DemoService, RecordCloner, Redis queue | `initial_investigation_review.md` 3.3 |
| **Search Indexing** | pg_search configuration, triggers | `initial_investigation_review.md` 4.3 |
| **File Storage Backends** | 5 options: Disk, S3, DO Spaces, GCS | `initial_investigation_review.md` 3.5 |

### 4.2 Services Mentioned But Not Documented

From `loomio_initial_investigation.md`'s 46+ services list, these lack detail:

| Service | Inferred Purpose |
|---------|------------------|
| `LinkPreviewService` | URL metadata fetching |
| `RecordCloner` | Deep copy with associations |
| `CleanupService` | Orphan record deletion |
| `TranslationService` | Content translation |
| `GroupExportService` | Data export generation |

### 4.3 Areas Needing Deeper Investigation

1. **ActionCable vs Socket.io** - Rails has ActionCable but deployment uses separate Node.js server. Why? Any ActionCable usage remaining?

2. **Stimulus Controllers** - Vue frontend documented, but Stimulus controllers in `app/javascript/controllers/` not investigated.

3. **Turbo/Hotwire** - Any usage? Would affect Go template rendering approach.

4. **API Versioning** - Only v1 documented. Is there v2 planning? Deprecation notices?

---

## 5. Cross-Repository Alignment

### 5.1 Environment Variable Coverage

| Variable | In `env_template` | In Rails Code | In Channel Server |
|----------|-------------------|---------------|-------------------|
| `SECRET_COOKIE_SECRET` | ✓ | ✓ | ✓ (for auth) |
| `REDIS_URL` | ✓ | ✓ | ✓ |
| `CANONICAL_HOST` | ✓ | ✓ | ✓ |
| `CHANNELS_URI` | ✓ | ✓ (boot_controller) | N/A (is the server) |
| `HOCUSPOCUS_URI` | ✓ | ✓ (boot_controller) | N/A (is the server) |

**Status:** CONSISTENT across all three repos.

### 5.2 Redis Key/Channel Contract

| Key/Channel | Producer | Consumer |
|-------------|----------|----------|
| `/records` | Rails (MessageChannelService) | Channel Server |
| `/system_notice` | Rails (MessageChannelService) | Channel Server |
| `/current_users/{token}` | Rails (BootController) | Channel Server |
| `chatbot/*` | Rails (ChatbotService) | Channel Server |

**Status:** WELL-DOCUMENTED across deploy + channel server investigations.

---

## 6. Document Quality Assessment

| Document | Accuracy | Completeness | Usefulness |
|----------|----------|--------------|------------|
| `loomio_initial_investigation.md` | Medium (corrections needed) | Medium | High |
| `schema_investigation.md` | High (minor errors) | High | High |
| `initial_investigation_review.md` | High | High | Very High |
| `initial_meta_analysis.md` | High | Medium | High |
| `loomio_channel_server_initial_investigation.md` | High | High | Very High |
| `loomio_deploy_initial_investigation.md` | High | High | Very High |

**Recommendation:** Treat `initial_investigation_review.md` and `initial_meta_analysis.md` as authoritative corrections to earlier documents.

---

## 7. Recommendations

### 7.1 Immediate Verification Needed

1. **Attachments JSONB default** - Run `SELECT column_default FROM information_schema.columns WHERE column_name = 'attachments'` to resolve three-way conflict.

2. **Webhook permissions** - Grep for `permissions` usage in webhook-related code.

3. **pg_search triggers** - Check for database triggers or ActiveRecord callbacks.

### 7.2 Additional Investigation Areas

1. **StanceService** - How `latest` boolean is managed atomically.
2. **Email-to-Thread parsing** - Full address format specification.
3. **TaskService** - Task due reminder scheduling.
4. **OAuth/SAML flows** - Complete attribute mapping documentation.

### 7.3 Documentation Consolidation

Consider creating a single "Loomio Architecture Reference" that:
- Consolidates all resolved information
- Removes outdated claims from early investigations
- Provides authoritative Go implementation guidance

---

## 8. Key Takeaways for Go Rewrite

1. **Poll Types**: Support all 9 types, not just 7.
2. **Event Kinds**: Implement all 42 event kinds; only 14 are webhook-eligible.
3. **Attachments**: Verify default before implementing struct.
4. **Hocuspocus**: Ephemeral is intentional; don't try to persist documents.
5. **Counter Caches**: Groups alone has 17; full inventory needed.
6. **Redis Contract**: `/records`, `/system_notice`, `/current_users/{token}` patterns are well-defined.
7. **Volume Levels**: 0-3 with distinct notification behaviors.
8. **Environment Variables**: `env_template` is the source of truth for required config.

---

*Document generated: 2026-01-31*
*Review scope: All files in `research/` directory*
