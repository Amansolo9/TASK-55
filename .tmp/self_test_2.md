# ClubOps Static Delivery Acceptance + Architecture Audit

## 1. Verdict
- **Overall conclusion: Partial Pass**

## 2. Scope and Static Verification Boundary
- Reviewed: repository structure, docs/config (`README.md`, `go.mod`, `Dockerfile`, `docker-compose.yml`), entrypoints/routes, middleware/authz, core services/stores, templates, unit/API/E2E test code.
- Not reviewed: runtime behavior in browser/server, container orchestration behavior, actual DB/file-system side effects during execution, performance under load.
- Intentionally not executed: app startup, tests, Docker, Playwright, any external service calls (per audit boundary).
- Manual verification required for: true offline operation in target environment, session timeout behavior under real idle timing, UI behavior under real HTMX interactions.

## 3. Repository / Requirement Mapping Summary
- Prompt goal mapped: offline governance + finance portal with HTMX server-rendered flows, local auth/RBAC/audit/security, member lifecycle, review moderation/appeals, finance approval workflows, MDM, credit engine, and local feature flags.
- Main implementation areas mapped: route layer (`internal/handlers/handler.go`), auth/RBAC/CSRF/audit middleware (`internal/middleware/*.go`), domain services (`internal/services/*.go`), SQLite schema/access (`internal/store/*.go`), UI templates (`views/**/*.html`), tests (`unit_tests`, `API_tests`, `e2e`).
- Highest-risk findings: credit issuance data model conflicts with prompt semantics (High), plus several medium-severity security/validation/professionalism gaps.

## 4. Section-by-section Review

### 1. Hard Gates

#### 1.1 Documentation and static verifiability
- **Conclusion: Pass**
- **Rationale:** Startup/config/test instructions exist and are largely consistent with entrypoint/env usage.
- **Evidence:** `README.md:5`, `README.md:13`, `README.md:38`, `README.md:66`, `cmd/main.go:21`, `cmd/main.go:25`, `cmd/main.go:53`, `internal/store/seed.go:26`
- **Note:** `run_tests.sh` requires Docker while README also provides direct `go test` commands.
- **Evidence:** `run_tests.sh:4`, `README.md:47`

#### 1.2 Material deviation from Prompt
- **Conclusion: Partial Pass**
- **Rationale:** Most prompt flows are implemented; however, issued-credit persistence enforces one credit per member per rule version, which materially narrows expected transaction-based credit behavior.
- **Evidence:** `internal/store/migrations.go:144`, `internal/store/store.go:65`, `API_tests/credit_test.go:47`

### 2. Delivery Completeness

#### 2.1 Core explicit requirements coverage
- **Conclusion: Partial Pass**
- **Rationale:** Authentication, RBAC, member CRUD/import-export, reviews moderation/appeals, finance workflows, MDM imports, feature flags, and audit retention are present; credit issuance model conflict remains material.
- **Evidence:** `internal/handlers/handler.go:62`, `internal/services/auth_service.go:53`, `internal/handlers/members_admin.go:123`, `internal/services/review_service.go:57`, `internal/services/finance_service.go:42`, `internal/services/mdm_service.go:83`, `internal/services/flag_service.go:21`, `internal/store/migrations.go:188`, `internal/store/migrations.go:144`

#### 2.2 End-to-end 0->1 deliverable (not demo fragment)
- **Conclusion: Pass**
- **Rationale:** Complete project structure with entrypoint, persistence, templates, middleware, tests, and docs.
- **Evidence:** `cmd/main.go:20`, `internal/store/migrations.go:5`, `views/layouts/main.html:1`, `README.md:1`, `unit_tests/test_helpers_test.go:10`, `API_tests/test_helpers_test.go:23`

### 3. Engineering and Architecture Quality

#### 3.1 Structure and module decomposition
- **Conclusion: Pass**
- **Rationale:** Reasonable decomposition across handlers/services/store/middleware/models with clear routing and persistence boundaries.
- **Evidence:** `internal/handlers/handler.go:12`, `internal/services/finance_service.go:13`, `internal/store/store_finance.go:10`, `internal/middleware/auth.go:12`

#### 3.2 Maintainability/extensibility
- **Conclusion: Partial Pass**
- **Rationale:** Generally maintainable; however, some business constraints are encoded in schema in a way that reduces extensibility (credit issuance uniqueness constraint).
- **Evidence:** `internal/store/migrations.go:134`, `internal/store/migrations.go:144`, `internal/services/credit_service.go:56`

### 4. Engineering Details and Professionalism

#### 4.1 Error handling, logging, validation, API design
- **Conclusion: Partial Pass**
- **Rationale:** Standardized API error shape and structured logging exist; notable validation/security gaps remain (file type validation by extension only, weak date-format validation on some finance fields).
- **Evidence:** `internal/handlers/helpers_errors.go:32`, `cmd/main.go:67`, `internal/services/review_service.go:85`, `internal/services/finance_service.go:25`, `internal/services/finance_service.go:21`

#### 4.2 Product-like organization vs demo shape
- **Conclusion: Pass**
- **Rationale:** Includes role-aware UI, guarded routes, persistence, audit middleware, and multiple test layers.
- **Evidence:** `views/layouts/main.html:18`, `internal/middleware/auth.go:26`, `internal/middleware/audit.go:18`, `API_tests/auth_test.go:16`

### 5. Prompt Understanding and Requirement Fit

#### 5.1 Business goal and constraint fidelity
- **Conclusion: Partial Pass**
- **Rationale:** Strong alignment overall on offline HTMX flows, local auth, RBAC, moderation, finance, MDM, and feature flags; significant mismatch on transaction-level credit issuance flexibility.
- **Evidence:** `README.md:3`, `internal/handlers/handler.go:96`, `internal/handlers/handler.go:91`, `internal/store/migrations.go:144`

### 6. Aesthetics (frontend/full-stack)

#### 6.1 Visual and interaction quality
- **Conclusion: Pass**
- **Rationale:** Consistent layout hierarchy, spacing, responsive grids, and HTMX form feedback patterns; no major static visual defects found.
- **Evidence:** `views/layouts/main.html:12`, `views/budgets.html:1`, `views/recruiting.html:6`, `static/css/app.css:1`, `views/layouts/main.html:90`
- **Manual verification note:** interactive polish/visual regressions across browsers are **Cannot Confirm Statistically**.

## 5. Issues / Suggestions (Severity-Rated)

### Blocker / High

1) **Severity: High**
- **Title:** Credit issuance model blocks repeated transactions under same active rule version
- **Conclusion:** Fail
- **Evidence:** `internal/store/migrations.go:144`, `internal/store/store.go:65`, `API_tests/credit_test.go:47`
- **Impact:** Real operations can only issue one credit per member per rule version; subsequent valid transactions conflict, which deviates from transaction-date driven issuance expectations.
- **Minimum actionable fix:** Replace `UNIQUE(member_id, rule_version_id)` with a transaction-level uniqueness model (e.g., unique per `member_id + transaction_ref/date + source`) and persist explicit transaction identity used for immutability.

### Medium

2) **Severity: Medium**
- **Title:** Uploaded image validation trusts filename extension only
- **Conclusion:** Partial Fail
- **Evidence:** `internal/services/review_service.go:85`, `internal/handlers/helpers_files.go:14`
- **Impact:** Non-image payloads renamed as `.jpg/.png` can be accepted and served from static paths, raising content-safety risk.
- **Minimum actionable fix:** Validate MIME/content signature (magic bytes) server-side before write; optionally re-encode images.

3) **Severity: Medium**
- **Title:** Admin password reset does not verify target user existence
- **Conclusion:** Partial Fail
- **Evidence:** `internal/handlers/auth_pages.go:75`, `internal/services/auth_service.go:140`, `internal/store/store_auth.go:82`
- **Impact:** API may return success for non-existent `user_id` (silent no-op), weakening operational auditability and admin UX reliability.
- **Minimum actionable fix:** After update, verify rows affected > 0 and return `404 not_found` when target does not exist.

4) **Severity: Medium**
- **Title:** Finance period fields are weakly validated
- **Conclusion:** Partial Fail
- **Evidence:** `internal/services/finance_service.go:25`, `internal/services/finance_service.go:21`, `internal/handlers/finance.go:37`
- **Impact:** Invalid `period_start` formats can enter persisted budgets, reducing data quality for reporting/projections.
- **Minimum actionable fix:** Enforce strict format validation (`YYYY-MM` for monthly, `YYYY-Q[1-4]` or equivalent for quarterly) before insert.

### Low

5) **Severity: Low**
- **Title:** Test runner script is Docker-coupled despite direct Go test docs
- **Conclusion:** Partial Fail
- **Evidence:** `run_tests.sh:4`, `README.md:47`
- **Impact:** Contributors without Docker cannot use default script even though native commands exist.
- **Minimum actionable fix:** Make `run_tests.sh` run native `go test` by default and keep Docker variant as optional.

## 6. Security Review Summary

- **Authentication entry points: Partial Pass**
  - Evidence: `internal/handlers/auth_pages.go:32`, `internal/services/auth_service.go:61`, `internal/services/auth_service.go:53`, `internal/services/auth_service.go:67`
  - Reasoning: local auth, bcrypt, lockout, forced-change are implemented; some operational edge cases remain (admin reset target existence handling).

- **Route-level authorization: Pass**
  - Evidence: `internal/handlers/handler.go:71`, `internal/handlers/handler.go:88`, `internal/handlers/handler.go:90`, `internal/handlers/handler.go:113`
  - Reasoning: protected route groups and role gates are consistently applied.

- **Object-level authorization: Partial Pass**
  - Evidence: `internal/handlers/finance.go:54`, `internal/store/store_finance.go:156`, `internal/handlers/members_admin.go:56`, `internal/handlers/reviews_credit.go:29`
  - Reasoning: many object-scope checks exist; no critical bypass found statically, but coverage is uneven across all endpoints.

- **Function-level authorization: Pass**
  - Evidence: `internal/services/finance_service.go:107`, `internal/services/finance_service.go:71`, `internal/handlers/reviews_credit.go:239`
  - Reasoning: business-critical actions (budget approvals, >10% change rules, feature-gated credit issuing) enforce role/function constraints.

- **Tenant/user data isolation: Partial Pass**
  - Evidence: `internal/middleware/auth.go:66`, `internal/handlers/helpers.go:13`, `internal/handlers/finance.go:167`, `internal/handlers/pages.go:88`
  - Reasoning: club scoping is broadly enforced; statically no direct cross-club bypass identified, but some endpoints rely on handler-level discipline.

- **Admin/internal/debug protection: Pass**
  - Evidence: `internal/handlers/handler.go:81`, `internal/handlers/handler.go:82`, `internal/handlers/handler.go:90`
  - Reasoning: admin surfaces are role-guarded; no exposed debug route found.

## 7. Tests and Logging Review

- **Unit tests: Pass**
  - Evidence: `unit_tests/auth_test.go:11`, `unit_tests/finance_test.go:10`, `unit_tests/reviews_test.go:12`, `unit_tests/mdm_test.go:10`, `unit_tests/credit_test.go:9`
  - Notes: good domain validation coverage, including auth lockout, budget workflow, review constraints, MDM rules, credit rule date selection.

- **API/integration tests: Pass**
  - Evidence: `API_tests/auth_test.go:16`, `API_tests/members_test.go:117`, `API_tests/reviews_test.go:229`, `API_tests/finance_test.go:235`, `API_tests/clubs_users_test.go:16`
  - Notes: includes 401/403/404/conflict and HTMX-oriented responses across major modules.

- **Logging categories / observability: Partial Pass**
  - Evidence: `cmd/main.go:67`, `cmd/main.go:85`, `internal/handlers/helpers_errors.go:22`, `internal/store/store_audit.go:9`
  - Notes: structured categories exist; no centralized severity taxonomy beyond formatted logs.

- **Sensitive-data leakage risk in logs/responses: Partial Pass**
  - Evidence: `internal/middleware/audit.go:57`, `internal/middleware/audit.go:132`, `internal/handlers/helpers_errors.go:21`, `API_tests/members_test.go:287`
  - Notes: audit payload redaction is implemented and tested; file upload type checks remain weak (see issue #2).

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview
- Unit tests exist under `unit_tests` and API/integration tests under `API_tests`; optional browser tests in `e2e`.
- Frameworks: Go `testing` (`unit_tests`, `API_tests`) and Playwright (`e2e`).
- Test entry points documented in README.
- Evidence: `README.md:38`, `README.md:47`, `README.md:51`, `unit_tests/test_helpers_test.go:10`, `API_tests/test_helpers_test.go:23`, `e2e/package.json:6`

### 8.2 Coverage Mapping Table

| Requirement / Risk Point | Mapped Test Case(s) | Key Assertion / Fixture / Mock | Coverage Assessment | Gap | Minimum Test Addition |
|---|---|---|---|---|---|
| Auth: 12+ char passwords, lockout, forced change | `unit_tests/auth_test.go:11`, `API_tests/auth_test.go:69`, `API_tests/auth_test.go:131` | Short password rejected and lockout/forced-change behavior asserted | sufficient | None major | Add explicit 15-minute unlock boundary test with controlled clock |
| Session inactivity + sliding refresh | `unit_tests/auth_test.go:49`, `unit_tests/auth_test.go:156` | `SetNowFunc` and expired-session DB mutation checks | basically covered | No full end-to-end inactivity scenario | Add API-level test for idle timeout redirect/401 after TTL |
| Route RBAC (401/403) | `API_tests/auth_test.go:16`, `API_tests/auth_test.go:236`, `API_tests/finance_test.go:77` | Unauth and role-denied endpoints return expected codes | sufficient | Limited breadth for all admin endpoints | Add table-driven sweep over all admin-only routes |
| Object-level auth (club scope) | `API_tests/clubs_users_test.go:16`, `API_tests/members_test.go:48`, `API_tests/reviews_test.go:105` | Cross-club and ownership denials verified | basically covered | Not exhaustive across every mutable endpoint | Add tests for `/api/budgets/:id/spend` and `/api/members/:id` cross-club denials |
| Members CSV import/export + row limit/errors | `API_tests/members_test.go:117`, `API_tests/members_test.go:155`, `API_tests/members_test.go:244` | Invalid JSON, downloadable report link, and 5000-row cap assertions | sufficient | No malformed CSV quoting edge-case coverage | Add malformed CSV line and partial-row edge tests |
| Reviews validation and moderation/appeal | `unit_tests/reviews_test.go:12`, `API_tests/reviews_test.go:55`, `API_tests/reviews_test.go:20` | stars/comment/file constraints + moderation role checks | basically covered | No explicit test for 7-day appeal after hidden timestamp at API layer | Add API test setting `hidden_at` beyond 7 days |
| Finance >10% approval workflow | `unit_tests/finance_test.go:10`, `API_tests/finance_test.go:235` | organizer->admin flow and self-approval denial | sufficient | No explicit concurrent duplicate-approval race test | Add transaction/race resilience test for change-review endpoint |
| Credit rule-by-date and immutability | `unit_tests/credit_test.go:67`, `unit_tests/credit_test.go:9`, `API_tests/credit_test.go:47` | historical date selection + duplicate issuance conflict | insufficient | Tests encode a restrictive uniqueness behavior that conflicts with prompt intent | Add tests for multiple transactions per member under same rule version with distinct transaction refs |
| Audit logging + redaction + append-only retention | `API_tests/finance_test.go:148`, `API_tests/auth_test.go:390`, `unit_tests/flags_audit_test.go:84` | audit insertion, auth payload redaction, append-only trigger behavior | sufficient | Coverage for all critical mutation paths not exhaustive | Add assertions for `/api/credit_rules` and `/api/auth/admin-reset` audit entries |

### 8.3 Security Coverage Audit
- **Authentication: basically covered** by unit/API tests for lockout, password policy, forced change, CSRF.
  - Evidence: `unit_tests/auth_test.go:11`, `API_tests/auth_test.go:39`
- **Route authorization: basically covered** with representative 401/403 checks.
  - Evidence: `API_tests/auth_test.go:16`, `API_tests/finance_test.go:77`
- **Object-level authorization: insufficient-to-basically-covered** (good critical examples, not exhaustive).
  - Evidence: `API_tests/clubs_users_test.go:16`, `API_tests/reviews_test.go:105`
- **Tenant/data isolation: basically covered** for key club scope flows.
  - Evidence: `API_tests/members_test.go:48`, `API_tests/finance_test.go:108`
- **Admin/internal protection: basically covered** for selected routes.
  - Evidence: `API_tests/auth_test.go:236`
- Severe defects could still remain undetected in untested endpoint combinations (especially less-exercised object-level cases).

### 8.4 Final Coverage Judgment
- **Partial Pass**
- Major happy-path and many security/error-path cases are covered statically.
- Uncovered/undercovered risks remain: full endpoint-by-endpoint object-level auth matrix and prompt-aligned credit issuance semantics.

## 9. Final Notes
- This assessment is strictly static and evidence-based.
- No runtime claims were made without code/test evidence.
- High-priority remediation should start with the credit issuance data model mismatch, then upload validation hardening.
