package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"sync"
	"sync/atomic"
)

const maxPayloadSize = 8192
const defaultBufferSize = 1024 * 1024

const (
	stateRouting = iota
	stateOpen
	stateClosed
)

var _ astral.SecureWriteCloser = &conn{}

type conn struct {
	Nonce          astral.Nonce
	RemoteIdentity id.Identity
	Outbound       bool
	Query          string
	res            chan bool

	state  atomic.Int32 // connection state
	stream *Stream      // stream the connection is attached to

	rcond *sync.Cond // sync for read functions
	rsize int        // read buffer size
	rused int        // used read buffer
	rbuf  [][]byte   // read buffer

	wcond *sync.Cond // sync for write functions
	wsize int        // remote buffer left
}

func (c *conn) Identity() id.Identity {
	return c.RemoteIdentity
}

func newConn(n astral.Nonce) *conn {
	if n == 0 {
		n = astral.NewNonce()
	}

	return &conn{
		Nonce: n,
		res:   make(chan bool, 1),
		wcond: sync.NewCond(&sync.Mutex{}),
		rcond: sync.NewCond(&sync.Mutex{}),
		rsize: defaultBufferSize,
	}
}

func (c *conn) growRemoteBuffer(s int) {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	c.wsize += s
	c.wcond.Broadcast()
}

func (c *conn) Write(p []byte) (n int, err error) {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	for {
		if len(p) == 0 {
			return
		}

		if c.state.Load() != stateOpen {
			err = errors.New("invalid state")
		}

		if c.wsize == 0 {
			c.wcond.Wait()
			continue
		}

		var l = min(c.wsize, len(p), maxPayloadSize)

		err = c.stream.Write(&frames.Data{
			Nonce:   c.Nonce,
			Payload: p[0:l],
		})
		if err != nil {
			return
		}

		c.wsize -= l
		n = n + l
		p = p[l:]
	}
}

func (c *conn) pushRead(b []byte) error {
	c.rcond.L.Lock()
	defer c.rcond.L.Unlock()

	if c.rused+len(b) > c.rsize {
		return errors.New("buffer overflow")
	}

	c.rbuf = append(c.rbuf, b)
	c.rused += len(b)

	c.rcond.Broadcast()

	return nil
}

func (c *conn) Read(p []byte) (n int, err error) {
	c.rcond.L.Lock()
	defer c.rcond.L.Unlock()

	for {
		if len(p) == 0 {
			return
		}

		if c.state.Load() != stateOpen {
			err = errors.New("invalid state")
			return
		}

		if len(c.rbuf) == 0 {
			c.rcond.Wait()
			continue
		}

		b := c.rbuf[0]
		n = min(len(p), len(b))
		copy(p, b[:n])
		b = b[n:]
		c.rused -= n
		if len(b) > 0 {
			c.rbuf[0] = b
		} else {
			c.rbuf = c.rbuf[1:]
		}

		c.stream.Write(&frames.Read{
			Nonce: c.Nonce,
			Len:   uint32(n),
		})

		return
	}
}

func (c *conn) Close() error {
	if !c.swapState(stateOpen, stateClosed) {
		return errors.New("invalid state")
	}

	c.stream.Write(&frames.Reset{Nonce: c.Nonce})

	return nil
}

func (c *conn) swapState(old, new int) bool {
	if !c.state.CompareAndSwap(int32(old), int32(new)) {
		return false
	}
	c.rcond.Broadcast()
	c.wcond.Broadcast()
	return true
}
