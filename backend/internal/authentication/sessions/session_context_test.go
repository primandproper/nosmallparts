package sessions

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchContextFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), SessionContextDataKey, &ContextData{})
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)
		require.NotNil(t, req)

		actual, err := FetchContextDataFromRequest(req)
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	T.Run("missing data", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)
		require.NotNil(t, req)

		actual, err := FetchContextDataFromRequest(req)
		require.Error(t, err)
		require.Nil(t, actual)
	})
}
