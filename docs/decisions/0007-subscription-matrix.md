# ADR 0007 – Subscription Matrix (Tier × Interval)

## Status
Accepted

## Context
We need flexible billing periods while keeping a clean mapping to NowPayments subscription plans.

## Decision
- Global **Intervals** (monthly, quarterly…).
- **Plans** = Cartesian product of active Tiers × active Intervals.
- The system enforces that every combination has a configured NowPayments plan.
- Incomplete matrix → permanent red banner in admin.
- Upgrade flow is a multi-step modal that handles wallet balance, deposit via QR, then subscription creation, then status webhook, with timeout handling.
- Every webhook payload is stored raw for audit.

## Consequences
- Admin must configure the full matrix before the banner disappears.
- Exactly N×M plans exist; no more, no less for the active set.
- Subscription attempts are first-class entities for debugging.
