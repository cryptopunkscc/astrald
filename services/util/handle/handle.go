package handle

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func Requests(
	ctx context.Context,
	core api.Core,
	port string,
	authorize auth.Authorize,
	handle request.Handler,
) error {
	handler, err := register.Port(ctx, core, port)
	if err != nil {
		return err
	}
	for conn := range handler.Requests() {
		if !authorize(core, conn) {
			log.Println(port, "cannot authorize", conn.Caller())
			conn.Reject()
			continue
		}
		caller := conn.Caller()
		query := conn.Query()
		stream := accept.Request(ctx, conn)

		go func() {
			defer func() { _ = stream.Close() }()
			err := handle(caller, query, stream)
			if err != nil {
				log.Println(port, "cannot handle", err)
			}
		}()
	}
	return nil
}
