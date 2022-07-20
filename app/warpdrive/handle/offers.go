package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/legacy/enc"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func (c Client) Offers(filter api.Filter) (offers []api.Offer, err error) {
	// Connect to service
	conn, err := c.query(api.QueryOffers)
	if err != nil {
		return
	}
	defer conn.Close()
	err = enc.WriteL8String(conn, string(filter))
	if err != nil {
		c.Println("Cannot send filter", err)
		return
	}
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		c.Println("Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		c.Println("Cannot send ok", err)
		return
	}
	return
}

func Offers(srv handler.Context, request astral.Request) {
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
	f, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read filter", err)
		return
	}
	var offers []api.Offer
	switch api.Filter(f) {
	case api.FilterIn:
		offers = append(offers, service.Incoming(srv.Core).List()...)
	case api.FilterOut:
		offers = append(offers, service.Outgoing(srv.Core).List()...)
	case api.FilterAll:
		offers = append(offers, service.Incoming(srv.Core).List()...)
		offers = append(offers, service.Outgoing(srv.Core).List()...)
	}
	// Send outgoing files
	err = json.NewEncoder(conn).Encode(offers)
	if err != nil {
		srv.Println("Cannot send incoming offers", err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
	srv.Println("Send incoming offers")
}
