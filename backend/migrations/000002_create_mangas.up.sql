CREATE TYPE manga_status AS ENUM ('ongoing', 'completed', 'hiatus');

CREATE TABLE mangas (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    cover_url   TEXT NOT NULL DEFAULT '',
    status      manga_status NOT NULL DEFAULT 'ongoing',
    tags        TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mangas_owner_id ON mangas(owner_id);
CREATE INDEX idx_mangas_status ON mangas(status);
CREATE INDEX idx_mangas_tags ON mangas USING GIN(tags);
CREATE INDEX idx_mangas_created_at ON mangas(created_at DESC);
