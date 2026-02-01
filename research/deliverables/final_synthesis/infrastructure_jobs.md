# Infrastructure & Background Jobs - Implementation Synthesis

## Executive Summary

Loomio uses Sidekiq with 38 workers across 5 queue priorities.

---

## Confirmed Configuration

### Queue Priorities

| Priority | Queue | Purpose |
|----------|-------|---------|
| 10 | critical | Time-sensitive operations |
| 6 | high | Important but less urgent |
| 3 | default | Standard processing |
| 1 | low | Background maintenance |
| 1 | mailers | Email delivery |

### Sidekiq Concurrency

- **Production:** 20 threads
- **Development:** 5 threads

---

## Worker Inventory (38)

### Core Workers (Implement First)

| Worker | Purpose | Queue |
|--------|---------|-------|
| `publish_event_worker.rb` | Event publishing | default |
| `generic_worker.rb` | Generic job wrapper | default |
| `close_expired_poll_worker.rb` | Close lapsed polls | default |
| `send_daily_catch_up_email_worker.rb` | Daily digest | mailers |

### Notification Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `announce_discussion_worker.rb` | Discussion announcements | high |

### Poll Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `update_poll_counts_worker.rb` | Counter cache update | default |
| `reset_poll_stance_data_worker.rb` | Reset poll data | low |

### Cleanup Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `destroy_discussion_worker.rb` | Async discussion delete | low |
| `destroy_group_worker.rb` | Async group delete | low |
| `destroy_record_worker.rb` | Generic record delete | low |
| `destroy_tag_worker.rb` | Tag deletion | low |
| `destroy_user_worker.rb` | User deletion | low |

### User Management Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `deactivate_user_worker.rb` | User deactivation | default |
| `redact_user_worker.rb` | GDPR data redaction | low |
| `geo_location_worker.rb` | Geolocation lookup | low |

### Export Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `group_export_worker.rb` | Full export | low |
| `group_export_csv_worker.rb` | CSV export | low |

### Content Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `move_comments_worker.rb` | Move comments | default |
| `append_transcript_worker.rb` | Append transcripts | default |
| `attach_document_worker.rb` | Attach documents | default |
| `download_attachment_worker.rb` | Download attachments | low |
| `add_heading_ids_worker.rb` | Content processing | low |
| `undelete_blob_worker.rb` | Restore deleted files | low |
| `update_tag_worker.rb` | Tag updates | default |

### Membership Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `accept_membership_worker.rb` | Accept membership | high |
| `revoke_memberships_of_deactivated_users_worker.rb` | Cleanup | low |

### Maintenance Workers

| Worker | Purpose | Queue |
|--------|---------|-------|
| `repair_thread_worker.rb` | Thread repair | low |
| `update_blocked_domains_worker.rb` | Update blocklist | low |

### Migration Workers (Can Skip)

| Worker | Purpose | Notes |
|--------|---------|-------|
| `add_group_id_to_documents_worker.rb` | One-time migration | Skip |
| `convert_discussion_templates_worker.rb` | One-time migration | Skip |
| `convert_poll_stances_in_discussion_worker.rb` | One-time migration | Skip |
| `migrate_discussion_readers_for_deactivated_members_worker.rb` | One-time migration | Skip |
| `migrate_guest_on_discussion_readers_and_stances.rb` | One-time migration | Skip |
| `migrate_poll_templates_worker.rb` | One-time migration | Skip |
| `migrate_tags_worker.rb` | One-time migration | Skip |
| `migrate_user_worker.rb` | One-time migration | Skip |
| `fix_stances_missing_from_threads_worker.rb` | One-time fix | Skip |
| `remove_poll_expired_from_threads_worker.rb` | One-time cleanup | Skip |

---

## Scheduled Tasks

### Hourly

| Task | Purpose |
|------|---------|
| Close expired polls | Close lapsed polls |
| Send reminders | Vote reminders |

### Daily (6 AM User Timezone)

| Task | Purpose | Notes |
|------|---------|-------|
| Daily catch-up emails | Daily digest | Per-user timezone |
| Outcome review | Review due outcomes | |
| Task due reminders | Task reminders | |

### Weekly (Monday)

| Task | Purpose |
|------|---------|
| Weekly catch-up emails | Weekly digest |
| Group statistics | Usage stats |

### Monthly

| Task | Purpose |
|------|---------|
| Update blocked domains | Update blocklist |
| Cleanup inactive | Remove stale data |

---

## Retry Configuration

### Sidekiq Defaults

- **Default retries**: 25
- **Backoff**: Exponential (sidekiq_retry_in formula)
- **Dead queue**: After 25 failures

### Custom Retry for Critical Jobs

Critical jobs should retry faster with custom backoff:
- Retry in 30 seconds, 1 min, 2 min, 5 min...

---

## Monitoring

A job monitoring UI should be available for tracking:
- Pending jobs by queue
- Running jobs
- Completed jobs
- Failed jobs
