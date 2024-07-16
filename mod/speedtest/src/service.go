package speedtest

import (
	"context"
	"crypto/rand"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const ServiceName = "speedtest"

// Min and max duration in seconds
const minTestDuration = 1
const maxTestDuration = 10

type Service struct {
	*Module
}

func (srv *Service) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(ServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(ServiceName)

	<-ctx.Done()
	return nil
}

func (srv *Service) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, srv.Serve)
}

func (srv *Service) Serve(conn net.Conn) {
	defer conn.Close()

	var seconds int
	cslq.Decode(conn, "c", &seconds)

	seconds = min(maxTestDuration, max(minTestDuration, seconds))

	var buf = make([]byte, 4*1024)
	rand.Read(buf)

	srv.log.Logv(1, "sending speedtest to %v for %v seconds...", conn.RemoteIdentity(), seconds)

	// send a nil error code first
	conn.Write([]byte{0})

	var sent int
	var startAt = time.Now()
	var stopAt = startAt.Add(time.Duration(seconds) * time.Second)
	for {
		if !time.Now().Before(stopAt) {
			break
		}
		n, err := conn.Write(buf)
		if err != nil {
			srv.log.Logv(1, "error sending speedtest to %v: %v", conn.RemoteIdentity(), err)
			return
		}
		sent += n
	}
	var elapsed = float64(time.Since(startAt)) / float64(time.Second)
	var speed = int(float64(sent) / elapsed)

	srv.log.Logv(1, "done sending speedtest to %v (%v bytes sent, %v/s)",
		conn.RemoteIdentity(),
		log.DataSize(sent).HumanReadable(),
		log.DataSize(speed).HumanReadable(),
	)
}
