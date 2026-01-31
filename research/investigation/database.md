# Database Schema

> PostgreSQL schema patterns and Go migration considerations.

## PostgreSQL Extensions

| Extension | Purpose | Go Implication |
|-----------|---------|----------------|
| `citext` | Case-insensitive text (email, handle, tags) | Use `LOWER()` or custom collation |
| `hstore` | Key-value (translations, email headers) | `map[string]string` |
| `pgcrypto` | UUID generation | `github.com/google/uuid` |
| `pg_stat_statements` | Query monitoring | Optional |

## Table Inventory (57 tables)

| Category | Tables |
|----------|--------|
| **Core Domain (13)** | users, groups, memberships, discussions, comments, polls, poll_options, stances, stance_choices, outcomes, events, notifications, tags |
| **Access Control (6)** | discussion_readers, membership_requests, login_tokens, omniauth_identities, group_identities, member_email_aliases |
| **Templates (2)** | discussion_templates, poll_templates |
| **Content (5)** | documents, attachments, taggings, reactions, translations |
| **Tasks (2)** | tasks, tasks_users |
| **Integrations (4)** | webhooks, chatbots, received_emails, forward_email_rules |
| **Billing (2)** | subscriptions, group_surveys |
| **Audit (3)** | versions, active_admin_comments, user_deactivation_responses |
| **Search (1)** | pg_search_documents |
| **Analytics (5)** | blazer_* tables |
| **Rails (5)** | active_storage_*, action_mailbox_inbound_emails, ar_internal_metadata |
| **OAuth (3)** | oauth_applications, oauth_access_tokens, oauth_access_grants |
| **Other (6)** | blocked_domains, demos, cohorts, default_group_covers, partition_sequences, schema_migrations |

## Index Strategy

### Unique Indexes

| Table | Columns | Notes |
|-------|---------|-------|
| users | email, username, key | citext for email |
| groups | handle, key, token | citext for handle |
| discussions, polls | key | URL keys |
| memberships | (group_id, user_id) | One per user per group |
| discussion_readers | (user_id, discussion_id) | |
| stances | (poll_id, participant_id, latest) | **Partial: WHERE latest = true** |
| events | (discussion_id, sequence_id) | Timeline ordering |
| tags | (group_id, name) | |

### Partial Indexes

```sql
-- Soft delete optimization
CREATE INDEX ON discussions (discarded_at) WHERE discarded_at IS NULL;
CREATE INDEX ON groups (archived_at) WHERE archived_at IS NULL;

-- Latest stance per user per poll
CREATE UNIQUE INDEX ON stances (poll_id, participant_id, latest) WHERE latest = true;

-- Guest filtering
CREATE INDEX ON discussion_readers (guest) WHERE guest = true;
CREATE INDEX ON stances (guest) WHERE guest = true;
```

### GIN Indexes

```sql
-- Tag arrays
CREATE INDEX ON discussions USING gin (tags);
CREATE INDEX ON polls USING gin (tags);

-- Full-text search
CREATE INDEX ON pg_search_documents USING gin (ts_content);
```

## JSONB Field Structures

### Attachments
**Tables:** discussions, comments, polls, outcomes, stances, users, groups

**Structure:**
```go
type Attachment struct {
    ID          int64  `json:"id"`           // blob_id
    SignedID    string `json:"signed_id"`
    Filename    string `json:"filename"`
    ContentType string `json:"content_type"`
    ByteSize    int64  `json:"byte_size"`
    PreviewURL  string `json:"preview_url"`  // Often missing from docs
    DownloadURL string `json:"download_url"` // Often missing from docs
    Icon        string `json:"icon"`         // Often missing from docs
}
```

**Note:** Default is `[]` (empty array), not `{}` (empty object).

### Link Previews
**Tables:** discussions, comments, polls, outcomes, stances, users, groups

```go
type LinkPreview struct {
    Title       string `json:"title"`       // max 240 chars
    Description string `json:"description"` // max 240 chars
    Image       string `json:"image"`       // Note: 'image' not 'image_url'
    URL         string `json:"url"`
    Hostname    string `json:"hostname"`
    Fit         string `json:"fit"`         // 'contain'
    Align       string `json:"align"`       // 'center'
}
```

### Poll Voting Data

```go
// polls.stance_counts - vote counts per option (ordered by priority)
StanceCounts []int `json:"stance_counts"` // e.g., [1, 0, 2]

// stances.option_scores - user's scores per option
OptionScores map[string]int `json:"option_scores"` // e.g., {"123": 5, "124": 3}

// poll_options.voter_scores - all voters' scores for this option
VoterScores map[string]int `json:"voter_scores"` // e.g., {"1": 5, "2": 3}
// Note: Cleared for anonymous polls
```

### Custom Fields

```go
// polls.custom_fields
type PollCustomFields struct {
    MeetingDuration *int    `json:"meeting_duration,omitempty"`
    TimeZone        *string `json:"time_zone,omitempty"`
    CanRespondMaybe *bool   `json:"can_respond_maybe,omitempty"`
}

// events.custom_fields
type EventCustomFields struct {
    PinnedTitle         *string `json:"pinned_title,omitempty"`
    RecipientUserIDs    []int64 `json:"recipient_user_ids,omitempty"`
    RecipientChatbotIDs []int64 `json:"recipient_chatbot_ids,omitempty"`
    RecipientMessage    *string `json:"recipient_message,omitempty"`
    RecipientAudience   *string `json:"recipient_audience,omitempty"`
    StanceIDs           []int64 `json:"stance_ids,omitempty"`
}
```

## Full-Text Search

**Table:** `pg_search_documents`

| Column | Purpose |
|--------|---------|
| content | Searchable text |
| ts_content | Pre-computed tsvector |
| author_id, group_id, discussion_id, poll_id | Filter columns |
| searchable_type, searchable_id | Polymorphic source |
| authored_at | Sort key |

**Search Query Pattern:**
```sql
SELECT *, ts_rank(ts_content, query) AS rank
FROM pg_search_documents, plainto_tsquery('simple', $1) query
WHERE ts_content @@ query AND group_id = $2
ORDER BY rank DESC, authored_at DESC
LIMIT 50;
```

## Common Query Patterns

### Discussion Visibility

```sql
SELECT * FROM discussions
LEFT JOIN discussion_readers dr ON dr.discussion_id = discussions.id AND dr.user_id = $1
WHERE discarded_at IS NULL
  AND (private = FALSE
       OR group_id IN ($2)  -- user's groups
       OR (dr.id IS NOT NULL AND dr.revoked_at IS NULL));
```

### Latest Stance Per User

```sql
SELECT * FROM stances
WHERE poll_id = $1 AND latest = TRUE AND revoked_at IS NULL;
```

### Poll Status

```sql
-- Active polls
WHERE discarded_at IS NULL AND closed_at IS NULL

-- Lapsed (should be closed)
WHERE closed_at IS NULL AND closing_at < NOW()
```

## Go Type Mappings

| PostgreSQL | Go Type | Notes |
|------------|---------|-------|
| serial/integer | int64 | Prefer int64 for IDs |
| character varying | string | |
| text | string | |
| boolean | bool | |
| timestamp | time.Time | `*time.Time` for nullable |
| jsonb | struct or map[string]any | |
| character varying[] | []string | pq.Array |
| citext | string | Handle in queries |
| hstore | map[string]string | |

## Unique Constraint Handling

```go
if pgErr, ok := err.(*pgconn.PgError); ok {
    if pgErr.Code == "23505" { // unique_violation
        return ErrDuplicate
    }
}
```

---
