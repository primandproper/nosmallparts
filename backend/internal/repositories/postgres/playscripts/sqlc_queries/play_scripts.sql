-- name: CreatePlayScript :exec
INSERT INTO play_scripts (
	id,
	title,
	description,
	author,
	is_public,
	preceded_by_id,
	precedes_id,
	created_by_user,
	belongs_to_account
) VALUES (
	sqlc.arg(id),
	sqlc.arg(title),
	sqlc.arg(description),
	sqlc.arg(author),
	sqlc.arg(is_public),
	sqlc.arg(preceded_by_id),
	sqlc.arg(precedes_id),
	sqlc.arg(created_by_user),
	sqlc.arg(belongs_to_account)
);

-- name: UpdatePlayScript :execrows
UPDATE play_scripts SET
	title = sqlc.arg(title),
	description = sqlc.arg(description),
	author = sqlc.arg(author),
	is_public = sqlc.arg(is_public),
	preceded_by_id = sqlc.arg(preceded_by_id),
	precedes_id = sqlc.arg(precedes_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchivePlayScript :execrows
UPDATE play_scripts SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetPlayScript :one
SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.preceded_by_id,
	play_scripts.precedes_id,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account
FROM play_scripts
WHERE play_scripts.archived_at IS NULL
	AND play_scripts.id = sqlc.arg(id);

-- name: CheckPlayScriptExistence :one
SELECT EXISTS(
	SELECT play_scripts.id
	FROM play_scripts
	WHERE play_scripts.archived_at IS NULL
		AND play_scripts.id = sqlc.arg(id)
);

-- name: GetPlayScripts :many
SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.preceded_by_id,
	play_scripts.precedes_id,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND
			play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_before), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_after), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR play_scripts.archived_at = NULL)
	) AS filtered_count,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
	) AS total_count
FROM play_scripts
WHERE play_scripts.archived_at IS NULL
	AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND play_scripts.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY play_scripts.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetPlayScriptsForAccount :many
SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.preceded_by_id,
	play_scripts.precedes_id,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND
			play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_before), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_after), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR play_scripts.archived_at = NULL)
			AND play_scripts.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS filtered_count,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND play_scripts.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM play_scripts
WHERE play_scripts.archived_at IS NULL
	AND play_scripts.belongs_to_account = sqlc.arg(belongs_to_account)
	AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND play_scripts.belongs_to_account = sqlc.arg(belongs_to_account)
	AND play_scripts.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY play_scripts.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetPublicPlayScripts :many
SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.preceded_by_id,
	play_scripts.precedes_id,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND
			play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_before), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_after), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR play_scripts.archived_at = NULL)
			AND play_scripts.is_public = TRUE
	) AS filtered_count,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND play_scripts.is_public = TRUE
	) AS total_count
FROM play_scripts
WHERE play_scripts.archived_at IS NULL
	AND play_scripts.is_public = TRUE
	AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND play_scripts.is_public = TRUE
	AND play_scripts.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY play_scripts.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchPlayScriptsForXParticipants :many
SELECT
	play_scripts.id,
	play_scripts.title,
	play_scripts.description,
	play_scripts.author,
	play_scripts.is_public,
	play_scripts.preceded_by_id,
	play_scripts.precedes_id,
	play_scripts.created_at,
	play_scripts.last_updated_at,
	play_scripts.archived_at,
	play_scripts.created_by_user,
	play_scripts.belongs_to_account,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND
			play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_before), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				play_scripts.last_updated_at IS NULL
				OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_after), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR play_scripts.archived_at = NULL)
			AND (play_scripts.is_public = TRUE OR play_scripts.belongs_to_account = sqlc.arg(belongs_to_account))
			AND (SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)
	) AS filtered_count,
	(
		SELECT COUNT(play_scripts.id)
		FROM play_scripts
		WHERE play_scripts.archived_at IS NULL
			AND (play_scripts.is_public = TRUE OR play_scripts.belongs_to_account = sqlc.arg(belongs_to_account))
			AND (SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)
	) AS total_count
FROM play_scripts
WHERE play_scripts.archived_at IS NULL
	AND (SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)
	AND (play_scripts.is_public = TRUE OR play_scripts.belongs_to_account = sqlc.arg(belongs_to_account))
	AND play_scripts.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND play_scripts.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		play_scripts.last_updated_at IS NULL
		OR play_scripts.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (play_scripts.is_public = TRUE OR play_scripts.belongs_to_account = sqlc.arg(belongs_to_account))
	AND (SELECT COUNT(*) FROM play_script_cast_members WHERE play_script_cast_members.play_script_id = play_scripts.id) = sqlc.arg(participant_count)
	AND play_scripts.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY play_scripts.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
