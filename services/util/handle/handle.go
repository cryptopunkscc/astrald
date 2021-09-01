package handle

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
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
) {
	// Register port
	handler, err := register.Port(ctx, core, port)
	if err != nil {
		log.Println(port, "cannot register port")
		return
	}
	observers := map[sio.ReadWriteCloser]struct{}{}
	for conn := range handler.Requests() {
		// Authorize connection
		if !authorize(core, conn) {
			log.Println(port, "cannot authorize", conn.Caller())
			conn.Reject()
			continue
		}

		// Handle query
		rc := request.Context{
			ReadWriteCloser: accept.Request(ctx, conn),
			Caller:          conn.Caller(),
			Port:            conn.Query(),
			Observers:       observers,
		}
		go func() {
			defer func() {
				log.Println(port, "closing connection")
				_ = rc.Close()
			}()
			err := handle(rc)
			if err != nil {
				log.Println(port, "cannot handle", err)
			}
		}()
	}
}

func Using(
	handlers request.Handlers,
) request.Handler {
	return func(rc request.Context) error {
		requestType, err := rc.ReadByte()
		if err != nil {
			return err
		}
		handle, contains := handlers[requestType]
		if !contains {
			return fmt.Errorf("%s no handler for request type %d", rc.Port, requestType)
		}
		return handle(rc)
	}
}
