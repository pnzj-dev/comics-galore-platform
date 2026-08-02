---
name: comics-galore-scope
description: Enforce Comics Galore v1 scope freeze and build priority. Use when implementing features, planning phases, choosing IN vs LATER work, or when scope might expand mid-build.
---

# Comics Galore – Scope Gate

Before implementing anything, read `docs/v1-scope.md`.

## Priority order

1. **IN (v1)** — ship these first
2. **SOON (v1.1)** — only after v1 exit checklist passes
3. **LATER** — do not implement before v1 is demoable

## On conflict

`docs/v1-scope.md` wins over the full vision in `product.md` / `roadmap.md`.

## V1 spine (never skip)

- Encore + SvelteKit web (not desktop-first)
- Auth, roles, bootstrap admin check
- Manual comic create, presigned S3, `POST /comics`, `pending_review`
- Public browse + basic reader + Continue Reading
- Like/favorite, server-side download quota
- NowPayments path behind provider interface
- Minimal admin (approve/reject, users, basic KPIs)

## Allowed hooks without full UI

- Asset `kind` enum
- `PaymentsProvider` interface with only NowPayments
- Image resolver with mode `direct` only
- Plan matrix schema with few configured plans

## Refuse

- Building Wails offline suite before web spine works
- AI moderation, DMs, full support desk, coupons, impersonation as v1 blockers
- Second payment provider implementation in v1

## Exit

Only advance to SOON when the checklist in `docs/v1-scope.md` is satisfied.
