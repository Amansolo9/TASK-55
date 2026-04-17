<!-- project-type: fullstack -->

# ClubOps Governance & Finance Portal

**Project type:** `fullstack`

Go (Fiber) + HTMX + SQLite implementation for multi-site membership governance, moderation, and finance workflows. Everything (app, database, tests, e2e) runs inside Docker — no host-level Go/Node install is required.

---

## One-click start

```bash
docker-compose up
```

> The `docker-compose up` command builds the image on first run and starts the app container. Use `docker-compose up --build` to force a rebuild after code changes.

App URL: `http://localhost:8080`

All required environment variables are preset in `docker-compose.yml`:

| Variable | Purpose | Default in compose |
|---|---|---|
| `APP_ENCRYPTION_KEY` | AES-GCM key for PII columns | `local-dev-encryption-key-please-change` |
| `APP_DB_PATH` | SQLite file path inside container | `/data/fullstack.db` |
| `APP_BOOTSTRAP_ADMIN_PASSWORD` | Seeds the `admin` account | `ChangeMe12345!-replace-for-real-use` |
| `APP_SEED_DEMO_USERS` | Creates demo accounts for every role | `true` |
| `APP_DEMO_USER_PASSWORD` | Password for seeded demo accounts | `DemoPass12345!` |
| `APP_PORT` | Listen port | `8080` |

Persistence: the `sqlite_data` Docker volume holds the database; `./static/uploads` is bind-mounted for uploaded review/avatar images.

---

## Demo credentials (all roles)

When the container starts with the defaults above, the following accounts are seeded and usable immediately via `POST /login` or the `/login` UI:

| Role | Username | Password |
|---|---|---|
| `admin` | `admin` | `ChangeMe12345!-replace-for-real-use` |
| `organizer` | `organizer_demo` | `DemoPass12345!` |
| `team_lead` | `teamlead_demo` | `DemoPass12345!` |
| `member` | `member_demo` | `DemoPass12345!` |

The `admin` account is flagged for forced password change on first login (redirects to `/change-password` until rotated). The three demo accounts are pre-rotated for direct use.

Registration via public `POST /register` always creates `member`-role accounts only. Elevated roles must be assigned through admin workflows.

---

## How to verify the system works

After `docker-compose up`, run these checks from a host shell to confirm the full stack is healthy.

### 1. Landing / login page reachable

```bash
curl -sI http://localhost:8080/login | head -1
```

Expected: `HTTP/1.1 200 OK`

### 2. Authenticate as admin and capture session + CSRF cookies

```bash
curl -s -c cookies.txt \
  -d "username=admin&password=ChangeMe12345!-replace-for-real-use" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -o /dev/null -w "%{http_code}\n" \
  http://localhost:8080/login
```

Expected: `302`. The file `cookies.txt` must now contain both `session_token` and `csrf_token` entries.

### 3. Authenticated dashboard is served

```bash
curl -s -b cookies.txt -o /dev/null -w "%{http_code}\n" http://localhost:8080/
```

Expected: `200` (redirects to `/change-password` until admin rotates — that is also a 200 page).

### 4. API smoke call — list budgets (scoped)

```bash
CSRF=$(grep csrf_token cookies.txt | awk '{print $NF}')
curl -s -b cookies.txt -H "X-CSRF-Token: $CSRF" \
  -o /dev/null -w "%{http_code}\n" \
  http://localhost:8080/partials/budgets/list
```

Expected: `200`.

### 5. Feature-flag evaluate endpoint (admin only)

```bash
curl -s -b cookies.txt -H "X-CSRF-Token: $CSRF" \
  http://localhost:8080/api/flags/evaluate/credit_engine_v2
```

Expected JSON: `{"enabled":...,"flag":"credit_engine_v2"}`.

### 6. Non-admin role is blocked from admin surfaces

```bash
curl -s -c member.txt \
  -d "username=member_demo&password=DemoPass12345!" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -o /dev/null -w "%{http_code}\n" \
  http://localhost:8080/login   # expect 302

CSRF_M=$(grep csrf_token member.txt | awk '{print $NF}')
curl -s -b member.txt -H "X-CSRF-Token: $CSRF_M" \
  -o /dev/null -w "%{http_code}\n" \
  http://localhost:8080/api/flags/evaluate/credit_engine_v2
```

Expected: `403` (members cannot evaluate feature flags).

If all six checks return the expected status/response, the system is working end-to-end: authentication, CSRF, RBAC, session cookies, HTMX partials, API JSON endpoints, and DB-backed reads.

---

## Test commands (Docker-contained)

All tests run inside the official Go container — no host Go install required:

```bash
./run_tests.sh
```

This runs three layers inside Docker (no host toolchain required):

| Layer | Container | Location |
|---|---|---|
| Backend unit | `golang:1.26.1-alpine` | `./unit_tests` (services, store, domain rules) |
| Backend API (no-mock HTTP) | `golang:1.26.1-alpine` | `./API_tests` (Fiber app on `:memory:` SQLite, real middleware + routes) |
| Frontend unit (templates) | `node:22-alpine` | `./frontend_tests` (Vitest + cheerio parsing of HTML templates and partials) |

The API tests exercise the real route tree — no mocks, no stubs. The frontend suite parses `repo/views/*.html` and `repo/views/partials/*.html` and asserts structural invariants (form actions, required inputs, HTMX attributes, CSRF wiring, role-gated nav links).

---

## Architecture at a glance

| Layer | Location | Notes |
|---|---|---|
| HTTP entry | `repo/cmd/main.go` | Boots Fiber, middleware, services |
| Routes | `repo/internal/handlers/handler.go` `RegisterRoutes` | Public + `secured` + `/api` groups |
| Middleware | `repo/internal/middleware/` | `AttachCurrentUser`, `RequireAuth(roles...)`, `CSRFProtection`, `AuditTrail` |
| Services | `repo/internal/services/` | Auth, Finance, Credit, Review, MDM, Flags, Crypto, Audit |
| Store | `repo/internal/store/` | SQLite DAO layer + migrations + seed |
| Views | `repo/views/*.html` | Server-rendered templates + HTMX partials |

Key workflows: budget approval (>10% request → cross-role review), member PII encryption (AES-GCM), review moderation + appeal, MDM dimension/sales-fact import, feature-flag gating per role/club rollout.

---

## Security controls

- bcrypt cost `12`; 5 failed logins → 15-minute lockout.
- Session timeout `30m` with sliding refresh; `no-store` on authenticated responses.
- CSRF protection on all authenticated mutations (`csrf_token` cookie + `X-CSRF-Token` header).
- Club-scoped access for non-admin operators; object-level checks on budget/review mutations.
- App-layer AES-GCM encryption for `members.email`, `members.phone`, `members.custom_fields` (`enc:v1:` prefix; legacy plaintext lazily migrated on read/write).
- Audit logs: append-only, 2-year retention, path-allowlisted; sensitive fields (`email`, `phone`, `custom_fields`, `comment`, passwords/tokens) redacted.
- Member CSV export emits an explicit audit entry (action, row count, club scope).
- Frontend vendor assets served locally from `/static/vendor` (offline-capable).

## API error schema

```json
{ "error": "...", "error_code": "...", "message": "..." }
```

`error_code` values: `validation_error`, `conflict`, `forbidden`, `not_found`, `bad_request`.

---

## Additional workflows

- Reviews must reference a fulfilled order/service record (`fulfilled_order_id`).
- Region hierarchy management at `/regions` (admin/organizer).
- Dimension and sales-fact imports at `/mdm` (admin/organizer).
- Admin user provisioning at `/users`.
- Club avatars uploaded under `./static/uploads/avatars`.
- Budget spend updates recorded from `/budgets` drive execution % and threshold alerts.
- Member CSV import returns an inline downloadable error-report link on validation failure.
