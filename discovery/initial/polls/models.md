# Polls Domain: Models

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

The polls domain consists of five core models that work together to implement collaborative decision-making:

- **Poll**: The decision or question being asked
- **PollOption**: The choices voters can select
- **Stance**: A voter's participation record and vote
- **StanceChoice**: A specific option selection within a vote
- **Outcome**: The result summary after a poll closes

---

## Poll Model

**File:** `/app/models/poll.rb`

### Purpose

The Poll model represents a decision or question posed to a group. It supports multiple poll types with different voting mechanics, privacy settings, and result display options.

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `poll_type` | string | Type of poll (proposal, poll, meeting, etc.) |
| `title` | string | Poll title |
| `details` | text | Rich text description |
| `closing_at` | datetime | When voting ends |
| `closed_at` | datetime | When poll was actually closed |
| `anonymous` | boolean | Whether voter identities are hidden |
| `hide_results` | enum | off, until_vote, until_closed |
| `specified_voters_only` | boolean | Restrict voting to invited users |
| `voter_can_add_options` | boolean | Allow voters to add options |
| `notify_on_closing_soon` | enum | Who to notify: nobody, author, undecided_voters, voters |
| `stance_reason_required` | enum | disabled, optional, required |
| `quorum_pct` | integer | Percentage of voters needed for quorum |

### Poll Types

Defined in `/config/poll_types.yml`:

1. **proposal**: Agree/Abstain/Disagree/Block thumbs voting
2. **poll**: Choose one or more options
3. **meeting**: Time availability polling (dates as options)
4. **dot_vote**: Allocate limited dots across options
5. **score**: Rate each option on a scale
6. **ranked_choice**: Rank options by preference
7. **count**: Simple counting (opt-in/opt-out)
8. **check**: Sense check (looks good/not sure/concerned)
9. **question**: Reason-only response, no voting options

### Voting Configuration

Each poll type has defaults defined in poll_types.yml:

- `min_score` / `max_score`: Score range for voting
- `minimum_stance_choices` / `maximum_stance_choices`: How many options must/can be selected
- `dots_per_person`: For dot voting
- `require_all_choices`: Must vote on every option (meetings)
- `chart_type`: bar, pie, grid, none

### Poll Lifecycle States

1. **Draft/WIP**: `closing_at` is null
2. **Active**: `closing_at` is set, `closed_at` is null
3. **Closing Soon**: Active and within 24 hours of closing
4. **Closed**: `closed_at` is set

Query scopes:
- `Poll.active` - Not closed, not discarded
- `Poll.closed` - Has closed_at
- `Poll.lapsed_but_not_closed` - Past closing_at but not yet processed
- `Poll.closing_soon_not_published` - Due to close within timeframe, no event published

### Anonymous Voting

When `anonymous = true`:
- Participant identities are hidden during voting
- The `participant` method returns `AnonymousUser` instead of real user
- Event user_id is set to nil for stance events
- On close, participant_id is permanently scrubbed from stances
- Cannot be changed back to non-anonymous once set

### Hide Results Settings

Enum values for `hide_results`:
- `off` (0): Results always visible
- `until_vote` (1): Results hidden until user votes
- `until_closed` (2): Results hidden until poll closes

The `show_results?(voted:)` method determines visibility:
- Returns true if results are off (always show)
- Returns true if until_vote and user has voted or poll closed
- Returns true if until_closed and poll is closed

### Specified Voters vs Open Voting

- `specified_voters_only = false`: All group members automatically get a stance
- `specified_voters_only = true`: Only explicitly invited users can vote

When specified_voters_only is false, `PollService.create_anyone_can_vote_stances` creates stances for all group members and discussion guests.

### Key Associations

```
Poll
  belongs_to :author (User)
  belongs_to :discussion (optional)
  belongs_to :group (optional)
  has_many :poll_options (ordered by priority)
  has_many :stances
  has_many :outcomes
  has_one :current_outcome (latest: true)
  has_many :voters (through latest stances)
  has_many :undecided_voters
  has_many :decided_voters
```

### Counter Caches

- `voters_count`: Total stances (latest)
- `undecided_voters_count`: Stances without cast_at
- `stance_counts`: Array of scores per option
- `none_of_the_above_count`: Voters who chose none

### Concerns Included

- HasRichText (details field)
- HasMentions (for @mentions)
- HasEvents (activity tracking)
- HasTags
- Discard::Model (soft delete)
- Searchable (full-text search)
- ReadableUnguessableUrls (key generation)
- Reactable

---

## PollOption Model

**File:** `/app/models/poll_option.rb`

### Purpose

Represents a choice voters can select. Options have names, icons, meanings, and track voting statistics.

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Option text/identifier |
| `priority` | integer | Display order |
| `icon` | string | Visual icon (agree, disagree, etc.) |
| `meaning` | text | What selecting this option means |
| `prompt` | text | Guidance for voters |
| `voter_scores` | jsonb | Map of user_id to score |
| `total_score` | integer | Sum of all votes for this option |
| `voter_count` | integer | Number of voters who chose this |

### Threshold Testing

Poll options support pass/fail thresholds:
- `test_operator`: 'gte' or 'lte'
- `test_percent`: Target percentage
- `test_against`: 'score_percent' or 'voter_percent'

Used to determine if an option has passed or failed based on results.

### Score Tracking

The `update_counts!` method calculates:
- `voter_scores`: Hash mapping participant_id to their score
- `total_score`: Sum of all stance_choice scores
- `voter_count`: Count of distinct voters

---

## Stance Model

**File:** `/app/models/stance.rb`

### Purpose

A Stance represents a user's participation in a poll. It may or may not contain an actual vote (cast_at determines if voted).

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `participant_id` | integer | The voting user |
| `poll_id` | integer | The poll |
| `cast_at` | datetime | When vote was cast (null = invited but not voted) |
| `reason` | text | Voter's explanation |
| `option_scores` | jsonb | Map of poll_option_id to score |
| `latest` | boolean | Is this the most recent stance |
| `guest` | boolean | Is user a poll guest |
| `admin` | boolean | Is user a poll admin |
| `inviter_id` | integer | Who invited this voter |
| `revoked_at` | datetime | When access was revoked |
| `none_of_the_above` | boolean | Voted for none of the options |

### Voting States

- **Undecided**: `cast_at` is null (invited but hasn't voted)
- **Decided**: `cast_at` is set (has voted)
- **Latest**: `latest = true` (current stance, not superseded)
- **Revoked**: `revoked_at` is set (access removed)

Scopes:
- `Stance.latest` - Current active stances
- `Stance.decided` - Has cast_at
- `Stance.undecided` - No cast_at
- `Stance.revoked` - Has revoked_at

### Vote Representation

The stance stores votes in two ways:
1. `option_scores` jsonb: Quick lookup map of option_id to score
2. `stance_choices` association: Detailed choice records

The `choice=` method accepts:
- String: Single option name
- Array: Multiple option names
- Hash: Option name to score mapping

### Validation Rules

Stances validate based on poll configuration:
- `valid_minimum_stance_choices`: Must select enough options
- `valid_maximum_stance_choices`: Cannot select too many
- `valid_min_score` / `valid_max_score`: Score range limits
- `valid_dots_per_person`: Total dots allocated
- `valid_reason_required`: Reason presence if required
- `valid_require_all_choices`: Must vote on all options (meetings)

### Anonymous Handling

When poll is anonymous:
- `participant` method returns `AnonymousUser`
- `real_participant` returns actual user
- On poll close, `participant_id` is set to null

### Volume Preferences

Stances include HasVolume for notification preferences:
- mute (0), quiet (1), normal (2), loud (3)
- Cascades from DiscussionReader to Membership to User default

---

## StanceChoice Model

**File:** `/app/models/stance_choice.rb`

### Purpose

Represents a single option selection within a stance, with an associated score.

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `stance_id` | integer | Parent stance |
| `poll_option_id` | integer | Selected option |
| `score` | integer | Vote weight (default 1) |

### Scoring Semantics

- For choose-one polls: score = 1
- For score polls: score = user's rating
- For dot vote: score = dots allocated
- For ranked choice: score represents rank (higher = better rank)
- For meetings: score represents availability (0/1/2)

The `rank` method calculates rank from score for ranked_choice polls.

### Scope

- `StanceChoice.latest`: Choices from latest, non-revoked stances

---

## Outcome Model

**File:** `/app/models/outcome.rb`

### Purpose

Represents the declared result or summary after a poll closes. Created by poll admins to communicate decisions.

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `poll_id` | integer | The poll |
| `author_id` | integer | Who created the outcome |
| `statement` | text | Rich text outcome statement |
| `poll_option_id` | integer | The winning/chosen option (optional) |
| `latest` | boolean | Is current outcome |
| `review_on` | date | Scheduled review date |

### Calendar Integration

For meeting polls, outcomes can generate calendar invites:
- `calendar_invite` method returns ICS format
- Uses poll_option for the selected time
- Includes event_summary, event_description, event_location from custom_fields

### Review Due Scheduling

Outcomes can have a `review_on` date. The `OutcomeService.publish_review_due` method finds outcomes due for review and publishes reminder events.

---

## Model Relationships Summary

```
Group
  has_many :polls

Discussion
  has_many :polls

Poll
  has_many :poll_options
  has_many :stances
  has_many :outcomes
  has_many :stance_choices (through stances)

PollOption
  belongs_to :poll
  has_many :stance_choices

Stance
  belongs_to :poll
  belongs_to :participant (User)
  has_many :stance_choices
  has_many :poll_options (through stance_choices)

StanceChoice
  belongs_to :stance
  belongs_to :poll_option

Outcome
  belongs_to :poll
  belongs_to :author (User)
  belongs_to :poll_option (optional)
```

---

## Paper Trail Versioning

Both Poll and Stance use Paper Trail for edit history:

**Poll tracks:** title, details, closing_at, closed_at, group_id, discussion_id, anonymous, discarded_at, voter_can_add_options, specified_voters_only, stance_reason_required, tags, notify_on_closing_soon, poll_option_names, hide_results

**Stance tracks:** reason, option_scores, revoked_at, revoker_id, inviter_id

**Outcome tracks:** statement, statement_format, author_id, review_on
