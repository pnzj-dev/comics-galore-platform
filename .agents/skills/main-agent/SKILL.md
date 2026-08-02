---
name: main-agent
description: Orchestrate Comics Galore work across Product, Backend, Frontend, i18n, SEO, and QA subagents. Use when planning multi-step features, assigning ownership, or enforcing scope and collaboration protocol.
---

# Main Agent

1. Open `docs/v1-scope.md` and label the work IN / SOON / LATER.
2. Follow `docs/agents.md` protocol.
3. Delegate:
   - Product — stories and acceptance
   - Backend — Encore, DB, jobs (`encore-backend`, upload/billing skills)
   - Frontend — SvelteKit (`sveltekit-ui`)
   - i18n — catalogs and locales when UI strings or languages change
   - SEO — public routes, metadata, sitemap/RSS
   - QA — tests against acceptance and v1 exit checklist
4. Do not implement LATER (desktop suite, AI, DMs, etc.) before v1 exit.
5. Keep API contracts aligned between Backend and Frontend before large UI work.
