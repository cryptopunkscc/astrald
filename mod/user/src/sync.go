package user

import (
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/mod/user"
)

// NOTE: Legacy methods below are result of lack of universal solution to
// this set of problems

func (mod *Module) syncAssets(ctx *astral.Context, nodeID *astral.Identity) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return errors.New("no active contract")
	}

	nodePath := fmt.Sprintf("/mod/user/assets/%v/next_height", nodeID.String())
	var args any

	heightNode, err := tree.Query(ctx, mod.Tree.Root(), nodePath, true)
	if err != nil {
		return err
	}

	height, _ := tree.Get[*astral.Uint64](ctx, heightNode)
	if height != nil {
		args = opSyncAssetsArgs{Start: int(*height)}
	}

	var q = query.New(ac.UserID, nodeID, user.OpSyncAssets, args)
	conn, err := query.Route(ctx, mod.node, q)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := channel.New(conn)

	for {
		msg, err := ch.Receive()
		if err != nil {
			mod.log.Error("syncAssets: error reading from channel: %v", err)
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
				mod.log.Error("syncAssets: error syncing asset: %v", err)
				return err
			}

		case *astral.Uint64:
			return heightNode.Set(ctx, m)

		default:
			mod.log.Error("syncAssets: protocol error: unknown msg: %v", m.ObjectType())
			return err
		}
	}
}

func (mod *Module) syncAlias(ctx *astral.Context, nodeID *astral.Identity) (err error) {
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
	ch := channel.New(conn)

	obj, err := ch.Receive()
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

	mod.log.Info("syncAlias: updating %v alias %v", nodeID, info.NodeAlias)

	return mod.Dir.SetAlias(nodeID, string(info.NodeAlias))
}

func (mod *Module) syncApps(ctx *astral.Context, nodeID *astral.Identity) (err error) {
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

func (mod *Module) syncSiblings(ctx *astral.Context, with *astral.Identity) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return errors.New("no active contract")
	}

	contracts, err := mod.ActiveContractsOf(ac.UserID)
	if err != nil {
		return err
	}

	for _, contract := range contracts {
		if contract.NodeID.IsEqual(mod.node.Identity()) {
			continue
		}
		if contract.NodeID.IsEqual(with) {
			continue
		}
		mod.Objects.Push(ctx, with, contract)
	}

	return nil
}
