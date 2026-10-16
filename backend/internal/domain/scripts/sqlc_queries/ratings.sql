-- name: GetRatingStats :one
SELECT COUNT(*) AS rating_count, CAST(COALESCE(AVG(rating), 0) AS REAL) AS average_rating
FROM play_script_ratings
WHERE play_script_id = ?;

-- name: UpsertRating :exec
INSERT INTO play_script_ratings (play_script_id, user_id, rating, created_at, updated_at)
VALUES (sqlc.arg(play_script_id), sqlc.arg(user_id), sqlc.arg(rating), sqlc.arg(rated_at), sqlc.arg(rated_at))
ON CONFLICT (play_script_id, user_id) DO UPDATE SET rating = excluded.rating, updated_at = excluded.updated_at;

-- name: GetUserRating :one
SELECT rating FROM play_script_ratings WHERE play_script_id = ? AND user_id = ?;
