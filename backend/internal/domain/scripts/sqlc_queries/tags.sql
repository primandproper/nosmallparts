-- name: ListScriptTags :many
SELECT * FROM script_tags ORDER BY name;

-- name: GetScriptTagByName :one
-- name is collated NOCASE, so this matches without regard to case.
SELECT * FROM script_tags WHERE name = ?;

-- name: InsertScriptTag :exec
INSERT INTO script_tags (id, name, description, created_at) VALUES (?, ?, ?, ?);

-- name: ListPlayScriptTags :many
SELECT t.*
FROM script_tags t
JOIN play_script_tags pt ON pt.script_tag_id = t.id
WHERE pt.play_script_id = ?
ORDER BY t.name;

-- name: InsertPlayScriptTag :exec
INSERT INTO play_script_tags (play_script_id, script_tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING;

-- name: DeletePlayScriptTags :exec
DELETE FROM play_script_tags WHERE play_script_id = ?;
