package notify

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"log"
)

const (
	createChannel = "sys/notify/channel"
	notify        = "sys/notify"
)

var _ Api = Client{}

type Client struct {
	Identity string
}

func (c Client) Create(channel Channel) (err error) {
	log.Println("Creating channel", channel)
	conn, err := astral.Query(c.Identity, createChannel)
	if err != nil {
		return
	}
	bytes, err := json.Marshal(channel)
	if err != nil {
		return
	}
	err = enc.WriteL16Bytes(conn, bytes)
	if err != nil {
		return
	}
	_, err = enc.ReadUint8(conn)
	if err != nil {
		return
	}
	return
}

func (c Client) Notify(notifications ...Notification) (err error) {
	conn, err := astral.Query(c.Identity, notify)
	if err != nil {
		return
	}
	bytes, err := json.Marshal(notifications)
	if err != nil {
		return
	}
	err = enc.WriteL16Bytes(conn, bytes)
	if err != nil {
		return
	}
	_, err = enc.ReadUint8(conn)
	if err != nil {
		return
	}
	return
}
