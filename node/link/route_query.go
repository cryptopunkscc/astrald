package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (link *CoreLink) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (target net.SecureWriteCloser, err error) {
	if !query.Target().IsEqual(link.RemoteIdentity()) {
		return nil, errors.New("target/link identity mismatch")
	}
	if !query.Caller().IsEqual(link.LocalIdentity()) {
		return nil, errors.New("caller/link identity mismatch")
	}
	if !query.Caller().IsEqual(caller.Identity()) {
		return nil, errors.New("caller/writer identity mismatch")
	}

	link.health.Check()

	var responseHandler = &ResponseHandler{}
	localPort, err := link.mux.BindAny(responseHandler)
	if err != nil {
		return nil, err
	}

	// initialize buffer for response handling
	link.control.GrowBuffer(localPort, 1024*16)

	// set up response handler
	var done = make(chan struct{})
	responseHandler.Func = func(res Response, herr error) {
		defer close(done)

		// we only want to capture the first frame containing the response, so unbind
		link.mux.Unbind(localPort)

		if err = herr; err != nil {
			link.control.Reset(localPort)
			return
		}
		if err = codeToError(res.Error); err != nil {
			link.control.Reset(localPort)
			if res.Error == errRouteNotFound {
				err = &net.ErrRouteNotFound{Router: link}
			}
			return
		}

		err = link.Bind(localPort, caller)

		target = net.NewSecureWriteCloser(
			NewPortWriter(link, res.Port),
			link.RemoteIdentity(),
		)
	}

	// send the query to the remote peer
	if err := link.control.Query(query.Query(), localPort); err != nil {
		link.CloseWithError(err)
		return nil, err
	}
	select {
	case <-done:
		return
	case <-ctx.Done():
		go func() {
			<-done
			if target != nil {
				target.Close()
			}
		}()

		return nil, ctx.Err()
	}
}

func codeToError(code int) error {
	switch code {
	case errSuccess:
		return nil
	case errRejected:
		return net.ErrRejected
	case errRouteNotFound:
		return &net.ErrRouteNotFound{}
	case errUnexpected:
		return errors.New("unexpected error")
	default:
		return errors.New("invalid error")
	}
}
