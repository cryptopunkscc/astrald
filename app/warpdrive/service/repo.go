package warpdrive

import (
	"encoding/gob"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type repository struct {
	incomingDir string
	outgoingDir string
	peersFile   string
}

func newRepository() repository {
	dir := warpdriveDir()
	return repository{
		incomingDir: filepath.Join(dir, "incoming"),
		outgoingDir: filepath.Join(dir, "outgoing"),
		peersFile:   filepath.Join(dir, "peers"),
	}
}

func (r repository) init() {
	_ = os.MkdirAll(r.incomingDir, 0700)
	_ = os.MkdirAll(r.outgoingDir, 0700)
}

func warpdriveDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error fetching user's config dir:", err)
		os.Exit(0)
	}

	dir := filepath.Join(cfgDir, "warpdrive")
	os.MkdirAll(dir, 0700)

	return dir
}

func (r repository) saveIncoming(offer *Offer) {
	path := filepath.Join(r.incomingDir, string(offer.Id))
	saveOffer(path, offer)
}

func (r repository) saveOutgoing(offer *Offer) {
	path := filepath.Join(r.outgoingDir, string(offer.Id))
	saveOffer(path, offer)
}

func saveOffer(path string, offer *Offer) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		log.Panicln("cannot create file for incoming offer", err)
	}
	err = gob.NewEncoder(file).Encode(offer)
	if err != nil {
		log.Panicln("cannot write offer to file", err)
	}
	err = file.Close()
	if err != nil {
		log.Println("cannot close offer file", file.Name(), err)
	}
}

func (r repository) savePeers(peers []Peer) {
	file, err := os.OpenFile(r.peersFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		log.Panicln("cannot open peers file", err)
	}
	err = gob.NewEncoder(file).Encode(peers)
	if err != nil {
		log.Panicln("cannot write peers", err)
	}
}

func (r repository) listIncoming() Offers {
	return listOffers(r.incomingDir)
}

func (r repository) listOutgoing() Offers {
	return listOffers(r.outgoingDir)
}

func listOffers(dir string) (offers Offers) {
	offers = make(Offers)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		id := OfferId(info.Name())
		offer := &Offer{}
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

func (r repository) listPeers() (peers []Peer) {
	file, err := os.Open(r.peersFile)
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
