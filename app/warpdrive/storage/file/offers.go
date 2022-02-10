package file

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var _ api.OfferStorage = Offers{}

type Offers struct {
	api.Core
	dir string
}

func Incoming(core api.Core) Offers {
	return Offers{core, "incoming"}
}

func Outgoing(core api.Core) Offers {
	return Offers{core, "outgoing"}
}

func (r Offers) Init() {
	_ = os.MkdirAll(r.normalizePath(""), 0700)
}

func (r Offers) Save(offer *api.Offer) {
	path := r.normalizePath(string(offer.Id))

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		r.Panicln("cannot create file for incoming offer", err)
	}
	err = gob.NewEncoder(file).Encode(offer)
	if err != nil {
		r.Panicln("cannot write offer to file", err)
	}
	err = file.Close()
	if err != nil {
		r.Println("cannot close offer file", path, err)
	}
}

func (r Offers) Get() api.Offers {
	offers := make(api.Offers)
	dir := r.normalizePath("")
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		normalizedPath := r.normalizePath(path)
		file, err := os.Open(normalizedPath)
		if err != nil {
			r.Println("cannot open", err)
			return nil
		}
		id := api.OfferId(info.Name())
		offer := &api.Offer{}
		err = gob.NewDecoder(file).Decode(offer)
		if err != nil {
			r.Println("cannot decode", path, err)
			return nil
		}
		offers[id] = offer
		return nil
	})
	if err != nil {
		r.Println("Cannot list incoming offers", err)
	}
	return offers
}

func (r Offers) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(r.RepositoryDir, r.dir, path)
}
