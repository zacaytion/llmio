# Missing Features - Follow-up Analysis

## Executive Summary

Several major feature areas documented in baseline research **were not included in the third-party follow-up investigation**. This document consolidates questions for these omitted topics.

---

## 1. Email/Mailer System

### Baseline Coverage
Source: `research/initial_investigation_review.md` Section 3.1

**7 Mailers Identified:**
- `BaseMailer` - Common email functionality
- `UserMailer` - User-related emails
- `EventMailer` - Event notifications
- `GroupMailer` - Group-related emails
- `ContactMailer` - Contact form emails
- `TaskMailer` - Task reminders
- `ForwardMailer` - Email forwarding

**Catch-up Email System:**
- Daily/weekly/bi-weekly frequency options
- Runs at 6 AM in user's timezone
- Email bounce throttling (1 per hour)

**Special Email Formats:**
- Threading: `d=100&k=key&u=999@mail.loomio.com`
- New discussions: `group+u=99&k=key@mail.loomio.com`

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What email service is used? (SendGrid, SES, SMTP?) | HIGH |
| 2 | How is email bounce tracking implemented? | HIGH |
| 3 | What is the catch-up email content structure? | MEDIUM |
| 4 | How does email threading work with different clients? | MEDIUM |
| 5 | Is there email template versioning? | LOW |

---

## 2. File Storage Backends

### Baseline Coverage
Source: `research/initial_investigation_review.md` Section 3.5

**5 Storage Options:**
1. Disk (test/local development)
2. Amazon S3
3. DigitalOcean Spaces
4. S3-compatible (generic)
5. Google Cloud Storage

**ActiveStorage Integration:**
- Blob metadata storage
- Variant records for image processing
- Attachment associations on models

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What are the environment variables for each backend? | HIGH |
| 2 | Is there image processing (resizing, thumbnails)? | HIGH |
| 3 | How are signed URLs generated? | MEDIUM |
| 4 | What is the max file size limit? | MEDIUM |
| 5 | Is there virus scanning? | LOW |

---

## 3. Demo System / RecordCloner

### Baseline Coverage
Source: `research/initial_investigation_review.md` Section 3.3

**DemoService Methods:**
- `refill_queue` - Refill demo group queue
- `take_demo(actor)` - Assign demo to user
- `ensure_queue` - Ensure queue has demos
- `generate_demo_groups` - Create demo groups

**RecordCloner Service:**
Located at `app/services/record_cloner.rb`

```ruby
def create_clone_group_for_public_demo(group, handle)
  # Deep clones group with:
  # - Discussions
  # - Polls
  # - Stances
  # - Tags
  # - Permissions reset for demo mode
end
```

**Redis Queue:**
- Key: `demo_group_ids`
- Size: `ENV['FEATURES_DEMO_GROUPS_SIZE']` (default: 3)

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | Which associations are cloned with a group? | HIGH |
| 2 | How are IDs remapped during cloning? | HIGH |
| 3 | Is demo data anonymized? | MEDIUM |
| 4 | How long do demo groups persist? | MEDIUM |
| 5 | Can users convert demo to real group? | LOW |

---

## 4. Subscription/Billing System

### Baseline Coverage
Source: `research/initial_investigation_review.md` Section 3.4

**Plan Types:**
- `free` - Free tier
- `demo` - Demo/trial
- `was-gift`, `was-paid` - Legacy consolidated plans

**States:**
- `active`, `on_hold`, `pending`, `past_due`, `trialing`, `canceled`

**Payment Methods:**
- `chargify` - Primary billing
- `manual` - Manual invoicing
- `barter` - Exchange arrangements
- `paypal` - PayPal payments
- `none` - No payment

**Feature Flags:**
- `max_members` - Member limit
- `max_threads` - Thread limit
- `max_orgs` - Organization limit
- `allow_subgroups` - Subgroup feature
- `allow_guests` - Guest access

**Chargify Integration:**
- `info` JSONB field stores `chargify_management_link`

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What triggers subscription state transitions? | HIGH |
| 2 | How are feature limits enforced? | HIGH |
| 3 | Is there metered billing (pay per use)? | MEDIUM |
| 4 | How is Chargify webhook handled? | MEDIUM |
| 5 | Are there enterprise/custom plans? | LOW |

---

## 5. Tasks System

### Baseline Coverage
Source: `research/loomio_initial_investigation.md` Section 7

**TaskService Methods:**
- `send_task_reminders` - Send due reminders
- `mark_as_done` - Mark task complete
- `mark_as_not_done` - Reopen task
- `update_done` - Update status

**Task Model:**
- Polymorphic association to `record_type` (Discussion, Comment)
- `tasks_users` many-to-many for assignees

**Known Issue:**
- Tasks are **hard-deleted** without notification (TODO not fixed)

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What triggers task creation? | HIGH |
| 2 | How are task due dates set? | MEDIUM |
| 3 | Is there task recurrence? | LOW |
| 4 | Should Go implement soft delete? | MEDIUM |

---

## 6. Mention System

### Baseline Coverage
Source: `research/loomio_initial_investigation.md` Section 4

**Events:**
- `user_mentioned` - @user mentions
- `group_mentioned` - @group mentions

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | How are mentions parsed from content? | HIGH |
| 2 | What is the mention syntax? (@username? @[User Name]?) | HIGH |
| 3 | Are mentions resolved at write time or render time? | MEDIUM |
| 4 | Can you mention users outside the group? | MEDIUM |

---

## 7. Translation System

### Baseline Coverage
Source: `research/initial_investigation_review.md` Section 4.5

**Translations Table:**
- `translatable_type` - Polymorphic type
- `translatable_id` - Polymorphic ID
- `fields` - hstore of translated fields

**TranslationService:**
- Implementation unknown
- Auto-translation support unknown

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What translation service is used? (Google, DeepL?) | HIGH |
| 2 | Is translation automatic or on-demand? | HIGH |
| 3 | Which fields are translatable? | MEDIUM |
| 4 | Is there a translation cache? | LOW |

---

## 8. Blocked Domains System

### Baseline Coverage
Source: `research/loomio_initial_investigation.md` Section 7

**Worker:**
- `UpdateBlockedDomainsWorker` - Monthly job

**Table:**
- `blocked_domains` - Email domain blocklist

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What is the source of the blocked domains list? | HIGH |
| 2 | Is blocking applied at registration only or all emails? | MEDIUM |
| 3 | Can admins add custom blocked domains? | LOW |

---

## 9. Link Preview Service

### Baseline Coverage
Source: `research/initial_investigation_review.md` Section 3.6

**LinkPreviewService:**
- Fetches URL metadata (title, description, image, hostname)
- Triggered when URLs pasted into content

**JSONB Structure:**
```json
[{
  "title": "...",
  "url": "...",
  "description": "...",
  "image": "...",
  "fit": "...",
  "align": "...",
  "hostname": "..."
}]
```

**Limits:**
- Title and description max 240 chars

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | What service fetches previews? (Embed.ly? Custom?) | MEDIUM |
| 2 | Is there preview caching? | LOW |
| 3 | How are failed fetches handled? | LOW |

---

## 10. Data Integrity Issues (FIXME/TODO)

### Baseline Coverage
Source: `research/investigation/questions.md`

**Known Issues:**

1. **Guest Boolean Migration Incomplete**
   - Existing guest records have `guest = false` when should be `true`
   - Affects `discussion_readers` and `stances` tables

2. **Reaction Uniqueness Not Enforced**
   - Duplicate reactions from same user possible

3. **Task Deletion Without Notification**
   - Tasks are hard-deleted, not soft-deleted

4. **Anonymous Poll Email Display Incomplete**
   - Email rendering for anon polls may be broken

### Open Questions

| # | Question | Priority |
|---|----------|----------|
| 1 | Should Go fix these issues or preserve bugs for compatibility? | HIGH |
| 2 | Is there a data migration strategy for guest boolean? | HIGH |
| 3 | Should Go add unique constraint on reactions? | MEDIUM |

---

## Priority Summary

| Feature | Priority | Blocking Go Implementation? |
|---------|----------|----------------------------|
| Email System | HIGH | Yes - core notification path |
| File Storage | HIGH | Yes - user uploads |
| Demo/Cloner | MEDIUM | No - feature can be deferred |
| Subscriptions | MEDIUM | Yes if monetization required |
| Tasks | MEDIUM | No - minor feature |
| Mentions | MEDIUM | Yes - affects content rendering |
| Translations | LOW | No - can defer |
| Blocked Domains | LOW | No - can defer |
| Link Previews | LOW | No - can defer |
| Data Integrity | HIGH | Decision needed |
