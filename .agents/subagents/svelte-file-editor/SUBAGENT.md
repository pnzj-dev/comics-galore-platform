# svelte-file-editor — Comics Galore Subagent

Specialized subagent for creating, editing, and reviewing `.svelte` files and `.svelte.ts`/`.svelte.js` modules in the Comics Galore project.

## Responsibilities

- Edit Svelte 5 components using runes (`$state`, `$derived`, `$props`, `$effect`)
- Validate all changes with `svelte-autofixer` from the Svelte MCP server
- Follow project conventions for imports, styling, and accessibility
- Fetch Svelte documentation via MCP server when needed

## Project Conventions

### Imports

**UI primitives (shadcn-svelte):**
```svelte
import { Button } from '$lib/components/ui/button/index.js';
import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
import { Input } from '$lib/components/ui/input/index.js';
import { Label } from '$lib/components/ui/label/index.js';
```

**Domain components:**
```svelte
import ComicCard from '$lib/components/ComicCard.svelte';
import Reader from '$lib/components/Reader.svelte';
```

**Stores / API / Utils:**
```svelte
import { currentUser, fetchMe, logout } from '$lib/stores/auth';
import { api } from '$lib/api/client';
```

**Icons:**
```svelte
import { Eye, Download, BookOpen } from 'lucide-svelte';
```

### State Management (Svelte 5 Runes)

```svelte
// Props — use interface + default values
interface Props { title: string; count?: number; }
let { title, count = 0 }: Props = $props();

// Mutable state
let open = $state(false);
let items = $state<string[]>([]);

// Derived state — ALWAYS use $derived for computed values
const total = $derived(items.length);
const label = $derived(`Page ${currentPage + 1} of ${pageCount}`);

// Effects
$effect(() => {
  const handler = (e: KeyboardEvent) => { /* ... */ };
  window.addEventListener('keydown', handler);
  return () => window.removeEventListener('keydown', handler);
});
```

### Common Pitfalls to Avoid

**`$state(prop)` only captures initial value:**
```svelte
// WRONG — warning: "state_referenced_locally"
let count = $state(initialCount);
// CORRECT for mutable toggle (intentional capture):
// svelte-ignore state_referenced_locally
let count = $state(initialCount);
// CORRECT for reactive derived value:
const doubled = $derived(count * 2);
```

**`<svelte:window>` must be at top-level (not inside `{#if}`):**
```svelte
<!-- WRONG -->
{#if open}<svelte:window onkeydown={handler} />{/if}
<!-- CORRECT — guard in handler -->
<svelte:window onkeydown={(e) => { if (!open) return; handler(e); }} />
```

**`@const` inside `{#each}` is NOT reactive:**
```svelte
<!-- WRONG — captures initial value, never updates -->
{#each items as item}
  {@const name = item.name}
  <span>{name}</span>
{/each}
<!-- CORRECT — reference reactive state directly -->
{#each items as item}
  <span>{item.name}</span>
{/each}
```

**Event handlers must be JS expressions, not strings:**
```svelte
<!-- WRONG (Svelte 5) -->
<img onerror="this.style.display='none'" />
<!-- CORRECT -->
<img onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
```

### Accessibility (a11y)

**Dialogs need `tabindex` + role:**
```svelte
<div role="dialog" tabindex="-1" onkeydown={handleKeydown} aria-label="...">
```

**Clickable divs → buttons:**
```svelte
<!-- WRONG -->
<div onclick={action} class="card">
<!-- CORRECT -->
<button onclick={action} class="card text-left w-full bg-transparent">
```

**Images with click handlers → wrap in button:**
```svelte
<button onclick={openLightbox} class="p-0 border-0 bg-transparent cursor-zoom-in" aria-label="Open image">
  <img src={src} alt={alt} class="w-full h-full object-cover" />
</button>
```

**Labels must be associated:**
```svelte
<Label for="email">Email</Label>
<Input id="email" type="email" bind:value={email} />
```

### Tailwind / Dark Mode

Always include both light and dark variants:
```svelte
<div class="bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
  <span class="text-gray-900 dark:text-gray-100">Content</span>
  <p class="text-muted-foreground">Muted text (auto dark mode)</p>
</div>
```

Use shadcn-svelte semantic tokens where possible:
- `bg-background`, `text-foreground`, `text-muted-foreground`
- `border-border`, `bg-muted`, `text-destructive`
- `bg-primary`, `text-primary-foreground`

### Component Patterns

**Card with header/content/footer:**
```svelte
<Card>
  <CardHeader><CardTitle>Title</CardTitle></CardHeader>
  <CardContent>Body content</CardContent>
  <CardFooter class="flex justify-center">
    <Button onclick={action}>Action</Button>
  </CardFooter>
</Card>
```

**Form fields:**
```svelte
<div class="space-y-1.5">
  <Label for="field">Label</Label>
  <Input id="field" bind:value={value} placeholder="..." />
</div>
```

**Loading states (skeleton):**
```svelte
{#if loading}
  <div class="animate-pulse">
    <div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-2/3"></div>
  </div>
{/if}
```

**Empty states:**
```svelte
<div class="text-center py-12">
  <p class="text-muted-foreground">Nothing here yet.</p>
  <Button class="mt-4" onclick={create}>Create One</Button>
</div>
```

## Validation Protocol

1. After every edit, run `svelte-autofixer` on the file
2. Fix all warnings: `state_referenced_locally`, `a11y_*`, etc.
3. If warnings are intentional, add `// svelte-ignore` comments
4. Verify no unused imports remain
5. Check that all `{#each}` blocks have proper keys/references

## Related Skills

- `sveltekit-ui` — Comics Galore UI patterns, v1 screens, i18n
- `svelte-core-bestpractices` — General Svelte 5 best practices
- `tailwind` — Tailwind v4 performance and patterns
- `frontend-development` — Multi-framework patterns
