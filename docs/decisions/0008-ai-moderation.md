# ADR 0008 – AI Moderation

## Status
Accepted

## Context
Human moderators alone may not scale. We want optional AI assistance.

## Decision
- AI moderation is completely optional and can be disabled.
- Admin chooses the LLM model (any supported / OpenAI-compatible endpoint) and the prompt from the control panel.
- AI runs as a background step on new comments (and optionally on flags).
- Human moderators always retain final authority.
- Configuration lives in the global `settings` store.

## Consequences
- No hard dependency on a specific provider.
- Easy to turn off if costs or quality are unacceptable.
