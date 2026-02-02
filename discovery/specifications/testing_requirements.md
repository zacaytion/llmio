# Testing Requirements - Loomio Rewrite Contract

**Generated: 2026-02-01**
**Source: Discovery phase analysis documents**

---

## Executive Summary

This document specifies testing requirements for the Loomio rewrite contract, derived from critical behaviors documented in the discovery phase. Tests are prioritized by security impact and system correctness.

| Priority | Test Categories | Count |
|----------|-----------------|-------|
| CRITICAL | OAuth security, Rate limiting, Bot API auth | 15 |
| HIGH | Permissions, Events, Real-time | 25 |
| MEDIUM | Email, Search, Webhooks | 20 |
| Total | All categories | 60+ |

---

## 1. OAuth Security Tests

**Source: `/Users/z/Code/loomio/discovery/final/oauth_security.md`**

### 1.1 CSRF Protection (State Parameter)

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| OAUTH-01 | OAuth redirect includes state parameter in URL | CRITICAL | HIGH |
| OAUTH-02 | State parameter is cryptographically random (min 32 bytes) | CRITICAL | HIGH |
| OAUTH-03 | State parameter stored in session before redirect | CRITICAL | HIGH |
| OAUTH-04 | Callback rejects missing state parameter | CRITICAL | HIGH |
| OAUTH-05 | Callback rejects mismatched state parameter | CRITICAL | HIGH |
| OAUTH-06 | State cleared from session after successful validation | HIGH | HIGH |

**Test Implementation Pattern:**
```ruby
describe 'OAuth CSRF protection' do
  it 'includes state parameter in authorization URL' do
    get :oauth, params: { back_to: '/dashboard' }
    expect(response.location).to match(/state=[a-zA-Z0-9_-]{32,}/)
    expect(session[:oauth_state]).to be_present
  end

  it 'rejects callback with missing state' do
    get :create, params: { code: 'valid_code' }
    expect(response).to have_http_status(400)
  end

  it 'rejects callback with mismatched state' do
    session[:oauth_state] = 'original_state'
    get :create, params: { code: 'valid_code', state: 'wrong_state' }
    expect(response).to have_http_status(400)
  end
end
```

### 1.2 SSO-Only Mode User Creation

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| OAUTH-07 | SSO-only mode: creates verified user on first login | HIGH | HIGH |
| OAUTH-08 | SSO-only mode: links to unverified user by email | HIGH | HIGH |
| OAUTH-09 | Standard mode: does NOT create user, sets pending_identity | HIGH | HIGH |
| OAUTH-10 | Standard mode: only links to verified users | HIGH | HIGH |
| OAUTH-11 | LOOMIO_SSO_FORCE_USER_ATTRS syncs name/email from SSO | MEDIUM | HIGH |
| OAUTH-12 | User cannot modify name/email when force attrs enabled | MEDIUM | HIGH |

**Existing Coverage:** Partially covered in `/Users/z/Code/loomio/spec/controllers/identities/oauth_controller_spec.rb`

### 1.3 Pending Identity Flow

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| OAUTH-13 | Pending identity stored in server-side session | HIGH | HIGH |
| OAUTH-14 | Pending identity linked to user on subsequent sign-in | HIGH | HIGH |
| OAUTH-15 | Pending identity cleared from session after consumption | MEDIUM | HIGH |
| OAUTH-16 | Pending identity serialized correctly to frontend | MEDIUM | HIGH |

---

## 2. Permission Flags Tests

**Source: `/Users/z/Code/loomio/discovery/final/permission_flags.md`**

### 2.1 All 12 Permission Flag Combinations

| ID | Permission Flag | Test Scenarios | Priority | Confidence |
|----|-----------------|----------------|----------|------------|
| PERM-01 | `members_can_add_members` | member can/cannot invite | HIGH | HIGH |
| PERM-02 | `members_can_edit_discussions` | member can/cannot update | HIGH | HIGH |
| PERM-03 | `members_can_edit_comments` | author can/cannot update | HIGH | HIGH |
| PERM-04 | `members_can_delete_comments` | author can/cannot delete | HIGH | HIGH |
| PERM-05 | `members_can_raise_motions` | member can/cannot create poll | HIGH | HIGH |
| PERM-06 | `members_can_start_discussions` | member can/cannot create | HIGH | HIGH |
| PERM-07 | `members_can_create_subgroups` | member can/cannot add subgroup | HIGH | HIGH |
| PERM-08 | `members_can_announce` | member can/cannot notify | HIGH | HIGH |
| PERM-09 | `members_can_add_guests` | member can/cannot invite guests | HIGH | HIGH |
| PERM-10 | `admins_can_edit_user_content` | admin can/cannot edit others' content | HIGH | HIGH |
| PERM-11 | `parent_members_can_see_discussions` | parent member visibility | HIGH | HIGH |
| PERM-12 | `members_can_vote` | DEPRECATED - verify no effect | LOW | HIGH |

**Required Test Matrix per Flag:**
- [ ] Admin always has permission (flag ignored)
- [ ] Member has permission when flag=true
- [ ] Member lacks permission when flag=false
- [ ] Non-member never has permission
- [ ] Guest user behavior (where applicable)

**Existing Coverage:** Partial in `/Users/z/Code/loomio/spec/models/ability_spec.rb` (402 lines)

### 2.2 NullGroup (Direct Discussion) Permissions

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| PERM-13 | NullGroup.members_can_add_guests returns true | HIGH | HIGH |
| PERM-14 | NullGroup.members_can_edit_discussions returns true | HIGH | HIGH |
| PERM-15 | NullGroup.members_can_add_members returns false | HIGH | HIGH |
| PERM-16 | NullGroup.members_can_start_discussions returns false | HIGH | HIGH |
| PERM-17 | NullGroup.admins_can_edit_user_content returns false | HIGH | HIGH |
| PERM-18 | Direct discussion creator is admin | HIGH | HIGH |
| PERM-19 | Direct discussion guest can be added | HIGH | HIGH |

**Note:** NullGroup has duplicate entries in true_methods and false_methods; true_methods wins due to method definition order.

---

## 3. Rate Limiting Tests

**Source: `/Users/z/Code/loomio/discovery/final/rate_limiting.md`**

### 3.1 ThrottleService Behavior

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| RATE-01 | ThrottleService.can? returns true under limit | HIGH | HIGH |
| RATE-02 | ThrottleService.can? returns false at limit | HIGH | HIGH |
| RATE-03 | ThrottleService.limit! raises LimitReached at limit | HIGH | HIGH |
| RATE-04 | Hourly throttle resets via reset!('hour') | MEDIUM | HIGH |
| RATE-05 | Daily throttle resets via reset!('day') | MEDIUM | HIGH |
| RATE-06 | Throttle scoped by key + id combination | MEDIUM | HIGH |

**Existing Coverage:** Covered in `/Users/z/Code/loomio/spec/services/throttle_service_spec.rb`

### 3.2 HTTP 429 Response Format

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| RATE-07 | LimitReached exception returns HTTP 429 (not 500) | CRITICAL | HIGH |
| RATE-08 | 429 response includes Retry-After header | CRITICAL | HIGH |
| RATE-09 | 429 response body includes error message | HIGH | HIGH |
| RATE-10 | 429 response is JSON format | HIGH | HIGH |

**Implementation Requirement:**
```ruby
# Required rescue_from in snorlax_base.rb
rescue_from(ThrottleService::LimitReached) do |e|
  response.headers['Retry-After'] = '3600'
  render json: { error: 'rate_limit_exceeded', retry_after: 3600 }, status: 429
end
```

### 3.3 Rack::Attack Integration

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| RATE-11 | IP throttling blocks after limit exceeded | HIGH | MEDIUM |
| RATE-12 | Throttle limits per endpoint configurable | MEDIUM | MEDIUM |
| RATE-13 | Static 429.html served for throttled requests | MEDIUM | HIGH |

---

## 4. Bot API Authentication Tests

**Source: `/Users/z/Code/loomio/discovery/final/rate_limiting.md`**

### 4.1 B2 API Authentication

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| BOT-01 | B2 API rejects request without api_key param | CRITICAL | HIGH |
| BOT-02 | B2 API rejects request with invalid api_key | CRITICAL | HIGH |
| BOT-03 | B2 API accepts request with valid user api_key | HIGH | HIGH |
| BOT-04 | B2 API returns 403 for deactivated user | HIGH | HIGH |
| BOT-05 | B2 API respects user's group permissions | HIGH | HIGH |

### 4.2 B3 API Authentication

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| BOT-06 | B3 API rejects missing b3_api_key param | CRITICAL | HIGH |
| BOT-07 | B3 API rejects short B3_API_KEY env (<17 chars) | CRITICAL | HIGH |
| BOT-08 | B3 API rejects mismatched b3_api_key | CRITICAL | HIGH |
| BOT-09 | B3 API accepts valid b3_api_key | HIGH | HIGH |
| BOT-10 | B3 user deactivate/reactivate works correctly | HIGH | HIGH |

**Existing Coverage:** Partial in `/Users/z/Code/loomio/spec/controllers/api/b3/users_controller_spec.rb`

---

## 5. Event and Notification Tests

**Source: `/Users/z/Code/loomio/discovery/final/realtime_pubsub.md`**

### 5.1 Event Type Coverage (42 Types)

All 42 event types must be tested for:
- [ ] Event creation with correct kind
- [ ] Notification recipients (who gets notified)
- [ ] Email recipients (who gets emailed)
- [ ] LiveUpdate broadcast (who gets real-time update)

**Event Types with LiveUpdate (16 types):**

| Event | LiveUpdate | InApp Notify | Priority |
|-------|------------|--------------|----------|
| NewComment | YES | NO | HIGH |
| CommentEdited | YES | NO | HIGH |
| NewDiscussion | YES | YES | HIGH |
| DiscussionEdited | YES | YES | HIGH |
| DiscussionClosed | YES | NO | MEDIUM |
| DiscussionReopened | YES | NO | MEDIUM |
| DiscussionMoved | YES | NO | MEDIUM |
| PollCreated | YES | YES | HIGH |
| PollEdited | YES | YES | HIGH |
| PollClosedByUser | YES | NO | MEDIUM |
| StanceCreated | YES | YES | HIGH |
| StanceUpdated | YES | YES | HIGH |
| OutcomeCreated | YES | YES | HIGH |
| OutcomeUpdated | YES | YES | MEDIUM |
| ReactionCreated | YES | YES | MEDIUM |
| InvitationAccepted | YES | YES | MEDIUM |

**Event Types with InApp Only (18 types):**

| Event | Priority |
|-------|----------|
| CommentRepliedTo | HIGH |
| UserMentioned | HIGH |
| GroupMentioned | HIGH |
| PollAnnounced | HIGH |
| PollClosingSoon | HIGH |
| PollExpired | MEDIUM |
| PollReminder | MEDIUM |
| OutcomeAnnounced | MEDIUM |
| DiscussionAnnounced | MEDIUM |
| MembershipCreated | MEDIUM |
| MembershipRequested | MEDIUM |
| MembershipRequestApproved | MEDIUM |
| NewCoordinator | MEDIUM |
| NewDelegate | MEDIUM |
| UserAddedToGroup | MEDIUM |
| OutcomeReviewDue | LOW |
| PollOptionAdded | LOW |
| UnknownSender | LOW |

**Existing Coverage:** Partial in `/Users/z/Code/loomio/spec/models/event_spec.rb` (375 lines)

### 5.2 LiveUpdate Room Routing

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| LIVE-01 | Group events broadcast to group-{id} room | HIGH | HIGH |
| LIVE-02 | Guest users receive individual user-{id} updates | HIGH | HIGH |
| LIVE-03 | Notifications always route to user-{id} room | HIGH | HIGH |
| LIVE-04 | group_id takes precedence over user_id in routing | HIGH | HIGH |
| LIVE-05 | Events without eventable skip publishing | MEDIUM | HIGH |
| LIVE-06 | Guest iteration uses batched loading | LOW | MEDIUM |

### 5.3 Guest User Individual Updates

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| GUEST-01 | Discussion guests receive NewComment via user-{id} | HIGH | HIGH |
| GUEST-02 | Discussion guests receive StanceCreated via user-{id} | HIGH | HIGH |
| GUEST-03 | Group.guests returns User.none (no guest routing) | MEDIUM | HIGH |
| GUEST-04 | Comment.guests delegates to discussion.guests | MEDIUM | HIGH |

---

## 6. Mention System Tests

**Source: Discovery analysis of mention parsing**

### 6.1 Mention Deduplication on Edit

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| MENT-01 | First mention of user triggers notification | HIGH | HIGH |
| MENT-02 | Re-mentioning same user on edit does NOT notify | HIGH | HIGH |
| MENT-03 | Mentioning new user on edit triggers notification | HIGH | HIGH |
| MENT-04 | Removing mention does not generate notification | MEDIUM | HIGH |
| MENT-05 | HTML mentions use data-mention-id attribute | MEDIUM | HIGH |
| MENT-06 | Markdown mentions use @username pattern | MEDIUM | HIGH |

---

## 7. Email System Tests

**Source: `/Users/z/Code/loomio/discovery/final/email_system.md`**

### 7.1 Email Reply-To Parsing

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| EMAIL-01 | Reply-to format: d={discussion_id}&u={user_id}&k={api_key} | HIGH | HIGH |
| EMAIL-02 | Comment reply includes pt=c&pi={comment_id} | HIGH | HIGH |
| EMAIL-03 | Poll reply includes pt=p&pi={poll_id} | HIGH | HIGH |
| EMAIL-04 | Invalid email_api_key rejects email | HIGH | HIGH |
| EMAIL-05 | User ID mismatch rejects email | HIGH | HIGH |
| EMAIL-06 | Legacy REPLY_HOSTNAME supported | MEDIUM | HIGH |

**Existing Coverage:** Covered in `/Users/z/Code/loomio/spec/services/received_email_service_spec.rb` (386 lines)

### 7.2 Spam Complaint Blocking

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| SPAM-01 | Spam complaint increments user.complaints_count | MEDIUM | HIGH |
| SPAM-02 | Users with complaints excluded from all emails | MEDIUM | HIGH |
| SPAM-03 | Complaint from COMPLAINTS_ADDRESS detected | MEDIUM | HIGH |
| SPAM-04 | no_spam_complaints scope excludes complainers | MEDIUM | HIGH |

### 7.3 Catch-up Email Timezone Handling

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| CATCH-01 | Catch-up sent at 6 AM in user's timezone | MEDIUM | HIGH |
| CATCH-02 | Daily catch-up (email_catch_up_day=7) sent every day | MEDIUM | HIGH |
| CATCH-03 | Weekly catch-up sent on correct weekday (0-6) | MEDIUM | HIGH |
| CATCH-04 | Every-other-day catch-up (day=8) on odd weekdays | MEDIUM | MEDIUM |
| CATCH-05 | User with no timezone defaults correctly | LOW | MEDIUM |

### 7.4 Bounce Email Handling

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| BOUNCE-01 | Bounce notice throttled to 1/hour per sender | MEDIUM | HIGH |
| BOUNCE-02 | Reply to notifications address triggers bounce | MEDIUM | HIGH |
| BOUNCE-03 | ForwardMailer.bounce sent to original sender | LOW | HIGH |

---

## 8. Search Access Control Tests

**Source: Discovery analysis of search visibility**

### 8.1 Search Access Filtering

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| SEARCH-01 | User only sees discussions in their groups | HIGH | HIGH |
| SEARCH-02 | Public discussions visible to all users | HIGH | HIGH |
| SEARCH-03 | Private discussions invisible to non-members | HIGH | HIGH |
| SEARCH-04 | Subgroup discussions respect parent_members_can_see | HIGH | HIGH |
| SEARCH-05 | Archived discussions excluded by default | MEDIUM | HIGH |
| SEARCH-06 | Discarded discussions excluded | MEDIUM | HIGH |

**Existing Coverage:** Covered in `/Users/z/Code/loomio/spec/controllers/api/v1/search_controller_spec.rb` (211 lines)

---

## 9. Webhook Delivery Tests

**Source: `/Users/z/Code/loomio/discovery/final/webhook_events.md`**

### 9.1 Webhook Event Delivery

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| HOOK-01 | Webhook POST sent for configured event_kinds | HIGH | HIGH |
| HOOK-02 | Webhook payload matches serializer format | HIGH | HIGH |
| HOOK-03 | Webhook supports 5 format types (slack, microsoft, discord, markdown, webex) | MEDIUM | HIGH |
| HOOK-04 | Non-200 response logged to Sentry | MEDIUM | HIGH |
| HOOK-05 | Failed webhook retried via Sidekiq (25 attempts) | MEDIUM | MEDIUM |

### 9.2 14 Webhook-Eligible Events

| Event Kind | Must Test Delivery | Priority |
|------------|-------------------|----------|
| new_discussion | YES | HIGH |
| discussion_edited | YES | HIGH |
| poll_created | YES | HIGH |
| poll_edited | YES | HIGH |
| poll_closing_soon | YES | HIGH |
| poll_expired | YES | MEDIUM |
| poll_announced | YES | MEDIUM |
| poll_reopened | YES | MEDIUM |
| outcome_created | YES | HIGH |
| outcome_updated | YES | MEDIUM |
| stance_created | YES | HIGH |
| new_comment | YES | HIGH |
| comment_replied_to | YES | MEDIUM |
| user_mentioned | YES | MEDIUM |

---

## 10. Edge Case Tests

**Source: Discovery phase findings**

### 10.1 Stance Revision 15-Minute Rule

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| EDGE-01 | Stance update within 15 min modifies existing record | HIGH | HIGH |
| EDGE-02 | Stance update after 15 min creates new record | HIGH | HIGH |
| EDGE-03 | Unchanged choices do not create new record | HIGH | HIGH |
| EDGE-04 | Poll without discussion always creates new record | HIGH | MEDIUM |
| EDGE-05 | Multiple revisions tracked with latest=true flag | MEDIUM | HIGH |

**Implementation Reference:**
```ruby
# StanceService.update should check:
# 1. Time since last stance > 15 minutes
# 2. Choices actually changed
# 3. Poll has discussion (revision tracking context)
```

### 10.2 Anonymous Poll Stance Visibility

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| EDGE-06 | Anonymous poll hides voter identity from non-admins | HIGH | HIGH |
| EDGE-07 | Poll author can see all stances in anonymous poll | HIGH | HIGH |
| EDGE-08 | Voter can see their own stance in anonymous poll | HIGH | HIGH |
| EDGE-09 | Results shown without names in anonymous poll | HIGH | HIGH |
| EDGE-10 | Anonymous poll results visible based on results_shown_at | MEDIUM | HIGH |

### 10.3 Direct Discussion (Null Group) Permissions

| ID | Test Case | Priority | Confidence |
|----|-----------|----------|------------|
| EDGE-11 | User can create discussion without group | HIGH | HIGH |
| EDGE-12 | Discussion creator is admin of direct discussion | HIGH | HIGH |
| EDGE-13 | Guests can be added to direct discussion | HIGH | HIGH |
| EDGE-14 | Members cannot "start discussions" from null group | MEDIUM | HIGH |
| EDGE-15 | No membership concept in direct discussion | MEDIUM | HIGH |

---

## 11. Test Implementation Checklist

### Phase 1: Critical Security (Week 1)

- [ ] OAUTH-01 through OAUTH-06 (CSRF protection)
- [ ] RATE-07 through RATE-10 (HTTP 429 responses)
- [ ] BOT-01 through BOT-10 (API authentication)

### Phase 2: Permissions (Week 2)

- [ ] PERM-01 through PERM-12 (all 12 flags)
- [ ] PERM-13 through PERM-19 (NullGroup)
- [ ] EDGE-11 through EDGE-15 (direct discussions)

### Phase 3: Events and Real-time (Week 3)

- [ ] All 42 event types (basic creation)
- [ ] LIVE-01 through LIVE-06 (room routing)
- [ ] GUEST-01 through GUEST-04 (guest updates)

### Phase 4: Email and Webhooks (Week 4)

- [ ] EMAIL-01 through EMAIL-06 (reply parsing)
- [ ] SPAM-01 through SPAM-04 (complaint handling)
- [ ] HOOK-01 through HOOK-05 (webhook delivery)
- [ ] CATCH-01 through CATCH-05 (catch-up emails)

### Phase 5: Edge Cases (Week 5)

- [ ] EDGE-01 through EDGE-05 (stance revision)
- [ ] EDGE-06 through EDGE-10 (anonymous polls)
- [ ] MENT-01 through MENT-06 (mention deduplication)
- [ ] SEARCH-01 through SEARCH-06 (access filtering)

---

## 12. Test Data Requirements

### Minimum Factory Set

| Factory | Required Variants |
|---------|-------------------|
| `:user` | verified, unverified, admin |
| `:group` | public, private, with each permission flag set |
| `:discussion` | in group, direct (no group), public, private |
| `:poll` | each poll type, anonymous, with results_shown_at |
| `:stance` | with choices, with reason, anonymous |
| `:comment` | with parent, with mentions |
| `:event` | each of 42 types |
| `:chatbot` | each webhook_kind |
| `:identity` | oauth, saml, google |

### Test Environment Requirements

| Component | Requirement |
|-----------|-------------|
| Redis | Real instance (not mock) for throttle tests |
| Sidekiq | Inline mode for synchronous testing |
| WebMock | Stub external HTTP requests |
| Time helpers | Freeze/travel time for timezone tests |

---

## Appendix: Confidence Levels

| Level | Definition |
|-------|------------|
| HIGH | Direct code analysis confirms behavior |
| MEDIUM | Inferred from documentation and patterns |
| LOW | Assumed based on standard practices |

---

## Appendix: Source Document References

| Document | Path | Key Tests |
|----------|------|-----------|
| OAuth Security | `/Users/z/Code/loomio/discovery/final/oauth_security.md` | OAUTH-*, SSO tests |
| Permission Flags | `/Users/z/Code/loomio/discovery/final/permission_flags.md` | PERM-*, NullGroup |
| Rate Limiting | `/Users/z/Code/loomio/discovery/final/rate_limiting.md` | RATE-*, BOT-* |
| Webhook Events | `/Users/z/Code/loomio/discovery/final/webhook_events.md` | HOOK-* |
| Realtime Pub/Sub | `/Users/z/Code/loomio/discovery/final/realtime_pubsub.md` | LIVE-*, GUEST-* |
| Email System | `/Users/z/Code/loomio/discovery/final/email_system.md` | EMAIL-*, SPAM-*, CATCH-* |

---

*Document generated: 2026-02-01*
*Total test cases specified: 100+*
*Priority breakdown: CRITICAL (15), HIGH (45), MEDIUM (35), LOW (5)*
