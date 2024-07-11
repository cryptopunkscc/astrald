package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/tasks"
	"github.com/jxskiss/base62"
	"gorm.io/gorm"
	"strings"
	"time"
)

const DefaultWorkerCount = 8
const DefaultTimeout = time.Minute
const infoPrefix = "node1"

type NodeInfo nodes.NodeInfo

var _ nodes.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	assets resources.Resources

	dir  dir.Module
	keys keys.Module
	db   *gorm.DB
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Service{Module: mod},
	).Run(ctx)
}

func (mod *Module) ParseInfo(s string) (*nodes.NodeInfo, error) {
	trimmed := strings.TrimPrefix(s, infoPrefix)
	data, err := base62.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}

	info, err := (&InfoEncoder{mod}).Unpack(data)
	if err != nil {
		return nil, err
	}

	return (*nodes.NodeInfo)(info), nil
}

func (mod *Module) InfoString(info *nodes.NodeInfo) string {
	packed, err := (&InfoEncoder{mod}).Pack((*NodeInfo)(info))
	if err != nil {
		return ""
	}

	return infoPrefix + base62.EncodeToString(packed)
}

func (mod *Module) AcceptLink(ctx context.Context, conn net.Conn) (net.Link, error) {
	l, err := muxlink.Accept(ctx, conn, mod.node.Identity(), mod.node.LocalRouter())
	if err != nil {
		return nil, err
	}

	err = mod.node.Network().AddLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) InitLink(ctx context.Context, conn net.Conn, remoteID id.Identity) (net.Link, error) {
	l, err := muxlink.Open(ctx, conn, remoteID, mod.node.Identity(), mod.node.LocalRouter())
	if err != nil {
		return nil, err
	}

	err = mod.node.Network().AddLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) Link(ctx context.Context, remoteIdentity id.Identity, opts nodes.LinkOpts) (net.Link, error) {
	l, err := (&Linker{mod}).LinkOpts(ctx, remoteIdentity, opts)
	if err != nil {
		return nil, err
	}

	err = mod.node.Network().AddLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) Resolve(ctx context.Context, identity id.Identity, opts *nodes.ResolveOpts) ([]net.Endpoint, error) {
	return mod.Endpoints(identity), nil
}

func (mod *Module) Nodes() (nodes []id.Identity) {
	mod.db.Model(&dbEndpoint{}).Distinct("identity").Find(&nodes)
	return
}
