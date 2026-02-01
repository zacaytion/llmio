# Follow-up Investigation Synthesis

**Generated:** 2026-02-01
**Investigations Completed:** 7 of 7
**Overall Confidence:** 4.7/5

---

## Executive Summary

All 7 discrepancies identified in client feedback have been investigated and resolved. Key findings:

1. **CRITICAL SECURITY ISSUE**: OAuth CSRF vulnerability confirmed - custom OAuth implementation lacks state parameter
2. **Documentation corrections**: Multiple counts and classifications corrected across all investigations
3. **Ground truth established**: Definitive answers with code evidence for all questions

---

## Resolution Summary

| Investigation | Priority | Resolution | Confidence |
|--------------|----------|------------|------------|
| OAuth Security | HIGH | **VULNERABLE** - No CSRF protection in OAuth flow | 5/5 |
| Rate Limiting | HIGH | **MULTI-LAYERED** - Both Rack::Attack AND ThrottleService exist | 5/5 |
| OAuth Providers | HIGH | **4 PROVIDERS** - Discovery correct, Research incorrect | 5/5 |
| Permission Flags | MEDIUM | **12 FLAGS** - 11 `members_can_*` + 1 `admins_can_*` | 5/5 |
| Stance Revision | MEDIUM | **15 MINUTES** - Hardcoded threshold confirmed | 5/5 |
| Webhook Events | MEDIUM | **14 UI-EXPOSED** - 16 code-capable, Discovery had 2 errors | 5/5 |
| Attachments JSONB | LOW | **`[]` ARRAY** - Consistent across all tables | 5/5 |

---

## Security Findings (HIGH Priority)

### OAuth CSRF Vulnerability

**Status:** VULNERABILITY CONFIRMED

**Details:**
- Loomio does NOT use OmniAuth gem
- Custom OAuth implementation at `app/controllers/identities/base_controller.rb`
- No `state` parameter in OAuth flow
- Attack vector: Attacker can force victim to link attacker's OAuth account

**Evidence:**
- `app/controllers/identities/base_controller.rb:93-95` - No state parameter in `oauth_params`
- `Gemfile` - No omniauth gems present

**Remediation Required:**
1. Add state parameter generation in `oauth` action
2. Store state in session
3. Validate state in `create` action before processing code
4. Consider changing OAuth initiation from GET to POST

### Rate Limiting Gaps

**Status:** FUNCTIONAL BUT INCOMPLETE

**Strengths:**
- Rack::Attack protects 21 endpoints (10-1000/hour limits)
- ThrottleService handles invitations (500/day trial, 50k/day paid)
- Devise lockable protects login (20 attempts)

**Gaps Identified:**
- `ThrottleService::LimitReached` exception returns 500 instead of 429
- No Retry-After headers
- GET requests unprotected
- Bot API endpoints (`/api/b1/`, `/api/b2/`, `/api/b3/`) not visible in config

---

## Updates to Confirmed Architecture

### OAuth/SSO Providers (Update `feedback/synthesis/authorization.md`)

**Confirmed providers (4):**
1. Google OAuth 2.0
2. Generic OAuth 2.0 (configurable)
3. SAML 2.0
4. Nextcloud OAuth 2.0

**NOT implemented:** Facebook, Slack, Microsoft (Research was incorrect - confused webhook serializers with SSO)

### Permission Flags (Update `feedback/synthesis/authorization.md`)

**Correct count: 12 permission flags**

| Category | Flags |
|----------|-------|
| Member permissions | `members_can_add_members`, `members_can_add_guests`, `members_can_announce`, `members_can_create_subgroups`, `members_can_start_discussions`, `members_can_edit_discussions`, `members_can_edit_comments`, `members_can_delete_comments`, `members_can_raise_motions`, `members_can_vote` (UNUSED) |
| Admin permissions | `admins_can_edit_user_content` |
| Visibility | `parent_members_can_see_discussions` |

**NOT permission flags (configuration):**
- `new_threads_max_depth`
- `new_threads_newest_first`

### Webhook Events (Update `feedback/synthesis/core_models.md`)

**Correct enumeration (14 UI-exposed):**
1. `new_discussion`
2. `discussion_edited`
3. `new_comment`
4. `poll_created`
5. `poll_edited`
6. `poll_closing_soon`
7. `poll_expired`
8. `poll_closed_by_user`
9. `poll_reopened` *(missed by Discovery)*
10. `stance_created`
11. `stance_updated`
12. `outcome_created`
13. `outcome_updated`
14. `outcome_review_due` *(missed by Discovery)*

**Discovery errors corrected:**
- `user_added_to_group` - NOT webhook-eligible
- `membership_requested` - NOT webhook-eligible

### Stance Revision (Add to `feedback/synthesis/core_models.md`)

**Vote revision behavior:**
- 15-minute threshold (hardcoded at `app/services/stance_service.rb:39`)
- Creates new stance only when ALL conditions met:
  1. Previous vote exists
  2. Poll is in a discussion
  3. Vote choices changed
  4. More than 15 minutes since last update
- Otherwise updates in place

### Attachments JSONB (Update `feedback/synthesis/infrastructure.md`)

**Confirmed default:** `[]` (empty array)
- Changed from `{}` to `[]` in migration `20190926001607`
- Consistent across all 8 tables with attachments column

---

## New Discrepancies Discovered

| Discovery | New Finding | Recommendation |
|-----------|-------------|----------------|
| OAuth Security | SAML signature verification disabled (`saml_controller.rb:97-100`) | Review SAML security posture |
| OAuth Security | OAuth tokens stored plaintext in database | Consider encryption at rest |
| Rate Limiting | ThrottleService exception not caught | Add rescue_from handler for 429 |
| Permission Flags | `members_can_vote` column exists but unused | Consider migration to remove |
| Stance Revision | No unit tests for 15-minute threshold | Add test coverage |

---

## Recommendations for Go Rewrite

### Security Patterns to Implement

1. **OAuth CSRF Protection (CRITICAL)**
   - Generate cryptographically random state parameter
   - Store in session, validate on callback
   - Consider PKCE for additional security

2. **Rate Limiting**
   - Implement middleware with configurable limits per endpoint
   - Return proper 429 responses with Retry-After headers
   - Use Redis for distributed rate limiting

### Thresholds and Constants to Preserve

| Constant | Value | Location |
|----------|-------|----------|
| Stance revision threshold | 15 minutes | `StanceService:39` |
| Login attempt limit | 20 | `devise.rb:148` |
| Unlock time | 6 hours | `devise.rb:154` |
| Trial invitation limit | 500/day | `User#invitations_rate_limit` |
| Paid invitation limit | 50,000/day | `User#invitations_rate_limit` |

### Behavior to Replicate Exactly

1. **Stance revision logic** - Create new record only when: is_update AND discussion_id AND option_scores_changed AND updated_at < 15.minutes.ago

2. **Webhook event filtering** - Support both `event_kinds` subscription AND `recipient_chatbot_ids` targeting

3. **Permission flag inheritance** - Subgroups inherit parent permissions unless explicitly overridden

4. **Attachments format** - Always output `[]` for empty attachments, never `null` or `{}`

---

## Files Produced

```
discovery/follow_up/
├── oauth_security/
│   ├── findings.md      # CSRF vulnerability analysis
│   ├── models.md        # Identity model documentation
│   ├── services.md      # OAuth flow documentation
│   └── confidence.md    # 5/5 confidence
├── rate_limiting/
│   ├── findings.md      # Multi-layered rate limiting
│   ├── services.md      # ThrottleService documentation
│   ├── controllers.md   # Protected endpoints
│   └── confidence.md    # 5/5 confidence
├── oauth_providers/
│   ├── findings.md      # 4 providers enumeration
│   ├── models.md        # Provider models
│   ├── controllers.md   # Provider controllers
│   └── confidence.md    # 5/5 confidence
├── permission_flags/
│   ├── findings.md      # 12 flags enumeration
│   ├── models.md        # Group model permissions
│   └── confidence.md    # 5/5 confidence
├── stance_revision/
│   ├── findings.md      # 15-minute threshold
│   ├── services.md      # StanceService logic
│   ├── tests.md         # Test coverage analysis
│   └── confidence.md    # 5/5 confidence
├── webhook_events/
│   ├── README.md        # Quick reference
│   ├── findings.md      # 14/16 event enumeration
│   ├── models.md        # Event model documentation
│   ├── services.md      # ChatbotService documentation
│   ├── payloads.md      # Webhook payload formats
│   └── confidence.md    # 5/5 confidence
├── attachments_jsonb/
│   ├── findings.md      # [] default confirmed
│   └── confidence.md    # 5/5 confidence
└── synthesis.md         # This document
```

---

## Verification Checklist

- [x] Each HIGH priority discrepancy has definitive resolution
- [x] Each MEDIUM priority discrepancy has documented answer
- [x] LOW priority findings documented
- [x] All findings.md files include code evidence (file:line references)
- [x] All confidence.md files have verification checklist
- [x] synthesis.md summarizes cross-investigation findings
- [x] NEW discrepancies flagged for future investigation
