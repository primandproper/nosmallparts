package pubsub

import (
	"context"
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func buildPubSubBackedPublisher(t *testing.T, ctx context.Context) (publisher messagequeue.Publisher, shutdownFunc func(context.Context) error) {
	t.Helper()

	randomID, err := random.GenerateHexEncodedString(ctx, 8)
	require.NoError(t, err)
	projectID := "project-" + randomID
	topicID := "topic-" + randomID

	pubsubContainer, err := gcloud.RunPubsub(
		ctx,
		"google/cloud-sdk:latest",
		gcloud.WithProjectID(projectID),
	)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(pubsubContainer.URI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	require.NotNil(t, conn)

	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	require.NoError(t, err)

	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	pubSubTopic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
	require.NoError(t, err)
	require.NotNil(t, pubSubTopic)

	logger := logging.NewNoopLogger()
	provider := ProvidePubSubPublisherProvider(logger, tracing.NewNoopTracerProvider(), client)
	require.NotNil(t, provider)

	publisher, err = provider.ProvidePublisher(ctx, pubSubTopic.GetName())
	assert.NotNil(t, publisher)
	assert.NoError(t, err)

	return publisher, pubsubContainer.Terminate
}

func Test_pubSubPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		publisher, shutdownFunc := buildPubSubBackedPublisher(t, ctx)
		defer func() {
			require.NoError(t, shutdownFunc(ctx))
		}()

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		assert.NoError(t, publisher.Publish(ctx, inputData))
	})
}
