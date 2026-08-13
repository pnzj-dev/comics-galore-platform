# ADR 0016 – Cross-Service Communication & Shared NowPayments Package

## Status
Accepted

## Context
Comics Galore runs on Encore, which gives **each service its own PostgreSQL database**. Early billing code read `users`, `plans`, and `tiers` directly from `auth` and `tiers` databases, which breaks at runtime (tables don't exist in the billing database) and couples services to each other's schemas.

Separately, the NowPayments HTTP client lived inside the `billing` service, but the `tiers` service also needed it (`CreatePlan` for the auto-link wizard). This created an awkward `tiers` → `billing` import that would have become a **cycle** once `billing` needed `tiers` data.

We also needed to fix **when** a user's NowPayments customer (`sub_partner_id`) is created, so we don't create "ghost" customers for unverified users.

## Decision

### 1. No direct cross-service table access
Services **never** read or write another service's tables. Cross-service communication happens only through:

- **Typed private API calls** — `//encore:api private` endpoints callable from other services, or
- **Pub/Sub events** for asynchronous work.

The concrete v1 endpoints introduced by this decision:

| Endpoint | Service | Consumer | Purpose |
|----------|---------|----------|---------|
| `POST /auth/ensure-sub-partner-id` | auth | billing | Get-or-create the user's NowPayments `sub_partner_id` |
| `POST /auth/set-user-tier` | auth | billing | Set a user's tier after subscription webhook activation |
| `GET /internal/plans/:id` | tiers | billing | Fetch plan details (provider plan id, interval, tier name, price) |

### 2. Shared `nowpayments` package
The NowPayments REST client, its request/response types, the `PaymentsProvider` interface, and the `BuildCallbackURL` helper moved to a **shared, non-service package**: `backend/nowpayments/`.

It is consumed by `auth` (customer creation), `billing` (estimate/balance/subscribe/deposit), and `tiers` (plan auto-link). This removes the `tiers` → `billing` import and the potential cycle.

### 3. `sub_partner_id` ownership & lifecycle
- `sub_partner_id` lives on the `users` table and is **owned by the `auth` service**.
- Provisioned **eagerly** on email verification (`VerifyEmail` calls `ensureSubPartnerID` synchronously; failures are non-fatal).
- Provisioned **lazily** on demand via `auth.EnsureSubPartnerID` (called by `billing` at subscription/deposit/balance-check time) for users who never verified email.
- Saved atomically (`UPDATE ... WHERE sub_partner_id IS NULL`) to guard against concurrent creation.
- Enforced unique via a **partial index**: `CREATE UNIQUE INDEX ... ON users(sub_partner_id) WHERE sub_partner_id IS NOT NULL`.

## Consequences
- Billing no longer contains any SQL touching `users`, `plans`, or `tiers`; it calls `auth` and `tiers` through typed APIs.
- `tiers` no longer imports `billing`; `billing` imports `tiers` — the cycle is broken.
- Adding a second payment provider stays behind the `PaymentsProvider` interface in the shared package (see ADR 0013).
- The billing seed no longer writes to `users`; the `sub_partner_id` seed value moved to the `auth` seed (which owns the table).

## References
- ADR `0005-nowpayments.md`, `0013-payments-provider-abstraction.md`
