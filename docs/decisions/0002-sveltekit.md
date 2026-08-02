# ADR 0002 – Frontend: SvelteKit + shadcn-svelte

## Status
Accepted (replaces previous React decision)

## Context
The original stack used React + Vite + shadcn/ui. React is excellent for complex admin UIs but heavier than necessary for a content-oriented public comics experience. We also want a single coherent frontend.

## Decision
- Use **SvelteKit** as the single frontend framework for both public and admin.
- Use **shadcn-svelte** as the component library.
- Tailwind CSS for styling.
- TanStack Table (Svelte) for advanced admin tables.
- Apache ECharts or Chart.js for charts.
- Superforms + Zod for forms.
- Prefer the generated Encore client for data fetching.

## Consequences
- Official SvelteKit and shadcn-svelte CLIs must be used for scaffolding.
- Public routes can leverage SvelteKit SSR and progressive enhancement for better performance and SEO.
- Admin routes remain fully capable.
- No React, no original shadcn/ui.
