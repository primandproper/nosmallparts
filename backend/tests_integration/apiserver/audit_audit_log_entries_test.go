package integration

import (
	"testing"

	auditgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLogEntries_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be able to be read in a list", func(t *testing.T) {
		t.Parallel()
		t.SkipNow()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		forUser, err := userClient.GetAuditLogEntriesForUser(ctx, &auditgrpc.GetAuditLogEntriesForUserRequest{})
		require.NoError(t, err)

		assert.Equal(t, 4, len(forUser.Results))

		forAccount, err := userClient.GetAuditLogEntriesForAccount(ctx, &auditgrpc.GetAuditLogEntriesForAccountRequest{})
		require.NoError(t, err)

		assert.Equal(t, 2, len(forAccount.Results))
	})
}

/*

func (s *TestSuite) TestWebhooks_Retrieving_Returns404ForNonexistentAuditLogEntry() {
	s.runTest("should error when archiving a non-existent auditLogEntry", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.userClient.GetAuditLogEntryByID(ctx, nonexistentID)
			assert.Nil(t, actual)
			assert.Error(t, err)
		}
	})
}

*/
