CREATE TABLE pages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    number     INT NOT NULL,
    object_key TEXT NOT NULL,
    width      INT NOT NULL DEFAULT 0,
    height     INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_page_chapter_number UNIQUE (chapter_id, number)
);

CREATE INDEX idx_pages_chapter_id ON pages(chapter_id);
