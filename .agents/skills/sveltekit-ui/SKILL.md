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

## References

- `docs/ui.md`, `docs/v1-scope.md`, ADR `0002-sveltekit.md`
