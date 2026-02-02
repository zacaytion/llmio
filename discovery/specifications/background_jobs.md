# Background Jobs Analysis

Investigation of Sidekiq configuration, workers, queues, and scheduled tasks.

---

## 1. Sidekiq Retry Configuration

### Global Defaults

**Sidekiq 7.x defaults** (no explicit override found):
- **Default retries**: 25 retries with exponential backoff
- **Backoff formula**: `(retry_count ** 4) + 15 + (rand(30) * (retry_count + 1))` seconds
- **Dead job retention**: 6 months

**Source**: `Gemfile:24` - `gem 'sidekiq', '~> 7.0'`

### Global Options Set

| Setting | Value | Location |
|---------|-------|----------|
| `backtrace` | `true` | `config/initializers/sidekiq.rb:1` |

```ruby
# config/initializers/sidekiq.rb:1
Sidekiq.default_job_options = { 'backtrace' => true }
```

**Confidence**: HIGH - Explicit configuration found in initializer.

### Worker-Specific Retry Overrides

| Worker | Retry Setting | Queue | Location |
|--------|---------------|-------|----------|
| `SendDailyCatchUpEmailWorker` | `false` | default | `app/workers/send_daily_catch_up_email_worker.rb:3` |
| `RepairThreadWorker` | `false` | default | `app/workers/repair_thread_worker.rb:3` |
| `ConvertPollStancesInDiscussionWorker` | `false` | low | `app/workers/convert_poll_stances_in_discussion_worker.rb:3` |
| `RemovePollExpiredFromThreadsWorker` | `false` | low | `app/workers/remove_poll_expired_from_threads_worker.rb:3` |
| `ResetPollStanceDataWorker` | `false` | low | `app/workers/reset_poll_stance_data_worker.rb:3` |
| `UpdatePollCountsWorker` | `false` | low | `app/workers/update_poll_counts_worker.rb:3` |

**Note**: 6 workers explicitly disable retries. All other workers use Sidekiq's default of 25 retries.

**Confidence**: HIGH - All `sidekiq_options` declarations found via grep search.

---

## 2. Queue Configuration

### Queue List (Order = Priority)

From `config/sidekiq.yml:6-16`:

```yaml
:queues:
  - critical
  - login_emails
  - mailers
  - notification_emails
  - default
  - action_mailbox_routing
  - active_storage_analysis
  - active_storage_purge
  - low
  - low_priority
```

**Queue Processing Order** (higher = processed first):
1. `critical` - Highest priority
2. `login_emails` - Login/authentication emails
3. `mailers` - General email delivery (ActionMailer default)
4. `notification_emails` - Event notification emails
5. `default` - Standard Sidekiq default queue
6. `action_mailbox_routing` - Rails Action Mailbox
7. `active_storage_analysis` - File analysis jobs
8. `active_storage_purge` - File cleanup jobs
9. `low` - Low priority batch jobs
10. `low_priority` - Lowest priority jobs

**Note**: Sidekiq 7 processes queues in list order (first = highest priority). No weighted queue configuration found.

**Confidence**: HIGH - Direct configuration file.

### Queue Usage by Framework

| Queue | Used By | Configuration |
|-------|---------|---------------|
| `mailers` | ActionMailer `.deliver_later` | `config/initializers/new_framework_defaults_6_1.rb:63` |
| `default` | Most workers (no explicit queue) | Sidekiq default |
| `low` | Migration/batch workers | Explicit `sidekiq_options queue: :low` |

**Confidence**: HIGH for documented queues, MEDIUM for `login_emails`/`notification_emails` (defined but no explicit worker usage found - may be used by external configuration or Devise).

---

## 3. Complete Worker Inventory

### All 38 Workers

| Worker | Queue | Retry | Purpose |
|--------|-------|-------|---------|
| `AcceptMembershipWorker` | default | 25 | Redeem pending membership invitations |
| `AddGroupIdToDocumentsWorker` | default | 25 | Migration: Add group_id to orphaned documents |
| `AddHeadingIdsWorker` | default | 25 | Migration: Add IDs to HTML headings |
| `AnnounceDiscussionWorker` | default | 25 | Send discussion invitations |
| `AppendTranscriptWorker` | default | 25 | Append transcription text to records |
| `AttachDocumentWorker` | default | 25 | Migration: Attach S3 documents to ActiveStorage |
| `CloseExpiredPollWorker` | default | 25 | Close polls past their closing date |
| `ConvertDiscussionTemplatesWorker` | default | 25 | Migration: Convert template discussions |
| `ConvertPollStancesInDiscussionWorker` | **low** | **false** | Migration: Convert poll stances to discussion items |
| `DeactivateUserWorker` | default | 25 | Deactivate user account and revoke memberships |
| `DestroyDiscussionWorker` | default | 25 | Permanently delete discarded discussions |
| `DestroyGroupWorker` | default | 25 | Permanently delete archived groups |
| `DestroyRecordWorker` | default | 25 | Generic record destruction (scheduled cleanup) |
| `DestroyTagWorker` | default | 25 | Delete tags and remove from discussions/polls |
| `DestroyUserWorker` | default | 25 | Permanently delete user record |
| `DownloadAttachmentWorker` | default | 25 | Download attachments during export |
| `FixStancesMissingFromThreadsWorker` | default | 25 | Migration: Fix missing stance events in threads |
| `GenericWorker` | default | 25 | Execute arbitrary class/method combinations |
| `GeoLocationWorker` | default | 25 | Update user country from IP address |
| `GroupExportCsvWorker` | default | 25 | Generate CSV export and email download link |
| `GroupExportWorker` | default | 25 | Generate full group export (JSON/files) |
| `MigrateDiscussionReadersForDeactivatedMembersWorker` | default | 25 | Migration: Update reader records for deactivated users |
| `MigrateGuestOnDiscussionReadersAndStances` | default | 25 | Migration: Mark guest users on readers/stances |
| `MigratePollTemplatesWorker` | default | 25 | Migration: Convert poll templates |
| `MigrateTagsWorker` | default | 25 | Migration: Convert taggings to JSONB arrays |
| `MigrateUserWorker` | default | 25 | Merge two user accounts |
| `MoveCommentsWorker` | default | 25 | Move comments between discussions |
| `PublishEventWorker` | default | 25 | Trigger event notifications |
| `RedactUserWorker` | default | 25 | Permanently anonymize user data (GDPR) |
| `RemovePollExpiredFromThreadsWorker` | **low** | **false** | Remove poll_expired events from threads |
| `RepairThreadWorker` | default | **false** | Repair discussion event sequence |
| `ResetPollStanceDataWorker` | **low** | **false** | Recalculate poll stance/option data |
| `RevokeMembershipsOfDeactivatedUsersWorker` | default | 25 | Batch revoke memberships for deactivated users |
| `SendDailyCatchUpEmailWorker` | default | **false** | Send daily/weekly catch-up digest emails |
| `UndeleteBlobWorker` | default | 25 | Restore S3 object from delete marker |
| `UpdateBlockedDomainsWorker` | default | 25 | Refresh blocked domain list from external source |
| `UpdatePollCountsWorker` | **low** | **false** | Recalculate poll vote counts |
| `UpdateTagWorker` | default | 25 | Rename/recolor tags across group hierarchy |

**Statistics**:
- Total workers: 38
- Using `default` queue: 33
- Using `low` queue: 5
- With `retry: false`: 6
- Migration workers: ~12

**Confidence**: HIGH - All 38 workers read and analyzed.

---

## 4. Scheduled Tasks

### Primary Scheduler: External Cron/Heroku Scheduler

Loomio uses an **external scheduler** (Heroku Scheduler or system cron) to trigger a single rake task that orchestrates all scheduled work.

**Evidence**:
- No `config/schedule.rb` (Whenever gem)
- No `config/sidekiq_cron.yml` (Sidekiq-Cron)
- Single `hourly_tasks` rake task contains all scheduling logic

**Confidence**: HIGH - No scheduler gem configuration found; rake task structure confirms external triggering.

### Scheduled Task List

#### Hourly Tasks (Every Hour)

| Task | Implementation | Description |
|------|----------------|-------------|
| Throttle reset (hourly) | `ThrottleService.reset!('hour')` | Reset hourly rate limit counters |
| Expire lapsed polls | `GenericWorker` -> `PollService.expire_lapsed_polls` | Close polls past closing_at |
| Publish closing soon | `GenericWorker` -> `PollService.publish_closing_soon` | Notify users of polls closing soon |
| Send task reminders | `GenericWorker` -> `TaskService.send_task_reminders` | Email task due date reminders |
| Route received emails | `GenericWorker` -> `ReceivedEmailService.route_all` | Process inbound emails |
| Clean expired login tokens | `LoginToken.where(...).delete_all` | Delete tokens >1 hour old |
| GeoLocation updates | `GeoLocationWorker.perform_async` | Update user countries from IPs |
| Daily catch-up emails | `SendDailyCatchUpEmailWorker.perform_async` | Send digests (checks timezone internally) |
| Ensure demo queue | `GenericWorker` -> `DemoService.ensure_queue` | Maintain demo group pool |

**Source**: `lib/tasks/loomio.rake:222-243`

#### Daily Tasks (At Midnight UTC, Hour 0)

| Task | Implementation | Description |
|------|----------------|-------------|
| Throttle reset (daily) | `ThrottleService.reset!('day')` | Reset daily rate limit counters |
| Delete expired demos | `Group.expired_demo.delete_all` | Remove expired demo groups |
| Generate demo groups | `GenericWorker` -> `DemoService.generate_demo_groups` | Create new demo groups |
| Clean orphan records | `GenericWorker` -> `CleanupService.delete_orphan_records` | Database hygiene |
| Publish review due | `GenericWorker` -> `OutcomeService.publish_review_due` | Notify of outcomes needing review |
| Delete old emails | `GenericWorker` -> `ReceivedEmailService.delete_old_emails` | Clean up processed inbound emails |

**Source**: `lib/tasks/loomio.rake:234-241`

**Trigger Condition**: `Time.now.hour == 0`

#### Monthly Tasks (1st of Month at Midnight)

| Task | Implementation | Description |
|------|----------------|-------------|
| Update blocked domains | `UpdateBlockedDomainsWorker.perform_async` | Refresh spam domain blocklist |

**Source**: `lib/tasks/loomio.rake:245-247`

**Trigger Condition**: `Time.now.hour == 0 && Time.now.mday == 1`

#### Weekly Tasks (Sunday, External Trigger)

| Task | Implementation | Description |
|------|----------------|-------------|
| Refresh Chargify links | `GenericWorker` -> `SubscriptionService.refresh_expiring_management_links` | Refresh subscription management URLs |
| Populate Chargify links | `GenericWorker` -> `SubscriptionService.populate_management_links` | Generate missing management URLs |

**Source**: `lib/tasks/loomio.rake:262-273`

**Note**: These tasks check `Date.today.sunday?` internally, so the rake task must be called separately or as part of hourly_tasks on Sundays.

### GenericWorker Pattern

Most scheduled tasks use `GenericWorker` which dynamically invokes service methods:

```ruby
# lib/tasks/loomio.rake
GenericWorker.perform_async('PollService', 'expire_lapsed_polls')
```

```ruby
# app/workers/generic_worker.rb
class GenericWorker
  include Sidekiq::Worker

  def perform(class_name, method_name, arg1 = nil, arg2 = nil, arg3 = nil, arg4 = nil, arg5 = nil)
    class_name.constantize.send(method_name, *([arg1, arg2, arg3, arg4, arg5].compact))
  end
end
```

**Confidence**: HIGH - Explicit implementation found.

### Scheduled Deletion Pattern

Deferred deletions use `perform_at` for scheduled future execution:

```ruby
# app/workers/group_export_worker.rb:12
DestroyRecordWorker.perform_at(1.week.from_now, 'Document', document.id)

# app/services/group_service.rb:125
DestroyGroupWorker.perform_in(2.weeks, group.id)
```

**Confidence**: HIGH - Pattern found in multiple locations.

---

## 5. Active Job Configuration

```ruby
# config/application.rb:26
config.active_job.queue_adapter = :sidekiq
```

**Mailer Queue**:
```ruby
# config/initializers/new_framework_defaults_6_1.rb:63
Rails.application.config.action_mailer.deliver_later_queue_name = :mailers
```

**Confidence**: HIGH - Direct configuration.

---

## 6. Development vs Production

### Development Mode (Default)

```ruby
# config/initializers/sidekiq.rb:11-13
if !Rails.env.production? && !ENV['USE_SIDEKIQ']
  require 'sidekiq/testing'
  Sidekiq::Testing.inline!  # Jobs execute synchronously
end
```

**Effect**: Background jobs run immediately in-process during development unless `USE_SIDEKIQ=true`.

**Confidence**: HIGH - Explicit configuration.

---

## 7. Redis Configuration

| Connection | Environment Variable | Default |
|------------|---------------------|---------|
| Sidekiq queue | `REDIS_QUEUE_URL` or `REDIS_URL` | `redis://localhost:6379/0` |
| Cache/channels | `REDIS_CACHE_URL` or `REDIS_URL` | `redis://localhost:6379/0` |
| Pool size | `REDIS_POOL_SIZE` | 30 |

**Source**: `config/initializers/sidekiq.rb:3-7`

**Confidence**: HIGH - Direct configuration.

---

## Summary

| Finding | Confidence |
|---------|------------|
| Retry defaults (Sidekiq 7 standard: 25 retries) | HIGH |
| 6 workers with retry disabled | HIGH |
| 10 queues defined in priority order | HIGH |
| Queue weighting (order-based, not weighted) | HIGH |
| 38 total workers | HIGH |
| External scheduler pattern (rake task) | HIGH |
| Hourly tasks rake command: `rake loomio:hourly_tasks` | HIGH |
| `login_emails`/`notification_emails` queue usage | MEDIUM (defined but no explicit worker assignment found) |

---

## File References

| File | Lines | Content |
|------|-------|---------|
| `/Users/z/Code/loomio/config/sidekiq.yml` | 1-16 | Queue definitions |
| `/Users/z/Code/loomio/config/initializers/sidekiq.rb` | 1-23 | Sidekiq global config |
| `/Users/z/Code/loomio/lib/tasks/loomio.rake` | 222-275 | Scheduled tasks |
| `/Users/z/Code/loomio/config/application.rb` | 26 | Active Job adapter |
| `/Users/z/Code/loomio/config/initializers/new_framework_defaults_6_1.rb` | 63 | Mailer queue |
| `/Users/z/Code/loomio/Gemfile` | 24 | Sidekiq version |
| `/Users/z/Code/loomio/app/workers/*.rb` | - | All 38 workers |
