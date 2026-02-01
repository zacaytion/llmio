# Templates Domain: Frontend

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

The frontend template system provides:
- Template selection when starting new discussions/polls
- Template management for group admins
- Public template browsing and importing
- Real-time template ordering via drag-and-drop

---

## Models (LokiJS Records)

### DiscussionTemplateModel

**Location:** `/vue/src/shared/models/discussion_template_model.js`

**Unique Indices:** id, key
**Indices:** groupId

**Default Values:**
- description: null, descriptionFormat: 'html'
- processIntroduction: null, processIntroductionFormat: 'html'
- title: null, tags: [], maxDepth: 3, newestFirst: false
- pollTemplateKeysOrIds: [], recipientAudience: null
- discardedAt: null

**Relationships:**
- belongsTo author (from users)
- belongsTo group

**Key Methods:**

**buildDiscussion()** - Creates a new Discussion record from template
```
PSEUDO-CODE:
discussion = Records.discussions.build()
copy template attributes to discussion
set discussionTemplateId = this.id
set discussionTemplateKey = this.key
set authorId = current user
return discussion
```

**pollTemplates()** - Returns associated PollTemplate records
```
PSEUDO-CODE:
return pollTemplateKeysOrIds.map(keyOrId =>
  Records.pollTemplates.find(keyOrId)
).compact()
```

**pollTemplateIds()** - Extracts only numeric IDs (custom templates)

**pollTemplateKeys()** - Extracts only string keys (system templates)

### PollTemplateModel

**Location:** `/vue/src/shared/models/poll_template_model.js`

**Unique Indices:** id, key
**Indices:** groupId

**Default Values:**
- Most poll configuration fields with sensible defaults
- pollType: null, defaultDurationInDays: 7
- pollOptions: [], stanceReasonRequired: 'optional'
- notifyOnClosingSoon: 'undecided_voters'

**Key Methods:**

**config()** - Returns poll type configuration from AppConfig

**buildPoll()** - Creates a new Poll record from template
```
PSEUDO-CODE:
poll = Records.polls.build()
copy template attributes
set pollTemplateId = this.id
set pollTemplateKey = this.key
set closingAt = now + defaultDurationInDays
set pollOptionsAttributes from pollOptions
return poll
```

**pollOptionsAttributes()** - Converts pollOptions to API format

**translatedPollType()** - Returns localized poll type name

---

## Records Interfaces

### DiscussionTemplateRecordsInterface

**Location:** `/vue/src/shared/interfaces/discussion_template_records_interface.js`

Basic interface extending BaseRecordsInterface. Inherits standard CRUD methods.

### PollTemplateRecordsInterface

**Location:** `/vue/src/shared/interfaces/poll_template_records_interface.js`

**Additional Methods:**

**fetchByGroupId(groupId)** - Loads all templates for a group

**findOrFetchByKeyOrId(keyOrId)** - Finds cached or fetches from API

---

## Services

### DiscussionTemplateService (Frontend)

**Location:** `/vue/src/shared/services/discussion_template_service.js`

Note: Named "PollTemplateService" in file but handles discussion templates.

**actions(discussionTemplate, group)** - Returns available actions for a template:

| Action | Icon | Condition | Behavior |
|--------|------|-----------|----------|
| edit_default_template | mdi-pencil | No id, is admin | Navigate to new template form with key |
| edit_template | mdi-pencil | Has id, is admin | Navigate to edit form |
| move | mdi-arrow-up-down | Not discarded, is admin | Emit sortThreadTemplates event |
| discard | mdi-eye-off | Has id, not discarded, is admin | POST discard API |
| undiscard | mdi-eye | Has id, discarded, is admin | POST undiscard API |
| destroy | mdi-delete | Has id, is admin | Confirm modal then destroy |
| hide | mdi-eye-off | No id, has key, not discarded | POST hide API |
| unhide | mdi-eye | No id, has key, discarded | POST unhide API |

### PollTemplateService (Frontend)

**Location:** `/vue/src/shared/services/poll_template_service.js`

Same pattern as DiscussionTemplateService with equivalent actions for poll templates.

---

## Components

### Poll Template Selection

**ChooseTemplate**
**Location:** `/vue/src/components/poll/common/choose_template.vue`

Main component for selecting poll templates when creating a poll.

**Features:**
- Filter tabs by category (proposal/poll/meeting) or show all
- "Recommended" filter when discussion has associated templates
- Admin settings tab showing hidden templates
- Drag-and-drop reordering
- Action menus per template

**Props:** discussion, group

**Data State:**
- pollTemplates: Current filtered list
- isSorting: Drag-drop mode active
- filter: Current category filter
- singleList: Toggle for category mode vs flat list

**Key Methods:**

**query()** - Refreshes template list based on filters
```
PSEUDO-CODE:
if filter == 'recommended':
  templates = discussionTemplate.pollTemplates()
else if categorized:
  filter by pollType category
else:
  show all non-discarded
sort by position
```

**cloneTemplate(template)** - Creates poll from template
```
PSEUDO-CODE:
poll = template.buildPoll()
if in discussion context:
  poll.discussionId = discussion.id
  poll.groupId = discussion.groupId
emit('setPoll', poll)
```

**sortEnded()** - Saves new positions after drag
```
PSEUDO-CODE:
ids = templates.map(t => t.id || t.key)
POST poll_templates/positions with ids
```

**ChooseTemplateWrapper**
**Location:** `/vue/src/components/poll/common/choose_template_wrapper.vue`

Wrapper component for template chooser.

**TemplateBanner**
**Location:** `/vue/src/components/poll/template_banner.vue`

Shows info banner when viewing a poll that is itself a template.

### Discussion Template Pages

**IndexPage**
**Location:** `/vue/src/components/thread_template/index_page.vue`

Lists all discussion templates for a group.

**Features:**
- Settings toggle for hidden templates view
- Drag-drop reordering
- Admin menu for creating/managing templates
- Link to browse public templates

**Key Interactions:**
- sortThreadTemplates event triggers sort mode
- reloadThreadTemplates refreshes list
- Templates link to `/d/new?template_id=X&group_id=Y`

**FormPage**
**Location:** `/vue/src/components/thread_template/form_page.vue`

Wrapper for template edit/create form.

**Form**
**Location:** `/vue/src/components/thread_template/form.vue`

Full template editing form.

**Sections:**
1. Process metadata (name, subtitle, introduction)
2. Default content (title, placeholder, tags, description)
3. Poll templates (sortable list with add/remove)
4. Reply arrangement (newestFirst, maxDepth)
5. Public sharing toggle

**Poll Template Association:**
- Select dropdown to add poll templates
- Drag-drop to reorder
- Remove button per template
- Saves as pollTemplateKeysOrIds array

**BrowsePage**
**Location:** `/vue/src/components/thread_template/browse_page.vue`

Public template gallery browser.

---

## Frontend Routes

**Location:** `/vue/src/routes.js`

| Path | Component | Purpose |
|------|-----------|---------|
| /thread_templates | IndexPage | List group's discussion templates |
| /thread_templates/new | FormPage | Create new discussion template |
| /thread_templates/:id | FormPage | Edit existing template |
| /thread_templates/browse | BrowsePage | Browse public templates |
| /poll_templates/new | PollTemplateForm | Create poll template |
| /poll_templates/:id/edit | PollTemplateForm | Edit poll template |

---

## Template Usage Flow

### Creating a Discussion from Template

1. User navigates to `/thread_templates?group_id=X`
2. IndexPage loads group's templates
3. User clicks template, navigates to `/d/new?template_id=X&group_id=Y`
4. Discussion form loads with template values pre-filled
5. User modifies and saves

### Creating a Poll from Template

1. User clicks "New poll" or opens poll chooser
2. ChooseTemplate component displays
3. User clicks template
4. cloneTemplate() creates poll with template values
5. Poll form opens with pre-filled values
6. User modifies and saves

### Template Admin Flow

1. Admin accesses template settings via cog menu
2. Can reorder by dragging (positions endpoint)
3. Can hide/unhide system templates
4. Can discard/undiscard custom templates
5. Can delete custom templates permanently
6. Can toggle categorized vs flat display

---

## Real-time Updates

Templates use WatchRecords mixin for reactivity:

```
PSEUDO-CODE:
watchRecords({
  collections: ['pollTemplates'],
  query: () => this.refreshTemplateList()
})
```

Changes to template records automatically refresh UI.

Position and settings changes broadcast via MessageChannelService for multi-tab sync.
