package notify

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/legacy/enc"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"log"
)

var _ Api = Client{}

type Client struct{ Identity id.Identity }

func (c Client) Create(channel Channel) (err error) {
	conn, err := astral.Dial(c.Identity, createChannel)
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
	conn, err := astral.Dial(c.Identity, notify)
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

func (c Client) Notifier() (dispatch Notify) {
	conn, err := astral.Dial(c.Identity, notify)
	if err != nil {
		return
	}
	nc := make(chan []Notification, 128)
	dispatch = nc
	go func() {
		defer conn.Close()
		encoder := gob.NewEncoder(conn)
		for notifications := range nc {
			if err == nil {
				err = encoder.Encode(notifications)
				if err != nil {
					log.Println("cannot encode notification", err)
					continue
				}
				_, err = enc.ReadUint8(conn)
				if err != nil {
					log.Println("notifier connection lost", err)
					continue
				}
			}
		}
	}()
	return
}
