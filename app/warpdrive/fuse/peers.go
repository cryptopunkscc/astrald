package fuse

import (
	"context"
	app "github.com/cryptopunkscc/astrald/services/apphost/client"
	warp "github.com/cryptopunkscc/astrald/services/warpdrive/client"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type peersDir struct {
	fs.Inode
}

var _ fs.NodeOnAdder = new(peersDir)

func (r *peersDir) OnAdd(ctx context.Context) {
	warpClient := warp.NewWarpDriveClient(ctx, app.NewCoreAdapter())
	peers, err := warpClient.ListPeers()
	if err != nil {
		panic(err)
	}
	for _, peer := range peers {
		rd := &peerDir{name: peer, client: warpClient}
		attr := fs.StableAttr{Mode: fuse.S_IFDIR}
		inode := r.NewInode(ctx, rd, attr)
		_ = r.AddChild(peer, inode, true)
	}
}
