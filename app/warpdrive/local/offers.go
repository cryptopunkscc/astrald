package local

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type offersRepo struct {
	dir string
}

func (r *offersRepo) Save(offer *api.Offer) {
	path := filepath.Join(r.dir, string(offer.Id))

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
		log.Println("cannot close offer file", path, err)
	}
}

func (r *offersRepo) List() api.Offers {
	offers := make(api.Offers)
	err := filepath.Walk(r.dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(r.normalizePath(path))
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
		return nil
	}
	return offers
}

func (r *offersRepo) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(r.dir, path)
}
