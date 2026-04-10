package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/auth"
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
		return nil
	case signed.ExpiresAt.Time().Before(time.Now()):
		return errors.New("contract is expired")
	case !signed.Subject.IsEqual(mod.node.Identity()):
		return errors.New("local node is not the subject of the contract")
	}

	mod.log.Info("hello, %v!", signed.Issuer)
	mod.activeContract = signed
	return nil
}

// StoreContract stores the signed contract in the auth module
func (mod *Module) StoreContract(signed *auth.SignedContract) error {
	return mod.Auth.StoreContract(mod.ctx, signed)
}

// ActiveContractsOf returns all active SwarmAccess contracts issued by userID
func (mod *Module) ActiveContractsOf(userID *astral.Identity) ([]*auth.SignedContract, error) {
	contracts, err := mod.Auth.FindContractsWithIssuer(mod.ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []*auth.SignedContract
	for _, c := range contracts {
		if len(c.HasPermit(user.ActionSwarmAccess)) > 0 {
			result = append(result, c)
		}
	}
	return result, nil
}

// ActiveUsers returns all users with an active SwarmAccess contract on the given node
func (mod *Module) ActiveUsers(nodeID *astral.Identity) (users []*astral.Identity) {
	contracts, err := mod.Auth.FindContractsWithActor(mod.ctx, nodeID)
	if err != nil {
		mod.log.Error("db error: %v", err)
		return
	}
	for _, c := range contracts {
		if len(c.HasPermit(user.ActionSwarmAccess)) > 0 {
			users = append(users, c.Issuer)
		}
	}
	return
}

// ActiveNodes returns all nodes with an active SwarmAccess contract from the given user
func (mod *Module) ActiveNodes(userID *astral.Identity) (nodes []*astral.Identity) {
	contracts, err := mod.Auth.FindContractsWithIssuer(mod.ctx, userID)
	if err != nil {
		mod.log.Error("db error: %v", err)
		return
	}

	for _, c := range contracts {
		if len(c.HasPermit(user.ActionSwarmAccess)) > 0 {
			nodes = append(nodes, c.Subject)
		}
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
	permits := astral.NewBundle()
	_ = permits.Append(&auth.Permit{Action: astral.String8(user.ActionSwarmAccess)})
	contract := &auth.Contract{
		Issuer:    userID,
		Subject:   nodeID,
		Permits:   permits,
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

	// sign as issuer: try ASN1 first, fall back to BIP137
	issuerKey := secp256k1.FromIdentity(userID)
	if signer, signerErr := mod.Crypto.ObjectSigner(issuerKey); signerErr == nil {
		signed.IssuerSig, err = signer.SignObject(ctx, contract)
	} else if signer, signerErr := mod.Crypto.TextObjectSigner(issuerKey); signerErr == nil {
		signed.IssuerSig, err = signer.SignTextObject(ctx, contract)
	} else {
		return nil, fmt.Errorf("no signing scheme available for issuer key")
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
