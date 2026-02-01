# Integrations Domain: Tests

**Generated:** 2026-02-01
**Confidence Rating:** 3/5 (Limited test coverage found for chatbots)

---

## Overview

The integrations domain has comprehensive test coverage for email processing but limited explicit test coverage for chatbots. Tests are located in the standard Rails test directories.

---

## 1. ReceivedEmailService Tests

**Location:** `/spec/services/received_email_service_spec.rb`

### Reply Body Extraction

Tests the `extract_reply_body` method that strips quoted content from email replies.

**Test Cases:**

| Scenario | Input Pattern | Expected Behavior |
|----------|---------------|-------------------|
| Spanish text reply | "Me aparece...El lun, 3 de jun...escribió:" | Strips Spanish "wrote" indicator |
| Spanish HTML reply | Content with "escribiÃ³:" | Handles HTML-encoded Spanish |
| Joshua's reply | "Yep...On Tue, 7 Mar...wrote:" | Strips English date+wrote pattern |
| Loomio address pattern | "On someday (Loomio) notifications@... said:" | Strips Loomio-specific pattern |
| Generic wrote pattern | "On someday bobo@... wrote:" | Strips generic "wrote" pattern |
| Hidden delimiter | Content with EventMailer::REPLY_DELIMITER | Strips hidden delimiter chars |
| Signature delimiter | Content with "--" separator | Strips signature block |
| Author name signature | Content starting with author name | Strips signature starting with author name |

### Subject Line Stripping

Tests regex for removing Re:/Fwd: prefixes from subjects.

**Test Cases:**
- "Re: repairing..." -> "repairing..."
- "Fwd: repairing..." -> "repairing..."
- "RE: FW: repairing..." -> "repairing..."
- Multiple nested Re/Fwd combinations

### Notifications Address Routing

Tests handling of emails sent to the notifications "From" address.

**Test Cases:**

| Scenario | Expected Behavior |
|----------|-------------------|
| First reply to notifications | Send delivery failure notice, destroy email |
| Throttled reply (1+ per hour) | Skip notice but still destroy email |
| No valid route, not notifications | Destroy email |

### Email Routing Cases

**Complaint Handling:**
- Increment complaints_count on user
- Mark email as released

**Forwarding Rule:**
- Match handle to ForwardEmailRule
- Forward to target email
- Destroy original

**Group Handle Discussion:**
- When not blocked and actor authorized
- Create discussion
- Mark as released

**Blocked Address:**
- When address is blocked
- Do not create discussion

**Banned Sender:**
- When sender hostname is banned
- Destroy email

### Integration Tests

**Personal Email-to-Thread:**
- Route pattern: `d=<discussion_id>&u=<user_id>&k=<key>@host`
- Creates comment with correct discussion, user, and parent
- Marks email as released

**Throttled Bounce Notices:**
- First notification address reply: sends notice
- Second within hour: no additional notice

**Forwarding Rules:**
- Create ForwardEmailRule
- Email to handle@host forwards to rule's email
- Original email destroyed

**Group Handle Discussion Creation:**
- Email to grouphandle@host
- Creates discussion in group
- Sets author to sender
- Marks email as released

---

## 2. ReceivedEmailsController Tests

**Location:** `/spec/controllers/received_emails_controller_spec.rb`

### Test Helper

`mailin_params` function builds test payload in Mailin format:
- token, to, from, subject, body parameters
- Wraps in mailinMsg JSON structure

### Test Cases

**Ignore Self-Referential Emails:**
- Emails from reply_hostname are ignored
- Prevents mail loops

**Forward Rule Processing:**
- ForwardEmailRule.create with handle
- POST with matching address
- Last email delivered to rule's target

**Reply to Comment:**
- Route: `c=<comment_id>&d=<discussion_id>&u=<user_id>&k=<key>`
- Creates comment
- Sets parent to original comment
- Sets author to user

**Reply to Poll:**
- Route: `pt=p&pi=<poll_id>&d=<discussion_id>&u=<user_id>&k=<key>`
- Creates comment
- Sets parent to poll

**Reply to Discussion:**
- Route: `d=<discussion_id>&u=<user_id>&k=<key>`
- Creates comment
- Sets parent to discussion

**Start Discussion in Group:**
- Route: `<handle>+u=<user_id>&k=<key>`
- Creates discussion
- Sets author to user

**Group Handle Routes:**

| Scenario | Behavior |
|----------|----------|
| Invalid handle | No discussion created, email destroyed |
| Member email | Discussion created, email released |
| Unknown sender | Email kept unreleased, UnknownSender event published |
| Valid member alias | Discussion created by aliased user |
| Blocked member alias | No discussion created |

---

## 3. Chatbot Tests

**Observation:** No dedicated chatbot spec files were found in the codebase. Testing coverage appears to rely on:

- Integration through event trigger chain testing
- Manual testing via the test connection feature
- E2E tests via Nightwatch (not investigated in detail)

### Implicit Testing

Chatbot functionality is indirectly tested through:

1. **Event System Tests:** Events that include the Chatbots notify concern
2. **Service Pattern Tests:** Following the standard service pattern
3. **Authorization Tests:** CanCanCan ability specs

---

## 4. Ability Tests (Authorization)

**Expected Location:** `/spec/models/ability/chatbot_spec.rb` (not found)

The Ability::Chatbot module defines permissions but explicit specs were not located.

**Permissions to Test:**

| Action | Condition |
|--------|-----------|
| :create | User is group admin |
| :update | User is group admin |
| :destroy | User is group admin |
| :test | User is group admin |

---

## 5. Bot API Tests

**Expected Location:** `/spec/controllers/api/b2/` (not fully investigated)

### B2 API Test Considerations

**Authentication:**
- Valid api_key returns user
- Invalid api_key raises AccessDenied

**Discussions:**
- Create requires valid group membership
- Show respects visibility rules

**Polls:**
- Create and invite flow
- Show respects visibility

**Memberships:**
- Index returns group members
- Create syncs email list
- Requires admin access

### B3 API Test Considerations

**Authentication:**
- Requires B3_API_KEY env var of 16+ chars
- Validates params[:b3_api_key]

**User Management:**
- Deactivate finds active user, enqueues worker
- Reactivate finds deactivated user, calls service

---

## 6. OAuth/Identity Tests

**Expected Location:** `/spec/controllers/identities/`

### Test Considerations

**OAuth Flow:**
- Redirect to provider with correct params
- Callback handles code exchange
- Identity creation/linking logic
- User sign-in on successful link

**Identity Model:**
- force_user_attrs! updates user
- assign_logo! downloads and attaches avatar

---

## 7. Webhook Serializer Tests

**Expected Location:** `/spec/serializers/webhook/`

### Serializer Test Considerations

**Format-Specific Output:**
- Slack: text attribute with proper formatting
- Discord: content attribute, truncated to 1900 chars
- Microsoft: MessageCard format with @type, @context, themeColor
- Webex: markdown attribute
- Markdown: base format with text, icon_url, username

---

## 8. Test Patterns for Integrations

### Email Processing Test Pattern

```pseudo
describe "email routing" do
  before do
    set up ENV['REPLY_HOSTNAME']
    create user with email_api_key
    create group with handle
    add user to group
  end

  it "routes email correctly" do
    email = ReceivedEmail.create!(
      headers: { from: sender, to: route_address, subject: subject },
      body_text: body_content
    )

    expect { ReceivedEmailService.route(email) }
      .to change { Discussion.count }.by(1)
      .or change { Comment.count }.by(1)

    expect(email.reload.released).to eq true
  end
end
```

### Chatbot Service Test Pattern (Suggested)

```pseudo
describe ChatbotService do
  describe ".publish_event!" do
    let(:group) { create(:group) }
    let(:chatbot) { create(:chatbot, group: group, event_kinds: ['new_comment']) }
    let(:discussion) { create(:discussion, group: group) }
    let(:comment) { create(:comment, discussion: discussion) }
    let(:event) { Events::NewComment.create!(eventable: comment) }

    before do
      stub Clients::Webhook.post
    end

    it "sends notification to matching chatbot" do
      expect(Clients::Webhook).to receive(:post).with(chatbot.server, anything)
      ChatbotService.publish_event!(event.id)
    end

    it "skips chatbot when event kind not in event_kinds" do
      chatbot.update(event_kinds: ['poll_created'])
      expect(Clients::Webhook).not_to receive(:post)
      ChatbotService.publish_event!(event.id)
    end
  end
end
```

---

## 9. Test Coverage Summary

| Component | Coverage Level | Notes |
|-----------|---------------|-------|
| ReceivedEmailService | High | Comprehensive unit and integration tests |
| ReceivedEmailsController | High | Good coverage of routing scenarios |
| ChatbotService | Low | No dedicated specs found |
| Chatbot model | Low | No dedicated specs found |
| Ability::Chatbot | Low | Authorization logic not explicitly tested |
| Bot API B2 | Unknown | Not fully investigated |
| Bot API B3 | Unknown | Not fully investigated |
| Webhook Serializers | Unknown | Not fully investigated |
| OAuth Controllers | Unknown | Not fully investigated |

---

## 10. Recommendations for Test Improvement

1. **Add ChatbotService specs:**
   - Test publish_event! with various event types
   - Test chatbot filtering by event_kinds and recipient_chatbot_ids
   - Test error handling for failed webhook POSTs

2. **Add Chatbot model specs:**
   - Validation testing
   - Association testing

3. **Add Ability::Chatbot specs:**
   - Permission testing for admin vs non-admin

4. **Add Webhook serializer specs:**
   - Output format validation per platform
   - Template rendering testing

5. **Add Bot API controller specs:**
   - Authentication testing
   - Authorization testing
   - Success/error response formats
