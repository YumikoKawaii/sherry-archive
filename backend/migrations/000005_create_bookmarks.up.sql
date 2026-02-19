CREATE TABLE bookmarks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id         UUID NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    chapter_id       UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    last_page_number INT NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_bookmark_user_manga UNIQUE (user_id, manga_id)
);

CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
