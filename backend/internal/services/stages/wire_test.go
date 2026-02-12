package stages

import (
	"context"
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

	tracerProvider := tracing.NewNoopTracerProvider()
	logger := logging.NewNoopLogger()
	mgr := ProvideConnectionManager(tracerProvider, logger)
	require.NotNil(t, mgr)
	assert.False(t, mgr.GroupHasConnections(context.Background(), "any"))
	assert.Equal(t, 0, mgr.GetConnectionCount(context.Background(), "any"))
}

func TestProvideHandler(t *testing.T) {
	t.Parallel()

	logger := logging.NewNoopLogger()
	tracerProvider := tracing.NewNoopTracerProvider()
	stagesManager := NewManager(&fakeStagesRepo{})
	playScriptsRepo := &playscriptsmock.Repository{}
	connectionManager := platformws.NewConnectionManager(tracerProvider, logger)

	h := ProvideHandler(logger, tracerProvider, stagesManager, playScriptsRepo, connectionManager)
	require.NotNil(t, h)
	assert.NotNil(t, h.stagesManager)
	assert.NotNil(t, h.playScriptsRepo)
	assert.Equal(t, connectionManager, h.connections)
}
