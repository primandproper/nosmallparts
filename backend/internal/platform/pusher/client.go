package pusher

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	pushersdk "github.com/pusher/pusher-http-go/v5"
)

const (
	observabilityName = "pusher_client"
)

// Client wraps the Pusher SDK client.
type Client struct {
	inner   *pushersdk.Client
	config  *Config
	logger  logging.Logger
	tracer  tracing.Tracer
	enabled bool
}

// NewClient creates a new Pusher client wrapper.
func NewClient(
	cfg *Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *Client {
	l := logging.EnsureLogger(logger).WithName(observabilityName)

	if cfg == nil || !cfg.Enabled() {
		l.Info("Pusher not configured, client will be a no-op")
		return &Client{
			config:  cfg,
			logger:  l,
			tracer:  tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(observabilityName)),
			enabled: false,
		}
	}

	inner := &pushersdk.Client{
		AppID:   cfg.AppID,
		Key:     cfg.Key,
		Secret:  cfg.Secret,
		Cluster: cfg.Cluster,
		Secure:  true,
	}

	return &Client{
		inner:   inner,
		config:  cfg,
		logger:  l,
		tracer:  tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(observabilityName)),
		enabled: true,
	}
}

// Enabled returns whether the Pusher client is configured and active.
func (c *Client) Enabled() bool {
	return c.enabled
}

// Key returns the Pusher app key for client-side use.
func (c *Client) Key() string {
	if c.config == nil {
		return ""
	}
	return c.config.Key
}

// Cluster returns the Pusher cluster for client-side use.
func (c *Client) Cluster() string {
	if c.config == nil {
		return ""
	}
	return c.config.Cluster
}

// stageChannel returns the presence channel name for a stage.
func stageChannel(stageID string) string {
	return fmt.Sprintf("presence-stage-%s", stageID)
}

// castChannel returns the private channel name for cast-only messages.
func castChannel(stageID string) string {
	return fmt.Sprintf("private-cast-%s", stageID)
}

// TriggerStageEvent triggers an event on the stage's presence channel.
func (c *Client) TriggerStageEvent(ctx context.Context, stageID, event string, data any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if !c.enabled {
		c.logger.Debug("Pusher disabled, skipping TriggerStageEvent")
		return nil
	}

	if err := c.inner.Trigger(stageChannel(stageID), event, data); err != nil {
		c.logger.Error("failed to trigger stage event", err)
		return err
	}

	return nil
}

// TriggerCastEvent triggers an event on the stage's cast-only private channel.
func (c *Client) TriggerCastEvent(ctx context.Context, stageID, event string, data any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if !c.enabled {
		c.logger.Debug("Pusher disabled, skipping TriggerCastEvent")
		return nil
	}

	if err := c.inner.Trigger(castChannel(stageID), event, data); err != nil {
		c.logger.Error("failed to trigger cast event", err)
		return err
	}

	return nil
}

// PresenceUserData is the data attached to a presence channel member.
type PresenceUserData struct {
	UserInfo map[string]string `json:"user_info,omitempty"`
	UserID   string            `json:"user_id"`
}

// AuthorizePresenceChannel generates an auth response for a presence channel subscription.
func (c *Client) AuthorizePresenceChannel(socketID, channelName string, userData PresenceUserData) ([]byte, error) {
	if !c.enabled {
		return nil, fmt.Errorf("pusher not configured")
	}

	presenceData := pushersdk.MemberData{
		UserID:   userData.UserID,
		UserInfo: userData.UserInfo,
	}

	params, err := json.Marshal(map[string]string{"socket_id": socketID, "channel_name": channelName})
	if err != nil {
		return nil, fmt.Errorf("marshaling params: %w", err)
	}

	response, err := c.inner.AuthorizePresenceChannel(params, presenceData)
	if err != nil {
		return nil, fmt.Errorf("authorizing presence channel: %w", err)
	}

	return response, nil
}

// AuthorizePrivateChannel generates an auth response for a private channel subscription.
func (c *Client) AuthorizePrivateChannel(socketID, channelName string) ([]byte, error) {
	if !c.enabled {
		return nil, fmt.Errorf("pusher not configured")
	}

	params, err := json.Marshal(map[string]string{"socket_id": socketID, "channel_name": channelName})
	if err != nil {
		return nil, fmt.Errorf("marshaling params: %w", err)
	}

	response, err := c.inner.AuthorizePrivateChannel(params)
	if err != nil {
		return nil, fmt.Errorf("authorizing private channel: %w", err)
	}

	return response, nil
}

// ValidateWebhook validates a Pusher webhook signature.
func (c *Client) ValidateWebhook(signature string, body []byte) bool {
	if !c.enabled || c.config == nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(c.config.Secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// WebhookEvent represents a single event in a Pusher webhook payload.
type WebhookEvent struct {
	Name    string `json:"name"`
	Channel string `json:"channel"`
	UserID  string `json:"user_id,omitempty"`
}

// WebhookPayload represents the body of a Pusher webhook request.
type WebhookPayload struct {
	Events []WebhookEvent `json:"events"`
	TimeMS int64          `json:"time_ms"`
}

// ParseWebhookPayload parses the webhook body into a structured payload.
func ParseWebhookPayload(body []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parsing webhook payload: %w", err)
	}
	return &payload, nil
}
