package user

import (
	"crypto/ecdsa"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

var _ astral.Object = &SignedNodeContract{}

type SignedNodeContract struct {
	*NodeContract
	UserSig []byte
	NodeSig []byte
}

func (SignedNodeContract) ObjectType() string {
	return "mod.users.signed_node_contract"
}

func (c *SignedNodeContract) VerifySigs() (err error) {
	if c.ExpiresAt.Time().IsZero() {
		return errors.New("expiry time missing")
	}

	err = c.VerifyUserSig()
	if err != nil {
		return
	}

	err = c.VerifyNodeSig()

	return
}

func (c *SignedNodeContract) VerifyUserSig() error {
	switch {
	case c.UserID.IsZero():
		return errors.New("user identity missing")
	case c.UserSig == nil:
		return errors.New("user signature missing")
	case !ecdsa.VerifyASN1(
		c.UserID.PublicKey().ToECDSA(),
		c.Hash(),
		c.UserSig,
	):
		return errors.New("user signature is invalid")
	}

	return nil
}

func (c *SignedNodeContract) VerifyNodeSig() error {
	switch {
	case c.NodeID.IsZero():
		return errors.New("node identity missing")
	case c.NodeSig == nil:
		return errors.New("node signature missing")
	case !ecdsa.VerifyASN1(
		c.NodeID.PublicKey().ToECDSA(),
		c.Hash(),
		c.NodeSig,
	):
		return errors.New("node signature is invalid")
	}

	return nil
}

func (c *SignedNodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	c.NodeContract = &NodeContract{}
	n, err = c.NodeContract.ReadFrom(r)
	if err != nil {
		return
	}

	err = cslq.Decode(r, "[c]c[c]c", &c.UserSig, &c.NodeSig)
	if err != nil {
		n += int64(len(c.UserSig) + len(c.NodeSig) + 2)
	}

	return
}

func (c SignedNodeContract) WriteTo(w io.Writer) (n int64, err error) {
	n, err = c.NodeContract.WriteTo(w)
	if err != nil {
		return
	}

	err = cslq.Encode(w, "[c]c[c]c", c.UserSig, c.NodeSig)
	if err != nil {
		n += int64(len(c.UserSig) + len(c.NodeSig) + 2)
	}

	return
}
