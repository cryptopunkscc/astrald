package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
)

func (c Client) Subscribe(filter api.Filter) (offers <-chan api.Offer, err error) {
	// Connect to local service
	conn, err := c.query(api.QuerySubscribe)
	if err != nil {
		return
	}
	err = cslq.Encode(conn, "[c]c", filter)
	if err != nil {
		c.Println("Cannot send filter", err)
		conn.Close()
		return
	}
	ofs := make(chan api.Offer)
	offers = ofs
	go func(conn io.ReadWriteCloser, offers chan api.Offer) {
		defer close(offers)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &api.Offer{}
		c.Println("Start listening offers")
		for {
			err := dec.Decode(files)
			if err != nil {
				c.Println("Finish listening offers", err)
				return
			}
			offers <- *files
		}
	}(conn, ofs)
	return
}

func Subscribe(srv handler.Context, request astral.Request) {
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
	var filter api.Filter
	err = cslq.Decode(conn, "[c]c", &filter)
	if err != nil {
		srv.Println("Cannot read filter", err)
		return
	}
	c := api.WriteChannel(conn)
	defer close(c)
	switch filter {
	case api.FilterIn:
		unsubIn := service.Incoming(srv.Core).OfferSubs.SubscribeChan(c)
		defer unsubIn()
	case api.FilterOut:
		unsubOut := service.Outgoing(srv.Core).OfferSubs.SubscribeChan(c)
		defer unsubOut()
	default:
		unsubIn := service.Incoming(srv.Core).OfferSubs.SubscribeChan(c)
		defer unsubIn()
		unsubOut := service.Outgoing(srv.Core).OfferSubs.SubscribeChan(c)
		defer unsubOut()
	}
	// Wait for close
	var code byte
	err = cslq.Decode(conn, "c", &code)
}
