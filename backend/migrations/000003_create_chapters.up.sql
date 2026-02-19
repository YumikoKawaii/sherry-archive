CREATE TABLE chapters (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manga_id   UUID NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    number     NUMERIC(8,1) NOT NULL,
    title      TEXT NOT NULL DEFAULT '',
    page_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_chapter_manga_number UNIQUE (manga_id, number)
);

CREATE INDEX idx_chapters_manga_id ON chapters(manga_id);
