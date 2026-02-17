package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/user"
)

// SetActiveContract sets the contract under which the node operates
func (mod *Module) SetActiveContract(signed *user.SignedNodeContract) (err error) {
	// verify signatures
	err = mod.VerifySignedNodeContract(signed)
	if err != nil {
		return
	}

	mod.config.ActiveContract.Set(mod.ctx, signed)

	// synchronize siblings & broadcast
	err = mod.Nearby.Broadcast()
	if err != nil {
		mod.log.Error("error broadcasting presence after setting contract: %v", err)
	}

	mod.runSiblingLinker()
	return
}

// ActiveContract returns the active contract
func (mod *Module) ActiveContract() *user.SignedNodeContract {
	return mod.activeContract
}

func (mod *Module) Identity() *astral.Identity {
	ac := mod.ActiveContract()
	if ac == nil {
		return nil
	}
	return ac.UserID
}

func (mod *Module) setActiveContract(signed *user.SignedNodeContract) error {
	// viability checks
	switch {
	case signed.IsNil():
		mod.activeContract = nil
		return nil
	case !signed.ActiveAt(time.Now()):
		return errors.New("contract is not active")
	case !signed.NodeID.IsEqual(mod.node.Identity()):
		return errors.New("local node is not a party to the contract")
	case mod.VerifySignedNodeContract(signed) != nil:
		return errors.New("contract is invalid")
	}

	mod.log.Info("hello, %v!", signed.UserID)
	mod.activeContract = signed
	return nil
}

// ActiveContractsOf returns a list of all active contracts of the specified user
func (mod *Module) ActiveContractsOf(userID *astral.Identity) (contracts []*user.SignedNodeContract, err error) {

	rows, err := mod.db.ActiveContractsOf(userID)
	if err != nil {
		return contracts, err
	}

	var errs []error
	for _, row := range rows {
		contract, err := objects.Load[*user.SignedNodeContract](mod.ctx, mod.Objects.ReadDefault(), row.ObjectID)

		if err != nil {
			errs = append(errs, fmt.Errorf("error loading %s: %w", row.ObjectID.String(), err))
			continue
		}

		contracts = append(contracts, contract)
	}

	err = errors.Join(errs...)

	return contracts, err
}

// IndexSignedNodeContract adds a signed contract to the index
func (mod *Module) IndexSignedNodeContract(signed *user.SignedNodeContract) (err error) {
	// check if active
	if !signed.ActiveAt(time.Now()) {
		return errors.New("contract is not active")
	}

	// verify signatures
	err = mod.VerifySignedNodeContract(signed)
	if err != nil {
		return
	}

	signedID, err := astral.ResolveObjectID(signed)
	if err != nil {
		return
	}

	// check if already indexed
	if mod.db.signedNodeContractExists(signedID) {
		return nil
	}

	// save to db
	err = mod.db.Create(&dbSignedNodeContract{
		ObjectID:  signedID,
		UserID:    signed.UserID,
		NodeID:    signed.NodeID,
		StartsAt:  signed.StartsAt.Time().UTC(),
		ExpiresAt: signed.ExpiresAt.Time().UTC(),
	}).Error
	if err != nil {
		return fmt.Errorf(`db error: %w`, err)
	}

	mod.runSiblingLinker()

	return
}

func (mod *Module) RemoveFromIndex(objectID *astral.ObjectID) error {
	if mod.db.signedNodeContractExists(objectID) {
		return mod.db.deleteSignedNodeContract(objectID)
	}

	return errors.New("object not found in any index")
}

func (mod *Module) GetNodeContract(contractID *astral.ObjectID) (*user.SignedNodeContract, error) {
	// fast fail so we dont need to load the contract if it does not exist in db
	if !mod.db.signedNodeContractExists(contractID) {
		return nil, user.ErrContractNotExists
	}

	return objects.Load[*user.SignedNodeContract](mod.ctx, mod.Objects.ReadDefault(), contractID)
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

func (mod *Module) InviteNode(ctx *astral.Context, nodeID *astral.Identity) (signed *user.SignedNodeContract, err error) {
	// get local userID
	ac := mod.ActiveContract()
	if ac == nil {
		return nil, errors.New("no active contract")
	}
	userID := ac.UserID

	// open a channel to the target
	ch, err := astrald.WithTarget(nodeID).QueryChannel(ctx, user.OpInvite, nil)
	if err != nil {
		return
	}
	defer ch.Close()

	// generate a new contract with the default length
	contract := user.NewNodeContract(userID, nodeID)
	signed = &user.SignedNodeContract{
		NodeContract: contract,
	}

	// send the contract to the target
	err = ch.Send(contract)
	if err != nil {
		return
	}

	// expect node's signature
	err = ch.Switch(channel.Expect(&signed.NodeSig))
	if err != nil {
		return
	}

	// verify node's signature
	err = mod.Crypto.VerifyHashSignature(
		secp256k1.FromIdentity(signed.NodeID),
		signed.NodeSig,
		signed.SignableHash(),
	)
	if err != nil {
		return
	}

	// sign the contract with user's key
	userSigner, err := mod.Crypto.TextObjectSigner(secp256k1.FromIdentity(contract.UserID))
	if err != nil {
		return nil, err
	}

	signed.UserSig, err = userSigner.SignTextObject(ctx, contract)
	if err != nil {
		return nil, err
	}

	// final verification
	err = mod.VerifySignedNodeContract(signed)
	if err != nil {
		return signed, err
	}

	// send the user's signature back
	err = ch.Send(signed.UserSig)
	if err != nil {
		return signed, err
	}

	return signed, nil
}

func (mod *Module) FindNodeContract(contractID *astral.ObjectID) (*user.NodeContract, error) {
	dbRecord, err := mod.db.findSignedNodeContract(contractID)
	if err != nil {
		return nil, err
	}

	return &user.NodeContract{
		UserID:    dbRecord.UserID,
		NodeID:    dbRecord.NodeID,
		StartsAt:  astral.Time(dbRecord.StartsAt),
		ExpiresAt: astral.Time(dbRecord.ExpiresAt),
	}, nil
}
