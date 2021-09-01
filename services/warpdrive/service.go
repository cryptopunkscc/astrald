package warpdrive

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/storage"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"os"
)

const dirName = "wdrive"

const (
	PortLocal  = "wdrive-local"
	PortRemote = "wdrive-remote"
)

const (
	List       = 1
	SendPath   = 2
	SendStream = 3
)

const (
	Ok       = 0
	Rejected = 1
)

type remoteService struct {
	store storage.ReadWriteStorage
}

type localService struct {
	ctx   context.Context
	core  api.Core
	store storage.ReadStorage
}

func RunRemote(ctx context.Context, core api.Core) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	srv := remoteService{
		store: file.NewReadWriteDir(homeDir, dirName),
	}
	handle.Requests(ctx, core, PortRemote, auth.All, srv.receive)
	return nil
}

func RunLocal(ctx context.Context, core api.Core) error {
	srv := localService{
		ctx:   ctx,
		core:  core,
		store: file.NewReadStorage("/"),
	}
	handlers := request.Handlers{
		List:       srv.listPeers,
		SendPath:   srv.sendFromPath,
		SendStream: srv.sendFromStream,
	}
	handle.Requests(ctx, core, PortLocal, auth.Local, handle.Using(handlers))
	return nil
}
