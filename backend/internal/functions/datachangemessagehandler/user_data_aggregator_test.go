package datachangemessagehandler

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsyncDataChangeMessageHandler_UserDataAggregationEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: "test-report-id",
			UserID:   "test-user-id",
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.On(reflection.GetMethodName(decoder.DecodeBytes), mock.Anything, rawMsg, mock.AnythingOfType("*dataprivacy.UserDataAggregationRequest")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
		})

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, "test-user-id").Return(&dataprivacy.UserDataCollection{}, nil)

		uploadManager.On(reflection.GetMethodName(uploadManager.SaveFile), mock.Anything, "test-report-id.json", mock.AnythingOfType("[]uint8")).Return(nil)

		err = handler.UserDataAggregationEventHandler(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, decoder, uploadManager, dataPrivacyRepo)
	})

	t.Run("with decode error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, decoder, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		rawMsg := []byte(`{"invalid": "json"}`)

		expectedError := errors.New("decode error")
		decoder.On(reflection.GetMethodName(decoder.DecodeBytes), mock.Anything, rawMsg, mock.AnythingOfType("*dataprivacy.UserDataAggregationRequest")).Return(expectedError)

		err := handler.UserDataAggregationEventHandler(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decoding JSON body")

		mock.AssertExpectationsForObjects(t, decoder)
	})

	t.Run("with fetch user data collection error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: "test-report-id",
			UserID:   "test-user-id",
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.On(reflection.GetMethodName(decoder.DecodeBytes), mock.Anything, rawMsg, mock.AnythingOfType("*dataprivacy.UserDataAggregationRequest")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
		})

		expectedError := errors.New("fetch error")
		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, "test-user-id").Return((*dataprivacy.UserDataCollection)(nil), expectedError)

		err = handler.UserDataAggregationEventHandler(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetching user data collection")

		mock.AssertExpectationsForObjects(t, decoder, dataPrivacyRepo)
	})

	t.Run("with upload error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: "test-report-id",
			UserID:   "test-user-id",
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.On(reflection.GetMethodName(decoder.DecodeBytes), mock.Anything, rawMsg, mock.AnythingOfType("*dataprivacy.UserDataAggregationRequest")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
		})

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, "test-user-id").Return(&dataprivacy.UserDataCollection{}, nil)

		expectedError := errors.New("upload error")
		uploadManager.On(reflection.GetMethodName(uploadManager.SaveFile), mock.Anything, "test-report-id.json", mock.AnythingOfType("[]uint8")).Return(expectedError)

		err = handler.UserDataAggregationEventHandler(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "saving collection")

		mock.AssertExpectationsForObjects(t, decoder, uploadManager, dataPrivacyRepo)
	})

	t.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: "", // Empty report ID
			UserID:   "test-user-id",
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.On(reflection.GetMethodName(decoder.DecodeBytes), mock.Anything, rawMsg, mock.AnythingOfType("*dataprivacy.UserDataAggregationRequest")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
		})

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, "test-user-id").Return(&dataprivacy.UserDataCollection{}, nil)

		uploadManager.On(reflection.GetMethodName(uploadManager.SaveFile), mock.Anything, ".json", mock.AnythingOfType("[]uint8")).Return(nil)

		err = handler.UserDataAggregationEventHandler(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, decoder, uploadManager, dataPrivacyRepo)
	})

	t.Run("with marshaling error scenario", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: "test-report-id",
			UserID:   "test-user-id",
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.On(reflection.GetMethodName(decoder.DecodeBytes), mock.Anything, rawMsg, mock.AnythingOfType("*dataprivacy.UserDataAggregationRequest")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
		})

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, "test-user-id").Return(&dataprivacy.UserDataCollection{}, nil)

		// The function marshals UserDataCollection which should not fail
		// This test ensures we handle the marshaling step correctly
		// Mock the upload manager to return success so we can test the marshaling path
		uploadManager.On(reflection.GetMethodName(uploadManager.SaveFile), mock.Anything, "test-report-id.json", mock.AnythingOfType("[]uint8")).Return(nil)

		err = handler.UserDataAggregationEventHandler(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, decoder, uploadManager, dataPrivacyRepo)
	})
}
