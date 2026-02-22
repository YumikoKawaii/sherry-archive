CREATE TABLE comments (
    id         UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    manga_id   UUID        NOT NULL REFERENCES mangas(id)   ON DELETE CASCADE,
    chapter_id UUID                 REFERENCES chapters(id) ON DELETE CASCADE,
    content    TEXT        NOT NULL CHECK (char_length(content) BETWEEN 1 AND 2000),
    edited     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_comments_manga   ON comments (manga_id, chapter_id, created_at DESC);
CREATE INDEX idx_comments_user    ON comments (user_id);
