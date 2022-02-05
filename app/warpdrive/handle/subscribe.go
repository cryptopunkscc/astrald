package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
)

func (c Client) Subscribe(filter api.Filter) (offers <-chan api.Offer, err error) {
	// Connect to local service
	conn, err := c.query(api.QuerySubscribe)
	if err != nil {
		return
	}
	err = enc.WriteL8String(conn, string(filter))
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
	f, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read filter", err)
		return
	}
	c := api.WriteChannel(conn)
	defer close(c)
	switch api.Filter(f) {
	case api.FilterIn:
		unsubIn := service.Incoming(srv.Core).Offers().SubscribeChan(c)
		defer unsubIn()
	case api.FilterOut:
		unsubOut := service.Outgoing(srv.Core).Offers().SubscribeChan(c)
		defer unsubOut()
	default:
		unsubIn := service.Incoming(srv.Core).Offers().SubscribeChan(c)
		defer unsubIn()
		unsubOut := service.Outgoing(srv.Core).Offers().SubscribeChan(c)
		defer unsubOut()
	}
	// Wait for close
	_, _ = enc.ReadUint8(conn)
}
