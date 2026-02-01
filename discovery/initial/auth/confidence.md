# Auth Domain - QA Confidence Report

**Generated:** 2026-02-01
**Reviewer:** QA Agent
**Overall Domain Confidence:** 3.8/5

---

## Executive Summary

The auth domain documentation is largely accurate and well-structured, with models and services documentation being the strongest. The frontend documentation has notable gaps, particularly around state management details and error handling flows. Test documentation is comprehensive in coverage enumeration but reveals significant gaps in the actual test suite.

---

## 1. Checklist Results

### 1.1 Model Documentation (`models.md`)

| Checklist Item | Status | Notes |
|----------------|--------|-------|
| All attributes listed with types | PASS | Comprehensive attribute tables with types |
| Associations documented | PASS | All associations correctly documented |
| Validations described | PASS | Validations match source code |
| Callbacks and side effects noted | PASS | before_validation, before_save callbacks documented |
| Scopes listed with purpose | PASS | All scopes documented with descriptions |
| Concerns/mixins identified | PASS | Concerns listed, though some details sparse |

**Verified Claims:**
- User model Devise modules match source (line 25-26 in user.rb)
- LoginToken EXPIRATION constant confirmed (line 8 in login_token.rb)
- Identity validations confirmed (lines 5-6 in identity.rb)
- User scopes verified (lines 115-133 in user.rb)

**Issues Found:**
- LoginToken code generation described as "between 100000 and 999999" but code shows it generates values >= 100000 and < 999999 via `Random.new.rand(999999)` with retry - slight inaccuracy
- User model lists `failed_attempts` and `unlock_token` attributes but documentation doesn't explain when these are used (Devise lockable)

**Score: 4/5**

---

### 1.2 Service Documentation (`services.md`)

| Checklist Item | Status | Notes |
|----------------|--------|-------|
| Public methods listed with signatures | PASS | All methods documented with signatures |
| Trigger conditions documented | PASS | "Triggered by" section for each method |
| Side effects documented | PASS | Side effects clearly listed |
| Events emitted documented | PASS | EventBus events table at end |
| Error conditions documented | PASS | Exceptions documented |
| Pseudo-code for complex logic | PASS | Logic steps documented for each method |

**Verified Claims:**
- UserService.create logic matches source (lines 5-18)
- UserService.verify logic matches source (lines 20-30)
- UserService.deactivate authorization pattern confirmed (lines 32-35)
- EventBus.broadcast calls verified in update method (line 76)

**Issues Found:**
- Documentation states `UserService.create` "returns user (check errors for validation failures)" but doesn't mention that `save` (not `save!`) is used, so errors are on the user object not raised
- Missing documentation for `UserService.delete_spam_user` method (referenced in tests but not in service docs)

**Score: 4/5**

---

### 1.3 Controller Documentation (`controllers.md`)

| Checklist Item | Status | Notes |
|----------------|--------|-------|
| HTTP method and path | PASS | All endpoints documented with method/path |
| Authentication requirements | PASS | Auth requirements stated per endpoint |
| Request parameters with types | PASS | Parameter tables included |
| Response structure with examples | PASS | JSON examples provided |
| Error responses | PASS | Error response examples included |
| Authorization rules | PARTIAL | Some authorization details missing |

**Verified Claims:**
- Sessions#create authentication strategies match source (lines 36-44)
- Sessions#destroy secret_token regeneration confirmed (line 19)
- Error response structure matches source (lines 28-34)
- Permitted parameters confirmed (lines 51-55)

**Issues Found:**
- Documentation mentions `user[name]` parameter for sessions#create but source shows it updates name AFTER login (line 9), not as auth parameter
- Missing `pending_login_token` helper method documentation - referenced but not explained
- OAuth state parameter concern raised in "Open Questions" is valid - no state parameter visible in identity controllers

**Score: 4/5**

---

### 1.4 Frontend Documentation (`frontend.md`)

| Checklist Item | Status | Notes |
|----------------|--------|-------|
| Component names and purposes | PASS | All components documented |
| Props/inputs with types | PARTIAL | Props listed but some missing type details |
| API calls made | PASS | API calls documented per component |
| User interactions | PASS | Interaction flows documented |

**Verified Claims:**
- All 9 auth components exist in `vue/src/components/auth/`
- signup_form.vue props match documentation (line 7-8)
- termsUrl, privacyUrl computed properties confirmed (lines 32-33)
- allow computed uses `AppConfig.features.app.create_user` as documented (lines 35-37)

**Issues Found:**
- **signup_form.vue discrepancy**: Documentation says `vars.legalAccepted` and `vars.emailNewsletter` are in data(), but actual component has `vars: {name: this.user.name, site_name: AppConfig.theme.site_name}` in data() - legal/newsletter checkboxes use v-model directly on vars but not initialized
- **Missing component**: Documentation doesn't mention all the actual data properties (e.g., `siteName`)
- **AuthService not verified**: File location `/vue/src/shared/services/auth_service.js` not verified
- **State management section** is speculative ("Auth state is managed through...") without concrete code verification
- **Self-assigned confidence of 3/5** in the document header is accurate

**Score: 3/5**

---

### 1.5 Test Documentation (`tests.md`)

| Checklist Item | Status | Notes |
|----------------|--------|-------|
| Scenarios covered | PASS | Comprehensive test scenario enumeration |
| Factories used | PASS | Factories table included |
| Gaps identified | PASS | Extensive gap analysis provided |

**Verified Claims:**
- Sessions controller test structure matches actual file
- Password auth tests confirmed (lines 14-30)
- Token auth tests confirmed (lines 33-82)
- Test scenarios accurately documented

**Issues Found:**
- Test documentation self-identifies confidence as 4/5, which is appropriate
- Gap analysis is valuable but some recommendations may already be covered by Devise's built-in tests
- Documentation doesn't distinguish between tests that exist vs. tests that should exist in some sections

**Score: 4/5**

---

## 2. Confidence Scores Summary

| File | Score | Status |
|------|-------|--------|
| models.md | 4/5 | PASS |
| services.md | 4/5 | PASS |
| controllers.md | 4/5 | PASS |
| frontend.md | 3/5 | **NEEDS REVISION** |
| tests.md | 4/5 | PASS |

**Overall Domain Score: 3.8/5**

---

## 3. Issues Found

### Critical Issues
None identified.

### Major Issues

1. **Frontend data property discrepancy** (`frontend.md`): The signup_form.vue documentation claims `vars.legalAccepted` and `vars.emailNewsletter` are initialized in data(), but the actual component only initializes `{name: this.user.name, site_name: ...}`. These properties are created dynamically by v-model.

2. **Missing spam user deletion documentation** (`services.md`): UserService has a `delete_spam_user` method referenced in tests but not documented in services.md.

### Minor Issues

1. **LoginToken code generation** (`models.md`): Minor inaccuracy in describing the code generation range.

2. **Session create name parameter** (`controllers.md`): Documentation implies name is an auth parameter, but it's actually applied post-authentication.

3. **Frontend AuthService** (`frontend.md`): Service file existence and method implementations not verified against source.

4. **OAuth state parameter** (`controllers.md`): Potential security concern noted but not investigated.

---

## 4. Uncertainties

### Cannot Be Answered From Documentation

1. **Session invalidation timing**: How does secret_token regeneration propagate to invalidate existing sessions? The sessions controller regenerates on logout, but the mechanism for other sessions isn't clear.

2. **OAuth state parameter implementation**: Is CSRF protection actually missing from OAuth flows, or is it handled elsewhere (middleware, provider library)?

3. **Rate limiting implementation**: Documentation notes absence of rate limiting but doesn't verify if it exists at infrastructure level (nginx, rack-attack, etc.).

4. **Token cleanup**: Neither documentation nor source review found a scheduled job for cleaning expired LoginTokens.

5. **Frontend session refresh**: How does the frontend detect and handle session expiration?

### Partially Answered

1. **Email verification flow**: UserService.verify is called during sign_in, but the exact trigger paths through various auth flows could be clearer.

2. **Pending identity lifecycle**: How long do pending identities persist? When are they cleaned up?

---

## 5. Revision Recommendations

### frontend.md (Priority: HIGH)

The frontend documentation should be revised to:

1. **Verify AuthService implementation**: Read and document the actual `auth_service.js` file rather than inferring behavior.

2. **Fix signup_form.vue data properties**: Correct the documentation to show:
   ```javascript
   data() {
     return {
       siteName: AppConfig.theme.site_name,
       vars: {name: this.user.name, site_name: AppConfig.theme.site_name},
       loading: false
     };
   }
   ```

3. **Document actual state flow**: Trace through actual component code to document how `user` object mutations flow between components.

4. **Verify all file paths**: Confirm existence of:
   - `/vue/src/shared/services/auth_service.js`
   - `/vue/src/shared/services/session.js`
   - `/vue/src/shared/interfaces/user_model.js`
   - `/vue/src/mixins/auth_modal.js`

5. **Add missing component details**: Document `siteName` data property and other actual component internals.

### services.md (Priority: MEDIUM)

1. **Add delete_spam_user documentation**: Document this method referenced in tests.

2. **Clarify return values**: Note that `create` uses `save` not `save!`, so validation errors are on the object.

### controllers.md (Priority: LOW)

1. **Clarify name parameter timing**: Note that name update happens after successful authentication, not as part of it.

2. **Investigate OAuth state**: Either verify CSRF protection exists or document as actual security gap.

### models.md (Priority: LOW)

1. **Fix code generation description**: Minor correction to rand behavior.

---

## 6. Verification Methodology

### Files Reviewed
- `/Users/z/Code/loomio/app/models/user.rb`
- `/Users/z/Code/loomio/app/models/login_token.rb`
- `/Users/z/Code/loomio/app/models/identity.rb`
- `/Users/z/Code/loomio/app/services/user_service.rb`
- `/Users/z/Code/loomio/app/controllers/api/v1/sessions_controller.rb`
- `/Users/z/Code/loomio/spec/controllers/api/v1/sessions_controller_spec.rb`
- `/Users/z/Code/loomio/vue/src/components/auth/signup_form.vue`
- All auth component files via glob

### Verification Approach
1. Read documentation claims
2. Locate corresponding source code
3. Compare documented behavior to actual implementation
4. Note discrepancies and gaps
5. Cross-reference with test files for behavior confirmation

---

## Appendix: Document Self-Assessments vs QA Assessment

| Document | Self-Assessment | QA Assessment | Delta |
|----------|-----------------|---------------|-------|
| models.md | 4/5 | 4/5 | 0 |
| services.md | 4/5 | 4/5 | 0 |
| controllers.md | 4/5 | 4/5 | 0 |
| frontend.md | 3/5 | 3/5 | 0 |
| tests.md | 4/5 | 4/5 | 0 |

The documentation's self-assessed confidence levels align with QA findings. The frontend.md self-assessment of 3/5 was appropriately conservative.
