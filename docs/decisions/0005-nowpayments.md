# ADR 0005 – Payments: NowPayments (Crypto Only)

## Status
Accepted (see also ADR 0013)

## Context
The product requires subscription payments in cryptocurrency. Fiat processors (e.g. Stripe) are out of scope.

## Decision
- **NowPayments** is the sole payment processor **implemented in v1**.
- Stripe and other fiat processors are forbidden.
- Backend creates invoices / payment requests, stores NowPayments IDs, processes webhooks idempotently.
- Subscription status and tier changes happen only after confirmed webhook events (or explicit admin grant).

## Consequences
- All payment-related orchestration lives behind a provider abstraction (ADR 0013).
- Webhook endpoint verifies signatures and is idempotent; raw payloads are stored for audit.
- Frontend never handles private keys; it only displays addresses / QR codes or redirects as instructed by the backend.
