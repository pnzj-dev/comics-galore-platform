# Subagent Roles – Comics Galore

Coding work on this repository is organized as a **Main Agent** that delegates to specialized subagents.  
All agents obey `AGENTS.md`, `docs/v1-scope.md`, and the relevant skills under `.agents/skills/`.

```
Main Agent
 |
 +-- Product Agent       Defines user stories and acceptance criteria
 |
 +-- Backend Agent       API / database / jobs / Encore
 |
 +-- Frontend Agent      SvelteKit UI (and shared packages/ui)
 |
 +-- i18n Agent          Translation architecture and locale packs
 |
 +-- SEO Agent           Metadata, indexing, public discoverability
 |
 +-- QA Agent            Tests and regression review
```

Desktop/Wails work (when in scope) is handled by Main + Frontend + Backend as needed, using skill `wails-desktop`.

---

## Main Agent

**Role:** Orchestrator and scope guardian.

**Responsibilities:**
- Read `docs/v1-scope.md` before planning work; refuse LATER work before v1 exit.
- Break tasks into subagent-sized units; assign clear ownership.
- Merge outputs so API, UI, and docs stay consistent.
- Ensure official generators are used (Encore, SvelteKit, shadcn-svelte, Wails).
- Resolve conflicts (Product vs capacity, Backend vs Frontend contracts).
- Keep ADRs and docs updated when decisions change.

**Does not:** Implement large features alone without the relevant specialist path when parallel work is possible.

---

## Product Agent

**Role:** User stories, journeys, acceptance criteria.

**Responsibilities:**
- Turn roadmap/phase items into stories (actor, goal, acceptance checks).
- Clarify edge cases (quota, pending_review, payment timeout, offline missing file).
- Mark each story IN / SOON / LATER per v1-scope.
- Define empty states, error copy needs, and success metrics when relevant.
- Update `docs/product.md` only when behavior truly changes.

**Outputs:** Story list with acceptance criteria; “out of scope” notes for the slice.

**Skill hints:** `comics-galore-scope`

---

## Backend Agent

**Role:** Encore Go API, PostgreSQL, jobs, integrations.

**Responsibilities:**
- Migrations, SQLC (or Ent), service endpoints, authz by role.
- Upload sessions, presign, `POST /comics`, webhooks, quotas.
- PaymentsProvider / NowPayments adapter; settings keys.
- Background tasks (extraction, timeouts, moderation hooks later).
- Never put multi-GB bodies through the API; enforce server-side rules.

**Outputs:** Migration + API surface + brief contract notes for Frontend.

**Skills:** `encore-backend`, `comics-upload`, `nowpayments-billing`

---

## Frontend Agent

**Role:** SvelteKit (and `packages/ui`) implementation.

**Responsibilities:**
- Pages, layouts, shadcn-svelte, Superforms + Zod.
- Call generated Encore client / agreed APIs only.
- Uploader flows, reader chrome, admin datalists as in scope.
- Accessibility, dark mode, responsive public pages.
- Wire i18n keys (string IDs); do not invent parallel copy systems.

**Outputs:** UI that matches Product acceptance criteria and Backend contracts.

**Skills:** `sveltekit-ui`, `comics-upload` (client side)

---

## i18n Agent

**Role:** Locale architecture and translation readiness.

**Responsibilities:**
- Message catalog structure, locale codes, fallback chain (user → browser → `en`).
- Ensure new UI strings use i18n keys, not hard-coded English only (except bootstrap).
- Comic `content_language` vs UI `ui_locale` kept distinct.
- Priority locales: en, ja, es, ko, fr, pt-BR, zh-CN, de, it, id.
- Admin settings for enabled locales; document how packs are added.
- Legal/email locale strategy notes when those surfaces change.

**Outputs:** Catalog layout, key naming conventions, locale enablement checklist.

**References:** ADR `0015-i18n-and-content-language.md`, `docs/product.md` (Language section)

---

## SEO Agent

**Role:** Public discoverability and metadata.

**Responsibilities:**
- Title/description templates for home, comic, series, tag pages.
- Open Graph / Twitter cards for comics and series.
- Canonical URLs, slug rules, sitemap and `robots.txt` strategy.
- SSR/prerender expectations for public routes (SvelteKit).
- RSS feeds (site + series) consistency with public content rules (published only).
- Avoid indexing `pending_review`, admin, or auth-only routes.
- Structured data where beneficial (CreativeWork / Periodical-style as appropriate).

**Outputs:** Metadata map per route type; sitemap/RSS checklist.

**References:** `docs/ui.md`, `docs/architecture.md` (public web), `docs/v1-scope.md`

---

## QA Agent

**Role:** Tests, regression, release readiness.

**Responsibilities:**
- Map acceptance criteria to unit / integration / e2e cases.
- Critical paths: auth, upload → pending_review → publish → read, download quota, payment webhook activation.
- Regression after API or schema changes.
- Check role gates (uploader/moderator/admin).
- Verify i18n fallbacks and SEO noindex rules when those areas change.
- Against v1 exit checklist before calling v1 done.

**Outputs:** Test plan for the slice; failed cases with repro steps; sign-off vs checklist.

**References:** `docs/v1-scope.md` exit checklist, Product acceptance criteria

---

## Collaboration protocol

1. **Main** confirms scope label (IN / SOON / LATER).  
2. **Product** writes or trims stories + acceptance.  
3. **Backend** and **Frontend** agree on contract (types/endpoints).  
4. **i18n** / **SEO** consult when the slice touches chrome strings or public routes.  
5. **QA** reviews against acceptance + exit checklist.  
6. **Main** merges and updates docs if behavior changed.

## Anti-patterns

- Frontend inventing API shapes without Backend.
- Implementing LATER desktop/AI/DMs during v1 spine work.
- Hard-coding copy in components when i18n keys are required for the slice.
- Indexing private or pending content.
- Shipping without QA on pay/upload/review paths.
