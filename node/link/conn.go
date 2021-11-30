package link

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/sig"
	"time"
)

type Conn struct {
	activity sig.Activity

	*link.Conn
}

func wrapConn(linkConn *link.Conn) *Conn {
	c := &Conn{Conn: linkConn}
	c.activity.Touch()
	return c
}

func (conn *Conn) Idle() time.Duration {
	return conn.activity.Idle()
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	defer conn.activity.Touch()

	return conn.Conn.Read(p)
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	defer conn.activity.Touch()

	return conn.Conn.Write(p)
}
