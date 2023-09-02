package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (link *CoreLink) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (target net.SecureWriteCloser, err error) {
	// validate identities
	if !query.Target().IsEqual(link.RemoteIdentity()) {
		return nil, errors.New("target/link identity mismatch")
	}
	if !query.Caller().IsEqual(link.LocalIdentity()) {
		return nil, errors.New("caller/link identity mismatch")
	}
	if !query.Caller().IsEqual(caller.Identity()) {
		return nil, errors.New("caller/writer identity mismatch")
	}

	// request a health check to make sure the link is responsive
	link.health.Check()

	// get a free port and bind a response handler to it
	var responseHandler = &ResponseHandler{}
	localPort, err := link.mux.BindAny(responseHandler.HandleMux)
	if err != nil {
		return nil, err
	}

	// set up response handler
	var done = make(chan struct{})
	responseHandler.Func = func(res Response, herr error) {
		defer close(done)

		// we have the response, so unbind the port so that it can be bound to the caller
		link.Unbind(localPort)

		if err = herr; err != nil {
			link.CloseWithError(err)
			return
		}

		// check error response
		err = codeToError(res.Error)
		if err != nil {
			if res.Error == errRouteNotFound {
				err = &net.ErrRouteNotFound{Router: link}
			}
			return
		}

		// rebind the port to the caller
		var binding *PortBinding
		binding, err = link.Bind(localPort, caller)
		if err != nil {
			return
		}

		if sourced, ok := net.FinalWriter(caller).(net.Sourcer); ok {
			sourced.SetSource(binding)
		}

		// grow the remote buffer for the port
		link.remoteBuffers.grow(res.Port, res.Buffer)

		// prepare the target
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
