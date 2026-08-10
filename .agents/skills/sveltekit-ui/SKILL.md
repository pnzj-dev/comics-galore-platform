---
name: sveltekit-ui
description: Build Comics Galore SvelteKit web UI with shadcn-svelte, Tailwind, Superforms, and Zod. Use when creating pages, components, forms, or admin datalists on the web client.
---

# SvelteKit UI – Comics Galore

## Stack

- SvelteKit + Tailwind + shadcn-svelte (official CLIs)
- Superforms + Zod for forms
- TanStack Table (Svelte) for admin tables
- Charts only where needed (ECharts or Chart.js)

## Rules

- Do not introduce React or shadcn/ui (React).
- Prefer shared code under `packages/ui` when desktop will consume the same components.
- Public routes lean on SSR and minimal JS.
- Admin routes may be richer.
- **Data loading is done via Encore generated client** — all `+page.server.ts` files use `getEncoreClient(token)` with `try/catch` for fallback empty data. No raw `fetch()` or legacy `api.get()` remains. Migration complete.

## V1 screens priority

Public home/detail/reader, uploader manual create, admin pending queue and basic lists, auth flows, Terms/Privacy stubs.

## i18n

- Message catalogs; default `en`.
- Respect `users.ui_locale` and enabled locales from settings.
- Comic forms require content language.

## Patterns

- Resolve image URLs via media helper (mode `direct` in v1).
- Upgrade CTAs open the plans modal flow.
- Empty and quota-blocked states must be explicit.

### Svelte 5 Runes & Common Warnings

**`$state(prop)` only captures initial value**
When initializing mutable state from a prop, the value is captured once. If the prop must stay reactive, use `$derived()`. If the initial-value-only behavior is intentional, suppress:
```svelte
<script>
  // Intentionally initial value only — suppress warning
  // svelte-ignore state_referenced_locally
  let count = $state(initialCount);

  // Reactive — recomputes when the prop changes
  const derived = $derived(propValue * 2);
</script>
```

**`<svelte:window>` cannot be inside `{#if}` blocks**
Move the window listener outside the conditional and check the flag in the handler:
```svelte
<script>
  let open = $state(false);
  function onKeydown(e: KeyboardEvent) {
    if (!open) return;       // guard inside handler
    if (e.key === 'Escape') open = false;
  }
</script>

<svelte:window onkeydown={onKeydown} />  <!-- always rendered, guarded by flag -->

{#if open}
  <div>modal content</div>
{/if}
```
Alternative: use `onkeydown` on the backdrop `<div>` directly.

**Event handlers must be JS expressions, not strings (Svelte 5)**
```svelte
✗  <img onerror="this.style.display='none'" />   // string — fails in Svelte 5
✓  <img onerror={(e) => { const img = e.target; img.style.display = 'none' }} />
```

**Accessibility: dialogs**
```svelte
✗  <div role="dialog" onclick={onClose}>         // missing tabindex
✓  <div role="dialog" tabindex="-1" onclick={onClose} onkeydown={onKeydown}>
```

**Accessibility: clickable non-interactive elements**
Prefer `<button>` over `<div>` with `onclick`. If a `<div>` is necessary, add `role="button" tabindex="0" onkeydown`.
```svelte
✗  <div class="card" onclick={handleClick}>       // no role, no keyboard
✓  <button class="card text-left w-full bg-transparent" onclick={handleClick}>
```
The inner content div can use `<div role="presentation" onclick={(e) => e.stopPropagation()}>`.

**Accessibility: carousels / regions with keyboard nav**
When a `<div>` needs `tabindex="0"` + `onkeydown` for interactive regions (image carousels, galleries), suppress the false-positive a11y warnings since `role="region"` makes it accessible:
```svelte
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="carousel" onkeydown={handleKeydown} tabindex="0" role="region" aria-label="Image carousel">
```
These two warnings fire because Svelte cannot statically verify `role="region"` constitutes proper interactivity — the region role paired with keyboard handlers is a valid ARIA pattern. Prefer a `<button>` wrapper when possible; use these ignores only for carousels/galleries.

### Inline SVG for Toggleable Icons

When an icon needs to switch between filled/outlined states (like, dislike, favorite), use inline SVG with `{#if}`:
```svelte
{#if active}
  <svg class="size-3.5" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="..."/></svg>
{:else}
  <svg class="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="..."/></svg>
{/if}
```
Prefer lucide-svelte for all non-toggle icons; use inline SVGs only when fill toggling is needed.

### Admin sidebar layout

Fixed sidebar for admin apps — no sticky top navbar:
```svelte
<aside class="w-60 flex-shrink-0 bg-slate-900 dark:bg-slate-950 text-white flex flex-col">
  <nav class="flex-1 p-3 overflow-y-auto">
    {#each navItems as item}
      <a href={item.href}
         class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm
           {isActive(item.href) ? 'bg-white/10 text-white font-medium' : 'text-slate-300 hover:text-white hover:bg-white/5'}">
        <item.icon class="size-4" />
        {item.label}
      </a>
    {/each}
  </nav>
  <div class="p-3 border-t border-slate-700/50">
    <!-- user email + sign out button at bottom -->
  </div>
</aside>
```
Active state detected from `$page.url.pathname` via `$app/state` (no `$` prefix — Svelte 5 rune).

### `$app/state` not `$app/stores`

Svelte 5: `import { page } from '$app/state'` → access as `page.params.slug` (no `$` prefix). The old `$app/stores` export is deprecated in SvelteKit 2.

### Settings with Form/JSON toggle

Segmented toggle pattern:
```svelte
<label class="{mode === 'form' ? 'bg-primary/10' : 'hover:bg-muted'}" onclick={() => switchMode('form')}>Form</label>
<label class="{mode === 'json' ? 'bg-primary/10' : 'hover:bg-muted'}" onclick={() => switchMode('json')}>JSON</label>
```
JSON mode: `<textarea>` with `JSON.stringify(settings, null, 2)`, monospace font. Validate before saving. On switching back to form, parse the textarea back. If invalid JSON, show error and stay in JSON mode.

### Upload tabs via query params

Use `goto('/upload?tab=list')` for tab switches, not local `$state`. The `+page.server.ts` reads `url.searchParams.get('tab')`, validates against allowed set (`['list', 'manual', 'archive']`), and returns `{ tab }`. Benefits: bookmarkable, browser back/forward, survives page refresh.

### Standardized submit button

Replaces individual `saved` boolean + `setTimeout` patterns:
```svelte
let submitting = $state(false);
let error = $state('');

<Button disabled={submitting}>
  {submitting ? 'Publishing...' : 'Publish Comic'}
</Button>
{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
```
Use `submitting` boolean, change button text, disable while in-flight. Show error banner on failure.

### Comment threading

Comment tree is fetched server-side as a threaded structure (root comments with nested `replies` arrays). After submit/delete/SSE event, re-fetch the full threaded tree via `ListComments()` instead of manually inserting into the array. This ensures correct nesting and oldest-first ordering.

## References

- `docs/ui.md`, `docs/v1-scope.md`, ADR `0002-sveltekit.md`
