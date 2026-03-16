CREATE INDEX IF NOT EXISTS idx_mangas_tags     ON mangas USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_mangas_author   ON mangas(author);
CREATE INDEX IF NOT EXISTS idx_mangas_category ON mangas(category);
