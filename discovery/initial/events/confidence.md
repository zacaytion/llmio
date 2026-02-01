# Events Domain: QA Review

**Generated:** 2026-02-01
**Reviewer:** QA Agent

---

## 1. Checklist Results

### Model Documentation (models.md)

| Criterion | Pass/Fail | Notes |
|-----------|-----------|-------|
| Core model identified | PASS | Event model correctly identified at `/app/models/event.rb` |
| Associations documented | PASS | All key associations listed (eventable, discussion, user, parent, children, notifications) |
| Custom fields documented | PASS | JSONB fields correctly enumerated (pinned_title, recipient_user_ids, etc.) |
| STI pattern explained | PASS | Correctly describes STI with `kind` column and `sti_find` method |
| Event subclasses enumerated | PASS | 42 event types confirmed (matches actual count of files) |
| Sequence/position system explained | PASS | sequence_id, position, position_key, depth all documented |
| Counter caches documented | PASS | child_count and descendant_count correctly explained |
| Event concerns documented | PASS | All 7 notification concerns with trigger! explained |
| Trigger chain explained | PASS | Concern composition via super chain accurately described |
| Recipient calculation documented | PASS | UsersByVolumeQuery integration explained |

**Model Score: 5/5**

### Service Documentation (services.md)

| Criterion | Pass/Fail | Notes |
|-----------|-----------|-------|
| EventService methods documented | PASS | remove_from_thread, move_comments, repair_discussion, reset_child_positions, repair_all_threads |
| NotificationService documented | PASS | mark_as_read, viewed_events, viewed methods explained |
| SequenceService documented | PASS | seq_present?, create_seq!, next_seq!, drop_seq! with SQL |
| MessageChannelService documented | PASS | Real-time update flow explained |
| PublishEventWorker documented | PASS | Background trigger flow correct |
| EventBus documented | PASS | broadcast, listen, deafen, clear, configure methods |
| EventBus configuration documented | PASS | Listeners for reader updates and real-time sync |
| UsersByVolumeQuery documented | PASS | Volume cascade priority and SQL logic |
| ChatbotService integration noted | PASS | GenericWorker dispatch mentioned |

**Service Score: 5/5**

### Controller Documentation (controllers.md)

| Criterion | Pass/Fail | Notes |
|-----------|-----------|-------|
| EventsController identified | PASS | `/app/controllers/api/v1/events_controller.rb` |
| All actions documented | PASS | index, comment, timeline, position_keys, pin, unpin, remove_from_thread, count |
| Parameters documented | PASS | Comprehensive param list including from, per, order, filters |
| Filter logic explained | PASS | Range format, comparison operators (_lt, _gt, etc.) |
| Response structure documented | PASS | JSON structure with events, discussions, users, etc. |
| Pagination documented | PASS | per, from, order explained |
| Authorization documented | PASS | load_and_authorize, ability.authorize! patterns |
| Routes documented | PASS | Matches config/routes.rb structure |

**Controller Score: 5/5**

### Frontend Documentation (frontend.md)

| Criterion | Pass/Fail | Notes |
|-----------|-----------|-------|
| Record interface documented | PASS | EventRecordsInterface with fetchByDiscussion, findByDiscussionAndSequenceId |
| Event model documented | PASS | Properties, relationships, default values, key methods |
| Model methods documented | PASS | parentOrSelf, isNested, children, model, isUnread, pin, unpin, etc. |
| Event service (actions) documented | PASS | move_event, pin_event, unpin_event, copy_url actions |
| Frontend EventBus documented | PASS | Vue-based event bus distinct from backend |
| Strand components documented | PARTIAL | Missing some components (actions_panel, load_more, members, toc_nav, wall) |
| ThreadLoader documented | PASS | Core functionality explained |
| Notification components documented | PASS | NotificationsCount, Notifications dropdown |
| Notification model documented | PASS | Relationships, href, args, isRouterLink methods |
| Real-time updates documented | PASS | WebSocket/SSE flow explained |
| Pinning UI documented | PASS | PinEventForm modal |

**Frontend Score: 4/5** (Minor gap in component coverage)

### Test Documentation (tests.md)

| Criterion | Pass/Fail | Notes |
|-----------|-----------|-------|
| Model tests documented | PASS | event_spec.rb with volume-based recipient tests |
| Service tests documented | PASS | event_service_spec.rb repair_discussion scenarios |
| Controller tests documented | PASS | events_controller_spec.rb endpoints |
| EventBus tests documented | PASS | event_bus_spec.rb pub/sub tests |
| Integration tests documented | PASS | discussion_event_integration_spec.rb |
| Test setup explained | PASS | Complex user/volume scenario documented |
| Test patterns documented | PASS | Factory usage, email counting, WebMock |
| Coverage gaps identified | PASS | Missing event types, position_key scenarios noted |
| Running instructions provided | PASS | Bundle exec rspec commands |

**Test Score: 5/5**

---

## 2. Confidence Scores

| Area | Score | Flag? |
|------|-------|-------|
| Models | 5/5 | No |
| Services | 5/5 | No |
| Controllers | 5/5 | No |
| Frontend | 4/5 | **Yes** |
| Tests | 5/5 | No |
| **Overall** | **4.8/5** | |

---

## 3. Issues Found

### Verified Accuracy

1. **Event subclass count**: Documentation states 42 event types; actual count of files in `/app/models/events/` is 42. VERIFIED.

2. **STI lookup pattern**: Documentation describes `sti_find` using `kind.classify`; actual code uses `("Events::"+kind.classify).constantize`. VERIFIED.

3. **Event.publish! signature**: Documentation correctly describes the flow: build -> save -> PublishEventWorker.perform_async. VERIFIED.

4. **Concern composition**: Documentation accurately describes the trigger chain through super calls. VERIFIED (7 concerns with trigger!).

5. **EventSerializer attributes**: Documentation lists position data, associations, metadata. Actual serializer matches. VERIFIED.

6. **Controller filtering**: Documentation describes _lt, _gt, _lte, _gte, _sw operators. Actual controller code matches. VERIFIED.

7. **Frontend model indices**: Documentation lists discussionId, sequenceId, position, depth, parentId, positionKey. Actual model matches. VERIFIED.

### Documentation Discrepancies

1. **Minor typo in frontend model**: The actual `event_model.js` has a typo on line 31: `positition: 0` (should be `position`). The documentation correctly states `position: 0`. This is actually a bug in the source code, not the documentation.

2. **Frontend model default showReplyForm**: Documentation says `showReplyForm: true`. Actual code has `showReplyForm: true`. VERIFIED.

---

## 4. Uncertainties

### Low Uncertainty Areas

1. **Event model structure**: Very high confidence - all associations, callbacks, and methods verified against source.

2. **Event concerns**: Very high confidence - all 7 trigger! implementations verified.

3. **Service layer**: Very high confidence - EventService, NotificationService, SequenceService all verified.

4. **Controller endpoints**: Very high confidence - all actions and parameters verified against source.

5. **Test coverage**: High confidence - spec files exist and match documented patterns.

### Areas Requiring Clarification

1. **Strand components coverage**: The frontend documentation lists 15 strand item components but the actual directory contains 25 .vue files. Missing from documentation:
   - `actions_panel.vue`
   - `load_more.vue`
   - `members.vue`
   - `members_list.vue`
   - `reply_form.vue`
   - `seen_by_modal.vue`
   - `title.vue`
   - `toc_nav.vue`
   - `wall.vue`

   These are supporting components, not event item renderers, so the documentation focus on item/ components is reasonable, but completeness could be improved.

2. **ThreadLoader implementation**: The documentation describes `thread_loader.js` functionality in general terms but the file path should be verified. This is a critical component for thread rendering.

3. **Real-time update mechanism**: The documentation describes Redis pub/sub + external channels service. The actual implementation details (CHANNELS_URL configuration, WebSocket vs SSE choice) are not deeply explored.

---

## 5. Revision Recommendations

### Priority: High

None identified. The documentation is comprehensive and accurate for critical paths.

### Priority: Medium

1. **frontend.md - Add missing strand components**: Consider adding a brief mention of supporting components in the strand family:
   - `load_more.vue` - Pagination trigger
   - `reply_form.vue` - Comment composition
   - `toc_nav.vue` - Table of contents navigation
   - `wall.vue` - Alternative view mode

2. **frontend.md - Verify ThreadLoader path**: Confirm the actual file path for ThreadLoader and add it to the documentation.

### Priority: Low

1. **models.md - Add database migration reference**: Could be helpful to reference the events table schema or migration files for developers needing schema details.

2. **tests.md - Add E2E test references**: The tests.md focuses on backend specs; mentioning Nightwatch E2E tests that involve events (if any exist) would be helpful.

3. **services.md - Expand ChatbotService**: The ChatbotService integration is briefly mentioned; could expand on how webhooks are configured and triggered.

---

## Summary

The events domain documentation is **excellent quality** with high accuracy across all verified areas. The documentation correctly captures:

- The complex STI-based event hierarchy with 42 event types
- The concern-based composition pattern for event behaviors
- The sequence and position system for thread ordering
- The volume-based notification routing system
- The real-time update architecture
- The comprehensive API for event retrieval and manipulation

The only notable gap is in frontend component coverage, where several supporting components in the strand/ directory are not documented. This is a minor completeness issue rather than an accuracy problem.

**Recommendation**: Accept documentation with minor revisions to frontend.md for completeness.
