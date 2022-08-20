package notify

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"log"
)

type Client struct{ Identity id.Identity }

func (c Client) Create(channel Channel) (err error) {
	conn, err := astral.Dial(c.Identity, PortChannel)
	if err != nil {
		return
	}
	err = json.NewEncoder(conn).Encode(channel)
	if err != nil {
		return
	}
	var code byte
	err = cslq.Decode(conn, "c", &code)
	return
}

func (c Client) Notify(notifications ...Notification) (err error) {
	conn, err := astral.Dial(c.Identity, Port)
	if err != nil {
		return
	}
	err = json.NewEncoder(conn).Encode(notifications)
	if err != nil {
		return
	}
	var code byte
	err = cslq.Decode(conn, "c", &code)
	return
}

func (c Client) Notifier() (dispatch Notify) {
	conn, err := astral.Dial(c.Identity, Port)
	if err != nil {
		return
	}
	enc := json.NewEncoder(conn)
	nc := make(chan []Notification, 128)
	dispatch = nc
	go func() {
		defer conn.Close()
		for notifications := range nc {
			if err == nil {
				err = enc.Encode(notifications)
				if err != nil {
					log.Println("cannot encode notification", err)
					continue
				}
				var code byte
				err = cslq.Decode(conn, "c", &code)
				if err != nil {
					log.Println("notifier connection lost", err)
					continue
				}
			}
		}
	}()
	return
}
