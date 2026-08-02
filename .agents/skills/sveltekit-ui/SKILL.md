---
name: sveltekit-ui
description: Build Comics Galore SvelteKit web UI with shadcn-svelte, Tailwind, Superforms, and Zod. Use when creating pages, components, forms, or admin datalists on the web client.
---

# SvelteKit UI – Comics Galore

## Stack

- SvelteKit + Tailwind + shadcn-svelte (official CLIs)
- Superforms + Zod for forms
- TanStack Table (Svelte) for admin tables
- Charts only where needed (ECharts or Chart.js)

## Rules

- Do not introduce React or shadcn/ui (React).
- Prefer shared code under `packages/ui` when desktop will consume the same components.
- Public routes lean on SSR and minimal JS.
- Admin routes may be richer.

## V1 screens priority

Public home/detail/reader, uploader manual create, admin pending queue and basic lists, auth flows, Terms/Privacy stubs.

## i18n

- Message catalogs; default `en`.
- Respect `users.ui_locale` and enabled locales from settings.
- Comic forms require content language.

## Patterns

- Resolve image URLs via media helper (mode `direct` in v1).
- Upgrade CTAs open the plans modal flow.
- Empty and quota-blocked states must be explicit.

## References

- `docs/ui.md`, `docs/v1-scope.md`, ADR `0002-sveltekit.md`
