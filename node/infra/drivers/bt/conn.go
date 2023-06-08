package bt

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
	"golang.org/x/sys/unix"
	"io"
	"sync"
)

var _ net.Conn = &Conn{}

type Conn struct {
	mu             sync.RWMutex
	nfd            int
	outbount       bool
	localEndpoint  Endpoint
	remoteEndpoint Endpoint
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	fd := unix.PollFd{
		Fd:      int32(conn.nfd),
		Events:  unix.POLLIN | unix.POLLRDHUP,
		Revents: 0,
	}

	for {
		n, err = unix.Poll([]unix.PollFd{fd}, -1)

		// retry only on interrupt
		if err != unix.EINTR {
			break
		}
	}

	if n != 1 {
		return 0, err
	}

	conn.mu.RLock()
	defer conn.mu.RUnlock()

	n, err = unix.Read(conn.nfd, p)

	if n < 0 {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}

	return n, err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	n, err = unix.Write(conn.nfd, p)

	if n < 0 {
		n = 0
		err = fmt.Errorf("write error: %w", err)
		conn.Close()
	}
	return
}

func (conn *Conn) Close() error {
	return unix.Shutdown(conn.nfd, unix.SHUT_RDWR)
}

func (conn *Conn) Outbound() bool {
	return conn.outbount
}

func (conn *Conn) LocalEndpoint() net.Endpoint {
	return conn.localEndpoint
}

func (conn *Conn) RemoteEndpoint() net.Endpoint {
	return conn.remoteEndpoint
}
