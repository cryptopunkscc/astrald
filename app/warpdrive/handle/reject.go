package handle

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func (r Recipient) Reject(id api.OfferId) (err error) {
	// Connect to local service
	conn, err := r.query(api.RecReject)
	if err != nil {
		return
	}
	// Send accepted request id to service
	defer conn.Close()
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		r.Println("Cannot send rejected request id", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		r.Println("Cannot read ok", err)
		return
	}
	return
}

func RecipientReject(srv handler.Context, request astral.Request) {
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
	id, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read request id", err)
		return
	}
	err = reject(srv, api.OfferId(id))
	if err != nil {
		srv.Println("Cannot reject incoming files", id, err)
		return
	}
	err = enc.Write(conn, uint8(0))
	if err != nil {
		srv.Println("Cannot send ok", err)
		return
	}
	srv.Println("Rejected incoming files", id)
}

func reject(srv handler.Context, id api.OfferId) (err error) {
	srv.LogPrefix("<", api.Reject)
	// Get cached incoming files by id
	offer := service.Incoming(srv.Core).Get()[id]
	if offer == nil {
		err = errors.New(fmt.Sprint("No incoming file with id ", id))
		srv.Println("Cannot find incoming file", err)
		return
	}
	// Connect to service
	conn, err := srv.Query(string(offer.Peer), api.Reject)
	if err != nil {
		srv.Println("Cannot connect", err)
		return
	}
	defer conn.Close()
	// Send rejected offer id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		srv.Println("Cannot send rejected offer id", id, err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", id, err)
		return
	}
	service.Incoming(srv.Core).Update(offer, "rejected", true)
	return
}

func ServiceReject(srv handler.Context, request astral.Request) {
	// Accept incoming connection
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept connection from", request.Caller(), err)
		return
	}
	defer conn.Close()
	// Read id of rejected outgoing files
	offerId, err := enc.ReadL8String(conn)
	if err != nil {
		srv.Println("Cannot read request id", request.Caller(), err)
		return
	}
	// Reject outgoing files
	offer := service.Outgoing(srv.Core).Get()[api.OfferId(offerId)]
	if offer != nil {
		service.Outgoing(srv.Core).Update(offer, "rejected", true)
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		srv.Println("Cannot send ok", request.Caller(), err)
		return
	}
}
