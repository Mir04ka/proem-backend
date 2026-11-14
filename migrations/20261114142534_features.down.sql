DROP INDEX IF EXISTS idx_articles_moderation_pending;
DROP INDEX IF EXISTS idx_articles_public_feed;

ALTER TABLE articles DROP COLUMN IF EXISTS preview;
ALTER TABLE articles DROP COLUMN IF EXISTS moderation_reason;
ALTER TABLE articles DROP COLUMN IF EXISTS moderation_status;
ALTER TABLE articles DROP COLUMN IF EXISTS visibility;

DROP INDEX IF EXISTS idx_comments_parent_id;
ALTER TABLE comments DROP COLUMN IF EXISTS parent_id;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_format;
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(50);
ALTER TABLE users DROP COLUMN IF EXISTS role;