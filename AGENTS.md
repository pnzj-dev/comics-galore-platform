# AGENTS.md – Comics Galore

This repository is an **architecture-first** specification for Comics Galore.  
It exists so that any AI coding agent (or human developer) can implement the product with minimal additional guidance.

**Do not write application code until the architecture documents in `/docs` are understood and followed.**

**Scope gate:** For prioritization, `docs/v1-scope.md` overrides the full vision. Implement **IN (v1)** first; do not build **LATER** features before v1 exit criteria are met.

---

## 1. Project Mission

Comics Galore is a production-grade comics sharing and blogging platform available as:

- A **web application** (SvelteKit)
- A **desktop application** (Wails + Svelte)

Both clients share as much Svelte UI, form logic, validation and business code as possible.  
They talk to the same Encore (Go) backend.

Core product capabilities remain: uploaders, tiers × intervals subscription matrix (NowPayments), human + AI moderation, messaging, support, feature-rich admin, etc.

---

## 2. Strict Technology Stack

| Concern              | Choice                                      | Notes |
|----------------------|---------------------------------------------|-------|
| Backend API          | Encore (Go)                                 | Official Encore CLI only |
| Web frontend         | SvelteKit + shadcn-svelte + Tailwind        | Official scaffolding |
| Desktop frontend     | **Wails** (Go) + Svelte + shadcn-svelte     | Reuses shared Svelte packages |
| Shared UI / logic    | Internal package(s)                         | Components, stores, Zod schemas, API client wrappers |
| Forms                | Superforms + Zod                            | Shared |
| Tables               | TanStack Table (Svelte)                     | Admin |
| Charts               | Apache ECharts or Chart.js                  | Admin |
| Object Storage       | S3-compatible (pages + archives)            | Presigned URLs |
| Images (cover/preview) | Cloudflare Images                         | Kind-flagged assets |
| Image delivery       | direct \| imgproxy \| cloudflare_images     | Admin setting; default direct |
| Auth                 | Encore Auth                                 | Same for web & desktop |
| Payments             | NowPayments (v1)                            | Crypto only; **PaymentsProvider** interface, multi-provider-ready |
| AI Moderation        | Configurable LLM                            | Optional |
| Emails               | Resend                                      | |
| Background Jobs      | Encore background tasks                     | |

**Never introduce React.** Desktop must reuse the Svelte component layer.

---

## 3. Monorepo Layout (mandatory)

```
Comics-Galore/
├── AGENTS.md
├── README.md
├── docs/
├── backend/                     # Encore Go API
├── frontend-public/             # SvelteKit web app (comics-galore.com)
├── frontend-admin/              # SvelteKit web app (admin.comics-galore.com)
├── desktop/                     # Wails application
│   ├── main.go / app.go         # Wails Go bindings
│   └── frontend/                # Svelte (Vite) app that consumes shared UI
└── packages/
    └── ui/                      # Shared Svelte components, stores, schemas, utils
        ├── components/          # shadcn-svelte based + domain components
        ├── forms/
        ├── lib/                 # Zod schemas, API helpers, stores
        └── ...
```

Rules:
- All reusable Svelte UI and client logic lives in `packages/ui`.
- `frontend-public/` (SvelteKit) and `frontend-admin/` (SvelteKit) both import from `packages/ui`.
- Public app handles all consumer-facing features (home, browse, reader, auth, upload, settings).
- Admin app handles admin-only features (dashboard, moderation, users, subscriptions, comics).
- Do not add admin features to public app or public features to admin app.
- The Encore backend remains the single source of truth for business logic, auth, payments and data.

---

## 4. Non-Negotiable Rules

### Generators First
- `encore app create`
- Official SvelteKit scaffolding
- `npx shadcn-svelte@latest`
- Official Wails CLI (`wails init`, `wails generate`, etc.)

Never hand-write core config files that a generator produces.


### Desktop Offline & Native (Wails only)
- Offline library: user folder, per-comic `.cbz`, bulk series/Continue Reading, auto-next, multiple profiles.
- Reader opens local CBZ directly; export/backup supported.
- System tray, native notifications, global hotkeys, start-with-OS.
- “Open with” for CBZ/CBR, Jump List/dock menu, drag-and-drop import.
- Fullscreen reader, dual-page, touch/pen/gamepad, Quick Look, local reading stats.
- File I/O, tray, notifications, hotkeys live in Wails Go; UI in Svelte.
- Web has none of these offline/native features.

### Code Reuse
- Prefer moving any component or form used by both web and desktop into `packages/ui`.
- Avoid duplicating Zod schemas, comic cards, reader chrome, upload session logic, etc.

### Data & Uploads
- Same presigned-URL + Upload Session model for both clients.
- Same unified `POST /comics` payload.
- Desktop may additionally use native file dialogs, but still uploads via presigned URLs.

### Bootstrap & Matrix
- System refuses to start without at least one admin.
- Incomplete Tier × Interval plan matrix → red banner in admin (web and desktop admin views).

### Quality
SOLID, DRY, KISS, accessible, dark-mode ready.

---

## 5. Implementation Order (high level)

1. Read all `/docs`.
2. Scaffold Encore backend + SvelteKit web + shared `packages/ui`.
3. Scaffold Wails desktop shell that consumes `packages/ui`.
4. Auth + roles + bootstrap check.
5. Core comics + public experience (web first, then desktop parity).
6. Uploader 3-tab workspace (shared UI).
7. Review pipeline, subscriptions, moderation, messaging, support, admin.
8. Desktop-specific enhancements (native menus, better file picking, offline reading later…).
9. Hardening & tests.

---

## 6. Workflow for Every Change

1. Explain what is being created.
2. Prefer official CLIs.
3. Put shared Svelte code in `packages/ui`.
4. Only add platform-specific code in `frontend/` or `desktop/`.

---

## 7. Definition of Done

- Web and desktop feel like one product.
- Maximum practical reuse of Svelte UI and logic.
- Official generators used for Encore, SvelteKit, shadcn-svelte and Wails.
- All previous product rules (pending_review, plan matrix, AI moderation, etc.) still hold.



## Subagent roles

Work is coordinated by a **Main Agent** with specialized subagents:

- **Product** — user stories and acceptance criteria
- **Backend** — Encore API, DB, jobs
- **Frontend** — SvelteKit / shared UI
- **i18n** — translation architecture and locales
- **SEO** — metadata, indexing, public discovery
- **QA** — tests and regression review

Full responsibilities and protocol: [`docs/agents.md`](./docs/agents.md).

## Project AI skills

Load from `.agents/skills/` when relevant:

| Skill | Use for |
|-------|---------|
| `comics-galore-scope` | Scope gate (always for build priority) |
| `encore-backend` | Encore Go backend |
| `sveltekit-ui` | SvelteKit + shadcn-svelte UI |
| `comics-upload` | Presigned upload + `POST /comics` |
| `nowpayments-billing` | Crypto subscriptions & webhooks |
| `wails-desktop` | Desktop/offline (after v1) |
| `tier-gated-gallery` | Tier-limited gallery/lightbox + upgrade CTA |

See `.agents/skills/README.md`.
