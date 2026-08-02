# Database Design – Comics Galore

## Principles

- PostgreSQL + Encore SQL migrations only.
- SQLC preferred.
- Soft-delete where valuable.

## Important Tables for Upload & Review

### comics
```sql
id, uploader_id, series_id, title, slug, description, ...
status          TEXT NOT NULL DEFAULT 'pending_review'
                -- pending_review | published | rejected | scheduled | …
file_key        TEXT                  -- main archive or primary object
cover_key       TEXT
page_keys       JSONB                 -- optional ordered list of page objects
file_size_bytes BIGINT
min_tier_id     UUID
published_at    TIMESTAMPTZ
rejection_reason TEXT
...
```

### upload_sessions (crash recovery)
```sql
id              UUID PRIMARY KEY
user_id         UUID NOT NULL REFERENCES users(id)
mode            TEXT NOT NULL          -- manual | archive
status          TEXT NOT NULL          -- active | finalising | completed | failed | expired
metadata        JSONB                  -- partial form data or parsed JSON
s3_prefix       TEXT NOT NULL
multipart_upload_id TEXT               -- if using S3 multipart
parts           JSONB                  -- list of {part_number, etag, size, completed}
expires_at      TIMESTAMPTZ
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
```

### Other existing tables
users, tiers, intervals, plans, series, tags, likes, favorites, ratings, comments, flags, reading_progress, reading_lists, follows, download_logs, subscriptions, subscription_attempts, payments, webhook_events, settings, support_tickets, conversations, messages, audit_logs, notifications…

## Status Rules

- New comics from either creation mode are always born as `pending_review`.
- Public queries only return `status = 'published'` (and optionally `scheduled` once the time is reached).
- Uploaders can always see their own comics regardless of status.

## Bootstrap Rule

At process start: if zero admin users exist → log clear error and exit.


## Public-facing additions

### comics
- `age_rating` TEXT  -- e.g. all_ages | teen | mature | explicit
- (optional) `content_warnings` TEXT[]

### series_follows
```sql
user_id UUID NOT NULL REFERENCES users(id)
series_id UUID NOT NULL REFERENCES series(id)
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
PRIMARY KEY (user_id, series_id)
```

### notification_preferences
```sql
user_id UUID PRIMARY KEY REFERENCES users(id)
email_new_from_following BOOLEAN DEFAULT true
email_support_replies BOOLEAN DEFAULT true
email_marketing BOOLEAN DEFAULT false
in_app_enabled BOOLEAN DEFAULT true
-- extend as needed
updated_at TIMESTAMPTZ
```

### users
- `email_verified_at` TIMESTAMPTZ


## Media assets

### comic_assets (recommended)
```sql
id              UUID PRIMARY KEY
comic_id        UUID NOT NULL REFERENCES comics(id)
kind            TEXT NOT NULL  -- cover | preview | page | original
storage         TEXT NOT NULL  -- cloudflare_images | s3
external_id     TEXT           -- CF Images id or S3 key
url             TEXT           -- canonical stored URL if useful
position        INTEGER        -- page order for kind=page / preview
width           INTEGER
height          INTEGER
created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
```

Constraints:
- `kind = cover | preview` → `storage = cloudflare_images`
- `kind = page | original` → `storage = s3`

### settings
Include keys such as:
- `image_serving_mode` = `direct` | `imgproxy` | `cloudflare_images` (default `direct`)
- `imgproxy_base_url`
- `cloudflare_images_account` / delivery base (secrets via env, non-secret config in settings)


## Language fields

### comics
- `content_language` TEXT NOT NULL DEFAULT 'en'  -- ISO 639-1 (or BCP 47 tag where needed, e.g. pt-BR)
- Index for filter/facet queries on `content_language`

### users
- `ui_locale` TEXT NOT NULL DEFAULT 'en'  -- preferred UI locale (BCP 47)

### settings
- `default_ui_locale` (default `en`)
- `enabled_ui_locales` (JSON array of enabled locale codes)
- `default_content_language` (default `en`)

Archive metadata JSON (`comic.json`) should allow a `language` / `content_language` field mapped on ingest.
