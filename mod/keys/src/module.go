package keys

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"strings"
)

var _ keys.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Content content.Module
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

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) CreateKey(alias string) (identity *astral.Identity, objectID object.ID, err error) {
	_, err = mod.Dir.ResolveIdentity(alias)
	if err == nil {
		return identity, objectID, errors.New("alias already in use")
	}

	identity, err = astral.GenerateIdentity()
	if err != nil {
		return
	}

	objectID, err = mod.SaveKey(identity)
	if err != nil {
		return
	}

	err = mod.Dir.SetAlias(identity, alias)

	return
}

func (mod *Module) SaveKey(key *astral.Identity) (object.ID, error) {
	if key.PrivateKey() == nil {
		return object.ID{}, errors.New("private key is nil")
	}

	pk := &keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: key.PrivateKey().Serialize(),
	}

	w, err := mod.Objects.Create(&objects.CreateOpts{Alloc: 70})
	if err != nil {
		return object.ID{}, err
	}

	err = astral.EncodeObject(w, pk)
	if err != nil {
		return object.ID{}, nil
	}

	objectID, err := w.Commit()
	if err != nil {
		mod.log.Errorv(1, "error importing private key %v: %v", key, err)
	}

	return objectID, mod.IndexKey(objectID)
}

func (mod *Module) IndexKey(objectID object.ID) error {
	var row dbPrivateKey
	var err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err == nil {
		return ErrAlreadyIndexed
	}

	r, err := mod.Objects.Open(context.Background(), objectID, objects.DefaultOpenOpts())
	if err != nil {
		return err
	}
	defer r.Close()

	var pk keys.PrivateKey

	err = astral.DecodeObject(r, &pk)
	if err != nil {
		return err
	}

	if pk.Type != keys.KeyTypeIdentity {
		return errors.New("unsupported key type")
	}

	identity, err := astral.IdentityFromPrivKeyBytes(pk.Bytes)
	if err != nil {
		return err
	}

	err = mod.db.Create(&dbPrivateKey{
		DataID:    objectID,
		Type:      pk.Type,
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

func (mod *Module) LoadPrivateKey(objectID object.ID) (*keys.PrivateKey, error) {
	r, err := mod.Objects.Open(context.Background(), objectID, objects.DefaultOpenOpts())
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var pk keys.PrivateKey
	err = astral.DecodeObject(r, &pk)

	return &pk, nil
}

func (mod *Module) FindIdentity(hex string) (*astral.Identity, error) {
	var row dbPrivateKey

	tx := mod.db.Where("type = ? and public_key = ?", keys.KeyTypeIdentity, hex).First(&row)
	if tx.Error != nil {
		return &astral.Identity{}, tx.Error
	}

	pk, err := mod.LoadPrivateKey(row.DataID)
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
