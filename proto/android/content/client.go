package content

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"io"
)

type Client struct {
	wrapper.Api
	Identity id.Identity
}

func (c Client) Info(uri string) (files Info, err error) {
	conn, err := c.Query(c.Identity, PortInfo)
	if err != nil {
		return
	}
	err = cslq.Encode(conn, "[c]c", uri)
	if err != nil {
		return
	}
	err = json.NewDecoder(conn).Decode(&files)
	if err != nil {
		return
	}
	_ = cslq.Encode(conn, "c", 0)
	return
}

func (c Client) Reader(uri string) (reader io.ReadCloser, err error) {
	conn, err := c.Query(c.Identity, Port)
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
	reader = fileReader{conn}
	return
}

type fileReader struct {
	io.ReadWriteCloser
}

func (f fileReader) Close() error {
	_ = cslq.Encode(f, "c", 0)
	return f.ReadWriteCloser.Close()
}
