package integration

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc/converters"

	"github.com/stretchr/testify/assert"
)

func TestUsers_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createUserAndClientForTest(t)
	})

	T.Run("rejects duplicate registration", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: identityconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input)})
		assert.NoError(t, err)

		_, err = testClient.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: identityconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input)})
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		input.Username = ""

		testClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := testClient.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: identityconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input)})
		assert.Error(t, err)
	})
}

func TestUsers_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		u, _ := createUserAndClientForTest(t)

		user, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: u.ID})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, u.ID, user.Result.Id)
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		u, _ := createUserAndClientForTest(t)

		user, err := c.GetUser(ctx, &identitysvc.GetUserRequest{UserId: u.ID})
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUsers_PermissionChecking(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		response, err := testClient.CheckPermissions(ctx, &authsvc.UserPermissionsRequestInput{Permissions: []string{
			authorization.ImpersonateUserPermission.ID(),
			authorization.ReadWebhooksPermission.ID(), // permission everyone has
		}})
		assert.NoError(t, err)
		assert.NotNil(t, response)

		assert.Equal(t, response.Permissions, map[string]bool{
			authorization.ImpersonateUserPermission.ID(): false,
			authorization.ReadWebhooksPermission.ID():    true,
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		testClient := buildUnauthenticatedGRPCClientForTest(t)

		response, err := testClient.CheckPermissions(ctx, &authsvc.UserPermissionsRequestInput{Permissions: []string{authorization.ReadWebhooksPermission.ID()}})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestUsers_Searching(T *testing.T) {
	T.Parallel()

	// create some users to search from
	createdUsers := []*identity.User{}
	for range exampleQuantity {
		u, _ := createUserAndClientForTest(T)
		createdUsers = append(createdUsers, u)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.SearchForUsers(ctx, &identitysvc.SearchForUsersRequest{
			Query: createdUsers[0].Username[:2],
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= 1)
	})

	T.Run("only admins can do it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.SearchForUsers(ctx, &identitysvc.SearchForUsersRequest{
			Query: createdUsers[0].Username[:2],
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestUsers_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		_, err := adminClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: nonexistentID,
		})
		assert.Error(t, err)
	})

	T.Run("only admins can archive another user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
	})
}
