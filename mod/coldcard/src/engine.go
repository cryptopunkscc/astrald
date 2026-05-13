package coldcard

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/coldcard"
	"github.com/cryptopunkscc/astrald/mod/coldcard/ckcc"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	dcrdSecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Engine implements TextSignerFactory for hardware signing via Coldcard devices.
type Engine struct {
	mod *Module
}

// --- TextSignerFactory ---

func (e *Engine) NewTextSigner(ctx *astral.Context, key *crypto.PrivateKey, scheme string) (crypto.TextSigner, error) {
	pubKey := &crypto.PublicKey{
		Type: secp256k1.KeyType,
		Key:  dcrdSecp.PrivKeyFromBytes(key.Key).PubKey().SerializeCompressed(),
	}

	pubKeyHex := hex.EncodeToString(pubKey.Key)

	device := e.mod.deviceForPublicKeyHex(pubKeyHex)
	if device != nil {
		return &MessageSigner{dev: device, path: coldcard.BIP44Path}, nil
	}

	return nil, crypto.ErrUnsupported
}

type MessageSigner struct {
	dev  *ckcc.Device
	path string
}

func (m *MessageSigner) SignText(ctx *astral.Context, msg string) (*crypto.Signature, error) {
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
