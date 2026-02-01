# Discussions Domain - Tests

**Generated:** 2026-02-01
**Domain:** Test patterns for discussions and comments
**Confidence:** 4/5 (High - Based on test file inspection)

---

## Overview

The discussions domain has tests at multiple levels:
- **Service Specs** - RSpec tests for DiscussionService and CommentService
- **E2E Tests** - Nightwatch browser tests for discussion workflows

---

## Backend Tests (RSpec)

### DiscussionService Spec

**File:** `/spec/services/discussion_service_spec.rb`

#### Test Setup

```
PSEUDO-CODE:
let(:user) - Factory user
let(:another_user) - Factory user
let(:admin) - Factory user
let(:group) - Factory group
let(:another_group) - Factory group (private)
let(:discussion) - Factory discussion with author: user, group: group
let(:poll) - Factory poll in discussion
let(:discussion_params) - {title: "new title", description: "new description", private: true}
```

#### Create Tests

**Authorization:**
- Verifies authorize! is called with :create and discussion

**Email Behavior:**
- Confirms no emails are sent on create (just events)

**Mentions:**
- When description contains @username of group member, creates UserMentioned event
- When mentioning non-member, does NOT create mention event

**Volume Setting:**
- If user has email_on_participation enabled, sets reader volume to 'loud'
- Otherwise, does not set loud volume

**Return Value:**
- Returns Event on success

#### Update Tests

**Authorization:**
- Verifies authorize! is called with :update

**Mentions:**
- New mentions in description trigger notifications
- Editing to add mention to member -> notification created
- Re-editing same mention -> no duplicate notification

**Versioning:**
- Creates Paper Trail version with changed title/description

**Invalid Discussion:**
- Returns false when discussion invalid

#### Update Reader Tests

**Volume:**
- Can save volume attribute to DiscussionReader
- Also updates volume on related Stances
- Raises AccessDenied when user cannot access discussion

#### Move Tests

**Permissions:**
- User must be member of source group (can edit discussion)
- User must be member of destination group

**Privacy Adjustment:**
- When destination is public_only, discussion becomes public
- When destination is private_only, discussion becomes private

**Author Access:**
- Discussion author can move their discussion

**Attribute Protection:**
- Moving does NOT update other attributes like title

**Poll Updates:**
- Polls in discussion get new group assignment

### CommentService Spec

**File:** `/spec/services/comment_service_spec.rb`

#### Test Setup

```
PSEUDO-CODE:
let(:user) - Factory user
let(:another_user) - Factory user in group
let(:group) - Factory group
let(:discussion) - Factory discussion in group by user
let(:comment) - Factory comment by user
let(:reader) - DiscussionReader for user/discussion
```

#### Destroy Tests

**Authorization:**
- Verifies authorize! with :destroy

**Deletion:**
- Calls destroy on comment

**Reply Protection:**
- Raises AccessDenied when comment has replies

#### Create Tests

**Authorization:**
- Verifies authorize! with :create

**Volume:**
- If user has email_on_participation, sets reader volume to 'loud'

**Saving:**
- Calls save! on comment

**Event:**
- Fires Events::NewComment.publish!
- Returns Event on success

**Reader Update:**
- Updates DiscussionReader appropriately

**Mentions with Reply:**
- When replying to comment that mentions you, marks that mention notification as read

#### Create (Invalid) Tests

- Returns false for invalid comment
- Does NOT create NewComment event
- Does NOT update discussion

#### Update Tests

**Content:**
- Updates comment body

**Mention Handling:**
- New mentions trigger notifications
- Re-mentioning same user does NOT re-notify

**Validation:**
- Invalid body prevents update

---

## E2E Tests (Nightwatch)

**File:** `/vue/tests/e2e/specs/discussion.js`

### Public Thread Display

**Test: should_display_content_for_a_public_thread**
```
PSEUDO-CODE:
1. Load open group as visitor (not logged in)
2. Verify group name visible
3. Verify thread preview visible
4. Click thread preview
5. Verify thread heading shows
```

**Test: should_display_timestamps_on_content**
```
PSEUDO-CODE:
1. Load open group as non-member
2. Click thread preview
3. Verify time-ago element exists
```

### Thread State Management

**Test: can_close_and_reopen_a_thread**
```
PSEUDO-CODE:
1. Load setup with open and closed discussions
2. Verify open discussion shows
3. Switch filter to closed
4. Verify closed discussion shows
5. Switch to open, click discussion
6. Open action menu
7. Click close action
8. Verify flash: "Discussion closed"
```

### Editing

**Test: lets_you_edit_title_and_context**
```
PSEUDO-CODE:
1. Load discussion setup
2. Open action menu
3. Click edit action
4. Fill in new title
5. Fill in new description
6. Submit
7. Verify new title/description displayed
```

### Deletion

**Test: lets_coordinators_and_thread_authors_delete_threads**
```
PSEUDO-CODE:
1. Load discussion setup
2. Open action menu
3. Click discard action
4. Confirm deletion
5. Verify flash: "Discussion deleted"
6. Verify redirected to group page
7. Verify discussion NOT in list
```

### Commenting

**Test: adds_a_comment**
```
PSEUDO-CODE:
1. Load discussion setup
2. Fill in comment form
3. Click submit
4. Verify comment text appears
```

**Test: replies_to_a_comment**
```
PSEUDO-CODE:
1. Load discussion setup
2. Add original comment
3. Verify flash
4. Open action menu on comment
5. Click reply action
6. Fill in reply
7. Submit
8. Verify reply appears in strand
```

### Guest Access

**Test: allows_guests_to_comment_and_view_thread_in_dashboard**
```
PSEUDO-CODE:
1. Load discussion as guest
2. Fill in comment
3. Submit
4. Verify flash: "Comment added"
5. Navigate to sidebar recent
6. Verify discussion appears
```

### Joining and Commenting

**Test: allows_logged_in_users_to_join_a_group_and_comment**
```
PSEUDO-CODE:
1. Load open group as non-member
2. Click thread preview
3. Click join group button
4. Verify flash: "You are now a member"
5. Fill in comment
6. Submit
7. Verify flash: "Comment added"
```

### Reactions

**Test: can_react_to_a_discussion**
```
PSEUDO-CODE:
1. Load discussion setup
2. Verify no reactions
3. Click emoji picker
4. Click heart emoji
5. Verify reaction display appears
```

### Mentions

**Test: mentions_a_user_in_wysiwyg**
```
PSEUDO-CODE:
1. Load discussion setup
2. Type @jennifer in comment
3. Verify suggestion list shows
4. Click suggestion
5. Submit
6. Verify mention displayed
```

**Test: mentions_a_user_in_markdown**
```
PSEUDO-CODE:
1. Load discussion setup
2. Switch to markdown mode
3. Accept confirm dialog
4. Type @jennifer
5. Click suggestion
6. Submit
7. Verify @jennifergrey displayed
```

### Comment Actions

**Test: edits_a_comment**
```
PSEUDO-CODE:
1. Load discussion setup
2. Add comment
3. Click edit action
4. Modify text
5. Submit
6. Verify new text displayed
```

**Test: deletes_a_comment**
```
PSEUDO-CODE:
1. Load discussion setup
2. Add comment
3. Open action menu
4. Click discard action
5. Verify comment removed
6. Verify "Item removed" placeholder
```

**Test: discards_restores_deletes_a_comment**
```
PSEUDO-CODE:
1. Add comment
2. Discard comment
3. Verify "Item removed"
4. Open menu on removed item
5. Click undiscard
6. Discard again
7. Click delete (permanent)
8. Confirm
```

### Version History

**Test: lets_you_view_comment_revision_history**
```
PSEUDO-CODE:
1. Load comment with versions
2. Verify current text
3. Click show history action
4. Verify diff shows del/ins changes
```

**Test: lets_you_view_discussion_revision_history**
```
PSEUDO-CODE:
1. Load discussion with versions
2. Open action menu
3. Click show history
4. Verify diff shows del/ins changes
```

### Direct (Private) Discussions

**Test: private_thread**
```
PSEUDO-CODE:
1. Load discussion setup
2. Navigate to private threads
3. Click new thread button
4. Add recipient by email
5. Set title
6. Submit
7. Verify flash: "Discussion started"
8. Add comment
9. Verify comment added
```

### Email Integration

**Test: sign_in_from_discussion_announced_email**
```
PSEUDO-CODE:
1. Load mailer preview
2. Verify headline: "invited you to a discussion"
3. Verify body content
4. Click title link
5. Verify discussion page loads
6. Add comment
7. Verify comment appears
```

**Test: sign_up_from_invitation_created_email**
```
PSEUDO-CODE:
1. Load invitation email preview
2. Click title link
3. Verify email pre-filled
4. Complete signup
5. Verify discussion page loads
6. Verify existing comment visible
```

---

## Test Dev Routes

E2E tests use development routes defined in `Dev::NightwatchController`:

| Route | Purpose |
|-------|---------|
| `setup_discussion` | Create basic discussion with user |
| `setup_discussion_as_guest` | Create discussion with guest access |
| `setup_open_and_closed_discussions` | Create both open and closed discussions |
| `setup_discussion_with_versions` | Create discussion with edit history |
| `setup_comment_with_versions` | Create comment with edit history |
| `setup_unread_discussion` | Create discussion with unread content |
| `view_open_group_as_visitor` | Load public group without auth |
| `view_open_group_as_non_member` | Load public group with different user |

---

## Key Testing Patterns

### Factory Usage

Tests use FactoryBot factories:
- `:user` - Creates user record
- `:group` - Creates group record
- `:discussion` - Creates discussion with author and group
- `:comment` - Creates comment with author and discussion
- `:membership` - Links user to group

### Authorization Testing

Services are tested for proper authorization:
```
PSEUDO-CODE:
user.ability.should_receive(:authorize!).with(:action, resource)
Service.action(resource: resource, actor: user)
```

### Event Testing

Event creation is verified:
```
PSEUDO-CODE:
Events::SomeEvent.should_receive(:publish!).with(args)
Service.action(...)
```

### Notification Testing

Notification side effects checked:
```
PSEUDO-CODE:
expect { Service.action(...) }.to change { user.notifications.count }.by(1)
expect(user.notifications.last.kind).to eq 'some_kind'
```

### E2E Helpers

PageHelper provides test utilities:
- `loadPath(route)` - Navigate to dev route
- `expectText(selector, text)` - Assert text content
- `fillIn(selector, value)` - Enter text
- `click(selector)` - Click element
- `expectFlash(message)` - Verify flash notification
- `expectElement(selector)` - Assert element exists
- `expectNoText(selector, text)` - Assert text not present
