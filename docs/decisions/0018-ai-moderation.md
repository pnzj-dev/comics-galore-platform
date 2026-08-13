# ADR 0018 – AI Moderation (Generic LLM Provider)

## Status
Accepted (supersedes the high-level ADR 0008)

## Context
Human moderators alone may not scale. We want optional AI assistance for new comics and comments, but with no vendor lock-in and human moderators always in control (ADR 0008).

## Decision

### Generic, configurable provider
- New shared package `backend/aiprovider/` (mirrors the `nowpayments` package pattern) exposing a minimal client for any OpenAI-compatible chat-completions endpoint.
- Configuration lives in the auth `AppSettings` blob: `ai_moderation_enabled`, `ai_model`, `ai_endpoint`, `ai_api_key` (secret), `ai_prompt`, `ai_auto_approve_threshold`, `ai_auto_reject_threshold`.
- Fully disableable; no hard dependency on any vendor.

### Processing model
- Encore `//encore:worker` subscribes to Pub/Sub topics `comic-created` and `comment-created`.
- Worker calls the AI provider, writes an `ai_decisions` row, then applies thresholds:
  - confidence ≥ approve threshold → auto-approve (log decision).
  - confidence ≤ reject threshold → auto-reject (log decision).
  - in-between → mark `uncertain` and push to `ai_review_queue` for a human.

### Data model (comics service)
- `ai_decisions(id, target_type comic|comment, target_id, decision approved|rejected|uncertain, confidence numeric, reason, model, created_at)`.
- `ai_review_queue(id, target_type, target_id, status pending|resolved, created_at, resolved_by, resolved_at)`.

### API surface (admin/moderator)
- `GET /admin/ai/queue`, `POST /admin/ai/queue/:id/resolve`, `GET /admin/ai/decisions`.

### Authority
Human moderators always retain final authority: any AI decision can be overturned, and uncertain items require a human.

## Consequences
- AI work happens off the request path via workers; the API stays responsive.
- Adding/removing a vendor is configuration-only.
- New tables and endpoints in the comics service; settings keys in auth.

## References
- ADR `0008-ai-moderation.md`, `0016-service-communication.md`, `docs/product.md` (AI Moderation)
