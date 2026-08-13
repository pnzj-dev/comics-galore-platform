# Roadmap – Comics Galore

> **Prioritization:** see [v1-scope.md](./v1-scope.md) for the frozen **IN / SOON / LATER** split. This roadmap describes the full journey; v1-scope decides what ships first.

## Phase 0 – Architecture Repository (current)
- [x] Web + Desktop (Wails) architecture
- [x] Shared Svelte package strategy
- [x] Scaffold with official tools

## Phase 1 – Foundation
- Encore backend scaffolding
- SvelteKit web scaffolding + shadcn-svelte
- `packages/ui` skeleton (shared components, Zod, theme)
- Wails desktop shell that consumes `packages/ui`
- Auth + roles + bootstrap admin check
- Dark mode working on both clients

## Phase 2 – Core Comics & Public Web Experience
- Public listing, detail, series, reader (web)
- Shared reader + comic card components in `packages/ui`
- Like, favorite, rating, comments, Continue Reading
- Reader: thumbnails/scrubber, fit modes, optional dual-page
- Discovery rails (New this week, Popular this month, Comic of the Day)
- Tag pages; series follow; series progress % + missing issues
- Age rating + age gate; Terms / Privacy / DMCA pages
- Email verification, password reset, notification preferences
- Empty states, quota-blocked upgrade UX; RSS + Open Graph

## Phase 3 – Uploader Workspace (shared)
- 3-tab New Comic UI in `packages/ui`
- Presigned uploads + Upload Session recovery
- Manual + Archive (libarchive.js) both build same payload → `POST /comics`
- Desktop uses the same workspace (native file dialog optional)

## Phase 4 – Review Pipeline
- pending_review → AI / human approve-reject
- Works identically from web and desktop

## Phase 5 – Series, Library, Progress
- Shared library & series components

## Phase 6 – Subscriptions & NowPayments
- Plan matrix, multi-step modal (shared), wallets, webhooks

## Phase 7 – Comments, Moderation, Messaging, Support
- Shared UI; desktop parity

## Phase 8 – Feature-Complete Admin Control Panel
- KPI dashboard + charts + activity feeds + quick actions
- Datalists with search, filter, saved views, bulk actions, CSV export
- User detail drawer, ban/suspend, audited impersonation
- Recycle bin, scheduled comics, tags/series management, Staff Picks ordering
- Plan matrix UI + red banner; manual subscription grant/extend; coupons; past-due list
- AI settings, AI review queue, AI decision log
- Background jobs status + dead-letter retry; storage usage
- Broadcast announcements; email template preview
- Audit log; general settings


## Phase 9 – Desktop Offline + Native Polish
- Offline library: user folder, per-comic CBZ, bulk series / Continue Reading, auto-next, multiple profiles
- Desktop reader reads local CBZ; export/backup
- System tray, start-with-OS, native notifications
- Global hotkeys, “Open with” CBZ/CBR, Jump List / dock menu
- Drag-and-drop import (archives + image folders)
- Fullscreen distraction-free reader, dual-page, touch/pen/gamepad
- Quick Look preview, local reading stats
- Native menus & window state

## Phase 10 – Production Hardening
- CI for web + desktop builds
- Observability, backups, tests, performance

## Phase 11 – V1.1 (shipped-after-v1 refinements)
- [x] i18n foundation (catalog, `en`, `ui_locale`, locale detection) — first
- [x] Comment flagging / report flow
- [x] Archive `comic.json` metadata extraction (libarchive.js)
- [x] Cloudflare Images wiring in upload forms + `image_serving_mode` resolver
- [x] Soft quota warning (~80%)
- [x] Series progress % / missing-issue gaps
- [x] Language facet UI on public browse
- [x] Real notification emails (Resend) + uploader follow

See `docs/v1-scope.md` for the authoritative IN (v1.1) list and the original SOON items already shipped in v1.
