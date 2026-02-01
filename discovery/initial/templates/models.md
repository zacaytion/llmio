# Templates Domain: Models

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

The templates system in Loomio provides pre-configured starting points for discussions and polls. There are two template types:

1. **DiscussionTemplate** - Templates for starting discussions/threads
2. **PollTemplate** - Templates for creating polls/decisions

Both types exist in two forms:
- **System templates** - Defined in YAML configuration files, available to all groups
- **Custom templates** - Stored in the database, created by group admins

---

## DiscussionTemplate Model

**Location:** `/app/models/discussion_template.rb`

### Database Schema

| Column | Type | Description |
|--------|------|-------------|
| id | integer | Primary key (only for custom templates) |
| key | string | Identifier for system templates (blank for custom) |
| group_id | integer | Owning group |
| author_id | integer | Creator user |
| position | integer | Display order within group |
| process_name | string | Template title (required) |
| process_subtitle | string | Short description (required) |
| process_introduction | string | Rich text introduction shown before use |
| process_introduction_format | string | Format: html or md |
| title | string | Default title for new discussions |
| title_placeholder | string | Placeholder text for title field |
| description | text | Default description content |
| description_format | string | Format: html or md (default: html) |
| tags | string[] | Default tags to apply |
| max_depth | integer | Reply nesting level (1=linear, 2=nested once, 3=nested twice) |
| newest_first | boolean | Sort order for replies |
| recipient_audience | string | Default audience for notifications (null or "group") |
| poll_template_keys_or_ids | jsonb | Array of associated poll template identifiers |
| public | boolean | Whether template is shared in public gallery |
| discarded_at | datetime | Soft delete timestamp |
| discarded_by | integer | User who discarded |
| attachments | jsonb | File attachments |
| link_previews | jsonb | Link preview data |
| content_locale | string | Content language |

### Concerns Included

- **Discard::Model** - Soft delete support via discarded_at timestamp
- **HasRichText** - Rich text handling for description field
- **CustomCounterCache::Model** - Updates group.discussion_templates_count

### Key Relationships

- `belongs_to :author` - The user who created the template
- `belongs_to :group` - The group owning this template

### Important Methods

**poll_templates** - Returns associated PollTemplate records by filtering poll_template_keys_or_ids for integer IDs

**poll_template_ids** - Extracts only integer IDs from poll_template_keys_or_ids (custom templates)

**dump_i18n** - Exports template content for internationalization purposes

### Paper Trail Versioning

Tracks changes to: public, title, process_name, process_subtitle, process_introduction, process_introduction_format, description, description_format, group_id, tags, discarded_at, attachments

---

## PollTemplate Model

**Location:** `/app/models/poll_template.rb`

### Database Schema

| Column | Type | Description |
|--------|------|-------------|
| id | integer | Primary key (only for custom templates) |
| key | string | Identifier for system templates |
| group_id | integer | Owning group (required) |
| author_id | integer | Creator user (required) |
| position | integer | Display order (default: 0) |
| poll_type | string | Underlying poll type (required) |
| process_name | string | Template title (required) |
| process_subtitle | string | Short description (required) |
| process_introduction | string | Rich text introduction |
| process_introduction_format | string | Format: md or html |
| title | string | Default poll title |
| title_placeholder | string | Placeholder for title field |
| details | text | Default poll description |
| details_format | string | Format: md (default) |
| default_duration_in_days | integer | Default voting period (required, default: 7) |
| poll_options | jsonb | Array of option configurations |
| anonymous | boolean | Anonymous voting (default: false) |
| specified_voters_only | boolean | Restrict to specified voters |
| shuffle_options | boolean | Randomize option order |
| show_none_of_the_above | boolean | Add NOTA option |
| hide_results | enum | When to show results: off, until_vote, until_closed |
| notify_on_closing_soon | enum | Who to notify: nobody, author, undecided_voters, voters |
| stance_reason_required | enum | Reason requirement: disabled, optional, required |
| limit_reason_length | boolean | Cap reason length |
| reason_prompt | string | Custom prompt for reasons |
| chart_type | string | Visualization type |
| min_score, max_score | integer | Score range (for score polls) |
| minimum_stance_choices, maximum_stance_choices | integer | Choice limits |
| dots_per_person | integer | Dot vote allocation |
| meeting_duration | integer | Meeting length in minutes |
| can_respond_maybe | boolean | Allow maybe responses |
| poll_option_name_format | string | Option formatting: plain (default) |
| agree_target | integer | Target agreement percentage |
| quorum_pct | integer | Required participation percentage (0-100) |
| outcome_statement | string | Pre-filled outcome text |
| outcome_statement_format | string | Format for outcome |
| outcome_review_due_in_days | integer | Days until outcome review |
| tags | string[] | Default tags |
| discarded_at | datetime | Soft delete timestamp |
| public | boolean | Shared in gallery |

### Concerns Included

- **Discard::Model** - Soft delete support
- **HasRichText** - Rich text for details field
- **CustomCounterCache::Model** - Updates group.poll_templates_count

### Key Relationships

- `belongs_to :author` - Template creator
- `belongs_to :group` - Owning group

### Enums

The model uses database-backed enums for:

**notify_on_closing_soon:**
- 0: nobody
- 1: author
- 2: undecided_voters
- 3: voters

**hide_results:**
- 0: off
- 1: until_vote
- 2: until_closed

**stance_reason_required:**
- 0: disabled
- 1: optional
- 2: required

### Validations

- poll_type must be in AppConfig.poll_types.keys
- details length capped at max_message_length
- process_name and process_subtitle required
- default_duration_in_days required
- quorum_pct normalized to 0-100 range

---

## System Templates Configuration

### Poll Templates

**Location:** `/config/poll_templates.yml`

Defines 14 built-in poll templates including:
- practice_proposal, proposal, check, advice, consent, consensus
- question, count, poll, dot_vote, score, ranked_choice, meeting
- gradients_of_agreement

Each template specifies:
- poll_type - Which poll engine to use
- Internationalized text via `_i18n` suffixes
- default_duration_in_days
- poll_options - Array of voting options with name, meaning, prompt, icon, color

### Discussion Templates

**Location:** `/config/discussion_templates.yml`

Defines 13 built-in discussion templates including:
- blank, practice_thread, discuss_a_topic
- onboarding_to_loomio, advice_process, consent_process
- decision_by_consensus, approve_a_document, share_links_and_info
- prepare_for_a_meeting, funding_decision, elect_a_governance_position
- build_a_policy

Each template specifies:
- process_name, process_subtitle, process_introduction (via i18n)
- title/description defaults
- recipient_audience
- max_depth, newest_first
- position for ordering
- tags

---

## Group Settings for Templates

Groups store template configuration in the `info` JSONB column:

### Poll Template Settings

**poll_template_positions** - Hash mapping template key to position number
- Default ordering defined for standard templates

**hidden_poll_templates** - Array of template keys to hide
- Defaults from AppConfig.app_features[:hidden_poll_templates]
- Default: ["proposal", "question"]

**categorize_poll_templates** - Boolean for UI display mode
- true: Show templates in categories (proposal/poll/meeting)
- false: Show all templates in one list

### Discussion Template Settings

**hidden_discussion_templates** - Array of hidden template keys
- Managed similarly to poll templates

**discussion_template_positions** - Hash of position overrides
- Used for reordering system templates

---

## Template Instantiation

Templates are converted to actual discussions/polls through these mechanisms:

### Discussion Creation

When creating a discussion from a template:
1. Frontend calls `discussionTemplate.buildDiscussion()`
2. Picks template attributes and copies to new Discussion record
3. Sets `discussionTemplateId` and `discussionTemplateKey` on the discussion
4. User can modify values before saving

### Poll Creation

When creating a poll from a template:
1. Frontend calls `pollTemplate.buildPoll()`
2. Picks template attributes
3. Sets `pollTemplateId` and `pollTemplateKey` on the poll
4. Calculates `closingAt` from current time + defaultDurationInDays
5. Converts pollOptions to pollOptionsAttributes format

### Referential Integrity

Discussions and Polls store:
- `poll_template_id` / `discussion_template_id` - Reference to custom template
- `poll_template_key` / `discussion_template_key` - Key for system templates

The Poll model provides a `poll_template` method that looks up templates:
1. First checks poll_template_id for custom templates
2. Falls back to poll_template_key for system templates

---

## Null Group Template Defaults

**Location:** `/app/models/concerns/null/group.rb`

The NullGroup provides default values for users without a group context:
- poll_template_positions with standard ordering
- discussion_template_positions with basic templates
- Empty hidden template arrays

This enables template features for standalone discussions/polls.
