-- script_tags table
CREATE TABLE IF NOT EXISTS script_tags (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- play_script_tags junction table (many-to-many with soft delete)
CREATE TABLE IF NOT EXISTS play_script_tags (
    play_script_id TEXT NOT NULL REFERENCES play_scripts(id) ON DELETE CASCADE,
    script_tag_id TEXT NOT NULL REFERENCES script_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (play_script_id, script_tag_id)
);

-- Partial unique index: one active tag assignment per script+tag
CREATE UNIQUE INDEX IF NOT EXISTS idx_play_script_tags_active ON play_script_tags (play_script_id, script_tag_id) WHERE archived_at IS NULL;

-- Index for lookup by tag name
CREATE INDEX IF NOT EXISTS idx_script_tags_name ON script_tags(name);

-- Index for finding scripts by tag
CREATE INDEX IF NOT EXISTS idx_play_script_tags_script_tag_id ON play_script_tags(script_tag_id) WHERE archived_at IS NULL;
