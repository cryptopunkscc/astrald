package storage

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

type PurgerService struct {
	purger storage.Purger
}

func NewPurgerService(purger storage.Purger) *PurgerService {
	return &PurgerService{purger: purger}
}

func (s *PurgerService) Handle(conn io.ReadWriter, args *cslq.Decoder) (err error) {
	var dataID data.ID
	var opts storage.PurgeOpts
	if err = args.Decodef("v v", &dataID, &opts); err != nil {
		return
	}
	i, err := s.purger.Purge(dataID, &opts)
	if err != nil {
		return
	}
	return cslq.Encode(conn, "l", i)
}
