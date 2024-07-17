package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
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
	node   node2.Node
	log    *log.Logger
	assets resources.Resources

	exonet exonet.Module
	dir    dir.Module
	keys   keys.Module
	db     *gorm.DB

	links sig.Set[net.Link]
}

func (mod *Module) Peers() (peers []id.Identity) {
	var r map[string]struct{}

	for _, link := range mod.links.Clone() {
		if _, found := r[link.RemoteIdentity().PublicKeyHex()]; found {
			continue
		}
		r[link.RemoteIdentity().PublicKeyHex()] = struct{}{}
		peers = append(peers, link.RemoteIdentity())
	}

	return
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

func (mod *Module) AcceptLink(ctx context.Context, conn exonet.Conn) (net.Link, error) {
	l, err := muxlink.Accept(ctx, conn, mod.node.Identity(), mod.node.Router())
	if err != nil {
		return nil, err
	}

	err = mod.addLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) InitLink(ctx context.Context, conn exonet.Conn, remoteID id.Identity) (net.Link, error) {
	l, err := muxlink.Open(ctx, conn, remoteID, mod.node.Identity(), mod.node.Router())
	if err != nil {
		return nil, err
	}

	err = mod.addLink(l)
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

	err = mod.addLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) addLink(link net.Link) error {
	link.SetLocalRouter(mod.node.Router())
	err := mod.links.Add(link)
	if err != nil {
		return err
	}

	mod.log.Logv(1, "added link with %v (%s)", link.RemoteIdentity(), exonet.Network(link))

	go func() {
		link.Run(context.Background())
		mod.links.Remove(link)
		mod.log.Logv(1, "removed link with %v (%s)", link.RemoteIdentity(), exonet.Network(link))
	}()

	return nil
}

func (mod *Module) Resolve(ctx context.Context, identity id.Identity) ([]exonet.Endpoint, error) {
	return mod.Endpoints(identity), nil
}

func (mod *Module) Nodes() (nodes []id.Identity) {
	mod.db.Model(&dbEndpoint{}).Distinct("identity").Find(&nodes)
	return
}
