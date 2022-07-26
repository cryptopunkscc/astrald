package file

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/core"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"log"
	"os"
	"path/filepath"
)

func Peers(core core.Component) storage.Peer {
	return peers{
		core.Logger,
		filepath.Join(core.RepositoryDir, "peers"),
	}
}

type peers struct {
	*log.Logger
	path string
}

func (r peers) Save(peers []warpdrive.Peer) {
	file, err := os.OpenFile(r.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		r.Panicln("cannot open peers file", err)
	}
	err = gob.NewEncoder(file).Encode(peers)
	if err != nil {
		r.Panicln("cannot write peers", err)
	}
}

func (r peers) Get() (peers warpdrive.Peers) {
	list := r.List()
	peers = warpdrive.Peers{}
	for _, peer := range list {
		p := peer
		peers[peer.Id] = &p
	}
	return
}

func (r peers) List() (peers []warpdrive.Peer) {
	file, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		r.Panicln("cannot open peers file", err)
	}
	err = gob.NewDecoder(file).Decode(&peers)
	if err != nil {
		r.Panicln("cannot read peers file", err)
	}
	return
}
