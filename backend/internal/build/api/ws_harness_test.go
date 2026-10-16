package api_test

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/guests"
	"github.com/primandproper/nosmallparts/backend/internal/proto/scriptspb"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"
	"github.com/primandproper/nosmallparts/backend/internal/services/stages/ws"

	"github.com/primandproper/primitives-go/v2/database"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
	"github.com/samber/do/v2"
	"github.com/shoenig/test/must"
	"google.golang.org/protobuf/encoding/protojson"
)

// The stage cases drive real sockets against a real server, so where the
// TypeScript suite awaited each message's handling, these wait for what the
// message should have caused, up to waitFor. The timed waits are the old
// suite's, against the same shortened timers.
const (
	waitFor        = 5 * time.Second
	countdownWait  = 120 * time.Millisecond
	vacancyWait    = 1200 * time.Millisecond
	cleanupWait    = 1250 * time.Millisecond
	pollEvery      = 5 * time.Millisecond
	vacancyMessage = "is vacant"
)

// wsIdentity is who a test socket connects as: a signed-in user's token, a
// guest cookie, both or neither, from an address.
type wsIdentity struct {
	user    *testUser
	guestID string
	ip      string
}

func guestIdentity() wsIdentity {
	return wsIdentity{guestID: ulid.Make().String()}
}

func (h *harness) userIdentity(t *testing.T) wsIdentity {
	t.Helper()

	return wsIdentity{user: h.signUpUser(t, "")}
}

// wsClient is a test's socket, recording everything the server sends it.
type wsClient struct {
	t        *testing.T
	conn     *websocket.Conn
	done     chan struct{}
	received []*stagepb.ServerMessage
	mu       sync.Mutex
	closed   bool
}

// connect opens a stage socket as the identity.
func (h *harness) connect(t *testing.T, who wsIdentity) *wsClient {
	t.Helper()

	header := http.Header{}
	if who.guestID != "" {
		header.Set("Cookie", (&http.Cookie{Name: guests.CookieName, Value: who.guestID}).String())
	}

	if who.ip != "" {
		header.Set("X-Forwarded-For", who.ip)
	}

	dialer := *websocket.DefaultDialer
	dialer.Subprotocols = []string{ws.Subprotocol}

	if who.user != nil {
		dialer.Subprotocols = append(dialer.Subprotocols, ws.BearerPrefix+who.user.Token)
	}

	conn, res, err := dialer.DialContext(t.Context(), "ws"+strings.TrimPrefix(h.httpURL, "http")+ws.Route, header)
	must.NoError(t, err)
	must.NoError(t, res.Body.Close())
	must.EqOp(t, ws.Subprotocol, conn.Subprotocol())

	c := &wsClient{t: t, conn: conn, done: make(chan struct{})}

	go c.read()

	t.Cleanup(c.disconnect)

	return c
}

func (c *wsClient) read() {
	defer close(c.done)

	for {
		_, frame, err := c.conn.ReadMessage()
		if err != nil {
			c.mu.Lock()
			c.closed = true
			c.mu.Unlock()

			return
		}

		msg := &stagepb.ServerMessage{}
		if err = protojson.Unmarshal(frame, msg); err != nil {
			c.t.Errorf("the server sent a frame that is not a ServerMessage: %v: %s", err, frame)

			continue
		}

		c.mu.Lock()
		c.received = append(c.received, msg)
		c.mu.Unlock()
	}
}

// joinClient connects and joins in one call, and waits to be admitted.
func (h *harness) joinClient(t *testing.T, who wsIdentity, stageIDOrCode, name string) *wsClient {
	t.Helper()

	c := h.connect(t, who)
	c.join(stageIDOrCode, name)
	c.waitFor("to be admitted", func() bool { return c.participantID() != "" })

	return c
}

func (c *wsClient) join(stageIDOrCode, name string) {
	c.send(&stagepb.ClientMessage{Message: &stagepb.ClientMessage_Join_{Join: &stagepb.ClientMessage_Join{
		StageId:         stageIDOrCode,
		ParticipantName: name,
	}}})
}

func (c *wsClient) send(msg *stagepb.ClientMessage) {
	c.t.Helper()

	frame, err := protojson.Marshal(msg)
	must.NoError(c.t, err)
	c.sendRaw(string(frame))
}

func (c *wsClient) sendRaw(frame string) {
	c.t.Helper()

	must.NoError(c.t, c.conn.WriteMessage(websocket.TextMessage, []byte(frame)))
}

// disconnect closes the socket from this side, as a browser tab closing does,
// and waits for the server to hang up.
func (c *wsClient) disconnect() {
	_ = c.conn.Close()
	<-c.done
}

func (c *wsClient) messages() []*stagepb.ServerMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	return slices.Clone(c.received)
}

func (c *wsClient) closedByServer() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}

func (c *wsClient) participantID() string {
	for _, msg := range c.messages() {
		if joined := msg.GetJoined(); joined != nil {
			return joined.GetParticipantId()
		}
	}

	return ""
}

// lastState is the newest stage state this socket was sent, joined or
// broadcast.
func (c *wsClient) lastState() *stagepb.StageState {
	var last *stagepb.StageState

	for _, msg := range c.messages() {
		if state := msg.GetStageState(); state != nil {
			last = state
		} else if joined := msg.GetJoined(); joined != nil {
			last = joined.GetStageState()
		}
	}

	return last
}

func (c *wsClient) me() *stagepb.Participant {
	return participantIn(c.lastState(), c.participantID())
}

func participantIn(state *stagepb.StageState, id string) *stagepb.Participant {
	for _, p := range state.GetParticipants() {
		if p.GetId() == id {
			return p
		}
	}

	return nil
}

func (c *wsClient) errors() []string {
	var out []string

	for _, msg := range c.messages() {
		if e := msg.GetError(); e != nil {
			out = append(out, e.GetMessage())
		}
	}

	return out
}

func (c *wsClient) systemMessages() []string {
	var out []string

	for _, msg := range c.messages() {
		if sys := msg.GetSystemMessage(); sys != nil {
			out = append(out, sys.GetText())
		}
	}

	return out
}

func (c *wsClient) chats() []string {
	var out []string

	for _, msg := range c.messages() {
		if chat := msg.GetChatMessage(); chat != nil {
			out = append(out, chat.GetText())
		}
	}

	return out
}

// count is how many messages this socket was sent that kind matches.
func (c *wsClient) count(kind func(*stagepb.ServerMessage) bool) int {
	n := 0

	for _, msg := range c.messages() {
		if kind(msg) {
			n++
		}
	}

	return n
}

func (c *wsClient) hasError(text string) bool {
	return slices.Contains(c.errors(), text)
}

func (c *wsClient) hasErrorContaining(fragment string) bool {
	return slices.ContainsFunc(c.errors(), func(e string) bool { return strings.Contains(e, fragment) })
}

func (c *wsClient) hasSystemMessageContaining(fragment string) bool {
	return slices.ContainsFunc(c.systemMessages(), func(m string) bool { return strings.Contains(m, fragment) })
}

// waitFor fails the test unless cond comes true within waitFor.
func (c *wsClient) waitFor(what string, cond func() bool) {
	c.t.Helper()

	eventually(c.t, what, cond)
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(waitFor)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting %s for %s", waitFor, what)
		}

		time.Sleep(pollEvery)
	}
}

// sync makes sure the server has handled everything this socket sent: it sends
// a heartbeat and waits for the state the heartbeat answers with. Messages on
// one socket are handled in order, so anything sent before is done.
func (c *wsClient) sync() {
	c.t.Helper()

	isState := func(msg *stagepb.ServerMessage) bool { return msg.GetStageState() != nil }
	before := c.count(isState)

	c.send(heartbeat())
	c.waitFor("a heartbeat's answer", func() bool { return c.count(isState) > before })
}

func heartbeat() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_Heartbeat_{Heartbeat: &stagepb.ClientMessage_Heartbeat{}}}
}

func toggleCastMember(castMemberID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_ToggleCastMember_{
		ToggleCastMember: &stagepb.ClientMessage_ToggleCastMember{CastMemberId: castMemberID},
	}}
}

func setAudience(audience bool) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_SetAudience_{
		SetAudience: &stagepb.ClientMessage_SetAudience{IsAudience: audience},
	}}
}

func setReady(ready bool) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_SetReady_{SetReady: &stagepb.ClientMessage_SetReady{IsReady: ready}}}
}

func selectPlay(playID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_SelectPlay_{SelectPlay: &stagepb.ClientMessage_SelectPlay{PlayId: playID}}}
}

func startPlay() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_StartPlay_{StartPlay: &stagepb.ClientMessage_StartPlay{}}}
}

func nextLine() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_NextLine_{NextLine: &stagepb.ClientMessage_NextLine{}}}
}

func resetPlay() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_ResetPlay_{ResetPlay: &stagepb.ClientMessage_ResetPlay{}}}
}

func hostResume() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostResume_{HostResume: &stagepb.ClientMessage_HostResume{}}}
}

func hostEndPlay() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostEndPlay_{HostEndPlay: &stagepb.ClientMessage_HostEndPlay{}}}
}

func audienceClaimRole(castMemberID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_AudienceClaimRole_{
		AudienceClaimRole: &stagepb.ClientMessage_AudienceClaimRole{CastMemberId: castMemberID},
	}}
}

func sendChat(text string, castOnly bool) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_SendChat_{
		SendChat: &stagepb.ClientMessage_SendChat{Text: text, CastOnly: castOnly},
	}}
}

func isKind[T any](get func(*stagepb.ServerMessage) *T) func(*stagepb.ServerMessage) bool {
	return func(msg *stagepb.ServerMessage) bool { return get(msg) != nil }
}

// makeStage creates a stage the way the TypeScript suite's helper did, straight
// through the store, so a case about joining needs nobody signed in.
func (h *harness) makeStage(t *testing.T, playID string) *stages.Stage {
	t.Helper()

	client := do.MustInvoke[database.Client](h.injector)
	store := do.MustInvoke[stages.Store](h.injector)

	stage := &stages.Stage{ID: ulid.Make().String(), Code: stages.NewCode(), PlayScriptID: playID}
	must.NoError(t, client.WithTransaction(t.Context(), func(tx database.Tx) error {
		return store.CreateStage(t.Context(), tx, stage)
	}))

	return stage
}

// makePlay is the sample script, owned by somebody new: cast "Alice" and
// "Bob", lines spoken Alice, Bob, Alice.
func (h *harness) makePlay(t *testing.T) *scriptspb.Play {
	t.Helper()

	return h.createScript(t, h.signUpUser(t, ""), nil)
}

func castIDs(play *scriptspb.Play) (a, b string) {
	return play.GetCast()[0].GetId(), play.GetCast()[1].GetId()
}

func phaseOf(c *wsClient) stagepb.StagePhase {
	return c.lastState().GetPhase()
}

func describe(c *wsClient) string {
	return fmt.Sprintf("errors %q, last state %v", c.errors(), c.lastState())
}
