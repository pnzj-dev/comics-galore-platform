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
users, tiers, plans, series, tags, likes, favorites, ratings, comments, flags, reading_progress, reading_lists, follows, download_logs, subscriptions, webhook_events, settings, support_tickets, conversations, messages, audit_logs, notifications, deposits…

## Billing & Payments

### tiers
```sql
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
name            TEXT NOT NULL                    -- Free / Bronze / Silver / Gold / Platinum
description     TEXT NOT NULL DEFAULT ''
sort_order      INT NOT NULL DEFAULT 0
created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
```

Seed data (5 tiers):
| Name | Sort | Description |
|------|------|-------------|
| Free | 0 | Limited access, browse and read comments |
| Bronze | 1 | Entry-level paid tier |
| Silver | 2 | Mid-range tier |
| Gold | 3 | Premium tier |
| Platinum | 4 | Top tier with full access |

### plans
```sql
id                      UUID PRIMARY KEY DEFAULT gen_random_uuid()
tier_id                 UUID NOT NULL REFERENCES tiers(id)
name                    TEXT NOT NULL
interval                TEXT NOT NULL
                        CHECK (interval IN ('monthly', 'quarterly', 'yearly'))
price_usd_cents         INT NOT NULL DEFAULT 0
currency                TEXT NOT NULL DEFAULT 'USD'
features                JSONB DEFAULT '[]'          -- cumulative string array
quota_downloads         INT NOT NULL DEFAULT 0
is_active               BOOLEAN NOT NULL DEFAULT true
provider_plan_id        TEXT                        -- NowPayments plan ID
provider_interval_days  INT DEFAULT 0               -- for NowPayments sync
created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
```

Interval rules:
- No `lifetime` interval.
- Free tier: only `monthly` plan (price 0).
- Paid tiers: `monthly`, `quarterly`, `yearly` plans each.
- Interval pricing multipliers: quarterly (0.9×), yearly (0.8×).

Seed plans (3 intervals × 4 paid tiers = 12 paid plans + 1 free):

| Tier | Monthly | Quarterly (0.9×) | Yearly (0.8×) |
|------|---------|-----------------|---------------|
| Free | $0 | — | — |
| Bronze | $3.99 | $10.77 | $38.84 |
| Silver | $6.99 | $18.87 | $68.06 |
| Gold | $9.99 | $26.97 | $97.22 |
| Platinum | $19.99 | $53.97 | $193.90 |

Features are cumulative per tier (higher tiers inherit all lower-tier features):

| Tier | Features |
|------|----------|
| Free | Browse comics, Read comments, 1 GB download quota |
| Bronze | + Write comments, Download archives, Web reader, Full preview gallery, 10 GB download quota |
| Silver | + 50 GB download quota |
| Gold | + Premium & exclusive posts, 200 GB download quota |
| Platinum | + 1 TB download quota |

### subscriptions
```sql
id                        UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id                   UUID NOT NULL
plan_id                   UUID NOT NULL
provider                  TEXT NOT NULL DEFAULT 'nowpayments'
provider_subscription_id  TEXT
provider_invoice_id       TEXT
status                    TEXT NOT NULL DEFAULT 'pending'
                          CHECK (status IN ('pending', 'active', 'expired', 'cancelled', 'failed'))
active                    BOOLEAN NOT NULL DEFAULT false   -- set to true only on confirmed webhook
tier                      TEXT NOT NULL                    -- the tier this subscription grants
activated_at              TIMESTAMPTZ
expires_at                TIMESTAMPTZ
created_at                TIMESTAMPTZ NOT NULL DEFAULT now()
updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
```

### deposits
```sql
id                  UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id             UUID NOT NULL
provider            TEXT NOT NULL DEFAULT 'nowpayments'
provider_deposit_id TEXT
amount_crypto       TEXT
currency_crypto     TEXT NOT NULL
amount_usd_cents    INT
status              TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'completed', 'expired', 'failed'))
pay_address         TEXT
qr_code_url         TEXT
created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
completed_at        TIMESTAMPTZ
```

### webhook_events
```sql
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
provider        TEXT NOT NULL DEFAULT 'nowpayments'
event_type      TEXT NOT NULL                        -- subscription | deposit
external_id     TEXT                                 -- NowPayments subscription/deposit ID
payload         JSONB NOT NULL                       -- full raw webhook body
created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
```

### users (billing additions)
```sql
sub_partner_id  TEXT                        -- NowPayments sub-partner account ID
tier            TEXT NOT NULL DEFAULT 'free'
                CHECK (tier IN ('free', 'bronze', 'silver', 'gold', 'platinum'))
```

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

Global application settings stored as a JSON blob in `app_settings` (key `'defaults'`). Includes:
- `site_name` — public-facing site name
- `maintenance_mode` — boolean, blocks non-admin access
- `registrations_open` — boolean, toggles registration
- `contact_email` — from-address for Resend transactional emails
- `image_serving_mode` = `direct` | `imgproxy` | `cloudflare_images` (default `direct`)
- `imgproxy_base_url`
- `cloudflare_images_account` / delivery base (secrets via env, non-secret config in settings)
- `require_email_verify` — boolean, email verification gate
- `hide_mature_default` — boolean, default for anonymous users
- `enable_comments` — boolean, global comment kill switch
- `default_meta_description` — SEO fallback for public pages
- Rate limit, presigned TTLs, per-tier quotas, boost prices (see `AppSettings` struct)


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
