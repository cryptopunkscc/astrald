package link

import (
	"github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"time"
)

type BasicConn interface {
	io.ReadWriteCloser
	Link() *Link
	Query() string
	Outbound() bool
}

type Conn struct {
	*link.Conn

	link     *Link
	activity sig.Activity
}

func wrapConn(lconn *link.Conn) *Conn {
	c := &Conn{Conn: lconn}
	c.activity.Touch()
	go func() {
		<-c.Wait()
		if c.link != nil {
			c.link.events.Emit(EventConnClosed{c})
			c.link.remove(c)
		}
	}()
	return c
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	defer conn.activity.Touch()

	return conn.Conn.Read(p)
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	defer conn.activity.Touch()

	return conn.Conn.Write(p)
}

func (conn *Conn) Link() *Link {
	return conn.link
}

func (conn *Conn) Idle() time.Duration {
	return conn.activity.Idle()
}

func (conn *Conn) Wait() <-chan struct{} {
	return conn.Conn.Wait()
}

func (conn *Conn) Attach(link *Link) {
	if conn.link != nil {
		conn.link.remove(conn)
		conn.link = nil
	}
	if link != nil {
		conn.link = link
		conn.link.add(conn)
	}
}
