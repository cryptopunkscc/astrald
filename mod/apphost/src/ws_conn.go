package apphost

import (
	"context"
	"net"
	"sync"

	"github.com/coder/websocket"
)

// wsConn wraps the net.Conn returned by websocket.NetConn with an idempotent Close.
// The native protocol's post-accept handover passes the conn to streams.Join, which
// closes both sides on EOF — and Serve also closes on return. Without sync.Once we
// would call websocket.Conn.Close twice and the second call would return an error.
type wsConn struct {
	net.Conn
	closeOnce sync.Once
	closeErr  error
}

// newWSConn adapts an accepted websocket.Conn into an io.ReadWriteCloser that streams
// each Write as one frame of msgType (binary or text). Reads stream bytes across frames.
// websocket.NetConn already sets ReadLimit to -1 (unlimited).
func newWSConn(ctx context.Context, c *websocket.Conn, msgType websocket.MessageType) *wsConn {
	return &wsConn{Conn: websocket.NetConn(ctx, c, msgType)}
}

func (c *wsConn) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.Conn.Close()
	})
	return c.closeErr
}
