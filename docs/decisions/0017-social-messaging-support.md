# ADR 0017 – Messaging, Support & Broadcasts (`social` service)

## Status
Proposed

## Context
Users need to message uploaders and each other, open support tickets, and admins need to reach users (announcements). Comments are public and threaded; messaging is private and pairwise. Support is a distinct, staff-facing queue. None of this exists yet.

Existing conventions (ADR 0016): each Encore service owns its own PostgreSQL database; cross-service communication happens only through typed private API calls or Pub/Sub. Email delivery and user preferences live in the `auth` service.

## Decision

### New service: `social`
- `backend/social/` with its own database `socialdb`.
- Owns private messaging, support tickets, and broadcasts.

### Data model
- `conversations(id, participant_a UUID, participant_b UUID, created_at, last_message_at)` — unique index on `(participant_a, participant_b)` (normalised ordering) to prevent duplicate pairwise threads.
- `messages(id, conversation_id FK, sender_id UUID, body TEXT, read_at, created_at)`.
- `support_tickets(id, user_id UUID, subject, status pending|open|resolved|closed, priority, assigned_to UUID, created_at, resolved_at)`.
- `support_messages(id, ticket_id FK, sender_id UUID, is_staff bool, body, created_at)`.
- `broadcasts(id, title, body, target all|tier, tier, sent_at, created_at)`.

### API surface (all `//encore:api auth` unless noted)
Messaging:
- `GET /messages/conversations` — list my conversations (with last message + unread count).
- `POST /messages/:userId` — get-or-create a conversation with a user.
- `GET /messages/conversation/:id` — list messages (marks as read).
- `POST /messages/conversation/:id` — send a message.
- `POST /messages/conversation/:id/read` — mark read.
- `GET /messages-stream/:userId` (raw SSE) — live new-message stream (mirrors the comments SSE pattern).

Support:
- `POST /support/tickets`, `GET /support/tickets` (my tickets), `GET /support/tickets/:id`, `POST /support/tickets/:id/reply` (user).
- Admin/moderator: `GET /admin/support/tickets`, `POST /admin/support/tickets/:id/reply`, `POST /admin/support/tickets/:id/assign`, `POST /admin/support/tickets/:id/resolve`.

Broadcasts:
- `POST /admin/broadcasts`, `GET /admin/broadcasts` (admin).

### Cross-service
- `social` calls `auth.NotifySupportReply` (private endpoint, reusing the `NotifyFollowersNewComic` pattern) to email users on staff replies, honouring the `email_support_replies` preference.
- Broadcasts likewise fan out via `auth.NotifyBroadcast` for `email_marketing`.

## Consequences
- One new database; no existing service schema changes.
- Messaging scales independently and stays out of the comics read path.
- SSE provides real-time messaging without WebSockets (consistent with comments).
- Admin gains a Support queue and Broadcasts tool in `frontend-admin`.

## References
- ADR `0016-service-communication.md`, `docs/product.md` (Roles, Admin), `docs/api.md`
