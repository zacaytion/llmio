# Shared Concerns Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/concerns/*.rb`
- `/app/models/concerns/events/*.rb`

---

## Overview

Loomio uses ActiveSupport::Concern modules extensively to share behavior across models. This document catalogs all shared concerns and their functionality.

---

## Core Concerns

### HasRichText

**File:** `/app/models/concerns/has_rich_text.rb`

Provides rich text editing support with HTML sanitization, file attachments, and task extraction.

**Used By:** User, Group, Discussion, Comment, Poll, Stance, Outcome

**Configuration:**
```ruby
is_rich_text on: :description  # Field name
```

**Provides:**

| Feature | Description |
|---------|-------------|
| HTML sanitization | Whitelist-based with allowed tags/attributes |
| File attachments | `has_many_attached :files` |
| Image attachments | `has_many_attached :image_files` |
| Task extraction | Parses checklist items from content |
| Link previews | `sanitize_link_previews` callback |
| Content locale | Language detection via CLD |
| Visible text | `{field}_visible_text` method |
| Blank check | `body_is_blank?` method |

**Allowed HTML Tags:**
```
strong, em, b, i, p, s, code, pre, big, div, small, hr, br, span, mark,
h1, h2, h3, ul, ol, li, abbr, a, img, video, audio, blockquote, table,
thead, th, tr, td, iframe, u
```

**Allowed Attributes:**
```
href, src, alt, title, data-type, data-iframe-container, data-done,
data-mention-id, poster, controls, data-author-id, data-uid, data-checked,
data-due-on, data-color, data-remind, width, height, target, colspan,
rowspan, data-text-align
```

**Callbacks:**
- `before_save :sanitize_{field}!`
- `before_save :update_content_locale`
- `before_save :build_attachments`
- `before_save :sanitize_link_previews`
- `after_save :parse_and_update_tasks_{field}!`
- `after_discard` - Discards associated tasks
- `after_undiscard` - Undiscards associated tasks

---

### HasMentions

**File:** `/app/models/concerns/has_mentions.rb`

Extracts @mentions from content and provides mention-related queries.

**Used By:** Discussion, Comment, Poll, Stance, Outcome

**Configuration:**
```ruby
is_mentionable on: :description  # Field name
```

**Provides:**

| Method | Description |
|--------|-------------|
| `mentioned_usernames` | Extracts usernames from content |
| `mentioned_user_ids` | Extracts user IDs from HTML data attributes |
| `mentioned_users` | Users who are mentioned |
| `mentioned_groups` | Groups who are @mentioned |
| `mentioned_group_users` | Members of mentioned groups |
| `newly_mentioned_users` | Users mentioned for first time (avoids re-notifying on edit) |
| `newly_mentioned_groups` | Groups mentioned for first time |
| `already_mentioned_user_ids` | Previously notified user IDs |
| `already_mentioned_group_ids` | Previously notified group IDs |

**Mention Parsing:**
- **Markdown:** Uses twitter-text gem to extract `@username`
- **HTML:** Parses `span[data-mention-id]` attributes via Nokogiri

---

### HasEvents

**File:** `/app/models/concerns/has_events.rb`

Provides polymorphic event associations.

**Used By:** Discussion, Comment, Poll, Stance, Outcome, Membership, Group

**Provides:**
```ruby
has_many :events, -> { includes :user, :eventable }, as: :eventable, dependent: :destroy
has_many :notifications, through: :events
has_many :users_notified, -> { distinct }, through: :notifications, source: :user
```

---

### HasCreatedEvent

**File:** `/app/models/concerns/has_created_event.rb`

Tracks the "created" event for a record.

**Used By:** Discussion, Comment, Poll, Stance, Outcome

**Provides:**

| Method | Description |
|--------|-------------|
| `created_event` | Finds event with `kind: created_event_kind` |
| `created_event_kind` | Returns symbol like `:discussion_created` |
| `create_missing_created_event!` | Creates event if missing |

Default `created_event_kind`:
```ruby
def created_event_kind
  :"#{self.class.name.downcase}_created"
end
```

---

### HasVolume

**File:** `/app/models/concerns/has_volume.rb`

Provides notification volume settings.

**Used By:** Membership, DiscussionReader, Stance

**Provides:**

| Feature | Description |
|---------|-------------|
| `enum :volume` | 0: mute, 1: quiet, 2: normal, 3: loud |
| Scopes | `volume`, `volume_at_least`, `email_notifications`, `app_notifications` |
| Instance methods | `set_volume!`, `volume_is_*?` predicates |

**Scopes:**
```ruby
scope :volume, ->(volume) { where(volume: volumes[volume]) }
scope :volume_at_least, ->(volume) { where('volume >= ?', volumes[volume]) }
scope :email_notifications, -> { where('volume >= ?', volumes[:normal]) }
scope :app_notifications, -> { where('volume >= ?', volumes[:quiet]) }
```

---

### HasAvatar

**File:** `/app/models/concerns/has_avatar.rb`

Provides avatar management with Gravatar and upload support.

**Used By:** User

**Includes:** AvatarInitials, Routing, Gravtastic

**Provides:**

| Method | Description |
|--------|-------------|
| `avatar_kind` | Returns: initials, uploaded, gravatar, mdi-duck (deactivated), mdi-email-outline (no name) |
| `avatar_url(size)` | Returns appropriate URL based on avatar_kind |
| `thumb_url` | 128px avatar |
| `uploaded_avatar_url(size)` | ActiveStorage representation URL |
| `avatar_initials_url(size)` | ui-avatars.com URL |
| `has_gravatar?` | Checks Gravatar availability |
| `set_default_avatar_kind` | Sets initial avatar type |

---

### HasExperiences

**File:** `/app/models/concerns/has_experiences.rb`

Manages feature flag/tutorial completion tracking in JSONB.

**Used By:** User, Membership

**Provides:**
```ruby
def experienced!(key, toggle = true)
  experiences[key] = toggle
  save
end
```

---

### HasTags

**File:** `/app/models/concerns/has_tags.rb`

Provides tag management with group-scoped tags.

**Used By:** Discussion, Poll

**Provides:**

| Method | Description |
|--------|-------------|
| `tag_models` | Tag records matching this record's tags |
| `update_group_tags` | Async updates group tag counts |

**Callbacks:**
- `after_save :update_group_tags`

---

### HasTimeframe

**File:** `/app/models/concerns/has_timeframe.rb`

Provides time-based query scopes.

**Used By:** Membership, Discussion, DiscussionReader

**Provides:**
```ruby
scope :within, ->(since, till, field = nil) {
  where("#{table_name}.#{field || :created_at} BETWEEN ? AND ?",
        since || 100.years.ago, till || 100.years.from_now)
}
scope :until, ->(till) { within(nil, till) }
scope :since, ->(since) { within(since, nil) }

def self.has_timeframe?
  true
end
```

---

### HasTokens

**File:** `/app/models/concerns/has_tokens.rb`

Provides secure token initialization.

**Used By:** User, Group, Membership, DiscussionReader, Stance

**Configuration:**
```ruby
extend HasTokens
initialized_with_token :token
initialized_with_token :unsubscribe_token
```

**Implementation:**
```ruby
def initialized_with_token(column, method = nil)
  after_initialize do
    send(:"#{column}=", send(column) || method&.call || self.class.generate_unique_secure_token) if respond_to?("#{column}=")
  end
end
```

---

### HasDefaults

**File:** `/app/models/concerns/has_defaults.rb`

Provides default value initialization.

**Used By:** User

**Configuration:**
```ruby
extend HasDefaults
initialized_with_default :some_field, -> { default_value }
```

---

### HasCustomFields

**File:** `/app/models/concerns/has_custom_fields.rb`

Provides dynamic accessors for JSONB custom_fields.

**Used By:** Poll, Outcome, Event, Identity

**Configuration:**
```ruby
extend HasCustomFields
set_custom_fields :meeting_duration, :time_zone, :can_respond_maybe
```

**Generates:**
```ruby
def meeting_duration
  self[:custom_fields]['meeting_duration']
end

def meeting_duration=(value)
  self[:custom_fields]['meeting_duration'] = value
end
```

---

### ReadableUnguessableUrls

**File:** `/app/models/concerns/readable_unguessable_urls.rb`

Provides 8-character random URL keys.

**Used By:** User, Group, Discussion, Poll

**Provides:**

| Feature | Description |
|---------|-------------|
| `key` attribute | 8-char alphanumeric string |
| FriendlyId integration | Find by key |
| Auto-generation | `before_validation :set_key` |
| Uniqueness check | Regenerates if collision |

**Key Generation:**
```ruby
KEY_LENGTH = 8

def generate_key
  (('a'..'z').to_a + ('A'..'Z').to_a + (0..9).to_a).sample(KEY_LENGTH).join
end
```

---

### MessageChannel

**File:** `/app/models/concerns/message_channel.rb`

Provides real-time pub/sub channel naming.

**Used By:** User, Group, Discussion, Poll

**Provides:**
```ruby
def message_channel
  "/#{self.class.to_s.downcase}-#{self.key}"
end
```

Examples:
- User: `/user-abc123XY`
- Group: `/group-def456ZW`

---

### SelfReferencing

**File:** `/app/models/concerns/self_referencing.rb`

Provides self-referential methods for polymorphic contexts.

**Used By:** User, Group, Discussion, Poll

**Provides:**
```ruby
# For a Discussion model:
def discussion
  self
end

def discussion_id
  self.id
end
```

---

### Reactable

**File:** `/app/models/concerns/reactable.rb`

Provides emoji reaction support.

**Used By:** Discussion, Comment, Poll, Stance, Outcome

**Provides:**
```ruby
has_many :reactions, -> { joins(:user).where("users.deactivated_at": nil) },
         dependent: :destroy, as: :reactable
has_many :reactors, through: :reactions, source: :user
```

---

### Translatable

**File:** `/app/models/concerns/translatable.rb`

Provides content translation support.

**Used By:** Group, Discussion, Comment, Poll, Stance, Outcome, PollOption

**Configuration:**
```ruby
is_translatable on: [:title, :description], load_via: :find_by_key!, id_field: :key
```

**Provides:**

| Feature | Description |
|---------|-------------|
| `translations` association | Has many translations |
| `translatable_fields_modified?` | Checks if translatable fields changed |
| `update_translations` | Async translation update |

**Callbacks:**
- `before_update :update_translations, if: :translatable_fields_modified?`

---

### Searchable

**File:** `/app/models/concerns/searchable.rb`

Provides full-text search integration with pg_search.

**Used By:** Discussion, Comment, Poll, Stance, Outcome

**Provides:**
```ruby
include PgSearch::Model
multisearchable

def self.rebuild_pg_search_documents
  connection.execute pg_search_insert_statement
end

def self.pg_search_insert_statement(id: nil, author_id: nil, discussion_id: nil)
  raise "expected to be overwritten"  # Each model implements custom SQL
end
```

**Overrides PgSearch default:**
```ruby
def update_pg_search_document
  PgSearch::Document.where(searchable: self).delete_all
  ActiveRecord::Base.connection.execute(self.class.pg_search_insert_statement(id: self.id))
end
```

---

### GroupPrivacy

**File:** `/app/models/concerns/group_privacy.rb`

Provides privacy settings management for groups.

**Used By:** Group

**Constants:**
```ruby
DISCUSSION_PRIVACY_OPTIONS = ['public_only', 'private_only', 'public_or_private'].freeze
MEMBERSHIP_GRANTED_UPON_OPTIONS = ['request', 'approval', 'invitation'].freeze
```

**Provides:**

| Method | Description |
|--------|-------------|
| `group_privacy=` | Sets privacy mode: 'open', 'closed', 'secret' |
| `group_privacy` | Returns computed privacy mode |
| `is_hidden_from_public?` | Visibility check |
| `private_discussions_only?` | Discussion privacy check |
| `public_discussions_only?` | Discussion privacy check |
| `membership_granted_upon_*?` | Membership grant checks |
| `discussion_private_default` | Default for new discussions |

**Validations:**
- `validate_parent_members_can_see_discussions`
- `validate_is_visible_to_parent_members`
- `validate_discussion_privacy_options`
- `validate_trial_group_cannot_be_public`

---

### NoSpam

**File:** `/app/models/concerns/no_spam.rb`

Provides spam content filtering.

**Used By:** User, Group, Poll

**Configuration:**
```ruby
extend NoSpam
no_spam_for :name, :description
```

**Implementation:**
```ruby
SPAM_REGEX = Regexp.new(ENV.fetch('SPAM_REGEX', "(diide\.com|gusronk\.com)"), 'i')

def no_spam_for(*fields)
  Array(fields).each do |field|
    validates field, format: { without: SPAM_REGEX, message: "no spam" }
  end
end
```

---

### NoForbiddenEmails

**File:** `/app/models/concerns/no_forbidden_emails.rb`

Excludes system email addresses from user registration.

**Used By:** User

**Provides:**
```ruby
FORBIDDEN_EMAIL_ADDRESSES = [ENV.fetch('DECIDE_EMAIL', "decide@#{ENV['CANONICAL_HOST']}")]

validates_exclusion_of :email, in: FORBIDDEN_EMAIL_ADDRESSES
```

---

### AvatarInitials

**File:** `/app/models/concerns/avatar_initials.rb`

Computes display initials from name.

**Used By:** User (via HasAvatar)

**Provides:**
```ruby
def set_avatar_initials
  self.avatar_initials = get_avatar_initials[0..2]
end

def get_avatar_initials
  if deactivated_at
    "DU"
  elsif name.blank? || name == email
    email.to_s[0..1]
  else
    name.split.map(&:first).join
  end.upcase.gsub(/(\W|\d)/, "")
end
```

---

### Routing

**File:** `/app/models/concerns/routing.rb`

Provides URL helper access in models.

**Used By:** User (via HasAvatar)

---

### UsesOrganisationScope

**File:** `/app/models/concerns/uses_organisation_scope.rb`

Provides organization-scoped queries.

---

## Event Notification Concerns

### Events::LiveUpdate

**File:** `/app/models/concerns/events/live_update.rb`

Publishes real-time updates to connected clients.

```ruby
def trigger!
  super
  notify_clients!
end

def notify_clients!
  return unless eventable
  if eventable.group_id
    MessageChannelService.publish_models([self], group_id: eventable.group.id)
  end
  if eventable.respond_to?(:guests)
    eventable.guests.find_each do |user|
      MessageChannelService.publish_models([self], user_id: user.id)
    end
  end
end
```

---

### Events::Notify::InApp

**File:** `/app/models/concerns/events/notify/in_app.rb`

Generates in-app notifications.

```ruby
def trigger!
  super
  notify_users!
end

def notify_users!
  notifications.import(built_notifications)
  built_notifications.each { |n|
    MessageChannelService.publish_models(Array(n), user_id: n.user_id)
  }
end

def notification_for(recipient)
  I18n.with_locale(recipient.locale) do
    notifications.build(
      user: recipient,
      actor: notification_actor,
      url: notification_url,
      translation_values: notification_translation_values
    )
  end
end
```

---

### Events::Notify::ByEmail

**File:** `/app/models/concerns/events/notify/by_email.rb`

Sends email notifications via EventMailer.

```ruby
def trigger!
  super
  email_users!
end

def email_users!
  email_recipients.active.no_spam_complaints.uniq.pluck(:id).each do |recipient_id|
    EventMailer.event(recipient_id, self.id).deliver_later
  end
end
```

---

### Events::Notify::Mentions

**File:** `/app/models/concerns/events/notify/mentions.rb`

Handles @mention notifications, publishing separate UserMentioned/GroupMentioned events.

```ruby
def trigger!
  super
  return if silence_mentions?
  notify_mentioned_groups!
  notify_mentioned_users!
end

def notify_mentioned_users!
  return if eventable.newly_mentioned_users.empty?
  Events::UserMentioned.publish!(eventable, user, eventable.newly_mentioned_users.pluck(:id))
end
```

---

### Events::Notify::Subscribers

**File:** `/app/models/concerns/events/notify/subscribers.rb`

Notifies subscribed/participating users.

---

### Events::Notify::Chatbots

**File:** `/app/models/concerns/events/notify/chatbots.rb`

Sends webhook notifications to configured chatbots.

---

### Events::Notify::Author

**File:** `/app/models/concerns/events/notify/author.rb`

Notifies content authors.

---

## Model Export Relations

### DiscussionExportRelations

**File:** `/app/models/concerns/discussion_export_relations.rb`

Provides optimized includes for discussion exports.

---

### GroupExportRelations

**File:** `/app/models/concerns/group_export_relations.rb`

Provides optimized includes for group exports.

---

## Concern Usage Summary

| Concern | User | Group | Membership | Discussion | Comment | Poll | Stance | Outcome | Event |
|---------|:----:|:-----:|:----------:|:----------:|:-------:|:----:|:------:|:-------:|:-----:|
| HasRichText | X | X | | X | X | X | X | X | |
| HasMentions | | | | X | X | X | X | X | |
| HasEvents | | X | X | X | X | X | X | X | |
| HasCreatedEvent | | | | X | X | X | X | X | |
| HasVolume | | | X | | | | X | | |
| HasAvatar | X | | | | | | | | |
| HasExperiences | X | | X | | | | | | |
| HasTags | | | | X | | X | | | |
| HasTimeframe | | | X | X | | | | | |
| HasTokens | X | X | X | | | | X | | |
| HasCustomFields | | | | | | X | | X | X |
| ReadableUnguessableUrls | X | X | | X | | X | | | |
| MessageChannel | X | X | | X | | X | | | |
| SelfReferencing | X | X | | X | | X | | | |
| Reactable | | | | X | X | X | X | X | |
| Translatable | | X | | X | X | X | X | X | |
| Searchable | | | | X | X | X | X | X | |
| GroupPrivacy | | X | | | | | | | |
| NoSpam | X | X | | | | X | | | |
| Discard::Model | | | | X | X | X | | | |

---

## Uncertainties

1. **Routing concern** - Implementation details not visible in model files
2. **UsesOrganisationScope** - Limited visibility into usage patterns
3. **Export relations** - Complex eager loading logic not fully documented

**Confidence Level:** HIGH for core concerns, MEDIUM for less commonly used ones.
