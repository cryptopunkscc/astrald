package scanner

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/storage/scanner"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/accept"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/register"
	"log"
	"time"
)

func Run(ctx context.Context, core api.Core) error {
	handler, err := register.Port(ctx, core, Port)
	if err != nil {
		return err
	}
	r := repo.NewRepoClient(ctx, core)

	for request := range handler.Requests() {
		if !auth.Local(core, request) {
			request.Reject()
			continue
		}

		s := accept.Request(ctx, request)

		go func() {
			path, err := s.ReadStringWithSize16()
			if err != nil {
				log.Println(Port, "cannot read path", err)
				return
			}

			_ = s.Close()

			err = scanner.Scan(path, func(path string, modTime time.Time) {
				_, err := r.Map(path)
				if err != nil {
					log.Println(Port, "cannot map path", path, err)
				}
			})
			if err != nil {
				log.Println(Port, "scanning error", err)
				return
			}
		}()
	}

	return nil
}
