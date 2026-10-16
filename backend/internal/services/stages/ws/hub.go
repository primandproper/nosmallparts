// Package ws is the stage socket: the real-time half of a stage, where
// participants join, take roles, perform and chat.
//
// Every socket on a stage is handled under that stage's lock, so a stage
// processes one thing at a time, as the TypeScript server's event loop did:
// a message, a disconnect, or a timer firing. What has to outlive a socket is
// in the stages store; what lives only as long as the process (the sockets,
// the timers, a vacancy's countdown) is here, in a room per stage.
//
// The wire is protojson: a ClientMessage per frame in, a ServerMessage per
// frame out.
package ws

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/clientip"
	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/guests"
	"github.com/primandproper/nosmallparts/backend/internal/livekit"

	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"github.com/gorilla/websocket"
)

const hubName = "stage_socket"

// Route is where the stage socket is served.
const Route = "/ws/stage"

// MaxParticipants is how many people may be connected to one stage.
const MaxParticipants = 12

// The socket subprotocol, and the prefix of the second one a signed-in client
// offers to carry its access token. A browser cannot set an Authorization
// header on a WebSocket, and a token in the URL ends up in every proxy's log,
// so it rides in Sec-WebSocket-Protocol: ["nsp.v1", "bearer.<token>"]. Only
// "nsp.v1" is echoed back.
const (
	Subprotocol  = "nsp.v1"
	BearerPrefix = "bearer."
)

// Timings are how long the timed parts of a stage take.
type Timings struct {
	HostGrace      time.Duration
	VacancyClaim   time.Duration
	CleanupDelay   time.Duration
	IdleAfter      time.Duration
	ReaperInterval time.Duration
	Countdown      time.Duration
}

// Authenticator turns an access token into the signed-in caller it names.
// signingrpc.PrincipalExtractor is one, and checks the login is still going.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*signingrpc.Caller, error)
}

// Hub is every stage's socket state, and the reaper that deletes stages left
// idle. It is a service.Runner: Run is the reaper, and Close stops it and
// closes every socket.
type Hub struct {
	o11y   observability.Observer
	client database.Client
	store  stages.Store
	plays  scripts.Store
	auth   Authenticator
	video  livekit.Admin

	// ctx is what timers and the reaper run under; Close cancels it.
	ctx     context.Context //nolint:containedctx // a Runner's work outlives any one call's context.
	library *scripts.Service
	cancel  context.CancelFunc

	rooms   map[string]*room
	sockets map[*conn]struct{}

	stop     chan struct{}
	stopped  chan struct{}
	upgrader websocket.Upgrader
	timings  Timings

	closeOnce sync.Once

	mu      sync.Mutex
	closing bool
}

// room is the in-process state of one stage, guarded by its own lock.
type room struct {
	conns     map[string]*conn // by participant ID
	cleanup   *time.Timer
	countdown *time.Timer
	vacancies map[string]*vacancy // by cast member ID
	offers    map[string]*offer   // by offer ID

	// hostGrace runs while a disconnected host may still come back, and hands
	// the role on when it fires.
	hostGrace *time.Timer

	id string
	mu sync.Mutex

	// gone is set once the stage is deleted; a socket that was waiting for the
	// lock finds the stage no longer there.
	gone bool
}

// NewHub builds the hub. The reaper does not start until Run.
func NewHub(
	client database.Client,
	store stages.Store,
	plays scripts.Store,
	library *scripts.Service,
	auth Authenticator,
	video livekit.Admin,
	timings Timings,
	pillars *observability.Pillars,
) (*Hub, error) {
	if client == nil || store == nil || plays == nil || library == nil || auth == nil || video == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "building the stage socket hub")
	}

	o11y := observability.NewObserver(hubName, nil, nil)
	if pillars != nil {
		o11y = observability.NewObserver(hubName, pillars.Logger, pillars.TracerProvider)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		o11y:    o11y,
		client:  client,
		store:   store,
		plays:   plays,
		library: library,
		auth:    auth,
		video:   video,
		timings: timings,
		// CheckOrigin is left at gorilla's default, which admits a same-origin
		// browser and a client that sends no Origin, and refuses anything
		// else: the web app is served from this origin.
		upgrader: websocket.Upgrader{Subprotocols: []string{Subprotocol}},
		ctx:      ctx,
		cancel:   cancel,
		rooms:    map[string]*room{},
		sockets:  map[*conn]struct{}{},
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}, nil
}

// Run is the idle reaper: every ReaperInterval it deletes the stages nobody is
// connected to and nothing has happened on for IdleAfter. It reads the
// database rather than the rooms, so it also finds stages a restart orphaned.
func (h *Hub) Run() {
	defer close(h.stopped)

	ticker := time.NewTicker(h.timings.ReaperInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stop:
			return
		case <-ticker.C:
			if _, err := h.ReapOnce(h.ctx); err != nil {
				h.o11y.Logger().Error("reaping idle stages", err)
			}
		}
	}
}

// Close stops the reaper, cancels every timer, and closes every socket. A
// socket closed this way leaves its participant as it was: a restart is not a
// participant walking away.
func (h *Hub) Close(ctx context.Context) error {
	h.closeOnce.Do(func() {
		h.mu.Lock()
		h.closing = true
		sockets := make([]*conn, 0, len(h.sockets))

		for c := range h.sockets {
			sockets = append(sockets, c)
		}
		h.mu.Unlock()

		h.cancel()
		close(h.stop)

		for _, c := range sockets {
			c.close()
		}
	})

	select {
	case <-h.stopped:
		return nil
	case <-ctx.Done():
		return platformerrors.Wrap(ctx.Err(), "waiting for the stage reaper to stop")
	}
}

func (h *Hub) isClosing() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.closing
}

// ReapOnce deletes every idle stage nobody is connected to, and answers which.
func (h *Hub) ReapOnce(ctx context.Context) ([]string, error) {
	ctx, op := h.o11y.Begin(ctx)
	defer op.End()

	ids, err := h.store.IdleStageIDs(ctx, h.client.Reader(), h.client.CurrentTime().Add(-h.timings.IdleAfter))
	if err != nil {
		return nil, op.Error(err, "listing idle stages")
	}

	var deleted []string

	for _, id := range ids {
		r := h.room(id)

		r.mu.Lock()

		if r.gone || len(r.conns) > 0 {
			r.mu.Unlock()

			continue
		}

		err = h.deleteStage(ctx, r)
		r.mu.Unlock()

		if err != nil {
			return deleted, op.Error(err, "deleting idle stage %s", id)
		}

		deleted = append(deleted, id)
	}

	if len(deleted) > 0 {
		op.Logger().WithValue("stages", len(deleted)).Info("reaped idle stages")
	}

	return deleted, nil
}

// room is the stage's room, made the first time anybody asks.
func (h *Hub) room(stageID string) *room {
	h.mu.Lock()
	defer h.mu.Unlock()

	r, ok := h.rooms[stageID]
	if !ok {
		r = &room{id: stageID, conns: map[string]*conn{}, vacancies: map[string]*vacancy{}, offers: map[string]*offer{}}
		h.rooms[stageID] = r
	}

	return r
}

// existingRoom is the stage's room if it has one.
func (h *Hub) existingRoom(stageID string) (*room, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	r, ok := h.rooms[stageID]

	return r, ok
}

// deleteStage deletes a stage and everything on it, and forgets its room. The
// caller holds the room's lock.
func (h *Hub) deleteStage(ctx context.Context, r *room) error {
	r.gone = true
	r.stopTimers()

	h.mu.Lock()
	if h.rooms[r.id] == r {
		delete(h.rooms, r.id)
	}
	h.mu.Unlock()

	return h.client.WithTransaction(ctx, func(tx database.Tx) error {
		return h.store.DeleteStage(ctx, tx, r.id)
	})
}

func (r *room) stopTimers() {
	for _, timer := range []*time.Timer{r.cleanup, r.countdown, r.hostGrace} {
		if timer != nil {
			timer.Stop()
		}
	}

	r.cleanup, r.countdown, r.hostGrace = nil, nil, nil
	clear(r.offers)

	for id := range r.vacancies {
		r.clearVacancy(id)
	}
}

// later runs fn after d under the room's lock, unless the stage has been
// deleted or the hub is closing by then.
func (h *Hub) later(r *room, d time.Duration, fn func(ctx context.Context)) *time.Timer {
	return time.AfterFunc(d, func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.gone || h.isClosing() {
			return
		}

		fn(h.ctx)
	})
}

// ServeHTTP upgrades a request to the stage socket and serves it until it
// closes. Who is connecting is settled here, once: a signed-in user by the
// access token in the subprotocol header, or a guest by their cookie.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, op := h.o11y.Begin(r.Context())

	identity := h.identify(ctx, r)
	ip := clientip.FromRequest(r)

	op.End()

	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// The upgrader has answered the request with what went wrong.
		h.o11y.Logger().Debug("refusing a stage socket: " + err.Error())

		return
	}

	c := newConn(socket, identity, ip, h.o11y.Logger())

	h.mu.Lock()
	if h.closing {
		h.mu.Unlock()
		c.close()
		c.writeLoop()

		return
	}

	h.sockets[c] = struct{}{}
	h.mu.Unlock()

	go c.writeLoop()

	h.readLoop(c)

	h.mu.Lock()
	delete(h.sockets, c)
	h.mu.Unlock()
}

// identify is who a socket request is from, or nobody. A token that does not
// authenticate is ignored rather than refused, and the request falls back to
// the guest cookie, as a browser holding a stale token and a guest cookie
// should be let in as the guest.
func (h *Hub) identify(ctx context.Context, r *http.Request) stages.Identity {
	for _, offered := range websocket.Subprotocols(r) {
		token, isBearer := strings.CutPrefix(offered, BearerPrefix)
		if !isBearer || token == "" {
			continue
		}

		caller, err := h.auth.Authenticate(ctx, token)
		if err == nil && caller != nil {
			return stages.Identity{UserID: caller.UserID()}
		}

		if err != nil && !errors.Is(err, signingrpc.ErrUnauthenticated) {
			h.o11y.Logger().Error("authenticating a stage socket", err)
		}
	}

	if guestID, ok := guests.FromRequest(r); ok {
		return stages.Identity{GuestID: guestID}
	}

	return stages.Identity{}
}

// readLoop handles the socket's frames in order until it closes, then runs
// the departure.
func (h *Hub) readLoop(c *conn) {
	c.ws.SetReadLimit(maxFrameBytes)

	if err := c.ws.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.close()

		return
	}

	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, frame, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		if err = c.ws.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			break
		}

		h.handle(h.ctx, c, frame)
	}

	c.close()
	h.disconnected(h.ctx, c)
}
