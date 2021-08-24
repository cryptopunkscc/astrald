package identifier

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/identifier/internal"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
)

const Port = "identifier"

type service struct {
	repository repo2.LocalRepository
	observers  map[sio.ReadWriteCloser]string
	resolvers  []resolve
}

type resolve func(prefix []byte) (string, error)

func run(ctx context.Context, core api.Core) error {
	srv := service{
		repository: repo.NewRepoClient(ctx, core),
		observers:  map[sio.ReadWriteCloser]string{},
		resolvers: []resolve{
			internal.GetStoryType,
			internal.GetMimeType,
		},
	}

	// Observe repo changes
	go srv.observeRepo()

	// Handle incoming connections
	handle.Requests(ctx, core, Port, auth.Local, srv.Observe)

	return nil
}
