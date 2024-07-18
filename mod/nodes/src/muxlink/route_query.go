package muxlink

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
)

func (link *Link) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (target astral.SecureWriteCloser, err error) {
	// validate identities
	switch {
	case !query.Target().IsEqual(link.RemoteIdentity()):
		return astral.RouteNotFound(link, errors.New("target/link identity mismatch"))

	case !query.Caller().IsEqual(link.LocalIdentity()):
		return astral.RouteNotFound(link, errors.New("caller/link identity mismatch"))

	case !query.Caller().IsEqual(caller.Identity()):
		return astral.RouteNotFound(link, errors.New("caller/writer identity mismatch"))
	}

	// request a health check to make sure the link is responsive
	link.health.Check()

	// get a free port and bind a response handler to it
	var responseHandler = &ResponseHandler{}
	localPort, err := link.mux.BindAny(responseHandler.HandleMux)
	if err != nil {
		return astral.RouteNotFound(link, err)
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
			return
		}

		// rebind the port to the caller
		_, err = link.Bind(localPort, caller)
		if err != nil {
			return
		}

		// grow the remote buffer for the port
		link.remoteBuffers.grow(res.Port, res.Buffer)

		// prepare the target
		target = NewPortWriter(link, res.Port)
	}

	// send the query to the remote peer
	if err := link.control.Query(uint64(query.Nonce()), query.Query(), localPort); err != nil {
		link.CloseWithError(err)
		return astral.RouteNotFound(link, err)
	}

	select {
	case <-done:
		return

	case <-link.Done():
		return astral.Abort()

	case <-ctx.Done():
		go func() {
			<-done
			if target != nil {
				target.Close()
			}
		}()

		return astral.Abort()
	}
}

func codeToError(code int) error {
	switch code {
	case errSuccess:
		return nil
	case errRejected:
		return astral.ErrRejected
	case errRouteNotFound:
		return astral.ErrRejected
	case errUnexpected:
		return errors.New("unexpected error")
	default:
		return errors.New("invalid error")
	}
}
