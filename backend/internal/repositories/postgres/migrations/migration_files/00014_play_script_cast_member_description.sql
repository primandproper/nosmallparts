-- Add description column to play_script_cast_members
ALTER TABLE play_script_cast_members ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
