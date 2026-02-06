package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConnHelper starts a minimal WebSocket server and returns a client connection.
func testConnHelper(t *testing.T) (conn *websocket.Conn, shutdown func()) {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = upgrader.Upgrade(w, r, nil)
	}))

	var err error
	conn, _, err = websocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		server.Close()
	}
}

func TestNewConnectionContext(t *testing.T) {
	t.Parallel()

	conn, cleanup := testConnHelper(t)
	defer cleanup()

	ctx := NewConnectionContext(conn)
	require.NotNil(t, ctx)
	assert.Equal(t, conn, ctx.Conn)
	assert.True(t, ctx.IsAlive)
	assert.NotNil(t, ctx.Done())
	assert.Empty(t, ctx.GroupID)
	assert.Empty(t, ctx.MemberID)
}

func TestConnectionContext_SetGroupAndMember(t *testing.T) {
	t.Parallel()

	conn, cleanup := testConnHelper(t)
	defer cleanup()

	ctx := NewConnectionContext(conn)
	ctx.SetGroupAndMember("group-1", "member-1")
	assert.Equal(t, "group-1", ctx.GroupID)
	assert.Equal(t, "member-1", ctx.MemberID)

	ctx.SetGroupAndMember("group-2", "member-2")
	assert.Equal(t, "group-2", ctx.GroupID)
	assert.Equal(t, "member-2", ctx.MemberID)
}

func TestConnectionContext_SetHeartbeatTicker(t *testing.T) {
	t.Parallel()

	conn, cleanup := testConnHelper(t)
	defer cleanup()

	ctx := NewConnectionContext(conn)
	ticker := time.NewTicker(time.Hour)
	ctx.SetHeartbeatTicker(ticker)
	// Close should stop the ticker and not panic
	ctx.Close()
	cleanup() // close underlying conn
}

func TestConnectionContext_MarkAlive_MarkDead_CheckAlive(t *testing.T) {
	t.Parallel()

	conn, cleanup := testConnHelper(t)
	defer cleanup()

	ctx := NewConnectionContext(conn)
	assert.True(t, ctx.CheckAlive())

	ctx.MarkDead()
	assert.False(t, ctx.CheckAlive())

	ctx.MarkAlive()
	assert.True(t, ctx.CheckAlive())
}

func TestConnectionContext_Done_closed_on_Close(t *testing.T) {
	t.Parallel()

	conn, cleanup := testConnHelper(t)
	defer cleanup()

	ctx := NewConnectionContext(conn)
	done := ctx.Done()
	ctx.Close()
	select {
	case <-done:
		// expected: channel closed
	case <-time.After(time.Second):
		t.Fatal("Done() channel was not closed after Close()")
	}
}

func TestConnectionContext_WriteJSON(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	received := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		var m map[string]string
		_ = conn.ReadJSON(&m)
		close(received)
	}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
	require.NoError(t, err)
	defer conn.Close()

	ctx := NewConnectionContext(conn)
	err = ctx.WriteJSON(map[string]string{"key": "value"})
	require.NoError(t, err)

	select {
	case <-received:
		// server received the message
	case <-time.After(time.Second):
		t.Fatal("server did not receive WriteJSON message")
	}
}
