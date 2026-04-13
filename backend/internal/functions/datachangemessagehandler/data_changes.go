package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	localemail "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/keys"
	textsearch "github.com/verygoodsoftwarenotvirus/platform/v4/search/text"
	coreemails "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/emails"
	coreindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"
)

func (a *AsyncDataChangeMessageHandler) DataChangesEventHandler(ctx context.Context, rawMsg []byte) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	start := time.Now()

	var dataChangeMessage audit.DataChangeMessage
	if err := json.NewDecoder(bytes.NewReader(rawMsg)).Decode(&dataChangeMessage); err != nil {
		return fmt.Errorf("decoding JSON body: %w", err)
	}

	if err := a.handleDataChangeMessage(ctx, &dataChangeMessage); err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "handling data change message")
	}

	a.dataChangesExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()))

	return nil
}

func (a *AsyncDataChangeMessageHandler) handleDataChangeMessage(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)

	if changeMessage == nil {
		return errRequiredDataIsNil
	}

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	if changeMessage.UserID != "" && changeMessage.EventType != "" {
		if err := a.analyticsEventReporter.EventOccurred(ctx, changeMessage.EventType, changeMessage.UserID, changeMessage.Context); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "notifying customer data platform")
		}
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		a.nonWebhookEventTypesHat.RLock()
		eventTypeIsValid := slices.Contains(a.nonWebhookEventTypes, changeMessage.EventType)
		a.nonWebhookEventTypesHat.RUnlock()

		if changeMessage.AccountID != "" && !eventTypeIsValid {
			relevantWebhooks, err := a.webhookRepo.GetWebhooksForAccountAndEvent(ctx, changeMessage.AccountID, changeMessage.EventType)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "getting webhooks")
				return
			}

			for _, webhook := range relevantWebhooks {
				if err = a.webhookExecutionRequestPublisher.Publish(ctx, &webhooks.WebhookExecutionRequest{
					WebhookID: webhook.ID,
					AccountID: changeMessage.AccountID,
					Payload:   changeMessage,
				}); err != nil {
					observability.AcknowledgeError(err, logger, span, "publishing webhook execution request")
				}
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if err := a.handleOutboundNotifications(ctx, changeMessage); err != nil {
			observability.AcknowledgeError(err, logger, span, "notifying customer(s)")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if err := a.handleSearchIndexUpdates(ctx, changeMessage); err != nil {
			observability.AcknowledgeError(err, logger, span, "updating search index)")
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (a *AsyncDataChangeMessageHandler) handleSearchIndexUpdates(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	switch changeMessage.EventType {
	case identity.UserSignedUpServiceEventType,
		identity.UserArchivedServiceEventType,
		identity.EmailAddressChangedEventType,
		identity.UsernameChangedEventType,
		identity.UserDetailsChangedEventType,
		identity.UserEmailAddressVerifiedEventType:
		if changeMessage.UserID == "" {
			observability.AcknowledgeError(errRequiredDataIsNil, logger, span, "updating search index for User")
		}

		if err := a.searchDataIndexPublisher.Publish(ctx, &textsearch.IndexRequest{
			RowID:     changeMessage.UserID,
			IndexType: coreindexing.IndexTypeUsers,
			Delete:    changeMessage.EventType == identity.UserArchivedServiceEventType,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return nil
	default:
		logger.Debug("event type not handled for search indexing")
		return nil
	}
}

func (a *AsyncDataChangeMessageHandler) handleOutboundNotifications(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if changeMessage == nil {
		return errors.New("nil data change message")
	}

	envCfg := localemail.GetConfigForEnvironment(os.Getenv("NO_SMALL_PARTS_SERVICE_ENVIRONMENT"))
	if envCfg == nil {
		return observability.PrepareAndLogError(localemail.ErrMissingEnvCfg, a.logger, span, "getting environment queuesConfig")
	}

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	user, err := a.identityRepo.GetUser(ctx, changeMessage.UserID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "getting user")
	}

	var (
		emailType             string
		msg                   *email.OutboundEmailMessage
		outboundEmailMessages []*email.OutboundEmailMessage
	)

	switch changeMessage.EventType {
	case identity.UserSignedUpServiceEventType:
		emailType = "user signup"
		if err = a.analyticsEventReporter.AddUser(ctx, changeMessage.UserID, changeMessage.Context); err != nil {
			observability.AcknowledgeError(err, logger, span, "notifying customer data platform")
		}

		evf, parseError := parseValueFromEventContext[string](ctx, changeMessage, a.decoder, keys.UserEmailVerificationTokenKey)
		if parseError != nil {
			return observability.PrepareAndLogError(parseError, logger, span, "parsing email verification token")
		}

		msg, err = coreemails.BuildVerifyEmailAddressEmail(user, *evf, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building address verification email")
		}
		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.UserEmailAddressVerificationEmailRequestedEventType:
		emailType = "email address verification"
		evf, parseError := parseValueFromEventContext[string](ctx, changeMessage, a.decoder, keys.UserEmailVerificationTokenKey)
		if parseError != nil {
			return observability.PrepareAndLogError(parseError, logger, span, "parsing email verification token")
		}

		msg, err = coreemails.BuildVerifyEmailAddressEmail(user, *evf, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building address verification email")
		}
		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.PasswordResetTokenCreatedEventType:
		emailType = "password reset request"
		prt, parseError := parseValueFromEventContext[auth.PasswordResetToken](ctx, changeMessage, a.decoder, keys.PasswordResetTokenKey)
		if parseError != nil {
			return observability.PrepareAndLogError(parseError, logger, span, "parsing email verification token")
		}

		if prt == nil {
			return observability.PrepareError(fmt.Errorf("password reset token is nil"), span, "publishing password reset token email")
		}

		msg, err = coreemails.BuildGeneratedPasswordResetTokenEmail(user, prt, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building password reset token created email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.UsernameReminderRequestedEventType:
		emailType = "username reminder"
		msg, err = coreemails.BuildUsernameReminderEmail(user, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building username reminder email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.PasswordResetTokenRedeemedEventType:
		emailType = "password reset token redeemed"
		msg, err = coreemails.BuildPasswordResetTokenRedeemedEmail(user, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building password reset token redemption email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.PasswordChangedEventType:
		emailType = "password reset token redeemed"
		msg, err = coreemails.BuildPasswordChangedEmail(user, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building password reset token email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.AccountInvitationCreatedServiceEventType:
		emailType = "account invitation created"
		accountInvite, parseError := parseValueFromEventContext[identity.AccountInvitation](ctx, changeMessage, a.decoder, keys.AccountInvitationKey)
		if parseError != nil {
			return observability.PrepareAndLogError(parseError, logger, span, "parsing email verification token")
		}

		if accountInvite == nil {
			return observability.PrepareError(fmt.Errorf("account invitation is nil"), span, "publishing password reset token redemption email")
		}

		msg, err = coreemails.BuildInviteMemberEmail(user, accountInvite, envCfg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "building email message")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)
	}

	if len(outboundEmailMessages) == 0 {
		logger.WithValue("email_type", emailType).WithValue("outbound_emails_to_send", len(outboundEmailMessages)).Info("publishing email requests")
	}

	for _, oem := range outboundEmailMessages {
		if err = a.outboundEmailsPublisher.Publish(ctx, oem); err != nil {
			observability.AcknowledgeError(err, logger, span, "publishing %s request email", emailType)
		}
	}

	return nil
}

func parseValueFromEventContext[T any](ctx context.Context, changeMessage *audit.DataChangeMessage, decoder encoding.ServerEncoderDecoder, key string) (*T, error) {
	var x T
	if z, ok := changeMessage.Context[key]; ok {
		switch y := z.(type) {
		case string:
			if err := decoder.DecodeBytes(ctx, []byte(y), &z); err != nil {
				return nil, err
			}
		case []byte:
			z = string(y)
		}
	}

	return &x, nil
}
