-- name: CreatePlayScriptCastMember :exec
INSERT INTO play_script_cast_members (
	id,
	play_script_id,
	name,
	description,
	sort_order
) VALUES (
	sqlc.arg(id),
	sqlc.arg(play_script_id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(sort_order)
);

-- name: GetPlayScriptCastMember :one
SELECT
	play_script_cast_members.id,
	play_script_cast_members.play_script_id,
	play_script_cast_members.name,
	play_script_cast_members.description,
	play_script_cast_members.sort_order,
	play_script_cast_members.created_at
FROM play_script_cast_members
WHERE play_script_cast_members.id = sqlc.arg(id);

-- name: GetPlayScriptCastMembersForPlayScript :many
SELECT
	play_script_cast_members.id,
	play_script_cast_members.play_script_id,
	play_script_cast_members.name,
	play_script_cast_members.description,
	play_script_cast_members.sort_order,
	play_script_cast_members.created_at
FROM play_script_cast_members
WHERE play_script_cast_members.play_script_id = sqlc.arg(play_script_id)
ORDER BY play_script_cast_members.sort_order ASC;

-- name: UpdatePlayScriptCastMember :execrows
UPDATE play_script_cast_members SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	sort_order = sqlc.arg(sort_order)
WHERE id = sqlc.arg(id);

-- name: DeletePlayScriptCastMember :execrows
DELETE FROM play_script_cast_members
WHERE id = sqlc.arg(id);

-- name: DeletePlayScriptCastMembersForPlayScript :execrows
DELETE FROM play_script_cast_members
WHERE play_script_id = sqlc.arg(play_script_id);
