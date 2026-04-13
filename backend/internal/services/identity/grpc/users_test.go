package grpc

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/reflection"
	"github.com/verygoodsoftwarenotvirus/platform/v4/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceImpl_CreateUser(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInput := identityfakes.BuildFakeUserCreationInput()
		exampleResponse := &identity.UserCreationResponse{
			CreatedUserID:   identityfakes.BuildFakeID(),
			Username:        exampleInput.Username,
			EmailAddress:    exampleInput.EmailAddress,
			FirstName:       exampleInput.FirstName,
			LastName:        exampleInput.LastName,
			TwoFactorSecret: "secret",
			TwoFactorQRCode: "qr_code",
			AccountStatus:   identity.UnverifiedAccountStatus.String(),
			CreatedAt:       identityfakes.BuildFakeTime(),
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateUser), testutils.ContextMatcher, mock.MatchedBy(func(input *identity.UserRegistrationInput) bool {
			return input.Username == exampleInput.Username &&
				input.EmailAddress == exampleInput.EmailAddress &&
				input.FirstName == exampleInput.FirstName &&
				input.LastName == exampleInput.LastName
		})).Return(exampleResponse, nil)

		request := &identitysvc.CreateUserRequest{
			Input: &identitysvc.UserRegistrationInput{
				Username:              exampleInput.Username,
				EmailAddress:          exampleInput.EmailAddress,
				FirstName:             exampleInput.FirstName,
				LastName:              exampleInput.LastName,
				Password:              exampleInput.Password,
				AccountName:           exampleInput.AccountName,
				AcceptedTos:           exampleInput.AcceptedTOS,
				AcceptedPrivacyPolicy: exampleInput.AcceptedPrivacyPolicy,
			},
		}

		result, err := service.CreateUser(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Created)
		assert.Equal(t, exampleResponse.CreatedUserID, result.Created.CreatedUserId)
		assert.Equal(t, exampleResponse.Username, result.Created.Username)
		assert.Equal(t, exampleResponse.EmailAddress, result.Created.EmailAddress)
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInput := identityfakes.BuildFakeUserCreationInput()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateUser), testutils.ContextMatcher, mock.AnythingOfType("*identity.UserRegistrationInput")).Return((*identity.UserCreationResponse)(nil), errors.New("database error"))

		request := &identitysvc.CreateUserRequest{
			Input: &identitysvc.UserRegistrationInput{
				Username:              exampleInput.Username,
				EmailAddress:          exampleInput.EmailAddress,
				FirstName:             exampleInput.FirstName,
				LastName:              exampleInput.LastName,
				Password:              exampleInput.Password,
				AccountName:           exampleInput.AccountName,
				AcceptedTos:           exampleInput.AcceptedTOS,
				AcceptedPrivacyPolicy: exampleInput.AcceptedPrivacyPolicy,
			},
		}

		result, err := service.CreateUser(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_ArchiveUser(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveUser), testutils.ContextMatcher, exampleUserID).Return(nil)

		request := &identitysvc.ArchiveUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.ArchiveUser(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveUser), testutils.ContextMatcher, exampleUserID).Return(errors.New("database error"))

		request := &identitysvc.ArchiveUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.ArchiveUser(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetUser(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUser := identityfakes.BuildFakeUser()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUser), testutils.ContextMatcher, exampleUser.ID).Return(exampleUser, nil)

		request := &identitysvc.GetUserRequest{
			UserId: exampleUser.ID,
		}

		result, err := service.GetUser(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Result)
		assert.Equal(t, exampleUser.ID, result.Result.Id)
		assert.Equal(t, exampleUser.Username, result.Result.Username)
		assert.Equal(t, exampleUser.EmailAddress, result.Result.EmailAddress)
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUser), testutils.ContextMatcher, exampleUserID).Return((*identity.User)(nil), errors.New("database error"))

		request := &identitysvc.GetUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.GetUser(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetUsers(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
				identityfakes.BuildFakeUser(),
			},
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUsers), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleUsers, nil)

		pageSize := uint32(25)
		request := &identitysvc.GetUsersRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetUsers(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleUsers.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUsers), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.User])(nil), errors.New("database error"))

		pageSize := uint32(25)
		request := &identitysvc.GetUsersRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetUsers(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_SearchForUsers(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
				identityfakes.BuildFakeUser(),
			},
		}
		exampleQuery := "test search"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SearchForUsers), testutils.ContextMatcher, exampleQuery, false, testutils.QueryFilterMatcher).Return(exampleUsers, nil)

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: false,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleUsers.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	t.Run("with search service enabled", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
			},
		}
		exampleQuery := "search query"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SearchForUsers), testutils.ContextMatcher, exampleQuery, true, testutils.QueryFilterMatcher).Return(exampleUsers, nil)

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: true,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleUsers.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleQuery := "test search"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SearchForUsers), testutils.ContextMatcher, exampleQuery, false, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.User])(nil), errors.New("search error"))

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: false,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateUserDetails(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserDetails), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.UserDetailsUpdateRequestInput")).Return(nil)

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName:       "Updated First",
				LastName:        "Updated Last",
				CurrentPassword: "password",
			},
		}

		result, err := service.UpdateUserDetails(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName: "Updated First",
			},
		}

		result, err := service.UpdateUserDetails(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserDetails), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.UserDetailsUpdateRequestInput")).Return(errors.New("update error"))

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName: "Updated First",
			},
		}

		result, err := service.UpdateUserDetails(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateUserEmailAddress(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		newEmail := "new@example.com"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserEmailAddress), testutils.ContextMatcher, mock.AnythingOfType("string"), newEmail).Return(nil)

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: newEmail,
		}

		result, err := service.UpdateUserEmailAddress(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: "new@example.com",
		}

		result, err := service.UpdateUserEmailAddress(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserEmailAddress), testutils.ContextMatcher, mock.AnythingOfType("string"), "new@example.com").Return(errors.New("update error"))

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: "new@example.com",
		}

		result, err := service.UpdateUserEmailAddress(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateUserUsername(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		newUsername := "newusername"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserUsername), testutils.ContextMatcher, mock.AnythingOfType("string"), newUsername).Return(nil)

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: newUsername,
		}

		result, err := service.UpdateUserUsername(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: "newusername",
		}

		result, err := service.UpdateUserUsername(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserUsername), testutils.ContextMatcher, mock.AnythingOfType("string"), "newusername").Return(errors.New("update error"))

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: "newusername",
		}

		result, err := service.UpdateUserUsername(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UploadUserAvatar(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		base64Data := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UploadUserAvatar), testutils.ContextMatcher, mock.AnythingOfType("string"), base64Data).Return(nil)

		request := &identitysvc.UploadUserAvatarRequest{
			Base64EncodedData: base64Data,
		}

		result, err := service.UploadUserAvatar(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UploadUserAvatarRequest{
			Base64EncodedData: "test-data",
		}

		result, err := service.UploadUserAvatar(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UploadUserAvatar), testutils.ContextMatcher, mock.AnythingOfType("string"), "test-data").Return(errors.New("upload error"))

		request := &identitysvc.UploadUserAvatarRequest{
			Base64EncodedData: "test-data",
		}

		result, err := service.UploadUserAvatar(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}
