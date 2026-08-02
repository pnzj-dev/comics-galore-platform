# ADR 0013 – Single Crypto Provider (v1) + Multi-Provider-Ready Design

## Status
Accepted

## Context
Comics Galore monetises exclusively via cryptocurrency. NowPayments is the chosen processor. Supporting multiple processors in v1 would double webhook, reconciliation and support complexity while the rest of the product is already large.

## Decision

### v1
- **Only NowPayments** is implemented and enabled in production.
- User-facing copy speaks of “Pay with crypto”, not a single brand-heavy dependency where avoidable.

### Architecture (mandatory)
- Introduce a **PaymentsProvider** interface (or equivalent Go interface) used by the subscription service.
- v1 has a single implementation: `NowPaymentsProvider`.
- Domain tables already store provider-agnostic fields plus provider-specific IDs:
  - `provider` (e.g. `nowpayments`)
  - external payment / subscription / invoice IDs
  - raw `webhook_events.payload` (already required)
- Webhooks enter through a generic ingress, are verified by the active provider adapter, then normalized into internal events.
- Admin may expose a “Payment provider” setting even if only one option exists in v1.

### Future
- A second provider is added as another adapter + config, not a rewrite of the upgrade modal or plan matrix.
- Plan matrix (tiers × intervals) remains owned by Comics Galore; providers only execute charge/subscribe/deposit flows.

## Consequences
- Faster, safer v1 delivery.
- No hard lock-in at the application boundary.
- Switching or dual-running providers later is feasible without changing public UX flows.
