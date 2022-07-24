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

func (c Client) Status(filter api.Filter) (status <-chan api.OfferStatus, err error) {
	// Connect to local service
	conn, err := c.query(api.QueryStatus)
	if err != nil {
		return
	}
	err = cslq.Encode(conn, "[c]c", filter)
	if err != nil {
		c.Println("Cannot send filter", err)
		conn.Close()
		return
	}
	statChan := make(chan api.OfferStatus)
	status = statChan
	go func(conn io.ReadWriteCloser, status chan api.OfferStatus) {
		defer close(status)
		defer conn.Close()
		dec := json.NewDecoder(conn)
		files := &api.OfferStatus{}
		c.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				c.Println("Cannot decode status", err)
				return
			}
			status <- *files
		}
	}(conn, statChan)
	return
}

func Status(srv handler.Context, request astral.Request) {
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
		unsub := service.Incoming(srv.Core).StatusSubs.SubscribeChan(c)
		defer unsub()
	case api.FilterOut:
		unsub := service.Outgoing(srv.Core).StatusSubs.SubscribeChan(c)
		defer unsub()
	default:
		unsub1 := service.Incoming(srv.Core).StatusSubs.SubscribeChan(c)
		defer unsub1()
		unsub2 := service.Outgoing(srv.Core).StatusSubs.SubscribeChan(c)
		defer unsub2()
	}
	// Wait for close
	var code byte
	err = cslq.Decode(conn, "c", &code)
}
