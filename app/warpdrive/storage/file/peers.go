package file

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"os"
	"path/filepath"
)

var _ api.PeerStorage = Peers{}

type Peers api.Core

func (r Peers) Save(peers []api.Peer) {
	file, err := os.OpenFile(r.peers(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		r.Panicln("cannot open peers file", err)
	}
	err = gob.NewEncoder(file).Encode(peers)
	if err != nil {
		r.Panicln("cannot write peers", err)
	}
}

func (r Peers) Get() (peers api.Peers) {
	list := r.List()
	peers = api.Peers{}
	for _, peer := range list {
		p := peer
		peers[peer.Id] = &p
	}
	return
}

func (r Peers) List() (peers []api.Peer) {
	file, err := os.Open(r.peers())
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

func (r Peers) peers() string {
	return filepath.Join(r.RepositoryDir, "peers")
}
