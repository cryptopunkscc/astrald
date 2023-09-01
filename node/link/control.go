package link

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"math/rand"
	"time"
)

const pingTimeout = 15 * time.Second
const maxConcurrentPings = 10

type Control struct {
	*CoreLink
	notify map[int][]chan struct{}
	pings  map[int]chan struct{}
	nonce  int
}

func NewControl(link *CoreLink) *Control {
	return &Control{
		CoreLink: link,
		notify:   map[int][]chan struct{}{},
		pings:    map[int]chan struct{}{},
	}
}

func (c *Control) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *Control) HandleFrame(frame mux.Frame) {
	c.Touch()

	if frame.IsEmpty() {
		c.CloseWithError(io.EOF)
		return
	}

	var r = bytes.NewReader(frame.Data[1:])
	switch frame.Data[0] {
	case codePing:
		cslq.Invoke(r, c.handlePing)
	case codePong:
		cslq.Invoke(r, c.handlePong)
	case codeGrowBuffer:
		cslq.Invoke(r, c.handleGrowBuffer)
	case codeReset:
		cslq.Invoke(r, c.handleReset)
	case codeQuery:
		cslq.Invoke(r, c.handleQuery)
	default:
		c.CloseWithError(ErrProtocolError)
	}
}

func (c *Control) AfterUnbind() {
	c.CloseWithError(ErrLinkClosedByPeer)
}

// Ping sends a ping request and waits for the response. Returns roundtrip time or an error.
// Errors: ErrTooManyPings, ErrPingTimeout.
func (c *Control) Ping() (time.Duration, error) {
	if len(c.pings) > maxConcurrentPings {
		return 0, ErrTooManyPings
	}

	var nonce = rand.Int() & 0x7fffffff
	var pingFrame = &bytes.Buffer{}
	cslq.Encode(pingFrame, "cv", codePing, Ping{Nonce: nonce})

	var ch = make(chan struct{})
	c.pings[nonce] = ch
	var pingAt = time.Now()

	c.mux.Write(mux.Frame{Data: pingFrame.Bytes()})

	select {
	case <-ch:
		return time.Since(pingAt), nil
	case <-time.After(pingTimeout):
		return 0, ErrPingTimeout
	}
}

// handlePing is called when a Ping message is received
func (c *Control) handlePing(msg Ping) error {
	return c.Pong(msg.Nonce)
}

// Pong sends a Pong message with the provided nonce
func (c *Control) Pong(nonce int) error {
	var buf = &bytes.Buffer{}
	cslq.Encode(buf, "cv", codePong, Pong{Nonce: nonce})
	return c.mux.Write(mux.Frame{Data: buf.Bytes()})
}

// handlePong is called when a Pong message is received
func (c *Control) handlePong(msg Pong) error {
	ping, found := c.pings[msg.Nonce]
	if !found {
		return c.CloseWithError(ErrInvalidNonce)
	}
	delete(c.pings, msg.Nonce)
	close(ping)
	return nil
}

// GrowBuffer sends a GrowBuffer message to indicate that there is more space in port's receive buffer
func (c *Control) GrowBuffer(port int, size int) error {
	var buf = &bytes.Buffer{}
	cslq.Encode(buf, "cv", codeGrowBuffer, GrowBuffer{
		Port: port,
		Size: size,
	})
	return c.mux.Write(mux.Frame{Data: buf.Bytes()})
}

func (c *Control) handleGrowBuffer(msg GrowBuffer) error {
	c.remoteBuffers.grow(msg.Port, msg.Size)

	return nil
}

// Reset sends a Reset message to indicate that local port has closed and should not be sent any data
func (c *Control) Reset(port int) error {
	var buf = &bytes.Buffer{}
	cslq.Encode(buf, "cv", codeReset, Reset{
		Port: port,
	})
	return c.mux.Write(mux.Frame{Data: buf.Bytes()})
}

func (c *Control) handleReset(msg Reset) error {
	c.remoteBuffers.reset(msg.Port)
	return nil
}

// Query sends a Query messsage to the remote party
func (c *Control) Query(query string, localPort int) error {
	var buf = &bytes.Buffer{}
	cslq.Encode(buf, "cv", codeQuery, Query{
		Service: query,
		Port:    localPort,
		Buffer:  portBufferSize,
	})
	return c.mux.Write(mux.Frame{Data: buf.Bytes()})
}

func (c *Control) handleQuery(msg Query) error {
	// queries can take a long time to finish, so run them in a goroutine
	go func() {
		defer debug.SaveLog(func(p any) {
			c.Close()
		})
		c.executeQuery(msg)
	}()

	return nil
}

// executeQuery executes an incoming query
func (c *Control) executeQuery(msg Query) error {
	var query = net.NewQuery(c.RemoteIdentity(), c.LocalIdentity(), msg.Service)

	var portWriter = NewPortWriter(c.CoreLink, msg.Port)

	// lock the port writer so that the target cannot write to it before we get a chance to send the query response
	portWriter.Lock()
	defer portWriter.Unlock()

	caller := net.NewSecureWriteCloser(portWriter, c.RemoteIdentity())

	// route the query upstream
	target, err := c.uplink.RouteQuery(c.ctx, query, caller, net.Hints{Origin: net.OriginNetwork})
	if err != nil {
		var code = errRejected
		if errors.Is(err, &net.ErrRouteNotFound{}) {
			code = errRouteNotFound
		}
		return c.WriteResponse(msg.Port, &Response{Error: code})
	}

	c.remoteBuffers.grow(msg.Port, msg.Buffer)

	// asign a local port to the target
	binding, err := c.BindAny(target)
	if err != nil {
		target.Close()
		return c.WriteResponse(msg.Port, &Response{Error: errUnexpected})
	}

	if sourcer, ok := net.FinalWriter(target).(net.Sourcer); ok {
		sourcer.SetSource(binding)
	}

	return c.WriteResponse(msg.Port, &Response{Port: binding.port, Buffer: portBufferSize})
}

func (c *Control) WriteResponse(port int, r *Response) error {
	var buf = &bytes.Buffer{}

	if err := cslq.Encode(buf, "v", r); err != nil {
		return err
	}

	return c.mux.Write(mux.Frame{
		Port: port,
		Data: buf.Bytes(),
	})
}
