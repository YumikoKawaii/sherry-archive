CREATE TYPE manga_type AS ENUM ('series', 'oneshot');
ALTER TABLE mangas ADD COLUMN type manga_type NOT NULL DEFAULT 'series';
