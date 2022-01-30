package content

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
)

const (
	read = "sys/content"
	info = "sys/content/info"
)

var _ Api = Client{}

type Client struct{ Identity string }

func (c Client) Reader(uri string) (reader io.ReadCloser, err error) {
	conn, err := astral.Query(c.Identity, read)
	err = enc.WriteL8String(conn, uri)
	if err != nil {
		return
	}
	_, err = enc.ReadUint8(conn)
	if err != nil {
		return nil, err
	}
	reader = conn
	return
}

func (c Client) Info(uri string) (files []Info, err error) {
	conn, err := astral.Query(c.Identity, info)
	err = enc.WriteL8String(conn, uri)
	if err != nil {
		return
	}
	err = json.NewDecoder(conn).Decode(&files)
	if err != nil {
		return
	}
	_ = enc.Write(conn, 0)
	return
}
