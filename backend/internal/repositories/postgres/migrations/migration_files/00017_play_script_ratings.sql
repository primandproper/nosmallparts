-- play_script_ratings table: user ratings for play scripts (1-5)
-- One active rating per (user, play_script). Changing vote archives old and inserts new.
CREATE TABLE IF NOT EXISTS play_script_ratings (
    id TEXT NOT NULL PRIMARY KEY,
    play_script_id TEXT NOT NULL REFERENCES play_scripts(id) ON DELETE CASCADE,
    belongs_to_user TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE
);

-- One active rating per user per play script
CREATE UNIQUE INDEX IF NOT EXISTS idx_play_script_ratings_active ON play_script_ratings (play_script_id, belongs_to_user) WHERE archived_at IS NULL;

-- Index for aggregate queries (rating stats by play)
CREATE INDEX IF NOT EXISTS idx_play_script_ratings_play_script_id ON play_script_ratings(play_script_id) WHERE archived_at IS NULL;

-- Index for user lookup (user's rating for a play)
CREATE INDEX IF NOT EXISTS idx_play_script_ratings_user_play ON play_script_ratings(play_script_id, belongs_to_user) WHERE archived_at IS NULL;
