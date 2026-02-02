# Group Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/group.rb`
- `/app/models/formal_group.rb`
- `/app/models/guest_group.rb`
- `/app/models/null_group.rb`
- `/app/models/concerns/group_privacy.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Group model represents organization containers with hierarchical structure. Groups can have subgroups (one level of nesting only), configurable privacy settings, and member permission flags.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `name` | string(255) | - | YES | max 250, presence required | Group name |
| `full_name` | string(255) | - | YES | - | Computed: "Parent - Child" for subgroups |
| `description` | text | - | YES | max via AppConfig | Rich text description |
| `description_format` | string(10) | "md" | NO | "md" or "html" | Description format |
| `handle` | citext | - | YES | UNIQUE, parameterized | URL slug |
| `key` | string(255) | - | YES | UNIQUE | Public URL key (8 chars) |
| `token` | string | - | YES | UNIQUE | Secret API token |

### Hierarchy

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `parent_id` | integer | - | YES | FK to groups | Parent group (self-referential) |
| `creator_id` | integer | - | YES | FK to users | User who created the group |
| `subscription_id` | integer | - | YES | FK to subscriptions, must be null for subgroups | Billing subscription |

### Privacy & Visibility

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `is_visible_to_public` | boolean | true | - | - | Listed in explore |
| `is_visible_to_parent_members` | boolean | false | - | - | Visible to parent group members |
| `discussion_privacy_options` | string | "private_only" | - | "public_only", "private_only", "public_or_private" | Allowed discussion privacy |
| `membership_granted_upon` | string | "approval" | - | "approval", "request", "invitation" | How membership is granted |
| `listed_in_explore` | boolean | false | - | - | Show in public explore page |
| `parent_members_can_see_discussions` | boolean | false | - | - | Parent group visibility |

### Permission Flags

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `members_can_add_members` | boolean | false | Members can invite new members |
| `members_can_add_guests` | boolean | true | Members can invite guests |
| `members_can_edit_discussions` | boolean | true | Members can edit any discussion |
| `members_can_edit_comments` | boolean | true | Members can edit own comments |
| `members_can_delete_comments` | boolean | true | Members can delete own comments |
| `members_can_raise_motions` | boolean | true | Members can create polls |
| `members_can_start_discussions` | boolean | true | Members can create discussions |
| `members_can_create_subgroups` | boolean | false | Members can create child groups |
| `members_can_announce` | boolean | true | Members can send notifications |
| `members_can_vote` | boolean | true | **DEPRECATED** - unused |
| `admins_can_edit_user_content` | boolean | true | Admins can edit others' content |

### Thread Defaults

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `new_threads_max_depth` | integer | 3 | Default reply nesting depth |
| `new_threads_newest_first` | boolean | false | Default sort order |
| `can_start_polls_without_discussion` | boolean | false | Standalone polls allowed |

### Counter Caches

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `memberships_count` | integer | 0 | Total memberships |
| `admin_memberships_count` | integer | 0 | Admin count |
| `pending_memberships_count` | integer | 0 | Pending invitations |
| `delegates_count` | integer | 0 | Delegate count |
| `discussions_count` | integer | 0 | Total discussions |
| `open_discussions_count` | integer | 0 | Open discussions |
| `closed_discussions_count` | integer | 0 | Closed discussions |
| `public_discussions_count` | integer | 0 | Public discussions |
| `polls_count` | integer | 0 | Total polls |
| `closed_polls_count` | integer | 0 | Closed polls |
| `proposal_outcomes_count` | integer | 0 | Outcomes count |
| `invitations_count` | integer | 0 | Sent invitations |
| `subgroups_count` | integer | 0 | Child groups |
| `discussion_templates_count` | integer | 0 | Discussion templates |
| `poll_templates_count` | integer | 0 | Poll templates |
| `closed_motions_count` | integer | 0 | Legacy counter |
| `recent_activity_count` | integer | 0 | Activity metric |

### Media Attachments

| Column | Type | Description |
|--------|------|-------------|
| `cover_photo_*` | Paperclip fields | Legacy cover photo |
| `logo_*` | Paperclip fields | Legacy logo |

**Note:** ActiveStorage attachments `cover_photo` and `logo` replace Paperclip.

### Location

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `country` | string | - | GeoIP country |
| `region` | string | - | GeoIP region |
| `city` | string | - | GeoIP city |

### Status

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `created_at` | datetime | - | Creation timestamp |
| `updated_at` | datetime | - | Last update |
| `archived_at` | datetime | - | Soft delete timestamp |

### JSONB Fields

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |
| `info` | jsonb | {} | Extensible metadata |

**info structure:**
```json
{
  "poll_template_positions": {
    "practice_proposal": 0,
    "check": 1,
    "advice": 2,
    ...
  },
  "hidden_poll_templates": [],
  "categorize_poll_templates": true
}
```

### Other Fields

| Column | Type | Description |
|--------|------|-------------|
| `category` | string | Group category |
| `category_id` | integer | Legacy category FK |
| `theme_id` | integer | Legacy theme FK |
| `cohort_id` | integer | Analytics cohort |
| `default_group_cover_id` | integer | Default cover FK |
| `admin_tags` | string | Internal admin tags |
| `content_locale` | string | Content locale |
| `is_referral` | boolean | Referral tracking |
| `request_to_join_prompt` | string | Join request prompt (max 280) |

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `name` | presence required | always |
| `name` | max length 250 | always |
| `name`, `description` | no spam regex | NoSpam concern |
| `description` | max length via AppConfig | always |
| `request_to_join_prompt` | max length 280 | always |
| `handle` | uniqueness | allow nil |
| `subscription` | must be absent | if subgroup |
| `discussion_privacy_options` | inclusion in DISCUSSION_PRIVACY_OPTIONS | always |
| `membership_granted_upon` | inclusion in MEMBERSHIP_GRANTED_UPON_OPTIONS | always |

### Custom Validations

```ruby
validate :limit_inheritance
# Subgroups cannot have subgroups (only one level of nesting)

validate :handle_is_valid
# Handle must be parameterized and start with parent handle for subgroups

validate :validate_parent_members_can_see_discussions
# Consistency check for parent visibility settings

validate :validate_is_visible_to_parent_members
# Consistency check for visibility settings

validate :validate_discussion_privacy_options
# Discussions must be public if group is open
# Discussions must be private if group is hidden

validate :validate_trial_group_cannot_be_public
# Trial subscription groups cannot be public
```

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `creator` | User | - | User who created the group |
| `parent` | Group | - | Parent group for subgroups |
| `subscription` | Subscription | - | Billing subscription |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `discussions` | Discussion | dependent: :destroy | Group discussions |
| `discussion_templates` | DiscussionTemplate | dependent: :destroy | Discussion templates |
| `public_discussions` | Discussion | `-> { visible_to_public }` | Public discussions |
| `comments` | Comment | through: :discussions | All comments |
| `polls` | Poll | dependent: :destroy | Group polls |
| `poll_templates` | PollTemplate | dependent: :destroy | Poll templates |
| `all_memberships` | Membership | dependent: :destroy | All memberships |
| `all_members` | User | through: :all_memberships | All members |
| `memberships` | Membership | `-> { active }` | Active memberships |
| `members` | User | through: :memberships | Active members |
| `delegate_memberships` | Membership | `-> { active.delegates }` | Delegate memberships |
| `delegates` | User | through: :delegate_memberships | Delegate users |
| `accepted_memberships` | Membership | `-> { active.accepted }` | Accepted memberships |
| `accepted_members` | User | through: :accepted_memberships | Accepted members |
| `admin_memberships` | Membership | `-> { active.where(admin: true) }` | Admin memberships |
| `admins` | User | through: :admin_memberships | Admin users |
| `membership_requests` | MembershipRequest | dependent: :destroy | Join requests |
| `pending_membership_requests` | MembershipRequest | `-> { where(response: nil) }` | Pending requests |
| `subgroups` | Group | foreign_key: :parent_id, `-> { where(archived_at: nil) }` | Active subgroups |
| `all_subgroups` | Group | foreign_key: :parent_id, dependent: :destroy | All subgroups |
| `documents` | Document | as: :model, dependent: :destroy | Attached documents |
| `chatbots` | Chatbot | dependent: :destroy | Webhook/bot configs |
| `tags` | Tag | foreign_key: :group_id | Group tags |

### Document Associations (through)

| Association | Through | Description |
|-------------|---------|-------------|
| `discussion_documents` | discussions | Documents on discussions |
| `poll_documents` | polls | Documents on polls |
| `comment_documents` | comments | Documents on comments |

---

## Scopes

```ruby
scope :archived, -> { where('archived_at IS NOT NULL') }
scope :published, -> { where(archived_at: nil) }
scope :parents_only, -> { where(parent_id: nil) }
scope :visible_to_public, -> { published.where(is_visible_to_public: true) }
scope :hidden_from_public, -> { published.where(is_visible_to_public: false) }
scope :with_serializer_includes, -> { includes(:subscription) }

scope :in_organisation, ->(group) { where(id: group.id_and_subgroup_ids) }

scope :search_for, ->(query) { where("name ilike :q", q: "%#{query}%") }
scope :explore_search, ->(query) {
  where("name ilike :q or description ilike :q", q: "%#{query}%")
}
scope :mention_search, lambda { |q|
  where("groups.name ilike :first OR groups.name ilike :other OR groups.handle ilike :first",
        first: "#{q}%", other: "% #{q}%")
}

scope :dangling, -> {
  joins('left join groups parents on parents.id = groups.parent_id')
    .where('groups.parent_id is not null and parents.id is null')
}

scope :empty_no_subscription, -> {
  joins('left join subscriptions on subscription_id = groups.subscription_id')
    .where('subscriptions.id is null and groups.parent_id is null')
    .where('memberships_count < 2 AND discussions_count < 3 and polls_count < 2 and subgroups_count = 0')
    .where('groups.created_at < ?', 1.year.ago)
}

scope :expired_trial, -> {
  joins(:subscription)
    .where('subscriptions.plan = ?', 'trial')
    .where('subscriptions.expires_at < ?', 12.months.ago)
}

scope :any_trial, -> { joins(:subscription).where('subscriptions.plan = ?', 'trial') }

scope :expired_demo, -> {
  joins(:subscription)
    .where('subscriptions.plan = ?', 'demo')
    .where('groups.created_at < ?', 7.days.ago)
}

scope :not_demo, -> { joins(:subscription).where('subscriptions.plan != ?', 'demo') }

scope :by_slack_team, ->(team_id) {
  joins(:identities)
    .where("(omniauth_identities.custom_fields->'slack_team_id')::jsonb ? :team_id", team_id: team_id)
}
```

---

## Callbacks

### Before Validation
- `ensure_handle_is_not_empty` - Sets handle to nil if blank

### After Initialize (GroupPrivacy concern)
- `set_privacy_defaults` - Sets default privacy settings

### Before Validation (GroupPrivacy concern)
- `set_discussions_private_only` - If group is hidden from public

---

## Instance Methods

### Hierarchy Methods

```ruby
def is_parent?
  parent_id.blank?
end

def is_subgroup?
  !is_parent?
end

def parent_or_self
  parent || self
end

def self_and_subgroups
  Group.where(id: [id].concat(subgroup_ids))
end

def id_and_subgroup_ids
  subgroup_ids.concat([id]).compact.uniq
end

def is_subgroup_of_hidden_parent?
  is_subgroup? && parent.is_hidden_from_public?
end
```

### Privacy Methods (GroupPrivacy concern)

```ruby
def group_privacy
  if is_visible_to_public?
    public_discussions_only? ? 'open' : 'closed'
  elsif parent_id && is_visible_to_parent_members?
    'closed'
  else
    'secret'
  end
end

def group_privacy=(term)
  # Sets multiple privacy flags based on term: 'open', 'closed', 'secret'
  case term
  when 'open'
    self.is_visible_to_public = true
    self.discussion_privacy_options = 'public_only'
    # ...
  when 'closed'
    self.is_visible_to_public = true
    # ...
  when 'secret'
    self.is_visible_to_public = false
    self.listed_in_explore = false
    self.discussion_privacy_options = 'private_only'
    self.membership_granted_upon = 'invitation'
    self.is_visible_to_parent_members = false
  end
end

def is_hidden_from_public?
  !is_visible_to_public?
end

def private_discussions_only?
  discussion_privacy_options == 'private_only'
end

def public_discussions_only?
  discussion_privacy_options == 'public_only'
end

def public_or_private_discussions_allowed?
  discussion_privacy_options == 'public_or_private'
end

def discussion_private_default
  discussion_privacy_options != "public_only"
end

def membership_granted_upon_approval?
  membership_granted_upon == 'approval'
end

def membership_granted_upon_request?
  membership_granted_upon == 'request'
end

def membership_granted_upon_invitation?
  membership_granted_upon == 'invitation'
end
```

### Membership Methods

```ruby
def add_member!(user, inviter: nil)
  # Adds user as member, reactivates if previously revoked
  # Triggers GenericWorker for poll notification
  save! unless persisted?
  user.save! unless user.persisted?

  if membership = Membership.find_by(user_id: user.id, group_id: id)
    if membership.revoked_at
      membership.update(admin: false, revoked_at: nil, revoker_id: nil,
                       accepted_at: DateTime.now, inviter: inviter)
    end
  else
    membership = Membership.create!(user_id: user.id, group_id: id,
                                    inviter: inviter, accepted_at: DateTime.now)
  end

  GenericWorker.perform_async('PollService', 'group_members_added', self.id)
  membership
rescue ActiveRecord::RecordNotUnique
  retry
end

def add_members!(users, inviter: nil)
  users.map { |user| add_member!(user, inviter: inviter) }
end

def add_admin!(user)
  add_member!(user).tap do |m|
    m.make_admin!
    update(creator: user) if creator.blank?
  end.reload
end

def membership_for(user)
  memberships.find_by(user_id: user.id)
end

def existing_member_ids
  member_ids
end
```

### Archive Methods

```ruby
def archive!
  Group.where(id: id_and_subgroup_ids).update_all(archived_at: DateTime.now)
  reload
end

def unarchive!
  Group.where(id: id_and_subgroup_ids).update_all(archived_at: nil)
  reload
end
```

### Count Methods

```ruby
def org_members_count
  Membership.active.where(group_id: id_and_subgroup_ids).count('distinct user_id')
end

def org_accepted_members_count
  Membership.active.accepted.where(group_id: id_and_subgroup_ids).count('distinct user_id')
end

def org_discussions_count
  Group.where(id: id_and_subgroup_ids).sum(:discussions_count)
end

def org_polls_count
  Group.where(id: id_and_subgroup_ids).sum(:polls_count)
end

def accepted_memberships_count
  memberships_count - pending_memberships_count
end
```

### URL Methods

```ruby
def full_name
  if is_subgroup?
    [parent&.name, name].compact.join(' - ')
  else
    name
  end
end

def title
  full_name
end

def message_channel
  "/group-#{self.key}"
end
```

### Media Methods

```ruby
def logo_url(size = 512)
  return nil unless logo.attached?
  # Returns ActiveStorage representation URL
end

def cover_url(size = 512)
  return nil unless cover_photo.attached?
  # Returns ActiveStorage representation URL
end

def self_or_parent_logo_url(size = 512)
  logo_url(size) || (parent && parent.logo_url(size))
end

def self_or_parent_cover_url(size = 512)
  cover_url(size) || (parent && parent.cover_url(size))
end

def custom_cover_photo?
  !GroupService::DEFAULT_COVER_PHOTO_FILENAMES.include? cover_photo.filename.to_s
end
```

### Template Methods

```ruby
def poll_template_positions
  self[:info]['poll_template_positions'] ||= {
    'practice_proposal' => 0,
    'check' => 1,
    'advice' => 2,
    'consent' => 3,
    'consensus' => 4,
    'poll' => 5,
    'score' => 6,
    'dot_vote' => 7,
    'ranked_choice' => 8,
    'meeting' => 9,
  }
end

def hidden_poll_templates
  self[:info]['hidden_poll_templates'] ||= AppConfig.app_features.fetch(:hidden_poll_templates, [])
end

def categorize_poll_templates
  self[:info].fetch('categorize_poll_templates', true)
end
```

### Subscription Methods

```ruby
def is_trial_or_demo?
  parent_group = parent_or_self
  subscription = Subscription.for(parent_group)
  ['trial', 'demo'].include?(subscription.plan)
end
```

### Delegate Methods

```ruby
delegate :locale, to: :creator, allow_nil: true
delegate :time_zone, to: :creator, allow_nil: true
delegate :date_time_pref, to: :creator, allow_nil: true
delegate :include?, to: :users, prefix: true
delegate :members, to: :parent, prefix: true
```

---

## Counter Cache Definitions

```ruby
define_counter_cache(:polls_count) { |g| g.polls.count }
define_counter_cache(:closed_polls_count) { |g| g.polls.closed.count }
define_counter_cache(:poll_templates_count) { |g| g.poll_templates.kept.count }
define_counter_cache(:memberships_count) { |g| g.memberships.count }
define_counter_cache(:pending_memberships_count) { |g| g.memberships.pending.count }
define_counter_cache(:admin_memberships_count) { |g| g.admin_memberships.count }
define_counter_cache(:delegates_count) { |g| g.memberships.delegates.count }
define_counter_cache(:public_discussions_count) { |g| g.discussions.visible_to_public.count }
define_counter_cache(:discussions_count) { |g| g.discussions.kept.count }
define_counter_cache(:open_discussions_count) { |g| g.discussions.is_open.count }
define_counter_cache(:closed_discussions_count) { |g| g.discussions.is_closed.count }
define_counter_cache(:discussion_templates_count) { |g| g.discussion_templates.kept.count }
define_counter_cache(:subgroups_count) { |g| g.subgroups.published.count }

update_counter_cache(:parent, :subgroups_count)
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `HasRichText` | Rich text description with sanitization |
| `CustomCounterCache::Model` | Counter cache definitions |
| `ReadableUnguessableUrls` | 8-char key generation |
| `SelfReferencing` | `group` and `group_id` methods |
| `MessageChannel` | Real-time pub/sub channel |
| `GroupPrivacy` | Privacy settings and validation |
| `HasEvents` | Event associations |
| `Translatable` | Translation support |
| `HasTokens` | Token initialization |
| `NoSpam` | Spam validation |
| `GroupExportRelations` | Export query optimizations |

---

## Paper Trail Tracking

Tracked fields:
- `name`
- `parent_id`
- `description`
- `description_format`
- `handle`
- `archived_at`
- `parent_members_can_see_discussions`
- `key`
- `is_visible_to_public`
- `is_visible_to_parent_members`
- `discussion_privacy_options`
- `members_can_add_members`
- `membership_granted_upon`
- `members_can_edit_discussions`
- `members_can_edit_comments`
- `members_can_delete_comments`
- `members_can_raise_motions`
- `members_can_start_discussions`
- `members_can_create_subgroups`
- `creator_id`
- `subscription_id`
- `members_can_announce`
- `new_threads_max_depth`
- `new_threads_newest_first`
- `admins_can_edit_user_content`
- `listed_in_explore`
- `attachments`

**Note:** `members_can_add_guests` is NOT tracked (documented gotcha).

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `handle` | UNIQUE | |
| `key` | UNIQUE | |
| `token` | UNIQUE | |
| `parent_id` | INDEX | |
| `subscription_id` | INDEX | |
| `name` | INDEX | |
| `full_name` | INDEX | |
| `created_at` | INDEX | |
| `archived_at` | PARTIAL | where IS NULL |

---

## Related Classes

### NullGroup

Null object pattern for discussions without a group.

```ruby
class NullGroup
  # Returns safe defaults for all group methods
  # true_methods and false_methods define behavior
  # NOTE: true_methods overrides false_methods due to method definition order (gotcha)
end
```

### FormalGroup / GuestGroup

Legacy STI subclasses that appear unused in current code.

---

## Uncertainties

1. **subgroup_ids method** - Referenced but definition not visible in model (likely from GroupExportRelations)
2. **GroupService::DEFAULT_COVER_PHOTO_FILENAMES** - Default cover photos list location unclear
3. **admin_email method** - Assumes at least one admin exists, may fail on empty group
4. **Legacy Paperclip fields** - Cover photo and logo fields exist for both Paperclip and ActiveStorage

**Confidence Level:** HIGH for core attributes, MEDIUM for some privacy validation edge cases.
