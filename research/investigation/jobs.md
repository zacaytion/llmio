# Background Jobs

> Sidekiq workers and scheduled tasks.

## Configuration

**Source:** `orig/loomio/config/sidekiq.yml`

- **Concurrency:** 20 threads (production), 5 (development)
- **Redis:** Uses same Redis as cache/pub-sub

## Queue Priorities

| Priority | Queue | Purpose |
|----------|-------|---------|
| 10 | critical | Time-sensitive operations |
| 6 | high | Important but less urgent |
| 3 | default | Standard processing |
| 1 | low | Background maintenance |
| 1 | mailers | Email delivery |

## Workers Inventory (38)

**Source:** `orig/loomio/app/workers/`

### Email Workers

| Worker | Purpose |
|--------|---------|
| DeliverAnnouncementWorker | Send announcements |
| SendDailyCatchUpEmailWorker | Daily digest |
| SendMorningCatchUpEmailWorker | Morning digest |
| DeliverEmailWorker | Generic email delivery |

### Event Workers

| Worker | Purpose |
|--------|---------|
| EmitEventWorker | Trigger EventBus broadcasts |
| AnnouncementCreatedWorker | Process new announcements |
| PollOutcomeCreatedWorker | Handle outcome creation |
| WebhookWorker | Fire webhook requests |

### Poll Workers

| Worker | Purpose |
|--------|---------|
| PollClosingWorker | Close expired polls |
| PollRemindingWorker | Send vote reminders |
| ProcessStanceWorker | Post-vote processing |

### Group Workers

| Worker | Purpose |
|--------|---------|
| GroupDestroyWorker | Async group deletion |
| GroupExportWorker | Generate data exports |
| LeaveGroupWorker | Handle member departure |

### Maintenance Workers

| Worker | Purpose |
|--------|---------|
| CleanupCacheWorker | Clear old cache entries |
| DiscardOldChatbotsWorker | Remove unused chatbots |
| ScheduledTaskWorker | Run scheduled tasks |
| CalendarInviteWorker | Send calendar invites |

### Search Workers

| Worker | Purpose |
|--------|---------|
| SearchIndexWorker | Update search index |
| ReindexGroupWorker | Full group reindex |

## Scheduled Tasks

**Source:** `orig/loomio/config/schedule.yml`

### Hourly

| Task | Purpose |
|------|---------|
| PollClosingWorker | Close lapsed polls |
| SendRemindersWorker | Vote reminders |
| CleanupWorker | Maintenance tasks |

### Daily (6 AM user timezone)

| Task | Purpose |
|------|---------|
| SendDailyCatchUpEmailWorker | Daily digest emails |
| OutcomeReviewWorker | Review due outcomes |
| TaskDueReminderWorker | Task due reminders |

### Weekly (Monday)

| Task | Purpose |
|------|---------|
| SendWeeklyCatchUpEmailWorker | Weekly digest |
| GroupStatisticsWorker | Usage stats |

### Monthly

| Task | Purpose |
|------|---------|
| SubscriptionWorker | Billing checks |
| CleanupInactiveWorker | Remove stale data |

## Key Worker Implementations

### PollClosingWorker

```ruby
# Finds polls past closing_at, not yet closed
Poll.lapsed.find_each do |poll|
  PollService.close(poll: poll)
end
```

### DeliverAnnouncementWorker

```ruby
def perform(event_id, actor_id, user_ids)
  event = Event.find(event_id)
  users = User.where(id: user_ids)

  users.each do |user|
    Notification.create!(user: user, event: event)
    UserMailer.event(user, event).deliver_later
  end
end
```

### WebhookWorker

```ruby
def perform(webhook_id, event_id)
  webhook = Webhook.find(webhook_id)
  event = Event.find(event_id)

  response = HTTP.post(webhook.url, json: payload(webhook, event))
  # Handle response, retry on failure
end
```

## Go Implementation Notes

**Options:**
- **Asynq:** Redis-based, similar to Sidekiq
- **River:** PostgreSQL-based job queue
- **Temporal:** For complex workflows

**Key Patterns:**
1. Retry with exponential backoff
2. Dead letter queue for failed jobs
3. Scheduled jobs via cron-like scheduler
4. Priority queues

---
