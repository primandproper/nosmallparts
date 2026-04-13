package datachangemessagehandler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/httptracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
)

func (a *AsyncDataChangeMessageHandler) WebhookExecutionRequestsEventHandler(ctx context.Context, rawMsg []byte) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	start := time.Now()

	var webhookExecutionRequest webhooks.WebhookExecutionRequest
	if err := a.decoder.DecodeBytes(ctx, rawMsg, &webhookExecutionRequest); err != nil {
		return fmt.Errorf("decoding JSON body: %w", err)
	}

	if err := a.handleWebhookExecutionRequest(ctx, &webhookExecutionRequest); err != nil {
		return fmt.Errorf("handling webhook execution request: %w", err)
	}

	a.webhookExecutionTimestampHistogram.Record(ctx, float64(time.Since(start).Milliseconds()))

	return nil
}

func (a *AsyncDataChangeMessageHandler) handleWebhookExecutionRequest(
	ctx context.Context,
	webhookExecutionRequest *webhooks.WebhookExecutionRequest,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if webhookExecutionRequest == nil {
		return errRequiredDataIsNil
	}

	logger := a.logger.WithValue(keys.RequestIDKey, webhookExecutionRequest.RequestID)

	account, err := a.identityRepo.GetAccount(ctx, webhookExecutionRequest.AccountID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "getting account")
	}

	webhook, err := a.webhookRepo.GetWebhook(ctx, webhookExecutionRequest.WebhookID, webhookExecutionRequest.AccountID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting webhook")
		return nil
	}

	var payloadBody []byte
	switch webhook.ContentType {
	case encoding.ContentTypeToString(encoding.ContentTypeJSON):
		payloadBody, err = json.Marshal(webhookExecutionRequest.Payload)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "marshaling webhook payload")
		}
	case encoding.ContentTypeToString(encoding.ContentTypeXML):
		payloadBody, err = xml.Marshal(webhookExecutionRequest.Payload)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "marshaling webhook payload")
		}
	}

	req, err := http.NewRequestWithContext(ctx, webhook.Method, webhook.URL, bytes.NewReader(payloadBody))
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating webhook request")
	}

	req.Header.Set("Content-Type", webhook.ContentType)

	decryptedKey, err := hex.DecodeString(account.WebhookEncryptionKey)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "decoding webhook encryption key")
	}

	digest := hmac.New(sha256.New, decryptedKey)
	digest.Write(payloadBody)
	req.Header.Set("X-Dinner-Done-Better-Signature", hex.EncodeToString(digest.Sum(nil)))

	res, err := httptracing.BuildTracedHTTPClient().Do(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "executing webhook request")
		return nil
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Error("closing response body", err)
		}
	}()

	logger = logger.WithResponse(res)
	tracing.AttachResponseToSpan(span, res)

	if res.StatusCode < 200 || res.StatusCode > 299 {
		observability.AcknowledgeError(err, logger, span, "invalid response type")
		return nil
	}

	return nil
}
