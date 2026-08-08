ALTER TABLE comics ADD COLUMN dislike_count INT NOT NULL DEFAULT 0;

CREATE TABLE dislikes (
    user_id UUID NOT NULL,
    comic_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, comic_id)
);

CREATE INDEX idx_dislikes_comic ON dislikes(comic_id);
CREATE INDEX idx_dislikes_user ON dislikes(user_id);
