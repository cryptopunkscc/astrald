package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/user"
	userClient "github.com/cryptopunkscc/astrald/mod/user/client"
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
		return auth.ErrContractExpired
	case !signed.Subject.IsEqual(mod.node.Identity()):
		return errors.New("local node is not the subject of the contract")
	}

	mod.log.Info("hello, %v!", signed.Issuer)
	mod.activeContract = signed
	mod.Nearby.Broadcast()
	return nil
}

// ActiveNodeContracts returns all active SwarmAccess contracts issued by userID
func (mod *Module) ActiveNodeContracts(userID *astral.Identity) ([]*auth.SignedContract, error) {
	return mod.Auth.SignedContracts().WithIssuer(userID).WithAction(&user.SwarmAccessAction{}).Find(mod.ctx)
}

// ActiveNodes returns all nodes with an active SwarmAccess contract from the given user
func (mod *Module) ActiveNodes(userID *astral.Identity) (nodes []*astral.Identity) {
	contracts, err := mod.Auth.
		SignedContracts().
		WithIssuer(userID).
		WithAction(&user.SwarmAccessAction{}).
		Find(mod.ctx)

	if err != nil {
		mod.log.Error("error getting active nodes: %v", err)
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
	ac := mod.ActiveContract()
	if ac == nil {
		return nil, errors.New("no active contract")
	}

	contract, err := user.NewNodeContract(ac.Issuer, nodeID, defaultContractValidity)
	if err != nil {
		return nil, err
	}

	signed = &auth.SignedContract{Contract: contract}

	issuerSig, err := mod.Auth.SignIssuer(ctx, signed)
	if err != nil {
		return nil, fmt.Errorf("sign as issuer: %w", err)
	}

	subjectSig, err := userClient.New(nodeID, nil).Invite(ctx, contract, issuerSig)
	if err != nil {
		return nil, err
	}

	err = mod.Crypto.VerifyObjectSignature(
		secp256k1.FromIdentity(nodeID),
		subjectSig,
		contract,
	)
	if err != nil {
		return nil, fmt.Errorf("subject sig verification: %w", err)
	}

	signed.IssuerSig = issuerSig
	signed.SubjecSig = subjectSig
	return signed, nil
}
