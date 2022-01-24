package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func (r recipient) Received() (offers api.Offers, err error) {
	// Connect to service
	conn, err := r.query(api.RecReceived)
	if err != nil {
		return
	}
	defer conn.Close()
	// Receive outgoing offers
	err = json.NewDecoder(conn).Decode(&offers)
	if err != nil {
		r.Println("Cannot read outgoing offers", err)
		return
	}
	// Send OK
	err = enc.Write(conn, uint8(0))
	if err != nil {
		r.Println("Cannot send ok", err)
		return
	}
	return
}

func RecipientReceived(srv service.Context, request astral.Request) {
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
	err = json.NewEncoder(conn).Encode(srv.Incoming().List())
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
