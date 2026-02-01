# Integrations Domain: Frontend

**Generated:** 2026-02-01
**Confidence Rating:** 4/5

---

## Overview

The frontend provides UI components for group admins to configure chatbot/webhook integrations. The chatbot configuration is accessed through group settings.

---

## 1. Vue Components

### ChatbotList

**Location:** `/vue/src/components/chatbot/list.vue`

**Purpose:** Display and manage all chatbot integrations for a group.

**Props:**
- group: Object - the group to manage chatbots for

**Behavior:**
1. On mount:
   - Fetch chatbots from API for the group
   - Set up record watching for real-time updates
2. Display loading state while fetching
3. Show list of configured chatbots with name, kind, server, channel
4. Clicking a chatbot opens edit modal (WebhookForm or MatrixForm based on kind)
5. Action menu provides "Add chatbot" options for each supported platform

**Supported Platforms:**
- matrix (Matrix protocol)
- slack (Slack webhook)
- discord (Discord webhook)
- mattermost (Mattermost webhook)
- teams (Microsoft Teams webhook)
- webex (Webex webhook)

**Icons:**
- Matrix: mdi-matrix
- Slack: mdi-slack
- Discord: mdi-discord
- Mattermost: mdi-chat-processing
- Teams: mdi-microsoft-teams
- Webex: custom webex icon

---

### ChatbotWebhookForm

**Location:** `/vue/src/components/chatbot/webhook_form.vue`

**Purpose:** Configure webhook-based chatbot integrations.

**Props:**
- chatbot: Object - the chatbot model to edit/create

**Form Fields:**

| Field | Description |
|-------|-------------|
| name | User-friendly name for the integration |
| server | Webhook URL (displayed as "Webhook URL") |
| notificationOnly | Checkbox for minimal notifications vs full content |
| eventKinds | Multi-select checkboxes for which events to notify |

**Available Event Kinds:**
Loaded from AppConfig.webhookEventKinds, includes:
- new_discussion
- discussion_edited
- new_comment
- poll_created
- poll_edited
- poll_closing_soon
- poll_expired
- poll_closed_by_user
- poll_reopened
- outcome_created
- outcome_updated
- outcome_review_due
- stance_created
- stance_updated

**Actions:**
- Save: chatbot.save() - creates or updates via API
- Delete: chatbot.destroy() - removes integration (only shown for existing)
- Test Connection: POSTs to /api/v1/chatbots/test with server URL

**Help Links:**
Contextual documentation links based on webhook_kind:
- Slack: loomio help /integrations/slack
- Discord: loomio help /integrations/discord
- Microsoft Teams: loomio help /integrations/microsoft_teams
- Mattermost: loomio help /integrations/mattermost
- Webex: loomio help /integrations/webex

---

### ChatbotMatrixForm

**Location:** `/vue/src/components/chatbot/matrix_form.vue`

**Purpose:** Configure Matrix protocol chatbot integrations.

**Props:**
- chatbot: Object - the chatbot model to edit/create

**Form Fields:**

| Field | Description |
|-------|-------------|
| name | User-friendly name for the integration |
| server | Matrix homeserver URL (e.g., https://example.com) |
| accessToken | Bot user's access token from Matrix |
| channel | Room identifier (e.g., #general:example.com or !roomid:example.com) |
| notificationOnly | Checkbox for minimal notifications |
| eventKinds | Multi-select for event types |

**Hints:**
- Access token found at: User menu > All settings > Help & about > Access token
- Room ID found at: Room options > Settings > Advanced > Internal room ID

**Actions:**
- Save: chatbot.save()
- Delete: chatbot.destroy() (for existing chatbots)
- Test Connection: Currently commented out in template

---

## 2. Models

### ChatbotModel

**Location:** `/vue/src/shared/models/chatbot_model.js`

**Purpose:** Client-side model for chatbot records.

**Default Values:**
```pseudo
{
  groupId: null,
  name: null,
  server: null,
  accessToken: null,
  eventKinds: [],
  kind: null,
  webhookKind: null,
  errors: {},
  notificationOnly: false
}
```

**Relationships:**
- belongsTo('group') - parent group

---

## 3. Services

### ChatbotService

**Location:** `/vue/src/shared/services/chatbot_service.js`

**Purpose:** Provides action definitions for adding new chatbot integrations.

**Method: addActions(group)**

Returns action definitions for each supported platform:

| Platform | kind | webhookKind | Component |
|----------|------|-------------|-----------|
| matrix | matrix | - | ChatbotMatrixForm |
| slack | webhook | slack | ChatbotWebhookForm |
| discord | webhook | discord | ChatbotWebhookForm |
| microsoft | webhook | microsoft | ChatbotWebhookForm |
| mattermost | webhook | markdown | ChatbotWebhookForm |
| webex | webhook | webex | ChatbotWebhookForm |

Each action:
1. Creates a new chatbot record with appropriate kind/webhookKind
2. Opens the corresponding form in a modal

---

### ChatbotRecordsInterface

**Location:** `/vue/src/shared/interfaces/chatbot_records_interface.js`

**Purpose:** LokiJS record store interface for chatbots.

**API Endpoints:**
- GET /api/v1/chatbots?group_id=X - fetch chatbots
- POST /api/v1/chatbots - create chatbot
- PATCH /api/v1/chatbots/:id - update chatbot
- DELETE /api/v1/chatbots/:id - destroy chatbot

---

## 4. Access via Group Settings

### GroupService Integration

**Location:** `/vue/src/shared/services/group_service.js`

The chatbot list is accessed through group admin actions:

```pseudo
chatbots: {
  name: 'chatbot.chatbots',
  icon: 'mdi-robot',
  menu: true,
  canPerform() {
    return group.adminsInclude(Session.user());
  },
  perform() {
    return openModal({
      component: 'ChatbotList',
      props: { group }
    });
  }
}
```

**Access Control:** Only visible to group admins.

---

## 5. Configuration

### AppConfig.webhookEventKinds

**Source:** Loaded from server during boot, from `/config/webhook_event_kinds.yml`

**Contents:**
```yaml
- new_discussion
- discussion_edited
- new_comment
- poll_created
- poll_edited
- poll_closing_soon
- poll_expired
- poll_closed_by_user
- poll_reopened
- outcome_created
- outcome_updated
- outcome_review_due
- stance_created
- stance_updated
```

---

## 6. UI Flow

### Adding a Chatbot

```pseudo
1. User navigates to Group Settings
2. User is a group admin
3. User clicks "Chatbots" action (mdi-robot icon)
4. ChatbotList modal opens
5. User clicks "Add chatbot" menu
6. User selects platform (Slack, Discord, Matrix, etc.)
7. Appropriate form modal opens (WebhookForm or MatrixForm)
8. User fills in configuration:
   - Name
   - Webhook URL / Server + Token + Channel
   - Notification only toggle
   - Event types to notify
9. User clicks Save
10. API creates chatbot record
11. Modal closes, list updates
```

### Testing a Webhook

```pseudo
1. In WebhookForm, user clicks "Test Connection"
2. Frontend POSTs to /api/v1/chatbots/test with:
   - server: webhook URL
   - kind: 'slack_webhook'
3. Backend sends test message to webhook URL
4. Flash message: "Check for test message"
5. User verifies message appeared in external platform
```

### Editing/Deleting a Chatbot

```pseudo
1. In ChatbotList, user clicks on existing chatbot
2. Edit form opens with current values
3. User can modify and Save, or Delete
4. Delete confirmation removes integration
```

---

## 7. Internationalization

Translation keys used:

- chatbot.chatbots - modal title
- chatbot.name - name field label
- chatbot.webhook_url - server field label (webhook)
- chatbot.homeserver_url - server field label (matrix)
- chatbot.access_token - token field label
- chatbot.channel - channel field label
- chatbot.notification_only_label - checkbox label
- chatbot.event_kind_helptext - event selection help
- chatbot.test_connection - test button label
- chatbot.check_for_test_message - test success message
- chatbot.saved - save success message
- webhook.event_kinds.[kind] - individual event type labels
- webhook.formats.[format] - webhook format labels
