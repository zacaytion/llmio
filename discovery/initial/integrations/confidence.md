# Integrations Domain: QA Review

**Reviewed:** 2026-02-01
**Reviewer:** QA Agent

---

## 1. Checklist Results

### models.md

| Claim | Verified | Notes |
|-------|----------|-------|
| Chatbot model location at `/app/models/chatbot.rb` | PASS | Confirmed |
| Chatbot belongs_to group, author | PASS | Line 2-3 in chatbot.rb |
| Chatbot validations (server, name required; kind/webhook_kind inclusion) | PASS | Lines 5-8 |
| Chatbot config method returns server/access_token/channel hash | PASS | Lines 10-17 |
| ReceivedEmail model location | PASS | `/app/models/received_email.rb` |
| ReceivedEmail associations (belongs_to group, has_many_attached attachments) | PASS | Lines 2-3 |
| ReceivedEmail scopes (unreleased, released) | PASS | Lines 5-6 |
| ReceivedEmail key methods (header, sender_email, route_address, etc.) | PASS | All methods present |
| Identity model location | PASS | `/app/models/identity.rb` |
| Identity table name omniauth_identities | PASS | Line 3: `self.table_name = :omniauth_identities` |
| Identity associations and validations | PASS | Lines 5-8 |
| Identity methods (force_user_attrs!, assign_logo!) | PASS | Lines 14-27 |
| ForwardEmailRule model exists | PASS | `/app/models/forward_email_rule.rb` |
| MemberEmailAlias model exists | PASS | `/app/models/member_email_alias.rb` |
| MemberEmailAlias scopes (blocked, allowed) | PASS | Lines 6-7 |
| User api_key and email_api_key tokens | PASS | user.rb lines 109-110 |

### services.md

| Claim | Verified | Notes |
|-------|----------|-------|
| ChatbotService location | PASS | `/app/services/chatbot_service.rb` |
| ChatbotService.create method flow | PASS | Lines 2-7 |
| ChatbotService.update preserves empty access_token | PASS | Line 11 |
| ChatbotService.destroy method | PASS | Lines 17-20 |
| ChatbotService.publish_event! flow | PASS | Lines 22-70, matches description |
| ChatbotService.publish_test! method | PASS | Lines 72-83 |
| ReceivedEmailService location | PASS | `/app/services/received_email_service.rb` |
| ReceivedEmailService.route method flow | PASS | Lines 27-116, comprehensive routing logic verified |
| ReceivedEmailService.extract_reply_body method | PASS | Lines 118-128 |
| ReceivedEmailService.delete_old_emails (60 days) | PASS | Lines 130-132 |
| Clients::Webhook location and methods | PASS | `/app/extras/clients/webhook.rb` |
| Webhook serializer selection chain | PASS | Lines 16-21 in webhook.rb |
| Events::Notify::Chatbots concern | PASS | `/app/models/concerns/events/notify/chatbots.rb` |
| Event trigger chain for chatbots via GenericWorker | PASS | Line 4 calls GenericWorker.perform_async |

### controllers.md

| Claim | Verified | Notes |
|-------|----------|-------|
| ChatbotsController location | PASS | `/app/controllers/api/v1/chatbots_controller.rb` |
| ChatbotsController index action with show_chatbots permission | PASS | Line 3 |
| ChatbotsController test action | PASS | Lines 8-11 |
| ChatbotsController index_scope with current_user_is_admin | PASS | Lines 13-15 |
| ReceivedEmailsController location | PASS | `/app/controllers/received_emails_controller.rb` |
| B2 BaseController authentication via api_key | PASS | `/app/controllers/api/b2/base_controller.rb` lines 6-12 |
| B3 UsersController authentication via B3_API_KEY | PASS | `/app/controllers/api/b3/users_controller.rb` lines 6-8 |
| B3 UsersController deactivate/reactivate actions | PASS | Lines 11-20 |
| Identities::BaseController OAuth flow | PASS | `/app/controllers/identities/base_controller.rb` |
| OAuth create action finds/creates identity | PASS | Lines 7-57 |
| ChatbotSerializer conditional includes (server, channel) | PASS | `/app/serializers/chatbot_serializer.rb` lines 4-9 |

**Note:** Documentation states B3 API key must be "16+ characters" but code shows `length > 16` (line 7), meaning it requires more than 16, not 16 or more. Minor discrepancy.

### frontend.md

| Claim | Verified | Notes |
|-------|----------|-------|
| ChatbotList component location | PASS | `/vue/src/components/chatbot/list.vue` |
| ChatbotList supported platforms | PASS | Line 18: `kinds: ['matrix', 'slack', 'discord', 'mattermost', 'teams', 'webex']` |
| ChatbotList icons mapping | PASS | Lines 21-28 |
| ChatbotList loads chatbots on mount | PASS | Lines 42-44 |
| ChatbotList editChatbot opens correct form | PASS | Lines 52-67 |
| ChatbotWebhookForm component exists | PASS | `/vue/src/components/chatbot/webhook_form.vue` |
| ChatbotMatrixForm component exists | PASS | `/vue/src/components/chatbot/matrix_form.vue` |

### tests.md

| Claim | Verified | Notes |
|-------|----------|-------|
| ReceivedEmailService specs location | PASS | `/spec/services/received_email_service_spec.rb` |
| Reply body extraction tests | PASS | Multiple test cases verified |
| Subject stripping tests | PASS | Lines 70-85 |
| Notifications address routing tests | PASS | Lines 87-131 |
| Email routing integration tests | PASS | Lines 253-385 |
| ReceivedEmailsController specs location | PASS | `/spec/controllers/received_emails_controller_spec.rb` |
| No dedicated chatbot specs | PASS | Glob search found no chatbot spec files |
| B2 API tests exist | PASS | Found 4 spec files in `/spec/controllers/api/b2/` |
| B2 tests include authentication testing | PASS | Verified in discussions_controller_spec.rb |

---

## 2. Confidence Scores

| Document | Score | Assessment |
|----------|-------|------------|
| models.md | **4/5** | Accurate model documentation. Minor issues with ForwardEmailRule description (says "Defined via database table" but model file exists). |
| services.md | **4/5** | Service flows accurately documented. publish_event! flow verified. Minor: publish_test! uses MAIN_REDIS_POOL for matrix, not CACHE_REDIS_POOL as in publish_event!. |
| controllers.md | **4/5** | Controller documentation accurate. B3 key length check is "> 16" not ">= 16". Route configuration section uses pseudo-code that matches actual routes. |
| frontend.md | **4/5** | Component documentation accurate. Platform list matches code. UI flow descriptions are logical. |
| tests.md | **3/5** | **FLAG** - Stated "Observation: No dedicated chatbot spec files were found" is correct, but test coverage assessment for B2 API says "Unknown" when specs clearly exist. |

**Overall Domain Score: 4/5**

---

## 3. Issues Found

### Accuracy Issues

1. **models.md Line ~156**: ForwardEmailRule description says "Defined via database table, used in ReceivedEmailService" suggesting no model file, but `/app/models/forward_email_rule.rb` exists (though minimal).

2. **controllers.md Line ~189**: B3 API key requirement described as "16+ characters" but code requires `length > 16`, meaning 17+ characters minimum.

3. **services.md**: Inconsistent Redis pool naming - publish_event! uses `CACHE_REDIS_POOL` while publish_test! uses `MAIN_REDIS_POOL`. Not documented.

4. **tests.md**: Claims "Bot API B2: Unknown - Not fully investigated" but comprehensive specs exist in `/spec/controllers/api/b2/` with authentication, authorization, and functional tests.

### Missing Information

1. **models.md**: Does not document the `PROVIDERS` constant in Identity model loaded from `config/providers.yml`.

2. **services.md**: Does not mention error handling in ChatbotService.publish_event! (line 54 logs to Sentry on non-200 responses).

3. **controllers.md**: Missing documentation for `show_chatbots` ability check on group (line 3 of chatbots_controller).

4. **frontend.md**: Missing documentation for help button path (`en/user_manual/groups/integrations/chatbots`).

### Completeness Gaps

1. No documentation of the `Events::UnknownSender` event published when sender cannot be authenticated.

2. No documentation of the `ForwardMailer` used for bounce notices and forwarding.

3. No documentation of `ThrottleService` integration for bounce notice rate limiting.

---

## 4. Uncertainties

1. **Webhook Format Variations**: The documentation mentions serializers for different platforms but doesn't detail the actual payload differences between Slack, Discord, Microsoft, and other formats. Need to verify actual serializer implementations.

2. **Matrix Integration**: The publish_event! method publishes to Redis with `chatbot/publish` key, suggesting an external service handles Matrix API calls. The architecture of this external service is not documented.

3. **OAuth Controllers**: The documentation covers the base OAuth flow but the specific behavior differences between Google, Nextcloud, OAuth (generic), and SAML controllers are not detailed.

4. **Ability::Group#show_chatbots**: The chatbots_controller uses `load_and_authorize(:group, :show_chatbots)` but this permission is not documented in the Ability section.

5. **LOOMIO_SSO_FORCE_USER_ATTRS**: Environment variable mentioned in identity controller but not documented in integrations domain.

---

## 5. Revision Recommendations

### High Priority

1. **tests.md**: Update B2 API test coverage assessment from "Unknown" to "Moderate" and document the existing test coverage:
   - `discussions_controller_spec.rb`: Tests create with valid/invalid key, permissions
   - `memberships_controller_spec.rb`: Tests sync and admin requirements
   - `polls_controller_spec.rb`: Tests create flow
   - `comments_controller_spec.rb`: Tests comment creation

2. **controllers.md**: Correct B3 API key length requirement from "16+" to "17+" (strictly greater than 16).

### Medium Priority

3. **models.md**: Update ForwardEmailRule section to acknowledge the model file exists at `/app/models/forward_email_rule.rb`.

4. **services.md**: Document the error handling in ChatbotService.publish_event! that logs to Sentry on webhook failures.

5. **services.md**: Document the Events::UnknownSender event published when sender authentication fails.

6. **controllers.md**: Add documentation for the `show_chatbots` ability used in chatbots_controller index action.

### Low Priority

7. **frontend.md**: Add help button path documentation.

8. **models.md**: Document the Identity.PROVIDERS constant and config/providers.yml dependency.

9. **services.md**: Clarify the Redis pool usage differences between publish_event! and publish_test!.

10. **Add new section**: Consider adding documentation about ForwardMailer and ThrottleService integration for email processing.

---

## Summary

The integrations domain documentation is generally accurate and comprehensive. The main area needing attention is the test coverage documentation which understates the actual testing present for the B2 API. The models, services, and controllers documentation accurately reflects the codebase with only minor discrepancies. The frontend documentation correctly describes the chatbot configuration UI components.

The confidence score of 4/5 reflects solid documentation with room for improvement in completeness and accuracy of edge cases.
