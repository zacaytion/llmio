# Polls Domain: Frontend

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

The polls frontend is a Vue 3 application that uses LokiJS for client-side data storage. Key components handle poll display, voting forms, and results visualization.

---

## Client-Side Models

### PollModel

**File:** `/vue/src/shared/models/poll_model.js`

Client-side representation of polls with computed properties and methods.

#### Key Properties

```javascript
defaultValues: {
  discussionId: null,
  title: '',
  closingAt: null,
  details: '',
  specifiedVotersOnly: false,
  pollOptionNames: [],
  pollType: 'proposal',
  hideResults: 'off',
  stanceCounts: [],
  results: []
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `config()` | Returns poll type configuration from AppConfig |
| `pollOptions()` | Returns ordered poll options from Records |
| `pollOptionsForVoting()` | Options, shuffled if shuffle_options is true |
| `myStance()` | Current user's latest stance |
| `iHaveVoted()` | Whether current user has cast a vote |
| `showResults()` | Whether to display results based on hide_results setting |
| `iCanVote()` | Whether current user can vote |
| `isVotable()` | Poll is not discarded and not closed |
| `isClosed()` | Poll has closed_at |
| `adminsInclude(user)` | Whether user can admin this poll |
| `membersInclude(user)` | Whether user can view this poll |
| `pieSlices()` | Computed pie chart data |
| `decidedVoters()` | Users who have voted |
| `outcome()` | Latest outcome record |
| `latestStances()` | All latest stances, sorted |
| `latestCastStances()` | Only stances with votes |
| `clonePoll()` | Creates a copy for templating |

#### Relationships

```javascript
relationships() {
  this.belongsTo('author', {from: 'users'});
  this.belongsTo('discussion');
  this.belongsTo('group');
  this.hasMany('stances');
  this.hasMany('versions');
}
```

#### Results Display Logic

```javascript
showResults() {
  switch (this.hideResults) {
    case "until_closed": return this.closedAt;
    case "until_vote": return this.closedAt || this.iHaveVoted();
    default: return true;
  }
}
```

### StanceModel

**File:** `/vue/src/shared/models/stance_model.js`

Client-side representation of votes.

#### Key Properties

```javascript
defaultValues: {
  reason: '',
  reasonFormat: 'html',
  optionScores: {},
  castAt: null,
  guest: false
}
```

#### Key Methods

| Method | Description |
|--------|-------------|
| `pollOptions()` | Options this stance voted for |
| `pollOptionIds()` | IDs of selected options |
| `sortedChoices()` | Choices sorted by score/rank |
| `scoreFor(option)` | Score given to specific option |
| `totalScore()` | Sum of all scores |
| `votedFor(option)` | Whether this option was selected |
| `participantName()` | Voter name (or "Anonymous") |
| `singleChoice()` | Whether poll is single-choice |
| `hasOptionIcon()` | Whether to show icons |

#### Choice Representation

```javascript
sortedChoices() {
  // Returns array of {score, rank, show, pollOption}
  // Sorted by priority (meetings), rank (ranked_choice), or score
}
```

---

## Frontend Services

### PollService

**File:** `/vue/src/shared/services/poll_service.js`

Provides action definitions for polls.

#### Actions Object

Returns a set of action definitions for poll context menus and action docks:

| Action | Dock | Description |
|--------|------|-------------|
| `view_all_votes` | 2 | Navigate to full poll page |
| `translate_poll` | 3 | Translate poll content |
| `edit_stance` | 2 | Change user's vote |
| `uncast_stance` | menu | Remove user's vote |
| `edit_poll` | menu | Edit poll settings |
| `make_a_copy` | menu | Clone poll as template |
| `add_poll_to_thread` | menu | Attach to discussion |
| `announce_poll` | 2 | Invite voters |
| `remind_poll` | 2 | Send reminders |
| `close_poll` | 2 | Close early |
| `reopen_poll` | 3 | Reopen closed poll |
| `show_history` | menu | View edit history |
| `notification_history` | menu | View announcement history |
| `move_poll` | menu | Move to different group |
| `export_poll` | menu | Export as CSV |
| `print_poll` | menu | Export as HTML |
| `verify_participants` | menu | View receipts |
| `discard_poll` | menu | Delete poll |

Each action has:
- `canPerform()`: Visibility check
- `perform()`: Action handler (modal, navigation, etc.)
- `icon`: Material icon name
- `name`: i18n key

### StanceService

**File:** `/vue/src/shared/services/stance_service.js`

Handles voting operations.

Key methods:
- `canUpdateStance(stance)`: Check if user can modify stance
- `updateStance(stance)`: Open voting modal
- `uncastStance(stance)`: Remove vote with confirmation

---

## Component Structure

### Poll Components

Located in `/vue/src/components/poll/`:

#### Common Components
- `poll/common/poll_option_form.vue` - Option editing form
- `poll/common/stance.vue` - Display a single stance
- `poll/common/stance_choice.vue` - Display a choice within stance
- `poll/common/stance_choices.vue` - List of choices
- `poll/common/stance_icon.vue` - Icon for vote type
- `poll/common/stance_reason.vue` - Reason text display

#### Poll Type Specific
- `poll/meeting/stance_icon.vue` - Meeting availability icon

### Strand Components

Located in `/vue/src/components/strand/item/`:

- `poll_created.vue` - Poll creation event display
- `poll_edited.vue` - Poll edit event display
- `stance_created.vue` - New vote event
- `stance_updated.vue` - Vote change event

### Dashboard Components

- `dashboard/polls_panel.vue` - Polls list on dashboard
- `dashboard/polls_to_vote_on_page.vue` - Pending votes page
- `group/polls_panel.vue` - Polls within a group

### Thread Components

- `thread/current_poll_banner.vue` - Active poll indicator in discussion

---

## Records Interface

### Poll Records

**File:** `/vue/src/shared/interfaces/poll_records_interface.js`

Defines how poll records are fetched and stored:

```javascript
// Fetching
Records.polls.fetch({key: 'abc123'})
Records.polls.fetchByGroup(groupKey, params)
Records.polls.fetchByDiscussion(discussionKey)

// Remote actions
Records.polls.remote.postMember(key, 'close')
Records.polls.remote.postMember(key, 'reopen', {poll: {...}})
Records.polls.remote.patchMember(key, 'add_to_thread', {...})
```

### Stance Records

**File:** `/vue/src/shared/interfaces/stance_records_interface.js`

Defines stance operations:

```javascript
// Creating/updating votes
Records.stances.remote.create({stance: {...}})
Records.stances.remote.update(id, {stance: {...}})

// Fetching stances for a poll
Records.stances.fetchForPoll(pollId, params)
```

---

## Data Flow

### Voting Flow

1. User opens poll or voting modal
2. Component loads poll and poll_options from Records
3. User selects options, sets scores via UI
4. Stance model built with optionScores map
5. POST to /api/v1/stances (or PATCH for update)
6. Server creates/updates stance, returns event
7. Records imports event and updated stance
8. UI reactively updates to show new vote

### Results Display Flow

1. Poll model checks showResults()
2. If true, results array is available
3. Results mapped to chart data (pieSlices, bars)
4. Components render with voter counts, percentages
5. Voter avatars fetched from voter_ids in results

### Real-time Updates

Poll and stance changes are pushed via MessageChannelService:

1. Server publishes to Redis channel
2. External channels service (WebSocket/SSE) pushes to client
3. Client receives model updates
4. LokiJS store updates records
5. Vue reactivity triggers component re-render

---

## Poll Type Configuration

Poll type configurations are loaded from the server into AppConfig.pollTypes.

Frontend accesses via:
```javascript
poll.config() // Returns type config
poll.config().has_option_icon // Boolean
poll.config().has_options // Boolean
poll.defaulted('minScore') // Gets value or falls back to config default
```

Key configuration properties used by frontend:
- `has_option_icon`: Show icons on options
- `has_options`: Whether poll has selectable options
- `material_icon`: Icon to display for poll type
- `vote_method`: How voting works (show_thumbs, choose, allocate, etc.)
- `can_shuffle_options`: Whether shuffle is supported
- `allow_none_of_the_above`: Whether "none" option is available

---

## Internationalization

Poll-related i18n keys are organized under:

- `poll_types.*`: Poll type names
- `poll_common.*`: Shared poll UI text
- `poll_proposal_options.*`: Proposal option labels
- `poll_templates.*`: Template descriptions
- `action_dock.*`: Action button labels
- `poll_common_form.*`: Form labels

Option names can use `i18n` format to reference translation keys.

---

## Chart Rendering

Poll results support multiple chart types:

| Type | Usage |
|------|-------|
| `pie` | Proposal, check polls |
| `bar` | Poll, dot_vote, score, ranked_choice |
| `grid` | Meeting polls (time availability) |
| `none` | Question polls |

The `pieSlices()` method on PollModel computes chart data:
- For count polls with target: Shows progress toward goal
- For other polls: Shows proportion of votes per option
