# Comics Galore – project AI skills

**Location:** `.agents/skills/` (versioned with the repo).


These skills specialize coding agents on this repository’s conventions. They complement (do not replace) `AGENTS.md` and `docs/`.

| Skill | When to use |
|-------|-------------|
| `main-agent` | Orchestrate subagents (Product, Backend, Frontend, i18n, SEO, QA) |
| `comics-galore-scope` | Any implementation or prioritization — enforces v1 freeze |
| `encore-backend` | Encore Go APIs, migrations, auth, jobs |
| `sveltekit-ui` | SvelteKit pages, shadcn-svelte, forms |
| `comics-upload` | Presigned uploads, creation payload, pending_review |
| `nowpayments-billing` | Plans, subscriptions, webhooks, provider interface |
| `wails-desktop` | Desktop/offline only (LATER vs v1-scope) |
| `tier-gated-gallery` | Svelte lightbox/gallery with tier-blurred images + upgrade CTA |

## Related bundled environment skills (optional)

Not project-specific, but useful when relevant:

- **ffmpeg** — extract pages/covers from archives, transcode previews (if you ever process media server-side)
- **imagemagick** — image convert/resize for covers
- **pdf** — only if you add PDF comics later
- **skill-creator** — add or update skills in `.agents/skills/`

## Adding a skill

Follow the skill-creator format (`SKILL.md` with YAML `name` + `description`, no colon-space in description). Keep instructions project-specific and short.
