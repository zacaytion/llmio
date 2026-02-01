# Search Domain - Frontend

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Search Modal Component

**Location:** `/vue/src/components/search/modal.vue`

The SearchModal is a Vue 3 component that provides the primary search interface in Loomio. It is displayed as a modal dialog accessible from multiple locations in the application.

---

## Component Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| initialOrgId | Number | null | Pre-select an organization to search within |
| initialGroupId | Number | null | Pre-select a specific group to search within |
| initialType | String | null | Pre-select a content type filter |
| initialQuery | String | null | Pre-populate the search query |

---

## User Interface Elements

### Search Input

- Text field with magnifying glass icon
- Submits on Enter key or icon click
- Shows loading indicator during API calls

### Filter Controls

Row of dropdown selects for narrowing results:

1. **Organization Selector**
   - "All groups" (search everything)
   - "Direct discussions" (no group, org_id = 0)
   - List of user's parent groups

2. **Subgroup Selector** (appears when org selected)
   - "All subgroups" (search entire org)
   - "Parent only" (exclude subgroups)
   - List of visible subgroups

3. **Tag Selector** (appears when group selected)
   - "Any tag" (no filter)
   - List of tags from selected group

4. **Content Type Selector**
   - "All content"
   - "Discussions"
   - "Comments"
   - "Decisions" (Polls)
   - "Votes" (Stances)
   - "Outcomes"

5. **Sort Order Selector**
   - "Best match" (relevance)
   - "Newest"
   - "Oldest"

### Results List

Each result displays:
- **Icon:** Poll icon for Poll/Outcome, author avatar for others
- **Title:** Poll title or discussion title
- **Tags:** Combined tags from poll and discussion
- **Timestamp:** When content was authored
- **Highlight:** Snippet with matched terms in bold
- **Metadata line:** Content type, author name, group name

---

## Data Flow

### Search Execution

1. User enters query and/or changes filters
2. Component builds request parameters:
   - query: search text
   - type: content type filter
   - org_id: organization scope
   - group_id: specific group scope
   - order: sort preference
   - tag: tag filter
3. Calls Records.remote.get('search', params)
4. API returns search_results array with denormalized data
5. Results stored in component's local state (not LokiJS)

### Result Navigation

When user clicks a result, urlForResult() generates the navigation URL:

| Content Type | URL Pattern |
|--------------|-------------|
| Discussion | `/d/{discussion_key}/{slug}` |
| Comment | `/d/{discussion_key}/comment/{comment_id}` |
| Poll, Stance, Outcome (in discussion) | `/d/{discussion_key}/{slug}/{sequence_id}` |
| Poll, Stance, Outcome (standalone) | `/p/{poll_key}/{slug}` |

The sequence_id enables navigation directly to the item's position in the discussion timeline.

---

## Entry Points

The SearchModal can be opened from three locations:

### Navbar

**Location:** `/vue/src/components/common/navbar.vue`

- Search icon button in top navigation
- Opens modal with no filters pre-applied
- Available on all pages when logged in

### Discussions Panel

**Location:** `/vue/src/components/group/discussions_panel.vue`

- "Search" action button on group discussions page
- Opens modal with:
  - initialOrgId set to parent group
  - initialGroupId set to current group
  - initialType set to "Discussion"

### Polls Panel

**Location:** `/vue/src/components/group/polls_panel.vue`

- "Search" action button on group polls page
- Opens modal with:
  - initialOrgId set to parent group
  - initialGroupId set to current group
  - initialType set to "Poll"

---

## Modal Launcher Integration

**Location:** `/vue/src/components/modal/launcher.vue`

SearchModal is registered as a lazy-loaded component in the modal launcher:

```pseudo
SearchModal: asyncComponent loading from '@/components/search/modal'
```

Modals are opened via EventBus:

```pseudo
EventBus.emit('openModal', {
  component: 'SearchModal',
  props: { initialOrgId, initialGroupId, initialType }
})
```

---

## Related Records Loading

After receiving search results, the component can access related records:

- **userById(id)** - Loads user from Records.users collection
- **pollById(id)** - Loads poll from Records.polls collection
- **groupById(id)** - Loads group from Records.groups collection

These are used for displaying avatars, poll icons, and tag information.

---

## Tag Integration

When a group is selected, the component loads available tags:

1. Fetches the group record via Records.groups.find()
2. Calls group.tagsByName() to get sorted tag list
3. Populates tagItems dropdown with tag options

Tags are matched against both poll.tags and discussion.tags in search results.

---

## Responsive Behavior

The search modal:
- Uses Vuetify's v-card for dialog structure
- Filter controls stack horizontally with responsive spacing
- Results list uses v-list with two-line items
- Closes automatically when navigating to a result (watches $route.path)
