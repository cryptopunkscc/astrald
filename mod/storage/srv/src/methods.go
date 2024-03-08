package srv

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	client "github.com/cryptopunkscc/astrald/lib/storage"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
)

type Context struct {
	storage.Module
	Conn     io.ReadWriteCloser
	RemoteID id.Identity
}

func readAll(ctx Context, req proto.ReadAllReq) (resp proto.ReadAllResp, err error) {
	r := &storage.OpenOpts{
		Offset:  req.Offset,
		Virtual: req.Virtual,
		Network: req.Network,
	}
	if req.Filter != "" {
		var closer io.Closer
		if r.IdentityFilter, closer, err = client.NewClient(ctx.RemoteID).Port(req.Filter).IdFilter(); err != nil {
			return
		}
		defer closer.Close()
	}
	resp.Bytes, err = ctx.ReadAll(req.ID, r)
	return
}

func put(ctx Context, req proto.PutReq) (resp proto.PutResp, err error) {
	i, err := ctx.Put(req.Bytes, req.CreateOpts)
	if err != nil {
		return
	}
	resp.ID = &i
	return
}

func create(ctx Context, req proto.CreateReq) (_ any, err error) {
	creator, err := ctx.Create(req.CreateOpts)
	if err != nil {
		return
	}
	if err = NewWriterService(creator, ctx.Conn).Loop(); err != nil {
		return
	}
	return
}

func purge(ctx Context, req proto.PurgeReq) (resp proto.PurgeResp, err error) {
	resp.Total, err = ctx.Purge(req.ID, req.PurgeOpts)
	return
}

func open(ctx Context, req proto.OpenReq) (_ any, err error) {
	opts := &storage.OpenOpts{
		Offset:  req.Offset,
		Virtual: req.Virtual,
		Network: req.Network,
	}
	// inject id filter client if port was specified
	if req.Filter != "" {
		var closer io.Closer
		if opts.IdentityFilter, closer, err = client.NewClient(ctx.RemoteID).Port(req.Filter).IdFilter(); err != nil {
			return
		}
		defer closer.Close()
	}

	// handle requests in loop
	reader, err := ctx.Open(req.ID, opts)
	if err != nil {
		return
	}
	if err = NewReaderService(reader, ctx.Conn).Loop(); err != nil {
		return
	}
	return
}
