package stages

import (
	"testing"

	playscriptsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	pusherpkg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pusher"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideHandler(t *testing.T) {
	t.Parallel()

	logger := logging.NewNoopLogger()
	tracerProvider := tracing.NewNoopTracerProvider()
	stagesRepo := &fakeStagesRepo{}
	playScriptsRepo := &playscriptsmock.Repository{}
	pusherClient := pusherpkg.NewClient(nil, logger, tracerProvider)

	h := ProvideHandler(logger, tracerProvider, stagesRepo, playScriptsRepo, pusherClient)
	require.NotNil(t, h)
	assert.NotNil(t, h.stagesRepo)
	assert.NotNil(t, h.playScriptsRepo)
	assert.NotNil(t, h.pusher)
}
