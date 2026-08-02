# ADR 0002 – Frontend Stack (Historical)

## Status
**Superseded** by ADR 0002-sveltekit.md

## Context
Originally React + Vite + TanStack + shadcn/ui was chosen.

## Decision (original)
React was selected for its ecosystem and the complexity of the admin side.

## Reason for change
React is heavier than necessary for the public comics browsing experience. SvelteKit provides better default performance and SEO characteristics while still supporting a rich admin panel via shadcn-svelte.
