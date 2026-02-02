# Stance Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/stance.rb`
- `/app/models/stance_choice.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Stance model represents an individual vote/response to a poll. Each user can have multiple stances over time (for vote revision), with only the latest marked as `latest: true`. Stances contain stance_choices linking to specific poll_options with scores.

---

## Vote Revision Rules (Important Gotcha)

**A new stance record is created instead of updating existing when ALL conditions are met:**
1. More than 15 minutes elapsed since last stance
2. Stance choices have changed
3. Poll is attached to a discussion

Otherwise, the existing stance is updated in place.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `poll_id` | integer | - | NO | FK to polls | Parent poll |
| `participant_id` | integer | - | YES | FK to users | Voter |
| `token` | string | auto-generated | YES | UNIQUE | Invitation token |

### Vote Content

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `reason` | string | - | YES | - | Vote explanation |
| `reason_format` | string(10) | "md" | NO | "md" or "html" | Reason format |
| `option_scores` | jsonb | {} | NO | - | Option ID -> score mapping |
| `none_of_the_above` | boolean | false | NO | - | NOTA selection |

### Status & Tracking

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `latest` | boolean | true | NO | - | Most recent stance for user |
| `cast_at` | datetime | - | YES | - | When vote was submitted |
| `revoked_at` | datetime | - | YES | - | If unvoted/revoked |

### Role & Access

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `admin` | boolean | false | NO | - | Poll admin role |
| `guest` | boolean | false | NO | - | Guest voter (not group member) |
| `volume` | integer | 2 | NO | enum 0-3 | Notification volume |

### Invitation

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `inviter_id` | integer | - | YES | FK to users | Who invited |
| `revoker_id` | integer | - | YES | FK to users | Who revoked |
| `accepted_at` | datetime | - | YES | - | Invitation accepted |

### Counters & Metadata

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `versions_count` | integer | 0 | YES | Paper Trail versions |
| `content_locale` | string | - | YES | Content locale |
| `attachments` | jsonb | [] | NO | Rich text attachments |
| `link_previews` | jsonb | [] | NO | Cached link previews |

### Timestamps

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `created_at` | datetime | - | Creation timestamp |
| `updated_at` | datetime | - | Last update |

---

## Validations

All validations only apply when `cast_at` is set (vote has been submitted):

| Validation | Condition | Description |
|------------|-----------|-------------|
| `valid_minimum_stance_choices` | poll.validate_minimum_stance_choices | Min choices met |
| `valid_maximum_stance_choices` | poll.validate_maximum_stance_choices | Max choices not exceeded |
| `valid_max_score` | poll.validate_max_score | No score exceeds max |
| `valid_min_score` | poll.validate_min_score | No score below min |
| `valid_dots_per_person` | poll.validate_dots_per_person | Total dots within limit |
| `valid_reason_length` | poll.limit_reason_length | Reason under 500 chars |
| `valid_reason_required` | poll.stance_reason_required == 'required' | Reason provided if required |
| `valid_require_all_choices` | poll.require_all_choices | All options scored |
| `valid_none_of_the_above` | none_of_the_above checked | NOTA permitted and no other choices |
| `poll_options_must_match_stance_poll` | always | All choice options belong to poll |

**Note:** None-of-the-above bypasses most validations.

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `poll` | Poll | required: true | Parent poll |
| `participant` | User | required: true | Voter |
| `inviter` | User | - | Who sent invitation |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `stance_choices` | StanceChoice | dependent: :destroy | Option selections |
| `poll_options` | PollOption | through: :stance_choices | Selected options |

### Aliases

```ruby
alias user participant
alias author participant
```

### Concern Associations

| Association | Through | Description |
|-------------|---------|-------------|
| `events` | HasEvents | Eventable events |
| `notifications` | HasEvents | Through events |
| `reactions` | Reactable | Emoji reactions |
| `translations` | Translatable | Content translations |
| `tasks` | HasRichText | Embedded tasks |

---

## Scopes

```ruby
scope :latest, -> { where(latest: true, revoked_at: nil) }
scope :guests, -> { where(guest: true) }
scope :admins, -> { where(admin: true) }

scope :decided, -> { where("stances.cast_at IS NOT NULL") }
scope :undecided, -> { where("stances.cast_at IS NULL") }
scope :revoked, -> { where("revoked_at IS NOT NULL") }
scope :none_of_the_above, -> { where(none_of_the_above: true) }
scope :with_reason, -> { where("reason IS NOT NULL AND reason != '' AND reason != '<p></p>'") }

# Sorting
scope :newest_first, -> { order("cast_at DESC NULLS LAST") }
scope :undecided_first, -> { order("cast_at DESC NULLS FIRST") }
scope :oldest_first, -> { order(created_at: :asc) }
scope :priority_first, -> { joins(:poll_options).order('poll_options.priority ASC') }
scope :priority_last, -> { joins(:poll_options).order('poll_options.priority DESC') }

scope :in_organisation, ->(group) {
  joins(:poll).where("polls.group_id": group.id_and_subgroup_ids)
}

# Invitation redemption
scope :redeemable, -> { latest.guests.undecided.where('stances.accepted_at IS NULL') }
scope :redeemable_by, ->(user_id) {
  redeemable.joins(:participant)
    .where("stances.participant_id = ? or users.email_verified = false", user_id)
}

scope :dangling, -> {
  joins('left join polls on polls.id = poll_id').where('polls.id is null')
}
```

**ORDER_SCOPES constant:**
```ruby
ORDER_SCOPES = ['newest_first', 'oldest_first', 'priority_first', 'priority_last']
```

---

## Callbacks

### Before Save
- `assign_option_scores` - Builds option_scores JSONB from stance_choices

### After Save
- `update_versions_count!` - Updates versions_count column

---

## Instance Methods

### Choice Management

```ruby
def choice=(choice)
  self.cast_at ||= Time.zone.now

  if choice.kind_of?(Hash)
    # Hash: option_name => score
    self.stance_choices_attributes = poll.poll_options.where(name: choice.keys).map do |option|
      { poll_option_id: option.id, score: choice[option.name] }
    end
  else
    # Array/single: option names with default score
    options = poll.poll_options.where(name: choice)
    self.stance_choices_attributes = options.map do |option|
      { poll_option_id: option.id }
    end
  end
end

def assign_option_scores
  self.option_scores = build_option_scores
end

def build_option_scores
  stance_choices.map { |sc| [sc.poll_option_id.to_s, sc.score] }.to_h
end

def update_option_scores!
  update_columns(option_scores: assign_option_scores)
end

def score_for(option)
  option_scores[option.id] || 0
end
```

### Replacement Stance (for vote revision)

```ruby
def build_replacement
  Stance.new(
    poll_id: poll_id,
    participant_id: participant_id,
    inviter_id: inviter_id,
    reason_format: reason_format,
    latest: true
  )
end
```

### Participant Access

```ruby
def participant
  # Returns AnonymousUser for anonymous polls
  (!participant_id || poll.anonymous?) ? AnonymousUser.new : super()
end

def real_participant
  # Bypasses anonymity
  User.find_by(id: participant_id)
end

def author_id
  participant_id
end

def user_id
  participant_id
end

def author_name
  participant&.name
end
```

### Discussion Integration

```ruby
def add_to_discussion?
  poll.discussion_id &&
  poll.hide_results != 'until_closed' &&
  !body_is_blank? &&
  !Event.where(eventable: self,
               discussion_id: poll.discussion_id,
               kind: ['stance_created', 'stance_updated']).exists?
end

def parent_event
  poll.created_event
end
```

### Event Methods

```ruby
def created_event_kind
  :stance_created
end

def create_missing_created_event!
  events.create(
    kind: created_event_kind,
    user_id: (poll.anonymous? ? nil : author_id),
    created_at: created_at,
    discussion_id: (add_to_discussion? ? poll.discussion_id : nil)
  )
end
```

### Body Aliases

```ruby
def body
  reason
end

def body_format
  reason_format
end
```

### Status Methods

```ruby
def discarded?
  false  # Stances cannot be soft deleted
end

def locale
  author&.locale || group&.locale || poll.author.locale
end
```

### Delegate Methods

```ruby
%w[group mailer group_id discussion_id discussion members voters tags].each do |message|
  delegate(message, to: :poll)
end
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `CustomCounterCache::Model` | Counter cache updates |
| `HasMentions` | @mention extraction |
| `Reactable` | Emoji reactions |
| `HasEvents` | Event associations |
| `HasCreatedEvent` | Created event tracking |
| `HasVolume` | Notification volume enum |
| `Searchable` | Full-text search |
| `Translatable` | Translation support |
| `HasRichText` | Rich text with sanitization |
| `HasTokens` | Token initialization |

---

## Paper Trail Tracking

Tracked fields:
- `reason`
- `option_scores`
- `revoked_at`
- `revoker_id`
- `inviter_id`
- `attachments`

---

## Search Indexing

```ruby
def self.pg_search_insert_statement(...)
  # Only indexes stances where:
  # - polls.discarded_at IS NULL
  # - stances.cast_at IS NOT NULL (vote submitted)
  # - NOT (anonymous AND not closed)  -- hides anonymous votes until closed
  # - NOT (hide_results=until_closed AND not closed)
end
```

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `poll_id` | INDEX | |
| `participant_id` | INDEX | |
| `token` | UNIQUE | |
| `(poll_id, participant_id, latest)` | UNIQUE PARTIAL | where latest = true |
| `(poll_id, cast_at)` | INDEX | NULLS FIRST |
| `guest` | PARTIAL | where = true |

---

## StanceChoice Model

Individual option selections within a stance.

### Attributes

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `stance_id` | integer | - | YES | FK to stances |
| `poll_option_id` | integer | - | YES | FK to poll_options |
| `score` | integer | 1 | NO | Score/weight for choice |
| `created_at` | datetime | - | YES | |
| `updated_at` | datetime | - | YES | |

### Scopes

```ruby
scope :latest, -> { joins(:stance).merge(Stance.latest) }
```

### Associations

```ruby
belongs_to :stance
belongs_to :poll_option
```

---

## State Diagram

```
                           ┌───────────────────────────────┐
                           │                               │
                           ▼                               │
┌───────────┐   cast   ┌───────────┐   update   ┌───────────────┐
│ UNDECIDED │ ───────> │  DECIDED  │ ─────────> │ NEW_REVISION  │
└───────────┘          └───────────┘            └───────────────┘
     │                      │                          │
     │ revoke               │ revoke                   │
     ▼                      ▼                          │
┌───────────┐          ┌───────────┐                   │
│  REVOKED  │          │  REVOKED  │ <─────────────────┘
└───────────┘          └───────────┘

States:
- UNDECIDED: cast_at IS NULL, latest = true
- DECIDED: cast_at IS NOT NULL, latest = true
- REVOKED: revoked_at IS NOT NULL OR latest = false
- NEW_REVISION: New record created, old stance.latest = false
```

---

## option_scores JSONB Format

```json
{
  "123": 2,    // poll_option_id: score
  "124": 1,
  "125": 0
}
```

Keys are poll_option_id as strings, values are integer scores.

---

## Anonymous Poll Behavior

When `poll.anonymous? == true`:
- `participant` returns `AnonymousUser.new`
- Events are created with `user_id: nil`
- Search indexing is deferred until poll closes
- Voter identities hidden from results
- `real_participant` still returns actual user (for admin operations)

---

## Uncertainties

1. **Vote revision exact conditions** - The 15-minute rule is documented but exact implementation in StanceService should be verified
2. **StanceReceipt model** - Referenced in Poll but not fully documented
3. **Stance.guests scope defined twice** - One as `where(guest: true)`, another as `where('inviter_id is not null')` - need to verify correct behavior

**Confidence Level:** HIGH for core voting, MEDIUM for revision logic details.
