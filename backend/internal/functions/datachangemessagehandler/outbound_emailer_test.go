package datachangemessagehandler

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsyncDataChangeMessageHandler_OutboundEmailsEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		rawMsg, err := json.Marshal(emailMessage)
		assert.NoError(t, err)

		emailer.On(reflection.GetMethodName(emailer.SendEmail), mock.Anything, emailMessage).Return(nil)
		analyticsEventReporter.On(reflection.GetMethodName(analyticsEventReporter.EventOccurred), mock.Anything, email.SentEventType, emailMessage.UserID, mock.MatchedBy(func(props map[string]any) bool {
			return props["toAddress"] == emailMessage.ToAddress &&
				props["toName"] == emailMessage.ToName &&
				props["fromAddress"] == emailMessage.FromAddress &&
				props["fromName"] == emailMessage.FromName &&
				props["subject"] == emailMessage.Subject
		})).Return(nil)

		err = handler.OutboundEmailsEventHandler(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, emailer, analyticsEventReporter)
	})

	t.Run("with invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		rawMsg := []byte("invalid json")

		err := handler.OutboundEmailsEventHandler(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decoding JSON body")
	})

	t.Run("with email sending error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		rawMsg, err := json.Marshal(emailMessage)
		assert.NoError(t, err)

		expectedError := errors.New("email sending error")
		emailer.On(reflection.GetMethodName(emailer.SendEmail), mock.Anything, emailMessage).Return(expectedError)
		analyticsEventReporter.On(reflection.GetMethodName(analyticsEventReporter.EventOccurred), mock.Anything, email.SentEventType, emailMessage.UserID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

		err = handler.OutboundEmailsEventHandler(ctx, rawMsg)
		assert.NoError(t, err) // Should not return error, just log it

		mock.AssertExpectationsForObjects(t, emailer, analyticsEventReporter)
	})

	t.Run("with analytics error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		rawMsg, err := json.Marshal(emailMessage)
		assert.NoError(t, err)

		emailer.On(reflection.GetMethodName(emailer.SendEmail), mock.Anything, emailMessage).Return(nil)
		expectedError := errors.New("analytics error")
		analyticsEventReporter.On(reflection.GetMethodName(analyticsEventReporter.EventOccurred), mock.Anything, email.SentEventType, emailMessage.UserID, mock.AnythingOfType("map[string]interface {}")).Return(expectedError)

		err = handler.OutboundEmailsEventHandler(ctx, rawMsg)
		assert.NoError(t, err) // Should not return error, just log it

		mock.AssertExpectationsForObjects(t, emailer, analyticsEventReporter)
	})
}

func TestAsyncDataChangeMessageHandler_handleEmailRequest(t *testing.T) {
	t.Parallel()

	t.Run("with nil email message", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		err := handler.handleEmailRequest(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, errRequiredDataIsNil, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		emailer.On(reflection.GetMethodName(emailer.SendEmail), mock.Anything, emailMessage).Return(nil)
		analyticsEventReporter.On(reflection.GetMethodName(analyticsEventReporter.EventOccurred), mock.Anything, email.SentEventType, emailMessage.UserID, mock.MatchedBy(func(props map[string]any) bool {
			return props["toAddress"] == emailMessage.ToAddress &&
				props["toName"] == emailMessage.ToName &&
				props["fromAddress"] == emailMessage.FromAddress &&
				props["fromName"] == emailMessage.FromName &&
				props["subject"] == emailMessage.Subject
		})).Return(nil)

		err := handler.handleEmailRequest(ctx, emailMessage)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, emailer, analyticsEventReporter)
	})

	t.Run("with email sending error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		expectedError := errors.New("email sending error")
		emailer.On(reflection.GetMethodName(emailer.SendEmail), mock.Anything, emailMessage).Return(expectedError)
		analyticsEventReporter.On(reflection.GetMethodName(analyticsEventReporter.EventOccurred), mock.Anything, email.SentEventType, emailMessage.UserID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

		err := handler.handleEmailRequest(ctx, emailMessage)
		assert.NoError(t, err) // Should not return error, just log it

		mock.AssertExpectationsForObjects(t, emailer, analyticsEventReporter)
	})

	t.Run("with analytics error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		emailer.On(reflection.GetMethodName(emailer.SendEmail), mock.Anything, emailMessage).Return(nil)
		expectedError := errors.New("analytics error")
		analyticsEventReporter.On(reflection.GetMethodName(analyticsEventReporter.EventOccurred), mock.Anything, email.SentEventType, emailMessage.UserID, mock.AnythingOfType("map[string]interface {}")).Return(expectedError)

		err := handler.handleEmailRequest(ctx, emailMessage)
		assert.NoError(t, err) // Should not return error, just log it

		mock.AssertExpectationsForObjects(t, emailer, analyticsEventReporter)
	})
}
