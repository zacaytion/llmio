# Templates Domain: Services

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

The template services handle business logic for creating, updating, and managing templates. They follow the standard Loomio service pattern with authorization, validation, and persistence.

---

## DiscussionTemplateService

**Location:** `/app/services/discussion_template_service.rb`

### create(discussion_template:, actor:)

Creates a new custom discussion template.

**Flow:**
1. Authorize actor can create on the template
2. Assign actor as author
3. Validate template
4. Handle system template override:
   - If template has a key (based on system template), add that key to group's hidden_discussion_templates
   - Clear the key so this becomes a custom template
5. Save and return template

**Authorization:** Actor must be group admin (checked via ability)

**Return Value:** DiscussionTemplate on success, false on validation failure

### update(discussion_template:, params:, actor:)

Updates an existing discussion template.

**Flow:**
1. Authorize actor can update
2. Assign attributes except group_id (cannot change ownership)
3. Validate and save

**Return Value:** DiscussionTemplate on success, false on validation failure

### initial_templates(category, parent_id)

Returns the set of default templates to show for a new group based on its category.

**Categories and their templates:**
- **board:** blank, discuss_a_topic, practice_thread, approve_a_document, prepare_for_a_meeting, funding_decision
- **membership:** blank, discuss_a_topic, practice_thread, share_links_and_info, decision_by_consensus, elect_a_governance_position
- **self_managing:** blank, discuss_a_topic, practice_thread, advice_process, consent_process
- **other:** blank, discuss_a_topic, practice_thread, approve_a_document, advice_process, consent_process

**Subgroups** (parent_id present) default to: blank only
**Top-level fallback:** blank, practice_thread

### default_templates

Loads and returns all system discussion templates from YAML configuration.

**Flow:**
1. Reads AppConfig.discussion_templates (from `/config/discussion_templates.yml`)
2. For each template configuration:
   - Sets the key attribute
   - Processes `_i18n` suffixed attributes through I18n.t()
   - Builds a DiscussionTemplate instance (not persisted)
3. Returns array in reverse order (for display priority)

### create_public_templates

Seeds the public template gallery.

**Flow:**
1. Find or create the "templates" group (owned by helper_bot)
2. Create DiscussionTemplate records for all default templates
3. Mark them as public and authored by helper_bot

This is called lazily when browsing public templates.

---

## PollTemplateService

**Location:** `/app/services/poll_template_service.rb`

### group_templates(group:)

Returns all poll templates available to a group, combining custom and system templates.

**Flow:**
1. Start with group's custom poll_templates (database records)
2. Append system templates with:
   - Position from group.poll_template_positions hash
   - group_id set to current group
   - discarded_at set if key is in group.hidden_poll_templates

This merges custom templates with system defaults, allowing group-specific overrides.

### default_templates

Loads all system poll templates from YAML configuration.

**Flow:**
1. Reads AppConfig.poll_templates from `/config/poll_templates.yml`
2. For each template:
   - Loads base defaults from the poll_type's defaults in `/config/poll_types.yml`
   - Overlays template-specific values
   - Processes `_i18n` suffixes through I18n.t()
   - Processes poll_options array similarly
3. Returns array of unsaved PollTemplate instances

### create(poll_template:, actor:)

Creates a new custom poll template.

**Flow:**
1. Authorize actor can create
2. Assign actor as author
3. Validate
4. Handle system template override:
   - If template has a key, add it to group.hidden_poll_templates
   - Clear the key for custom template
5. Save and return

**Return Value:** PollTemplate on success, false on failure

### update(poll_template:, params:, actor:)

Updates an existing poll template.

**Flow:**
1. Authorize
2. Assign attributes (excluding group_id)
3. Validate and save

---

## Authorization Pattern

Both services use the same authorization approach:

```
PSEUDO-CODE:
authorize(actor, action, template)
  ability = actor.ability
  ability.authorize!(action, template)
  // Raises CanCan::AccessDenied if not permitted
```

For templates, only group admins can create and update:
- Check: template.group.admins.exists?(user.id)

---

## Template Hiding vs Discarding

The services support two distinct hide mechanisms:

### System Template Hiding

For built-in templates (those with a key):
- Add key to group.hidden_poll_templates or hidden_discussion_templates
- Template is not deleted, just filtered from display
- Can be unhidden by removing from the array

### Custom Template Discarding

For database-stored templates (those with an id):
- Set discarded_at timestamp via Discard gem
- Template stays in database but excluded from normal queries
- Can be undiscarded by clearing the timestamp

### Template Override Pattern

When editing a system template:
1. Create form loaded with system template values
2. On save, system template key added to hidden list
3. New custom template created without key
4. Original system template hidden, custom version shown

This allows customization while preserving ability to restore defaults.

---

## Internationalization Support

Both services handle i18n in template loading:

**Pattern:**
```
PSEUDO-CODE:
for each attribute in template_config:
  if attribute ends with "_i18n":
    real_name = remove "_i18n" suffix
    if value is array:
      result[real_name] = value.map(v => I18n.t(v))
    else:
      result[real_name] = I18n.t(value)
  else:
    result[attribute] = value
```

This enables storing i18n keys in YAML and resolving them at runtime based on user locale.

---

## Service Integration Points

### With Group Model

- Services read/write to group.hidden_*_templates arrays
- Services use group.poll_template_positions for ordering
- Counter caches updated automatically via model callbacks

### With Controllers

- Controllers delegate to services for mutations
- Services handle authorization, controllers trust service responses
- Return values indicate success/failure

### With Frontend

- Services are not called directly from frontend
- API endpoints wrap service calls
- Frontend receives serialized results
