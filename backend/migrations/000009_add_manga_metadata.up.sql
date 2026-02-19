ALTER TABLE mangas
  ADD COLUMN author   TEXT NOT NULL DEFAULT '',
  ADD COLUMN artist   TEXT NOT NULL DEFAULT '',
  ADD COLUMN category TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_mangas_author   ON mangas (author);
CREATE INDEX idx_mangas_artist   ON mangas (artist);
CREATE INDEX idx_mangas_category ON mangas (category);
