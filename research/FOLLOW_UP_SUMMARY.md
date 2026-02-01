# Follow-up Investigation Summary

This document serves as an entry point to the `follow_up/` directory, which contains 7 investigation briefs documenting discrepancies between your discovery findings and our internal research. Each brief includes specific codebase targets for verification.

## How to Use This Document

1. **Review the Priority Matrix** below to identify which investigations matter most
2. **Read the relevant brief** in `follow_up/` for full context
3. **Use the investigation targets** in each brief as starting points for LLM-assisted codebase exploration
4. Each brief includes Rails documentation context to help interpret what you find

---

## Priority Matrix

| Priority | File | Discrepancy | First Files to Check |
|----------|------|-------------|---------------------|
| **HIGH** | `oauth_security.md` | CSRF state parameter validation may be missing | `Gemfile` (check for `omniauth-rails_csrf_protection`), `config/initializers/omniauth.rb` |
| **HIGH** | `rate_limiting.md` | Conflicting documentation - we claim ThrottleService exists, you found none | `app/services/throttle_service.rb`, `config/initializers/rack_attack.rb`, `Gemfile` |
| **HIGH** | `oauth_providers.md` | Provider lists differ (you: 4, us: 5) | `app/controllers/identities/`, `config/routes.rb:387-420` |
| **MEDIUM** | `permission_flags.md` | Flag count differs (you: 10, us: 11) | `app/models/group.rb:409-492`, `db/schema.rb` |
| **MEDIUM** | `stance_revision_threshold.md` | 15-minute vote revision window you documented is missing from our research | `app/services/stance_service.rb` |
| **MEDIUM** | `webhook_eligible_events.md` | Neither of us enumerated the 14 webhook-eligible events | `app/models/event.rb`, `app/models/concerns/events/notify/chatbots.rb` |
| **LOW** | `attachments_jsonb_default.md` | Conflicting info on JSONB default (`[]` vs `{}`) | `db/schema.rb`, `db/migrate/*attachments*` |

---

## Quick Reference by Topic

### Security Issues (Investigate First)
- `oauth_security.md` - Potential CSRF vulnerability in OAuth flow
- `rate_limiting.md` - Unclear if rate limiting is implemented

### Authentication & Authorization
- `oauth_providers.md` - Which OAuth providers are actually available?
- `oauth_security.md` - Is OmniAuth properly configured?

### Polls & Voting
- `stance_revision_threshold.md` - Your finding about 15-minute vote revision window needs our verification

### Events & Webhooks
- `webhook_eligible_events.md` - Complete enumeration of 14 webhook events needed

### Data Models
- `permission_flags.md` - Exact count and classification of group permission flags
- `attachments_jsonb_default.md` - Minor data initialization detail

---

## Brief Format

Each investigation brief follows this structure:

```
# [Topic] - Follow-up Investigation Brief

## Discrepancy Summary
What differs between findings

## Discovery Claims
Your documented findings (with file references)

## Our Research Claims
Our documented findings (with file references)

## Ground Truth Needed
Specific questions to answer

## Investigation Targets
- [ ] File paths and grep commands to run
- [ ] Specific lines/methods to examine

## Priority
HIGH/MEDIUM/LOW with rationale

## Rails Context
Relevant Rails patterns and documentation
```

---

## Recommended Investigation Order

1. **Security first**: `oauth_security.md` and `rate_limiting.md`
2. **Authentication**: `oauth_providers.md` (related to security)
3. **Core mechanics**: `stance_revision_threshold.md`, `webhook_eligible_events.md`
4. **Minor details**: `permission_flags.md`, `attachments_jsonb_default.md`

---

## What We Learned

Overall, our document sets are **highly consistent** on core architecture. The discrepancies are mostly:
- **Coverage gaps**: One team documented something the other missed
- **Classification differences**: How to categorize certain settings
- **Security review depth**: Your team flagged potential issues we didn't examine

Your discovery documentation excels at **operational flow details** (service methods, validation rules, template names). Our research excels at **schema-level analysis** (JSONB structures, index strategies). The combination is more complete than either alone.
