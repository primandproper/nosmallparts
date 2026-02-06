-- name: CreateStage :exec
INSERT INTO stages (
	id,
	code
) VALUES (
	sqlc.arg(id),
	sqlc.arg(code)
);

-- name: GetStageByID :one
SELECT
	stages.id,
	stages.code,
	stages.play_script_id,
	stages.current_line_index,
	stages.phase,
	stages.created_at,
	stages.last_updated_at,
	stages.archived_at
FROM stages
WHERE stages.archived_at IS NULL
	AND stages.id = sqlc.arg(id);

-- name: GetStageByCode :one
SELECT
	stages.id,
	stages.code,
	stages.play_script_id,
	stages.current_line_index,
	stages.phase,
	stages.created_at,
	stages.last_updated_at,
	stages.archived_at
FROM stages
WHERE stages.archived_at IS NULL
	AND stages.code = sqlc.arg(code);

-- name: UpdateStage :execrows
UPDATE stages SET
	play_script_id = COALESCE(sqlc.narg(play_script_id), play_script_id),
	current_line_index = COALESCE(sqlc.narg(current_line_index), current_line_index),
	phase = COALESCE(sqlc.narg(phase), phase),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveStage :execrows
UPDATE stages SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: CreateParticipant :exec
INSERT INTO stage_participants (
	id,
	stage_id,
	user_id,
	display_name
) VALUES (
	sqlc.arg(id),
	sqlc.arg(stage_id),
	sqlc.narg(user_id),
	sqlc.arg(display_name)
);

-- name: GetParticipant :one
SELECT
	stage_participants.id,
	stage_participants.stage_id,
	stage_participants.user_id,
	stage_participants.display_name,
	stage_participants.is_director,
	stage_participants.is_ready,
	stage_participants.is_audience,
	stage_participants.created_at,
	stage_participants.left_at
FROM stage_participants
WHERE stage_participants.id = sqlc.arg(id);

-- name: GetActiveParticipantsForStage :many
SELECT
	stage_participants.id,
	stage_participants.stage_id,
	stage_participants.user_id,
	stage_participants.display_name,
	stage_participants.is_director,
	stage_participants.is_ready,
	stage_participants.is_audience,
	stage_participants.created_at,
	stage_participants.left_at
FROM stage_participants
WHERE stage_participants.stage_id = sqlc.arg(stage_id)
	AND stage_participants.left_at IS NULL
ORDER BY stage_participants.created_at ASC;

-- name: UpdateParticipant :execrows
UPDATE stage_participants SET
	is_director = COALESCE(sqlc.narg(is_director), is_director),
	is_ready = COALESCE(sqlc.narg(is_ready), is_ready),
	is_audience = COALESCE(sqlc.narg(is_audience), is_audience)
WHERE id = sqlc.arg(id);

-- name: SetParticipantLeft :execrows
UPDATE stage_participants SET
	left_at = NOW()
WHERE id = sqlc.arg(id)
	AND left_at IS NULL;

-- name: AddParticipantRole :exec
INSERT INTO stage_participant_roles (
	stage_participant_id,
	cast_member_id
) VALUES (
	sqlc.arg(stage_participant_id),
	sqlc.arg(cast_member_id)
) ON CONFLICT DO NOTHING;

-- name: RemoveParticipantRole :execrows
DELETE FROM stage_participant_roles
WHERE stage_participant_id = sqlc.arg(stage_participant_id)
	AND cast_member_id = sqlc.arg(cast_member_id);

-- name: GetParticipantRoles :many
SELECT
	stage_participant_roles.cast_member_id
FROM stage_participant_roles
WHERE stage_participant_roles.stage_participant_id = sqlc.arg(stage_participant_id);
