-- name: CreatePlayScriptRating :exec
INSERT INTO play_script_ratings (
	id,
	play_script_id,
	belongs_to_user,
	rating
) VALUES (
	sqlc.arg(id),
	sqlc.arg(play_script_id),
	sqlc.arg(belongs_to_user),
	sqlc.arg(rating)
);

-- name: ArchiveRatingForUserAndPlay :execrows
UPDATE play_script_ratings SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE play_script_ratings.archived_at IS NULL
	AND play_script_ratings.play_script_id = sqlc.arg(play_script_id)
	AND play_script_ratings.belongs_to_user = sqlc.arg(belongs_to_user);

-- name: GetActiveRatingForUserAndPlay :one
SELECT
	play_script_ratings.id,
	play_script_ratings.play_script_id,
	play_script_ratings.belongs_to_user,
	play_script_ratings.rating,
	play_script_ratings.created_at,
	play_script_ratings.last_updated_at,
	play_script_ratings.archived_at
FROM play_script_ratings
WHERE play_script_ratings.archived_at IS NULL
	AND play_script_ratings.play_script_id = sqlc.arg(play_script_id)
	AND play_script_ratings.belongs_to_user = sqlc.arg(belongs_to_user);

-- name: GetRatingStatsForPlayScript :one
SELECT
	COALESCE(AVG(play_script_ratings.rating)::double precision, 0) AS average_rating,
	COALESCE(COUNT(play_script_ratings.id), 0)::bigint AS rating_count
FROM play_script_ratings
WHERE play_script_ratings.play_script_id = sqlc.arg(play_script_id)
	AND play_script_ratings.archived_at IS NULL;

-- name: GetRatingStatsForPlayScripts :many
SELECT
	play_script_ratings.play_script_id,
	COALESCE(AVG(play_script_ratings.rating)::double precision, 0) AS average_rating,
	COALESCE(COUNT(play_script_ratings.id), 0)::bigint AS rating_count
FROM play_script_ratings
WHERE play_script_ratings.play_script_id = ANY(sqlc.arg(play_script_ids)::text[])
	AND play_script_ratings.archived_at IS NULL
GROUP BY play_script_ratings.play_script_id;

-- name: GetUserRatingsForPlayScripts :many
SELECT
	play_script_ratings.play_script_id,
	play_script_ratings.rating
FROM play_script_ratings
WHERE play_script_ratings.belongs_to_user = sqlc.arg(belongs_to_user)
	AND play_script_ratings.play_script_id = ANY(sqlc.arg(play_script_ids)::text[])
	AND play_script_ratings.archived_at IS NULL;
