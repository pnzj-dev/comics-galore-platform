# ADR 0021 – Social Engagement (Reading Lists & Recommendations)

## Status
Proposed

## Context
Beyond likes/favorites/follows, the product vision includes shareable public shelves and a "People also liked" discovery rail. These extend the comics service and reading lists without new infrastructure.

## Decision

### Public reading lists (shelves)
- `reading_lists(id, user_id, name, is_public, created_at)` and `reading_list_items(list_id, comic_id, position)` in the comics service.
- `POST/GET /reading-lists`, `GET /reading-lists/:id` (public if `is_public`), `POST /reading-lists/:id/items`, `DELETE /reading-lists/:id/items/:comicId`.

### "People also liked"
- `GET /comics/:id/related` — co-occurrence over `likes`/`favorites` (no new table); returns a short ranked list for the comic detail rail.

## Consequences
- No new service or database.
- Shelves are tier-agnostic (public read; only owner mutates).
- Recommendations are computed on read with a bounded query; heavier personalization is deferred.

## References
- ADR `0016-service-communication.md`, `docs/product.md` (Social & library)
