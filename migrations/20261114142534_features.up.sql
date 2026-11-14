-- Nickname
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(20);
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_format;
ALTER TABLE users ADD CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9_]{3,20}$');

-- Comments tree
ALTER TABLE comments ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES comments(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);

-- Moderation
ALTER TABLE articles ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) NOT NULL DEFAULT 'unlisted';
ALTER TABLE articles ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(20) NOT NULL DEFAULT 'none';
ALTER TABLE articles ADD COLUMN IF NOT EXISTS moderation_reason TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS preview TEXT NOT NULL DEFAULT '';


UPDATE articles
SET visibility = 'public', moderation_status = 'approved'
WHERE is_published = TRUE AND visibility = 'unlisted' AND moderation_status = 'none';

CREATE INDEX IF NOT EXISTS idx_articles_public_feed
    ON articles(created_at DESC)
    WHERE is_published = TRUE AND visibility = 'public' AND moderation_status = 'approved';

CREATE INDEX IF NOT EXISTS idx_articles_moderation_pending
    ON articles(created_at ASC)
    WHERE is_published = TRUE AND moderation_status = 'pending';