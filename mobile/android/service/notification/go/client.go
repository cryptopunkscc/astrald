package notify

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

var _ Api = Client{}

type Client struct{ Identity string }

func (c Client) Create(channel Channel) (err error) {
	conn, err := astral.Query(c.Identity, createChannel)
	if err != nil {
		return
	}
	err = gob.NewEncoder(conn).Encode(channel)
	if err != nil {
		return
	}
	_, err = enc.ReadUint8(conn)
	return
}

func (c Client) Notify(notifications ...Notification) (err error) {
	conn, err := astral.Query(c.Identity, notify)
	if err != nil {
		return
	}
	err = gob.NewEncoder(conn).Encode(notifications)
	if err != nil {
		return
	}
	_, err = enc.ReadUint8(conn)
	return
}
