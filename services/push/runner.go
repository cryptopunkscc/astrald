package push

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/fid"
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
	handler, err := register.Port(ctx, core, Port)
	if err != nil {
		return
	}
	for request := range handler.Requests() {
		caller := request.Caller()
		stream := accept.Request(ctx, request)
		log.Println(Port, "accepted connection")
		go func() {
			var idBuff [fid.Size]byte
			for {
				// Read next id
				_, err := stream.Read(idBuff[:])
				if err != nil {
					log.Println(Port, "cannot read file id", err)
					return
				}
				id := fid.Unpack(idBuff)

				// Handle id
				go func() {

					// Obtain remote reader
					r, err := repo.NewFilesClient(ctx, core, caller).Reader(id)
					if err != nil {
						log.Println(Port, "cannot obtain remote reader", err)
						return
					}

					// Obtain local writer
					w, err := repo.NewRepoClient(ctx, core).Writer()
					if err != nil {
						log.Println(Port, "cannot obtain local writer", err)
						return
					}

					// Copy file into local file system
					_, err = io.Copy(w, r)
					if err != nil {
						log.Println(Port, "cannot write copy file", err)
						return
					}
				}()
			}
		}()
	}
	return
}
