// Package livekit is the stage's video: minting the token a participant joins a
// stage's room with, and the admin calls a host's kick and mute make.
//
// Video is optional. A deployment that names no LiveKit server runs without it:
// a token request is refused with a message the web app shows in place of the
// video, and the admin calls do nothing.
package livekit

import (
	"context"
	"strings"
	"time"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"github.com/livekit/protocol/auth"
	lkproto "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// TokenTTL is how long a token admits its holder. It is long so a participant
// whose network drops late in a long session can still reconnect; the room
// goes with the stage.
const TokenTTL = 12 * time.Hour

// ErrNotConfigured is a token asked for when no LiveKit server is configured.
// Its words are what the web app shows where the video would be.
var ErrNotConfigured = platformerrors.New("Video is not available on this server.")

// Config names the LiveKit server. Every field is required for video; leaving
// all three empty runs without it, and leaving some empty is refused.
type Config struct {
	_ struct{} `json:"-"`

	APIKey    string `env:"API_KEY"    json:"apiKey,omitempty"`
	APISecret string `env:"API_SECRET" json:"apiSecret,omitempty"`
	// URL is the server's WebSocket address, as a client connects to it.
	URL string `env:"URL" json:"url,omitempty"`
}

// Enabled reports whether video is configured.
func (cfg *Config) Enabled() bool {
	return cfg.APIKey != "" && cfg.APISecret != "" && cfg.URL != ""
}

// ValidateWithContext refuses a half-configured server: some fields set and
// others not, which would look like video and fail on the first token.
func (cfg *Config) ValidateWithContext(context.Context) error {
	set := 0

	for _, field := range []string{cfg.APIKey, cfg.APISecret, cfg.URL} {
		if field != "" {
			set++
		}
	}

	if set != 0 && set != 3 {
		return platformerrors.New("livekit needs apiKey, apiSecret and url together, or none of them")
	}

	return nil
}

// Grant is who a token admits, to where, and whether they may publish.
type Grant struct {
	_ struct{} `json:"-"`

	// Room is the stage's ID.
	Room string
	// Identity is the participant's ID, which the admin calls name them by.
	Identity string
	// Name is what the room shows them as.
	Name string
	// Publish is false for the audience, who watch and do not send.
	Publish bool
}

// Minter makes access tokens.
type Minter struct {
	cfg *Config
}

// NewMinter builds a minter over the config, which may name no server.
func NewMinter(cfg *Config) *Minter {
	return &Minter{cfg: cfg}
}

// URL is where a token's holder connects.
func (m *Minter) URL() string {
	return m.cfg.URL
}

// Mint signs a token for the grant, or answers ErrNotConfigured.
func (m *Minter) Mint(grant *Grant) (string, error) {
	if !m.cfg.Enabled() {
		return "", ErrNotConfigured
	}

	video := &auth.VideoGrant{RoomJoin: true, Room: grant.Room}
	video.SetCanSubscribe(true)
	video.SetCanPublish(grant.Publish)
	video.SetCanPublishData(grant.Publish)

	token, err := auth.NewAccessToken(m.cfg.APIKey, m.cfg.APISecret).
		SetIdentity(grant.Identity).
		SetName(grant.Name).
		SetValidFor(TokenTTL).
		SetVideoGrant(video).
		ToJWT()
	if err != nil {
		return "", platformerrors.Wrap(err, "signing a LiveKit token")
	}

	return token, nil
}

// Admin is what a host does to somebody else's video. Both calls are
// best-effort from the stage's side: a participant removed from the stage is
// removed whether or not LiveKit answered, so a failure is the implementation's
// to log rather than the caller's to handle.
type Admin interface {
	// RemoveParticipant disconnects a participant from a room.
	RemoveParticipant(ctx context.Context, room, identity string)
	// SetMuted mutes a participant's microphone and stops them publishing
	// audio until unmuted, or undoes both.
	SetMuted(ctx context.Context, room, identity string, muted bool)
}

// NoopAdmin is Admin for a deployment without video.
type NoopAdmin struct{}

var _ Admin = NoopAdmin{}

// RemoveParticipant does nothing.
func (NoopAdmin) RemoveParticipant(context.Context, string, string) {}

// SetMuted does nothing.
func (NoopAdmin) SetMuted(context.Context, string, string, bool) {}

const adminName = "livekit_admin"

// RoomAdmin is Admin over LiveKit's room service.
type RoomAdmin struct {
	o11y   observability.Observer
	client *lksdk.RoomServiceClient
}

var _ Admin = (*RoomAdmin)(nil)

// NewAdmin is a RoomAdmin when video is configured and a NoopAdmin otherwise.
func NewAdmin(cfg *Config, pillars *observability.Pillars) Admin {
	if !cfg.Enabled() {
		return NoopAdmin{}
	}

	o11y := observability.NewObserver(adminName, nil, nil)
	if pillars != nil {
		o11y = observability.NewObserver(adminName, pillars.Logger, pillars.TracerProvider)
	}

	return &RoomAdmin{o11y: o11y, client: lksdk.NewRoomServiceClient(httpURL(cfg.URL), cfg.APIKey, cfg.APISecret)}
}

// httpURL is the room service's address for a server a client reaches by
// WebSocket: the same host, over HTTP.
func httpURL(wsURL string) string {
	if rest, ok := strings.CutPrefix(wsURL, "wss://"); ok {
		return "https://" + rest
	}

	if rest, ok := strings.CutPrefix(wsURL, "ws://"); ok {
		return "http://" + rest
	}

	return wsURL
}

// RemoveParticipant implements Admin.
func (a *RoomAdmin) RemoveParticipant(ctx context.Context, room, identity string) {
	ctx, op := a.o11y.Begin(ctx, observability.WithValue("livekit.room", room), observability.WithValue("livekit.identity", identity))
	defer op.End()

	if _, err := a.client.RemoveParticipant(ctx, &lkproto.RoomParticipantIdentity{Room: room, Identity: identity}); err != nil {
		op.Acknowledge(err, "removing a participant from the video room")
	}
}

// SetMuted implements Admin. Muting mutes every audio track the participant is
// publishing and narrows what they may publish to their camera and screen, so
// they cannot simply publish their microphone again; unmuting restores every
// source.
func (a *RoomAdmin) SetMuted(ctx context.Context, room, identity string, muted bool) {
	ctx, op := a.o11y.Begin(ctx,
		observability.WithValue("livekit.room", room),
		observability.WithValue("livekit.identity", identity),
		observability.WithValue("livekit.muted", muted),
	)
	defer op.End()

	info, err := a.client.GetParticipant(ctx, &lkproto.RoomParticipantIdentity{Room: room, Identity: identity})
	if err != nil {
		op.Acknowledge(err, "reading a participant in the video room")

		return
	}

	for _, track := range info.GetTracks() {
		if track.GetType() != lkproto.TrackType_AUDIO {
			continue
		}

		if _, err = a.client.MutePublishedTrack(ctx, &lkproto.MuteRoomTrackRequest{
			Room:     room,
			Identity: identity,
			TrackSid: track.GetSid(),
			Muted:    muted,
		}); err != nil {
			op.Acknowledge(err, "muting a participant's track")
		}
	}

	sources := []lkproto.TrackSource{lkproto.TrackSource_CAMERA, lkproto.TrackSource_SCREEN_SHARE}
	if !muted {
		sources = append(sources, lkproto.TrackSource_MICROPHONE, lkproto.TrackSource_SCREEN_SHARE_AUDIO)
	}

	if _, err = a.client.UpdateParticipant(ctx, &lkproto.UpdateParticipantRequest{
		Room:     room,
		Identity: identity,
		Permission: &lkproto.ParticipantPermission{
			CanSubscribe:      true,
			CanPublish:        true,
			CanPublishData:    true,
			CanPublishSources: sources,
		},
	}); err != nil {
		op.Acknowledge(err, "narrowing what a participant may publish")
	}
}
