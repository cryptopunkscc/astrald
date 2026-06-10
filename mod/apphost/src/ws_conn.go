package apphost

import (
	"bytes"
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

	// In JSON (text) mode the subprotocol promises one JSON envelope per
	// frame, but the bytes reaching Write come from relays (streams.Join /
	// io.Copy) that chunk arbitrarily: one chunk can carry several
	// newline-terminated objects, or a fraction of one. frameLines makes
	// Write re-frame the byte stream on newlines.
	frameLines bool
	wmu        sync.Mutex
	pending    []byte // partial line buffered until its newline arrives
}

// newWSConn adapts an accepted websocket.Conn into an io.ReadWriteCloser that streams
// each Write as one frame of msgType (binary or text). Reads stream bytes across frames.
// In text (JSON) mode, writes are re-framed so that each newline-terminated JSON line
// becomes exactly one frame. websocket.NetConn already sets ReadLimit to -1 (unlimited).
func newWSConn(ctx context.Context, c *websocket.Conn, msgType websocket.MessageType) *wsConn {
	return &wsConn{
		Conn:       websocket.NetConn(ctx, c, msgType),
		frameLines: msgType == websocket.MessageText,
	}
}

func (c *wsConn) Write(p []byte) (n int, err error) {
	if !c.frameLines {
		return c.Conn.Write(p)
	}

	c.wmu.Lock()
	defer c.wmu.Unlock()

	c.pending = append(c.pending, p...)
	for {
		i := bytes.IndexByte(c.pending, '\n')
		if i < 0 {
			break
		}
		if _, err = c.Conn.Write(c.pending[:i+1]); err != nil {
			return len(p), err
		}
		c.pending = c.pending[i+1:]
	}
	if len(c.pending) == 0 {
		c.pending = nil
	}
	return len(p), nil
}

func (c *wsConn) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.Conn.Close()
	})
	return c.closeErr
}
