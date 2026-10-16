-- The script library. A script is owned by the identity user who wrote it, and
-- goes when they do.
--
-- enabled is the admin kill switch, distinct from is_public: a disabled script
-- is hidden from everyone but admins, its owner included.
CREATE TABLE IF NOT EXISTS play_scripts (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    author          TEXT,
    is_public       BOOLEAN NOT NULL DEFAULT FALSE,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_user TEXT NOT NULL REFERENCES identity_users (id) ON DELETE CASCADE,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS play_scripts_created_by_user_idx ON play_scripts (created_by_user);

-- position is the order the author listed the cast in.
CREATE TABLE IF NOT EXISTS play_script_cast_members (
    id             TEXT PRIMARY KEY,
    play_script_id TEXT NOT NULL REFERENCES play_scripts (id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    description    TEXT,
    position       INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS play_script_cast_members_play_script_id_idx ON play_script_cast_members (play_script_id);

-- Several lines may share a line_index, which is overlapping dialogue, so
-- position keeps the order the author wrote them in within one index.
CREATE TABLE IF NOT EXISTS play_script_lines (
    id              TEXT PRIMARY KEY,
    play_script_id  TEXT NOT NULL REFERENCES play_scripts (id) ON DELETE CASCADE,
    line_text       TEXT NOT NULL,
    line_index      INTEGER NOT NULL,
    position        INTEGER NOT NULL,
    direction       TEXT,
    scroll_mode     TEXT NOT NULL DEFAULT 'auto' CHECK (scroll_mode IN ('auto', 'manual')),
    auto_scroll_wpm INTEGER
);

CREATE INDEX IF NOT EXISTS play_script_lines_play_script_id_idx ON play_script_lines (play_script_id);

-- More than one speaker on a line is unison.
CREATE TABLE IF NOT EXISTS play_script_line_speakers (
    line_id        TEXT NOT NULL REFERENCES play_script_lines (id) ON DELETE CASCADE,
    cast_member_id TEXT NOT NULL REFERENCES play_script_cast_members (id) ON DELETE CASCADE,
    position       INTEGER NOT NULL,
    PRIMARY KEY (line_id, cast_member_id)
);

-- A tag's name is unique without regard to case, so "Drama" and "drama" are
-- one tag.
CREATE TABLE IF NOT EXISTS script_tags (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE COLLATE NOCASE,
    description TEXT,
    created_at  DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS play_script_tags (
    play_script_id TEXT NOT NULL REFERENCES play_scripts (id) ON DELETE CASCADE,
    script_tag_id  TEXT NOT NULL REFERENCES script_tags (id) ON DELETE CASCADE,
    PRIMARY KEY (play_script_id, script_tag_id)
);

-- One rating per user per script; a re-rating replaces it.
CREATE TABLE IF NOT EXISTS play_script_ratings (
    play_script_id TEXT NOT NULL REFERENCES play_scripts (id) ON DELETE CASCADE,
    user_id        TEXT NOT NULL REFERENCES identity_users (id) ON DELETE CASCADE,
    rating         INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    created_at     DATETIME NOT NULL,
    updated_at     DATETIME NOT NULL,
    PRIMARY KEY (play_script_id, user_id)
);
