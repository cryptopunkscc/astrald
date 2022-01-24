package handle

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func (s sender) Peers() (peers []api.Peer, err error) {
	// Connect to local service
	conn, err := s.query(api.SenPeers)
	if err != nil {
		return
	}
	defer conn.Close()
	// Read peers
	err = json.NewDecoder(conn).Decode(&peers)
	if err != nil {
		s.Println("Cannot read peers", err)
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

func SenderPeers(srv service.Context, request astral.Request) {
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
	peers := srv.Peer().List()
	// Send peers
	err = json.NewEncoder(conn).Encode(peers)
	if err != nil {
		srv.Println("Cannot send peers", err)
		return
	}
	// Read OK
	_, err = enc.ReadUint8(conn)
	if err != nil {
		srv.Println("Cannot read ok", err)
		return
	}
}
