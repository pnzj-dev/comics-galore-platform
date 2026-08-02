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

## Reliability

- Idempotent webhook handling.
- Subscription attempts for success/fail/timeout.
- Never trust the client for tier changes.

## Later

- Second provider adapter, coupons, rich past-due tooling — not v1 blockers.

## References

- ADR `0005-nowpayments.md`, `0007-subscription-matrix.md`, `0013-payments-provider-abstraction.md`
