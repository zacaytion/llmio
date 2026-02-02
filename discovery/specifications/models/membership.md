# Membership Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/membership.rb`
- `/app/models/membership_request.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Membership model represents a user's membership in a group, including their role (admin, delegate), notification settings (volume), and invitation state. Memberships can be pending (invited but not accepted) or revoked.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `group_id` | integer | - | YES | FK to groups | Group membership is for |
| `user_id` | integer | - | YES | FK to users | User who is a member |
| `token` | string | auto-generated | YES | UNIQUE | Invitation/membership token |

### Role & Permissions

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `admin` | boolean | false | NO | - | Admin role |
| `delegate` | boolean | false | NO | - | Delegate role |

### Invitation Tracking

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `inviter_id` | integer | - | YES | FK to users | User who invited |
| `revoker_id` | integer | - | YES | FK to users | User who revoked |
| `invitation_id` | integer | - | YES | FK (legacy) | Legacy invitation record |
| `accepted_at` | datetime | - | YES | - | When invitation was accepted |
| `revoked_at` | datetime | - | YES | - | When membership was revoked |

### Notification Settings

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `volume` | integer | - | YES | enum 0-3 | Notification volume level |

**Volume Enum Values (HasVolume concern):**
- 0: mute - No notifications
- 1: quiet - In-app only
- 2: normal - In-app + email
- 3: loud - All notifications

### User Experience

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `inbox_position` | integer | 0 | YES | - | Dashboard ordering |
| `title` | string | - | YES | - | Custom member title |
| `experiences` | jsonb | {} | NO | - | Feature tutorials seen |

### Session Management

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `saml_session_expires_at` | datetime | - | YES | - | SAML session timeout |

### Timestamps

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `created_at` | datetime | - | YES | When membership was created |
| `updated_at` | datetime | - | YES | Last update |

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `group` | presence required | always |
| `user` | presence required | always |
| `user_id` | uniqueness scoped to `group_id` | always |

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `group` | Group | - | The group |
| `user` | User | - | The member |
| `inviter` | User | class_name: 'User' | User who sent invitation |
| `revoker` | User | class_name: 'User' | User who revoked membership |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `events` | Event | as: :eventable, dependent: :destroy | Membership events |

---

## Scopes

```ruby
scope :active, -> { where(revoked_at: nil) }
scope :pending, -> { active.where(accepted_at: nil) }
scope :accepted, -> { where('accepted_at IS NOT NULL') }
scope :revoked, -> { where('revoked_at IS NOT NULL') }
scope :delegates, -> { where(delegate: true) }
scope :admin, -> { where(admin: true) }
scope :email_verified, -> { joins(:user).where("users.email_verified": true) }

scope :for_group, ->(group) { where(group_id: group) }

scope :in_organisation, ->(group) {
  includes(:user).where(group_id: group.id_and_subgroup_ids).active
}

scope :search_for, ->(query) {
  joins(:user).where(
    "users.name ilike :query or users.username ilike :query or users.email ilike :query",
    query: "%#{query}%"
  )
}

scope :dangling, -> {
  joins('left join groups g on memberships.group_id = g.id')
    .where('group_id is not null and g.id is null')
}

# From HasVolume concern
scope :volume, ->(volume) { where(volume: volumes[volume]) }
scope :volume_at_least, ->(volume) { where('volume >= ?', volumes[volume]) }
scope :email_notifications, -> { where('volume >= ?', volumes[:normal]) }
scope :app_notifications, -> { where('volume >= ?', volumes[:quiet]) }

# From HasTimeframe concern
scope :within, ->(since, till, field = nil) {
  where("#{table_name}.#{field || :created_at} BETWEEN ? AND ?",
        since || 100.years.ago, till || 100.years.from_now)
}
scope :until, ->(till) { within(nil, till) }
scope :since, ->(since) { within(since, nil) }
```

---

## Callbacks

### Before Create
- `set_volume` - Sets volume from user's `default_membership_volume` if nil

---

## Instance Methods

### Role Management

```ruby
def make_admin!
  update_attribute(:admin, true)
end

def remove_admin!
  update_attribute(:admin, false)
end
```

### Related Records

```ruby
def discussion_readers
  DiscussionReader
    .joins(:discussion)
    .where("discussions.group_id": group_id)
    .where("discussion_readers.user_id": user_id)
end

def stances
  Stance
    .joins(:poll)
    .where("polls.group_id": group_id)
    .where(participant_id: user_id)
end
```

### Volume Methods (HasVolume concern)

```ruby
def set_volume!(volume, persist: true)
  if self.class.volumes.include?(volume)
    self.volume = volume
    save if persist
  else
    errors.add :volume, I18n.t(:"activerecord.errors.messages.invalid")
    false
  end
end

def volume_is_normal_or_loud?
  volume_is_normal? || volume_is_loud?
end

def volume_is_loud?
  volume.to_s == 'loud'
end

def volume_is_normal?
  volume.to_s == 'normal'
end

def volume_is_quiet?
  volume.to_s == 'quiet'
end

def volume_is_mute?
  volume.to_s == 'mute'
end
```

### Identity Methods

```ruby
def title_model
  group
end

def author_id
  inviter_id
end

def author
  inviter
end

def message_channel
  "membership-#{token}"
end
```

### Delegate Methods

```ruby
delegate :name, :email, to: :user, prefix: :user, allow_nil: true
delegate :parent, to: :group, prefix: :group, allow_nil: true
delegate :name, :full_name, to: :group, prefix: :group
delegate :admins, to: :group, prefix: :group
delegate :name, to: :inviter, prefix: :inviter, allow_nil: true
delegate :mailer, to: :user
```

---

## Counter Cache Updates

```ruby
update_counter_cache :group, :memberships_count
update_counter_cache :group, :delegates_count
update_counter_cache :group, :pending_memberships_count
update_counter_cache :group, :admin_memberships_count
update_counter_cache :user, :memberships_count
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `CustomCounterCache::Model` | Counter cache updates |
| `HasVolume` | Notification volume enum and methods |
| `HasTimeframe` | Time-based scopes |
| `HasExperiences` | Feature flag management |
| `FriendlyId` | Token-based finding |
| `HasTokens` | Token initialization |

---

## Paper Trail Tracking

Tracked fields:
- `group_id`
- `user_id`
- `inviter_id`
- `admin`
- `title`
- `revoked_at`
- `revoker_id`
- `volume`
- `accepted_at`

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `(group_id, user_id)` | UNIQUE | Ensures one membership per user per group |
| `token` | UNIQUE | |
| `inviter_id` | INDEX | |
| `(user_id, volume)` | INDEX | |
| `volume` | INDEX | |
| `created_at` | INDEX | |

---

## Custom Exception

```ruby
class Membership::InvitationAlreadyUsed < StandardError
  attr_accessor :membership
  def initialize(obj)
    self.membership = obj
  end
end
```

Used when attempting to accept an invitation that's already been used.

---

## Related Model: MembershipRequest

Pending requests to join groups.

### Attributes

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `group_id` | integer | - | YES | FK to groups |
| `requestor_id` | integer | - | YES | FK to users (requester) |
| `responder_id` | integer | - | YES | FK to users (admin) |
| `name` | string(255) | - | YES | Requester name |
| `email` | string(255) | - | YES | Requester email |
| `introduction` | text | - | YES | Request message |
| `response` | string(255) | - | YES | "approved" or "ignored" |
| `responded_at` | datetime | - | YES | Response timestamp |
| `created_at` | datetime | - | NO | |
| `updated_at` | datetime | - | NO | |

---

## State Diagram

```
                    ┌──────────────────────────────────────┐
                    │                                      │
                    ▼                                      │
┌─────────┐   accept   ┌─────────┐   revoke   ┌─────────┐ │
│ PENDING │ ─────────> │ ACTIVE  │ ─────────> │ REVOKED │ │
└─────────┘            └─────────┘            └─────────┘ │
     │                      │                      │       │
     │                      │                      │       │
     │    revoke            │    re-add            │       │
     └──────────────────────┼──────────────────────┘       │
                            │                              │
                            └──────────────────────────────┘
```

**States:**
- **PENDING**: `accepted_at IS NULL AND revoked_at IS NULL` - Invited but not accepted
- **ACTIVE**: `accepted_at IS NOT NULL AND revoked_at IS NULL` - Full member
- **REVOKED**: `revoked_at IS NOT NULL` - Membership removed

---

## Uncertainties

1. **invitation_id field** - Legacy field referencing old invitations system, appears unused
2. **saml_session_expires_at** - SAML session management mechanism unclear
3. **inbox_position** - Dashboard ordering implementation unclear

**Confidence Level:** HIGH for attributes and methods, MEDIUM for some state transitions.
