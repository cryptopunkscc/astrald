package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func (c Client) Peers() (peers []api.Peer, err error) {
	// Connect to local service
	conn, err := c.query(api.QueryPeers)
	if err != nil {
		return
	}
	defer conn.Close()
	// Read peers
	err = json.NewDecoder(conn).Decode(&peers)
	if err != nil {
		c.Println("Cannot read peers", err)
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

func Peers(srv handler.Context, request astral.Request) {
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
	// Get peers
	peers := service.Peer(srv.Core).List()
	// Send peers
	err = json.NewEncoder(conn).Encode(peers)
	if err != nil {
		srv.Println("Cannot send peers", err)
		return
	}
	// Read OK
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
}
