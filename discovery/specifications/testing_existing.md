# Existing Test Infrastructure - Loomio Rewrite Contract

**Generated: 2026-02-01**
**Source: `/Users/z/Code/loomio/spec/` directory analysis**

---

## Executive Summary

| Metric | Value |
|--------|-------|
| Total Spec Files | 109 |
| Total Spec Lines | ~12,500 (estimated) |
| Factory Definitions | 31 |
| Support Helpers | 3 |
| Shared Examples | 0 (not used) |

---

## 1. Spec Directory Structure

```
spec/
├── benchmarks/
│   └── models/
│       └── group_benchmark.rb (15 lines)
├── controllers/
│   ├── api/
│   │   ├── b2/                       # Bot API v2 tests
│   │   │   ├── discussions_controller_spec.rb (54 lines)
│   │   │   ├── polls_controller_spec.rb (153 lines)
│   │   │   ├── comments_controller_spec.rb (77 lines)
│   │   │   └── memberships_controller_spec.rb (122 lines)
│   │   ├── b3/                       # Bot API v3 tests
│   │   │   └── users_controller_spec.rb (63 lines)
│   │   ├── v1/                       # Primary API tests
│   │   │   ├── announcements_controller_spec.rb (584 lines) [LARGEST]
│   │   │   ├── discussions_controller_spec.rb (994 lines) [LARGEST]
│   │   │   ├── polls_controller_spec.rb (509 lines)
│   │   │   ├── stances_controller_spec.rb (386 lines)
│   │   │   ├── memberships_controller_spec.rb (338 lines)
│   │   │   ├── groups_controller_spec.rb (314 lines)
│   │   │   ├── comments_controller_spec.rb (230 lines)
│   │   │   ├── profile_controller_spec.rb (226 lines)
│   │   │   ├── search_controller_spec.rb (211 lines)
│   │   │   ├── events_controller_spec.rb (178 lines)
│   │   │   ├── documents_controller_spec.rb (156 lines)
│   │   │   ├── outcomes_controller_spec.rb (144 lines)
│   │   │   ├── received_emails_controller_spec.rb (132 lines)
│   │   │   ├── registrations_controller_spec.rb (109 lines)
│   │   │   ├── tasks_controller_spec.rb (99 lines)
│   │   │   ├── tags_controller_spec.rb (96 lines)
│   │   │   ├── membership_requests_controller_spec.rb (95 lines)
│   │   │   ├── sessions_controller_spec.rb (84 lines)
│   │   │   ├── mentions_controller_spec.rb (76 lines)
│   │   │   ├── reactions_controller_spec.rb (65 lines)
│   │   │   ├── attachments_controller_spec.rb (64 lines)
│   │   │   ├── trials_controller_spec.rb (48 lines)
│   │   │   ├── versions_controller_spec.rb (32 lines)
│   │   │   └── login_tokens_controller_spec.rb (30 lines)
│   │   └── hocuspocus_controller_spec.rb (91 lines)
│   ├── identities/
│   │   ├── oauth_controller_spec.rb (417 lines) [COMPREHENSIVE]
│   │   └── saml_controller_spec.rb (398 lines) [COMPREHENSIVE]
│   ├── discussion_mailer_spec.rb (51 lines)
│   ├── discussions_controller_spec.rb (62 lines)
│   ├── email_actions_controller_spec.rb (171 lines)
│   ├── groups_controller_spec.rb (151 lines)
│   ├── login_tokens_controller_spec.rb (21 lines)
│   ├── manifest_controller_spec.rb (12 lines)
│   ├── memberships_controller_spec.rb (131 lines)
│   ├── merge_users_controller_spec.rb (28 lines)
│   ├── poll_mailer_spec.rb (191 lines)
│   ├── polls_controller_spec.rb (58 lines)
│   ├── received_emails_controller_spec.rb (211 lines)
│   ├── redirect_controller_spec.rb (22 lines)
│   └── users_controller_spec.rb (10 lines)
├── extras/
│   ├── queries/
│   │   ├── explore_groups_spec.rb (72 lines)
│   │   ├── users_by_volume_query_spec.rb (93 lines)
│   │   └── users_to_email_query_spec.rb (296 lines) [COMPREHENSIVE]
│   ├── event_bus_spec.rb (73 lines)
│   ├── model_locator_spec.rb (24 lines)
│   ├── range_set_spec.rb (73 lines)
│   ├── time_zone_to_city_spec.rb (11 lines)
│   └── username_generator_spec.rb (26 lines)
├── fixtures/
│   ├── files/           # Test file attachments
│   └── images/          # Test images
├── helpers/
│   ├── email_helper_spec.rb (48 lines)
│   ├── locales_helper_spec.rb (50 lines)
│   └── pretty_url_helper_spec.rb (19 lines)
├── mailboxes/
│   ├── received_email_mailbox_spec.rb (236 lines) [COMPREHENSIVE]
│   └── received_email_mailbox_mixed_spec.rb (58 lines)
├── mailers/
│   └── user_mailer_spec.rb (129 lines)
├── models/
│   ├── ability/
│   │   ├── discussion_spec.rb (94 lines)
│   │   └── poll_spec.rb (220 lines)
│   ├── concerns/
│   │   ├── events/
│   │   │   └── position_spec.rb (153 lines)
│   │   └── has_avatar_spec.rb (46 lines)
│   ├── events/
│   │   ├── comment_replied_to_spec.rb (32 lines)
│   │   ├── group_mentioned_spec.rb (69 lines)
│   │   ├── invitation_accepted_spec.rb (17 lines)
│   │   ├── new_comment_spec.rb (25 lines)
│   │   └── new_coordinator_spec.rb (17 lines)
│   ├── ability_spec.rb (402 lines) [COMPREHENSIVE]
│   ├── comment_spec.rb (85 lines)
│   ├── discussion_event_integration_spec.rb (80 lines)
│   ├── discussion_reader_spec.rb (109 lines)
│   ├── discussion_spec.rb (202 lines)
│   ├── event_spec.rb (375 lines) [COMPREHENSIVE]
│   ├── group_privacy_spec.rb (126 lines)
│   ├── group_spec.rb (153 lines)
│   ├── login_token_spec.rb (27 lines)
│   ├── membership_spec.rb (46 lines)
│   ├── outcome_spec.rb (19 lines)
│   ├── poll_option_spec.rb (30 lines)
│   ├── poll_spec.rb (118 lines)
│   ├── stance_choice_spec.rb (28 lines)
│   ├── stance_spec.rb (44 lines)
│   └── user_spec.rb (191 lines)
├── queries/
│   ├── discussion_query_spec.rb (208 lines)
│   ├── group_query_spec.rb (43 lines)
│   ├── poll_query_spec.rb (39 lines)
│   └── user_query_spec.rb (317 lines) [COMPREHENSIVE]
├── services/
│   ├── group_service/
│   │   └── privacy_change_spec.rb (74 lines)
│   ├── comment_service_spec.rb (159 lines)
│   ├── discussion_reader_service_spec.rb (33 lines)
│   ├── discussion_service_spec.rb (248 lines)
│   ├── event_service_spec.rb (82 lines)
│   ├── group_export_service_spec.rb (152 lines)
│   ├── group_service_spec.rb (127 lines)
│   ├── login_token_service_spec.rb (31 lines)
│   ├── membership_service_spec.rb (130 lines)
│   ├── outcome_service_spec.rb (49 lines)
│   ├── poll_service_spec.rb (350 lines) [COMPREHENSIVE]
│   ├── reaction_service_spec.rb (39 lines)
│   ├── received_email_service_spec.rb (386 lines) [COMPREHENSIVE]
│   ├── record_cloner_spec.rb (114 lines)
│   ├── retry_on_error_spec.rb (24 lines)
│   ├── stance_service_spec.rb (77 lines)
│   ├── task_service_spec.rb (109 lines)
│   ├── throttle_service_spec.rb (57 lines)
│   ├── translation_service_spec.rb (71 lines)
│   └── user_service_spec.rb (143 lines)
├── support/
│   ├── database_cleaner.rb (14 lines)
│   ├── devise.rb (16 lines)
│   └── mailer_macros.rb (9 lines)
├── workers/
│   └── migrate_user_worker_spec.rb (93 lines)
├── factories.rb (317 lines)
└── rails_helper.rb (102 lines)
```

---

## 2. Factory Definitions

**File: `/Users/z/Code/loomio/spec/factories.rb` (317 lines)**

### Core Domain Factories

| Factory | Class | Key Attributes | Lines |
|---------|-------|----------------|-------|
| `:user` | `User` | email, name, password, email_verified: true | 29-40 |
| `:unverified_user` | `User` | email_verified: false | 42-45 |
| `:admin_user` | `User` | is_admin: true | 51-59 |
| `:group` | `Group` | name, description, group_privacy: 'closed', auto-creates admin | 61-76 |
| `:discussion` | `Discussion` | author, group, title, description, private: true | 106-125 |
| `:comment` | `Comment` | user, discussion, body | 132-144 |
| `:poll` | `Poll` | poll_type: 'poll', author, poll_option_names, closing_at | 201-213 |
| `:stance` | `Stance` | poll, participant, reason | 295-299 |
| `:membership` | `Membership` | user, group, accepted_at | 3-7 |
| `:pending_membership` | `Membership` | unverified_user, no accepted_at | 9-12 |

### Poll Type Variations

| Factory | Poll Type | Special Attributes |
|---------|-----------|-------------------|
| `:poll` | `poll` | Single-choice poll |
| `:poll_proposal` | `proposal` | agree/abstain/disagree/block options |
| `:poll_dot_vote` | `dot_vote` | dots_per_person: 8 |
| `:poll_meeting` | `meeting` | Date-based options |
| `:poll_ranked_choice` | `ranked_choice` | minimum_stance_choices: 2 |

### Supporting Factories

| Factory | Class | Purpose |
|---------|-------|---------|
| `:invitation` | `Invitation` | Group invitations |
| `:membership_request` | `MembershipRequest` | Request to join group |
| `:reaction` | `Reaction` | Comment reactions |
| `:event` | `Event` | Activity events |
| `:discussion_event` | `Event` | Discussion-specific events |
| `:notification` | `Notification` | User notifications |
| `:document` | `Document` | Attached documents |
| `:attachment` | `Attachment` | File attachments |
| `:tag` | `Tag` | Group tags |
| `:identity` | `Identity` | OAuth identities |
| `:chatbot` | `Chatbot` | Webhook integrations |
| `:outcome` | `Outcome` | Poll outcomes |
| `:poll_option` | `PollOption` | Poll choice options |
| `:stance_choice` | `StanceChoice` | Stance selections |
| `:poll_template` | `PollTemplate` | Reusable poll templates |
| `:discussion_template` | `DiscussionTemplate` | Reusable discussion templates |
| `:discussion_reader` | `DiscussionReader` | Read state tracking |
| `:login_token` | `LoginToken` | Email login tokens |
| `:version` | `PaperTrail::Version` | Audit history |
| `:translation` | `Translation` | Content translations |
| `:search_result` | `SearchResult` | Search results |
| `:received_email` | `ReceivedEmail` | Inbound emails |

### Factory Callbacks

| Factory | Callback | Purpose |
|---------|----------|---------|
| `:user` | `after(:build)` | Generate username |
| `:group` | `after(:create)` | Create admin user and membership |
| `:discussion` | `before(:create)` | Add author to group/parent |
| `:discussion` | `after(:create)` | Create initial event |
| `:comment` | `before(:create)` | Add commenter to group |
| `:poll` | `after(:create)` | Create initial event |

---

## 3. Test Support Infrastructure

### Database Cleaner Configuration

**File: `/Users/z/Code/loomio/spec/support/database_cleaner.rb`**

```ruby
RSpec.configure do |config|
  config.before(:suite) do
    DatabaseCleaner[:active_record].strategy = :transaction
    DatabaseCleaner[:redis].strategy = :deletion
  end

  config.before(:each) { DatabaseCleaner.start }
  config.after(:each) { DatabaseCleaner.clean }
end
```

**Key Behaviors:**
- ActiveRecord: Transaction rollback (fast, no cleanup)
- Redis: Deletion between tests (clean slate)

### Devise/Controller Helpers

**File: `/Users/z/Code/loomio/spec/support/devise.rb`**

```ruby
module ControllerHelpers
  def sign_in(user = double('user'))
    if user.nil?
      request.env['warden'].stub(:authenticate!).and_throw(:warden, {:scope => :user})
      controller.stub :current_user => nil
    else
      request.env['warden'].stub :authenticate! => user
      controller.stub :current_user => user
    end
  end
end

RSpec.configure do |config|
  config.include Devise::Test::ControllerHelpers, type: :controller
  config.include ControllerHelpers, type: :controller
end
```

### Mailer Helpers

**File: `/Users/z/Code/loomio/spec/support/mailer_macros.rb`**

```ruby
module MailerMacros
  def last_email
    ActionMailer::Base.deliveries.last
  end

  def reset_email
    ActionMailer::Base.deliveries = []
  end
end
```

### Rails Helper Configuration

**File: `/Users/z/Code/loomio/spec/rails_helper.rb`**

Key configurations:
- WebMock for HTTP request stubbing
- Sidekiq::Testing.inline! (synchronous job execution)
- FactoryBot::Syntax::Methods included
- ActiveSupport::Testing::TimeHelpers included
- MailerMacros included
- Default stubs for Chargify, Slack, Facebook, Microsoft, Gravatar

### Helper Methods in rails_helper.rb

| Method | Purpose |
|--------|---------|
| `fixture_for(path)` | Load test fixture file |
| `described_model_name` | Get model name for test subject |
| `emails_sent_to(address)` | Filter emails by recipient |
| `last_email` | Get most recent email |
| `last_email_html_body` | Get HTML body of last email |

---

## 4. Test Coverage Analysis

### Coverage by Domain

| Domain | Spec Files | Total Lines | Coverage Level |
|--------|------------|-------------|----------------|
| API Controllers (v1) | 25 | ~5,100 | HIGH |
| Bot API (b2, b3) | 5 | ~470 | MEDIUM |
| OAuth/SAML | 2 | ~815 | HIGH |
| Models | 23 | ~2,100 | MEDIUM |
| Services | 20 | ~2,600 | HIGH |
| Queries | 4 | ~610 | MEDIUM |
| Mailers | 4 | ~580 | MEDIUM |
| Events | 6 | ~540 | MEDIUM |
| Workers | 1 | 93 | LOW |

### Largest Test Files (by line count)

| File | Lines | Description |
|------|-------|-------------|
| `api/v1/discussions_controller_spec.rb` | 994 | Discussion CRUD, visibility, permissions |
| `api/v1/announcements_controller_spec.rb` | 584 | Notification announcements |
| `api/v1/polls_controller_spec.rb` | 509 | Poll lifecycle, voting |
| `identities/oauth_controller_spec.rb` | 417 | OAuth flow comprehensive tests |
| `models/ability_spec.rb` | 402 | Permission flag combinations |
| `identities/saml_controller_spec.rb` | 398 | SAML SSO tests |
| `api/v1/stances_controller_spec.rb` | 386 | Vote creation/update |
| `services/received_email_service_spec.rb` | 386 | Email parsing/routing |
| `models/event_spec.rb` | 375 | Event notifications |
| `services/poll_service_spec.rb` | 350 | Poll business logic |

---

## 5. Coverage Gaps Identified

### CRITICAL Gaps

| Gap | Description | Priority | Confidence |
|-----|-------------|----------|------------|
| OAuth state parameter | No tests for CSRF protection via state param | CRITICAL | HIGH |
| Rate limit HTTP 429 | No tests for proper 429 response format | CRITICAL | HIGH |
| Bot API auth | Minimal tests for API key authentication | CRITICAL | HIGH |
| Stance 15-min rule | No tests for revision window logic | HIGH | HIGH |

### HIGH Priority Gaps

| Gap | Description | Files Missing |
|-----|-------------|---------------|
| Permission flag combinations | Only 4 of 12 flags tested in ability_spec | ability_spec.rb |
| NullGroup permissions | No tests for direct discussion permissions | None |
| LiveUpdate routing | No tests for pub/sub room routing | None |
| Guest user updates | No tests for individual guest notifications | None |
| Webhook delivery | No tests for chatbot webhook calls | None |
| Webhook retry | No tests for failed delivery retry | None |

### MEDIUM Priority Gaps

| Gap | Description |
|-----|-------------|
| Anonymous poll visibility | Stance visibility in anonymous polls |
| Catch-up email timezone | Email scheduling across timezones |
| Mention deduplication | Multiple mentions of same user on edit |
| Search access control | Visibility filtering in search results |
| Email reply parsing | Complex reply-to address parsing |

### LOW Priority Gaps

| Gap | Description |
|-----|-------------|
| Spam complaint blocking | User suppression after complaints |
| Paper trail tracking | Audit log for permission changes |
| Translation caching | Translation service caching |

---

## 6. Test Data Patterns

### Common Test Setup Patterns

**Group with Members:**
```ruby
let(:group) { create(:group) }
let(:user) { create(:user) }
before { group.add_member!(user) }
```

**Discussion in Group:**
```ruby
let(:discussion) { create(:discussion, group: group, author: user) }
before { group.add_member!(discussion.author) }
```

**Poll with Voters:**
```ruby
let(:poll) { create(:poll, discussion: discussion) }
let(:stance) { create(:stance, poll: poll, participant: user) }
```

### Permission Testing Pattern

```ruby
let(:ability) { Ability::Base.new(user) }
subject { ability }

context "members_can_add_members true" do
  before { group.update_attribute(:members_can_add_members, true) }
  it { should be_able_to(:add_members, group) }
end

context "members_can_add_members false" do
  before { group.update_attribute(:members_can_add_members, false) }
  it { should_not be_able_to(:add_members, group) }
end
```

### OAuth Testing Pattern

```ruby
before do
  stub_const('ENV', ENV.to_hash.merge({
    'OAUTH_AUTH_URL' => 'https://oauth.provider.com/authorize',
    'OAUTH_TOKEN_URL' => 'https://oauth.provider.com/token',
    # ... other OAuth ENV vars
  }))

  stub_request(:post, 'https://oauth.provider.com/token')
    .to_return(status: 200, body: { access_token: 'mock_token' }.to_json)
end
```

### Email Testing Pattern

```ruby
before { ActionMailer::Base.deliveries = [] }

it 'sends notification email' do
  expect { SomeService.create(...) }.to change { ActionMailer::Base.deliveries.count }.by(1)
  expect(last_email.to).to include(user.email)
end
```

---

## 7. Shared Examples Inventory

**Finding: No shared examples defined**

The codebase does not use RSpec shared examples. All test logic is duplicated where needed.

**Recommendation for Rewrite:**
Create shared examples for common patterns:
- Permission flag combinations
- CRUD controller actions
- Event notification flows
- OAuth authentication scenarios

---

## 8. Test Configuration Summary

### External Service Stubs (rails_helper.rb)

| Service | Stub Pattern | Purpose |
|---------|--------------|---------|
| Chargify | `/.chargifypay.com/`, `/.chargify.com/` | Subscription billing |
| Slack | `/slack.com\/api/` | Webhook integration |
| Facebook | `/graph.facebook.com/` | OAuth provider |
| Microsoft | `/api.cognitive.microsoft.com/`, `/api.microsofttranslator.com/` | Translation |
| Outlook | `/outlook.office.com/` | Webhook integration |
| Gravatar | `/www.gravatar.com/` | Avatar lookup |

### Test Database Configuration

**File: `/Users/z/Code/loomio/config/environments/test.rb`**

- Cache store: Redis (separate from production)
- Transactional fixtures: Enabled
- Eager loading: Disabled
- Mailer delivery: Test mode (in-memory)

---

## Appendix: File Reference Table

| Path | Lines | Category |
|------|-------|----------|
| `/Users/z/Code/loomio/spec/factories.rb` | 317 | Factories |
| `/Users/z/Code/loomio/spec/rails_helper.rb` | 102 | Configuration |
| `/Users/z/Code/loomio/spec/support/database_cleaner.rb` | 14 | Support |
| `/Users/z/Code/loomio/spec/support/devise.rb` | 16 | Support |
| `/Users/z/Code/loomio/spec/support/mailer_macros.rb` | 9 | Support |

---

*Document generated: 2026-02-01*
*Analysis scope: 109 spec files, 31 factories, 3 support modules*
