-- Create enum type for scroll mode
CREATE TYPE scroll_mode_enum AS ENUM ('auto', 'manual');

-- play_scripts table (main entity)
CREATE TABLE IF NOT EXISTS play_scripts (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    created_by_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    belongs_to_account TEXT NOT NULL REFERENCES accounts("id") ON DELETE CASCADE
);

-- play_script_cast_members table
CREATE TABLE IF NOT EXISTS play_script_cast_members (
    id TEXT NOT NULL PRIMARY KEY,
    play_script_id TEXT NOT NULL REFERENCES play_scripts("id") ON DELETE CASCADE,
    name TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- play_script_lines table
CREATE TABLE IF NOT EXISTS play_script_lines (
    id TEXT NOT NULL PRIMARY KEY,
    play_script_id TEXT NOT NULL REFERENCES play_scripts("id") ON DELETE CASCADE,
    cast_member_id TEXT NOT NULL REFERENCES play_script_cast_members("id") ON DELETE CASCADE,
    line_text TEXT NOT NULL,
    line_index INTEGER NOT NULL,
    direction TEXT,
    scroll_mode scroll_mode_enum NOT NULL DEFAULT 'auto',
    auto_scroll_wpm INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
