---
name: tier-gated-gallery
description: Build a Svelte lightbox/gallery for Comics Galore where visible images are limited by subscription tier and extra images are blurred with an upgrade CTA. Use when implementing comic previews, cover galleries, or tier-gated image carousels.
---

# Tier-gated gallery / lightbox (Svelte)

## Goal

Port the pure TS gallery in `references/gallery.ts` into a **Svelte** (shadcn-svelte + Tailwind) component used on public comic detail / preview surfaces.

Behavior:

- Show a main image, dots, thumbnail strip, prev/next, fullscreen lightbox (keyboard + swipe).
- Only the first **N** images are sharp and fully usable, where **N** comes from the user’s tier (`max_preview_pages` or a dedicated gallery limit).
- Images beyond N stay in the grid/lightbox structure but are **blurred** (and non-interactive or locked) with copy that invites **upgrade** (open plans modal).
- Anonymous / free users use the free tier limit.

## Reference

Adapt structure and UX from:

`references/gallery.ts`

Keep:

- Inline main image + dots + thumbs
- Fullscreen overlay with counter, caption, Esc / arrows, touch swipe
- Accessible labels on controls

Do **not** require that file at runtime — it is a behavior reference for a Svelte rewrite.

## Tier gating rules

1. Resolve `previewLimit` from the current user’s tier (Encore client / auth store). If logged out, use free tier limit from public tiers endpoint or settings.
2. For index `i`:
   - `i < previewLimit` → full resolution (via media URL resolver), normal interactions.
   - `i >= previewLimit` → blurred presentation; clicking opens upgrade CTA instead of fullscreen (or opens fullscreen still blurred with persistent upgrade banner — prefer **no sharp fullscreen** for gated indices).
3. Thumbs for gated indices also blurred + lock affordance.
4. Dots may remain for all images so users see total count (social proof), or only for unlocked — prefer **show all dots** so the catalog size is visible.
5. Never ship ungated full-resolution URLs in the DOM for locked pages if the backend can withhold them. Prefer:
   - Backend returns locked slots as `{ locked: true, placeholderSrc? }` without real page URLs, **or**
   - Low-res/blurred derivative only for locked indices.
6. Upgrade CTA uses the same plans modal flow as the rest of the app (i18n keys).

## Svelte component sketch

Suggested public API:

```ts
type GalleryImage = {
  src?: string;       // omit or use placeholder when locked
  alt?: string;
  caption?: string;
  locked?: boolean;   // true when above tier limit
};

props:
  images: GalleryImage[]
  previewLimit: number  // from tier
  onUpgrade?: () => void
  initialIndex?: number
```

Implementation notes:

- Prefer Svelte 5 runes or idiomatic SvelteKit patterns used in this repo.
- Tailwind only (match existing design tokens / dark mode).
- Use `aria-modal` on lightbox; restore focus on close.
- Blur via CSS (`blur-md`, `pointer-events-none` on image, overlay with upgrade text).
- i18n for all user-visible strings (i18n agent keys).

## Backend alignment

- Comic detail / reader preview endpoints should respect tier `max_preview_pages` (or gallery-specific perk).
- Asset `kind` (`preview` | `page` | `cover`) unchanged; gating is about **how many** are revealed, not storage.
- Media resolver still applies (`direct` / later imgproxy / CF).

## V1 scope

- Component + free/paid limit wiring on **web** is appropriate for SOON if not in the thin v1 spine; a simpler single-cover + “upgrade for more previews” is enough for strict v1.
- Do not block v1 exit on a full multi-image gated lightbox unless Product marks it IN.

## QA checklist

- Limit 0 / 1 / many images
- Limit equal to length (nothing blurred)
- Limit less than length (blur + CTA)
- Keyboard and swipe do not reveal locked sharp images
- Logged-out uses free tier limit
- Upgrade callback opens plans modal
- Dark mode + reduced motion if applicable

## References

- `references/gallery.ts` — baseline gallery UX
- `docs/product.md` — tiers, max preview pages
- `docs/v1-scope.md` — phase gate
- skills `sveltekit-ui`, `nowpayments-billing` (upgrade path)
