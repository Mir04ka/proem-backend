DROP INDEX IF EXISTS idx_tags_name_lower;
ALTER TABLE articles DROP COLUMN IF EXISTS views_count;