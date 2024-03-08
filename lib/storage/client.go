package storage

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/storage"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
)

type Client struct {
	target   id.Identity
	port     string
	query    queryFunc
	response responseFunc
}

type queryFunc func(req proto.Request) (io.ReadWriteCloser, error)
type responseFunc func(closer io.ReadWriteCloser, resp any) error

func NewClient(target id.Identity) (c *Client) {
	c = &Client{target: target, port: proto.Port}
	c.query = c.jsonQuery
	c.response = jsonResponse
	return
}

func (c *Client) Port(port string) *Client {
	c.port = port
	return c
}

func (c *Client) ReadAll(dataID data.ID, opts *storage.OpenOpts) (b []byte, err error) {
	req := proto.ReadAllReq{
		ID:      dataID,
		Offset:  opts.Offset,
		Virtual: opts.Virtual,
		Network: opts.Network,
	}

	// register id filtering service if needed
	if opts.IdentityFilter != nil {
		var h Handler
		if h, err = NewIdentityFilterService(opts.IdentityFilter); err != nil {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err = RegisterHandler(ctx, h); err != nil {
			return
		}
		req.Filter = h.String()
	}

	resp := proto.ReadAllResp{Bytes: b}
	err = c.request(req, &resp)
	b = resp.Bytes
	return
}

func (c *Client) Put(b []byte, opts *storage.CreateOpts) (i data.ID, err error) {
	req := proto.PutReq{Bytes: b, CreateOpts: opts}
	resp := proto.PutResp{}
	if err = c.request(req, &resp); err != nil {
		return
	}
	i = *resp.ID
	return
}

func (c *Client) Purge(dataID data.ID, opts *storage.PurgeOpts) (i int, err error) {
	req := proto.PurgeReq{ID: dataID, PurgeOpts: opts}
	resp := proto.PurgeResp{}
	err = c.request(req, &resp)
	i = resp.Total
	return
}

func (c *Client) Open(dataID data.ID, opts *storage.OpenOpts) (r storage.Reader, err error) {
	req := proto.OpenReq{
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
		var h Handler
		if h, err = NewIdentityFilterService(opts.IdentityFilter); err != nil {
			return
		}
		if err = RegisterHandler(ctx, h); err != nil {
			return
		}
		req.Filter = h.String()
	}

	// query open
	conn, err := c.query(req)
	if err != nil {
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
	req := proto.CreateReq{CreateOpts: opts}
	conn, err := c.query(req)
	if err != nil {
		return
	}
	w = NewWriterClient(conn)
	return
}

func (c *Client) request(req proto.Request, resp error) (err error) {
	conn, err := c.query(req)
	if err != nil {
		return
	}
	defer conn.Close()
	if err = c.response(conn, resp); err == nil {
		if resp.Error() != "" {
			err = resp
		}
	}
	return
}

func (c *Client) jsonQuery(req proto.Request) (conn io.ReadWriteCloser, err error) {
	query := c.port
	if req != nil {
		var bytes []byte
		if bytes, err = json.Marshal(req); err != nil {
			return
		}
		query = query + req.Query() + "?" + string(bytes)
	}
	conn, err = astral.Query(c.target, query)
	return
}

func jsonResponse(conn io.ReadWriteCloser, resp any) (err error) {
	return json.NewDecoder(conn).Decode(resp)
}
