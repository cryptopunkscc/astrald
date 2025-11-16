package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kos"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

// FIXME: react on contract (spawn maintain)

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
	mod.runSiblingLinker()
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
			mod.Objects.Root(),
			row.ObjectID,
			mod.Objects.Blueprints(),
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

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	objects.Save(ctx, c, mod.Objects.Root())

	err = mod.db.Create(&dbNodeContract{
		ObjectID:  contractID,
		UserID:    c.UserID,
		NodeID:    c.NodeID,
		ExpiresAt: c.ExpiresAt.Time().UTC(),
	}).Error
	if err != nil {
		return
	}

	mod.runSiblingLinker()

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
