# V1 Scope Freeze – Comics Galore

This document is the **authoritative split** between what we build first and what stays in the full vision docs as later work.

Full product vision remains in `product.md`, `architecture.md`, and the ADRs.  
**Builders and AI agents must treat this file as the scope gate for v1.**

Labels:

- **IN (v1)** — must ship for the first production-capable release  
- **SOON (v1.1)** — start right after v1; hooks may exist in v1  
- **LATER** — specified in architecture, not required for v1 exit  

---

## V1 goal (one sentence)

A working web app where users can discover and read comics, uploaders can submit comics for review, admins can approve content and manage basic billing, and at least one paid crypto tier works end-to-end via NowPayments.

---

## IN — V1 (must ship)

### Platform
- Encore backend + **SvelteKit web** client  
- Auth + roles: `user`, `uploader`, `moderator`, `admin`  
- Cookie-based JWT auth with server-side SvelteKit route guards (no page flash)  
- PostgreSQL + Encore migrations + SQLC (or chosen ORM)  
- Dark mode + basic accessibility on core flows  

### Comics & media
- Comic entity: metadata, status (`pending_review` → `published` / `rejected`)  
- **Manual** create flow for uploaders  
- Presigned **S3** uploads for archives/pages (no file body through API)  
- Unified `POST /comics` with file keys in payload  
- Covers: may live on S3 in v1 (Cloudflare Images path can wait)  
- Image delivery: **`direct` only** (resolver stub OK; no imgproxy/CF mode UI required)  
- Asset **kind** field on model (`cover` | `preview` | `page` | `original`) even if only S3 is used  
- Public list + detail (published only)
- Comic **content_language** field (default `en`); filter support on list API
- UI i18n infrastructure with **English** catalogs; locale switcher optional if only `en` enabled  
- Basic web reader (pages, keyboard, progress save, Continue Reading)  
- View + download counters; download quota check server-side (simple tier limits)  

### Social (minimal)
- Like / unlike  
- Dislike / undislike  
- Like/Dislike mutual exclusivity: liking removes an existing dislike, disliking removes an existing like
- Favorite / unfavorite  
- Per-user reaction status in list + detail responses  
- Reaction counts displayed on ComicCards and detail stats  

### Reader & media viewing
- Lightbox carousel: fullscreen overlay with keyboard nav, dot indicators, page counter  
- Cover section inline carousel: left/right arrows (disabled at bounds), thumbnail strip, page counter  
- Thumbnail strip below cover: click to select image, highlighted current selection  

### Tiers & payments
- Tiers table + seed (e.g. free + 1–2 paid)  
- Intervals + plan matrix **schema** (may only configure a few plans)  
- NowPayments **only** behind a `PaymentsProvider` interface  
- One working paid checkout path (invoice/deposit/subscribe as already designed, simplified UI OK)  
- Webhook ingest + raw payload storage + internal subscription activation  
- Red banner in admin if configured plans are incomplete (can be simple)  

### Moderation & admin (minimal viable control panel)
- Pending comics queue: approve / reject with reason  
- User list + role change + basic ban/suspend  
- Comic list + status management  
- Basic KPIs (users, comics, downloads, active subs) — charts nice-to-have  
- Settings: maintenance flag, registration open/closed, contact email, hide mature default, enable comments, default meta description, quotas, rate limit, image serving mode — all editable via Form/JSON toggle  
- Audit log for critical actions (approve, reject, ban, role change, manual grant if any)  

### Account & trust (minimal)
- Email verification **or** documented Auth capability equivalent  
- Password reset / recovery as supported by Encore Auth  
- Static pages: Terms, Privacy (DMCA page can be a short stub)  
- Simple age_rating field on comics (enum); simple gate or “hide mature” setting  

### Quality bar for v1 exit
- Happy paths work on web: register/login, browse, read, upload, approve, pay  
- No desktop requirement  
- No AI moderation requirement  
- Docs updated if implementation diverges  

---

## V1.1 scope

V1.1 builds on the shipped v1 web spine. The following **IN (v1.1)** items are the next build target; the rest of the original SOON list is already done in v1 (marked below).

### IN — V1.1 (must ship)

- **i18n foundation** — message catalog, default `en`, `ui_locale` wiring, locale detection (user → Accept-Language → `en`), locale switcher (optional if only `en` enabled). First item; foundational for all surfaces.
- **Comment flagging / report flow** — flags table + endpoints, moderator queue integration.
- **Real notification emails** (Resend) — a few key events (e.g. new from followed uploader, support replies) beyond verify/reset.
- **Archive `comic.json` metadata extraction** — libarchive.js client-side parse of `comic.json`/`metadata.json`, auto-building the same `POST /comics` payload.
- **Cloudflare Images wiring in upload forms** — cover/preview go through CF upload URLs (backend already exists); `image_serving_mode` resolver applied to cover/preview.
- **Soft quota warning (~80%)**.
- **Series progress % / missing-issue gaps** on series pages.
- **Language facet UI** on public browse (`language=` filter is backend-ready).

### SOON — V1.1.1 (next after v1.1)

- Additional UI locales (ja, es, ko, fr, pt-BR, …) as translation packs land.
- Language facet polish (facets, counts).

---

## Done in v1 (original SOON items already shipped)

- ~~Migrate all data-fetching pages from `onMount` + client-side `api.get` to SvelteKit `load` functions~~ — Encore client migration complete; all pages use `+page.server.ts` with `getEncoreClient(token)`.
- ~~Comments + SSE live comments~~ — `/comments`, `/comments-stream/:id`, CommentList/CommentForm.
- ~~Series entity + series pages + series follow~~ — `/series`, `/series/:id`, `/series/:id/follow`.
- ~~Reader thumbnails/scrubber + fit modes~~ — Reader.svelte has fit modes + thumbnail toggle.
- ~~Tag pages + home rails~~ — `/tags/[tag]`, "Popular This Month" / latest / random rails.
- ~~Admin recycle bin + moderation bulk + user detail drawer~~.
- ~~Notification preferences~~ — `/me/notification-preferences`.
- ~~Cloudflare Images backend~~ — upload URL + media proxy (frontend wiring deferred to v1.1).

---

## LATER — full vision (documented, not v1)

### Desktop (Wails)
- Shared `packages/ui` monorepo packaging  
- Offline CBZ library, user folder, per-comic offline toggle  
- System tray, native notifications, global hotkeys  
- Drag-and-drop, Open with CBZ/CBR, Jump List  
- Fullscreen dual-page, gamepad, Quick Look, local stats  
- Multiple offline library profiles  

### Billing & growth
- Full Tier × Interval matrix UX polish  
- Coupons, manual grant/extend UI polish, past-due tooling  
- Second payment provider adapter  

### Social & engagement
- Internal DMs  
- Full support ticket system  
- Broadcast announcements  
- AI moderation (configurable LLM) + decision log  
- Shareable public shelves polish, “People also liked”  

### Admin power tools
- Impersonation, CSV export, saved datalist views  
- Background job / dead-letter dashboard  
- Staff Picks ordering UI polish  
- Storage usage dashboard  

### Other
- imgproxy mode, advanced CDN strategy  
- OPDS / advanced offline networking  
- Creator payouts, remaining locale packs, native mobile  

---

## Explicit non-goals for v1

- Building web and desktop in parallel as equal priority  
- Feature-complete admin (as in the long product list)  
- Perfect subscription matrix with all intervals live  
- Client-side archive pipeline  
- AI anything  
- Messaging / full support desk  

---

## Phase mapping (aligned to roadmap)

| Phase focus | Scope label |
|-------------|-------------|
| Foundation (auth, scaffold, admin bootstrap) | IN (v1) |
| Core comics + public reader + manual upload + review | IN (v1) |
| Basic tiers + NowPayments path | IN (v1) |
| i18n foundation, flagging, notification emails, archive JSON, CF Images, quota/series polish | IN (v1.1) |
| Archive JSON upload, CF Images, comments, series | DONE (v1) |
| Full admin suite, AI, messaging, support | LATER |
| Wails + offline + native polish | LATER |

---

## Rules for agents & developers

1. Do **not** implement LATER features before the v1.1 exit criteria work on web.  
2. Prefer **hooks** (kind enum, provider interface, status `pending_review`) over full UI for LATER items.  
3. If a task is ambiguous, choose the **IN** interpretation.  
4. When v1.1 ships, move items from SOON → IN in a new revision of this file—do not silently expand v1.1 mid-build.  
5. Full vision docs stay valid as the north star; **this file wins on prioritization conflicts**.

---

## V1 exit checklist (demoable)

- [x] User can register/login  
- [x] Uploader submits a comic (manual) → appears as pending  
- [x] Admin/moderator publishes it  
- [x] Anonymous/user can open public detail + reader  
- [x] Progress/Continue Reading works when logged in  
- [x] Like/favorite/dislike work  
- [x] Free tier quota enforced on download  
- [~] At least one paid plan can be purchased via NowPayments and unlocks tier  
- [x] Webhook activates subscription; raw webhook stored  
- [x] Basic admin lists work (users, comics, payments/subs)  

When all boxes are checked, v1 is done—then execute SOON.
