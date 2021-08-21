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
			var requestType uint16

			if requestType, err = r.ReadUint16(); err != nil {
				log.Println(srv.Port, "cannot read request type", err)
				return
			}

			log.Println(srv.Port, "getting handler for request type", requestType)
			handle := srv.Handlers[byte(requestType)]
			if handle == nil {
				log.Println(srv.Port, "cannot obtain handler for request type", requestType, "len", len(srv.Handlers), srv.Handlers, err)
				return
			}
			log.Println(srv.Port, "handling request type", requestType)
			err = handle(r)
			if err != nil {
				log.Println(srv.Port, "cannot handle request type", requestType, err)
			}
		}()
	}
	return
}
