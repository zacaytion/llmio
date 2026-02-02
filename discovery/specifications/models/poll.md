# Poll Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/poll.rb`
- `/app/models/poll_option.rb`
- `/app/models/outcome.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Poll model represents decision-making tools with multiple poll types (proposal, poll, count, score, dot_vote, ranked_choice, meeting). Polls have options, collect stances (votes), and produce outcomes. Polls support rich configuration for voting behavior, anonymity, and result display.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `title` | string | - | NO | presence required | Poll title |
| `details` | text | - | YES | max via AppConfig | Rich text description |
| `details_format` | string(10) | "md" | NO | "md" or "html" | Details format |
| `key` | string | - | NO | UNIQUE | Public URL key (8 chars) |
| `poll_type` | string | - | NO | inclusion in AppConfig.poll_types | Type identifier |

### Relationships

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `author_id` | integer | - | NO | FK to users | Poll creator |
| `discussion_id` | integer | - | YES | FK to discussions | Parent discussion |
| `group_id` | integer | - | YES | FK to groups | Container group |

### Time & Status

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `closing_at` | datetime | - | YES | must be future if set | Scheduled close time |
| `closed_at` | datetime | - | YES | - | Actual close time |
| `discarded_at` | datetime | - | YES | - | Soft delete timestamp |
| `discarded_by` | integer | - | YES | FK to users | Who deleted |
| `created_at` | datetime | - | YES | - | |
| `updated_at` | datetime | - | YES | - | |

### Template Settings

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `template` | boolean | false | NO | - | Is a template |
| `poll_template_id` | integer | - | YES | FK | Source template |
| `poll_template_key` | string | - | YES | - | Template key reference |

### Voting Configuration

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `anonymous` | boolean | false | - | cannot change to false | Hide voter identities |
| `hide_results` | integer | 0 | - | enum | When to show results |
| `voter_can_add_options` | boolean | false | - | - | Voters can add options |
| `specified_voters_only` | boolean | false | - | - | Limit to invited voters |
| `shuffle_options` | boolean | false | - | - | Randomize option order |
| `multiple_choice` | boolean | false | - | - | Multiple selections |
| `show_none_of_the_above` | boolean | false | - | - | NOTA option available |

**hide_results enum:**
- 0: off - Always show results
- 1: until_vote - Show after voting
- 2: until_closed - Show after poll closes

### Scoring Configuration

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `min_score` | integer | - | YES | Minimum score value |
| `max_score` | integer | - | YES | Maximum score value |
| `minimum_stance_choices` | integer | - | YES | Min selections required |
| `maximum_stance_choices` | integer | - | YES | Max selections allowed |
| `dots_per_person` | integer | - | YES | Dot voting allocation |
| `agree_target` | integer | - | YES | Target agree percentage |
| `quorum_pct` | integer | - | YES | Quorum percentage (0-100) |

### Notification Settings

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `notify_on_closing_soon` | integer | 0 | - | enum | When to notify |
| `stance_reason_required` | integer | 1 | - | enum | Reason requirement |
| `limit_reason_length` | boolean | true | - | - | Cap reason at 500 chars |

**notify_on_closing_soon enum:**
- 0: nobody
- 1: author
- 2: undecided_voters
- 3: voters

**stance_reason_required enum:**
- 0: disabled - Hide reason field
- 1: optional - Show but optional
- 2: required - Must provide reason

### Aggregated Data

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `voters_count` | integer | 0 | Total stances (invited voters) |
| `undecided_voters_count` | integer | 0 | Stances without cast_at |
| `none_of_the_above_count` | integer | 0 | NOTA votes |
| `stance_counts` | jsonb | [] | Per-option total scores |
| `stance_data` | jsonb | {} | Aggregated stance data |
| `matrix_counts` | jsonb | [] | Matrix poll data |

### Display Settings

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `chart_type` | string | - | Visualization type |
| `poll_option_name_format` | string | - | Option display format |
| `reason_prompt` | string | - | Custom reason prompt |
| `process_name` | string | - | Process type name |
| `process_subtitle` | string | - | Process subtitle |
| `process_url` | string | - | Process documentation URL |
| `default_duration_in_days` | integer | - | Template default duration |

### Other Fields

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `versions_count` | integer | 0 | Paper Trail versions |
| `content_locale` | string | - | Content locale |
| `tags` | string[] | [] | Tag array |
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |
| `custom_fields` | jsonb | {} | Extensible metadata |

**custom_fields structure:**
```json
{
  "meeting_duration": 60,
  "time_zone": "Pacific/Auckland",
  "can_respond_maybe": true
}
```

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `poll_type` | inclusion in AppConfig.poll_types.keys | always |
| `details` | max length via AppConfig | always |
| `title` | presence required | unless discarded |
| `title`, `details` | no spam regex | NoSpam concern |
| `closing_at` | must be in future | if not closed |
| `anonymous` | cannot change from true to false | once set |
| `hide_results` | cannot change from 'until_closed' | once set |
| `group_id` | must match discussion.group_id | if both set |
| `quorum_pct` | normalized to 0-100 | always |
| `minimum_stance_choices` | clamped to poll_options.length | always |

**Custom Validations:**
```ruby
validate :closes_in_future
validate :discussion_group_is_poll_group
validate :cannot_deanonymize
validate :cannot_reveal_results_early
validate :title_if_not_discarded
```

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `author` | User | - | Poll creator |
| `discussion` | Discussion | - | Parent discussion |
| `group` | Group | - | Container group |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `outcomes` | Outcome | dependent: :destroy | Poll outcomes |
| `stances` | Stance | dependent: :destroy | All stances |
| `stance_choices` | StanceChoice | through: :stances | All choices |
| `poll_options` | PollOption | `-> { order('priority') }`, autosave | Poll options |
| `documents` | Document | as: :model, dependent: :destroy | Attached documents |
| `stance_receipts` | StanceReceipt | dependent: :destroy | Receipt records |

### Has One

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `current_outcome` | Outcome | `-> { where(latest: true) }` | Latest outcome |

### Voter Associations (through stances)

| Association | Filter | Description |
|-------------|--------|-------------|
| `voters` | `Stance.latest` | All latest stances |
| `admin_voters` | `Stance.latest.admin` | Admin stances |
| `undecided_voters` | `Stance.latest.undecided` | No cast_at |
| `decided_voters` | `Stance.latest.decided` | Has cast_at |
| `none_of_the_above_voters` | `Stance.latest.none_of_the_above` | NOTA votes |

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
scope :active, -> { kept.where('polls.closed_at': nil) }
scope :template, -> { kept.where('polls.template': true) }
scope :closed, -> { kept.where("polls.closed_at IS NOT NULL") }
scope :recent, -> { kept.where("polls.closed_at IS NULL or polls.closed_at > ?", 7.days.ago) }

scope :search_for, ->(fragment) { kept.where("polls.title ilike :fragment", fragment: "%#{fragment}%") }
scope :lapsed_but_not_closed, -> { active.where("polls.closing_at < ?", Time.now) }
scope :active_or_closed_after, ->(since) { kept.where("polls.closed_at IS NULL OR polls.closed_at > ?", since) }
scope :in_organisation, ->(group) { kept.where(group_id: group.id_and_subgroup_ids) }

scope :closing_soon_not_published, ->(timeframe, recency_threshold = 24.hours.ago) {
  active.distinct
    .where(closing_at: timeframe)
    .where("NOT EXISTS (SELECT 1 FROM events
            WHERE events.created_at > ? AND
                  events.eventable_id = polls.id AND
                  events.eventable_type = 'Poll' AND
                  events.kind = 'poll_closing_soon')", recency_threshold)
}

scope :dangling, -> {
  joins('left join groups g on polls.group_id = g.id')
    .where('group_id is not null and g.id is null')
}
```

---

## Callbacks

### Before Validation
- `clamp_minimum_stance_choices` - Ensures min choices doesn't exceed options count

---

## Instance Methods

### Status Methods

```ruby
def active?
  kept? && (closing_at && closing_at > Time.now) && !closed_at
end

def wip?
  closing_at.nil?  # Work in progress - not yet scheduled
end

def closed?
  !!closed_at
end

def show_results?(voted: false)
  case hide_results
  when 'until_closed'
    closed_at
  when 'until_vote'
    closed_at || voted
  else
    true
  end
end
```

### Voter Methods

```ruby
def decided_voters_count
  voters_count - undecided_voters_count
end

def cast_stances_pct
  return 0 if voters_count == 0
  ((decided_voters_count.to_f / voters_count) * 100).to_i
end

# Anonymous polls return User.none
def undecided_voters
  anonymous? ? User.none : super
end

def decided_voters
  anonymous? ? User.none : super
end

# Bypass anonymity for admin operations
def unmasked_voters
  User.where(id: stances.latest.pluck(:participant_id))
end

def unmasked_undecided_voters
  User.where(id: stances.latest.undecided.pluck(:participant_id))
end

def unmasked_decided_voters
  User.where(id: stances.latest.decided.pluck(:participant_id))
end
```

### Member/Admin Methods

```ruby
def admins
  # Complex query returning users who are:
  # - Group admins
  # - Poll author (and group member)
  # - Poll author (no group)
  # - Poll author (discussion guest)
  # - Discussion admin (group member)
  # - Discussion guest admin (not group member)
  # - Poll admin (group member)
  # - Poll admin guest
end

def members
  # Returns users who can read poll:
  # - Discussion guests
  # - Group members
  # - Poll guest voters
end

def existing_member_ids
  voter_ids
end
```

### Guest Management

```ruby
def add_guest!(user, author)
  stances.create!(
    participant_id: user.id,
    inviter: author,
    guest: true,
    volume: DiscussionReader.volumes[:normal]
  )
end

def add_admin!(user, author)
  stances.create!(
    participant_id: user.id,
    inviter: author,
    volume: DiscussionReader.volumes[:normal],
    admin: true
  )
end
```

### Options Management

```ruby
def poll_option_names
  poll_options.map(&:name)
end

def poll_option_names=(names)
  names = Array(names)
  existing = Array(poll_options.pluck(:name))
  names = names.sort if poll_type == 'meeting'

  names.each_with_index do |name, priority|
    option = poll_options.find_or_initialize_by(name: name)
    option.priority = priority
    # Apply common_poll_options config if matching
    if params = AppConfig.poll_types.dig(poll_type, 'common_poll_options')&.find { |o| o['key'] == name }
      option.name = I18n.t(params['name_i18n'])
      option.icon = params['icon']
      option.meaning = I18n.t(params['meaning_i18n'])
      option.prompt = I18n.t(params['prompt_i18n'])
    end
  end

  removed = (existing - names)
  poll_options.each { |option| option.mark_for_destruction if removed.include?(option.name) }
  names
end

alias options= poll_option_names=
alias options poll_option_names
```

### Scoring Methods

```ruby
def total_score
  stance_counts.sum
end

def update_counts!
  poll_options.reload.each(&:update_counts!)
  update_columns(
    stance_counts: poll_options.map(&:total_score),
    voters_count: stances.latest.count,
    undecided_voters_count: stances.latest.undecided.count,
    none_of_the_above_count: stances.latest.decided.where(none_of_the_above: true).count,
    versions_count: versions.count
  )
end

def reset_latest_stances!
  # Recomputes latest flag for all stances
  transaction do
    stances.update_all(latest: false)
    Stance.where("id IN (
      SELECT DISTINCT ON (participant_id) id
      FROM stances
      WHERE poll_id = #{id}
      ORDER BY participant_id, created_at DESC
    )").update_all(latest: true)
  end
end
```

### Quorum Methods

```ruby
def quorum_count
  (quorum_pct.to_f / 100 * voters_count).ceil
end

def quorum_reached?
  quorum_pct && quorum_count <= voters_count
end

def quorum_votes_required
  return 0 if quorum_pct.nil?
  (((quorum_pct.to_f - cast_stances_pct.to_f) / 100) * voters_count).ceil
end
```

### Poll Type Configuration

These methods return values from AppConfig.poll_types based on poll_type:

```ruby
# TEMPLATE_DEFAULT_FIELDS - stored in column or custom_fields, with AppConfig defaults
def poll_option_name_format  # e.g., "iso8601" for meeting polls
def max_score
def min_score
def dots_per_person
def chart_type
def default_duration_in_days

# TEMPLATE_VALUES - read-only from AppConfig
def has_option_icon
def order_results_by
def prevent_anonymous
def vote_method
def material_icon
def require_all_choices
def validate_minimum_stance_choices
def validate_maximum_stance_choices
def validate_min_score
def validate_max_score
def has_options
def validate_dots_per_person
```

### Choice Constraints

```ruby
def minimum_stance_choices
  if require_all_choices
    poll_options.length
  else
    self[:minimum_stance_choices] ||
    self[:custom_fields][:minimum_stance_choices] ||
    AppConfig.poll_types.dig(poll_type, 'defaults', 'minimum_stance_choices') ||
    0
  end
end

def maximum_stance_choices
  self[:maximum_stance_choices] ||
  self[:custom_fields][:maximum_stance_choices] ||
  AppConfig.poll_types.dig(poll_type, 'defaults', 'maximum_stance_choices') ||
  poll_options.length
end

def is_single_choice?
  minimum_stance_choices == 1 && maximum_stance_choices == 1
end
```

### Display Methods

```ruby
def results
  PollService.calculate_results(self, poll_options)
end

def result_columns
  # Returns array of column names for results display based on poll_type
  case poll_type
  when 'proposal' then %w[chart name votes votes_cast_percent voter_percent voters]
  when 'check' then %w[chart name voter_percent voter_count voters]
  # ... etc for each poll type
  end
end

def chart_column
  case poll_type
  when 'count' then (agree_target ? 'target_percent' : 'voter_percent')
  when 'check', 'proposal' then 'score_percent'
  else 'max_score_percent'
  end
end

def dates_as_options
  poll_option_name_format == 'iso8601'
end

def results_include_undecided
  poll_type != "meeting"
end
```

### Version Tracking

```ruby
def is_new_version?
  !poll_options.map(&:persisted?).all? ||
  (['title', 'details', 'closing_at'] & changes.keys).any?
end
```

### Discussion Syncing

```ruby
def discussion_id=(discussion_id)
  super.tap { self.group_id = discussion&.group_id }
end

def discussion=(discussion)
  super.tap { self.group_id = discussion&.group_id }
end
```

### Null Object Handling

```ruby
def group
  super || NullGroup.new
end
```

---

## Counter Cache Updates

```ruby
update_counter_cache :group, :polls_count
update_counter_cache :group, :closed_polls_count
update_counter_cache :discussion, :closed_polls_count
update_counter_cache :discussion, :anonymous_polls_count
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `HasCustomFields` | Custom field accessors |
| `CustomCounterCache::Model` | Counter cache definitions |
| `ReadableUnguessableUrls` | 8-char key generation |
| `HasEvents` | Event associations |
| `HasMentions` | @mention extraction |
| `MessageChannel` | Real-time pub/sub |
| `SelfReferencing` | `poll` and `poll_id` |
| `Reactable` | Emoji reactions |
| `HasCreatedEvent` | Created event tracking |
| `HasRichText` | Rich text with sanitization |
| `HasTags` | Tag management |
| `Discard::Model` | Soft delete |
| `Searchable` | Full-text search |
| `Translatable` | Translation support |
| `NoSpam` | Spam validation |

---

## Paper Trail Tracking

Tracked fields:
- `author_id`
- `title`
- `details`
- `details_format`
- `closing_at`
- `closed_at`
- `group_id`
- `discussion_id`
- `anonymous`
- `discarded_at`
- `discarded_by`
- `voter_can_add_options`
- `specified_voters_only`
- `stance_reason_required`
- `tags`
- `notify_on_closing_soon`
- `poll_option_names`
- `hide_results`
- `attachments`

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `key` | UNIQUE | |
| `author_id` | INDEX | |
| `discussion_id` | INDEX | |
| `group_id` | INDEX | |
| `(closed_at, closing_at)` | INDEX | |
| `(closed_at, discussion_id)` | INDEX | |
| `tags` | GIN | Array search |

---

## Poll Types

| Type | Description |
|------|-------------|
| `proposal` | Agree/disagree/abstain/block voting |
| `poll` | Simple multiple choice |
| `count` | Count/check-in |
| `score` | Score each option |
| `dot_vote` | Allocate dots to options |
| `ranked_choice` | Ranked choice voting |
| `meeting` | Time/date selection |

---

## PollOption Model

Options available for selection in polls.

### Attributes

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `id` | serial | auto | Primary key |
| `poll_id` | integer | - | FK to polls |
| `name` | string | - | Option name/value |
| `priority` | integer | 0 | Display order |
| `icon` | string | - | Emoji icon |
| `meaning` | string | - | Semantic meaning |
| `prompt` | string | - | Selection prompt |
| `voter_count` | integer | 0 | Voters selecting this |
| `total_score` | integer | 0 | Sum of all scores |
| `score_counts` | jsonb | {} | Score distribution |
| `voter_scores` | jsonb | {} | Per-voter scores |

### Test Conditions (for proposal outcomes)

| Column | Type | Description |
|--------|------|-------------|
| `test_operator` | string | Comparison: 'gte', 'lte' |
| `test_percent` | integer | Target percentage (0-100) |
| `test_against` | string | 'score_percent' or 'voter_percent' |

### Key Methods

```ruby
def update_counts!
  update_columns(
    voter_scores: poll.anonymous ? {} : stance_choices.latest...to_h,
    total_score: stance_choices.latest.sum(:score),
    voter_count: stances.latest.count
  )
end

def average_score
  return 0 if voter_count == 0
  (total_score.to_f / voter_count.to_f)
end

def voter_ids
  # Returns IDs of users who voted for this option
  # Meeting polls: all voters, others: non-zero scores only
end
```

---

## Outcome Model

Published results/decisions from polls.

### Attributes

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `id` | serial | auto | Primary key |
| `poll_id` | integer | - | FK to polls |
| `poll_option_id` | integer | - | Winning option |
| `author_id` | integer | - | Outcome author |
| `statement` | text | - | Outcome statement |
| `statement_format` | string(10) | "md" | Format |
| `latest` | boolean | true | Most recent outcome |
| `review_on` | date | - | Review reminder date |
| `versions_count` | integer | 0 | Paper Trail versions |
| `custom_fields` | jsonb | {} | event_summary, event_description, event_location |

### Key Methods

```ruby
def calendar_invite
  return nil unless poll_option && dates_as_options
  CalendarInvite.new(self).to_ical
end

def attendee_emails
  stances.joins(:participant).joins(:stance_choices)
    .where("stance_choices.poll_option_id": poll_option_id)
    .pluck(:"users.email").flatten.compact.uniq
end
```

---

## Uncertainties

1. **stance_counts vs stance_data** - Both aggregate voting data, exact difference unclear
2. **matrix_counts** - Matrix poll functionality not fully documented
3. **AppConfig.poll_types** - Configuration source for poll type definitions
4. **vote_method** - Used for display but exact mapping unclear

**Confidence Level:** HIGH for core voting functionality, MEDIUM for specialized poll types.
