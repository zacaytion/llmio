# Authorization - Confirmed Architecture

## Summary

Both Discovery and Research documentation **agree** on the authorization architecture. Loomio uses CanCanCan for resource-based permissions with a modular ability system.

## Key Details

### Authorization Framework

**CanCanCan** (formerly CanCan) provides:
- Resource-based authorization (`can :action, Resource`)
- Ability class composition via prepend/include
- Integration with controllers via `authorize!` and `load_and_authorize_resource`

### Ability Module Architecture

Both sources confirm **25 ability modules** composed via prepend:

```ruby
# app/models/ability/base.rb
class Ability::Base
  include CanCan::Ability

  prepend Ability::User
  prepend Ability::Group
  prepend Ability::Membership
  prepend Ability::Discussion
  prepend Ability::Comment
  prepend Ability::Poll
  prepend Ability::Stance
  prepend Ability::Outcome
  prepend Ability::Event
  prepend Ability::Notification
  prepend Ability::Reaction
  prepend Ability::Document
  prepend Ability::Attachment
  prepend Ability::Webhook
  prepend Ability::Chatbot
  # ... and more
end
```

### Authorization Flow

Both sources confirm the service-layer authorization pattern:

```ruby
# Services call authorize! before mutations
class DiscussionService
  def self.create(discussion:, actor:, params:)
    actor.ability.authorize!(:create, discussion)
    discussion.assign_attributes(params)
    discussion.save!
    EventBus.publish('discussion_create', discussion, actor)
  end
end
```

### Membership Roles (4 Confirmed)

| Role | Value | Permissions |
|------|-------|-------------|
| Guest | N/A | Read access, vote if invited to poll |
| Member | 0 | Based on group permission flags |
| Delegate | N/A | Extended regional/proxy permissions |
| Admin | 1 | Full group management |

### Group Permission Flags

Both sources confirm 9 core `members_can_*` flags:

| Flag | Default | Controls |
|------|---------|----------|
| `members_can_add_members` | true | Invite new members |
| `members_can_add_guests` | true | Invite guests to threads |
| `members_can_announce` | true | Send announcements |
| `members_can_create_subgroups` | false | Create child groups |
| `members_can_start_discussions` | true | Create new threads |
| `members_can_edit_discussions` | true | Edit thread content |
| `members_can_edit_comments` | true | Edit own comments |
| `members_can_delete_comments` | false | Delete own comments |
| `members_can_raise_motions` | true | Create polls/proposals |

Note: `admins_can_edit_user_content` and thread configuration (`new_threads_*`) need verification - see `follow_up/permission_flags.md`.

### Discussion Visibility

Both sources confirm visibility model:

| Visibility | Description |
|------------|-------------|
| Public | `private: false` - Anyone can view |
| Private | `private: true` - Only group members |
| Guest Access | Via `discussion_readers` with `guest: true` and token |

### Poll Access Control

Both sources confirm poll visibility:

| Scenario | Who Can Vote |
|----------|--------------|
| Group poll | All group members |
| Discussion poll | Discussion participants only |
| Announced poll | Specific users/groups via announcement |
| Anonymous poll | Votes recorded, but `participant_id` scrubbed on close |

### Guest Access Mechanism

Both sources confirm token-based guest access:

- `Membership.token` - Invitation redemption
- `DiscussionReader.token` - Guest thread access
- `Stance.token` - Voting link for non-members

## Source Alignment

| Aspect | Discovery | Research | Status |
|--------|-----------|----------|--------|
| Framework | CanCanCan | CanCanCan | ✅ Confirmed |
| Ability modules | 25 | 25 | ✅ Confirmed |
| Authorization location | Services | Services | ✅ Confirmed |
| Membership roles | 4 | 4 | ✅ Confirmed |
| Core permission flags | 9-10 | 9-11 | ⚠️ See follow_up |
| Discussion visibility | 3 modes | 3 modes | ✅ Confirmed |
| Guest token access | Documented | Documented | ✅ Confirmed |

## Implementation Notes

### Go Authorization Approach

Options for Go implementation:

**Option 1: Casbin (recommended for complex RBAC)**
```go
e, _ := casbin.NewEnforcer("model.conf", "policy.csv")
e.Enforce(user, resource, action)
```

**Option 2: Custom Domain-Specific Rules**
```go
type Ability struct {
    user *User
}

func (a *Ability) Can(action string, resource interface{}) bool {
    switch r := resource.(type) {
    case *Discussion:
        return a.canDiscussion(action, r)
    case *Poll:
        return a.canPoll(action, r)
    // ...
    }
    return false
}

func (a *Ability) canDiscussion(action string, d *Discussion) bool {
    membership := a.membershipFor(d.GroupID)
    if membership == nil {
        return d.Private == false && action == "read"
    }

    switch action {
    case "create":
        return d.Group.MembersCanStartDiscussions || membership.Admin
    case "update":
        return d.AuthorID == a.user.ID ||
               (d.Group.MembersCanEditDiscussions && membership != nil) ||
               membership.Admin
    // ...
    }
    return false
}
```

### Key Authorization Checks

For Go implementation, ensure these checks exist:

1. **Discussion create**: `members_can_start_discussions` OR admin
2. **Discussion update**: Author OR `members_can_edit_discussions` OR admin
3. **Poll create**: `members_can_raise_motions` OR admin
4. **Stance create**: Poll participant (via group, discussion, or announcement)
5. **Membership create**: `members_can_add_members` OR admin
6. **Guest invite**: `members_can_add_guests` OR admin

### Hierarchical Group Permissions

Subgroup permissions inherit from parent unless overridden:

```go
func (g *Group) EffectivePermission(flag string) bool {
    if g.hasExplicitSetting(flag) {
        return g.getSetting(flag)
    }
    if g.ParentID != nil {
        return g.Parent.EffectivePermission(flag)
    }
    return defaultPermissions[flag]
}
```
