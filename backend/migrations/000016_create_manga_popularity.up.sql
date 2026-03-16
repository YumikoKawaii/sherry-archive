CREATE TABLE manga_popularity (
    manga_id   UUID        NOT NULL PRIMARY KEY REFERENCES mangas(id) ON DELETE CASCADE,
    score      FLOAT       NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL
);
