package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func (c Client) Offers(filter api.Filter) (offers []api.Offer, err error) {
	// Connect to service
	conn, err := c.query(api.QueryOffers)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send filter
	err = cslq.Encode(conn, "[c]c", filter)
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
	err = cslq.Encode(conn, "c", 0)
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
	// Receive filter
	var filter api.Filter
	err = cslq.Decode(conn, "[c]c", &filter)
	if err != nil {
		srv.Println("Cannot read filter", err)
		return
	}
	// Collect file offers
	var offers []api.Offer
	switch filter {
	case api.FilterIn:
		offers = append(offers, service.Incoming(srv.Core).List()...)
	case api.FilterOut:
		offers = append(offers, service.Outgoing(srv.Core).List()...)
	case api.FilterAll:
		offers = append(offers, service.Incoming(srv.Core).List()...)
		offers = append(offers, service.Outgoing(srv.Core).List()...)
	}
	// Send filtered file offers
	err = json.NewEncoder(conn).Encode(offers)
	if err != nil {
		srv.Println("Cannot send incoming offers", err)
		return
	}
	// Wait for OK
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
	srv.Println("Sent incoming offers")
}
