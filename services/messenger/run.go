package messenger

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	repoService "github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

const Port = "messenger"

type service struct {
	context.Context
	api.Core
	repo.LocalRepository
}

func Run(ctx context.Context, core api.Core) error {
	observers := map[sio.ReadWriteCloser]struct{}{}
	srv := service{
		Context:         ctx,
		Core:            core,
		LocalRepository: repoService.NewRepoClient(ctx, core),
	}
	go ObserveLore(ctx, core, Port, observers)
	handle.Requests(ctx, core, Port, auth.All, func(rc request.Context) error {
		rc.Observers = observers
		log.Println(Port, "reading request type")
		requestType, err := rc.ReadByte()
		if err != nil {
			log.Println(Port, "cannot reading request type", err)
			return err
		}
		log.Println(Port, "handling request type", requestType)
		switch requestType {
		case Send:
			srv.Send(rc)
		case Observe:
			_ = handle.Observe(&rc)
		}
		return nil
	})
	return nil
}
