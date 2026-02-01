# Loomio Documentation Uncertainties

**Generated:** 2026-02-01
**Purpose:** Consolidated list of questions, gaps, and areas needing further investigation

---

## Priority Legend

- **HIGH**: Blocks development or could cause production issues
- **MEDIUM**: Important for full understanding, should be resolved
- **LOW**: Nice to know, minor gaps

---

## 1. HIGH Priority Uncertainties

### 1.1 Security Concerns

**OAuth State Parameter Missing** (Auth Domain)
- **Issue:** OAuth controllers do not visibly implement CSRF state parameter validation
- **Risk:** Potential OAuth CSRF vulnerability
- **Status:** Needs investigation to determine if protection exists elsewhere
- **Source:** auth/confidence.md

**Rate Limiting Implementation** (Auth Domain)
- **Issue:** No visible rate limiting on login attempts beyond Devise lockable
- **Risk:** Brute force attacks on authentication endpoints
- **Status:** May exist at infrastructure level (nginx, rack-attack) - needs verification
- **Source:** auth/confidence.md

### 1.2 Code Bugs Discovered

**DiscussionTemplatesController Bug** (Templates Domain)
- **Issue:** Lines 107 and 122 call `DiscussionTemplateService.group_templates()` but this method does not exist
- **Location:** `/app/controllers/api/v1/discussion_templates_controller.rb`
- **Impact:** hide/unhide actions for discussion templates likely produce runtime errors
- **Status:** Needs code fix or verification that feature is unused
- **Source:** templates/confidence.md

**DiscussionExportRelations Foreign Key Bug** (Export Domain)
- **Issue:** `exportable_polls` uses `foreign_key: :group_id` instead of `:discussion_id`
- **Location:** `/app/models/concerns/discussion_export_relations.rb` line 5
- **Impact:** Discussion-level export would export wrong polls
- **Status:** May be dead code - needs verification
- **Source:** export/confidence.md

### 1.3 Missing Test Coverage

**Anonymous Poll Privacy Filtering** (Export Domain)
- **Issue:** No tests verify that anonymous poll data is properly filtered in exports
- **Risk:** Sensitive voter data could be exposed
- **Status:** Needs test coverage added
- **Source:** export/confidence.md

**Chatbot Service Tests** (Integrations Domain)
- **Issue:** No dedicated chatbot spec files exist
- **Impact:** Webhook integration reliability unknown
- **Status:** Tests needed for chatbot creation, webhook delivery, and error handling
- **Source:** integrations/confidence.md

---

## 2. MEDIUM Priority Uncertainties

### 2.1 Architecture Questions

**Real-time Collaboration Sync** (Events Domain)
- **Question:** How does Hocuspocus sync Y.js state with the Rails database?
- **Context:** Collaborative editing uses Yjs CRDT, but persistence mechanism unclear
- **Status:** External service architecture not documented
- **Source:** broad_overview.md, events/confidence.md

**Session Invalidation Propagation** (Auth Domain)
- **Question:** When secret_token is regenerated on logout, how are other sessions invalidated?
- **Context:** User can be logged in on multiple devices
- **Status:** Mechanism unclear - may rely on token comparison
- **Source:** auth/confidence.md

**Token Cleanup Mechanism** (Auth Domain)
- **Question:** Is there a scheduled job for cleaning expired LoginTokens?
- **Context:** Neither documentation nor source review found cleanup job
- **Status:** Potential database bloat issue
- **Source:** auth/confidence.md

**Pagination Not Implemented** (Search Domain)
- **Issue:** Search controller limits to 20 results with no pagination
- **Context:** Commented code suggests pagination was planned
- **Impact:** Users cannot browse beyond first 20 results
- **Source:** search/confidence.md

### 2.2 Permission System Gaps

**Ability::Group Module Not Documented** (Groups Domain)
- **Issue:** Authorization module at `/app/models/ability/group.rb` has no documentation
- **Impact:** Group permission logic unclear
- **Status:** Critical module needs documentation
- **Source:** groups/confidence.md

**show_chatbots Permission** (Integrations Domain)
- **Issue:** Chatbots controller uses `load_and_authorize(:group, :show_chatbots)` but permission not documented
- **Status:** Needs ability module documentation
- **Source:** integrations/confidence.md

### 2.3 Model Documentation Gaps

**Missing Scopes Documentation** (Groups Domain)
- **Issue:** Group, Membership, MembershipRequest models missing scope documentation
- **Scopes missing:** `dangling`, `empty_no_subscription`, `expired_trial`, `archived`, `published`, etc.
- **Status:** Comprehensive scope listing needed
- **Source:** groups/confidence.md

**StanceReceipt Model** (Polls Domain)
- **Issue:** Model used for vote verification not documented
- **Location:** Referenced in services (build_receipts) but no models.md entry
- **Source:** polls/confidence.md

**NullPoll Not Documented** (Polls Domain)
- **Issue:** Null object pattern for polls exists but not mentioned
- **Location:** `/app/models/null_poll.rb`
- **Source:** polls/confidence.md

### 2.4 Service Layer Gaps

**Missing Service Methods** (Groups Domain)
- `MembershipService.add_users_to_group` - Not documented
- `MembershipService.save_experience` - Not documented
- `MembershipService.redeem_if_pending!` - Not documented
- `GroupService.destroy_without_warning!` - Not documented
- **Source:** groups/confidence.md

**UserService.delete_spam_user** (Auth Domain)
- **Issue:** Method referenced in tests but not documented
- **Source:** auth/confidence.md

### 2.5 Frontend Documentation Gaps

**AuthService Implementation** (Auth Domain)
- **Issue:** `/vue/src/shared/services/auth_service.js` not verified
- **Status:** File existence and methods need confirmation
- **Source:** auth/confidence.md

**Strand Components Incomplete** (Events Domain)
- **Issue:** 10 strand components not documented
- **Missing:** `actions_panel.vue`, `load_more.vue`, `members.vue`, `reply_form.vue`, `toc_nav.vue`, `wall.vue`, etc.
- **Source:** events/confidence.md

**FileUploader Return Value** (Documents Domain)
- **Issue:** Documentation states blob returns download_url/preview_url but these come from server
- **Status:** Needs clarification about data origin
- **Source:** documents/confidence.md

---

## 3. LOW Priority Uncertainties

### 3.1 Configuration Questions

**Subscription Plans** (Groups Domain)
- **Question:** What are the exact plan tiers, features, and limits?
- **Context:** SubscriptionService::PLANS hash not visible
- **Source:** groups/confidence.md

**Category Field Purpose** (Groups Domain)
- **Question:** What is the `category` and `category_id` field used for?
- **Context:** GroupSerializer includes but purpose unknown
- **Source:** groups/confidence.md

**info JSONB Field Usage** (Groups Domain)
- **Question:** What is the pattern for Group.info JSONB field?
- **Context:** Stores `poll_template_positions`, `categorize_poll_templates`, `new_host`
- **Source:** groups/confidence.md

### 3.2 Minor Documentation Inaccuracies

**LoginToken Code Generation** (Auth Domain)
- **Issue:** Documentation says "between 100000 and 999999" but code uses `Random.new.rand(999999)` with retry
- **Impact:** Minor - functionally correct
- **Source:** auth/confidence.md

**B3 API Key Length** (Integrations Domain)
- **Issue:** Documentation says "16+" but code requires `length > 16` (17+)
- **Status:** Minor correction needed
- **Source:** integrations/confidence.md

**pg_search Dictionary** (Search Domain)
- **Issue:** Documentation says "Uses 'simple' dictionary" but it's in SQL statements, not initializer
- **Status:** Functionally correct but mechanism different
- **Source:** search/confidence.md

### 3.3 Unverified File Paths

**Frontend File Paths** (Multiple Domains)
- `/vue/src/shared/services/auth_service.js` - Not verified
- `/vue/src/shared/mixins/has_documents.js` - Not verified
- `/vue/src/shared/services/attachment_service.js` - Not verified
- **Source:** Various confidence.md files

### 3.4 Missing Test Coverage (Non-Critical)

**CSV Export Tests** (Export Domain)
- **Issue:** No tests for CSV export functionality
- **Source:** export/confidence.md

**E2E Template Tests** (Templates Domain)
- **Issue:** Nightwatch E2E tests for templates not verified
- **Source:** templates/confidence.md

**E2E Search Tests** (Search Domain)
- **Issue:** No dedicated search E2E tests exist
- **Source:** search/confidence.md

---

## 4. Questions Requiring Codebase Investigation

### 4.1 Architecture Questions

1. **Event Replay:** How are events replayed for timeline construction? What is the sequence_id vs position_key relationship in detail?

2. **RecordCache Invalidation:** What is the cache invalidation strategy for the serialization cache? How does it interact with real-time updates?

3. **Template Versioning:** How do template changes affect existing discussions/polls? Are templates versioned?

4. **Webhook Retry Logic:** What happens when webhook delivery fails? Is there retry logic?

5. **Attachment Download Failures:** During import, what happens if attachment source URL is inaccessible?

### 4.2 Deployment Questions

1. **Environment Variables:** What is the complete list of required vs optional environment variables?

2. **loomio-deploy Relationship:** How does loomio-deploy repo relate to this codebase?

3. **Worker Configuration:** Which Sidekiq workers run on schedules? What is the job queue configuration?

4. **Export Size Limits:** What are the practical limits for group export size? Memory/disk requirements?

### 4.3 Domain-Specific Questions

**Auth:**
- How does frontend detect and handle session expiration?
- What is the pending identity lifecycle and cleanup?

**Groups:**
- When is GuestGroup instantiated vs FormalGroup?
- What are ThrottleService invitation limits?

**Polls:**
- How does ranked choice counting work (Borda vs IRV)?
- What is the exact stance revision 15-minute threshold logic?

**Events:**
- How does SequenceService prevent race conditions?
- What is the full EventBus listener list?

---

## 5. Potential Bugs to Verify

| Issue | Location | Severity | Status |
|-------|----------|----------|--------|
| Missing group_templates method | discussion_templates_controller.rb:107,122 | HIGH | Unverified |
| Wrong foreign key in exportable_polls | discussion_export_relations.rb:5 | MEDIUM | May be dead code |
| Frontend service naming confusion | discussion_template_service.js exports PollTemplateService | LOW | Confirmed but harmless |
| OAuth state parameter missing | identities/base_controller.rb | MEDIUM | Needs security review |

---

## 6. Documentation Revision Priorities

### Immediate (HIGH)
1. Investigate and document OAuth state parameter handling
2. Fix or document discussion templates controller bug
3. Add Ability::Group module documentation
4. Document anonymous poll export privacy filtering

### Soon (MEDIUM)
1. Add comprehensive scope documentation to models
2. Document missing service methods
3. Verify and fix frontend file paths
4. Add StanceReceipt model documentation

### Eventually (LOW)
1. Minor accuracy corrections (code ranges, API key lengths)
2. Complete strand component documentation
3. Add missing E2E test documentation
4. Document subscription plans and limits

---

## Confidence Score Summary by Domain

| Domain | Overall Score | Lowest Area |
|--------|---------------|-------------|
| Auth | 3.8/5 | Frontend (3/5) |
| Groups | 4/5 | Query Objects (3/5), E2E Tests (3/5) |
| Discussions | 4.6/5 | Tests (4/5) |
| Polls | 4.6/5 | Frontend (4/5), Tests (4/5) |
| Events | 4.8/5 | Frontend (4/5) |
| Documents | 4.2/5 | Frontend (3/5) |
| Search | 5/5 | Tests (4/5) |
| Export | 4.4/5 | Services (4/5) |
| Integrations | 4/5 | Tests (3/5) |
| Templates | 4/5 | Tests (3/5) |

---

*End of Uncertainties Document*
