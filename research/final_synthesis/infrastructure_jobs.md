# Infrastructure & Background Jobs - Implementation Synthesis

## Executive Summary

Loomio uses Sidekiq with 38 workers across 5 queue priorities. Go will use River (PostgreSQL-based job queue per CLAUDE.md approved stack) to implement equivalent functionality.

---

## Confirmed Configuration

### Queue Priorities

| Priority | Queue | Purpose | River Config |
|----------|-------|---------|--------------|
| 10 | critical | Time-sensitive operations | `MaxWorkers: 10` |
| 6 | high | Important but less urgent | `MaxWorkers: 5` |
| 3 | default | Standard processing | `MaxWorkers: 20` |
| 1 | low | Background maintenance | `MaxWorkers: 2` |
| 1 | mailers | Email delivery | `MaxWorkers: 5` |

### Sidekiq Concurrency

- **Production:** 20 threads
- **Development:** 5 threads

---

## Worker Inventory (38)

### Core Workers (Implement First)

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `publish_event_worker.rb` | Event publishing | default | `PublishEventJob` |
| `generic_worker.rb` | Generic job wrapper | default | N/A (River handles) |
| `close_expired_poll_worker.rb` | Close lapsed polls | default | `CloseExpiredPollsJob` |
| `send_daily_catch_up_email_worker.rb` | Daily digest | mailers | `SendDailyCatchUpJob` |

### Notification Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `announce_discussion_worker.rb` | Discussion announcements | high | `AnnounceDiscussionJob` |

### Poll Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `update_poll_counts_worker.rb` | Counter cache update | default | `UpdatePollCountsJob` |
| `reset_poll_stance_data_worker.rb` | Reset poll data | low | `ResetPollStanceDataJob` |

### Cleanup Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `destroy_discussion_worker.rb` | Async discussion delete | low | `DestroyDiscussionJob` |
| `destroy_group_worker.rb` | Async group delete | low | `DestroyGroupJob` |
| `destroy_record_worker.rb` | Generic record delete | low | `DestroyRecordJob` |
| `destroy_tag_worker.rb` | Tag deletion | low | `DestroyTagJob` |
| `destroy_user_worker.rb` | User deletion | low | `DestroyUserJob` |

### User Management Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `deactivate_user_worker.rb` | User deactivation | default | `DeactivateUserJob` |
| `redact_user_worker.rb` | GDPR data redaction | low | `RedactUserJob` |
| `geo_location_worker.rb` | Geolocation lookup | low | `GeoLocationJob` |

### Export Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `group_export_worker.rb` | Full export | low | `GroupExportJob` |
| `group_export_csv_worker.rb` | CSV export | low | `GroupExportCSVJob` |

### Content Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `move_comments_worker.rb` | Move comments | default | `MoveCommentsJob` |
| `append_transcript_worker.rb` | Append transcripts | default | `AppendTranscriptJob` |
| `attach_document_worker.rb` | Attach documents | default | `AttachDocumentJob` |
| `download_attachment_worker.rb` | Download attachments | low | `DownloadAttachmentJob` |
| `add_heading_ids_worker.rb` | Content processing | low | `AddHeadingIDsJob` |
| `undelete_blob_worker.rb` | Restore deleted files | low | `UndeleteBlobJob` |
| `update_tag_worker.rb` | Tag updates | default | `UpdateTagJob` |

### Membership Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `accept_membership_worker.rb` | Accept membership | high | `AcceptMembershipJob` |
| `revoke_memberships_of_deactivated_users_worker.rb` | Cleanup | low | `RevokeMembershipsJob` |

### Maintenance Workers

| Worker | Purpose | Queue | Go Job |
|--------|---------|-------|--------|
| `repair_thread_worker.rb` | Thread repair | low | `RepairThreadJob` |
| `update_blocked_domains_worker.rb` | Update blocklist | low | `UpdateBlockedDomainsJob` |

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

| Task | Purpose | Go Schedule |
|------|---------|-------------|
| Close expired polls | Close lapsed polls | `@hourly` |
| Send reminders | Vote reminders | `@hourly` |

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

## Go Implementation

### River Configuration

```go
package jobs

import (
    "context"
    "time"

    "github.com/riverqueue/river"
    "github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func NewRiverClient(pool *pgxpool.Pool) (*river.Client[pgx.Tx], error) {
    workers := river.NewWorkers()

    // Register all workers
    river.AddWorker(workers, &CloseExpiredPollsWorker{})
    river.AddWorker(workers, &PublishEventWorker{})
    river.AddWorker(workers, &SendDailyCatchUpWorker{})
    river.AddWorker(workers, &DestroyDiscussionWorker{})
    // ... register all workers

    return river.NewClient(riverpgxv5.New(pool), &river.Config{
        Queues: map[string]river.QueueConfig{
            "critical": {MaxWorkers: 10},
            "high":     {MaxWorkers: 5},
            "default":  {MaxWorkers: 20},
            "low":      {MaxWorkers: 2},
            "mailers":  {MaxWorkers: 5},
        },
        Workers: workers,
    })
}
```

### Job Definitions

```go
// CloseExpiredPollsJob - Hourly scheduled task
type CloseExpiredPollsJob struct{}

func (CloseExpiredPollsJob) Kind() string { return "close_expired_polls" }

type CloseExpiredPollsWorker struct {
    river.WorkerDefaults[CloseExpiredPollsJob]
    pollService *PollService
}

func (w *CloseExpiredPollsWorker) Work(ctx context.Context, job *river.Job[CloseExpiredPollsJob]) error {
    polls, err := w.pollService.FindLapsed(ctx)
    if err != nil {
        return err
    }

    for _, poll := range polls {
        if err := w.pollService.Close(ctx, poll); err != nil {
            slog.Error("failed to close poll", "poll_id", poll.ID, "error", err)
            // Continue with other polls
        }
    }

    return nil
}
```

```go
// PublishEventJob - Event notification delivery
type PublishEventJob struct {
    EventID int64 `json:"event_id"`
    ActorID int64 `json:"actor_id"`
    UserIDs []int64 `json:"user_ids"`
}

func (PublishEventJob) Kind() string { return "publish_event" }

type PublishEventWorker struct {
    river.WorkerDefaults[PublishEventJob]
    eventService *EventService
}

func (w *PublishEventWorker) Work(ctx context.Context, job *river.Job[PublishEventJob]) error {
    event, err := w.eventService.Find(ctx, job.Args.EventID)
    if err != nil {
        return err
    }

    return w.eventService.Deliver(ctx, event, job.Args.ActorID, job.Args.UserIDs)
}
```

```go
// DestroyDiscussionJob - Async deletion
type DestroyDiscussionJob struct {
    DiscussionID int64 `json:"discussion_id"`
}

func (DestroyDiscussionJob) Kind() string { return "destroy_discussion" }

type DestroyDiscussionWorker struct {
    river.WorkerDefaults[DestroyDiscussionJob]
    discussionService *DiscussionService
}

func (w *DestroyDiscussionWorker) Work(ctx context.Context, job *river.Job[DestroyDiscussionJob]) error {
    return w.discussionService.HardDelete(ctx, job.Args.DiscussionID)
}
```

### Scheduled Tasks with River

```go
func SetupPeriodicJobs(client *river.Client[pgx.Tx]) {
    // Hourly: Close expired polls
    client.PeriodicJobs().Add(&river.PeriodicJob{
        PeriodicInterval: time.Hour,
        Constructors: func() (river.JobArgs, *river.InsertOpts) {
            return CloseExpiredPollsJob{}, &river.InsertOpts{Queue: "default"}
        },
    })

    // Daily: Update blocked domains (at midnight UTC)
    client.PeriodicJobs().Add(&river.PeriodicJob{
        PeriodicInterval: 24 * time.Hour,
        Constructors: func() (river.JobArgs, *river.InsertOpts) {
            return UpdateBlockedDomainsJob{}, &river.InsertOpts{Queue: "low"}
        },
    })
}
```

### User Timezone-Aware Daily Jobs

```go
// Daily catch-up emails need per-user timezone handling
type SendDailyCatchUpJob struct {
    UserID   int64     `json:"user_id"`
    Timezone string    `json:"timezone"`
    SendAt   time.Time `json:"send_at"`
}

func (SendDailyCatchUpJob) Kind() string { return "send_daily_catch_up" }

// Schedule catch-up emails at 6 AM in each user's timezone
func ScheduleDailyCatchUpEmails(ctx context.Context, client *river.Client[pgx.Tx], userRepo *UserRepository) error {
    users, err := userRepo.FindWithCatchUpEnabled(ctx)
    if err != nil {
        return err
    }

    for _, user := range users {
        loc, err := time.LoadLocation(user.Timezone)
        if err != nil {
            loc = time.UTC
        }

        // Calculate next 6 AM in user's timezone
        now := time.Now().In(loc)
        sendAt := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, loc)
        if sendAt.Before(now) {
            sendAt = sendAt.Add(24 * time.Hour)
        }

        _, err = client.InsertTx(ctx, tx, SendDailyCatchUpJob{
            UserID:   user.ID,
            Timezone: user.Timezone,
            SendAt:   sendAt,
        }, &river.InsertOpts{
            Queue:       "mailers",
            ScheduledAt: sendAt,
        })
        if err != nil {
            slog.Error("failed to schedule catch-up", "user_id", user.ID, "error", err)
        }
    }

    return nil
}
```

---

## Retry Configuration

### River Defaults (match Sidekiq)

```go
// Default: 25 retries with exponential backoff
type PublishEventWorker struct {
    river.WorkerDefaults[PublishEventJob]
}

func (w *PublishEventWorker) Timeout(job *river.Job[PublishEventJob]) time.Duration {
    return 5 * time.Minute
}

func (w *PublishEventWorker) MaxRetries() int {
    return 25
}
```

### Custom Retry for Critical Jobs

```go
type CloseExpiredPollsWorker struct {
    river.WorkerDefaults[CloseExpiredPollsJob]
}

// Critical jobs retry faster
func (w *CloseExpiredPollsWorker) NextRetry(job *river.Job[CloseExpiredPollsJob]) time.Time {
    // Retry in 30 seconds, 1 min, 2 min, 5 min...
    delays := []time.Duration{30 * time.Second, time.Minute, 2 * time.Minute, 5 * time.Minute}
    attempt := min(job.Attempt-1, len(delays)-1)
    return time.Now().Add(delays[attempt])
}
```

---

## Job Enqueueing

### From Services

```go
// In discussion_service.go
func (s *DiscussionService) Announce(ctx context.Context, discussion *Discussion, actor *User, userIDs []int64) error {
    // Create announcement event
    event := &Event{
        Kind:        "discussion_announced",
        EventableID: discussion.ID,
        ActorID:     actor.ID,
    }
    s.eventRepo.Create(ctx, event)

    // Enqueue notification delivery
    _, err := s.riverClient.Insert(ctx, PublishEventJob{
        EventID: event.ID,
        ActorID: actor.ID,
        UserIDs: userIDs,
    }, &river.InsertOpts{
        Queue: "high",
    })

    return err
}
```

### Async Deletion

```go
func (s *DiscussionService) Delete(ctx context.Context, discussion *Discussion) error {
    // Soft delete immediately
    discussion.DiscardedAt = time.Now()
    s.discussionRepo.Update(ctx, discussion)

    // Schedule hard delete
    _, err := s.riverClient.Insert(ctx, DestroyDiscussionJob{
        DiscussionID: discussion.ID,
    }, &river.InsertOpts{
        Queue:       "low",
        ScheduledAt: time.Now().Add(30 * 24 * time.Hour), // 30 days later
    })

    return err
}
```

---

## Monitoring

### River UI

River provides a built-in UI at `/river` for job monitoring.

```go
riverUI := riverui.NewServer(riverClient, &riverui.ServerConfig{
    Prefix: "/admin/river",
})
r.Mount("/admin/river", riverUI)
```

### Metrics

```go
// Track job counts by queue and status
type JobMetrics struct {
    Pending   int64
    Running   int64
    Completed int64
    Failed    int64
}

func GetJobMetrics(ctx context.Context, pool *pgxpool.Pool) (*JobMetrics, error) {
    var m JobMetrics
    err := pool.QueryRow(ctx, `
        SELECT
            COUNT(*) FILTER (WHERE state = 'available') as pending,
            COUNT(*) FILTER (WHERE state = 'running') as running,
            COUNT(*) FILTER (WHERE state = 'completed') as completed,
            COUNT(*) FILTER (WHERE state = 'discarded') as failed
        FROM river_job
    `).Scan(&m.Pending, &m.Running, &m.Completed, &m.Failed)
    return &m, err
}
```

---

## Testing

```go
func TestCloseExpiredPolls(t *testing.T) {
    ctx := context.Background()

    // Create expired poll
    poll := &Poll{
        ClosingAt: time.Now().Add(-time.Hour),
        ClosedAt:  nil,
    }
    pollRepo.Create(ctx, poll)

    // Run worker
    worker := &CloseExpiredPollsWorker{pollService: pollService}
    err := worker.Work(ctx, &river.Job[CloseExpiredPollsJob]{Args: CloseExpiredPollsJob{}})
    require.NoError(t, err)

    // Verify poll is closed
    poll, _ = pollRepo.Find(ctx, poll.ID)
    assert.NotNil(t, poll.ClosedAt)
}
```
