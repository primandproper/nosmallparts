-- name: CreatePlayScriptLine :exec
INSERT INTO play_script_lines (
	id,
	play_script_id,
	cast_member_id,
	line_text,
	line_index,
	direction,
	scroll_mode,
	auto_scroll_wpm
) VALUES (
	sqlc.arg(id),
	sqlc.arg(play_script_id),
	sqlc.arg(cast_member_id),
	sqlc.arg(line_text),
	sqlc.arg(line_index),
	sqlc.arg(direction),
	sqlc.arg(scroll_mode),
	sqlc.arg(auto_scroll_wpm)
);

-- name: GetPlayScriptLine :one
SELECT
	play_script_lines.id,
	play_script_lines.play_script_id,
	play_script_lines.cast_member_id,
	play_script_lines.line_text,
	play_script_lines.line_index,
	play_script_lines.direction,
	play_script_lines.scroll_mode,
	play_script_lines.auto_scroll_wpm,
	play_script_lines.created_at
FROM play_script_lines
WHERE play_script_lines.id = sqlc.arg(id);

-- name: GetPlayScriptLinesForPlayScript :many
SELECT
	play_script_lines.id,
	play_script_lines.play_script_id,
	play_script_lines.cast_member_id,
	play_script_lines.line_text,
	play_script_lines.line_index,
	play_script_lines.direction,
	play_script_lines.scroll_mode,
	play_script_lines.auto_scroll_wpm,
	play_script_lines.created_at
FROM play_script_lines
WHERE play_script_lines.play_script_id = sqlc.arg(play_script_id)
ORDER BY play_script_lines.line_index ASC;

-- name: UpdatePlayScriptLine :execrows
UPDATE play_script_lines SET
	cast_member_id = sqlc.arg(cast_member_id),
	line_text = sqlc.arg(line_text),
	line_index = sqlc.arg(line_index),
	direction = sqlc.arg(direction),
	scroll_mode = sqlc.arg(scroll_mode),
	auto_scroll_wpm = sqlc.arg(auto_scroll_wpm)
WHERE id = sqlc.arg(id);

-- name: DeletePlayScriptLine :execrows
DELETE FROM play_script_lines
WHERE id = sqlc.arg(id);

-- name: DeletePlayScriptLinesForPlayScript :execrows
DELETE FROM play_script_lines
WHERE play_script_id = sqlc.arg(play_script_id);
