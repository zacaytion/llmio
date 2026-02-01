# Polls Domain: Tests

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

The polls domain has comprehensive test coverage across models, services, controllers, and queries. Tests use RSpec with FactoryBot for fixtures.

---

## Test Files

| File | Coverage |
|------|----------|
| `/spec/models/poll_spec.rb` | Poll model validations and methods |
| `/spec/models/poll_option_spec.rb` | Poll option behavior |
| `/spec/models/stance_spec.rb` | Stance validation and choice handling |
| `/spec/models/ability/poll_spec.rb` | Poll authorization rules |
| `/spec/services/poll_service_spec.rb` | Poll service operations |
| `/spec/controllers/api/v1/polls_controller_spec.rb` | Poll API endpoints |
| `/spec/controllers/api/v1/stances_controller_spec.rb` | Stance API endpoints |
| `/spec/queries/poll_query_spec.rb` | Poll visibility queries |
| `/spec/controllers/poll_mailer_spec.rb` | Poll email notifications |

---

## Model Tests

### Poll Spec

**File:** `/spec/models/poll_spec.rb`

Tests poll validation and behavior:

#### Validation Tests
- Validates correctly with no poll option changes
- Does not allow changing poll options if template disallows
- Clamps minimum_stance_choices to poll_options length
- Allows closing dates in the future
- Disallows closing dates in the past
- Allows past closing dates if poll is already closed

#### Poll Option Ordering
- Orders by priority for non-meeting polls
- Orders by name (chronologically) for meeting polls

#### Membership Tests
- Members includes guests
- Members includes formal group members

#### Voter Tracking
- Increments voters count when stance created
- Undecided voters increases for uncast stances
- Decided voters increases for cast stances
- Cast votes do not increment undecided

#### Time Zone
- Defaults to author's time zone

### Stance Spec

**File:** `/spec/models/stance_spec.rb`

Tests stance validation:

#### Choice Validation
- Allows no stance choices for meetings/polls (uncast)
- Requires a stance choice for proposals
- Requires minimum number of choices for ranked_choice

#### Reason Validation
- Enforces length limit (500 characters)

#### Choice Shorthand
- Accepts string for single choice
- Accepts array for multiple choices
- Updates stance_counts correctly

---

## Service Tests

### PollService Spec

**File:** `/spec/services/poll_service_spec.rb`

Tests all poll service operations:

#### create_stances
- Creates stance by user ID
- Creates stance by email
- Creates stance by audience
- Avoids duplicate stances for existing voters
- Reinvites revoked users (clears revoked_at)
- Uses normal volume by default
- Uses DiscussionReader volume if set to quiet
- Uses loud volume if reader is loud
- Uses Membership volume if no reader
- Prefers reader volume over membership volume

#### create
- Creates a new poll
- Populates poll options correctly
- Does not create invalid polls (validation)
- Raises AccessDenied for unauthorized users
- Does not send emails on create
- Notifies @mentions in details

#### update
- Updates existing poll
- Raises AccessDenied for unauthorized users
- Does not save invalid changes
- Does not send emails on update
- Creates PollEdited event for option changes
- Creates PollEdited event for major changes (title)

#### close
- Closes a poll (sets closed_at)
- Cannot change anonymous to non-anonymous
- Cannot reveal results early (hide_results validation)

#### Anonymous Poll Closing
- Removes user from stance after close
- Removes user from event after close
- Creates stance receipts before scrubbing

#### Non-Anonymous Closing
- Does not remove user from stance

#### Hide Results
- Hides stance_counts until closed
- Reveals stance_counts after close

#### Stance Creation After Close
- Disallows creating new stances

#### expire_lapsed_polls
- Expires polls past closing_at
- Does not expire active polls
- Does not touch already closed polls

#### group_members_added
- Adds new group members to open polls
- Adds discussion guests to polls
- Does not add bot users
- Does not re-add revoked users

---

## Controller Tests

### Polls Controller Spec

**File:** `/spec/controllers/api/v1/polls_controller_spec.rb`

Tests API endpoints:

#### Index
- Returns polls visible to user
- Filters by group, discussion, status
- Respects visibility rules

#### Show
- Returns poll with associations
- Handles guest access
- Handles member access

#### Create
- Creates poll with valid params
- Rejects unauthorized creation
- Validates required fields

#### Update
- Updates poll with valid params
- Rejects unauthorized updates
- Publishes edit event

#### Close/Reopen
- Closes active polls
- Reopens closed polls
- Validates permissions

### Stances Controller Spec

**File:** `/spec/controllers/api/v1/stances_controller_spec.rb`

Tests voting endpoints:

#### Create
- Creates stance for authorized voter
- Rejects unauthorized voting
- Handles duplicate stance gracefully

#### Update
- Updates existing stance
- Preserves vote history
- Validates poll is still active

#### Revoke
- Removes user access
- Updates poll counts

---

## Query Tests

### PollQuery Spec

**File:** `/spec/queries/poll_query_spec.rb`

Tests visibility logic:

#### visible_to
- Returns polls user authored
- Returns polls in user's groups
- Returns polls with guest access
- Returns public group polls
- Excludes private polls without access
- Excludes discarded polls

#### filter
- Filters by group_key
- Filters by discussion
- Filters by status
- Filters by tags
- Filters by author_id
- Filters by poll_type
- Searches by query

---

## Factory Definitions

**File:** `/spec/factories.rb`

Key poll-related factories:

```ruby
factory :poll do
  title { Faker::Lorem.sentence }
  details { Faker::Lorem.paragraph }
  poll_type { 'poll' }
  closing_at { 1.week.from_now }
  association :author, factory: :user
  association :group
  poll_option_names { ['apple', 'banana', 'orange'] }
end

factory :poll_proposal, parent: :poll do
  poll_type { 'proposal' }
  poll_option_names { %w[agree abstain disagree block] }
end

factory :poll_meeting, parent: :poll do
  poll_type { 'meeting' }
  poll_option_names { ['2020-01-01', '2020-01-02'] }
end

factory :poll_ranked_choice, parent: :poll do
  poll_type { 'ranked_choice' }
  poll_option_names { ['apple', 'banana', 'orange'] }
end

factory :stance do
  association :poll
  association :participant, factory: :user
end

factory :poll_option do
  name { Faker::Lorem.word }
  association :poll
end

factory :outcome do
  statement { Faker::Lorem.paragraph }
  association :poll
  association :author, factory: :user
end
```

---

## Test Helpers

### Dev Routes for E2E

**File:** `/app/controllers/dev/nightwatch_controller.rb`

Provides test scenario setup for E2E tests:

- `setup_poll`: Creates poll in various states
- `setup_closed_poll`: Creates closed poll with stances
- `setup_anonymous_poll`: Creates anonymous poll
- `setup_poll_in_discussion`: Creates poll within thread

### Test Data Patterns

Common test patterns:

```ruby
# Create poll with group and user
let(:group) { create :group }
let(:user) { create :user }
let(:discussion) { create :discussion, group: group }
let(:poll) { create :poll, discussion: discussion }

before { group.add_member!(user) }

# Create stance for testing
let(:stance) { create :stance, poll: poll, participant: user }

# Test voting
stance.choice = 'agree'
StanceService.create(stance: stance, actor: user)
```

---

## Test Categories by Behavior

### Authorization Tests
- User cannot create poll without membership
- User cannot update closed poll
- User cannot close already closed poll
- User cannot reopen anonymous poll
- Guest can vote in open-voting poll
- Admin can close poll early

### State Transition Tests
- Poll closes at closing_at
- Poll can be closed early
- Closed poll can be reopened (if not anonymous)
- Anonymous poll cannot be de-anonymized

### Data Integrity Tests
- Stance counts update correctly
- Voter counts reflect latest stances
- Receipts created on close
- Anonymous stances scrubbed on close

### Notification Tests
- Mentions trigger notifications
- Closing soon notifications sent
- Reminders reach correct audience
- Volume preferences respected

---

## Running Tests

```bash
# All poll tests
bundle exec rspec spec/models/poll_spec.rb spec/services/poll_service_spec.rb spec/controllers/api/v1/polls_controller_spec.rb

# Single test file
bundle exec rspec spec/services/poll_service_spec.rb

# Single test by line
bundle exec rspec spec/services/poll_service_spec.rb:188

# With coverage
COVERAGE=true bundle exec rspec spec/

# Fast mode (skip slow tests)
bundle exec rspec --tag ~slow
```

---

## E2E Tests

**Directory:** `/vue/tests/e2e/specs/`

Nightwatch tests cover:

- Poll creation flow
- Voting interface
- Results display
- Poll closing
- Outcome creation
- Anonymous voting
- Hide results modes

E2E tests use dev routes to set up scenarios, then interact via browser automation.
