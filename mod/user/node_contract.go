package user

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"time"
)

var _ astral.Object = &NodeContract{}

type NodeContract struct {
	UserID        id.Identity
	NodeID        id.Identity
	ExpiresAt     time.Time
	UserSignature []byte
	NodeSignature []byte
}

func (NodeContract) ObjectType() string {
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

func (contract NodeContract) WriteTo(w io.Writer) (n int64, err error) {
	var buf = &bytes.Buffer{}
	err = cslq.Encode(buf, "vvv[c]c[c]c",
		contract.UserID,
		contract.NodeID,
		cslq.Time(contract.ExpiresAt),
		contract.UserSignature,
		contract.NodeSignature,
	)
	if err != nil {
		return
	}
	n2, err := w.Write(buf.Bytes())
	return int64(n2), err
}

func (contract *NodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	var expiresAt cslq.Time

	var c = streams.NewReadCounter(r)

	err = cslq.Decode(c, "vvv[c]c[c]c",
		&contract.UserID,
		&contract.NodeID,
		&expiresAt,
		&contract.UserSignature,
		&contract.NodeSignature,
	)

	contract.ExpiresAt = expiresAt.Time()

	n = c.Total()

	return
}
