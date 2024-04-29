package keys

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"strings"
)

var _ keys.Module = &Module{}

type Module struct {
	config  Config
	node    node.Node
	log     *log.Logger
	assets  assets.Assets
	objects objects.Module
	content content.Module
	db      *gorm.DB
}

var ErrAlreadyIndexed = errors.New("already indexed")
var privateKeyHeader = adc.Header(keys.PrivateKeyDataType)

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) CreateKey(alias string) (identity id.Identity, objectID object.ID, err error) {
	if _, err := mod.node.Tracker().IdentityByAlias(alias); err == nil {
		return identity, objectID, errors.New("alias already in use")
	}

	identity, err = id.GenerateIdentity()
	if err != nil {
		return
	}

	objectID, err = mod.SaveKey(identity)
	if err != nil {
		return
	}

	err = mod.node.Tracker().SetAlias(identity, alias)

	return
}

func (mod *Module) SaveKey(key id.Identity) (object.ID, error) {
	if key.PrivateKey() == nil {
		return object.ID{}, errors.New("private key is nil")
	}

	pk := keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: key.PrivateKey().Serialize(),
	}

	w, err := mod.objects.Create(&objects.CreateOpts{Alloc: 70})
	if err != nil {
		return object.ID{}, err
	}

	err = adc.WriteHeader(w, privateKeyHeader)
	if err != nil {
		return object.ID{}, nil
	}

	err = cslq.Encode(w, "v", &pk)
	if err != nil {
		return object.ID{}, err
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

	r, err := mod.objects.Open(objectID, &objects.OpenOpts{Virtual: true})
	if err != nil {
		return err
	}
	defer r.Close()

	err = adc.ExpectHeader(r, keys.PrivateKeyDataType)
	if err != nil {
		return err
	}

	var pk keys.PrivateKey
	if err = cslq.Decode(r, "v", &pk); err != nil {
		return err
	}

	if pk.Type != keys.KeyTypeIdentity {
		return errors.New("unsupported key type")
	}

	identity, err := id.ParsePrivateKey(pk.Bytes)
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
	r, err := mod.objects.Open(objectID, &objects.OpenOpts{Virtual: true})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	err = adc.ExpectHeader(r, keys.PrivateKeyDataType)
	if err != nil {
		return nil, err
	}

	var pk keys.PrivateKey
	if err = cslq.Decode(r, "v", &pk); err != nil {
		return nil, err
	}

	return &pk, nil
}

func (mod *Module) FindIdentity(hex string) (id.Identity, error) {
	var row dbPrivateKey

	tx := mod.db.Where("type = ? and public_key = ?", keys.KeyTypeIdentity, hex).First(&row)
	if tx.Error != nil {
		return id.Identity{}, tx.Error
	}

	pk, err := mod.LoadPrivateKey(row.DataID)
	if err != nil {
		return id.Identity{}, err
	}

	return id.ParsePrivateKey(pk.Bytes)
}

func (mod *Module) Sign(identity id.Identity, hash []byte) ([]byte, error) {
	var err error

	if identity.PrivateKey() == nil {
		identity, err = mod.FindIdentity(identity.PublicKeyHex())
		if err != nil {
			return nil, err
		}
	}

	return ecdsa.SignASN1(rand.Reader, identity.PrivateKey().ToECDSA(), hash)
}
