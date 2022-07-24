package content

import (
	"encoding/gob"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
)

var _ Api = Client{}

type Client struct{ Identity id.Identity }

func (c Client) Info(uri string) (files Info, err error) {
	conn, err := astral.Dial(c.Identity, info)
	if err != nil {
		return
	}
	err = cslq.Encode(conn, "[c]c", uri)
	if err != nil {
		return
	}
	err = gob.NewDecoder(conn).Decode(&files)
	if err != nil {
		return
	}
	_ = cslq.Encode(conn, "c", 0)
	return
}

func (c Client) Reader(uri string) (reader io.ReadCloser, err error) {
	conn, err := astral.Dial(c.Identity, content)
	if err != nil {
		return
	}
	err = cslq.Encode(conn, "[c]c", uri)
	if err != nil {
		return
	}
	var code byte
	err = cslq.Decode(conn, "c", &code)
	if err != nil {
		return
	}
	reader = conn
	return
}
