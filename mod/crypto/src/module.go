package crypto

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
	Dir     dir.Module
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	db     *DB
	scope  ops.Set
	ctx    *astral.Context

	engines sig.Set[crypto.Engine]
}

var _ crypto.Module = &Module{}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	for _, repoName := range mod.config.Repos {
		repo := mod.Objects.GetRepository(repoName)
		if repo == nil {
			mod.log.Logv(1, "cannot index %v: repository not found", repoName)
		}

		go func() {
			err := mod.indexRepo(ctx, repo)
			if err != nil {
				mod.log.Logv(1, "indexing %v disabled: %v", repo.Label(), err)
			}
		}()
	}

	<-ctx.Done()
	return nil
}

func (mod *Module) PrivateKeyID(key *crypto.PublicKey) (*astral.ObjectID, error) {
	keyText, err := key.MarshalText()
	if err != nil {
		return nil, err
	}

	row, err := mod.db.findPrivateKeyByPublicKey(string(keyText))
	if err != nil {
		return nil, err
	}

	return row.KeyID, nil
}

func (mod *Module) PrivateKey(ctx *astral.Context, key *crypto.PublicKey) (*crypto.PrivateKey, error) {
	keyText, err := key.MarshalText()
	if err != nil {
		return nil, err
	}

	row, err := mod.db.findPrivateKeyByPublicKey(string(keyText))
	if err != nil {
		return nil, err
	}

	object, err := mod.Objects.Load(ctx, mod.Objects.ReadDefault(), row.KeyID)
	switch object := object.(type) {
	case nil:
		return nil, err

	case *crypto.PrivateKey:
		return object, nil

	default:
		return nil, astral.NewErrUnexpectedObject(object)
	}
}

func (mod *Module) PublicKey(ctx *astral.Context, key *crypto.PrivateKey) (*crypto.PublicKey, error) {
	for _, engine := range mod.engines.Clone() {
		pubKey, err := engine.PublicKey(ctx, key)
		if err == nil {
			return pubKey, nil
		}
	}

	return nil, crypto.ErrUnsupported
}

func (mod *Module) HashSigner(key *crypto.PublicKey, scheme string) (crypto.HashSigner, error) {
	for _, engine := range mod.engines.Clone() {
		signer, err := engine.HashSigner(key, scheme)
		if err == nil {
			return signer, nil
		}
	}

	return nil, crypto.ErrUnsupported
}

func (mod *Module) NodeSigner() crypto.HashSigner {
	signer, err := mod.HashSigner(&crypto.PublicKey{
		Type: "secp256k1",
		Key:  mod.node.Identity().PublicKey().SerializeCompressed(),
	}, crypto.SchemeASN1)
	if err != nil {
		panic(err)
	}

	return signer
}

func (mod *Module) VerifyHashSignature(key *crypto.PublicKey, sig *crypto.Signature, hash []byte) error {
	// check args
	switch {
	case key == nil:
		return errors.New("public key is nil")
	case len(key.Key) == 0:
		return errors.New("public key data is empty")
	case len(key.Type) == 0:
		return errors.New("public key type is empty")
	case sig == nil:
		return errors.New("signature is nil")
	case len(sig.Data) == 0:
		return errors.New("signature data is empty")
	case len(sig.Scheme) == 0:
		return errors.New("signature scheme is empty")
	case len(hash) == 0:
		return errors.New("hash is empty")
	}

	// find an engine that can verify the signature
	for _, engine := range mod.engines.Clone() {
		err := engine.VerifyHashSignature(key, sig, hash)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, crypto.ErrInvalidSignature):
			return err
		default:
			continue
		}
	}

	return crypto.ErrUnsupported
}

func (mod *Module) TextSigner(key *crypto.PublicKey, scheme string) (crypto.TextSigner, error) {
	for _, engine := range mod.engines.Clone() {
		signer, err := engine.TextSigner(key, scheme)
		if err == nil {
			return signer, nil
		}
	}

	return nil, crypto.ErrUnsupported
}

func (mod *Module) VerifyTextSignature(key *crypto.PublicKey, sig *crypto.Signature, msg string) error {
	// check args
	switch {
	case key == nil:
		return errors.New("public key is nil")
	case len(key.Key) == 0:
		return errors.New("public key data is empty")
	case len(key.Type) == 0:
		return errors.New("public key type is empty")
	case sig == nil:
		return errors.New("signature is nil")
	case len(sig.Data) == 0:
		return errors.New("signature data is empty")
	case len(sig.Scheme) == 0:
		return errors.New("signature scheme is empty")
	case len(msg) == 0:
		return errors.New("hash is empty")
	}

	// find an engine that can verify the signature
	for _, engine := range mod.engines.Clone() {
		err := engine.VerifyTextSignature(key, sig, msg)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, crypto.ErrInvalidSignature):
			return err
		default:
			continue
		}
	}

	return crypto.ErrUnsupported
}

func (mod *Module) ObjectSigner(key *crypto.PublicKey) (crypto.ObjectSigner, error) {
	return &ObjectSigner{
		mod:    mod,
		scheme: crypto.SchemeASN1,
		key:    key,
	}, nil
}

func (mod *Module) VerifyObjectSignature(key *crypto.PublicKey, signature *crypto.Signature, object crypto.SignableObject) error {
	return mod.VerifyHashSignature(key, signature, object.SignableHash())
}

func (mod *Module) TextObjectSigner(key *crypto.PublicKey) (crypto.TextObjectSigner, error) {
	return &TextObjectSigner{
		mod:    mod,
		scheme: crypto.SchemeBIP137,
		key:    key,
	}, nil
}

func (mod *Module) VerityTextObjectSignature(key *crypto.PublicKey, signature *crypto.Signature, object crypto.SignableTextObject) error {
	return mod.VerifyTextSignature(key, signature, mod.formatSignableText(object))
}

func (mod *Module) formatSignableText(object crypto.SignableTextObject) string {
	commitment := base64.StdEncoding.EncodeToString(object.SignableHash()[0:15])

	return fmt.Sprintf("[%s] %s", commitment, object.SignableText())
}

// indexRepo scans and follows the given repo and attempts to index all private keys it encounters
func (mod *Module) indexRepo(ctx *astral.Context, repo objects.Repository) error {
	scan, err := repo.Scan(ctx, true)
	if err != nil {
		return err
	}

	mod.log.Logv(1, "auto-indexing %v", repo.Label())

	for objectID := range scan {
		if objectID.Size > maxObjectSize {
			// skip large objects
			continue
		}

		object, err := mod.Objects.Load(ctx, repo, objectID)
		switch object := object.(type) {
		case nil:
			switch {
			case errors.Is(err, &astral.ErrBlueprintNotFound{}):
				// ignore missing blueprints
			default:
				// log other errors
				mod.log.Logv(2, "indexRepo: error loading object %v from %v: %v", objectID, repo, err)
			}

		case *crypto.PrivateKey:
			// attempt to index private keys
			err = mod.indexPrivateKey(object)
			if err == nil {
				mod.log.Logv(2, "indexed private key %v (%v)", objectID, object.Type)
			}
		}
	}

	return nil
}

// indexPrivateKey attempts to index the given private key
func (mod *Module) indexPrivateKey(key *crypto.PrivateKey) error {
	keyID, err := astral.ResolveObjectID(key)
	if err != nil {
		return err
	}

	pubKey, err := mod.PublicKey(mod.ctx, key)
	if err != nil {
		return err
	}

	pubKeyID, err := astral.ResolveObjectID(pubKey)
	if err != nil {
		return err
	}

	pubKeyText, err := pubKey.MarshalText()
	if err != nil {
		return err
	}

	_, err = mod.db.createPrivateKey(keyID, string(key.Type), pubKeyID, string(pubKeyText))

	return err
}

func (mod *Module) AddEngine(engine crypto.Engine) {
	mod.engines.Add(engine)
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.scope
}

func (mod *Module) String() string {
	return crypto.ModuleName
}
