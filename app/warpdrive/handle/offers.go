package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
)

func (r recipient) Offers() (offers <-chan api.Offer, err error) {
	// Connect to local service
	conn, err := r.query(api.RecIncoming)
	if err != nil {
		return
	}
	ofs := make(chan api.Offer)
	offers = ofs
	go func(conn io.ReadWriteCloser, offers chan api.Offer) {
		defer close(offers)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &api.Offer{}
		r.Println("Start listening offers")
		for {
			err := dec.Decode(files)
			if err != nil {
				r.Println("Finish listening offers", err)
				return
			}
			offers <- *files
		}
	}(conn, ofs)
	return
}

func RecipientOffers(srv service.Context, request astral.Request) {
	if srv.IsRejected(request) {
		return
	}
	// Accept connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept request", err)
		return
	}
	defer conn.Close()
	remove := srv.FilesOffers().Subscribe(conn)
	defer remove()
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}
