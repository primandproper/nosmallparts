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

func TestNewConnectionManager(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	m := NewConnectionManager(nil, nil)
	require.NotNil(t, m)

	assert.False(t, m.GroupHasConnections(ctx, "any"))
	assert.Equal(t, 0, m.GetConnectionCount(ctx, "any"))
	assert.Nil(t, m.Get(ctx, "any", "any"))
	assert.Empty(t, m.GetGroupConnections(ctx, "any"))
}

func TestConnectionManager_Add_Get_Remove(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	conn, cleanup := testConnHelper(t)
	defer cleanup()

	connCtx := NewConnectionContext(conn)
	m := NewConnectionManager(nil, nil)

	m.Add(ctx, "g1", "m1", connCtx)
	assert.True(t, m.GroupHasConnections(ctx, "g1"))
	assert.Equal(t, 1, m.GetConnectionCount(ctx, "g1"))
	assert.Equal(t, connCtx, m.Get(ctx, "g1", "m1"))
	assert.Len(t, m.GetGroupConnections(ctx, "g1"), 1)
	assert.Equal(t, connCtx, m.GetGroupConnections(ctx, "g1")[0])

	m.Remove(ctx, "g1", "m1")
	assert.False(t, m.GroupHasConnections(ctx, "g1"))
	assert.Equal(t, 0, m.GetConnectionCount(ctx, "g1"))
	assert.Nil(t, m.Get(ctx, "g1", "m1"))
	assert.Empty(t, m.GetGroupConnections(ctx, "g1"))
}

func TestConnectionManager_Remove_empties_group(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	conn1, cleanup1 := testConnHelper(t)
	defer cleanup1()
	conn2, cleanup2 := testConnHelper(t)
	defer cleanup2()

	m := NewConnectionManager(nil, nil)
	m.Add(ctx, "g1", "m1", NewConnectionContext(conn1))
	m.Add(ctx, "g1", "m2", NewConnectionContext(conn2))
	assert.Equal(t, 2, m.GetConnectionCount(ctx, "g1"))

	m.Remove(ctx, "g1", "m1")
	assert.Equal(t, 1, m.GetConnectionCount(ctx, "g1"))
	assert.NotNil(t, m.Get(ctx, "g1", "m2"))

	m.Remove(ctx, "g1", "m2")
	assert.False(t, m.GroupHasConnections(ctx, "g1"))
	assert.Equal(t, 0, m.GetConnectionCount(ctx, "g1"))
}

func TestConnectionManager_Get_nonexistent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	m := NewConnectionManager(nil, nil)
	assert.Nil(t, m.Get(ctx, "g1", "m1"))
	assert.Nil(t, m.Get(ctx, "", ""))
}

func TestConnectionManager_BroadcastToGroup(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	ch := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		var s string
		_ = conn.ReadJSON(&s)
		ch <- s
	}))
	defer server.Close()

	dialConn := func() *websocket.Conn {
		conn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		return conn
	}
	conn1 := dialConn()
	defer conn1.Close()
	conn2 := dialConn()
	defer conn2.Close()

	m := NewConnectionManager(nil, nil)
	m.Add(ctx, "g1", "m1", NewConnectionContext(conn1))
	m.Add(ctx, "g1", "m2", NewConnectionContext(conn2))

	m.BroadcastToGroup(ctx, "g1", "hello")

	got1 := <-ch
	got2 := <-ch
	assert.Equal(t, "hello", got1)
	assert.Equal(t, "hello", got2)
}

func TestConnectionManager_BroadcastToGroupFiltered(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	ch := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		var s string
		_ = conn.ReadJSON(&s)
		ch <- s
	}))
	defer server.Close()

	dialConn := func() *websocket.Conn {
		conn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		return conn
	}
	conn1 := dialConn()
	defer conn1.Close()
	conn2 := dialConn()
	defer conn2.Close()

	m := NewConnectionManager(nil, nil)
	m.Add(ctx, "g1", "m1", NewConnectionContext(conn1))
	m.Add(ctx, "g1", "m2", NewConnectionContext(conn2))

	// only include m2
	m.BroadcastToGroupFiltered(ctx, "g1", "only-m2", func(memberID string) bool {
		return memberID == "m2"
	})

	got := <-ch
	assert.Equal(t, "only-m2", got)
	select {
	case <-ch:
		t.Fatal("expected only one message (filter should exclude m1)")
	case <-time.After(100 * time.Millisecond):
		// no second message, as expected
	}
}

func TestConnectionManager_BroadcastToGroupFiltered_none_match(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	conn, cleanup := testConnHelper(t)
	defer cleanup()

	m := NewConnectionManager(nil, nil)
	m.Add(ctx, "g1", "m1", NewConnectionContext(conn))
	m.BroadcastToGroupFiltered(ctx, "g1", "nobody", func(memberID string) bool { return false })
	// no panic, no message sent (conn would block if we read; we don't read so we're good)
	cleanup()
}

func TestConnectionManager_SendToMember(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	ch := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		var s string
		_ = conn.ReadJSON(&s)
		ch <- s
	}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
	require.NoError(t, err)
	defer conn.Close()

	ctx := t.Context()

	m := NewConnectionManager(nil, nil)
	m.Add(ctx, "g1", "m1", NewConnectionContext(conn))

	err = m.SendToMember(ctx, "g1", "m1", "direct")
	require.NoError(t, err)
	assert.Equal(t, "direct", <-ch)

	// nonexistent member returns nil error (no send)
	err = m.SendToMember(ctx, "g1", "m2", "x")
	assert.NoError(t, err)
	err = m.SendToMember(ctx, "g2", "m1", "x")
	assert.NoError(t, err)
}

func TestConnectionManager_BroadcastToGroup_empty_group(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	m := NewConnectionManager(nil, nil)
	m.BroadcastToGroup(ctx, "nonexistent", "msg") // no panic
}
