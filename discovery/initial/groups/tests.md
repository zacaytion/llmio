# Groups Domain: Tests

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Table of Contents

1. [Test File Locations](#test-file-locations)
2. [GroupService Tests](#groupservice-tests)
3. [MembershipService Tests](#membershipservice-tests)
4. [Group Model Tests](#group-model-tests)
5. [Membership Model Tests](#membership-model-tests)
6. [Key Test Patterns](#key-test-patterns)

---

## Test File Locations

| Test File | Location |
|-----------|----------|
| GroupService spec | `/spec/services/group_service_spec.rb` |
| MembershipService spec | `/spec/services/membership_service_spec.rb` |
| Group model spec | `/spec/models/group_spec.rb` |
| Membership model spec | `/spec/models/membership_spec.rb` |

---

## GroupService Tests

**Location:** `/spec/services/group_service_spec.rb`

### Invite Tests

**Test Setup:**
- User with verified email
- Parent group with subgroup
- Subscription with configurable max_members

**Test Cases:**

1. **Does not mark as accepted if user not in organization**
   - Invites user by email to parent group
   - Expects membership accepted_at to be nil (pending)

2. **Marks membership as accepted if already in parent group**
   - User already member of parent group
   - Invites user to subgroup
   - Expects subgroup membership accepted_at to be set (auto-accepted)

3. **Marks membership as accepted if already in a subgroup**
   - User already member of subgroup
   - Invites user to parent group
   - Expects parent membership accepted_at to be set

4. **Invites a user by email**
   - Creates membership for new email address
   - Verifies membership count increases

5. **Restricts group to subscription.max_members (single)**
   - Sets max_members to 1 (current admin only)
   - Attempts to invite one user
   - Expects `Subscription::MaxMembersExceeded` error
   - Verifies no new membership created

6. **Restricts group to subscription.max_members (multiple)**
   - Sets max_members to 2
   - Attempts to invite two users
   - Expects `MaxMembersExceeded` error
   - Verifies no new memberships created

### Create Tests

1. **Creates a new group**
   - Calls GroupService.create with group and actor
   - Expects Group count to increase by 1
   - Expects creator to be set to actor

### Move Tests

**Test Setup:**
- Group with subscription_id set
- Target parent group
- Site admin user
- Regular user

**Test Cases:**

1. **Moves a group to a parent as an admin**
   - Site admin moves group
   - Expects parent to be set
   - Expects subscription_id to be cleared
   - Expects parent's subgroups to include moved group

2. **Does not allow non-admins to move groups**
   - Regular user attempts move
   - Expects `CanCan::AccessDenied` error

### Merge Tests

**Test Setup:**
- Source and target groups
- Site admin user
- Shared user (member of both)
- Distinct users for each group
- Source has: subgroup, discussion, comment, poll, stance, membership request

**Test Cases:**

1. **Can merge two groups**
   - Merges source into target
   - Verifies all content moved:
     - Subgroups
     - Members (including deduplication)
     - Discussions
     - Polls
     - Membership requests
     - Stances reference new group
     - Comments reference new group
   - Expects source group to be destroyed

2. **Does not allow non-admins to merge**
   - Regular user attempts merge
   - Expects `CanCan::AccessDenied` error

---

## MembershipService Tests

**Location:** `/spec/services/membership_service_spec.rb`

### Test Setup

- Open group with public discussions
- Verified and unverified users
- Admin added to group
- Unverified user with pending membership

### Revoke Tests

**Setup:**
- Subgroup of main group
- Discussion in each group
- Poll in main group
- Active membership for user
- User added as guest to discussions and poll

**Test Cases:**

1. **Cascade deletes memberships**
   - User has: subgroup membership, discussion guest access, poll guest access
   - Revokes main group membership
   - Expects user removed from:
     - Subgroup members
     - Subgroup discussion members
     - Main discussion members
     - Poll members
   - Expects guest flags cleared on readers/stances

2. **Marks discussion readers as revoked**
   - Creates DiscussionReader for user
   - Revokes membership
   - Expects reader to have revoked_at set

### Redeem Tests

**Setup:**
- Another subgroup for testing
- Discussion and poll in main group
- Two inviters (to test inviter tracking)

**Test Cases:**

1. **Sets accepted_at**
   - Redeems unverified user's membership
   - Expects accepted_at to be present

2. **Handles simple case**
   - New membership created
   - Redeems membership
   - Expects: user_id, accepted_at, inviter_id all correct
   - Expects revoked_at nil

3. **Handles existing memberships**
   - User already has accepted membership from first inviter
   - New invitation from second inviter
   - Redeems new invitation
   - Expects original membership preserved
   - Expects accepted_at updated
   - Expects original inviter preserved (not overwritten)

4. **Handles revoked memberships**
   - User has revoked membership from first inviter
   - New invitation from second inviter
   - Redeems new invitation
   - Expects revoked_at cleared
   - Expects new inviter recorded
   - Expects accepted_at set

5. **Unrevokes discussion readers and stances**
   - User has revoked guest access to discussion and poll
   - Redeems membership invitation
   - Expects guest flags cleared
   - Expects revoked_at cleared on both

6. **Notifies the inviter of acceptance**
   - Redeems membership
   - Expects last event to be 'invitation_accepted'

### Alien Group Tests

**Setup:**
- Alien group (outside organization)
- Membership with invited_group_ids including alien group

**Test Case:**

1. **Cannot invite user to alien group**
   - Redeems membership
   - Expects alien group to not include user
   - (Security test: invitation can't add to unrelated groups)

---

## Group Model Tests

**Location:** `/spec/models/group_spec.rb`

### Membership Association Tests

1. **Deletes memberships associated with it**
   - Creates group with member
   - Destroys group
   - Expects membership record to raise `RecordNotFound`

### Subgroup Tests

1. **Subgroup full_name contains parent name**
   - Format: "Parent Name - Subgroup Name"

2. **Updates if parent_name changes**
   - Changes parent name
   - Expects subgroup full_name to reflect new parent name

### Member Management Tests

1. **Can promote existing member to admin**
   - Adds member, then makes admin
   - Expects user in admins list

2. **Can add a member**
   - Adds user to group
   - Expects user in members list

3. **Updates the memberships_count**
   - Adds member
   - Expects count to increase by 1

4. **Sets the first admin to be the creator**
   - New group with no creator
   - Adds first admin
   - Expects creator set to that user

### Privacy Validation Tests

1. **Errors for hidden_from_everyone subgroup with parent_members_can_see_discussions**
   - Invalid combination: secret + parent can see
   - Expects validation error

2. **Does not error for visible to parent subgroup**
   - Valid: visible to parent + parent can see
   - Expects no error

### Discussion Count Tests

1. **Does not count a discarded discussion**
   - Regular and discarded discussion in group
   - Expects counts to exclude discarded:
     - public_discussions_count = 0
     - open_discussions_count = 1
     - closed_discussions_count = 0
     - discussions_count = 1

### Archival Tests

1. **archive! sets archived_at on the group**
   - Archives group
   - Expects archived_at present

2. **unarchive! restores archived_at to nil**
   - Archives then unarchives
   - Expects archived_at nil

### Hierarchy Tests

1. **id_and_subgroup_ids returns empty for new group**
   - Unsaved group
   - Expects empty array

2. **id_and_subgroup_ids returns the id for groups with no subgroups**
   - Saved group without subgroups
   - Expects [group.id]

3. **id_and_subgroup_ids returns id and subgroup ids**
   - Group with subgroup
   - Expects array containing both IDs

### Org Member Count Tests

1. **Returns total number of memberships in the org**
   - Parent group with 2 members
   - Subgroup with 1 additional member
   - Total memberships = 3
   - Unique org members = 2 (creator shared)

---

## Membership Model Tests

**Location:** `/spec/models/membership_spec.rb`

### Validation Tests

1. **Cannot have duplicate memberships**
   - Creates membership for user/group
   - Attempts second membership for same user/group
   - Expects validation error on user_id

### Token Tests

1. **Generates a token on initialize**
   - New membership object
   - Expects token to be present

### Inviter Tests

1. **Can have an inviter**
   - Creates membership with inviter set
   - Expects inviter to be retrievable

### Volume Tests

1. **Responds to volume**
   - Membership with normal volume
   - Expects volume.to_sym == :normal

2. **Can change its volume**
   - Sets volume to quiet
   - Reloads and expects :quiet

---

## Key Test Patterns

### Factory Usage

Tests use FactoryBot factories defined in `/spec/factories.rb`:

```ruby
create(:user)                    # Basic user
create(:user, email_verified: false)  # Unverified user
create(:group)                   # Basic group
create(:group, parent: parent)   # Subgroup
create(:membership, user: user, group: group)
create(:discussion, group: group)
create(:poll, group: group)
create(:stance, poll: poll)
create(:membership_request, group: group)
```

### Common Setup Patterns

**Group with subscription:**
```ruby
let(:group) { create(:group) }
let(:subscription) { Subscription.create(max_members: nil) }
before do
  group.subscription = subscription
  group.save!
end
```

**Group with members:**
```ruby
let(:admin) { create(:user) }
before { group.add_admin!(admin) }
```

**Subgroup structure:**
```ruby
let(:group) { create(:group) }
let(:subgroup) { create(:group, parent: group) }
```

### Assertion Patterns

**Checking membership state:**
```ruby
expect(membership.accepted_at).to be nil       # pending
expect(membership.accepted_at).to be_present   # accepted
expect(membership.revoked_at).to be_present    # revoked
```

**Checking authorization:**
```ruby
expect { service_call }.to raise_error(CanCan::AccessDenied)
```

**Checking subscription limits:**
```ruby
expect { service_call }.to raise_error(Subscription::MaxMembersExceeded)
```

---

## Test Coverage Assessment

### Well Covered Areas

- Invitation flow with auto-accept logic
- Subscription member limits
- Revocation cascade to subgroups
- Redeem handling for various membership states
- Group merge functionality
- Privacy validation rules

### Areas Needing More Coverage

1. **MembershipRequestService:** No dedicated spec file found
2. **Ability modules:** Authorization logic tests not explored
3. **GroupQuery/MembershipQuery:** Query object specs not found
4. **E2E tests:** Frontend behavior tests (Nightwatch) not documented

### Suggested Additional Tests

1. Delegate role assignment and permissions
2. Parent admin making self admin of subgroup
3. Handle generation and uniqueness
4. Privacy mode transitions
5. Volume cascade when apply_to_all is true

---

## Open Questions

1. **Request/Response specs:** Controller tests are not documented - they may exist but weren't found in initial search.

2. **Ability specs:** Authorization rules have complex conditions that would benefit from dedicated tests.

3. **Integration tests:** The relationship between invitation flow and email delivery needs test coverage documentation.

**Confidence Breakdown:**
- Service test coverage: 5/5
- Model test coverage: 4/5
- Test patterns: 4/5
- Coverage gaps identified: 3/5 (may have missed some spec files)
