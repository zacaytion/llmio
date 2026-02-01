# Infrastructure & Background Jobs - Follow-up Analysis

## Executive Summary

Background job processing was documented in `research/investigation/jobs.md` but **was not included in the third-party follow-up investigation**. This document captures the 38 worker inventory and open questions.

---

## Source Code Verification

### Worker Files Inventory

Located at `orig/loomio/app/workers/`:

| Worker | Purpose | Category |
|--------|---------|----------|
| `accept_membership_worker.rb` | Accept membership request | Membership |
| `add_group_id_to_documents_worker.rb` | Migration helper | Migration |
| `add_heading_ids_worker.rb` | Content processing | Content |
| `announce_discussion_worker.rb` | Discussion announcements | Notifications |
| `append_transcript_worker.rb` | Append transcripts | Content |
| `attach_document_worker.rb` | Attach documents | Content |
| `close_expired_poll_worker.rb` | Close lapsed polls | Polls |
| `convert_discussion_templates_worker.rb` | Migration helper | Migration |
| `convert_poll_stances_in_discussion_worker.rb` | Migration helper | Migration |
| `deactivate_user_worker.rb` | User deactivation | User Mgmt |
| `destroy_discussion_worker.rb` | Async discussion delete | Cleanup |
| `destroy_group_worker.rb` | Async group delete | Cleanup |
| `destroy_record_worker.rb` | Generic record delete | Cleanup |
| `destroy_tag_worker.rb` | Tag deletion | Cleanup |
| `destroy_user_worker.rb` | User deletion | User Mgmt |
| `download_attachment_worker.rb` | Attachment download | Content |
| `fix_stances_missing_from_threads_worker.rb` | Data fix | Migration |
| `generic_worker.rb` | Generic job wrapper | Core |
| `geo_location_worker.rb` | Geolocation lookup | User Mgmt |
| `group_export_csv_worker.rb` | CSV export | Export |
| `group_export_worker.rb` | Full export | Export |
| `migrate_discussion_readers_for_deactivated_members_worker.rb` | Migration helper | Migration |
| `migrate_guest_on_discussion_readers_and_stances.rb` | Guest migration | Migration |
| `migrate_poll_templates_worker.rb` | Template migration | Migration |
| `migrate_tags_worker.rb` | Tag migration | Migration |
| `migrate_user_worker.rb` | User migration | Migration |
| `move_comments_worker.rb` | Move comments between threads | Content |
| `publish_event_worker.rb` | Event publishing | Events |
| `redact_user_worker.rb` | User data redaction (GDPR) | User Mgmt |
| `remove_poll_expired_from_threads_worker.rb` | Cleanup | Cleanup |
| `repair_thread_worker.rb` | Thread repair | Maintenance |
| `reset_poll_stance_data_worker.rb` | Reset poll data | Polls |
| `revoke_memberships_of_deactivated_users_worker.rb` | Cleanup | Cleanup |
| `send_daily_catch_up_email_worker.rb` | Daily digest | Email |
| `undelete_blob_worker.rb` | Restore deleted files | Content |
| `update_blocked_domains_worker.rb` | Update blocklist | Maintenance |
| `update_poll_counts_worker.rb` | Counter cache update | Polls |
| `update_tag_worker.rb` | Tag updates | Content |

**Total: 38 workers**

---

## Open Questions for Third Party

### HIGH Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 1 | **What are the Sidekiq retry settings?** | Reliability | Default retries, backoff formula |
| 2 | **Is there a dead letter queue?** | Observability | Failed job handling |
| 3 | **Which workers use which queues?** | Priority | Queue assignment per worker |

### MEDIUM Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 4 | What is the scheduled task configuration? | Operations | `config/schedule.yml` or equivalent |
| 5 | Are there job dependencies? (A must complete before B) | Ordering | Job orchestration patterns |
| 6 | How is the daily 6 AM user timezone calculated? | Email timing | Timezone-aware scheduling |

### LOW Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 7 | Are there job metrics/monitoring? | Observability | Sidekiq dashboard, APM |
| 8 | What happens if Redis is unavailable? | Failure mode | Fallback behavior |

---

## Confirmed Configuration

### Queue Priorities

From `investigation/jobs.md`:

| Priority | Queue | Purpose |
|----------|-------|---------|
| 10 | critical | Time-sensitive operations |
| 6 | high | Important but less urgent |
| 3 | default | Standard processing |
| 1 | low | Background maintenance |
| 1 | mailers | Email delivery |

### Scheduled Tasks

| Frequency | Task | Purpose |
|-----------|------|---------|
| Hourly | PollClosingWorker | Close lapsed polls |
| Hourly | SendRemindersWorker | Vote reminders |
| Daily (6 AM) | SendDailyCatchUpEmailWorker | Daily digest |
| Daily | OutcomeReviewWorker | Review due outcomes |
| Daily | TaskDueReminderWorker | Task due reminders |
| Weekly (Monday) | SendWeeklyCatchUpEmailWorker | Weekly digest |
| Monthly | UpdateBlockedDomainsWorker | Update blocklist |

---

## Discrepancies

### Worker Count Mismatch

**Investigation claimed:** 38 workers
**Verified count:** 38 workers (confirmed)

### Missing Workers from Investigation

The investigation/jobs.md lists categories but not all workers. Missing from docs:
- Migration workers (9 total)
- `generic_worker.rb` - Important wrapper pattern
- `redact_user_worker.rb` - GDPR compliance

### Unknown: Retry Configuration

**Not documented:** Sidekiq retry settings

Default Sidekiq behavior:
- 25 retries over ~21 days
- Exponential backoff

**Question:** Does Loomio use defaults or custom?

---

## Files Requiring Investigation

| File | Purpose | Priority |
|------|---------|----------|
| `config/sidekiq.yml` | Queue/concurrency config | HIGH |
| `config/schedule.yml` | Scheduled tasks | HIGH |
| `app/workers/generic_worker.rb` | Base worker pattern | MEDIUM |
| `app/workers/publish_event_worker.rb` | Event publishing | MEDIUM |

---

## Priority Assessment

| Area | Priority | Blocking? |
|------|----------|-----------|
| Queue assignment | HIGH | Yes - affects job ordering |
| Retry configuration | HIGH | Yes - affects reliability |
| Scheduled task timing | MEDIUM | No - can adjust post-launch |
| Migration workers | LOW | No - one-time use |
