-- name: CreateScriptTag :exec
INSERT INTO script_tags (
	id,
	name,
	description
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description)
);

-- name: GetScriptTag :one
SELECT
	script_tags.id,
	script_tags.name,
	script_tags.description,
	script_tags.created_at
FROM script_tags
WHERE script_tags.id = sqlc.arg(id);

-- name: GetScriptTagByName :one
SELECT
	script_tags.id,
	script_tags.name,
	script_tags.description,
	script_tags.created_at
FROM script_tags
WHERE script_tags.name = sqlc.arg(name);

-- name: ListScriptTags :many
SELECT
	script_tags.id,
	script_tags.name,
	script_tags.description,
	script_tags.created_at
FROM script_tags
ORDER BY script_tags.name ASC;

-- name: AddTagToPlayScript :exec
INSERT INTO play_script_tags (
	play_script_id,
	script_tag_id,
	created_at
) VALUES (
	sqlc.arg(play_script_id),
	sqlc.arg(script_tag_id),
	NOW()
)
ON CONFLICT (play_script_id, script_tag_id) DO UPDATE SET
	archived_at = NULL;

-- name: RemoveTagFromPlayScript :execrows
UPDATE play_script_tags SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND play_script_id = sqlc.arg(play_script_id)
	AND script_tag_id = sqlc.arg(script_tag_id);

-- name: GetTagsForPlayScript :many
SELECT
	script_tags.id,
	script_tags.name,
	script_tags.description,
	script_tags.created_at
FROM script_tags
JOIN play_script_tags ON play_script_tags.script_tag_id = script_tags.id
	AND play_script_tags.archived_at IS NULL
WHERE play_script_tags.play_script_id = sqlc.arg(play_script_id);

-- name: GetPlayScriptsWithTag :many
SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		JOIN play_script_tags ON play_script_tags.play_script_id = play_scripts.id AND play_script_tags.archived_at IS NULL
		JOIN script_tags ON script_tags.id = play_script_tags.script_tag_id AND script_tags.name = sqlc.arg(name)
		WHERE play_scripts.archived_at IS NULL
			AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	) AS filtered_count,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		JOIN play_script_tags ON play_script_tags.play_script_id = play_scripts.id AND play_script_tags.archived_at IS NULL
		JOIN script_tags ON script_tags.id = play_script_tags.script_tag_id AND script_tags.name = sqlc.arg(name)
		WHERE play_scripts.archived_at IS NULL
	) AS total_count
FROM play_scripts
JOIN play_script_tags ON play_script_tags.play_script_id = play_scripts.id AND play_script_tags.archived_at IS NULL
JOIN script_tags ON script_tags.id = play_script_tags.script_tag_id AND script_tags.name = sqlc.arg(name)
WHERE play_scripts.archived_at IS NULL
	AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND play_scripts.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY play_scripts.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
