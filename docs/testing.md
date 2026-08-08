# Testing Strategy — Comics Galore

## Layers

| Layer | Framework | Command | Coverage |
|-------|-----------|---------|----------|
| **Backend unit** | Go `testing` + Encore | `encore test ./...` | Service logic, auth, CRUD |
| **Frontend unit** | Vitest + @testing-library/svelte | `bun run test:unit` | Stores, API client, utils, components |
| **E2E** | Playwright | `bun run test:e2e` | Critical user flows |

## Running Tests

```bash
# Backend (requires Docker running)
cd backend && encore test ./...

# Frontend unit
cd frontend && bun run test:unit --run

# Frontend unit (watch mode)
cd frontend && bun run test:unit

# E2E
cd frontend && bun run test:e2e

# All frontend tests
cd frontend && bun run test
```

## CI Pipeline

GitHub Actions runs on every push and PR to `main`:
- `backend` job: runs `encore test ./...`
- `frontend` job: runs `bun run test:unit --run`
- `e2e` job: runs `bun run test:e2e` with headless Chromium

## Test Targets by Priority

### P0 — Critical Paths (must pass pre-commit)

**Auth:**
- Register with valid/invalid/duplicate credentials
- Login with valid/wrong credentials
- JWT contains correct tier
- Role-based access control

**Comics:**
- Create comic → `pending_review` status
- List only published comics
- Publish/reject by moderator
- Like/favorite toggles with correct counts

**Billing:**
- Plan estimate returns valid price
- Subscription creation is atomic
- Deposit creation returns pay address
- Webhook activates subscription

### P1 — Important (must pass pre-push)

**Reading:**
- Save progress → retrieve progress
- Continue reading returns incomplete items
- Download quota enforcement

**Upload:**
- Create session → presign URL → confirm part
- Session abort

**Moderation:**
- Audit log entries on approve/reject
- Admin role checks

### P2 — Coverage (target: 70% line coverage)

**Frontend components:**
- ComicCard renders with all props
- PlanGrid loads tiers and plans
- CheckoutModal screen transitions

**Edge cases:**
- Empty states (no comics, no users, no subscriptions)
- Deleted/expired sessions
- Invalid plan IDs
- Rate limiting on auth endpoints

## Subagent Protocol

After each code modification:
```
Main Agent → QA Subagents (parallel)
  ├─ Backend QA: runs `encore test ./...`, reports:
  │   ├─ Test file                                       │ Results
  │   ├─ backend/auth/auth_test.go                        PASS/FAIL
  │   ├─ backend/comics/comics_test.go                    PASS/FAIL
  │   ├─ backend/billing/billing_test.go                  PASS/FAIL
  │   ├─ backend/tiers/tiers_test.go                      PASS/FAIL
  │   ├─ backend/reading/reading_test.go                  PASS/FAIL
  │   └─ backend/upload/upload_test.go                    PASS/FAIL
  ├─ Frontend QA: runs `bun run test:unit --run`, reports:
  │   ├─ tests/unit/api/client.test.ts                    PASS/FAIL
  │   ├─ tests/unit/stores/auth.test.ts                   PASS/FAIL
  │   ├─ tests/unit/utils.test.ts                         PASS/FAIL
  │   ├─ tests/unit/components/Lightbox.test.ts           PASS/FAIL
  │   ├─ tests/unit/components/AgeGate.test.ts            PASS/FAIL
  │   ├─ tests/unit/components/LikeButton.test.ts         PASS/FAIL
  │   ├─ tests/unit/components/DislikeButton.test.ts       PASS/FAIL
  │   └─ tests/unit/components/FavoriteButton.test.ts     PASS/FAIL
  └─ E2E QA: runs `bun run test:e2e`, reports:
      ├─ tests/e2e/public.spec.ts                         PASS/FAIL
      ├─ tests/e2e/auth.spec.ts                           PASS/FAIL
      ├─ tests/e2e/auth-pages.spec.ts                     PASS/FAIL
      ├─ tests/e2e/navigation.spec.ts                     PASS/FAIL
      └─ tests/e2e/comic-detail.spec.ts                   PASS/FAIL

Main Agent ← collects results → only proceeds if all green
```

## Adding New Tests

When adding a new backend endpoint, create (or update) the corresponding `_test.go` file in the same package. Use `et.NewTestDatabase()` for isolated test databases and `auth.WithContext()` for authenticated test contexts.

When adding a new frontend component, create a test file in `tests/unit/components/` with render tests using `@testing-library/svelte`.

When adding a new page, add an E2E test in `tests/e2e/` verifying the page loads and key interactions work.
