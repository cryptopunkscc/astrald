package service

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"github.com/cryptopunkscc/astrald/components/storage/repo"
	"github.com/cryptopunkscc/astrald/services"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func (srv *Service) Run(ctx context.Context, core api.Core) error {
	srv.repo = repo.NewAdapter(file.NewStorage(services.AstralHome))
	handler, err := core.Network().Register(srv.port)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = handler.Close()
	}()
	for r := range handler.Requests() {
		if !srv.authorize(core, r) {
			log.Println(srv.port, "rejected remote connection")
			continue
		}

		stream := accept.Request(ctx, r)
		log.Println(srv.port, "accepted connection")

		go func() {
			defer stream.Close()

			requestType, err := stream.ReadByte()
			if err != nil {
				log.Println(srv.port, "error reading type", err)
				return
			}

			handle := srv.handlers[requestType]
			if handle == nil {
				log.Println(srv.port, "unknown request type", requestType)
				return
			}
			log.Println(srv.port, "request type", requestType)

			handle(&request.Context{
				Serializer:          stream,
				Port:                srv.port,
				ReadWriteRepository: srv.repo,
				Observers:           srv.observers,
			})
		}()
	}
	return nil
}
