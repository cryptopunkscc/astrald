package keys

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ keys.Module = &Module{}

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
	db     *gorm.DB
	scope  shell.Scope
}

var ErrAlreadyIndexed = errors.New("already indexed")

func (mod *Module) Run(ctx *astral.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) CreateKey(alias string) (identity *astral.Identity, objectID *astral.ObjectID, err error) {
	_, err = mod.Dir.ResolveIdentity(alias)
	if err == nil {
		return identity, objectID, errors.New("alias already in use")
	}

	identity = astral.GenerateIdentity()

	objectID, err = mod.SaveKey(identity)
	if err != nil {
		return
	}

	err = mod.Dir.SetAlias(identity, alias)

	return
}

func (mod *Module) SaveKey(key *astral.Identity) (*astral.ObjectID, error) {
	if key.PrivateKey() == nil {
		return nil, errors.New("private key is nil")
	}

	pk := &keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: key.PrivateKey().Serialize(),
	}

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	objectID, err := objects.Save(ctx, pk, mod.Objects.Root())
	if err != nil {
		mod.log.Errorv(1, "error saving private key %v: %v", key, err)
	}

	return objectID, mod.IndexKey(objectID)
}

func (mod *Module) IndexKey(objectID *astral.ObjectID) error {
	var row dbPrivateKey
	var err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err == nil {
		return ErrAlreadyIndexed
	}

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	pk, err := objects.Load[*keys.PrivateKey](ctx, mod.Objects.Root(), objectID, mod.Objects.Blueprints())

	if pk.Type != keys.KeyTypeIdentity {
		return errors.New("unsupported key type")
	}

	identity, err := astral.IdentityFromPrivKeyBytes(pk.Bytes)
	if err != nil {
		return err
	}

	err = mod.db.Create(&dbPrivateKey{
		DataID:    objectID,
		Type:      string(pk.Type),
		PublicKey: identity,
	}).Error

	switch {
	case err == nil:
		return nil
	case strings.Contains(err.Error(), "UNIQUE constraint failed"):
		return nil
	default:
		return err
	}
}

func (mod *Module) FindIdentity(hex string) (*astral.Identity, error) {
	var row dbPrivateKey

	tx := mod.db.Where("type = ? and public_key = ?", keys.KeyTypeIdentity, hex).First(&row)
	if tx.Error != nil {
		return &astral.Identity{}, tx.Error
	}

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	pk, err := objects.Load[*keys.PrivateKey](ctx, mod.Objects.Root(), row.DataID, mod.Objects.Blueprints())
	if err != nil {
		return &astral.Identity{}, err
	}

	return astral.IdentityFromPrivKeyBytes(pk.Bytes)
}

func (mod *Module) SignASN1(identity *astral.Identity, hash []byte) ([]byte, error) {
	var err error

	if identity.PrivateKey() == nil {
		identity, err = mod.FindIdentity(identity.String())
		if err != nil {
			return nil, err
		}
	}

	return ecdsa.SignASN1(rand.Reader, identity.PrivateKey().ToECDSA(), hash)
}

func (mod *Module) VerifyASN1(signer *astral.Identity, hash []byte, sig []byte) error {
	// check args
	switch {
	case signer.IsZero():
		return errors.New("signer missing")
	case len(sig) == 0:
		return errors.New("signature missing")
	case len(hash) == 0:
		return errors.New("hash missing")
	}

	// verify sig
	if ecdsa.VerifyASN1(signer.PublicKey().ToECDSA(), hash, sig) {
		return nil
	}

	return errors.New("verification failed")
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.scope
}

func (mod *Module) String() string {
	return keys.ModuleName
}
