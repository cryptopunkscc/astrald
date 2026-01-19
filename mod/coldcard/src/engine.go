package coldcard

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/coldcard/ckcc"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type Engine struct {
	mod *Module
	crypto.NilEngine
}

func (e *Engine) MessageSigner(key *crypto.PublicKey, scheme string) (crypto.MessageSigner, error) {
	switch {
	case scheme != "bip137":
		return nil, crypto.ErrUnsupportedScheme
	case key.Type != "secp256k1":
		return nil, crypto.ErrUnsupportedKeyType
	}

	pubKeyHex := hex.EncodeToString(key.Key)

	device := e.mod.deviceForPublicKeyHex(pubKeyHex)
	if device != nil {
		return &MessageSigner{dev: device, path: ""}, nil
	}

	return nil, crypto.ErrUnsupported
}

type MessageSigner struct {
	dev  *ckcc.Device
	path string
}

func (m *MessageSigner) SignMessage(ctx *astral.Context, msg string) (*crypto.Signature, error) {
	sigBase64, err := m.dev.Msg(msg, m.path)
	if err != nil {
		return nil, err
	}

	sig, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return nil, err
	}

	return &crypto.Signature{
		Scheme: "bip137",
		Data:   sig,
	}, nil
}
