---
name: nowpayments-billing
description: Implement Comics Galore crypto billing with NowPayments, plan matrix, webhooks, and provider abstraction. Use when building subscriptions, deposits, plans, or payment admin tools.
---

# Billing – NowPayments

## V1 policy

- Single live provider NowPayments.
- Code against `PaymentsProvider` interface (ADR 0013).
- No Stripe or fiat processors.

## Domain

- Tiers × Intervals = plans matrix (schema in v1; few plans configured is OK).
- Incomplete matrix may show admin red banner.
- Store `provider`, external IDs, and full webhook payloads.

## Flow (conceptual)

Estimate → ensure balance/deposit if required → create subscription at provider → wait for status webhook → activate internal subscription and tier.

## CreatePlan (auto-link)

The `nowpayments.Provider.CreatePlan` method (shared package `backend/nowpayments`) calls `POST /v1/subscriptions/plans`.
**Never hard-code raw `strconv.Atoi` on period strings** — periods like "month" are not numeric.

### Request body contract
```json
{ "title": "Plan Name", "price_amount": 9.99, "price_currency": "usd", "period": 30 }
```
- `title` — plan name (string, required)
- `price_amount` — USD price (float, required)
- `price_currency` — always `"usd"`
- `period` — **integer days** (number, required), not a string

### Period mapping
Use a `periodToDays()` helper to convert period strings to day counts:
| Period | Days |
|--------|------|
| `day` | 1 |
| `week` | 7 |
| `month` | 30 |
| `quarter` | 90 |
| `semester` | 180 |
| `year` | 365 |

### Auto-link endpoint (backend)

`POST /admin/plans/link/:id` (tiers service, admin-only):
1. Fetch plan from DB → get `name`, `price_usd_cents`, `interval`
2. Map interval string via `intervalToPeriod()` ("monthly"→"month"), then `periodToDays()` ("month"→30)
3. Call `nowpayments.Provider.CreatePlan()` directly (shared package) → NowPayments API → get back `provider_plan_id`
4. Store `provider_plan_id` on the plan in DB
5. Return `{ provider_plan_id, plan_name }`

The tiers service constructs its own `nowpayments.Provider` from secrets and calls it directly — it does **not** route through the `billing` service (see ADR `0016-service-communication.md`). The callback URL is built with `nowpayments.BuildCallbackURL(host, ngrokURL, path)`.

### Lock guard

`PATCH /admin/plans/:id` refuses update if the plan already has a non-empty `provider_plan_id`. Returns error: `"plan already linked to provider plan ID: {id}"`. Prevents accidental re-linking.

### Admin link wizard (frontend)

Modal in `frontend-admin/src/lib/components/NowPaymentsLinkWizard.svelte`:
- **Manual mode**: admin types NowPayments plan IDs into a sortable table, clicks "Link" per row
- **Automatic mode**: admin clicks "Link All" → calls `POST /admin/plans/:id/auto-link` for each unlinked plan → shows progress bar (% complete) → log lines per result
- Plans display name includes interval (except "Free" tier): e.g. "Bronze - Monthly"
- Sortable table headers: Plan name, Interval, Price
- Free tier excluded from interval display
- Submit button uses standardized pattern: `autoLinking ? 'Linking...' : 'Link All'`

## Reliability

- Idempotent webhook handling.
- Subscription attempts for success/fail/timeout.
- Never trust the client for tier changes.

## Later

- Second provider adapter, coupons, rich past-due tooling — not v1 blockers.

## References

- ADR `0005-nowpayments.md`, `0007-subscription-matrix.md`, `0013-payments-provider-abstraction.md`
