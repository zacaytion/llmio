# Groups Domain: QA Review

**Reviewed:** 2026-02-01
**Overall Confidence:** 4/5

---

## Table of Contents

1. [Checklist Results](#checklist-results)
2. [Confidence Scores](#confidence-scores)
3. [Issues Found](#issues-found)
4. [Uncertainties](#uncertainties)
5. [Revision Recommendations](#revision-recommendations)

---

## Checklist Results

### Model Documentation (models.md)

| Criteria | Status | Notes |
|----------|--------|-------|
| All attributes listed with types | Partial | Core attributes documented, but missing some like `info` JSONB field, `new_threads_max_depth`, `new_threads_newest_first`, `listed_in_explore`, `can_start_polls_without_discussion` |
| Associations documented | Yes | All major associations covered including `memberships`, `discussions`, `polls`, `subgroups`, `subscription`, `chatbots`, `tags` |
| Validations described | Yes | Name presence/length, handle uniqueness, subscription absence for subgroups, parent inheritance limit documented |
| Callbacks and side effects noted | Partial | Counter cache updates mentioned, but `after_initialize :set_privacy_defaults` and `before_validation :set_discussions_private_only` callbacks not documented |
| Scopes listed | No | Missing scopes like `dangling`, `empty_no_subscription`, `expired_trial`, `any_trial`, `expired_demo`, `not_demo`, `archived`, `published`, `parents_only`, `visible_to_public`, `hidden_from_public`, `mention_search`, `in_organisation`, `explore_search`, `by_slack_team`, `search_for` |
| Concerns/mixins identified | Yes | HasRichText, CustomCounterCache, ReadableUnguessableUrls, SelfReferencing, MessageChannel, GroupPrivacy, HasEvents, Translatable all listed |

**Additional findings:**
- Missing `has_one_attached :cover_photo` and `has_one_attached :logo` Active Storage attachments
- Missing `has_paper_trail` configuration details
- `extend HasTokens` and `extend NoSpam` not mentioned
- Missing `is_translatable on: [:description, :name]` declaration

### Membership Model

| Criteria | Status | Notes |
|----------|--------|-------|
| All attributes listed with types | Yes | Core attributes (token, admin, delegate, title, accepted_at, revoked_at, volume) documented |
| Associations documented | Yes | group, user, inviter, revoker, events all listed |
| Validations described | Partial | Missing `validates_presence_of :group, :user` and `validates_uniqueness_of :user_id, scope: :group_id` |
| Callbacks and side effects noted | Partial | Missing `before_create :set_volume` callback |
| Scopes listed | Partial | Listed active, pending, accepted, revoked, delegates, admin, email_verified. Missing `dangling`, `search_for`, `for_group` |
| Concerns/mixins identified | Partial | HasVolume, HasTimeframe, HasExperiences documented. Missing `has_paper_trail` configuration |

### MembershipRequest Model

| Criteria | Status | Notes |
|----------|--------|-------|
| All attributes listed with types | Yes | Properly documented |
| Associations documented | Yes | group, requestor, responder, user, admins all listed |
| Validations described | Yes | Documented validation for not_in_group_already, unique_membership_request, responder presence |
| Callbacks and side effects noted | N/A | No callbacks present |
| Scopes listed | Yes | pending, responded_to documented. Missing `dangling`, `requested_by` |
| Concerns/mixins identified | Yes | HasEvents mentioned |

### Subscription Model

| Criteria | Status | Notes |
|----------|--------|-------|
| All attributes listed with types | Yes | plan, state, max_members, max_threads, expires_at, renews_at, payment_method, owner documented |
| Associations documented | Partial | Missing `has_many :groups` and `has_paper_trail` |
| Validations described | N/A | No validations present |
| Scopes listed | Yes | active, expired, canceled documented. Missing `dangling` |

---

### Service Documentation (services.md)

| Criteria | Status | Notes |
|----------|--------|-------|
| Public methods with signatures | Yes | All public methods documented with parameters |
| Trigger conditions | Yes | Authorization requirements clearly stated |
| Side effects | Yes | EventBus broadcasts, worker enqueuing documented |
| Events emitted | Yes | All events listed in table format |
| Error conditions | Yes | CanCan::AccessDenied, Subscription exceptions documented |
| Pseudo-code for complex logic | Yes | Invitation and redemption flows diagrammed |

**Additional findings:**
- `PrivacyChange` class is documented as being "not in the services directory" but it actually exists at `/app/services/group_service/privacy_change.rb`
- Missing `MembershipService.add_users_to_group` method
- Missing `MembershipService.save_experience` method
- Missing `MembershipService.redeem_if_pending!` method
- Missing `GroupService.destroy_without_warning!` method

---

### API Endpoint Documentation (controllers.md)

| Criteria | Status | Notes |
|----------|--------|-------|
| HTTP method and path | Yes | All endpoints documented with HTTP verbs and paths |
| Authentication requirements | Partial | `require_signed_in_user_for_explore` mentioned but auth patterns not fully explained |
| Request parameters | Yes | Parameters documented for each endpoint |
| Response structure | Yes | Example JSON responses provided |
| Error responses | Partial | Authorization errors mentioned but HTTP status codes not consistently documented |
| Authorization rules | Yes | Permission requirements listed for each endpoint |

**Additional findings:**
- Missing documentation of the `accept_pending_membership` side effect in `show` action
- GroupQuery.visible_to method usage not fully explained
- `Queries::ExploreGroups` is noted as being in `/app/extras/queries/` which is correct

---

### Frontend/Test Documentation (frontend.md, tests.md)

| Criteria | Status | Notes |
|----------|--------|-------|
| Components documented | Yes | All 26 components listed with purposes |
| Test scenarios covered | Yes | GroupService and MembershipService specs well documented |
| Gaps identified | Yes | MembershipRequestService spec noted as missing |

**Additional findings:**
- E2E tests exist at `/vue/tests/e2e/specs/group.js` but are only briefly mentioned
- Frontend model documentation is comprehensive
- User flows are well documented

---

## Confidence Scores

| Area | Score | Flag |
|------|-------|------|
| Group Model | 4/5 | - |
| Membership Model | 4/5 | - |
| MembershipRequest Model | 5/5 | - |
| Subscription Model | 3/5 | FLAG |
| GroupService | 4/5 | - |
| MembershipService | 4/5 | - |
| MembershipRequestService | 5/5 | - |
| UserInviter | 4/5 | - |
| GroupsController | 4/5 | - |
| MembershipsController | 4/5 | - |
| MembershipRequestsController | 5/5 | - |
| Query Objects | 3/5 | FLAG |
| Ability/Authorization | 3/5 | FLAG |
| Frontend Components | 4/5 | - |
| E2E Tests | 3/5 | FLAG |
| Service Tests | 5/5 | - |
| Model Tests | 4/5 | - |

**Flagged areas requiring attention:**

1. **Subscription Model (3/5):** Limited visibility into billing logic, SubscriptionConcern module loaded conditionally, plan configuration not visible
2. **Query Objects (3/5):** GroupQuery and MembershipQuery exist but internal logic not fully documented
3. **Ability/Authorization (3/5):** Ability::Group module not documented at all in any file
4. **E2E Tests (3/5):** Tests exist but not documented in tests.md

---

## Issues Found

### Factual Errors

1. **PrivacyChange Location:** Documentation states "PrivacyChange class used in GroupService.update is not in the services directory" but it exists at `/app/services/group_service/privacy_change.rb`

2. **FormalGroup/GuestGroup Purpose:** Documentation says these exist for "legacy STI purposes" but the actual model files show they are truly empty subclasses. The documentation could clarify that the `type` column in the groups table is what differentiates them.

3. **NullGroup Location:** Documentation lists both `/app/models/null_group.rb` and `/app/models/concerns/null/group.rb` but should clarify NullGroup inherits from Null::Group concern.

### Omissions

1. **Scopes:** Model documentation is missing most scopes defined on Group, Membership, and MembershipRequest

2. **Active Storage:** Group model has `has_one_attached :cover_photo` and `has_one_attached :logo` which are not documented

3. **Paper Trail:** Both Group and Membership have `has_paper_trail` configurations that are not documented

4. **Ability Module:** `/app/models/ability/group.rb` is not documented anywhere, despite being critical for understanding authorization

5. **Missing Service Methods:**
   - `MembershipService.add_users_to_group`
   - `MembershipService.save_experience`
   - `MembershipService.redeem_if_pending!`
   - `GroupService.destroy_without_warning!`

6. **E2E Tests:** The file `/vue/tests/e2e/specs/group.js` contains 18 test cases that are not documented in tests.md

7. **SubscriptionConcern:** The conditional inclusion `include SubscriptionConcern if Object.const_defined?('SubscriptionConcern')` indicates there may be private/enterprise billing logic not captured

### Inconsistencies

1. **Privacy Cascade Documentation:** The GroupPrivacy concern has detailed validation logic and privacy change handling that is only partially captured in models.md

2. **Invitation Flow:** The diagram in services.md mentions `Events::MembershipCreated.publish!` but the actual code shows specific parameters passed including `recipient_user_ids` and `recipient_message`

---

## Uncertainties

1. **Category Field:** Groups have `category` and `category_id` fields whose purpose is unclear. The GroupSerializer includes a `category` attribute but no documentation explains its use.

2. **GuestGroup Instantiation:** When and where GuestGroup is actually created vs FormalGroup remains unclear from the codebase scan.

3. **Subscription Plans:** The exact plan tiers (trial, demo, and paid plans), their features, and the full SubscriptionService::PLANS hash are not visible in the codebase scan.

4. **ThrottleService Limits:** The exact rate limits for invitations are referenced but not documented.

5. **Explore Groups Query:** The `Queries::ExploreGroups` class exists at `/app/extras/queries/explore_groups.rb` but its filtering logic is not documented.

6. **info JSONB Field:** The Group model uses `self[:info]` for `poll_template_positions`, `categorize_poll_templates`, and `hidden_poll_templates` but this pattern is not documented.

7. **new_host Field:** GroupSerializer includes `new_host` from `object.info['new_host']` but its purpose is not documented.

---

## Revision Recommendations

### High Priority

1. **Add Ability Module Documentation:** Create a new section in models.md or a separate ability.md file documenting:
   - All permission checks (`:show`, `:update`, `:destroy`, `:add_members`, etc.)
   - Permission factors (admin, member, subscription state, group settings)
   - The ability architecture and how modules compose

2. **Document Missing Scopes:** Add comprehensive scope listings to all model documentation with descriptions of what each scope filters

3. **Correct PrivacyChange Location:** Fix the open question about PrivacyChange - it exists at `/app/services/group_service/privacy_change.rb` and handles cascading privacy changes to discussions and subgroups

4. **Add Missing Service Methods:** Document the 4 missing service methods found during review

### Medium Priority

5. **Document Active Storage Attachments:** Add `has_one_attached :cover_photo` and `has_one_attached :logo` to Group model documentation

6. **Add E2E Test Documentation:** Expand tests.md to include the 18 E2E test scenarios from `/vue/tests/e2e/specs/group.js`:
   - Group joining (open, closed, secret)
   - Subgroup creation (open, closed, secret)
   - Group editing
   - Membership volume changes
   - Tag creation
   - Group deletion

7. **Document Query Objects:** Expand controllers.md query objects section with:
   - GroupQuery visibility rules implementation details
   - MembershipQuery search parameter handling
   - ExploreGroups filtering logic

8. **Clarify info JSONB Usage:** Add documentation about the `info` JSONB column pattern used for flexible attributes like `poll_template_positions`, `categorize_poll_templates`, `hidden_poll_templates`, and `new_host`

### Low Priority

9. **Add Paper Trail Configuration:** Document which attributes are tracked by Paper Trail on Group and Membership

10. **Document Callbacks:** Add callbacks section to model documentation covering `after_initialize`, `before_validation`, and `before_create` hooks

11. **Clarify Category Field:** Research and document the `category` and `category_id` fields if they are actively used

12. **Add HTTP Status Codes:** Include expected HTTP status codes for error responses in controller documentation

---

## Summary

The groups domain documentation is comprehensive and well-structured, with good coverage of the core models, services, and user flows. The main gaps are:

1. **Authorization layer not documented** - This is the most significant gap
2. **Scopes largely undocumented** - Important for understanding query patterns
3. **Some service methods missing** - Documentation appears to be from an earlier version
4. **E2E tests not documented** - Good test coverage exists but isn't captured

The documentation accurately captures the overall architecture and patterns. With the recommended revisions, it would reach 5/5 confidence across all areas.
