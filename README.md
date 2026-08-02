# Comics Galore

Architecture repository for a comics platform delivered as:

- **Web** – SvelteKit + shadcn-svelte
- **Desktop** – Wails + Svelte (reuses the same UI package)
- **Backend** – Encore (Go)

## Monorepo layout

```
Comics-Galore/
├── backend/          # Encore API
├── frontend/         # SvelteKit web
├── desktop/          # Wails desktop
└── packages/ui/      # Shared Svelte components, forms, schemas, upload logic
```

## Key ideas

- One backend, two clients.
- Feature-complete public app (age ratings, legal pages, series follow, reader scrubber/fit modes, notification prefs).
- Shared Svelte layer = maximum reuse (comic cards, reader, 3-tab New Comic workspace, admin pieces, Zod schemas…).
- Presigned S3 uploads + unified creation payload on both clients.
- New comics always start as `pending_review`.
- Strict Tier × Interval subscription matrix.
- Optional AI moderation.
- System refuses to start without an admin user.
- **Desktop**: offline CBZ library, system tray, native notifications, global hotkeys, drag-and-drop, fullscreen reader, “Open with” CBZ/CBR, and more.

## Docs

**Start here for build priority:** [docs/v1-scope.md](./docs/v1-scope.md)

**AI skills:** [`.agents/skills/`](./.agents/skills/) — scope, Encore, SvelteKit, upload, billing, Wails


See `docs/` and `AGENTS.md` for the full specification and rules.

**Next step:** scaffold Encore + SvelteKit + `packages/ui` + Wails using only official generators.

- [docs/agents.md](./docs/agents.md) — subagent roles
