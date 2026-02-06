package stages

import (
	"testing"

	playscriptsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	platformws "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideConnectionManager(t *testing.T) {
	t.Parallel()

	mgr := ProvideConnectionManager()
	require.NotNil(t, mgr)
	assert.False(t, mgr.GroupHasConnections("any"))
	assert.Equal(t, 0, mgr.GetConnectionCount("any"))
}

func TestProvideHandler(t *testing.T) {
	t.Parallel()

	logger := logging.NewNoopLogger()
	tracerProvider := tracing.NewNoopTracerProvider()
	stagesRepo := &fakeStagesRepo{}
	playScriptsRepo := &playscriptsmock.Repository{}
	connectionManager := platformws.NewConnectionManager()

	h := ProvideHandler(logger, tracerProvider, stagesRepo, playScriptsRepo, connectionManager)
	require.NotNil(t, h)
	assert.NotNil(t, h.stagesRepo)
	assert.NotNil(t, h.playScriptsRepo)
	assert.Equal(t, connectionManager, h.connections)
}
