# Email System - Loomio Reverse Engineering

## Overview

Loomio uses Rails Action Mailer with configurable SMTP for outbound email delivery, and Rails Action Mailbox for inbound email processing (reply-by-email functionality). The system includes bounce handling, spam complaint tracking, and digest emails.

---

## 1. Email Service Configuration

**Confidence: HIGH**

### Outbound Email (SMTP)

Loomio uses standard SMTP delivery configured via environment variables. There is no explicit integration with third-party services like SendGrid or AWS SES at the application level - it's provider-agnostic SMTP.

**Configuration Location**: `config/application.rb:61-75`

```ruby
if ENV['SMTP_SERVER']
  config.action_mailer.delivery_method = :smtp
  config.action_mailer.smtp_settings = {
    address: ENV['SMTP_SERVER'],
    port: ENV['SMTP_PORT'],
    authentication: ENV['SMTP_AUTH'],
    user_name: ENV['SMTP_USERNAME'],
    password: ENV['SMTP_PASSWORD'],
    domain: ENV['SMTP_DOMAIN'],
    ssl: ENV['SMTP_USE_SSL'].present?,
    openssl_verify_mode: ENV.fetch('SMTP_SSL_VERIFY_MODE', 'none')
  }.compact
else
  config.action_mailer.delivery_method = :test
end
```

### Environment Variables Required

| Variable | Purpose |
|----------|---------|
| `SMTP_SERVER` | SMTP server address |
| `SMTP_PORT` | SMTP port |
| `SMTP_AUTH` | Authentication type |
| `SMTP_USERNAME` | SMTP username |
| `SMTP_PASSWORD` | SMTP password |
| `SMTP_DOMAIN` | Domain for HELO |
| `SMTP_USE_SSL` | Enable SSL (presence check) |
| `SMTP_SSL_VERIFY_MODE` | SSL verification mode (default: 'none') |
| `NOTIFICATIONS_EMAIL_ADDRESS` | From address (default: `notifications@#{SMTP_DOMAIN}`) |

### Notification Email Address

**File**: `app/mailers/base_mailer.rb:10-11`

```ruby
NOTIFICATIONS_EMAIL_ADDRESS = ENV.fetch('NOTIFICATIONS_EMAIL_ADDRESS', "notifications@#{ENV['SMTP_DOMAIN']}")
default :from => "\"#{AppConfig.theme[:site_name]}\" <#{NOTIFICATIONS_EMAIL_ADDRESS}>"
```

### Inbound Email (Action Mailbox)

**File**: `config/application.rb:56`

```ruby
config.action_mailbox.ingress = :relay
```

Loomio uses the `:relay` ingress for Action Mailbox, meaning a mail relay (like Postfix or an MTA) forwards inbound email to a Rails endpoint.

---

## 2. Bounce Tracking Mechanism

**Confidence: HIGH**

### Bounce Detection Strategy

Loomio implements bounce handling through **email-based feedback loops**, not webhook integrations. Bounces are detected when:

1. Users reply to the non-replyable notifications address
2. Email providers forward spam complaints (Amazon SES format)

### Bounce Email Throttling

**File**: `app/services/received_email_service.rb:33-42`

When someone replies to the notifications address (indicating a mis-reply or bounce), the system sends a delivery failure notice, throttled to **1 per hour per sender**:

```ruby
if email.sent_to_notifications_address?
  if ThrottleService.can?(key: 'bounce', id: email.sender_email.downcase, max: 1, per: 'hour')
    Rails.logger.info("email bounced");
    ForwardMailer.bounce(to: email.sender_name_and_email).deliver_now
  else
    Rails.logger.info("bounce throttled for #{email.sender_email}");
  end
  return email.destroy
end
```

### Throttle Service Implementation

**File**: `app/services/throttle_service.rb:11-16`

```ruby
def self.can?(key: 'default-key', id: 1, max: 100, inc: 1, per: 'hour')
  raise "Throttle per is not hour or day: #{per}" unless ['hour', 'day'].include? per.to_s
  k = "THROTTLE-#{per.upcase}-#{key}-#{id}"
  Redis::Counter.new(k).increment(inc)
  Redis::Counter.new(k).value <= ENV.fetch('THROTTLE_MAX_'+key, max)
end
```

Uses Redis counters with hourly or daily TTL. The throttle limit can be overridden via `ENV['THROTTLE_MAX_bounce']`.

### Spam Complaint Tracking

**File**: `app/services/received_email_service.rb:44-50`

Spam complaints (from AWS SES abuse feedback loop) increment a counter on the user record:

```ruby
if email.is_complaint? && email.complainer_address.present?
  Rails.logger.info("complaint email recieved from #{email.complainer_address}");
  User.where(email: email.complainer_address).update_all("complaints_count = complaints_count + 1")
  email.update(released: true)
  return
end
```

**Complaint Detection**: `app/models/received_email.rb:119-121`

```ruby
def is_complaint?
  sender_email == ENV.fetch('COMPLAINTS_ADDRESS', "complaints@email-abuse.amazonses.com")
end
```

### User Complaint Count (Schema)

**File**: `db/schema.rb:1049`

```ruby
t.integer "complaints_count", default: 0, null: false
```

### Email Suppression Based on Complaints

**File**: `app/models/user.rb:116-117`

```ruby
scope :no_spam_complaints, -> { where(complaints_count: 0) }
scope :has_spam_complaints, -> { where("complaints_count > 0") }
```

**File**: `app/mailers/base_mailer.rb:30`

```ruby
return if User.has_spam_complaints.where(email: to).exists?
```

**File**: `app/models/concerns/events/notify/by_email.rb:9`

```ruby
email_recipients.active.no_spam_complaints.uniq.pluck(:id).each do |recipient_id|
```

Users with any spam complaints (`complaints_count > 0`) are excluded from all email delivery.

---

## 3. Catch-up Email Structure

**Confidence: HIGH**

### Schedule

**File**: `app/workers/send_daily_catch_up_email_worker.rb:6-23`

Catch-up emails are sent at **6 AM in the user's timezone**:

```ruby
if time_in_zone.hour == 6
  days = [7, time_in_zone.wday, (time_in_zone.wday % 2 == 1) ? 8 : nil].compact
  User.distinct.active.verified.where(time_zone: zone).where(email_catch_up_day: days).find_each do |user|
    period = case user.email_catch_up_day
      when 8 then 'other'
      when 7 then 'daily'
      else 'weekly'
    end
    UserMailer.catch_up(user.id, nil, period).deliver_now
  end
end
```

### Frequency Options

| `email_catch_up_day` | Period | Meaning |
|---------------------|--------|---------|
| 7 | daily | Every day at 6 AM |
| 8 | other | Every other day (odd wday) |
| 0-6 | weekly | Specific day of week (0=Sunday) |

### Content Selection

**File**: `app/mailers/user_mailer.rb:28-62`

The catch-up email includes:

1. **Time window**: Based on frequency (24h, 48h, or 1 week)
2. **Discussions**: Unread discussions with activity in the time window
3. **Groups**: User's groups, ordered by full_name

```ruby
@discussions = DiscussionQuery.visible_to(
  user: user,
  only_unread: true,
  or_public: false,
  or_subgroups: false
).kept.last_activity_after(@time_start)
```

### Email Template Structure

**File**: `app/views/user_mailer/catch_up.html.haml`

The email is structured as:

1. "Do not reply" notice
2. Subject/headline
3. **Table of Contents** (group headlines with discussion links)
4. **Direct discussions section** (if any)
5. **Per-group sections**, each containing:
   - Group name with link
   - Discussion details

### Per-Discussion Content

**File**: `app/views/user_mailer/catch_up/_discussion.html.haml:1-22`

Each discussion includes:

1. **Title** (linked)
2. **Author name** (if new discussion)
3. **Description** (if new discussion)
4. **Active/recently closed polls** with:
   - Poll title and tags
   - Poll summary
   - Vote button
   - Results
5. **Activity feed** (since last read):
   - New comments
   - New votes / stance created
   - Discussion closed/edited
   - Poll edited
6. **Reply link**

### Read Tracking Pixel

**File**: `app/views/user_mailer/catch_up.html.haml:21`

```haml
%img{src: mark_summary_as_read_url_for(@user, format: 'gif'), alt: '', width: 1, height: 1}
```

A 1x1 tracking pixel marks the summary as read when the email is opened.

---

## 4. Email Threading Format (Reply-To Address)

**Confidence: HIGH**

### Reply-To Address Structure

**File**: `app/helpers/email_helper.rb:91-107`

```ruby
def reply_to_address(model:, user:)
  letter = {
    'Comment' => 'c',
    'Poll' => 'p',
    'Stance' => 's',
    'Outcome' => 'o'
  }[model.class.to_s]

  address = {
    pt: letter,
    pi: letter ? model.id : nil,
    d: model.discussion_id,
    u: user.id,
    k: user.email_api_key
  }.compact.map { |k, v| [k,v].join('=') }.join('&')
  [address, ENV['REPLY_HOSTNAME']].join('@')
end
```

### Address Format Examples

| Context | Format | Example |
|---------|--------|---------|
| Discussion reply | `d={discussion_id}&u={user_id}&k={api_key}@{hostname}` | `d=100&u=999&k=abc123@mail.loomio.com` |
| Comment reply | `pt=c&pi={comment_id}&d={discussion_id}&u={user_id}&k={api_key}@{hostname}` | `pt=c&pi=50&d=100&u=999&k=abc123@mail.loomio.com` |
| Poll reply | `pt=p&pi={poll_id}&d={discussion_id}&u={user_id}&k={api_key}@{hostname}` | `pt=p&pi=75&d=100&u=999&k=abc123@mail.loomio.com` |

### Parameters

| Key | Meaning |
|-----|---------|
| `d` | Discussion ID |
| `u` | User ID |
| `k` | User's email_api_key (authentication token) |
| `pt` | Parent type (c=Comment, p=Poll, s=Stance, o=Outcome) |
| `pi` | Parent ID |

### Inbound Processing

**File**: `app/services/received_email_service.rb:62-76`

```ruby
case email.route_path
when /d=.+&u=.+&k=.+/
  # personal email-to-thread
  CommentService.create(comment: Comment.new(comment_params(email)), actor: actor_from_email(email))
when /[^\s]+\+u=.+&k=.+/
  # personal email-to-group (creating discussion)
  DiscussionService.create(discussion: Discussion.new(discussion_params(email)), actor: actor_from_email(email))
end
```

### User Authentication from Email

**File**: `app/services/received_email_service.rb:172-175`

```ruby
def self.actor_from_email(email)
  params = parse_route_params(email.route_path)
  User.find_by!(id: params['u'], email_api_key: params['k'])
end
```

The `email_api_key` is a per-user token stored on the user record (`db/schema.rb:1015`), providing authentication for reply-by-email.

---

## 5. Mailer Hierarchy

**Confidence: HIGH**

### Mailer Classes

| Mailer | Parent | Purpose |
|--------|--------|---------|
| `BaseMailer` | `ActionMailer::Base` | Common functionality, spam filtering |
| `UserMailer` | `BaseMailer` | User-specific emails (catch-up, login, etc.) |
| `EventMailer` | `BaseMailer` | Event notifications (comments, polls, etc.) |
| `GroupMailer` | `BaseMailer` | Group-related emails (destroy warning) |
| `TaskMailer` | `BaseMailer` | Task reminders |
| `ContactMailer` | `ActionMailer::Base` | Contact form (to support) |
| `ForwardMailer` | `ActionMailer::Base` | Email forwarding & bounce notices |

### BaseMailer Features

**File**: `app/mailers/base_mailer.rb`

1. **Default from address**: Site name + notifications email
2. **UTM tracking**: Adds `utm_medium=email` and `utm_campaign={action_name}`
3. **Spam filtering**: Blocks sends to:
   - Addresses matching `NoSpam::SPAM_REGEX`
   - The notifications address itself
   - Users with spam complaints
4. **Localization**: Respects user locale preference
5. **Error handling**: Catches SMTP errors

---

## Environment Variables Summary

| Variable | Purpose | Default |
|----------|---------|---------|
| `SMTP_SERVER` | SMTP host | (required for delivery) |
| `SMTP_PORT` | SMTP port | - |
| `SMTP_AUTH` | Auth type | - |
| `SMTP_USERNAME` | SMTP user | - |
| `SMTP_PASSWORD` | SMTP password | - |
| `SMTP_DOMAIN` | HELO domain | - |
| `SMTP_USE_SSL` | Enable SSL | false |
| `SMTP_SSL_VERIFY_MODE` | SSL verification | none |
| `NOTIFICATIONS_EMAIL_ADDRESS` | From address | `notifications@{SMTP_DOMAIN}` |
| `REPLY_HOSTNAME` | Reply-to domain | - |
| `OLD_REPLY_HOSTNAME` | Legacy reply domain | - |
| `COMPLAINTS_ADDRESS` | SES complaint sender | `complaints@email-abuse.amazonses.com` |
| `SUPPORT_EMAIL` | Support contact | - |
| `THROTTLE_MAX_bounce` | Override bounce throttle | 1 |

---

## Key File References

| File | Line | Purpose |
|------|------|---------|
| `config/application.rb` | 56-75 | Mail delivery configuration |
| `app/mailers/base_mailer.rb` | 10-42 | Base mailer with spam filtering |
| `app/mailers/user_mailer.rb` | 28-63 | Catch-up email |
| `app/mailers/event_mailer.rb` | 1-112 | Event notifications |
| `app/mailers/forward_mailer.rb` | 20-29 | Bounce notice |
| `app/helpers/email_helper.rb` | 91-107 | Reply-to address format |
| `app/services/received_email_service.rb` | 27-116 | Inbound email routing |
| `app/services/throttle_service.rb` | 1-25 | Redis-based throttling |
| `app/workers/send_daily_catch_up_email_worker.rb` | 1-24 | Catch-up email scheduling |
| `app/models/received_email.rb` | 119-128 | Complaint detection |
| `app/models/user.rb` | 116-117 | Complaint scopes |
| `db/schema.rb` | 1049 | complaints_count column |
