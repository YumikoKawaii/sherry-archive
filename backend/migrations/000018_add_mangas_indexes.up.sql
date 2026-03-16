CREATE INDEX idx_mangas_tags     ON mangas USING GIN(tags);
CREATE INDEX idx_mangas_author   ON mangas(author);
CREATE INDEX idx_mangas_category ON mangas(category);
