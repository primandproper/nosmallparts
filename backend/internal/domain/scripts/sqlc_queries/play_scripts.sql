-- name: GetPlayScript :one
SELECT * FROM play_scripts WHERE id = ?;

-- name: ListVisiblePlayScriptIDs :many
-- A viewer sees public scripts and their own, and nobody but an admin sees a
-- disabled one. An anonymous viewer passes the empty string, which no user has.
SELECT id FROM play_scripts
WHERE enabled AND (is_public OR created_by_user = sqlc.arg(viewer_id))
ORDER BY updated_at DESC, id DESC;

-- name: ListAllPlayScriptIDs :many
SELECT id FROM play_scripts ORDER BY updated_at DESC, id DESC;

-- name: InsertPlayScript :exec
INSERT INTO play_scripts (id, title, description, author, is_public, enabled, created_by_user, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateOwnedPlayScript :execrows
UPDATE play_scripts
SET title = sqlc.arg(title), description = sqlc.arg(description), author = sqlc.arg(author),
    is_public = sqlc.arg(is_public), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND created_by_user = sqlc.arg(created_by_user);

-- name: DeletePlayScript :execrows
DELETE FROM play_scripts WHERE id = ?;

-- name: SetPlayScriptEnabled :execrows
UPDATE play_scripts SET enabled = sqlc.arg(enabled), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: ListCastMembers :many
SELECT * FROM play_script_cast_members WHERE play_script_id = ? ORDER BY position;

-- name: InsertCastMember :exec
INSERT INTO play_script_cast_members (id, play_script_id, name, description, position) VALUES (?, ?, ?, ?, ?);

-- name: DeleteCastMembers :exec
DELETE FROM play_script_cast_members WHERE play_script_id = ?;

-- name: ListLines :many
SELECT * FROM play_script_lines WHERE play_script_id = ? ORDER BY line_index, position;

-- name: InsertLine :exec
INSERT INTO play_script_lines (id, play_script_id, line_text, line_index, position, direction, scroll_mode, auto_scroll_wpm)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteLines :exec
DELETE FROM play_script_lines WHERE play_script_id = ?;

-- name: ListLineSpeakers :many
SELECT s.line_id, s.cast_member_id
FROM play_script_line_speakers s
JOIN play_script_lines l ON l.id = s.line_id
WHERE l.play_script_id = ?
ORDER BY s.line_id, s.position;

-- name: InsertLineSpeaker :exec
INSERT INTO play_script_line_speakers (line_id, cast_member_id, position) VALUES (?, ?, ?);
