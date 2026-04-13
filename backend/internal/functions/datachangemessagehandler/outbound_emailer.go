package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	localemail "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
)

func (a *AsyncDataChangeMessageHandler) OutboundEmailsEventHandler(ctx context.Context, rawMsg []byte) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	start := time.Now()

	var emailMessage email.OutboundEmailMessage
	if err := json.NewDecoder(bytes.NewReader(rawMsg)).Decode(&emailMessage); err != nil {
		return fmt.Errorf("decoding JSON body: %w", err)
	}

	if err := a.handleEmailRequest(ctx, &emailMessage); err != nil {
		return fmt.Errorf("handling outbound email request: %w", err)
	}

	a.outboundEmailsExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()))

	return nil
}

func (a *AsyncDataChangeMessageHandler) handleEmailRequest(
	ctx context.Context,
	mail *email.OutboundEmailMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if mail == nil {
		return errRequiredDataIsNil
	}

	if err := a.emailer.SendEmail(ctx, mail); err != nil {
		observability.AcknowledgeError(err, a.logger, span, "sending email")
	}

	if err := a.analyticsEventReporter.EventOccurred(ctx, localemail.SentEventType, mail.UserID, map[string]any{
		"toAddress":   mail.ToAddress,
		"toName":      mail.ToName,
		"fromAddress": mail.FromAddress,
		"fromName":    mail.FromName,
		"subject":     mail.Subject,
	}); err != nil {
		observability.AcknowledgeError(err, a.logger, span, "notifying customer data platform")
	}

	return nil
}
