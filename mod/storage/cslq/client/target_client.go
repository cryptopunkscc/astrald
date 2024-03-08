package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

type TargetClient struct {
	target id.Identity
	port   string
	cmd    byte
}

func NewTargetClient(target id.Identity, port string) *TargetClient {
	return &TargetClient{target: target, port: port}
}

func (c *TargetClient) Route(cmd byte) *TargetClient {
	c.cmd = cmd
	return c
}

func (c *TargetClient) Purge(dataID data.ID, opts *storage.PurgeOpts) (i int, err error) {
	conn, err := c.query("v v", dataID, opts)
	if err != nil {
		return
	}
	defer conn.Close()
	if err = cslq.Decode(conn, "l", &i); err != nil {
		return
	}
	return
}

func (c *TargetClient) Open(dataID data.ID, opts *storage.OpenOpts) (r storage.Reader, err error) {
	// register id filtering service if needed
	idFilterPort := ""
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	if opts.IdentityFilter != nil {
		var h Handler
		if h, err = newIdentityFilterService(opts.IdentityFilter); err != nil {
			return
		}
		if err = RegisterHandler(ctx, h); err != nil {
			return
		}
		idFilterPort = h.String()
	}

	// query open
	conn, err := c.query("v v [c]c", dataID, opts, idFilterPort)
	if err != nil {
		return
	}

	// cancel filtering service along with connection
	w := rwClose{conn, func() error {
		cancel()
		return conn.Close()
	}}

	r = newReaderClient(w)
	return
}

func (c *TargetClient) Create(opts *storage.CreateOpts) (w storage.Writer, err error) {
	conn, err := c.query("v", opts)
	if err != nil {
		return
	}
	w = newWriterClient(conn)
	return
}

func (c *TargetClient) idFilter() (filter id.Filter, closer io.Closer, err error) {
	conn, err := c.query("")
	if err != nil {
		return
	}
	filter = func(identity id.Identity) (b bool) {
		if err := cslq.Encode(conn, "v", identity); err != nil {
			return
		}
		if err := cslq.Decode(conn, "c", &b); err != nil {
			return
		}
		return
	}
	closer = conn
	return
}

func (c *TargetClient) query(format string, args ...any) (buffer *astral.Conn, err error) {
	return query(c.target, c.port, c.cmd, format, args)
}
