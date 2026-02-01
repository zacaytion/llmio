# Export Domain: Services

**Generated:** 2026-02-01
**Domain:** Export functionality for groups, discussions, and polls

---

## Overview

The export domain has one main service class and two exporter helper classes that handle data collection and formatting for different export types.

---

## 1. GroupExportService

**Location:** `/app/services/group_export_service.rb`

The primary service for JSON export and import operations. This is a comprehensive service that handles full data portability.

### Constants

#### RELATIONS
List of association methods to iterate when exporting a group:
- `all_users`, `all_events`, `all_notifications`, `all_reactions`, `all_tags`
- `poll_templates`, `discussion_templates`
- `memberships`, `membership_requests`
- `discussions`, `comments`, `discussion_readers`
- `exportable_polls`, `exportable_poll_options`, `exportable_outcomes`
- `exportable_stances`, `exportable_stance_choices`, `poll_stance_receipts`

#### JSON_PARAMS
Field exclusions for sensitive data:
- Groups: excludes `token`
- Users: excludes `encrypted_password`, `reset_password_token`, `email_api_key`, `secret_token`, `unsubscribe_token`

#### BACK_REFERENCES
Mapping of foreign key relationships for import ID remapping. Defines which tables reference which other tables and through which columns. Critical for maintaining referential integrity during import.

### Class Methods

#### export(groups, group_name)

**Purpose:** Export one or more groups to a JSON file

**Process:**
1. Generate timestamped filename in /tmp directory
2. Open file for writing
3. For each group:
   - Write the group record as JSON
   - Iterate through all RELATIONS, writing each record
   - Use batch processing (20,000 records per batch) via `find_each`
   - Track already-written IDs to avoid duplicates
4. Collect and write all attachments:
   - User avatars
   - Group assets (cover photos, logos, files)
   - Related content attachments (comments, discussions, polls, outcomes, subgroups)
5. Return the filename

**Output Format:** JSON Lines (one JSON object per line)
```
{"table": "groups", "record": {...}}
{"table": "users", "record": {...}}
{"table": "discussions", "record": {...}}
...
{"table": "attachments", "record": {...}}
```

#### export_filename_for(group_name)

**Purpose:** Generate standardized export filename

**Format:** `/tmp/YYYY-MM-DD_HH-MM-SS_group-name-parameterized.json`

#### export_direct_threads(group_id)

**Purpose:** Export invite-only (direct) discussions created by group members

**Use Case:** Exporting standalone discussions that are not formally part of the group but were created by group members

**Note:** This method has a TODO comment indicating it should be integrated into the normal export process

#### import(filename_or_url, reset_keys: false)

**Purpose:** Import a previously exported JSON file

**Process:**
1. Read file (supports local path or URL)
2. Parse each line as JSON
3. For each table (except attachments):
   - Create new records
   - Track old ID to new ID mapping
   - Handle key regeneration if `reset_keys: true`
   - Regenerate security tokens
   - Suggest new handle for groups to avoid conflicts
4. Rewrite all foreign key references using the ID mapping
5. Process attachments asynchronously via DownloadAttachmentWorker
6. Update poll counts and stance option scores

**Transaction Safety:** Entire import runs within a database transaction

#### puts_record(record, file, ids)

**Purpose:** Write a single record to the export file

**Behavior:**
- Skips if record ID already written (deduplication)
- Applies JSON_PARAMS exclusions
- Writes as JSON with table name

#### puts_attachment(attachment, file)

**Purpose:** Write attachment metadata to export file

**Includes:**
- Attachment ID, filename, content type
- Download path and full URL
- Record association info

#### download_attachment(record_data, new_id)

**Purpose:** Download and re-attach an attachment during import

**Process:**
1. Open URL and download file
2. Create new ActiveStorage blob
3. Attach to the imported record
4. Rebuild attachments array if applicable

---

## 2. GroupExporter

**Location:** `/app/extras/group_exporter.rb`

A simpler exporter for CSV and HTML formats with a limited field set.

### Export Models Configuration

Defines exportable fields for each model type:
- `groups`: id, key, name, description, created_at
- `memberships`: group_id, user_id, user_name, user_email, admin, created_at, accepted_at
- `discussions`: id, group_id, author_id, author_name, title, description, created_at
- `comments`: id, group_id, discussion_id, author_id, author_name, title, body, created_at
- `polls`: id, key, discussion_id, group_id, author_id, author_name, title, details, closing_at, closed_at, created_at, poll_type, custom_fields
- `stances`: id, poll_id, participant_id, author_name, reason, latest, created_at, updated_at
- `outcomes`: id, poll_id, author_id, statement, created_at, updated_at

### Instance Methods

#### initialize(group)
Sets up the exporter with a target group

#### to_csv(opts = {})
Generates CSV output with all models as sections:
1. Header with group name
2. For each model type:
   - Section header with count
   - Column headers (humanized field names)
   - Data rows

#### Dynamic Methods
For each model in EXPORT_MODELS:
- `{model}`: Returns records via `{Model}.in_organisation(group)` scope
- `{model}_fields`: Returns field list for that model

---

## 3. PollExporter

**Location:** `/app/extras/poll_exporter.rb`

Specialized exporter for individual poll data in CSV format.

### Instance Methods

#### initialize(poll)
Sets up the exporter with a target poll

#### file_name
Returns formatted filename: `poll-{id}-{key}-{title-parameterized}.csv`

#### meta_table
Returns poll metadata hash:
- Poll attributes: id, group_id, discussion_id, author_id, title, author_name, dates, voter counts, details
- Context: group_name, discussion_title, poll_url
- Outcome data: author, statement, created_at

#### to_csv(opts = {})
Generates comprehensive poll export:
1. **Poll Section:** Metadata keys and values
2. **Poll Options Section:** Results with scores, percentages, rankings
3. **Votes Section:** Each stance with voter info, timestamps, reason, and score for each option

**Vote Data Columns:**
- id, poll_id, voter_id, voter_name, created_at, updated_at, reason, reason_format
- One column per poll option showing the voter's score for that option

---

## 4. GroupService (Export Integration)

**Location:** `/app/services/group_service.rb`

Contains the `export` method that bridges the controller to the worker.

### export(group:, actor:)

**Authorization:** Requires `:show` permission on group (verified via ability)

**Scope Limitation:** Only exports groups the actor is a member of within the target group hierarchy

**Process:**
1. Authorize actor can show the group
2. Filter to only include groups where actor has membership
3. Enqueue GroupExportWorker with group IDs, name, and actor ID

---

## 5. Data Flow Summary

### JSON Export Flow
```
User Request
    |
GroupsController#export
    |
GroupService.export
    |-- Authorize (show permission)
    |-- Filter to actor's groups
    |
GroupExportWorker (async)
    |
GroupExportService.export
    |-- Iterate groups and relations
    |-- Batch process (20k records)
    |-- Collect attachments
    |
Document created with file
    |
UserMailer.group_export_ready
    |
DestroyRecordWorker scheduled (1 week)
```

### CSV Export Flow
```
User Request
    |
GroupsController#export_csv
    |
GroupExportCsvWorker (async)
    |
GroupExporter.to_csv
    |
Document created with CSV
    |
UserMailer.group_export_ready
    |
DestroyRecordWorker scheduled (1 week)
```

### Poll Export Flow
```
User Request
    |
PollsController#export
    |
PollExporter.to_csv
    |
Direct CSV download (synchronous)
```

---

## 6. Size and Performance Considerations

### Batch Processing
- Uses `find_each(batch_size: 20000)` to process records in batches
- Prevents memory exhaustion on large groups
- Maintains constant memory footprint regardless of group size

### No Explicit Size Limits
- No hard limits on export size found in the codebase
- Large exports are handled by background workers to avoid timeout
- File storage is in /tmp, subject to server disk space

### Deduplication
- Tracks exported record IDs to prevent duplicate entries
- Particularly important for union queries that may overlap

---

## Confidence Rating: 5/5

The service layer is well-documented in code with clear responsibilities. The JSON export/import service is comprehensive, while the CSV exporters are straightforward data formatters.
