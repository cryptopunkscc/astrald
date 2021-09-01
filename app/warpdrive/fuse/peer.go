package fuse

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/warpdrive/client"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"os"
	"syscall"
)

type peerDir struct {
	fs.Inode
	client *client.WarpDriveClient
	name   string
}

var _ fs.NodeCreater = new(peerDir)

func (p *peerDir) Create(
	ctx context.Context,
	name string,
	flags uint32,
	mode uint32,
	out *fuse.EntryOut,
) (
	node *fs.Inode,
	fh fs.FileHandle,
	fuseFlags uint32,
	errno syscall.Errno,
) {
	writer, err := p.client.Writer(api.Identity(p.name), name)
	if err != nil {
		return nil, nil, 0, syscall.EACCES
	}
	s := &fileSender{name: name, writer: writer}
	attr := fs.StableAttr{Mode: uint32(os.O_CREATE | os.O_WRONLY | os.O_APPEND)}
	node = p.NewInode(ctx, s, attr)
	p.AddChild(name, node, true)
	fh = s
	return
}
