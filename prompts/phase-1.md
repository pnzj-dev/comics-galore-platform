You are building Comics Galore from the architecture repository in this workspace.

## Mandatory reading (in order)
1. AGENTS.md
2. docs/v1-scope.md          ← scope gate; IN only until v1 exit checklist passes
3. docs/agents.md            ← Main + Product/Backend/Frontend/i18n/SEO/QA protocol
4. docs/architecture.md
5. docs/product.md (skim)
6. Relevant skills under .agents/skills/ (especially comics-galore-scope, encore-backend, sveltekit-ui)

## Rules
- Official generators only (Encore CLI, SvelteKit scaffold, shadcn-svelte, Tailwind).
- Do NOT implement LATER features (Wails desktop suite, AI moderation, DMs, full support desk, full admin power tools) before the v1 exit checklist in docs/v1-scope.md is done.
- Prefer hooks (asset kind, PaymentsProvider interface, pending_review) over building full later UIs.
- Stack: Encore (Go) backend + SvelteKit + shadcn-svelte + Tailwind frontend. No React.

## Start with Phase 1 – Foundation only
1. Explain what you will create.
2. Scaffold backend with Encore CLI into backend/ (or repo layout as in AGENTS.md).
3. Scaffold frontend with official SvelteKit + Tailwind + shadcn-svelte into frontend/.
4. Wire basic project layout; empty packages/ui only if needed later.
5. Auth + roles (user, uploader, moderator, admin).
6. Bootstrap rule: refuse to start (or clearly fail) if no admin user exists in non-dev as specified.
7. PostgreSQL via Encore migrations; seed default tiers.
8. Dark mode baseline.

After Foundation works, stop and summarize. Do not jump to payments, desktop, or archive upload until I confirm.

Work as Main Agent: keep changes small, use generators, update docs only if behavior diverges from the spec.