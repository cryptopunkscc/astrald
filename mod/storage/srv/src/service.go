package srv

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	client "github.com/cryptopunkscc/astrald/lib/storage"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/storage/srv"
	"github.com/cryptopunkscc/astrald/node"
	jrpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"io"
	"log"
)

func RunStorageRpc(ctx context.Context, node node.Node, mod storage.Module) error {
	rpc := jrpc.NewModule(node, proto.Port)
	rpc.Interface(NewService(mod))
	return rpc.Routes(
		"*",
		"readAll*",
		"put*",
		"create*",
		"purge*",
		"open*",
	).Run(ctx)
}

type Service struct {
	mod storage.Module
}

func NewService(mod storage.Module) *Service {
	return &Service{mod: mod}
}

func (c *Service) ReadAllAuth() string { return storage.OpenAction }
func (c *Service) ReadAll(remoteID id.Identity, req proto.ReadAll) (resp []byte, err error) {
	r := &storage.OpenOpts{
		Offset:  req.Offset,
		Virtual: req.Virtual,
		Network: req.Network,
	}
	if req.Filter != "" {
		var closer io.Closer
		var conn jrpc.Conn
		conn, err = jrpc.QueryFlow(remoteID, req.Filter)
		if err != nil {
			return
		}
		if r.IdentityFilter, closer, err = client.NewClient(conn).IdFilter(); err != nil {
			return
		}
		defer closer.Close()
	}
	resp, err = c.mod.ReadAll(req.ID, r)
	return
}

func (c *Service) PutAuth() string { return storage.CreateAction }
func (c *Service) Put(req proto.Put) (resp data.ID, err error) {
	return c.mod.Put(req.Bytes, req.CreateOpts)
}

func (c *Service) CreateAuth() string { return storage.CreateAction }
func (c *Service) Create(conn jrpc.Conn, req proto.Create) (err error) {
	creator, err := c.mod.Create(req.CreateOpts)
	if err != nil {
		return
	}
	if err = NewWriterService(creator, conn).Loop(); err != nil {
		return
	}
	return
}

func (c *Service) PurgeAuth() string { return storage.PurgeAction }
func (c *Service) Purge(req proto.Purge) (total int, err error) {
	return c.mod.Purge(req.ID, req.PurgeOpts)
}

func (c *Service) OpenAuth() string { return storage.OpenAction }
func (c *Service) Open(remoteID id.Identity, conn jrpc.Conn, req proto.Open) (err error) {
	opts := &storage.OpenOpts{
		Offset:  req.Offset,
		Virtual: req.Virtual,
		Network: req.Network,
	}
	// inject id filter client if port was specified
	if req.Filter != "" {
		var closer io.Closer
		rpc := conn
		rpc, err = jrpc.QueryFlow(remoteID, req.Filter)
		if err != nil {
			return
		}
		rpc.Logger(log.New(log.Writer(), "filter client ", 0))
		if opts.IdentityFilter, closer, err = client.NewClient(rpc).IdFilter(); err != nil {
			return
		}
		defer closer.Close()
	}

	// handle requests in loop
	reader, err := c.mod.Open(req.ID, opts)
	defer reader.Close()
	if err != nil {
		return
	}
	if err = NewReaderService(reader, conn).Loop(); err != nil {
		return
	}
	return
}
