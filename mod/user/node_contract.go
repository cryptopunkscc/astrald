package user

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"time"
)

type NodeContract struct {
	UserID        id.Identity
	NodeID        id.Identity
	ExpiresAt     time.Time
	UserSignature []byte
	NodeSignature []byte
}

func (*NodeContract) ObjectType() string {
	return "mod.users.node_contract"
}

func (contract *NodeContract) Hash() []byte {
	var hash = sha256.New()
	var err = cslq.Encode(hash,
		"[c]cvvv",
		contract.ObjectType(),
		contract.UserID,
		contract.NodeID,
		cslq.Time(contract.ExpiresAt),
	)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
}

func (contract *NodeContract) IsExpired() bool {
	return time.Now().After(contract.ExpiresAt)
}

func (contract *NodeContract) IsActive() bool {
	if contract.IsExpired() {
		return false
	}
	if contract.Validate() != nil {
		return false
	}
	return true
}

func (contract *NodeContract) Validate() (err error) {
	if contract.ExpiresAt.IsZero() {
		return errors.New("expiry time missing")
	}

	err = contract.VerifyUserSignature()
	if err != nil {
		return
	}

	err = contract.VerifyNodeSignature()

	return
}

func (contract *NodeContract) VerifyUserSignature() error {
	switch {
	case contract.UserID.IsZero():
		return errors.New("user identity missing")
	case contract.UserSignature == nil:
		return errors.New("user signature missing")
	case !ecdsa.VerifyASN1(
		contract.UserID.PublicKey().ToECDSA(),
		contract.Hash(),
		contract.UserSignature,
	):
		return errors.New("user signature is invalid")
	}

	return nil
}

func (contract *NodeContract) VerifyNodeSignature() error {
	switch {
	case contract.NodeID.IsZero():
		return errors.New("node identity missing")
	case contract.NodeSignature == nil:
		return errors.New("node signature missing")
	case !ecdsa.VerifyASN1(
		contract.NodeID.PublicKey().ToECDSA(),
		contract.Hash(),
		contract.NodeSignature,
	):
		return errors.New("node signature is invalid")
	}

	return nil
}

func (contract NodeContract) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("vvv[c]c[c]c",
		contract.UserID,
		contract.NodeID,
		cslq.Time(contract.ExpiresAt),
		contract.UserSignature,
		contract.NodeSignature,
	)
}

func (contract *NodeContract) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var expiresAt cslq.Time
	err := dec.Decodef("vvv[c]c[c]c",
		&contract.UserID,
		&contract.NodeID,
		&expiresAt,
		&contract.UserSignature,
		&contract.NodeSignature,
	)
	contract.ExpiresAt = expiresAt.Time()
	return err
}
