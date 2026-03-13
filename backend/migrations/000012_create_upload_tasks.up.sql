CREATE TYPE upload_task_type AS ENUM ('zip', 'oneshot_zip');
CREATE TYPE upload_task_status AS ENUM ('pending', 'processing', 'done', 'failed');

CREATE TABLE upload_tasks (
    id         UUID               NOT NULL PRIMARY KEY,
    type       upload_task_type   NOT NULL,
    status     upload_task_status NOT NULL DEFAULT 'pending',
    owner_id   UUID               NOT NULL REFERENCES users(id),
    manga_id   UUID               NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    chapter_id UUID               REFERENCES chapters(id) ON DELETE SET NULL,
    s3_key     TEXT               NOT NULL,
    error      TEXT               NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);
