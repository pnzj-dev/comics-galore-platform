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

**Layout: two-column (metadata left, cover right), full-width previews + archives**

```
┌──────────────────────────────────────────────────────────────────────┐
│  Create New Comic                                                    │
│                                                                      │
│  ┌─────────────────────────────┐ ┌────────────────────────────────┐ │
│  │ Title *                     │ │                                │ │
│  │ Author                      │ │   Cover Image *                │ │
│  │ Description (2 rows)       │ │                                │ │
│  │ Synopsis (2 rows)          │ │   ┌──────────────────────┐    │ │
│  │ Language ▼  Age Rating ▼    │ │   │                      │    │ │
│  │ Category                    │ │   │   3 : 4 aspect       │    │ │
│  │ Tags (comma-separated)      │ │   │   upload or preview  │    │ │
│  └─────────────────────────────┘ │   │       × %            │    │ │
│                                  │   └──────────────────────┘    │ │
│                                  └────────────────────────────────┘ │
│                                                                      │
│  Preview Images (min 2)                                   [+ Add]    │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐                  │
│  │ 🌄 │ │ 🌄 │ │ 🌄 │ │ 🌄 │ │ 🌄 │ │  + │ │  + │                  │
│  │ 2:3│ │ 2:3│ │ 2:3│ │ 2:3│ │ 2:3│ │ 2:3│ │ 2:3│                  │
│  └────┘ └────┘ └────┘ └────┘ └────┘ └────┘ └────┘                  │
│  7-column grid, each slot: upload→preview→progress→× remove         │
│                                                                      │
│  Archive Files (min 1)                                   [+ Add]    │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐                       │
│  │  ⬇   │ │  ⬇   │ │ 50%  │ │  +   │ │  +   │                       │
│  │ name  │ │ name  │ │ name  │ │ Arch  │ │ Arch  │                       │
│  │size ×│ │size ×│ │size   │ │       │ │       │                       │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘                       │
│  5-column grid, each slot: upload→name+size→× remove                │
│                                                                      │
│  [Publish Comic] (full width)                                       │
└──────────────────────────────────────────────────────────────────────┘
```

**Metadata fields (left column, stacked):**
- Title (required)
- Author
- Description (2 rows textarea)
- Synopsis (2 rows textarea)
- Language selector + Age Rating selector (inline)
- Category text input
- Tags (comma-separated, full width in metadata column)

**Cover Image (right column, 320px, required)**
- 3:4 aspect ratio placeholder when empty
- File input overlay (`accept="image/*"`)
- On select: local `URL.createObjectURL()` preview fills the area
- Upload starts immediately via presigned URL with XHR progress tracking
- Progress state: dark overlay with spinner + percentage
- × button (top-right) to remove and clear

**Preview Images (full width, min 2, max 10)**
- 7-column responsive grid (3 on mobile, 5 on tablet, 7 on desktop)
- Each slot: 2:3 aspect, dashed border, file input overlay
- Empty: book icon placeholder
- Uploaded: image preview with × remove button (cannot go below 2)
- Progress overlay: spinner + percentage on upload
- [+ Add] button (hidden at 10 max)

**Archive Files (full width, min 1, max 10)**
- 5-column responsive grid (3 on mobile)
- Each slot: square aspect, dashed border, file input overlay
- Accepts: `.cbr,.cbz,.pdf,.zip,.rar,.7z`
- Empty: download icon + "Archive" label
- Uploaded: purple tint + download icon + filename (truncated) + file size + × remove (cannot go below 1)
- Progress: spinner + percentage on upload
- [+ Add] button (hidden at 10 max)

**Submit**
- Validates: title, cover, at least 1 archive
- Collects keys from uploaded files
- Sends `POST /comics` with unified payload
- On success: redirects to Tab 1 (My Comics)

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

**Route: `/pricing`** — informational only, shows all plans inline with interval selector and feature diffs. No selection buttons in page mode. Authenticated users see a "Subscribe" button that opens the checkout modal.

**`SubscribeButton`** — reusable component that opens the checkout modal. Import on any page (detail, navbar, upload).

**`PlanGrid`** — shared plan display used in both modal (`mode="modal"`) and page (`mode="page"`) contexts.

#### PlanGrid layout (one card per tier)

```
┌─ Free ────────────────────────────────┐
│                                        │
│  ✓ Browse comics                       │
│  ✓ Read comments                       │
│  ✓ 1 GB download quota                 │
│                                        │
│  Free                     (no button)  │
└────────────────────────────────────────┘
┌─ Silver ──────────────────────────────┐
│  Interval: [Monthly ▼]                 │
│                                        │
│  ✓ Browse comics                       │
│  ✓ Read comments                       │
│  ✓ Write comments                      │
│  ✓ Download archives                   │
│  ✓ Web reader                          │
│  ✓ Full preview gallery                │
│  + 50 GB download quota                │
│                                        │
│  $6.99/month                 [Select]  │
└────────────────────────────────────────┘
```

**Rules:**
- One card per tier (not per plan/interval)
- Interval chosen via `<select>` dropdown inside each card — **affects price only**, features are always fully visible
- Intervals: `monthly`, `quarterly`, `semesterly`, `yearly` (longer = bigger discount)
- Features: cumulative per tier with **`+` (new, blue)** vs **`✓` (inherited, green)** markers compared to previous tier
- Free tier: shows its 3 features, no interval selector, no select button, slightly muted opacity
- Select button: bottom-aligned (`mt-auto`, `border-t`), visible only in `mode="modal"`, calls `onSelect(planId)`
- `mode="page"`: no select buttons, display only, `/pricing` route
- `mode="modal"`: selectable, used inside `CheckoutModal`

**`SubscribeButton`** — reusable component wrapping `CheckoutModal`. Import on any page (detail, upload, navbar) for a one-click subscribe CTA.

#### Checkout flow (modal only)

**Screen A — PlanGrid** → interval dropdown (price updates) → [Select] → `onSelect(planId)`

**Screen B — CryptoSelector** → crypto grid → live estimate → [Continue] → balance check → Screen C or D

**Screen C — Processing** → 3s polling, 5min timeout → success: JWT refresh + reload

**Screen D — Deposit QR** → QR + address → 5s polling, 30min timeout → success: Screen C

### Crypto Icon Library
- Use `lucide-svelte` or inline SVG for crypto currency icons (BTC, ETH, USDT, LTC).
- Icons should be visually distinct, 32×32 minimum in the selection grid.

## Comic Detail Page

Route: `/comics/:slug`

### Layout (desktop, 60/40 grid)

```
┌─── Cover Column (~60%) ──────────────┐┌─── Info Column (~40%) ─────┐
│                                       ││  Title (H1, text-2xl)       │
│  ┌─────────────────────────────────┐  ││  [published] [all ages]     │
│  │  Cover image (3:4, object-cover)│  ││  [Premium]                  │
│  │  click → opens Lightbox         │  ││                             │
│  │  cursor-zoom-in                 │  ││  ● Author Name              │
│  │  onerror → book icon fallback  │  ││  (purple circle + name)     │
│  │  [page count] badge (bottom-r)  │  ││                             │
│  └─────────────────────────────────┘  ││  Description text           │
│                                       ││  (text-sm, leading-relaxed)  │
│  ┌────┐┌────┐┌────┐┌────┐            ││                             │
│  │ P1 ││ P2 ││ P3 ││ +N │            ││  ♥ 42   ★ 18               │
│  └────┘└────┘└────┘└────┘            ││  ──────────────────────     │
│  4-column thumbnail strip             ││                             │
│  click → opens Lightbox at index      ││  Metadata (3×2 grid):       │
│  +N overflow button                   ││  👁 1.2k    ⬇ 156          │
│                                       ││  📖 32      📅 Aug 2        │
│                                       ││  ⏱ 5 min    🌍 EN         │
│                                       ││                             │
│                                       ││  [Start Reading] (emerald)  │
│                                       ││  [Download] (outline)        │
│                                       ││                             │
│                                       ││  tags (purple pills)         │
│                                       ││  Rejection reason (red box)  │
└───────────────────────────────────────┘└─────────────────────────────┘
```

**Mobile:** stack vertically — title/badges first, then cover (full-width), thumbnails, description, metadata, buttons, tags.

### Lightbox component

| Feature | Detail |
|---------|--------|
| Trigger | Click cover image or any thumbnail |
| Keyboard | ArrowLeft/ArrowRight = navigate, Escape = close |
| Mouse | Click backdrop = close, click image = next |
| Navigation | Prev/Next arrow buttons (left/right edges) |
| Dot indicators | Full set of dots, click to jump, active highlighted |
| Header | "Page N / Total" counter + close (✕) button |
| Styling | Fixed overlay, z-60, bg-black/95 |

### Components used

| Component | Source | Purpose |
|-----------|--------|---------|
| `Lightbox` | custom | Fullscreen image viewer |
| `ComicCard` | shared | Related comics grid |
| `LikeButton` / `FavoriteButton` | shared | Reactions bar |
| `Reader` | shared | Fullscreen page reader |
| `Button` | shadcn-svelte | CTAs |
| `Eye, Download, BookOpen, Clock, Globe` | lucide-svelte | Metadata icons |

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

### Admin Control Panel (`frontend-admin/` — admin.comics-galore.com)

**Layout:**
- Sticky admin navbar: Dashboard, Moderation, Users, Subscriptions, Comics
- Persistent **red banner** when plan matrix is incomplete
- Dark mode support via theme toggle

**Table pattern (all admin pages):**
```
┌─ Table ────────────────────────────────────────────────────┐
│  border border-border rounded-xl overflow-hidden           │
│                                                            │
│  ── Header row (bg-muted/50, font-medium) ──               │
│  Col 1      Col 2      Col 3      Col 4      Actions      │
│  ── Data rows (divide-y, hover:bg-muted/30) ──            │
│  Value      Value      Value      Value      [Edit][Del]  │
│  Value      Value      Value      Value      [Edit][Del]  │
│  ── Empty state: text-center py-12 text-muted ──           │
└────────────────────────────────────────────────────────────┘
```

**Loading**: Skeleton table rows (8 rows with pulsing gray bars, matching column widths).

**Actions per row:**
- **Edit**: Inline editor (role dropdown for users, status change for comics) or opens edit modal
- **Delete**: Button → inline confirmation (Confirm / Cancel), calls DELETE API, removes row
- **Create**: Global create button opens a creation modal (reuses existing forms)

**Dashboard:**
- 6 KPI cards (users, comics, pending, active subs, downloads, views)
- Lucide icons per card, skeleton cards on loading

**Moderation:**
- Pending comics queue with checkboxes + bulk approve/reject
- Select all toggle, "N selected" action bar
- Per-item: title, submit date, approve/reject buttons
- Loading: 4 skeleton cards with checkbox + title placeholders

**Users:**
- Table: Email, Role (dropdown), Tier (badge), Created date, Actions
- Inline role change via `<select>` dropdown
- Skeleton: 8 table rows with column-width matches

**Subscriptions:**
- Table: User ID (truncated UUID), Plan ID, Status (badge), Tier, Expires, Actions
- Cancel button (disabled — API pending)
- Skeleton: 6 table rows

**Comics:**
- Table: Title (link to public page), Status (badge), Views, Downloads, Created, Actions
- Delete with inline confirmation (button → Confirm/Cancel)
- Edit button (disabled — API pending)
- Skeleton: 8 table rows

## Navigation & User Modals

### Navbar (both apps)

```
┌─ Navbar ─────────────────────────────────────────────────────────┐
│  [Logo]  Browse  Pricing  Upload  ...         ⚙  ☀  [U] email   │
│                                               │      │           │
│                                               │      └─ click → Profile modal
│                                               └─ click → Settings modal
└──────────────────────────────────────────────────────────────────┘
```

**Profile modal** — opened by clicking the user's email area in navbar:
- Avatar (purple circle, first letter)
- Email + member since date
- Role badge (always shown)
- Tier badge (hidden for admin role)
- Subscription info + "Upgrade Plan" button (opens `CheckoutModal` in-place)
- **No** logout button (logout is a `LogOut` icon in the navbar, directly accessible)

**Settings modal** (public, ⚙ icon):
- Language: default language + default content language selects
- Display: items per page + popular tags limit
- **Notifications**: email from following, support replies, marketing, in-app (4 checkboxes)
- Quota boosts: +5 GB/$5, +10 GB/$8, +20 GB/$12 (display only)
- **Persistence**: saved to `users.preferences` JSONB column via `PATCH /me/preferences`.
  First load falls back to global defaults (`app_settings` table) → hardcoded defaults.

**Logout flow:**
- `LogOut` icon in navbar → click shows inline "Yes, logout" (red) + "Cancel"
- Cancel restores the icon; no modal needed

### Settings Persistence (DB-backed)

Three-tier fallback — zero extra queries on auth: users.preferences → app_settings → hardcoded.

| Endpoint | Auth | Purpose |
|----------|:---:|---------|
| `GET /me/preferences` | Auth | User prefs merged with global defaults |
| `PATCH /me/preferences` | Auth | Saves user-specific overrides |
| `GET /admin/settings` | Admin | Returns all global system settings |
| `PATCH /admin/settings` | Admin | Updates global defaults |

### Admin Settings Page (`/settings`)

Full system-wide form (7 Card sections) loaded via `GET /admin/settings`:
- **Site**: name, maintenance toggle, registrations toggle
- **Content**: language selects, max upload size, image serving mode
- **Display**: items per page, popular tags limit
- **Quotas**: per-tier GB inputs (Free/Bronze/Silver/Gold/Platinum)
- **Quota Boosts**: 3 boost prices editable
- **Security**: rate limit, email verification toggle, S3/CF presigned TTL

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
