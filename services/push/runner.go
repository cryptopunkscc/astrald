package push

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"io"
	"log"
)

func init() {
	_ = node.RegisterService(Port, run)
}

const Port = "push"

func run(ctx context.Context, core api.Core) (err error) {
	var handler api.PortHandler
	if handler, err = register.Port(ctx, core, Port); err != nil {
		return
	}

	for request := range handler.Requests() {
		caller := request.Caller()
		stream := accept.Request(ctx, request)
		log.Println(Port, "accepted connection")

		go func() {
			for {
				var idBuff [fid.Size]byte

				// Read next id
				if _, err := stream.Read(idBuff[:]); err != nil {
					log.Println(Port, "cannot read file id", err)
					return
				}

				// Handle id
				go func() {
					var err error
					var r repo2.Reader
					var w repo2.Writer

					// Obtain remote reader
					id := fid.Unpack(idBuff)
					if r, err = repo.NewFilesClient(ctx, core, caller).Reader(id); err != nil {
						log.Println(Port, "cannot obtain remote reader", err)
						return
					}

					// Obtain local writer
					if w, err = repo.NewRepoClient(ctx, core).Writer(); err != nil {
						log.Println(Port, "cannot obtain local writer", err)
						return
					}

					// Copy file into local file system
					defer func() { _ = r.Close() }()
					defer func() { _, _ = w.Finalize() }()
					if _, err = io.Copy(w, r); err != nil {
						log.Println(Port, "cannot write copy file", err)
						return
					}
				}()
			}
		}()
	}
	return
}
