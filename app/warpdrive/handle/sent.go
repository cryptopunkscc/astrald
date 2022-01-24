package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func (s sender) Sent() (offers api.Offers, err error) {
	// Connect to service
	conn, err := s.query(api.SenSent)
	if err != nil {
		return
	}
	defer conn.Close()
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		s.Println("Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		s.Println("Cannot send ok", err)
		return
	}
	return
}

func SenderSent(srv service.Context, request astral.Request) {
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
	// Send outgoing files
	err = json.NewEncoder(conn).Encode(srv.Outgoing().List())
	if err != nil {
		srv.Println("Cannot send outgoing offers", err)
		return
	}
	// Wait for OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
	srv.Println("Send outgoing offers")
}
