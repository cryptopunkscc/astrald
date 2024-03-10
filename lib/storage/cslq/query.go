package storage

import (
	"bytes"
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

var BaseEncoding = base64.RawStdEncoding

func query(target id.Identity, port string, cmd byte, format string, args []any) (conn *astral.Conn, err error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.WriteString(port)
	e := base64.NewEncoder(BaseEncoding, buffer)
	if cmd != 0 {
		if err = cslq.Encode(e, "c", cmd); err != nil {
			return
		}
	}
	if format != "" {
		if err = cslq.Encode(e, format, args...); err != nil {
			return
		}
	}
	if err = e.Close(); err != nil {
		return
	}
	return astral.Query(target, buffer.String())
}
