package service

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"github.com/cryptopunkscc/astrald/components/storage/repo"
	"github.com/cryptopunkscc/astrald/services"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"log"
)

func (srv *Context) Run(ctx context.Context, core api.Core) error {
	srv.ReadWriteRepository = repo.NewAdapter(file.NewStorage(services.AstralHome))
	handler, err := register.Port(ctx, core, srv.Port)
	if err != nil {
		return err
	}

	for r := range handler.Requests() {
		if !srv.authorize(core, r) {
			log.Println(srv.Port, "rejected remote connection")
			continue
		}

		request := &Request{*srv}
		request.ReadWriteCloser = accept.Request(ctx, r)
		log.Println(srv.Port, "accepted connection")

		go func() {
			defer func() { _ = request.Close() }()

			var err error
			var requestType uint16
			var handle Handle

			if requestType, err = request.ReadUint16(); err != nil {
				log.Println(request.Port, "error reading type", err)
				return
			}
			log.Println(request.Port, "received request type", requestType, err)

			if handle = request.handlers[byte(requestType)]; handle == nil {
				log.Println(request.Port, "unknown request type", requestType)
				return
			}

			log.Println(request.Port, "handle request type", requestType)
			handle(request)
		}()
	}
	return nil
}
