-- stages table (main entity for real-time stage sessions)
CREATE TABLE IF NOT EXISTS stages (
    id TEXT NOT NULL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,  -- 6-char join code
    play_script_id TEXT REFERENCES play_scripts(id) ON DELETE SET NULL,
    current_line_index INTEGER NOT NULL DEFAULT 0,
    phase TEXT NOT NULL DEFAULT 'lobby',  -- lobby, countdown, performing, discussion
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE
);

-- stage_participants table (participants in a stage session)
CREATE TABLE IF NOT EXISTS stage_participants (
    id TEXT NOT NULL PRIMARY KEY,
    stage_id TEXT NOT NULL REFERENCES stages(id) ON DELETE CASCADE,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,  -- NULL for anonymous participants
    display_name TEXT NOT NULL,
    is_director BOOLEAN NOT NULL DEFAULT FALSE,
    is_ready BOOLEAN NOT NULL DEFAULT FALSE,
    is_audience BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE  -- NULL if still connected
);

-- stage_participant_roles table (role assignments for participants)
CREATE TABLE IF NOT EXISTS stage_participant_roles (
    stage_participant_id TEXT NOT NULL REFERENCES stage_participants(id) ON DELETE CASCADE,
    cast_member_id TEXT NOT NULL,
    PRIMARY KEY (stage_participant_id, cast_member_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_stages_code ON stages(code) WHERE archived_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stages_archived_at ON stages(archived_at) WHERE archived_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stage_participants_stage_id ON stage_participants(stage_id);
CREATE INDEX IF NOT EXISTS idx_stage_participants_user_id ON stage_participants(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_stage_participants_active ON stage_participants(stage_id) WHERE left_at IS NULL;
