# Core Models - Confirmed Architecture

## Summary

Both Discovery and Research documentation **fully agree** on the core domain model architecture. This document consolidates confirmed findings for the Go rewrite.

## Key Details

### Poll Types (9 Confirmed)

Both sources agree on exactly 9 poll types:

| Type | Purpose | Voting Mechanic |
|------|---------|-----------------|
| `proposal` | Consensus decisions | Yes/No/Abstain/Block thumbs |
| `poll` | Multiple choice | Choose one or more options |
| `count` | Simple headcount | Opt-in/opt-out counting |
| `score` | Numeric rating | Rate each option on scale |
| `ranked_choice` | Preference ordering | Rank options by preference |
| `meeting` | Time scheduling | Availability polling (0/1/2) |
| `dot_vote` | Budget allocation | Allocate limited dots |
| `check` | Sense check | Looks good/not sure/concerned |
| `question` | Open-ended | Reason-only, no voting options |

### Event Kinds (42 Confirmed)

Both sources agree on exactly 42 event kinds:

**Discussion Events (9):**
`new_discussion`, `discussion_edited`, `discussion_announced`, `discussion_closed`, `discussion_reopened`, `discussion_moved`, `discussion_forked`, `discussion_title_edited`, `discussion_description_edited`

**Comment Events (3):**
`new_comment`, `comment_edited`, `comment_replied_to`

**Poll Events (9):**
`poll_created`, `poll_edited`, `poll_announced`, `poll_reminder`, `poll_closing_soon`, `poll_closed_by_user`, `poll_expired`, `poll_reopened`, `poll_option_added`

**Stance Events (2):**
`stance_created`, `stance_updated`

**Outcome Events (4):**
`outcome_created`, `outcome_announced`, `outcome_updated`, `outcome_review_due`

**Membership Events (9):**
`membership_created`, `invitation_accepted`, `user_joined_group`, `user_added_to_group`, `membership_requested`, `membership_request_approved`, `membership_resent`, `new_coordinator`, `new_delegate`

**Mention Events (2):**
`user_mentioned`, `group_mentioned`

**Other Events (4):**
`reaction_created`, `announcement_resend`, `user_reactivated`, `unknown_sender`

### Webhook-Eligible Events (14)

Both sources agree 14 events trigger webhooks (see `follow_up/webhook_eligible_events.md` for enumeration verification).

### Soft Delete Patterns

Both sources confirm soft delete via timestamps:

| Pattern | Column | Used On |
|---------|--------|---------|
| Discard | `discarded_at` | Discussions, Comments, Polls, Templates |
| Revoke | `revoked_at` | Memberships, DiscussionReaders, Stances |
| Deactivate | `deactivated_at` | Users |
| Archive | `archived_at` | Groups |

### Volume Levels

Both sources confirm 4 volume levels for notification preferences:

| Level | Value | Behavior |
|-------|-------|----------|
| Mute | 0 | No notifications |
| Quiet | 1 | In-app only, no email |
| Normal | 2 | Standard notifications |
| Loud | 3 | All notifications, email digest |

### Counter Caches

Both sources confirm Groups have 17 counter cache columns:

- `memberships_count`
- `pending_memberships_count`
- `admin_memberships_count`
- `discussions_count`
- `open_discussions_count`
- `closed_discussions_count`
- `public_discussions_count`
- `polls_count`
- `closed_polls_count`
- `subgroups_count`
- `members_count` (may be alias)
- `invitations_count`
- And 5+ more for various aggregate counts

### Stance/Voting Mechanics

Both sources agree on vote storage:

**Stance Model:**
- `option_scores: JSONB` - Map of poll_option_id to score
- `latest: boolean` - Is this the current vote (partial unique index)
- `cast_at: timestamp` - When vote was cast (null = invited but not voted)
- `participant_id` - Voter reference (scrubbed for anonymous polls on close)

**StanceChoice Model:**
- Join table: `stance_id`, `poll_option_id`, `score`
- Score semantics vary by poll type

**Poll aggregates:**
- `stance_counts: JSONB[]` - Array indexed by option priority
- `voter_scores: JSONB{}` - Per-option map (cleared for anonymous)

## Source Alignment

| Aspect | Discovery | Research | Status |
|--------|-----------|----------|--------|
| Poll types count | 9 | 9 | ✅ Confirmed |
| Event kinds count | 42 | 42 | ✅ Confirmed |
| Webhook events count | 14 | 14 | ✅ Confirmed |
| Soft delete patterns | 4 patterns | 4 patterns | ✅ Confirmed |
| Volume levels | 4 (0-3) | 4 (0-3) | ✅ Confirmed |
| Counter caches on Group | 17 | 17 | ✅ Confirmed |
| Stance.latest pattern | Documented | Documented | ✅ Confirmed |
| option_scores JSONB | Documented | Documented | ✅ Confirmed |

## Implementation Notes

### Go Type Definitions

```go
// Poll types as enum
type PollType string

const (
    PollTypeProposal     PollType = "proposal"
    PollTypePoll         PollType = "poll"
    PollTypeCount        PollType = "count"
    PollTypeScore        PollType = "score"
    PollTypeRankedChoice PollType = "ranked_choice"
    PollTypeMeeting      PollType = "meeting"
    PollTypeDotVote      PollType = "dot_vote"
    PollTypeCheck        PollType = "check"
    PollTypeQuestion     PollType = "question"
)

// Event kinds - use string for STI pattern
type EventKind string

// Volume levels
type VolumeLevel int

const (
    VolumeMute   VolumeLevel = 0
    VolumeQuiet  VolumeLevel = 1
    VolumeNormal VolumeLevel = 2
    VolumeLoud   VolumeLevel = 3
)

// Stance with JSONB fields
type Stance struct {
    ID            int64                  `json:"id"`
    PollID        int64                  `json:"poll_id"`
    ParticipantID *int64                 `json:"participant_id"` // nullable for anonymous
    OptionScores  map[string]int         `json:"option_scores"`
    Latest        bool                   `json:"latest"`
    CastAt        *time.Time             `json:"cast_at"`
    CreatedAt     time.Time              `json:"created_at"`
}
```

### Database Considerations

- Use partial unique index: `CREATE UNIQUE INDEX ON stances (poll_id, participant_id) WHERE latest = true`
- Counter caches require triggers or application-level maintenance
- Event STI pattern maps to single `events` table with `kind` discriminator
