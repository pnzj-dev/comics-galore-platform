CREATE TABLE IF NOT EXISTS reading_lists (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL,
    name       TEXT NOT NULL,
    is_public  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reading_lists_user ON reading_lists(user_id);

CREATE TABLE IF NOT EXISTS reading_list_items (
    list_id   UUID NOT NULL REFERENCES reading_lists(id) ON DELETE CASCADE,
    comic_id  UUID NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
    position  INT NOT NULL DEFAULT 0,
    PRIMARY KEY (list_id, comic_id)
);

CREATE INDEX IF NOT EXISTS idx_reading_list_items_list ON reading_list_items(list_id);
