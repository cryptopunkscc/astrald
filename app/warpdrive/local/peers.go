package local

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"log"
	"os"
)

type peersRepo struct {
	file string
}

func (r *peersRepo) Save(peers []api.Peer) {
	file, err := os.OpenFile(r.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		log.Panicln("cannot open peers file", err)
	}
	err = gob.NewEncoder(file).Encode(peers)
	if err != nil {
		log.Panicln("cannot write peers", err)
	}
}

func (r *peersRepo) List() (peers []api.Peer) {
	file, err := os.Open(r.file)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Panicln("cannot open peers file", err)
	}
	err = gob.NewDecoder(file).Decode(&peers)
	if err != nil {
		log.Panicln("cannot read peers file", err)
	}
	return
}
