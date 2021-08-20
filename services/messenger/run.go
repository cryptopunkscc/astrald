package messenger

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/messenger/handle"
	"github.com/cryptopunkscc/astrald/services/messenger/job"
	repoService "github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	handle2 "github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

func Run(ctx context.Context, core api.Core) error {
	r := repoService.NewRepoClient(ctx, core)
	observers := map[sio.ReadWriteCloser]struct{}{}
	go job.ObserveLore(ctx, core, Port, observers)
	return handle2.Requests(ctx, core, Port, auth.Local, func(
		caller api.Identity,
		query string,
		stream sio.ReadWriteCloser,
	) error {
		requestType, err := stream.ReadUint16()
		if err != nil {
			return err
		}
		switch requestType {
		case Send:
			handle.Send(ctx, core, r, stream, Port)
		case Observe:
			req := &request.Context{
				Port:            Port,
				ReadWriteCloser: stream,
				Observers:       observers,
			}
			_ = handle2.Observe(req)
		}
		return nil
	})
}
