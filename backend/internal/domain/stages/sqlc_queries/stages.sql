-- name: InsertStage :exec
INSERT INTO stages (id, code, play_script_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?);

-- name: GetStageByID :one
SELECT * FROM stages WHERE id = ?;

-- name: GetStageByCode :one
SELECT * FROM stages WHERE code = ?;

-- name: StageCodeTaken :one
SELECT EXISTS (SELECT 1 FROM stages WHERE code = ?) AS taken;

-- name: DeleteStage :exec
DELETE FROM stages WHERE id = ?;

-- name: SetStagePlay :exec
UPDATE stages SET play_script_id = sqlc.arg(play_script_id), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: SetStagePhase :exec
UPDATE stages SET phase = sqlc.arg(phase), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: SetStagePhaseAndLine :exec
UPDATE stages
SET phase = sqlc.arg(phase), current_line_index = sqlc.arg(current_line_index), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: SetStageLine :exec
UPDATE stages SET current_line_index = sqlc.arg(current_line_index), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: SetStagePaused :exec
UPDATE stages SET paused_by_host = sqlc.arg(paused_by_host), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: SetStageDirectorMode :exec
UPDATE stages SET director_mode = sqlc.arg(director_mode), updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: SetStageHost :exec
UPDATE stages
SET host_user_id = sqlc.arg(host_user_id), host_guest_id = sqlc.arg(host_guest_id), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: ResetStage :exec
UPDATE stages SET phase = 'lobby', current_line_index = 0, updated_at = sqlc.arg(updated_at) WHERE id = sqlc.arg(id);

-- name: ListIdleStageIDs :many
-- A stage is idle when neither it nor any participant on it has done anything
-- since the cutoff: no update, no join and no leave. A stage nobody ever
-- joined has no participant timestamps, which COALESCE reads as long ago.
SELECT s.id FROM stages s
LEFT JOIN stage_participants p ON p.stage_id = s.id
WHERE s.updated_at < sqlc.arg(cutoff)
GROUP BY s.id
HAVING COALESCE(MAX(p.created_at), '') < sqlc.arg(cutoff)
   AND COALESCE(MAX(p.left_at), '') < sqlc.arg(cutoff);

-- name: InsertParticipant :exec
INSERT INTO stage_participants (id, stage_id, user_id, guest_id, display_name, is_audience, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetParticipant :one
SELECT * FROM stage_participants WHERE id = ?;

-- name: ListActiveParticipants :many
SELECT * FROM stage_participants WHERE stage_id = ? AND left_at IS NULL ORDER BY created_at, id;

-- name: FindLatestParticipantByUser :one
SELECT * FROM stage_participants WHERE stage_id = ? AND user_id = ? ORDER BY created_at DESC, id DESC LIMIT 1;

-- name: FindLatestParticipantByGuest :one
SELECT * FROM stage_participants WHERE stage_id = ? AND guest_id = ? ORDER BY created_at DESC, id DESC LIMIT 1;

-- name: SetParticipantLeftAt :exec
UPDATE stage_participants SET left_at = sqlc.arg(left_at) WHERE id = sqlc.arg(id);

-- name: SetParticipantDisplayName :exec
UPDATE stage_participants SET display_name = sqlc.arg(display_name) WHERE id = sqlc.arg(id);

-- name: SetParticipantReady :exec
UPDATE stage_participants SET is_ready = sqlc.arg(is_ready) WHERE id = sqlc.arg(id);

-- name: SetParticipantAudience :exec
UPDATE stage_participants SET is_audience = sqlc.arg(is_audience) WHERE id = sqlc.arg(id);

-- name: SetParticipantForceMuted :exec
UPDATE stage_participants SET force_muted_by_host = sqlc.arg(force_muted_by_host) WHERE id = sqlc.arg(id);

-- name: ClearStageReadiness :exec
UPDATE stage_participants SET is_ready = FALSE WHERE stage_id = ?;

-- name: ListStageRoles :many
SELECT r.stage_participant_id, r.cast_member_id
FROM stage_participant_roles r
JOIN stage_participants p ON p.id = r.stage_participant_id
WHERE p.stage_id = ?
ORDER BY r.stage_participant_id, r.cast_member_id;

-- name: ListParticipantRoles :many
SELECT cast_member_id FROM stage_participant_roles WHERE stage_participant_id = ? ORDER BY cast_member_id;

-- name: InsertParticipantRole :exec
INSERT INTO stage_participant_roles (stage_participant_id, cast_member_id) VALUES (?, ?) ON CONFLICT DO NOTHING;

-- name: DeleteParticipantRole :execrows
DELETE FROM stage_participant_roles WHERE stage_participant_id = ? AND cast_member_id = ?;

-- name: DeleteParticipantRoles :exec
DELETE FROM stage_participant_roles WHERE stage_participant_id = ?;

-- name: DeleteRoleOnStage :exec
-- Roles are exclusive on a stage: whoever held this one gives it up.
DELETE FROM stage_participant_roles
WHERE cast_member_id = sqlc.arg(cast_member_id)
  AND stage_participant_id IN (SELECT id FROM stage_participants WHERE stage_id = sqlc.arg(stage_id));

-- name: DeleteStageRoles :exec
DELETE FROM stage_participant_roles
WHERE stage_participant_id IN (SELECT id FROM stage_participants WHERE stage_id = ?);

-- name: InsertBan :exec
INSERT INTO stage_bans (id, stage_id, user_id, guest_id, banned_at) VALUES (?, ?, ?, ?, ?);

-- name: CountUserBans :one
SELECT COUNT(*) FROM stage_bans WHERE stage_id = ? AND user_id = ?;

-- name: CountGuestBans :one
SELECT COUNT(*) FROM stage_bans WHERE stage_id = ? AND guest_id = ?;

-- name: InsertIPBan :exec
INSERT INTO stage_ip_bans (id, stage_id, ip, banned_at) VALUES (?, ?, ?, ?);

-- name: CountIPBans :one
SELECT COUNT(*) FROM stage_ip_bans WHERE stage_id = ? AND ip = ?;
