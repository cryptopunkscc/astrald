package link

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"time"
)

const DefaultQueryTimeout = 30 * time.Second

func (l *Link) RouteQuery(ctx context.Context, query query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if l.closed.Load() {
		return nil, ErrLinkClosed
	}

	if !query.Target().IsEqual(l.RemoteIdentity()) {
		return nil, errors.New("target/link identity mismatch")
	}

	if !query.Caller().IsEqual(l.LocalIdentity()) {
		return nil, errors.New("caller/link identity mismatch")
	}

	if !query.Caller().IsEqual(remoteWriter.RemoteIdentity()) {
		return nil, errors.New("caller/writer identity mismatch")
	}

	// silent queries do not affect activity
	if query.Query()[0] != '.' {
		l.activity.Add(1)
		defer l.activity.Done()
	}

	remoteMonitor := &WriterMonitor{Target: remoteWriter}

	var portCh = make(chan int, 1)

	handler := NewCaptureFrameHandler(
		WriterFrameHandler{remoteMonitor},
		func(frame mux.Frame) {
			defer close(portCh)
			if frame.EOF() {
				return
			}
			if len(frame.Data) != 2 {
				return
			}
			var remotePort int

			err := cslq.Decode(bytes.NewReader(frame.Data), "s", &remotePort)
			if err != nil {
				return
			}

			portCh <- remotePort
		},
	)

	localPort, err := l.mux.BindAny(handler)
	if err != nil {
		return nil, err
	}

	if err = l.ctl.WriteQuery(query.Query(), localPort); err != nil {
		return nil, err
	}

	select {
	case remotePort, ok := <-portCh:
		if !ok {
			return nil, ErrRejected
		}

		localWriter := &WriterMonitor{Target: NewSecureFrameWriter(l, remotePort)}

		conn := NewConn(
			localPort, localWriter,
			remotePort, remoteMonitor,
			query.Query(), true,
		)

		l.add(conn)

		return localWriter, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
