package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	jrpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

type Client struct {
	conn jrpc.Conn
}

func NewClient(conn jrpc.Conn) *Client {
	return &Client{conn: conn}
}

func (c *Client) ReadAll(dataID data.ID, opts *storage.OpenOpts) (b []byte, err error) {
	req := proto.ReadAll{
		ID:      dataID,
		Offset:  opts.Offset,
		Virtual: opts.Virtual,
		Network: opts.Network,
	}

	// register id filtering service if needed
	if opts.IdentityFilter != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if req.Filter, err = RunIdentityFilterService(ctx, opts.IdentityFilter); err != nil {
			return
		}
	}

	b, err = jrpc.Query[[]byte](c.conn, "readAll", req)
	return
}

func (c *Client) Put(b []byte, opts *storage.CreateOpts) (i data.ID, err error) {
	req := proto.Put{Bytes: b, CreateOpts: opts}
	i, err = jrpc.Query[data.ID](c.conn, "put", req)
	return
}

func (c *Client) Purge(dataID data.ID, opts *storage.PurgeOpts) (i int, err error) {
	req := proto.Purge{ID: dataID, PurgeOpts: opts}
	i, err = jrpc.Query[int](c.conn, "purge", req)
	return
}

func (c *Client) Open(dataID data.ID, opts *storage.OpenOpts) (r storage.Reader, err error) {
	req := proto.Open{
		ID:      dataID,
		Offset:  opts.Offset,
		Virtual: opts.Virtual,
		Network: opts.Network,
	}

	// register id filtering service if needed
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	if opts.IdentityFilter != nil {
		if req.Filter, err = RunIdentityFilterService(ctx, opts.IdentityFilter); err != nil {
			return
		}
	}

	// query open
	conn := c.conn.Copy()
	if err = jrpc.Call(conn, "open", req); err != nil {
		return
	}

	// cancel filtering service along with connection
	w := rwClose{conn, func() error {
		cancel()
		return conn.Close()
	}}

	r = NewReaderClient(w)
	return
}

func (c *Client) Create(opts *storage.CreateOpts) (w storage.Writer, err error) {
	req := proto.Create{CreateOpts: opts}
	if err = jrpc.Call(c.conn, "create", req); err != nil {
		return
	}
	w = NewWriterClient(c.conn)
	return
}
