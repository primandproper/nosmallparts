package api_test

// The cases in this file are ported from the TypeScript server's
// cleanup.test.ts, each subtest under its original name.

import (
	"testing"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"
	"github.com/primandproper/nosmallparts/backend/internal/services/stages/ws"

	"github.com/primandproper/primitives-go/v2/database"

	"github.com/samber/do/v2"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

// longAgo is far enough back that anything stamped with it is idle.
func longAgo() time.Time {
	return time.Now().Add(-1000 * time.Minute).UTC()
}

func (h *harness) db() database.Client {
	return do.MustInvoke[database.Client](h.injector)
}

func (h *harness) stageStore() stages.Store {
	return do.MustInvoke[stages.Store](h.injector)
}

func (h *harness) stageExists(t *testing.T, id string) bool {
	t.Helper()

	_, err := h.stageStore().GetStage(t.Context(), h.db().Reader(), id)
	if err == nil {
		return true
	}

	must.ErrorIs(t, err, stages.ErrStageNotFound)

	return false
}

// age stamps a stage, and optionally its participants, as long idle.
func (h *harness) age(t *testing.T, stageID string, participantsToo bool) {
	t.Helper()

	old := longAgo()

	_, err := h.db().Writer().ExecContext(t.Context(), "UPDATE stages SET created_at = ?, updated_at = ? WHERE id = ?", old, old, stageID)
	must.NoError(t, err)

	if participantsToo {
		_, err = h.db().Writer().ExecContext(t.Context(), "UPDATE stage_participants SET created_at = ? WHERE stage_id = ?", old, stageID)
		must.NoError(t, err)
	}
}

func TestStageCleanup(T *testing.T) {
	T.Parallel()

	T.Run("deletes the stage shortly after the last participant leaves", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		a := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		b := h.joinClient(t, guestIdentity(), stage.ID, "Guesty")
		a.disconnect()
		b.disconnect()

		eventually(t, "the stage to be deleted", func() bool { return !h.stageExists(t, stage.ID) })

		// Its participants went with it.
		var rows int
		must.NoError(t, h.db().Reader().QueryRowContext(t.Context(), "SELECT COUNT(*) FROM stage_participants WHERE stage_id = ?", stage.ID).Scan(&rows))
		test.EqOp(t, 0, rows)

		// As if it never existed.
		res, err := h.stages.StageExists(t.Context(), &stagepb.StageExistsRequest{Code: stage.Code})
		must.NoError(t, err)
		test.False(t, res.GetExists())
	})

	T.Run("a joiner during the 5s window cancels cleanup", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		a := h.joinClient(t, guestIdentity(), stage.ID, "FirstIn")
		a.disconnect()

		// Join again at once, inside the cleanup delay.
		b := h.joinClient(t, guestIdentity(), stage.ID, "SecondIn")
		time.Sleep(cleanupWait)

		test.True(t, h.stageExists(t, stage.ID))
		test.NotEq(t, "", b.participantID())
	})

	T.Run("reaper deletes stages idle past the threshold, recovered from DB", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		// A stage a restart orphaned: rows in the database, nobody connected,
		// every timestamp old.
		stage := h.makeStage(t, "")
		h.age(t, stage.ID, false)

		fresh := h.makeStage(t, "")

		deleted, err := do.MustInvoke[*ws.Hub](h.injector).ReapOnce(t.Context())
		must.NoError(t, err)

		test.SliceContains(t, deleted, stage.ID)
		test.False(t, h.stageExists(t, stage.ID))
		test.True(t, h.stageExists(t, fresh.ID))
	})

	T.Run("reaper spares stages with live connections", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		h.joinClient(t, guestIdentity(), stage.ID, "StillHere")
		h.age(t, stage.ID, true)

		deleted, err := do.MustInvoke[*ws.Hub](h.injector).ReapOnce(t.Context())
		must.NoError(t, err)

		test.SliceNotContains(t, deleted, stage.ID)
		test.True(t, h.stageExists(t, stage.ID))
	})

	T.Run("leftAt is set on disconnect and filtered from stage state", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		a := h.joinClient(t, guestIdentity(), stage.ID, "Leaver")
		b := h.joinClient(t, guestIdentity(), stage.ID, "Stayer")
		a.disconnect()

		b.waitFor("the leaver to go", func() bool { return participantIn(b.lastState(), a.participantID()) == nil })

		row, err := h.stageStore().GetParticipant(t.Context(), h.db().Reader(), a.participantID())
		must.NoError(t, err)
		test.NotNil(t, row.LeftAt)

		var names []string
		for _, p := range b.lastState().GetParticipants() {
			names = append(names, p.GetName())
		}

		test.SliceNotContains(t, names, "Leaver")
		test.SliceContains(t, names, "Stayer")
	})
}
