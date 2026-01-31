# Authorization

> CanCan abilities and permission model.

## Overview

Loomio uses **CanCan** (via `cancancan` gem) for authorization. Each ability module defines permissions for a specific resource type.

## Ability Modules (25)

**Source:** `orig/loomio/app/models/ability/`

| Module | Resource | Key Actions |
|--------|----------|-------------|
| base.rb | Core | setup, helpers |
| ahoy_message.rb | Analytics | view |
| attachment.rb | Attachments | create, destroy |
| chatbot.rb | Chatbots | CRUD |
| comment.rb | Comments | CRUD, discard |
| contact_message.rb | Support | create |
| discussion.rb | Discussions | CRUD, move, announce, discard |
| discussion_reader.rb | Read state | update |
| discussion_template.rb | Templates | CRUD |
| document.rb | Documents | CRUD |
| event.rb | Events | update, destroy |
| forward_email_rule.rb | Email rules | CRUD |
| group.rb | Groups | CRUD, export, archive |
| group_identity.rb | SSO configs | CRUD |
| login_token.rb | Magic links | create |
| member_email_alias.rb | Email aliases | CRUD |
| membership.rb | Memberships | manage, invite, remove |
| membership_request.rb | Join requests | approve, ignore |
| notification.rb | Notifications | view |
| outcome.rb | Outcomes | CRUD |
| poll.rb | Polls | CRUD, close, remind |
| poll_template.rb | Templates | CRUD |
| reaction.rb | Reactions | CRUD |
| stance.rb | Stances | create, update |
| user.rb | Users | update, deactivate |
| webhook.rb | Webhooks | CRUD |

## Permission Pattern

```ruby
class Ability::Discussion
  def initialize(user)
    super(user)

    can :show, Discussion do |discussion|
      discussion.readers.include?(user) ||
        discussion.group.members.include?(user) ||
        discussion.public?
    end

    can :create, Discussion do |discussion|
      discussion.group.members_can_start_discussions? ||
        discussion.group.admins.include?(user)
    end
  end
end
```

## Group Permission Flags

Boolean columns on `groups` table:

| Flag | Default | Purpose |
|------|---------|---------|
| members_can_add_members | true | Invite others |
| members_can_add_guests | false | Add guest users |
| members_can_announce | true | Send announcements |
| members_can_create_subgroups | false | Create subgroups |
| members_can_start_discussions | true | Start threads |
| members_can_edit_discussions | true | Edit any thread |
| members_can_edit_comments | true | Edit any comment |
| members_can_delete_comments | false | Delete any comment |
| members_can_raise_motions | true | Create polls |
| new_threads_max_depth | 2 | Reply nesting depth |
| new_threads_newest_first | true | Comment ordering |

## Membership Roles

| Role | Level | Permissions |
|------|-------|-------------|
| guest | - | Read access, vote if invited |
| member | 0 | Based on group flags |
| delegate | - | Extended permissions (region-specific) |
| admin | 1 | Full group management |

**Special:** `participation_token` grants guests temporary discussion/poll access.

## Discussion Visibility

1. **Public:** Anyone can view
2. **Private:** Only group members
3. **Guest access:** Via `discussion_readers` with `guest = true`

**Revocation:** `discussion_readers.revoked_at` or `memberships.revoked_at`

## Poll Access Control

1. **Group poll:** All group members can vote
2. **Discussion poll:** Discussion participants
3. **Announced:** Specific users/groups via announcement
4. **Anonymous:** `stances.participant_id` hidden, `voter_scores` cleared

## Go Implementation Notes

Consider using:
- **Casbin:** Flexible RBAC/ABAC with PostgreSQL adapter
- **Custom:** Domain-specific rules may be simpler

Key patterns to preserve:
1. Resource-based permissions (can user X do Y on resource Z?)
2. Hierarchical groups (parent permissions cascade)
3. Guest access via tokens
4. Revocation tracking (not hard delete)

---
