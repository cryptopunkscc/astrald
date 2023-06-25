package link

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
	"sync"
	"time"
)

var _ net.Link = &CoreLink{}

const portBufferSize = 128 * 1024
const controlPort = 0

type CoreLink struct {
	sig.Activity  //TODO: this should be in a container
	transport     net.SecureConn
	uplink        net.Router
	mux           *mux.FrameMux
	control       *Control
	remoteBuffers *remoteBuffers
	ctx           context.Context
	ctxCancel     context.CancelFunc
	mu            sync.Mutex
	err           error
	lastCheck     time.Time
	checkMu       sync.Mutex
}

func NewCoreLink(transport net.SecureConn) *CoreLink {
	link := &CoreLink{
		transport: transport,
	}

	link.remoteBuffers = newRemoteBuffers(link)
	link.mux = mux.NewFrameMux(transport, &UnexpectedFrameHandler{CoreLink: link})
	link.control = NewControl(link)
	if err := link.mux.Bind(controlPort, link.control); err != nil {
		panic(err)
	}

	return link
}

func (link *CoreLink) Run(ctx context.Context) error {
	link.ctx, link.ctxCancel = context.WithCancel(ctx)

	var group = tasks.Group(link.mux, link.control)

	group.Run(link.ctx)

	return link.err
}

func (link *CoreLink) CloseWithError(e error) error {
	link.mu.Lock()
	defer link.mu.Unlock()

	if link.err != nil {
		return nil
	}
	link.err = e

	defer link.remoteBuffers.reset(0)
	defer link.ctxCancel()
	return link.transport.Close()
}

func (link *CoreLink) Bind(localPort int, wc io.WriteCloser) error {
	err := link.mux.Bind(localPort, &PortBinding{link: link, WriteCloser: wc, port: localPort})
	if err != nil {
		return err
	}

	link.control.GrowBuffer(localPort, portBufferSize)

	return nil
}

func (link *CoreLink) BindAny(wc io.WriteCloser) (int, error) {
	var err error
	var handler = &PortBinding{
		WriteCloser: wc,
		link:        link,
	}
	handler.port, err = link.mux.BindAny(handler)

	link.control.GrowBuffer(handler.port, portBufferSize)

	return handler.port, err
}

func (link *CoreLink) Close() error {
	return link.CloseWithError(ErrLinkClosed)
}

func (link *CoreLink) LocalIdentity() id.Identity {
	return link.transport.LocalIdentity()
}

func (link *CoreLink) RemoteIdentity() id.Identity {
	return link.transport.RemoteIdentity()
}

func (link *CoreLink) Check() {
	link.checkMu.Lock()
	defer link.checkMu.Unlock()

	if time.Since(link.lastCheck) < time.Second {
		return
	}

	_, err := link.control.Ping()
	if err != nil {
		link.CloseWithError(err)
	}

	link.lastCheck = time.Now()
}

func (link *CoreLink) Uplink() net.Router {
	return link.uplink
}

func (link *CoreLink) SetUplink(uplink net.Router) {
	link.uplink = uplink
}

func (link *CoreLink) Transport() net.SecureConn {
	return link.transport
}

func (link *CoreLink) Done() <-chan struct{} {
	return link.ctx.Done()
}

func (link *CoreLink) Err() error {
	return link.err
}

func (link *CoreLink) write(port int, frame []byte) error {
	link.mu.Lock()
	defer link.mu.Unlock()

	bufferSize, open := link.remoteBuffers.sizes[port]
	if !open {
		return ErrRemoteBufferOverflow
	}

	if len(frame) > bufferSize {
		return ErrRemoteBufferOverflow
	}

	err := link.mux.Write(mux.Frame{
		Port: port,
		Data: frame,
	})
	if err != nil {
		return err
	}

	link.remoteBuffers.grow(port, -len(frame))

	return nil
}
