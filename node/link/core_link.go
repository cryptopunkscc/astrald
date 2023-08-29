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

const portBufferSize = 4 * 1024 * 1024
const controlPort = 0

type CoreLink struct {
	sig.Activity
	transport     net.SecureConn
	uplink        net.Router
	mux           *mux.FrameMux
	control       *Control
	remoteBuffers *remoteBuffers
	ctx           context.Context
	ctxCancel     context.CancelFunc
	mu            sync.Mutex
	err           error
	health        *health
	running       chan struct{}
}

func NewCoreLink(transport net.SecureConn) *CoreLink {
	link := &CoreLink{
		transport: transport,
		running:   make(chan struct{}),
	}

	link.remoteBuffers = newRemoteBuffers(link)
	link.mux = mux.NewFrameMux(transport, &UnexpectedFrameHandler{CoreLink: link})
	link.control = NewControl(link)
	link.health = newHealth(link)
	if err := link.mux.Bind(controlPort, link.control); err != nil {
		panic(err)
	}

	return link
}

func (link *CoreLink) Run(ctx context.Context) error {
	link.ctx, link.ctxCancel = context.WithCancel(ctx)

	var group = tasks.Group(link.mux, link.control, link.health)

	close(link.running)
	group.Run(link.ctx)

	return link.err
}

// CloseWithError closes the link with provided error as the reason.
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

// Bind binds a specific port on the link's multiplexer to a WriteCloser.
func (link *CoreLink) Bind(localPort int, wc io.WriteCloser) error {
	err := link.mux.Bind(localPort, NewPortBinding(wc, link, localPort))
	if err != nil {
		return err
	}

	return nil
}

// BindAny binds any port on the link's multiplexer to a WriteCloser.
func (link *CoreLink) BindAny(wc io.WriteCloser) (int, error) {
	var err error
	var handler = NewPortBinding(wc, link, 0)
	handler.port, err = link.mux.BindAny(handler)

	link.control.GrowBuffer(handler.port, portBufferSize)

	return handler.port, err
}

// Close closes the link.
func (link *CoreLink) Close() error {
	return link.CloseWithError(ErrLinkClosed)
}

// LocalIdentity returns the identity of the local party.
func (link *CoreLink) LocalIdentity() id.Identity {
	return link.transport.LocalIdentity()
}

// RemoteIdentity returns the identity of the remote party.
func (link *CoreLink) RemoteIdentity() id.Identity {
	return link.transport.RemoteIdentity()
}

// Check marks the link for health check. This can be called many times without
// causing any congestion, as health checks are rate limited.
func (link *CoreLink) Check() {
	link.health.Check()
}

// Uplink returns the upstream router to which incoming queries will be sent
func (link *CoreLink) Uplink() net.Router {
	return link.uplink
}

// SetUplink sets the upstream router to which incoming queries will ne sent
func (link *CoreLink) SetUplink(uplink net.Router) {
	link.uplink = uplink
}

func (link *CoreLink) Transport() net.SecureConn {
	return link.transport
}

// Ping sends a new ping request and returns its roundtrip time
func (link *CoreLink) Ping() (time.Duration, error) {
	return link.control.Ping()
}

// Latency returns the last measured latency of the link. It does not trigger a new measurement. If the latency
// is not known, it returns -1.
func (link *CoreLink) Latency() time.Duration {
	return link.health.Latency()
}

// Done returns a channel that will be closed when the link closes
func (link *CoreLink) Done() <-chan struct{} {
	<-link.running
	return link.ctx.Done()
}

// Err returns the error that caused the link to close or nil if the link is open
func (link *CoreLink) Err() error {
	return link.err
}

func (link *CoreLink) write(port int, frame []byte) error {
	link.mu.Lock()
	defer link.mu.Unlock()

	bufferSize, open := link.remoteBuffers.size(port)
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
