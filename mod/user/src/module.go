package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/kos"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"slices"
	"sync"
	"time"
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
	sibs   sig.Map[string, context.CancelFunc]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	ac := mod.ActiveContract()
	if ac != nil {
		mod.log.Info("hello, %v!", ac.UserID)
	}

	mod.syncSibs()

	<-ctx.Done()
	return nil
}

// ActiveNodes returns a list of known active nodes of the specified user
func (mod *Module) ActiveNodes(userID *astral.Identity) (nodes []*astral.Identity) {
	nodes, err := mod.db.UniqueActiveNodesOfUser(userID)
	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return
}

// ActiveUsers returns a list of known active users of the specified node
func (mod *Module) ActiveUsers(nodeID *astral.Identity) (users []*astral.Identity) {
	users, err := mod.db.UniqueActiveUsersOnNode(nodeID)
	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return
}

// SetActiveContract sets the contract under which the node operates
func (mod *Module) SetActiveContract(contract *user.SignedNodeContract) (err error) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	// check if the contract meets necessary criteria
	if !contract.NodeID.IsEqual(mod.node.Identity()) {
		return errors.New("local node is not a party to the contract")
	}

	if time.Now().After(contract.ExpiresAt.Time()) {
		return errors.New("contract expired")
	}

	err = mod.Validate(contract)
	if err != nil {
		return
	}

	// store the contract
	err = mod.KOS.Set(mod.ctx, keyActiveContract, contract)
	if err != nil {
		return
	}

	mod.log.Info("hello, %v!", contract.UserID)

	// synchronize siblings
	mod.mu.Unlock()
	mod.syncSibs()
	mod.mu.Lock()

	return
}

// Validate checks if the contract has valid signatures from both the user and the node.
func (mod *Module) Validate(contract *user.SignedNodeContract) error {
	if contract.UserID.IsZero() {
		return errors.New("invalid contract: UserID is zero")
	}

	if contract.NodeID.IsZero() {
		return errors.New("invalid contract: NodeID is zero")
	}

	if contract.UserID.IsEqual(contract.NodeID) {
		return errors.New("invalid contract: UserID and NodeID are equal")
	}

	var hash = contract.Hash()

	// verify user signature
	err := mod.Keys.VerifyASN1(contract.UserID, hash, contract.UserSig)
	if err != nil {
		return fmt.Errorf("user sig verification: %w", err)
	}

	// verify node signature
	err = mod.Keys.VerifyASN1(contract.NodeID, hash, contract.NodeSig)
	if err != nil {
		return fmt.Errorf("node sig verification: %w", err)
	}

	return nil
}

// ActiveContract returns the active contract
func (mod *Module) ActiveContract() *user.SignedNodeContract {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	contract, err := kos.Get[*user.SignedNodeContract](mod.ctx, mod.KOS, keyActiveContract)
	if err != nil {
		return nil
	}

	if contract.IsExpired() {
		return nil
	}

	return contract
}

// ActiveContractsOf returns a list of all active contracts of the specified user
func (mod *Module) ActiveContractsOf(userID *astral.Identity) (contracts []*user.SignedNodeContract, err error) {
	rows, err := mod.db.ActiveContractsOf(userID)
	if err != nil {
		return
	}

	var errs []error
	for _, row := range rows {
		contract, err := objects.Load[*user.SignedNodeContract](
			mod.ctx,
			mod.Objects,
			row.ObjectID,
			astral.DefaultScope(),
		)

		if err != nil {
			errs = append(errs, fmt.Errorf("error loading %s: %w", row.ObjectID.String(), err))
			continue
		}

		contracts = append(contracts, contract)
	}
	err = errors.Join(errs...)

	return
}

// SetUserID looks for a signed contract between the specified user and the local node and sets it as active
func (mod *Module) SetUserID(userID *astral.Identity) error {
	contracts, err := mod.ActiveContractsOf(userID)
	if len(contracts) == 0 {
		return err
	}

	var best *user.SignedNodeContract

	for _, contract := range contracts {
		if !contract.NodeID.IsEqual(mod.node.Identity()) {
			continue
		}

		if best == nil {
			best = contract
			continue
		}

		if best.ExpiresAt.Time().Before(contract.ExpiresAt.Time()) {
			best = contract
		}
	}

	bestID, _ := astral.ResolveObjectID(best)

	mod.log.Log("using contract %v for user %v", bestID, userID)

	return mod.SetActiveContract(best)
}

// SaveSignedNodeContract validates and persists a signed node contract
func (mod *Module) SaveSignedNodeContract(c *user.SignedNodeContract) (err error) {
	contractID, err := astral.ResolveObjectID(c)
	if err != nil {
		return
	}

	// check if already saved
	if mod.db.ContractExists(contractID) {
		return nil
	}

	if c.IsExpired() {
		return errors.New("contract expired")
	}

	err = mod.Validate(c)
	if err != nil {
		return err
	}

	mod.Objects.Store(c)

	err = mod.db.Create(&dbNodeContract{
		ObjectID:  contractID,
		UserID:    c.UserID,
		NodeID:    c.NodeID,
		ExpiresAt: c.ExpiresAt.Time().UTC(),
	}).Error
	if err != nil {
		return
	}

	// synchronize siblings with contracts
	mod.syncSibs()

	return
}

// SignLocalContract creates, signs and stores a new node contract with the specified user
func (mod *Module) SignLocalContract(userID *astral.Identity) (contract *user.SignedNodeContract, err error) {
	// then create and sign a new contract
	contract = &user.SignedNodeContract{
		NodeContract: &user.NodeContract{
			UserID:    userID,
			NodeID:    mod.node.Identity(),
			ExpiresAt: astral.Time(time.Now().Add(defaultContractValidity).UTC()),
		},
	}

	// sign with node key
	contract.NodeSig, err = mod.Keys.SignASN1(contract.NodeID, contract.Hash())
	if err != nil {
		return
	}

	// sign with user key
	contract.UserSig, err = mod.Keys.SignASN1(contract.UserID, contract.Hash())
	if err != nil {
		return
	}

	err = mod.SaveSignedNodeContract(contract)

	return
}

// AddAsset adds an object to user's assets
func (mod *Module) AddAsset(objectID *object.ID) (err error) {
	_, err = mod.db.AddAsset(objectID, false)
	return
}

// RemoveAsset removes an object from user's assets
func (mod *Module) RemoveAsset(objectID *object.ID) (err error) {
	return mod.db.RemoveAsset(objectID)
}

// AssetsContain returns true if user's assets contain the object
func (mod *Module) AssetsContain(objectID *object.ID) bool {
	return mod.db.AssetsContain(objectID)
}

func (mod *Module) Assets() []*object.ID {
	assets, err := mod.db.Assets()
	if err != nil {
		mod.log.Error("error getting assets: %v", err)
	}

	return assets
}

func (mod *Module) String() string {
	return user.ModuleName
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

// siblings (other nodes of the same user)

func (mod *Module) addSib(nodeID *astral.Identity) error {
	if nodeID.IsEqual(mod.node.Identity()) {
		return errors.New("cannot add self")
	}

	ctx, cancel := context.WithCancel(mod.ctx)

	cancel, ok := mod.sibs.Set(nodeID.String(), cancel)
	if !ok {
		return errors.New("already added")
	}

	go func() {
		mod.linkSib(ctx, nodeID)
		mod.removeSib(nodeID)
	}()

	return nil
}

func (mod *Module) removeSib(nodeID *astral.Identity) error {
	cancel, found := mod.sibs.Delete(nodeID.String())
	if !found {
		return errors.New("not found")
	}

	cancel()

	return nil
}

func (mod *Module) listSibs() (list []*astral.Identity) {
	for _, idStr := range mod.sibs.Keys() {
		if id, err := astral.IdentityFromString(idStr); err == nil {
			list = append(list, id)
		}
	}
	return
}

func (mod *Module) linkSib(ctx context.Context, nodeID *astral.Identity) {
	mod.log.Info("added sibling %v", nodeID)
	defer mod.log.Info("removed sibling %v", nodeID)
	var count = 0
	for {
		// context check
		select {
		case <-ctx.Done():
			return
		default:
		}

		ac := mod.ActiveContract()
		if ac == nil {
			return
		}

		conn, err := query.Route(ctx, mod.node, query.New(
			ac.UserID,
			nodeID,
			"user.link",
			nil,
		))

		if err != nil {
			delay := min((1<<count)*time.Second, 15*time.Minute)
			count = min(count+1, 32)

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		}

		count = 0

		done := make(chan struct{})

		go func() {
			select {
			case <-ctx.Done():
				conn.Close()
			case <-done:
			}
		}()

		mod.log.Info("linked with %v", nodeID)
		io.Copy(io.Discard, conn)
		mod.log.Log("link with %v lost", nodeID)

		close(done)
	}
}

func (mod *Module) syncSibs() {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	sibs := mod.listSibs()
	nodes := mod.ActiveNodes(ac.UserID)

	// remove siblings that are not on our nodes list
	for _, sib := range sibs {
		if !slices.ContainsFunc(nodes, sib.IsEqual) {
			mod.removeSib(sib)
		}
	}

	// add siblings that are missing
	for _, node := range nodes {
		if !slices.ContainsFunc(sibs, node.IsEqual) {
			mod.addSib(node)
		}
	}
}
