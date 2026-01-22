package crypto

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
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
	scope  shell.Scope
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
		return nil, errors.New("unexpected object type: " + object.ObjectType())
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

func (mod *Module) VerifyHashSignature(key *crypto.PublicKey, sig *crypto.Signature, hash []byte) error {
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

func (mod *Module) MessageSigner(key *crypto.PublicKey, scheme string) (crypto.MessageSigner, error) {
	for _, engine := range mod.engines.Clone() {
		signer, err := engine.MessageSigner(key, scheme)
		if err == nil {
			return signer, nil
		}
	}

	return nil, crypto.ErrUnsupported
}

func (mod *Module) VerifyMessageSignature(key *crypto.PublicKey, sig *crypto.Signature, msg string) error {
	for _, engine := range mod.engines.Clone() {
		err := engine.VerifyMessageSignature(key, sig, msg)
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

func (mod *Module) Scope() *shell.Scope {
	return &mod.scope
}

func (mod *Module) String() string {
	return crypto.ModuleName
}
