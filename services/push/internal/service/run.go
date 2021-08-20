package service

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"log"
)

func (srv *Context) Run(ctx context.Context, core api.Core) (err error) {
	var handler api.PortHandler
	if handler, err = register.Port(ctx, core, srv.Port); err != nil {
		return
	}

	srv.Ctx = ctx
	srv.Core = core
	srv.Observers = map[sio.ReadWriteCloser]struct{}{}

	for conn := range handler.Requests() {
		r := &Request{
			Context:         *srv,
			ReadWriteCloser: accept.Request(ctx, conn),
			Caller:          conn.Caller(),
		}
		log.Println(srv.Port, "accepted connection")
		go func() {
			defer func() { _ = r.Close() }()
			var err error
			var requestType byte

			if requestType, err = r.ReadByte(); err != nil {
				log.Println(srv.Port, "cannot read request type", err)
				return
			}

			handle := srv.Handlers[requestType]
			if handle != nil {
				log.Println(srv.Port, "cannot obtain handler for request type", requestType, err)
				return
			}
			err = handle(r)
			if err != nil {
				log.Println(srv.Port, "cannot obtain handler for request type", requestType, err)
			}
		}()
	}
	return
}
