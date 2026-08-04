# UI / UX Specification – Comics Galore

## Design Principles

- Public site: fast, content-first, minimal JS (SvelteKit SSR).
- Uploader / admin areas: powerful but clean.
- Full dark mode + WCAG compliance.
- Excellent progress feedback for long uploads.

## Technology

- **Web**: SvelteKit + shadcn-svelte + Tailwind CSS
- **Desktop**: Wails + Svelte + shadcn-svelte (same design system)
- **Shared**: `packages/ui` (components, forms, reader, upload logic, Zod schemas)
- Superforms + Zod, TanStack Table (Svelte), ECharts/Chart.js, libarchive.js
- Superforms + Zod
- TanStack Table (Svelte)
- Apache ECharts or Chart.js
- libarchive.js (client-side archive parsing)

## Uploader – New Comic Workspace (3 tabs)

### Tab 1 – My Comics
- Paginated grid of simplified comic cards (cover, title, status badge, date).
- Newest first.
- Clicking a card can open edit / view status.
- Empty state encourages going to Tab 2 or 3.
- After successful creation the user lands here.

### Tab 2 – Manual Creation
- Superforms form with clear sections.
- Cover + archive file inputs that:
  - Request presigned URL(s) from the backend
  - Upload **directly from browser to S3**
  - Show progress, allow cancel, support resume via Upload Session
- “Save draft / Resume” indicator if an incomplete session exists.
- Submit → creates comic in `pending_review` → redirect to Tab 1.

### Tab 3 – Archive + JSON
- Large drop zone.
- On file select:
  1. libarchive.js parses the archive in the browser.
  2. Looks for `comic.json` / `metadata.json`.
  3. Validates against the public JSON Schema (client-side first, server-side authoritative).
  4. Shows a preview of extracted metadata + file list.
  5. Uploads each required object via presigned URLs (multipart for large files).
- Progress UI for both parsing and uploading.
- Crash recovery via the same Upload Session mechanism.
- Builds the same form payload as the manual path (metadata + file keys).
- Calls the same creation API.
- On success → redirect to Tab 1.

### Shared Upload UX Rules
- Never upload the file body through the Encore backend.
- Clear error messages for invalid JSON, missing files, size limits, network failures.
- Ability to resume an interrupted session after page reload.
- Visual status of the session (uploading, paused, finalising, failed).

## Other Screens (summary)

**Public**: Home, Comic Detail, Series, Reader (keyboard + swipe), Search, Profiles, Pricing, RSS.  
**Authenticated**: Library, Messaging, Support, Account.  
**Moderator**: Moderation queue + pending comic reviews.  
**Admin**: Full control panel with red banner, datalists, settings, plan matrix, charts, etc.

### Pricing & Checkout Flow

Route: `/pricing` — all screens render in a modal, closable via X button or Esc only.

**Screen A — Plan Grid**
- Responsive grid of tier × interval cards.
- Each card shows: tier name, interval badge, price in USD, cumulative features list (higher tiers inherit all lower-tier features).
- "Select" button on each card → advances the modal to Screen B.

**Screen B — Crypto Currency Selector**
- Grid of crypto options with icons: BTC, ETH, USDT, LTC.
- **Single-select**: selecting one deactivates the others. User must unselect to reactivate, which also clears the estimated price.
- Backend calls NowPayments live price estimate API for the selected plan + crypto.
- Displays estimated amount in the chosen crypto.
- "Continue" button → triggers balance check → routes to Screen C or D.

**Screen C — Processing (Step 4A: funded)**
- Spinner + "Processing subscription..." message.
- Backend created the subscription atomically (saved locally only if NowPayments succeeded).
- **Polling**: frontend polls the local subscription by ID every 3 seconds, waiting for the subscription webhook to set `active = true`.
- On success → re-sign JWT cookie with new tier → page refresh (user sees upgraded tier).
- On timeout (5 minutes) → error message + retry button.

**Screen D — Deposit QR (Step 4B: needs funds)**
- QR code image (from NowPayments response) + deposit address with copy button.
- Amount to send displayed in the selected crypto.
- **Polling**: frontend polls the local deposit by ID every 5 seconds, waiting for the deposit webhook to set status to `completed`.
- On success → transition to Screen C (Processing subscription).
- On timeout (30 minutes) → "Transaction not detected yet" + retry button.

### Crypto Icon Library
- Use `lucide-svelte` or inline SVG for crypto currency icons (BTC, ETH, USDT, LTC).
- Icons should be visually distinct, 32×32 minimum in the selection grid.

## Comic Detail Page

Route: `/comics/:slug`

### Layout (desktop, two-column)

```
┌──────────────────────────────────────────────────────────────────────┐
│  [Premium] badge    Title (H1)                                       │
│  [tag1] [tag2] [+N]                                                  │
│  Description text (muted)                                            │
│                                                                      │
│  ┌──────────────────┐  ┌───────────────────────────────────────────┐ │
│  │                  │  │ ● Author (circular avatar + name)          │ │
│  │  Cover image     │  │                                           │ │
│  │  (aspect-3/4)    │  │ Views        1.2k     Downloads    456    │ │
│  │                  │  │ Pages        32       Published   Jan 2   │ │
│  │                  │  │ Reading Time 5 min    Language    EN      │ │
│  │                  │  │                                           │ │
│  │ [■][■][■][■]    │  │ [♥ 45]  [★ 12]                            │ │
│  │ thumbnail strip  │  │                                           │ │
│  │                  │  │ [Start Reading]    (premium gate)          │ │
│  └──────────────────┘  │ [Download ▼]       (auth/quota gate)       │ │
│                        └───────────────────────────────────────────┘ │
├──────────────────────────────────────────────────────────────────────┤
│  Synopsis (v1.1: premium-gated, 400-char truncation + upsell banner) │
├──────────────────────────────────────────────────────────────────────┤
│  Related Comics (4-col grid, lazy-loaded, skeleton loading)          │
├──────────────────────────────────────────────────────────────────────┤
│  Comments (v1.1: thread, SSE real-time, reply forms)                 │
└──────────────────────────────────────────────────────────────────────┘
```

**Mobile:** Stack vertically — cover full-width, metadata below, synopsis → related → comments.

### Feature breakdown by version

#### v1 (IN — implement now)

**Cover Gallery**
- Main cover image (3:4 aspect, `object-cover`), click to open fullscreen overlay
- Thumbnail strip below (4-5 images), dot indicators for pagination
- Fullscreen overlay: keyboard nav (ArrowLeft / ArrowRight / Escape), image counter header
- No premium gating on images in v1 (all visible)

**Title, Tags & Badges**
- H1 title with inline Premium badge (gold pill, `bg-yellow-500 text-white`)
- Tags row: each tag as a purple pill (`text-[10px] rounded-full bg-primary/10 text-primary`)
- Age rating badge if `mature` or `explicit` (red pill)

**Description**
- Plain paragraph, `text-muted-foreground`

**Author Info**
- Circular avatar (`size-8 rounded-full bg-purple-100 dark:bg-purple-900`) with uppercase first letter
- Author name + "Author" label

**Metadata Panel** (6-row key-value)
| Label | Value | Format |
|-------|-------|--------|
| Views | `view_count` | compactNum (k/M) |
| Downloads | `download_count` | compactNum |
| Pages | `page_keys.length` | raw int |
| Published | `published_at` | "Jan 2, 2006" |
| Reading Time | 5 | "5 min" |
| Language | `content_language` | uppercased ISO |

**Reactions Bar** (inline row)
- Like button: thumbs-up icon + count, filled/outlined toggle, `text-blue-500` when active
- Favorite button: star icon + count, filled/outlined toggle, `text-yellow-500` when active
- Both require auth; unauthenticated users see counts but buttons redirect to login

**Reader Access**
- "Start Reading" button (green/emerald) — opens existing `Reader` component
- If single archive: direct link. If multiple: dropdown button listing all files with name/size
- Premium-gated: non-premium users see upsell banner instead

**Download Section**
- Gating ladder: not authenticated → "Sign in to Download" → premium upsell → quota exhausted → download button
- Multi-archive: dropdown with filename + file size per entry
- Quota exhausted state: "X of Y GB used (tier)" + "Upgrade Plan" button

**Related Comics** (4-col grid)
- Lazy-loaded on mount via API call
- Skeleton loading: 4 cards with pulsing gray 3:4 aspect placeholders
- Empty state: "No related comics found"
- Uses existing `ComicCard` component

**View Count** — incremented server-side on page load

#### v1.1 (SOON)

**Synopsis Premium Gate**
- Full synopsis text. For non-premium users on premium comics: truncated to 400 chars + "..." + blur gradient + "Subscribe to read more" banner with CTA to Plans modal

**Comments**
- Lazy-loaded list of threaded comments
- Server-Sent Events (SSE) for real-time updates when new comments are posted
- Comment form (auth-gated): textarea + "Post Comment" button with inline spinner
- Per-comment reply toggle: nested reply form, indented replies with left border
- Max nesting depth configurable (default 3)

**Plans Modal Trigger**
- Same `CheckoutModal` from `/pricing` flow, triggered from "Subscribe" upsell on detail page

**Boost Quota Modal**
- Shows current quota: "X of Y GB used (tier)"
- Three boost options: +5 GB ($5.00), +10 GB ($8.00), +20 GB ($12.00)
- Each button POSTs to `/billing/boost` with SKU

#### v2 (LATER)

**Emoji Picker** — 8-column grid popover, cursor insertion into comment textarea
**User Hover Popover** — avatar + display name + email + role + tier badges
**Language Selector** — header toggles for EN/FR/ES
**Image Gallery Premium Gating** — blurred thumbnails + "Subscribe" CTA overlay for non-premium users, locked images unclickable

### Components used (shadcn-svelte)
- `Card` — related comics container
- `Button` — all CTA buttons
- `Badge` — tags, premium, age rating
- `Dialog` or custom overlay — fullscreen image viewer
- `Tabs` (optional) — desktop layout sections

## Component Guidelines

- Prefer shadcn-svelte.
- Comic cards reusable; see detailed spec in "Comic Card Components" below.
- Skeletons + optimistic UI where appropriate.
- Reader: ← → Space Esc + touch swipe.
- All uploads show determinate progress when possible.

## Comic Card Components

### ComicCard (primary public card)

The main card used in public grids, search results, and series pages. Rich, content-first, hover-interactive.

**Layout (top to bottom):**

```
┌─────────────────────────────┐
│  Cover image (aspect-3/4)   │ ← hover: scale-105 transition, lazy load, fallback SVG on error
│  ░░ gradient overlay (bot)  │ ← black/30 gradient 40px from bottom
│  ░░ hover description       │ ← black/70 overlay, centered description text, z-20
│  [Premium] badge (top-right)│ ← gold pill, z-10, icon-only on compact view
├─────────────────────────────┤
│  AUTHOR NAME                 │ ← text-xs, bold, uppercase, purple-500
│  Title (2-line clamp)        │ ← text-sm, font-semibold, hover:purple-500
│  [tag1] [tag2] [+N]         │ ← text-[10px], rounded, flex-wrap, overflow tooltip
│  Date · Favorite ★ toggle   │ ← text-xs, reactive Datastar signal toggle
├─────────────────────────────┤
│  Stats bar (border-top)     │ ← text-xs, muted-foreground, Lucide icons
│  👁 1.2k  ⬇ 456  💬 23  📖 32│ ← compactNum() formatting (k/M)
│                      👤 →   │ ← user popover trigger (hover to show)
└─────────────────────────────┘
```

**Cover image rules:**
- Real `<img>` with `loading="lazy"`, `object-cover`, 3:4 aspect ratio container
- `onerror` fallback to SVG placeholder (book icon)
- Hover zoom via `group-hover:scale-105 transition-transform duration-300`
- S3 keys → resolved via presigned download URL from media helper

**Hover overlay:**
- Black/70 background, full-card opacity transition
- Shows `comic.description` centered, `line-clamp-4`, text-xs
- `opacity-0 group-hover:opacity-100` with transition

**Premium badge:**
- Only renders when `comic.is_premium === true`
- Gold background (`bg-yellow-400` / `bg-yellow-500`), white text
- Crown icon (Lucide) or "Premium" text
- Position: `absolute top-2 right-2`, z-10

**Author:**
- Only renders when `comic.author` is non-empty
- `text-xs font-bold uppercase tracking-wide text-purple-500`

**Title:**
- `<a>` wrapping the comic detail link (`/comics/{slug}`)
- `text-sm font-semibold`, `line-clamp-2`, `min-h-[2.5rem]`
- `hover:text-purple-500 transition-colors`

**Tags:**
- Show first 2 tags as rounded pills (`text-[10px]`)
- If >2 tags, show "+N" pill with full tag list as `title` attribute tooltip
- Wrap with `flex items-center gap-1 flex-wrap`

**Date:**
- Format: `Jan 2, 2006` (when `published_at` is set)
- `text-xs text-gray-500`

**Favorite toggle:**
- Star icon (Lucide), inside card row (not separate component)
- Reactive: Datastar `data-on:click @post` with signal `$_isFav_{id}`
- Active: `text-yellow-500`, Inactive: `text-gray-400 hover:text-yellow-500`
- Only rendered when `isAuthenticated === true`

**Stats bar:**
- Border-top separator (`border-t border-gray-100`)
- Row of icons + formatted numbers:
  - 👁 Views (`Eye` icon) → `compactNum(viewCount)` → "1.2k", "3.4M"
  - ⬇ Downloads (`Download` icon)
  - 💬 Comments (`MessageCircle` icon)
  - 📖 Pages (`BookOpen` icon) → raw number
- `text-xs`, muted-foreground

**User popover (right-aligned in stats bar):**
- Trigger: `User` icon button, hover to show popover
- Content: avatar (first-letter circle), display name, email, role badge, tier badge
- Only rendered when `comic.user.display_name` is non-empty
- See `Popover` component for implementation

**Card chrome:**
- Border: `border border-gray-200 dark:border-gray-700`
- Background: `bg-white dark:bg-gray-800`
- Hover: `hover:border-purple-300 dark:hover:border-purple-700 hover:shadow-lg`
- Radius: `rounded-xl`
- Transition: `transition-all`

**Stats number formatting (`compactNum` / `cardNumStr`):**
```
0       → "0"
1 - 999 → raw number
1,000+  → "1.2k"
1,000,000+ → "1.5M"
```

### CompactCard (admin / uploader workspace)

Minimal variant for admin lists, uploader "My Comics" tab, and moderation queues.

**Layout:**

```
┌─────────────────────┐
│ Cover (aspect-2/3)  │
│ [Premium ★]         │
├─────────────────────┤
│ Title (1-line clamp)│ ← hover:text-purple-500
│ [status] author     │ ← status badge + truncated author
│              👁 1.2k │ ← view count only
└─────────────────────┘
```

**Differences from ComicCard:**
- Aspect ratio: 2:3 (taller, narrower)
- No hover description overlay
- No stats bar — only view count in footer
- Status badge replaces full stats: color-coded pills
  - `published` → green (`bg-green-100`, `text-green-600 dark:text-green-400`)
  - `pending` / `pending_review` → yellow (`bg-yellow-100`, `text-yellow-600 dark:text-yellow-400`)
  - `draft` → gray (`bg-gray-100`, `text-gray-500 dark:text-gray-400`)
- No favorite toggle
- No tags
- No user popover
- Premium badge: icon-only (smaller, crown SVG)

### ComicGrid

Responsive grid container for ComicCard components.

**Breakpoints:**
```
xs:   grid-cols-1  (mobile)
sm:   grid-cols-2  (640px+)
lg:   grid-cols-3  (1024px+)
xl:   grid-cols-4  (1280px+)
```

**Spacing:** `gap-6` between cards.

**Props:**
- `posts: Comic[]` — comic list to render
- `total: number` — total count for pagination
- `page: number` — current page
- `pageSize: number` — items per page
- `favoriteIDs: Map<UUID, boolean>` — which comics the user has favorited
- `isAuthenticated: boolean` — controls favorite toggle visibility
- `gridID: string` — DOM id for pagination swap target

**Empty state:**
- When `posts.length === 0`: show empty state with title "No posts found" and guidance text
- Use `EmptyState` component

**Pagination:**
- Bottom of grid, `mt-auto pt-4`
- Server-driven via Datastar SSE: `@get('/api/posts/grid?page=...&selector={gridID}')`
- See `Pagination` component

### RelatedGrid

Grid for related/similar comics on detail pages. Identical card rendering to ComicCard, different layout.

**Differences from ComicGrid:**
- Fixed 4 columns on `lg:` (`grid-cols-4`)
- Manual page buttons (not SSE-driven like ComicGrid)
- `gap-4` spacing (tighter)

**Props:**
- `posts: Comic[]`, `total: number`, `page: number`
- `favoriteIDs: Map<UUID, boolean>`, `isAuthenticated: boolean`

### Dark Mode Reference

Every ComicCard element must have explicit `dark:` variants:

| Element | Light | Dark |
|---------|-------|------|
| Card bg | `bg-white` | `dark:bg-gray-800` |
| Card border | `border-gray-200` | `dark:border-gray-700` |
| Card hover border | `hover:border-purple-300` | `dark:hover:border-purple-700` |
| Separator | `border-gray-100` | `dark:border-gray-700/50` |
| Tags bg | `bg-gray-100` | `dark:bg-gray-700` |
| Tags text | `text-gray-500` | `dark:text-gray-400` |
| Premium gold | `bg-yellow-400` | no change |
| Cover placeholder | `text-gray-400` | `dark:text-gray-500` |
| Cover bg | `bg-gray-100` | `dark:bg-gray-700` |
| Popover bg | `bg-white` | `dark:bg-gray-800` |
| Popover text | `text-gray-900` | `dark:text-gray-100` |

### Skeleton States

When loading, show skeleton cards matching the ComicCard layout:
- Skeleton `aspect-[3/4]` cover block
- Skeleton text lines for author, title (2 lines), tags, stats
- Use `animate-pulse` with `bg-gray-200 dark:bg-gray-700`
- Match the grid columns of the target layout


## Desktop-only UI (Wails)

### Offline
- Per-comic **“Available offline”** toggle + progress.
- Bulk: “Download series”, “Download Continue Reading”.
- Settings: offline folder picker, multiple library profiles, auto-download next issue, storage warnings.
- Offline library manager (list, size, remove, export/backup).

### Native behaviour
- System tray menu (Continue Reading, active downloads, unread).
- Native notification toasts (download complete, new issues, support replies…).
- Global hotkey settings.
- Drag-and-drop zones on main window and New Comic tab.
- “Open with” / file association feedback.

### Reader polish
- True fullscreen (F11), auto-hide chrome, optional dual-page spread.
- Touch / pen swipe; optional gamepad hints.
- Quick Look overlay (cover + first pages).
- Local session stats visible in a discreet panel.

### Admin Control Panel
- Persistent **red banner** when plan matrix is incomplete
- Dashboard: KPI cards, charts, activity feeds, quick actions
- Datalists with search, filter, **saved views**, **bulk actions**, **CSV export**
- User detail drawer (subs, downloads, flags, tickets); ban/suspend; audited impersonation
- Comics/series/tags management; recycle bin; scheduled list; Staff Picks ordering
- Plan matrix editor; manual subscription grant/extend; coupons; past-due list
- AI settings + AI review queue + decision log
- System health: background jobs, dead-letter retry, storage usage
- Broadcast announcements; email template preview
- Support tickets; audit log; general settings

## Public experience (completed)

- Home rails: Staff Picks, New this week, Popular this month, Comic of the Day, Continue Reading
- Tag pages as public routes
- Series page: progress %, missing issues, series-follow button
- Reader: thumbnails/scrubber, fit modes (width/height/original), optional dual-page
- Age rating badge on cards/detail; age-gate modal when required
- Static routes: Terms, Privacy, DMCA
- Cookie consent banner (if analytics enabled)
- Auth: verify-email, reset-password flows; notification preferences screen
- Empty states + first-time guidance; quota-blocked upgrade state

- Admin media settings: image serving mode (default direct), imgproxy base URL.


## Language & i18n UI

- Locale switcher in header/user menu (enabled locales only).
- Comic forms: required **content language** select (defaults from settings).
- Browse/search: language facet / filter chips.
- Comic cards/detail show content language badge when useful (e.g. mixed-language grids).
- All chrome strings go through i18n catalogs (SvelteKit-friendly i18n library).
- Default locale `en`; fallback chain user → browser → en.

## Tier-gated image gallery

- Main image, dots, thumbnails, fullscreen lightbox (see skill `tier-gated-gallery` and reference TS gallery).
- Indices beyond the tier preview limit are blurred with upgrade CTA (plans modal).
- Locked images must not open sharp fullscreen; prefer withheld or placeholder URLs from the API.
