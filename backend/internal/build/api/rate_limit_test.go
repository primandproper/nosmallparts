package api_test

// The case in this file is ported from the TypeScript server's
// rate-limit.test.ts, under its original name. The lookup is an RPC now, so
// 429 is RESOURCE_EXHAUSTED and Retry-After is the google.rpc.RetryInfo the
// refusal carries.

import (
	"context"
	"testing"

	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestStageCodeLookupRateLimit(T *testing.T) {
	T.Parallel()

	T.Run("allows 30/min/IP then returns 429 with Retry-After; per-IP isolation", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		lookup := &stagepb.StageExistsRequest{Code: "ABCDEF"}

		from := func(ip string) context.Context {
			return metadata.AppendToOutgoingContext(t.Context(), "x-forwarded-for", ip)
		}

		for n := range 30 {
			_, err := h.stages.StageExists(from("192.0.2.55"), lookup)
			must.NoError(t, err, must.Sprintf("lookup %d", n+1))
		}

		_, err := h.stages.StageExists(from("192.0.2.55"), lookup)
		must.Error(t, err)

		refusal := status.Convert(err)
		test.EqOp(t, codes.ResourceExhausted, refusal.Code())

		var retry *errdetails.RetryInfo

		for _, detail := range refusal.Details() {
			if info, ok := detail.(*errdetails.RetryInfo); ok {
				retry = info
			}
		}

		must.NotNil(t, retry)
		test.Positive(t, retry.GetRetryDelay().AsDuration())

		// Another address is unaffected.
		_, err = h.stages.StageExists(from("192.0.2.56"), lookup)
		test.NoError(t, err)
	})
}
