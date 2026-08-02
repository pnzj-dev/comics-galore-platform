# ADR 0011 – Desktop Client with Wails + Shared Svelte UI

## Status
Accepted

## Context
We want a native desktop experience while avoiding a second UI codebase. Svelte already powers the web client.

## Decision
- Build the desktop application with **Wails** (Go + webview).
- Extract all reusable Svelte UI, forms, stores and client logic into `packages/ui`.
- Both the SvelteKit web app and the Wails frontend import from `packages/ui`.
- The Encore backend remains the single API; the Wails Go side is only for native shell features and does not duplicate business logic.

## Consequences
- Maximum UI reuse.
- Desktop gains native windowing, menus and file dialogs.
- Slightly more monorepo complexity, offset by far less duplicated frontend work.
- Official Wails CLI must be used for the desktop project scaffolding.
