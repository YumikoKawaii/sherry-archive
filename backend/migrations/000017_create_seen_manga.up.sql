CREATE TABLE seen_manga (
    identity_id UUID        NOT NULL,
    manga_id    UUID        NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    seen_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (identity_id, manga_id)
);

CREATE INDEX idx_seen_manga_identity_id ON seen_manga(identity_id);
