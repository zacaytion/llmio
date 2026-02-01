# Templates Domain: Controllers

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

Template controllers provide RESTful APIs for managing discussion and poll templates. They extend RestfulController for standard CRUD plus custom actions for hiding, discarding, and ordering.

---

## DiscussionTemplatesController

**Location:** `/app/controllers/api/v1/discussion_templates_controller.rb`

**Routes:** `/api/v1/discussion_templates`

### Standard Actions

**index** - List templates for a group

```
PSEUDO-CODE:
GET /api/v1/discussion_templates?group_id=X

group = current_user.groups.find(group_id) or NullGroup
if group has no kept templates:
  initialize group with DiscussionTemplateService.initial_templates(category, parent_id)

if id parameter present:
  return single template if in user's group
else:
  return all group discussion_templates
```

**show** - Get single template

```
PSEUDO-CODE:
GET /api/v1/discussion_templates/:id

find template where:
  group_id in current_user's groups OR public = true
return serialized template
```

**create** - Create custom template

Delegates to DiscussionTemplateService.create via RestfulController base

**update** - Update existing template

Delegates to DiscussionTemplateService.update

**destroy** - Permanently delete template

```
PSEUDO-CODE:
DELETE /api/v1/discussion_templates/:id

find template by id
verify current_user is admin of template's group
permanently destroy template
```

### Custom Collection Actions

**browse_tags** - Get popular tags from public templates

```
PSEUDO-CODE:
GET /api/v1/discussion_templates/browse_tags

count occurrences of each tag across public templates
sort by frequency
return top 20 tags
```

**browse** - Search public template gallery

```
PSEUDO-CODE:
GET /api/v1/discussion_templates/browse?query=X

if no public templates exist:
  call DiscussionTemplateService.create_public_templates

filter templates where:
  public = true
  AND (in 'templates' group OR subscription.plan != 'trial')
  AND (query matches process_name/subtitle OR in tags)

return up to 50 results with author/group metadata
```

**positions** - Update template ordering

```
PSEUDO-CODE:
POST /api/v1/discussion_templates/positions
  group_id: X
  ids: [id1, key1, id2, ...]

group = current_user.adminable_groups.find(group_id)

for each id at index in ids:
  if id is integer (custom template):
    update DiscussionTemplate position = index
  else (system template key):
    set group.discussion_template_positions[key] = index

save group
return updated index
```

**discard** - Soft-delete custom template

```
PSEUDO-CODE:
POST /api/v1/discussion_templates/discard
  group_id: X, id: Y

group = current_user.adminable_groups.find(group_id)
template = group.discussion_templates.kept.find(id)
template.discard!
return updated index
```

**undiscard** - Restore soft-deleted template

```
PSEUDO-CODE:
POST /api/v1/discussion_templates/undiscard
  group_id: X, id: Y

template = group.discussion_templates.discarded.find(id)
template.undiscard!
return updated index
```

**hide** - Hide system template

```
PSEUDO-CODE:
POST /api/v1/discussion_templates/hide
  group_id: X, key: Y

verify template with key exists in group's templates
add key to group.hidden_discussion_templates
return updated index
```

**unhide** - Show hidden system template

```
PSEUDO-CODE:
POST /api/v1/discussion_templates/unhide
  group_id: X, key: Y

verify template with key exists
remove key from group.hidden_discussion_templates
return updated index
```

---

## PollTemplatesController

**Location:** `/app/controllers/api/v1/poll_templates_controller.rb`

**Routes:** `/api/v1/poll_templates`

### Standard Actions

**index** - List templates for a group

```
PSEUDO-CODE:
GET /api/v1/poll_templates?group_id=X[&key_or_id=Y]

group = current_user.groups.find(group_id) or NullGroup

if key_or_id looks like integer:
  return single template by id
else:
  return PollTemplateService.group_templates(group)
```

Key difference from discussion templates: poll templates always use the service method which merges custom + system templates.

**show** - Get single template

```
PSEUDO-CODE:
GET /api/v1/poll_templates/:id

find template in current_user's group_ids
```

**update** - Update template (handles system template override)

```
PSEUDO-CODE:
PATCH /api/v1/poll_templates/:id

if id is not numeric (system template key):
  find by group_id and key
else:
  find by id

delegate to PollTemplateService.update
```

**destroy** - Permanently delete

```
PSEUDO-CODE:
DELETE /api/v1/poll_templates/:id

verify admin access
destroy template
```

### Custom Collection Actions

**positions** - Update ordering

```
PSEUDO-CODE:
POST /api/v1/poll_templates/positions
  group_id: X, ids: [...]

for each id at index:
  if integer: update PollTemplate.position
  else: update group.poll_template_positions[key]

save group
```

**settings** - Update template display settings

```
PSEUDO-CODE:
POST /api/v1/poll_templates/settings
  group_id: X, categorize_poll_templates: true/false

group.categorize_poll_templates = value
save and broadcast update via MessageChannelService
```

**discard/undiscard** - Same pattern as discussion templates

**hide/unhide** - Same pattern as discussion templates

---

## Route Configuration

**Location:** `/config/routes.rb`

### API Routes

```
resources :discussion_templates, only: [:create, :index, :show, :update, :destroy] do
  collection do
    get :browse_tags
    get :browse
    post :hide
    post :unhide
    post :discard
    post :undiscard
    post :positions
  end
end

resources :poll_templates, only: [:index, :create, :update, :show, :destroy] do
  collection do
    post :hide
    post :unhide
    post :discard
    post :undiscard
    post :positions
    post :settings
  end
end
```

### Frontend Routes

```
get 'poll_templates/new'         => 'application#index'
get 'poll_templates/:id'         => 'application#index'
get 'poll_templates/:id/edit'    => 'application#index'
get 'thread_templates/browse'    => 'application#index'
get 'thread_templates/new'       => 'application#index'
```

---

## Authorization Patterns

All mutating actions require group admin access:

```
PSEUDO-CODE:
// For custom templates
group = current_user.adminable_groups.find(group_id)

// For standard CRUD via services
actor.ability.authorize!(:create/:update, template)
  -> checks template.group.admins.exists?(user.id)
```

Read access is more permissive:
- Own group templates: always visible
- Public templates: visible to anyone

---

## Response Patterns

Most actions follow standard RESTful responses:

**Collection actions:** Return serialized collection with meta

**Resource actions:** Return single serialized resource

**Position/settings updates:** Return updated index for UI refresh

**Error cases:** Standard error responses via RestfulController

---

## Serialization

Controllers use the standard serialization scope:
- DiscussionTemplateSerializer includes group, poll_templates associations
- PollTemplateSerializer includes group association

RecordCache optimizes N+1 queries for related records.
