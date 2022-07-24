package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func (c Client) Update(
	peerId api.PeerId,
	attr string,
	value string,
) (err error) {
	// Connect to local service
	conn, err := c.query(api.QueryUpdate)
	if err != nil {
		return
	}
	defer conn.Close()
	// Send peers to update
	req := []string{string(peerId), attr, value}
	err = json.NewEncoder(conn).Encode(req)
	if err != nil {
		c.Println("Cannot send peer update", err)
		return
	}
	// Wait for OK
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		c.Println("Cannot read ok", err)
		return
	}
	return
}

func Update(srv handler.Context, request astral.Request) {
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
	// Read peer update
	var req []string
	err = json.NewDecoder(conn).Decode(&req)
	if err != nil {
		srv.Println("Cannot read peer update", err)
		return
	}
	peerId := req[0]
	attr := req[1]
	value := req[2]
	// Update peer
	service.Peer(srv.Core).Update(peerId, attr, value)
	// Send OK
	err = cslq.Encode(conn, "c", 0)
	if err != nil {
		srv.Println("Cannot send ok", err)
		return
	}
}
