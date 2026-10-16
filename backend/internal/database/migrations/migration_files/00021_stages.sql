-- A stage is one live room: who is on it, which play, where in it they are.
-- The host is an identity rather than a participant, so they keep the role
-- across a reconnect: a user, or (only ever cleared, never elected) nobody.
-- host_guest_id is kept beside host_user_id because a host set by a transfer
-- names whoever the target was, and the identity match reads both.
CREATE TABLE IF NOT EXISTS stages (
    id                 TEXT PRIMARY KEY,
    code               TEXT NOT NULL UNIQUE,
    play_script_id     TEXT REFERENCES play_scripts (id) ON DELETE SET NULL,
    current_line_index INTEGER NOT NULL DEFAULT 0,
    phase              TEXT NOT NULL DEFAULT 'lobby' CHECK (phase IN ('lobby', 'countdown', 'performing', 'discussion')),
    host_user_id       TEXT REFERENCES identity_users (id) ON DELETE SET NULL,
    host_guest_id      TEXT,
    paused_by_host     BOOLEAN NOT NULL DEFAULT FALSE,
    director_mode      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at         DATETIME NOT NULL,
    updated_at         DATETIME NOT NULL
);

-- A participant row outlives its connection: left_at marks a disconnect, and a
-- reconnect by the same identity clears it, keeping the roles they held.
CREATE TABLE IF NOT EXISTS stage_participants (
    id                  TEXT PRIMARY KEY,
    stage_id            TEXT NOT NULL REFERENCES stages (id) ON DELETE CASCADE,
    user_id             TEXT REFERENCES identity_users (id) ON DELETE SET NULL,
    guest_id            TEXT,
    display_name        TEXT NOT NULL,
    is_ready            BOOLEAN NOT NULL DEFAULT FALSE,
    is_audience         BOOLEAN NOT NULL DEFAULT FALSE,
    force_muted_by_host BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          DATETIME NOT NULL,
    left_at             DATETIME
);

CREATE INDEX IF NOT EXISTS stage_participants_stage_id_idx ON stage_participants (stage_id);

-- cast_member_id names a role in the stage's selected play. It is not a
-- foreign key: an author's edit replaces the cast, and a live stage keeps the
-- roles it handed out until the play is selected again.
CREATE TABLE IF NOT EXISTS stage_participant_roles (
    stage_participant_id TEXT NOT NULL REFERENCES stage_participants (id) ON DELETE CASCADE,
    cast_member_id       TEXT NOT NULL,
    PRIMARY KEY (stage_participant_id, cast_member_id)
);

-- Bans are per stage, by identity, and by the address a banned participant was
-- connected from, so clearing a guest cookie does not get them back in.
CREATE TABLE IF NOT EXISTS stage_bans (
    id        TEXT PRIMARY KEY,
    stage_id  TEXT NOT NULL REFERENCES stages (id) ON DELETE CASCADE,
    user_id   TEXT,
    guest_id  TEXT,
    banned_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS stage_bans_stage_id_idx ON stage_bans (stage_id);

CREATE TABLE IF NOT EXISTS stage_ip_bans (
    id        TEXT PRIMARY KEY,
    stage_id  TEXT NOT NULL REFERENCES stages (id) ON DELETE CASCADE,
    ip        TEXT NOT NULL,
    banned_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS stage_ip_bans_stage_id_idx ON stage_ip_bans (stage_id);
