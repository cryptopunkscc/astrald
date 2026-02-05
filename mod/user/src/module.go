package user

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/kos"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
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
	ops    ops.Set

	activeContract *user.SignedNodeContract

	sibs sig.Map[string, Sibling]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	<-mod.Scheduler.Ready()

	if userID := mod.Identity(); userID != nil {
		mod.log.Info("hello, %v!", userID)
	}

	mod.runSiblingLinker()
	<-ctx.Done()

	return nil
}

func (mod *Module) SignNodeContract(ctx *astral.Context, contract *user.NodeContract) (*user.SignedNodeContract, error) {
	// node signs the hash of the contract
	nodeSigner, err := mod.Crypto.ObjectSigner(secp256k1.FromIdentity(contract.NodeID))
	if err != nil {
		return nil, fmt.Errorf("sign as node: %w", err)
	}

	nodeSig, err := nodeSigner.SignObject(ctx, contract)
	if err != nil {
		return nil, fmt.Errorf("sign as node: %w", err)
	}

	// user signs the text of the contract
	userSigner, err := mod.Crypto.TextObjectSigner(secp256k1.FromIdentity(contract.UserID))
	if err != nil {
		return nil, fmt.Errorf("sign as user: %w", err)
	}

	userSig, err := userSigner.SignTextObject(ctx, contract)
	if err != nil {
		return nil, fmt.Errorf("sign as user: %w", err)
	}

	return &user.SignedNodeContract{
		NodeContract: contract,
		UserSig:      userSig,
		NodeSig:      nodeSig,
	}, nil
}

func (mod *Module) VerifySignedNodeContract(signed *user.SignedNodeContract) error {
	switch {
	case signed.UserSig == nil:
		return errors.New("user signature is missing")
	case signed.NodeSig == nil:
		return errors.New("node signature is missing")
	case signed.NodeSig.Scheme != crypto.SchemeASN1:
		return errors.New("node signature scheme is not supported")
	case !slices.Contains([]string{
		crypto.SchemeASN1,
		crypto.SchemeBIP137,
	}, signed.UserSig.Scheme.String()):
		return errors.New("user signature scheme is not supported")
	}

	// verify node signature (always hash)
	err := mod.Crypto.VerifyObjectSignature(
		secp256k1.FromIdentity(signed.NodeID),
		signed.NodeSig,
		signed.NodeContract,
	)
	if err != nil {
		return fmt.Errorf("node sig verification: %w", err)
	}

	// verify user signature
	switch signed.UserSig.Scheme {
	case crypto.SchemeASN1:
		// verify user signature via hash
		err = mod.Crypto.VerifyObjectSignature(
			secp256k1.FromIdentity(signed.UserID),
			signed.UserSig,
			signed.NodeContract,
		)

	case crypto.SchemeBIP137:
		// verify user signature via text
		err = mod.Crypto.VerityTextObjectSignature(
			secp256k1.FromIdentity(signed.UserID),
			signed.UserSig,
			signed.NodeContract,
		)

	default:
		err = fmt.Errorf("signature scheme %s is not supported", signed.UserSig.Scheme)
	}
	if err != nil {
		return err
	}

	return nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) TextObjectSigner() crypto.TextObjectSigner {
	signer, _ := mod.Crypto.TextObjectSigner(secp256k1.FromIdentity(mod.Identity()))
	return signer
}

func (mod *Module) String() string {
	return user.ModuleName
}

func (mod *Module) runSiblingLinker() {
	for _, node := range mod.LocalSwarm() {
		if node.IsEqual(mod.node.Identity()) {
			continue
		}

		maintainLinkAction := mod.NewMaintainLinkAction(node)
		scheduledAction, err := mod.Scheduler.Schedule(maintainLinkAction)
		if err != nil {
			mod.log.Error("error scheduling maintain link action: %v for node %v", err, node)
			continue
		}

		mod.addSibling(node, scheduledAction.Cancel)
	}
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

	ch := channel.New(conn)

	for {
		msg, err := ch.Receive()
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

func (mod *Module) GetSwarmJoinRequestPolicy() user.SwarmJoinRequestPolicy {
	return mod.SwarmJoinRequestAcceptAll
}

var _ user.SwarmJoinRequestPolicy = (*Module)(nil).SwarmJoinRequestAcceptAll

func (mod *Module) SwarmJoinRequestAcceptAll(requester *astral.Identity) bool {
	mod.log.Info("Accepting %v join request into swarm", requester)
	return true
}

func (mod *Module) GetSwarmInvitePolicy() user.SwarmInvitePolicy {
	return mod.SwarmInviteAcceptAll
}

var _ user.SwarmInvitePolicy = (*Module)(nil).SwarmInviteAcceptAll

func (mod *Module) SwarmInviteAcceptAll(inviter *astral.Identity, contract user.NodeContract) bool {
	mod.log.Info("Accepting invitation from %v for %v join swarm till %v", inviter, contract.NodeID, contract.ExpiresAt)
	return true
}
