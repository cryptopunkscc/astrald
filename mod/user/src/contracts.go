package user

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/user"
)

// SetActiveContract sets the contract under which the node operates
func (mod *Module) SetActiveContract(signed *auth.SignedContract) (err error) {
	if err = mod.Auth.VerifyContract(signed); err != nil {
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
func (mod *Module) ActiveContract() *auth.SignedContract {
	return mod.activeContract
}

func (mod *Module) Identity() *astral.Identity {
	ac := mod.ActiveContract()
	if ac == nil {
		return nil
	}
	return ac.Issuer
}

func (mod *Module) setActiveContract(signed *auth.SignedContract) error {
	switch {
	case signed.IsNil():
		mod.activeContract = nil
		go mod.Nearby.SetMode(mod.ctx, nearby.ModeVisible)
		return nil
	case signed.ExpiresAt.Time().Before(time.Now()):
		return errors.New("contract is expired")
	case !signed.Subject.IsEqual(mod.node.Identity()):
		return errors.New("local node is not the subject of the contract")
	}

	mod.log.Info("hello, %v!", signed.Issuer)
	mod.activeContract = signed
	mod.Nearby.Broadcast()
	return nil
}

// ActiveContractsOf returns all active SwarmAccess contracts issued by userID
func (mod *Module) ActiveContractsOf(userID *astral.Identity) ([]*auth.SignedContract, error) {
	return mod.Auth.FindContracts(mod.ctx, auth.ContractQuery{
		IssuerID: userID,
		Action:   user.ActionSwarmAccess,
		Active:   true,
	})
}

// StoreContract stores the signed contract in the auth module
func (mod *Module) StoreContract(signed *auth.SignedContract) error {
	return mod.Auth.StoreContract(mod.ctx, signed)
}

// ActiveUsers returns all users with an active SwarmAccess contract on the given node
func (mod *Module) ActiveUsers(nodeID *astral.Identity) (users []*astral.Identity) {
	contracts, err := mod.Auth.FindContracts(mod.ctx, auth.ContractQuery{
		SubjectID: nodeID,
		Action:    user.ActionSwarmAccess,
		Active:    true,
	})
	if err != nil {
		mod.log.Error("db error: %v", err)
		return
	}
	for _, c := range contracts {
		users = append(users, c.Issuer)
	}
	return
}

// ActiveNodes returns all nodes with an active SwarmAccess contract from the given user
func (mod *Module) ActiveNodes(userID *astral.Identity) (nodes []*astral.Identity) {
	contracts, err := mod.Auth.FindContracts(mod.ctx, auth.ContractQuery{
		IssuerID: userID,
		Action:   user.ActionSwarmAccess,
		Active:   true,
	})
	if err != nil {
		mod.log.Error("db error: %v", err)
		return
	}
	for _, c := range contracts {
		nodes = append(nodes, c.Subject)
	}
	return
}

// LocalSwarm returns a list of node identities with an active SwarmAccess contract with the current user
func (mod *Module) LocalSwarm() (list []*astral.Identity) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	return mod.ActiveNodes(ac.Issuer)
}

func (mod *Module) InviteNode(ctx *astral.Context, nodeID *astral.Identity) (signed *auth.SignedContract, err error) {
	// get local userID
	ac := mod.ActiveContract()
	if ac == nil {
		return nil, errors.New("no active contract")
	}
	userID := ac.Issuer

	// open a channel to the target
	ch, err := astrald.WithTarget(nodeID).QueryChannel(ctx, user.OpInvite, nil)
	if err != nil {
		return
	}
	defer ch.Close()

	// build the contract
	contract := &auth.Contract{
		Issuer:    userID,
		Subject:   nodeID,
		Permits:   []auth.Permit{{Action: astral.String8(user.ActionSwarmAccess)}},
		ExpiresAt: astral.Time(time.Now().Add(defaultContractValidity)),
	}

	signed = &auth.SignedContract{Contract: contract}

	// send the contract to the target
	err = ch.Send(contract)
	if err != nil {
		return
	}

	// expect subject (node) signature
	err = ch.Switch(channel.Expect(&signed.SubjecSig))
	if err != nil {
		return
	}

	// verify subject signature
	err = mod.Crypto.VerifyObjectSignature(
		secp256k1.FromIdentity(nodeID),
		signed.SubjecSig,
		contract,
	)
	if err != nil {
		return nil, fmt.Errorf("subject sig verification: %w", err)
	}

	// sign as issuer: prefer BIP137 if available
	issuerKey := secp256k1.FromIdentity(userID)
	schemes := mod.Crypto.AvailableSchemes(issuerKey)
	if slices.Contains(schemes, crypto.SchemeBIP137) {
		signer, signerErr := mod.Crypto.TextObjectSigner(issuerKey)
		if signerErr != nil {
			return nil, signerErr
		}
		signed.IssuerSig, err = signer.SignTextObject(ctx, contract)
	} else {
		signer, signerErr := mod.Crypto.ObjectSigner(issuerKey)
		if signerErr != nil {
			return nil, signerErr
		}
		signed.IssuerSig, err = signer.SignObject(ctx, contract)
	}
	if err != nil {
		return nil, fmt.Errorf("sign as issuer: %w", err)
	}

	// send the issuer's signature back
	err = ch.Send(signed.IssuerSig)
	if err != nil {
		return signed, err
	}

	return signed, nil
}

func (mod *Module) RemoveFromIndex(objectID *astral.ObjectID) error {
	return errors.New("not supported: use auth.Unban instead")
}
