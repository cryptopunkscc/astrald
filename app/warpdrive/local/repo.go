package local

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"os"
	"path/filepath"
)

func NewRepository(dir string) api.Repository {
	r := &repo{}
	r.incoming.dir = filepath.Join(dir, "incoming")
	r.outgoing.dir = filepath.Join(dir, "outgoing")
	r.peers.file = filepath.Join(dir, "peers")
	_ = os.MkdirAll(r.incoming.dir, 0700)
	_ = os.MkdirAll(r.outgoing.dir, 0700)
	return r
}

type repo struct {
	incoming offersRepo
	outgoing offersRepo
	peers    peersRepo
}

func (r *repo) Incoming() api.OffersRepo {
	return &r.incoming
}

func (r *repo) Outgoing() api.OffersRepo {
	return &r.outgoing
}

func (r *repo) Peers() api.PeersRepo {
	return &r.peers
}
