# Loomio Mention System - Technical Documentation

## Executive Summary

The Loomio mention system supports two distinct mention types (@user and @group) with dual parsing strategies based on content format (HTML vs Markdown). Mentions are resolved at **write time** with user/group identifiers stored directly in the content. Event-driven notifications are generated asynchronously via Sidekiq workers.

---

## 1. Mention Syntax

### Confidence: HIGH

### HTML Format (Tiptap/Rich Text Editor)

Mentions are stored as decorated `<span>` elements with `data-mention-id` attributes:

```html
<span class="mention" data-mention-id="username">@Display Name</span>
```

**Key observations:**
- The `data-mention-id` attribute stores either a **username** (string) or a **user ID** (numeric string)
- Legacy content may use numeric IDs; newer content uses usernames
- The visible text includes the `@` prefix followed by the display name

**Source:** `/Users/z/Code/loomio/vue/src/components/lmo_textarea/extension_mention.js:11-12`

```javascript
parseHTML: element => ( element.getAttribute('data-mention-id') ),
renderHTML: attributes => ( { 'data-mention-id': attributes.id } ),
```

### Markdown Format

Plain text `@username` syntax following Twitter conventions:

```
Hello @johndoe, please review this.
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:14`

```ruby
extract_mentioned_screen_names(mentionable_text).uniq - [self.author&.username]
```

The system uses the `twitter-text` gem (Loomio fork: `github: 'loomio/twitter-text'`) to extract mentions via `Twitter::Extractor#extract_mentioned_screen_names`.

---

## 2. Parsing Mechanism

### Confidence: HIGH

The parsing logic is centralized in `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb`.

### Format Detection

The system determines parsing strategy based on the `*_format` field (e.g., `body_format`, `description_format`):

```ruby
def text_format
  send("#{self.class.mentionable_fields.first}_format")
end
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:65-67`

### HTML Parsing (Default for Tiptap)

Uses Nokogiri to extract `data-mention-id` from span elements:

```ruby
def mentioned_usernames
  if text_format == "md"
    extract_mentioned_screen_names(mentionable_text).uniq - [self.author&.username]
  else
    Nokogiri::HTML::fragment(mentionable_text).search("span[data-mention-id]").map do |el|
      el['data-mention-id']
    end.filter { |id_or_username| id_or_username.to_i.to_s != id_or_username }
  end
end
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:12-19`

**Important:** The filter `id_or_username.to_i.to_s != id_or_username` separates usernames (strings) from numeric IDs.

### Markdown Parsing

Uses `twitter-text` gem's `extract_mentioned_screen_names` method:

```ruby
include Twitter::Extractor
# ...
extract_mentioned_screen_names(mentionable_text).uniq - [self.author&.username]
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:3,14`

### User Resolution

Mentions are resolved against the discussion/group membership:

```ruby
def mentioned_users
  members.where("users.username in (:usernames) or users.id in (:ids)",
                usernames: mentioned_usernames, ids: mentioned_user_ids)
end
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:31-33`

**Critical security constraint:** Only members of the relevant context (group/discussion) can be mentioned. Non-members are silently ignored.

---

## 3. Storage Format

### Confidence: HIGH

### Content Storage

Mentions are stored **inline within the content field** of the model:

| Model | Mentionable Field | Format Field |
|-------|-------------------|--------------|
| Comment | `body` | `body_format` |
| Discussion | `description` | `description_format` |
| Poll | `details` | `details_format` |
| Stance | `reason` | `reason_format` |
| Outcome | `statement` | `statement_format` |

**Source:** Grep results from `/Users/z/Code/loomio/app/models/`:
- `comment.rb:51`: `is_mentionable on: :body`
- `discussion.rb:70`: `is_mentionable on: :description`
- `poll.rb:140`: `is_mentionable on: :details`
- `stance.rb:56`: `is_mentionable on: :reason`
- `outcome.rb:67`: `is_mentionable on: :statement`

### Database Defaults

Format fields default to `'md'` (Markdown):

```ruby
add_column :comments, :body_format, :string, default: "md", null: false, limit: 10
```

**Source:** `/Users/z/Code/loomio/db/migrate/20190205050843_add_format_field_to_textareas.rb:4`

### HTML Sanitization

The `data-mention-id` attribute is whitelisted for sanitization:

```ruby
attributes = %w[... data-mention-id ...]
self[field] = Rails::Html::WhiteListSanitizer.new.sanitize(self[field], tags: tags, attributes: attributes)
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_rich_text.rb:19`

---

## 4. Event Generation

### Confidence: HIGH

### Event Flow

1. **Content created/updated** - Service layer saves model (e.g., `CommentService.create`)
2. **Base event published** - `Events::NewComment.publish!(comment)` creates event record
3. **Async trigger** - `PublishEventWorker.perform_async(event.id)` enqueues Sidekiq job
4. **Event triggered** - Worker calls `Event.sti_find(event_id).trigger!`
5. **Mention events created** - `Events::Notify::Mentions#trigger!` creates mention events

**Source flow:**
- `/Users/z/Code/loomio/app/services/comment_service.rb:9`
- `/Users/z/Code/loomio/app/models/event.rb:64`
- `/Users/z/Code/loomio/app/workers/publish_event_worker.rb:5`

### Mention Notification Module

The `Events::Notify::Mentions` concern is included by parent events:

```ruby
module Events::Notify::Mentions
  def trigger!
    super
    return if silence_mentions?

    notify_mentioned_groups!
    notify_mentioned_users!
  end

  def notify_mentioned_users!
    return if eventable.newly_mentioned_users.empty?
    Events::UserMentioned.publish! eventable, user, eventable.newly_mentioned_users.pluck(:id)
  end

  def notify_mentioned_groups!
    return if eventable.newly_mentioned_groups.empty?
    Events::GroupMentioned.publish! eventable, user, eventable.newly_mentioned_groups.pluck(:id), id
  end
end
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/events/notify/mentions.rb:1-26`

### Events That Support Mentions

| Event Class | Source File |
|-------------|-------------|
| `Events::NewComment` | `app/models/events/new_comment.rb:3` |
| `Events::CommentEdited` | `app/models/events/comment_edited.rb:3` |
| `Events::NewDiscussion` | `app/models/events/new_discussion.rb:5` |
| `Events::DiscussionEdited` | `app/models/events/discussion_edited.rb:5` |
| `Events::PollCreated` | `app/models/events/poll_created.rb:3` |
| `Events::PollEdited` | `app/models/events/poll_edited.rb:5` |
| `Events::StanceCreated` | `app/models/events/stance_created.rb:4` |
| `Events::OutcomeCreated` | `app/models/events/outcome_created.rb:2` |
| `Events::OutcomeUpdated` | `app/models/events/outcome_updated.rb:2` |

### Re-mention Prevention

The system tracks previously mentioned users to avoid duplicate notifications on edits:

```ruby
def newly_mentioned_users
  mentioned_users.where.not(id: already_mentioned_user_ids)
end

def already_mentioned_user_ids
  notifications.user_mentions.pluck(:user_id)
end
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:50-57`

### UserMentioned Event

Stores mentioned user IDs in `custom_fields`:

```ruby
def self.publish!(model, actor, user_ids)
  super model, user: actor, custom_fields: { user_ids: }
end
```

**Source:** `/Users/z/Code/loomio/app/models/events/user_mentioned.rb:5-6`

### GroupMentioned Event

Expands group membership for notifications, excluding already-mentioned individuals:

```ruby
def already_mentioned_user_ids
  eventable.mentioned_users.pluck(:id)
end

def scope
  Membership
    .active
    .accepted
    .where(group_id: group_ids)
    .where.not(user_id: already_mentioned_user_ids)
    .where.not(user_id: already_notified_user_ids)
end
```

**Source:** `/Users/z/Code/loomio/app/models/events/group_mentioned.rb:26-31`

---

## 5. Group Mentions

### Confidence: HIGH

Groups can be mentioned using their `handle` field:

```ruby
def mentioned_groups
  group_ids = Group.published.where(id: group.id).where(handle: mentioned_usernames).filter { |group|
    author.can? :notify, group
  }.map(&:id)
  Group.where(id: group_ids)
end
```

**Source:** `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb:40-44`

**Constraints:**
- Only the current group can be mentioned (not arbitrary groups)
- Author must have `:notify` permission on the group
- Group must have a non-nil `handle`

---

## 6. Frontend Mention UI

### Confidence: HIGH

### Tiptap Extension

Custom mention extension for the Tiptap editor:

```javascript
export const CustomMention = Mention.extend({
  addAttributes() {
    return {
      id: {
        default: null,
        parseHTML: element => ( element.getAttribute('data-mention-id') ),
        renderHTML: attributes => ( { 'data-mention-id': attributes.id } ),
      },
      label: {
        default: null,
        parseHTML: element => ( element.getAttribute('data-label') || element.innerText.split('@').join('') )
      },
    }
  },
  renderHTML({ node, HTMLAttributes }) {
    return ['span', mergeAttributes(this.options.HTMLAttributes, HTMLAttributes), `@${node.attrs.label}`]
  },
  parseHTML() {
    return [{ tag: 'span[data-mention-id]' }]
  },
})
```

**Source:** `/Users/z/Code/loomio/vue/src/components/lmo_textarea/extension_mention.js:6-31`

### Mention Autocomplete API

Endpoint: `GET /api/v1/mentions`

Returns mentionable users and groups based on context:

```ruby
def user_mention(user)
  { handle: user.username, name: user.name }
end

def group_mention(group)
  { handle: group.handle, name: group.full_name }
end
```

**Source:** `/Users/z/Code/loomio/app/controllers/api/v1/mentions_controller.rb:41-53`

### Frontend Mention Selection

When a user selects a mention from the dropdown:

```javascript
selectRow(row) {
  this.insertMention({
    id: row.handle,
    label: row.name
  });
  this.editor.chain().focus();
}
```

**Source:** `/Users/z/Code/loomio/vue/src/components/lmo_textarea/mentioning.js:143-148`

---

## 7. Architectural Diagram

```
+------------------+     +------------------+     +------------------+
|  Tiptap Editor   | --> |  CustomMention   | --> |  HTML Content    |
|  (Vue Frontend)  |     |  Extension       |     |  with data-*     |
+------------------+     +------------------+     +------------------+
                                                          |
                                                          v
+------------------+     +------------------+     +------------------+
|  Rails Model     | <-- |  HasRichText     | <-- |  API Request     |
|  (Comment/etc)   |     |  Sanitization    |     |  (body + format) |
+------------------+     +------------------+     +------------------+
         |
         v
+------------------+     +------------------+     +------------------+
|  Service Layer   | --> |  Event.publish!  | --> |  PublishEvent    |
|  (CommentService)|     |                  |     |  Worker (Sidekiq)|
+------------------+     +------------------+     +------------------+
                                                          |
                                                          v
+------------------+     +------------------+     +------------------+
|  Events::Notify  | --> |  HasMentions     | --> |  UserMentioned   |
|  ::Mentions      |     |  (parse mentions)|     |  GroupMentioned  |
+------------------+     +------------------+     +------------------+
                                                          |
                                                          v
                                                  +------------------+
                                                  |  Notifications   |
                                                  |  (in-app + email)|
                                                  +------------------+
```

---

## 8. Key Findings Summary

| Aspect | Finding | Confidence |
|--------|---------|------------|
| Mention syntax (HTML) | `<span data-mention-id="username">@Name</span>` | HIGH |
| Mention syntax (MD) | `@username` (Twitter-style) | HIGH |
| Parsing mechanism | Nokogiri (HTML) / twitter-text gem (MD) | HIGH |
| Resolution timing | Write time (stored with identifier) | HIGH |
| Storage location | Inline in content field | HIGH |
| Format detection | `*_format` column (html/md) | HIGH |
| Event trigger | Async via Sidekiq worker | HIGH |
| User event | `Events::UserMentioned` with user_ids in custom_fields | HIGH |
| Group event | `Events::GroupMentioned` with group_ids in custom_fields | HIGH |
| Re-mention prevention | Tracks via notifications table | HIGH |
| Security constraint | Only group/discussion members can be mentioned | HIGH |

---

## 9. Files Referenced

- `/Users/z/Code/loomio/app/models/concerns/has_mentions.rb` - Core mention parsing logic
- `/Users/z/Code/loomio/app/models/concerns/events/notify/mentions.rb` - Event notification module
- `/Users/z/Code/loomio/app/models/events/user_mentioned.rb` - User mention event
- `/Users/z/Code/loomio/app/models/events/group_mentioned.rb` - Group mention event
- `/Users/z/Code/loomio/app/controllers/api/v1/mentions_controller.rb` - Autocomplete API
- `/Users/z/Code/loomio/vue/src/components/lmo_textarea/extension_mention.js` - Tiptap extension
- `/Users/z/Code/loomio/vue/src/components/lmo_textarea/mentioning.js` - Vue mention handling
- `/Users/z/Code/loomio/app/models/concerns/has_rich_text.rb` - HTML sanitization whitelist
- `/Users/z/Code/loomio/app/workers/publish_event_worker.rb` - Async event processing
- `/Users/z/Code/loomio/app/services/comment_service.rb` - Example service layer

---

*Generated: 2026-02-01*
