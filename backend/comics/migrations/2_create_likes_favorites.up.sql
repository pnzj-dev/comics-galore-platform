CREATE TABLE likes (
    user_id UUID NOT NULL,
    comic_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, comic_id)
);

CREATE TABLE favorites (
    user_id UUID NOT NULL,
    comic_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, comic_id)
);

CREATE INDEX idx_likes_comic ON likes(comic_id);
CREATE INDEX idx_favorites_comic ON favorites(comic_id);
CREATE INDEX idx_favorites_user ON favorites(user_id);
