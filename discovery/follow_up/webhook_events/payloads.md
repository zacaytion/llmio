# Webhook Events: Payload Format Documentation

## Overview

Webhook payloads are generated via serializers in `/Users/z/Code/loomio/app/serializers/webhook/` and content templates in `/Users/z/Code/loomio/app/views/chatbot/`.

## Base Payload Structure

All webhook payloads share a common base structure:

```json
{
  "text": "...rendered markdown content...",
  "icon_url": "https://example.com/group-logo.png",
  "username": "Loomio"
}
```

| Field | Source | Description |
|-------|--------|-------------|
| `text` | Template rendering | Main message content in Markdown |
| `icon_url` | `group.self_or_parent_logo_url(128)` | Group logo URL (128px) |
| `username` | `AppConfig.theme[:site_name]` | Site name (default: "Loomio") |

## Format-Specific Variations

### Slack Format

**Serializer**: `/Users/z/Code/loomio/app/serializers/webhook/slack/event_serializer.rb`

Same as base, but:
- Uses Slack-specific templates from `app/views/chatbot/slack/`
- Omits `icon_url` and `username` (Slack handles these via webhook configuration)

```json
{
  "text": "...slack-formatted text..."
}
```

### Discord Format

**Serializer**: `/Users/z/Code/loomio/app/serializers/webhook/discord/event_serializer.rb`

```json
{
  "content": "...truncated to 1900 chars...",
  "text": "...full markdown text...",
  "icon_url": "https://example.com/group-logo.png",
  "username": "Loomio"
}
```

Note: `content` field is truncated to 1900 characters (Discord's 2000 char limit minus buffer).

### Microsoft Teams Format

**Serializer**: `/Users/z/Code/loomio/app/serializers/webhook/microsoft/event_serializer.rb`

Uses Microsoft MessageCard format:

```json
{
  "@type": "MessageCard",
  "@context": "http://schema.org/extensions",
  "themeColor": "#658AE7",
  "text": "...rendered markdown...",
  "sections": []
}
```

| Field | Value | Description |
|-------|-------|-------------|
| `@type` | `"MessageCard"` | Microsoft card type |
| `@context` | `"http://schema.org/extensions"` | Schema context |
| `themeColor` | Primary color from theme | Accent color for card |
| `sections` | `[]` | Empty (reserved for future use) |

### Webex Format

**Serializer**: `/Users/z/Code/loomio/app/serializers/webhook/webex/event_serializer.rb`

```json
{
  "markdown": "...same as text...",
  "text": "...rendered markdown...",
  "icon_url": "https://example.com/group-logo.png",
  "username": "Loomio"
}
```

Webex uses `markdown` field for message content.

## Text Content by Event Type

### Discussion Events

#### new_discussion

**Template**: `chatbot/{format}/discussion.text.erb`

Content includes:
1. Notification line: "[Actor] started a thread: [Title](url)"
2. Title with link
3. Description body (markdown rendered)
4. Attachments

Example:
```markdown
John Doe started a thread: [Project Planning](https://loomio.example.com/d/abc123)

**[Project Planning](https://loomio.example.com/d/abc123)**

Let's discuss the upcoming project milestones...

[attachment1.pdf](https://loomio.example.com/files/...)
```

#### discussion_edited

**Template**: `chatbot/{format}/discussion.text.erb`

Same as new_discussion, notification line reads: "[Actor] edited [Title](url)"

### Comment Events

#### new_comment

**Template**: `chatbot/{format}/comment.text.erb`

Content includes:
1. Notification line: "[Actor] commented on [Title](url)"
2. Comment body (markdown rendered)
3. Attachments

### Poll Events

#### poll_created

**Template**: `chatbot/{format}/poll.text.erb`

Content includes:
1. Notification line: "[Actor] started a [poll_type]: [Title](url)"
2. Title with link
3. Outcome (if exists)
4. Poll body/details
5. Vote options (if active, single choice)
6. Voting rules
7. Current results (if visible)

Example:
```markdown
Jane Smith started a proposal: [Approve budget](https://loomio.example.com/p/xyz789)

**[Approve budget](https://loomio.example.com/p/xyz789)**

Please review and vote on the Q1 budget proposal.

**Have your say**
- [Agree](https://loomio.example.com/p/xyz789?poll_option_id=1)
- [Abstain](https://loomio.example.com/p/xyz789?poll_option_id=2)
- [Disagree](https://loomio.example.com/p/xyz789?poll_option_id=3)

You have until January 15, 2024 at 5:00 PM
```

#### poll_edited

Same template as poll_created, notification line reads: "[Actor] edited [poll_type] [Title](url)"

#### poll_closing_soon

Same template, notification line reads: "[poll_type] closing soon: [Title](url)"

#### poll_expired

Same template, notification line reads: "[poll_type] has closed: [Title](url)"

Shows final results.

#### poll_closed_by_user

Same template, notification line reads: "[Actor] closed [poll_type]: [Title](url)"

Shows final results.

#### poll_reopened

Same template, notification line reads: "[Actor] re-opened [poll_type]: [Title](url)"

### Stance Events

#### stance_created

**Template**: `chatbot/{format}/stance.text.erb`

Content includes:
1. Notification line: "[Actor] voted on [Title](url)"
2. Poll title with link
3. Stance choices (votes/selections)
4. Reason (if provided)

Example for proposal:
```markdown
Bob Johnson voted on [Approve budget](https://loomio.example.com/p/xyz789)

**[Approve budget](https://loomio.example.com/p/xyz789)**

**Agree**

I support this proposal because it aligns with our goals.
```

Example for meeting poll:
```markdown
Alice Chen voted on [Team Meeting](https://loomio.example.com/p/abc456)

**[Team Meeting](https://loomio.example.com/p/abc456)**

**Can attend:**
- Monday 10am
- Tuesday 2pm

**Can't attend:**
- Wednesday 9am
```

#### stance_updated

Same as stance_created, notification line reads: "[Actor] changed their vote on [Title](url)"

### Outcome Events

#### outcome_created

**Template**: `chatbot/{format}/poll.text.erb`

Content includes:
1. Notification line: "[Actor] shared an outcome for [Title](url)"
2. Poll title
3. Outcome statement
4. Final results

Example:
```markdown
Jane Smith shared an outcome for [Approve budget](https://loomio.example.com/p/xyz789)

**[Approve budget](https://loomio.example.com/p/xyz789)**

**Outcome**
The budget has been approved. Implementation will begin next week.

**Results**
- Agree: 8 (80%)
- Abstain: 1 (10%)
- Disagree: 1 (10%)
```

#### outcome_updated

Same as outcome_created, notification line reads: "[Actor] updated the outcome for [Title](url)"

#### outcome_review_due

Same template, notification line reads: "Outcome review due for [Title](url)"

## Notification-Only Mode

When `chatbot.notification_only = true`, all events use:

**Template**: `chatbot/{format}/notification.text.erb`

Minimal content:
```markdown
[Actor] [action] [Title](url)
```

No body, results, or additional details.

## Template Files Reference

| Template | Events Using | Content |
|----------|-------------|---------|
| `discussion.text.erb` | new_discussion, discussion_edited | Notification + title + body + attachments |
| `comment.text.erb` | new_comment | Notification + body + attachments |
| `poll.text.erb` | poll_*, outcome_* | Notification + title + outcome + body + vote + rules + results |
| `stance.text.erb` | stance_created, stance_updated | Notification + title + choices + reason |
| `notification.text.erb` | Any (notification_only=true) | Notification line only |

## Partial Templates

| Partial | Purpose |
|---------|---------|
| `_notification.text.erb` | Event header with i18n notification text |
| `_title.text.erb` | Bold title with link |
| `_body.text.erb` | Rendered markdown body + attachments |
| `_results.text.erb` | Poll results display |
| `_vote.text.erb` | Vote options with links |
| `_rules.text.erb` | Poll rules/settings |
| `_outcome.text.erb` | Outcome statement |
| `_undecided.text.erb` | Undecided voters count |
| `_attachments.text.erb` | File attachment links |
| `_stance_choices.text.erb` | Vote selections |
| `_meeting_stance_choices.text.erb` | Meeting availability selections |
| `_simple.text.erb` | Simple poll results |
| `_meeting.text.erb` | Meeting poll results |

## Localization

All webhook content is localized using:
1. `chatbot.author.locale` (if author exists)
2. `chatbot.group.creator.locale` (fallback)

Notification text keys follow pattern: `notifications.without_title.{event_kind}`

## Data Available in Templates

| Variable | Type | Available In |
|----------|------|--------------|
| `@event` | Event | All templates |
| `@poll` | Poll | Poll/outcome/stance templates |
| `@recipient` | LoggedOutUser | All templates (for timezone/locale) |
| `event.user` | User | Actor who triggered event |
| `event.eventable` | Model | The affected record |
| `event.recipient_message` | String | Custom message from announcements |

## URL Generation

All URLs use `polymorphic_url` with `CANONICAL_HOST`:
- Discussions: `https://{host}/d/{key}`
- Polls: `https://{host}/p/{key}`
- Comments: Links to parent discussion with anchor

## Attachment Handling

Attachments are rendered as markdown links:
```markdown
[filename.pdf](https://loomio.example.com/files/abc123)
```

## Error Handling

If webhook delivery fails (non-200 response):
- Error logged to Sentry with:
  - Chatbot ID
  - Event ID
  - Response code
  - Response body
- No retry mechanism visible in code
