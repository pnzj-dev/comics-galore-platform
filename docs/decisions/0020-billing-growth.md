# ADR 0020 – Billing Growth (Coupons, Manual Grants, Reconciliation)

## Status
Accepted

## Context
v1 billing covers the NowPayments checkout path, webhooks, deposits, subscriptions, and a red banner for an incomplete plan matrix. Growth requires coupons, manual comp/revoke, past-due tracking, and richer revenue breakdown — all behind the existing `PaymentsProvider` abstraction (ADR 0013).

## Decision

### Coupons
- `coupons(id, code UNIQUE, percent_off, tier, max_uses, used, expires_at, created_at)` in the billing service.
- `GET/POST /admin/coupons`. Checkout accepts a `coupon` code; the discounted amount is displayed and stored with the subscription/deposit.

### Manual grant / extend / revoke
- `POST /admin/subscriptions/:id/grant` (comp subscription: tier, duration), `POST /admin/subscriptions/:id/revoke` (admin) — write directly to `subscriptions` and update `users.tier` via `auth.SetUserTier` (ADR 0016).

### Past-due & reconciliation
- `GET /admin/payments/past-due` — failed/expired recent payments.
- `POST /admin/subscriptions/:id/sync` — force re-sync a subscription's status from NowPayments via the provider.

### Revenue breakdown
- Extend `GET /admin/billing-stats` with `revenue_by_tier_interval` (tier × interval).

### Second provider
- Already enabled by the `PaymentsProvider` interface; a second provider is a new adapter + config, not a rewrite of checkout or the plan matrix.

## Consequences
- All new tables live in `billingdb`; no new service.
- Manual grants and provider sync stay server-side; the client never mutates tier directly.
- Coupons remain simple (percentage, optional tier scope) in the first pass.

## References
- ADR `0005-nowpayments.md`, `0007-subscription-matrix.md`, `0013-payments-provider-abstraction.md`, `0016-service-communication.md`
