package ws

import (
	"sync"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	"github.com/primandproper/primitives-go/v2/observability/logging"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
)

// The socket's liveness and limits. A peer that stops answering pings is
// dropped within pongWait, so a phone that walked out of range releases its
// participant rather than holding the stage open.
const (
	writeWait      = 10 * time.Second
	pongWait       = 75 * time.Second
	pingInterval   = 30 * time.Second
	maxFrameBytes  = 64 << 10
	sendQueueDepth = 64
)

// conn is one socket. Its reads happen on the goroutine ServeHTTP runs on, and
// its writes on a goroutine of its own, since a gorilla connection takes one
// writer at a time; everything else hands it bytes through send.
type conn struct {
	ws     *websocket.Conn
	logger logging.Logger
	send   chan []byte

	identity stages.Identity
	ip       string

	// stageID and participantID are set when the socket joins, on its read
	// goroutine and under the stage's lock, and read by that same goroutine.
	stageID       string
	participantID string

	// displaced is set, under the stage's lock, when a newer socket from the
	// same identity takes this one's place, so its close leaves the
	// participant where it is.
	displaced bool

	mu     sync.Mutex
	closed bool
}

func newConn(socket *websocket.Conn, identity stages.Identity, ip string, logger logging.Logger) *conn {
	return &conn{
		ws:       socket,
		logger:   logger,
		send:     make(chan []byte, sendQueueDepth),
		identity: identity,
		ip:       ip,
	}
}

// enqueue hands a frame to the writer. A socket whose queue is full is too
// slow to keep up with its stage and is closed rather than allowed to hold
// everybody else's broadcasts.
func (c *conn) enqueue(frame []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	select {
	case c.send <- frame:
	default:
		c.logger.Info("closing a socket that fell too far behind")
		c.closed = true
		close(c.send)
	}
}

// sendMessage marshals and enqueues one server message.
func (c *conn) sendMessage(msg *stagepb.ServerMessage) {
	frame, err := protojson.Marshal(msg)
	if err != nil {
		c.logger.Error("marshaling a server message", err)

		return
	}

	c.enqueue(frame)
}

// sendError tells this socket why what it asked for did not happen.
func (c *conn) sendError(text string) {
	c.sendMessage(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_Error_{Error: &stagepb.ServerMessage_Error{Message: text}}})
}

// close ends the socket once everything already queued has been written.
func (c *conn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.send)
	}
}

// writeLoop writes queued frames and pings until the queue is closed, then
// says goodbye and closes the socket, which ends the read loop too.
func (c *conn) writeLoop() {
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()

		if err := c.ws.Close(); err != nil {
			c.logger.Debug("closing a socket: " + err.Error())
		}
	}()

	for {
		select {
		case frame, ok := <-c.send:
			if err := c.ws.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if !ok {
				closing := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
				if err := c.ws.WriteMessage(websocket.CloseMessage, closing); err != nil {
					// The peer may already be gone, and then there is nobody to
					// tell.
					c.logger.Debug("saying goodbye on a socket: " + err.Error())
				}

				return
			}

			if err := c.ws.WriteMessage(websocket.TextMessage, frame); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}
