# ADR 0012 – Desktop Offline Library & Native Features (Wails)

## Status
Accepted

## Context
The desktop client should be clearly superior to the browser for reading and library management.

## Decision

### Offline
- Desktop-only offline library.
- User-chosen folder; comics stored as `.cbz`.
- Per-comic offline toggle; bulk download (series, Continue Reading); optional auto-download next issue.
- Multiple library profiles supported.
- Reader prefers local CBZ when available.
- File I/O and local index live in Wails Go; UI in Svelte.

### Native shell & integrations
- System tray, minimize-to-tray, optional start with OS.
- Native OS notifications.
- Global hotkeys.
- File associations (“Open with”) for CBZ/CBR.
- Jump List / dock menu.
- Drag-and-drop of archives or image folders onto the app.

### Reader polish
- True fullscreen distraction-free mode, optional dual-page spread.
- Touch / pen / gamepad page turns.
- Quick Look preview.
- Local reading stats; export/backup of offline library.

Web does not implement these offline or native features.

## Consequences
- Clear platform split: shared Svelte UI + desktop-only Go bindings for FS, tray, notifications, hotkeys.
- Offline path must stay reliable when the network is absent.
