# ADR 0023 – Series Association at Comic Creation

## Status
Accepted

## Context
Series existed as a first-class entity (the `series` table + `/series` pages) and
the home page already surfaces series rails, but uploaders had **no way to
attach a comic to a series** — `POST /comics` had no series field, and
`CreateSeries` wasn't reachable from the upload flow. Comics were only linked to
series via the dev seed.

## Decision

1. **Series association lives in `POST /comics`.** `CreateComicParams` accepts
   either `series_id` (attach to an existing series) or `series_title` (create a
   new series inline). When creating inline, the comic's `genre`/`category` seed
   the new series and optional `series_genre`/`series_category`/
   `series_schedule_day` override them; the comic's cover becomes the series
   `cover_key`.

2. **`series_order` is auto-incremented** (`MAX(series_order)+1`) so issues stay
   in reading order without manual bookkeeping.

3. **Series engagement is derived, not hand-entered.** After attachment the
   series `views_count`/`hearts_count` are re-aggregated from its comics (same
   backfill as migration `22_add_series_engagement`).

4. **`CreateSeries` gains `cover_key`** so a standalone series can be created
   with a cover.

5. **Series discovery is server-side.** New public `GET /series-search`
   (`search`, `category`, `page`, `limit` — DB-filtered) and
   `GET /series-categories` power the "Browse Series" page and the upload form's
   `SeriesPicker` (search + category + "load more"). `GET /admin/series` gives
   admin a paginated/sortable/filterable datalist with a details drawer.

## Consequences
- Uploaders can create comics that belong to a series (or start one) without a
  separate step.
- Series lists can scale (search + pagination instead of loading every series).
- No new tables; uses existing `series` + `comics.series_id`/`series_order`.

## References
- ADR `0024-comic-archive-build.md`, `docs/api.md` (Series section),
  `docs/architecture.md` (Series association).
