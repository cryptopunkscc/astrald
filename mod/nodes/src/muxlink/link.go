package muxlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
	"time"
)

var _ net.Link = &Link{}

var DefaultMuxHandler = func(event mux.Event) {}

const portBufferSize = 4 * 1024 * 1024
const controlPort = 0

type Link struct {
	sig.Activity
	transport     net.Conn
	localRouter   net.Router
	mux           *mux.FrameMux
	control       *Control
	remoteBuffers *remoteBuffers
	ctx           context.Context
	cancelCtx     context.CancelFunc
	mu            sync.Mutex
	err           error
	health        *health
	running       chan struct{}
}

func NewLink(transport net.Conn, localRouter net.Router) *Link {
	link := &Link{
		transport:   transport,
		running:     make(chan struct{}),
		localRouter: localRouter,
	}
	if link.localRouter == nil {
		link.localRouter = &net.NilRouter{}
	}

	link.remoteBuffers = newRemoteBuffers(link)
	link.mux = mux.NewFrameMux(transport, DefaultMuxHandler)
	link.control = NewControl(link)
	link.health = newHealth(link)
	if err := link.mux.Bind(controlPort, link.control.handleMux); err != nil {
		panic(err)
	}

	return link
}

func (link *Link) Run(ctx context.Context) error {
	link.ctx, link.cancelCtx = context.WithCancel(ctx)

	var group = tasks.Group(link.mux, link.control, link.health)

	close(link.running)
	group.Run(link.ctx)

	return link.err
}

// CloseWithError closes the link with provided error as the reason.
func (link *Link) CloseWithError(e error) error {
	link.mu.Lock()
	defer link.mu.Unlock()

	if link.err != nil {
		return nil
	}
	link.err = e

	defer link.remoteBuffers.reset(0)
	if link.cancelCtx != nil {
		defer link.cancelCtx()
	}
	return link.Transport().Close()
}

// Bind binds a specific port on the link's multiplexer to a WriteCloser.
func (link *Link) Bind(localPort int, output net.SecureWriteCloser) (binding *PortBinding, err error) {
	binding = NewPortBinding(output, link)
	err = link.mux.Bind(localPort, binding.HandleMux)

	return
}

// BindAny binds any port on the link's multiplexer to a WriteCloser.
func (link *Link) BindAny(output net.SecureWriteCloser) (binding *PortBinding, err error) {
	binding = NewPortBinding(output, link)
	_, err = link.mux.BindAny(binding.HandleMux)

	return
}

// Unbind unbinds a local port
func (link *Link) Unbind(port int) error {
	return link.mux.Unbind(port)
}

// Close closes the link.
func (link *Link) Close() error {
	return link.CloseWithError(ErrLinkClosed)
}

// LocalIdentity returns the identity of the local party.
func (link *Link) LocalIdentity() id.Identity {
	return link.transport.LocalIdentity()
}

// RemoteIdentity returns the identity of the remote party.
func (link *Link) RemoteIdentity() id.Identity {
	return link.transport.RemoteIdentity()
}

// Check marks the link for health check. This can be called many times without
// causing any congestion, as health checks are rate limited.
func (link *Link) Check() {
	link.health.Check()
}

// LocalRouter returns the upstream router to which incoming queries will be sent
func (link *Link) LocalRouter() net.Router {
	return link.localRouter
}

// SetLocalRouter sets the upstream router to which incoming queries will ne sent
func (link *Link) SetLocalRouter(uplink net.Router) {
	link.localRouter = uplink
}

func (link *Link) Transport() net.Conn {
	return link.transport
}

// Ping sends a new ping request and returns its roundtrip time
func (link *Link) Ping() (time.Duration, error) {
	return link.control.Ping()
}

// Latency returns the last measured latency of the link. It does not trigger a new measurement. If the latency
// is not known, it returns -1.
func (link *Link) Latency() time.Duration {
	return link.health.Latency()
}

// Done returns a channel that will be closed when the link closes
func (link *Link) Done() <-chan struct{} {
	<-link.running
	return link.ctx.Done()
}

// Err returns the error that caused the link to close or nil if the link is open
func (link *Link) Err() error {
	return link.err
}

func (link *Link) write(port int, frame []byte) error {
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
