# Export Domain: Models

**Generated:** 2026-02-01
**Domain:** Export functionality for groups, discussions, and polls

---

## Overview

The export domain does not have dedicated models. Instead, it relies on concerns that extend existing models (Group, Discussion) with export-specific associations and methods. These concerns provide the relationship chains needed to collect all related data for export.

---

## 1. GroupExportRelations Concern

**Location:** `/app/models/concerns/group_export_relations.rb`

This concern is included in the Group model and provides associations and methods to gather all exportable data from a group and its subgroups.

### Key Associations

#### Poll-Related (with privacy filter)
- `exportable_polls`: Polls that are either non-anonymous OR already closed (protects anonymous poll data until closed)
- `exportable_poll_options`: Poll options for exportable polls
- `exportable_outcomes`: Outcomes for exportable polls
- `exportable_stances`: Votes/stances for exportable polls
- `exportable_stance_choices`: Individual choice selections within stances
- `poll_stance_receipts`: Voting receipts for exportable polls

#### Attachment Collections
- `comment_files`, `comment_image_files`: Files attached to comments
- `discussion_files`, `discussion_image_files`: Files attached to discussions
- `poll_files`, `poll_image_files`: Files attached to exportable polls
- `outcome_files`, `outcome_image_files`: Files attached to outcomes
- `subgroup_files`, `subgroup_image_files`: Files from subgroups
- `subgroup_cover_photos`, `subgroup_logos`: Subgroup branding images

#### Document Collections
- `discussion_documents`, `exportable_poll_documents`, `comment_documents`
- `public_discussion_documents`, `public_comment_documents`

#### Reaction Collections (with user join for validity)
- `discussion_reactions`, `exportable_poll_reactions`, `exportable_stance_reactions`
- `comment_reactions`, `exportable_outcome_reactions`

#### User-Related
- `discussion_authors`, `comment_authors`, `exportable_poll_authors`
- `exportable_outcome_authors`, `exportable_stance_authors`, `reader_users`

#### Event Collections
- `membership_events`, `discussion_events`, `comment_events`
- `exportable_poll_events`, `exportable_outcome_events`, `exportable_stance_events`

### Aggregation Methods

These methods use `Queries::UnionQuery` to combine multiple association sources:

| Method | Purpose |
|--------|---------|
| `all_users` | All users related to the group (members, authors, voters, reactors, readers) |
| `all_tags` | Tags from the group and all subgroups |
| `all_groups` | The group and all its subgroups |
| `all_events` | All events from memberships, discussions, comments, polls, outcomes, stances |
| `all_notifications` | Notifications for all events in the group |
| `all_documents` | Documents from all sources |
| `all_reactions` | Reactions from discussions, polls, stances, comments, outcomes |
| `reaction_users` | Users who have reacted to anything in the group |

---

## 2. DiscussionExportRelations Concern

**Location:** `/app/models/concerns/discussion_export_relations.rb`

This concern is included in the Discussion model and provides associations for exporting discussion-specific data.

### Key Associations

Similar structure to GroupExportRelations but scoped to a single discussion:
- `exportable_polls` (with same privacy filter)
- `exportable_poll_options`, `exportable_outcomes`, `exportable_stances`, `exportable_stance_choices`
- File attachments for comments, polls, outcomes
- Reaction collections

### Aggregation Methods

| Method | Purpose |
|--------|---------|
| `all_reactions` | Union of discussion reactions, poll reactions, stance reactions, comment reactions, outcome reactions |

---

## 3. Document Model (Export Storage)

**Location:** `/app/models/document.rb`

The Document model is used to store generated export files before delivery. Key characteristics:

- Has file attachment via Active Storage
- Belongs to an author (User)
- Has a title (used for the filename)
- Export files are scheduled for automatic deletion after 1 week

---

## 4. Privacy Considerations in Model Design

### Anonymous Poll Protection

The `exportable_polls` scope uses a critical privacy filter:

```
Pseudo-code: where(anonymous = false OR closed_at is not null)
```

This ensures:
- Non-anonymous polls are always exportable
- Anonymous polls are ONLY exportable after they close
- While anonymous polls are active, individual votes cannot be linked to voters in exports

### User Data Exclusions

When exporting user records, sensitive fields are excluded:
- `encrypted_password`
- `reset_password_token`
- `email_api_key`
- `secret_token`
- `unsubscribe_token`

### Group Token Exclusion

Group records exclude the `token` field to prevent sharing invitation tokens.

---

## 5. Relationship Diagram

```
Group
  |
  +-- all_users (union query)
  |     +-- members
  |     +-- discussion_authors
  |     +-- comment_authors
  |     +-- poll_authors
  |     +-- outcome_authors
  |     +-- stance_authors
  |     +-- reaction_users
  |     +-- reader_users
  |
  +-- all_events (union query)
  |     +-- membership_events
  |     +-- discussion_events
  |     +-- comment_events
  |     +-- poll_events
  |     +-- outcome_events
  |     +-- stance_events
  |
  +-- all_notifications
  |
  +-- all_reactions (union query)
  |
  +-- all_tags
  |
  +-- subgroups (recursive)
  |
  +-- exportable_polls (privacy filtered)
        +-- poll_options
        +-- stances
        +-- stance_choices
        +-- outcomes
        +-- stance_receipts
```

---

## Confidence Rating: 5/5

The model layer for exports is well-structured with clear concerns. The privacy protections (anonymous poll filtering, sensitive field exclusions) are properly implemented at the model/concern level.
