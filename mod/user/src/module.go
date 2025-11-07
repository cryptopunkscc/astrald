package user

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/kos"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ user.Module = &Module{}

type Module struct {
	Deps
	ctx    *astral.Context
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	db     *DB
	mu     sync.Mutex
	ops    shell.Scope

	linkedSibs sig.Map[string, *astral.Identity]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	ac := mod.ActiveContract()
	if ac != nil {
		mod.log.Info("hello, %v!", ac.UserID)
	}

	mod.Scheduler.Schedule(ctx, mod.NewEnsureConnectivityAction())

	<-ctx.Done()
	return nil
}

// ActiveUsers returns a list of known active users of the specified node
func (mod *Module) ActiveUsers(nodeID *astral.Identity) (users []*astral.Identity) {
	users, err := mod.db.UniqueActiveUsersOnNode(nodeID)
	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return
}

// ActiveNodes returns a list of known active nodes of the specified user
func (mod *Module) ActiveNodes(userID *astral.Identity) (nodes []*astral.Identity) {
	nodes, err := mod.db.UniqueActiveNodesOfUser(userID)
	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return
}

// LocalSwarm returns a list of node identities with an active contract with the current user
func (mod *Module) LocalSwarm() (list []*astral.Identity) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	return mod.ActiveNodes(ac.UserID)
}

func (mod *Module) notifyLinkedSibs(event string) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	for _, sib := range mod.getLinkedSibs() {
		sib := sib
		go mod.Objects.Push(mod.ctx, sib, &user.Notification{Event: astral.String8(event)})
	}
}

func (mod *Module) pushToLinkedSibs(object astral.Object) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	for _, sib := range mod.linkedSibs.Values() {
		sib := sib
		go mod.Objects.Push(mod.ctx, sib, object)
	}
}

func (mod *Module) String() string {
	return user.ModuleName
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) getLinkedSibs() (list []*astral.Identity) {
	return mod.linkedSibs.Values()
}

// NOTE: Legacy methods below are result of lack of universal solution to
// this set of problems

func (mod *Module) SyncAssets(ctx *astral.Context, nodeID *astral.Identity) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return errors.New("no active contract")
	}

	key := fmt.Sprintf("mod.user.sync.%v.next_height", nodeID.String())
	var args any
	height, _ := kos.Get[*astral.Uint64](ctx, mod.KOS, key)
	if height != nil {
		args = opSyncAssetsArgs{Start: int(*height)}
	}

	var q = query.New(ac.UserID, nodeID, user.OpSyncAssets, args)
	conn, err := query.Route(ctx, mod.node, q)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := astral.NewChannel(conn)
	ch.Blueprints = astral.NewBlueprints(ch.Blueprints)
	ch.Blueprints.Add(&OpUpdate{})

	for {
		msg, err := ch.Read()
		if err != nil {
			mod.log.Error("SyncAssets: error reading from channel: %v", err)
			return err
		}

		switch m := msg.(type) {
		case *OpUpdate:
			if m.Removed {
				err = mod.db.RemoveAssetByNonce(m.Nonce, m.ObjectID)
			} else {
				err = mod.db.AddAssetWithNonce(m.ObjectID, m.Nonce)
			}
			if err != nil {
				mod.log.Error("SyncAssets: error syncing asset: %v", err)
				return err
			}

		case *astral.Uint64:
			return mod.KOS.Set(ctx, key, m)

		default:
			mod.log.Error("SyncAssets: protocol error: unknown msg: %v", m.ObjectType())
			return err
		}
	}
}

func (mod *Module) SyncAlias(ctx *astral.Context, nodeID *astral.Identity) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return errors.New("no active contract")
	}

	var q = query.New(ac.UserID, nodeID, user.OpInfo, nil)
	conn, err := query.Route(ctx, mod.node, q)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch := astral.NewChannel(conn)
	ch.Blueprints.Add(&user.Info{})

	obj, err := ch.Read()
	if err != nil {
		return err
	}

	info, ok := obj.(*user.Info)
	if !ok {
		return fmt.Errorf("protocol error: invalid object type %s", obj.ObjectType())
	}

	if len(info.NodeAlias) == 0 {
		return nil
	}

	if mod.Dir.DisplayName(ac.UserID) == "" {
		mod.Dir.SetAlias(ac.UserID, string(info.UserAlias))
	}

	mod.log.Info("SyncAlias: updating %v alias %v", nodeID, info.NodeAlias)

	return mod.Dir.SetAlias(nodeID, string(info.NodeAlias))
}

func (mod *Module) SyncApps(ctx *astral.Context, nodeID *astral.Identity) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return errors.New("no active contract")
	}

	contracts, err := mod.Apphost.ActiveLocalAppContracts()
	if err != nil {
		return err
	}

	for _, contract := range contracts {
		mod.Objects.Push(ctx, nodeID, contract)
	}

	return
}
