package handle

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func (s Sender) Status(id api.OfferId) (status string, err error) {
	// Connect to service
	conn, err := s.query(api.SenStatus)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send request id
	err = enc.WriteL8String(conn, string(id))
	if err != nil {
		s.Println("Cannot send request id", err)
		return
	}
	// Receive status
	status, err = enc.ReadL8String(conn)
	if err != nil {
		s.Println("Cannot read request status", err)
	}
	return
}

func SenderStatus(srv handler.Context, request astral.Request) {
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
	files := service.Outgoing(srv.Core).Get()[api.OfferId(id)]
	if files == nil {
		srv.Println("Cannot find outgoing files with id", id)
		return
	}
	err = enc.WriteL8String(conn, files.Status.Status)
	if err != nil {
		srv.Println("Cannot send file status", files.Status, err)
		return
	}
	srv.Println("Send file status", files.Status, err)
}
