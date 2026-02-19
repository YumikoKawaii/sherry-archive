DROP INDEX IF EXISTS idx_mangas_category;
DROP INDEX IF EXISTS idx_mangas_artist;
DROP INDEX IF EXISTS idx_mangas_author;

ALTER TABLE mangas
  DROP COLUMN category,
  DROP COLUMN artist,
  DROP COLUMN author;
