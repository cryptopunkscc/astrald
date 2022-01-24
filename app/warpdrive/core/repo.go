package core

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io/fs"
	"log"
	"path/filepath"
)

type repository struct {
	storage
	incomingDir string
	outgoingDir string
	peersFile   string
}

func newRepository(storage storage) repository {
	r := repository{
		storage:     storage,
		incomingDir: "incoming",
		outgoingDir: "outgoing",
		peersFile:   "peers",
	}
	r.init()
	return r
}

func (r *repository) init() {
	_ = r.MkDir(r.incomingDir, 0700)
	_ = r.MkDir(r.outgoingDir, 0700)
}

func (r *repository) saveIncoming(offer *api.Offer) {
	path := filepath.Join(r.incomingDir, string(offer.Id))
	r.saveOffer(path, offer)
}

func (r *repository) saveOutgoing(offer *api.Offer) {
	path := filepath.Join(r.outgoingDir, string(offer.Id))
	r.saveOffer(path, offer)
}

func (r *repository) saveOffer(path string, offer *api.Offer) {
	file, err := r.Writer(path, 0700)
	if err != nil {
		log.Panicln("cannot create file for incoming offer", err)
	}
	err = gob.NewEncoder(file).Encode(offer)
	if err != nil {
		log.Panicln("cannot write offer to file", err)
	}
	err = file.Close()
	if err != nil {
		log.Println("cannot close offer file", path, err)
	}
}

func (r *repository) savePeers(peers []api.Peer) {
	file, err := r.Writer(r.peersFile, 0700)
	if err != nil {
		log.Panicln("cannot open peers file", err)
	}
	err = gob.NewEncoder(file).Encode(peers)
	if err != nil {
		log.Panicln("cannot write peers", err)
	}
}

func (r *repository) listIncoming() api.Offers {
	return r.listOffers(r.incomingDir)
}

func (r *repository) listOutgoing() api.Offers {
	return r.listOffers(r.outgoingDir)
}

func (r *repository) listOffers(dir string) (offers api.Offers) {
	dir = r.storage.Absolute(dir)
	offers = make(api.Offers)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := r.Reader(path)
		if err != nil {
			return err
		}
		id := api.OfferId(info.Name())
		offer := &api.Offer{}
		err = gob.NewDecoder(file).Decode(offer)
		if err != nil {
			return err
		}
		offers[id] = offer
		return nil
	})
	if err != nil {
		log.Println("Cannot list incoming offers", err)
		return
	}
	return
}

func (r *repository) listPeers() (peers []api.Peer) {
	file, err := r.Reader(r.peersFile)
	if err != nil {
		if r.IsNotExist(err) {
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
