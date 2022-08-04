package contacts

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
)

type Client struct{ wrapper.Api }

func (c Client) List() (peers []Contact, err error) {
	conn, err := c.Query(id.Identity{}, Port)
	if err != nil {
		return
	}
	err = json.NewDecoder(conn).Decode(&peers)
	return
}
