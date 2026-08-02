---
name: wails-desktop
description: Implement Comics Galore Wails desktop client with shared Svelte UI and offline CBZ library. Use when working on desktop shell, offline downloads, tray, or native integrations. Not for v1 web spine work.
---

# Wails desktop – Comics Galore

## Scope warning

Desktop is **LATER** relative to `docs/v1-scope.md`. Do not prioritize this over the web v1 spine.

## Architecture

- Wails Go shell for native FS, tray, notifications, hotkeys, file dialogs.
- Svelte UI reuses `packages/ui` shared with web.
- Business data still from Encore API (no second backend).

## Offline

- User-chosen folder; per-comic offline as local `.cbz`.
- Reader prefers local CBZ when present.
- Index and file I/O on Go side.

## Native (when in scope)

- Tray, start-with-OS, notifications, global hotkeys
- Open-with CBZ/CBR, drag-and-drop, Jump List
- Fullscreen reader polish

## References

- ADR `0011-wails-desktop.md`, `0012-desktop-offline.md`, `docs/v1-scope.md` (LATER section)
