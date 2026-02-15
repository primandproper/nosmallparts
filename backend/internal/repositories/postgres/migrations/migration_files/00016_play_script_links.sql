-- Add self-referential links for chaining acts (preceded_by -> current -> precedes)
ALTER TABLE play_scripts
    ADD COLUMN IF NOT EXISTS preceded_by_id TEXT REFERENCES play_scripts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS precedes_id TEXT REFERENCES play_scripts(id) ON DELETE SET NULL;
