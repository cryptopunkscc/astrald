package keys

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ keys.Module = &Module{}

type Module struct {
	config  Config
	node    node.Node
	log     *log.Logger
	assets  assets.Store
	storage storage.Module
	data    data.Module
	db      *gorm.DB
}

var adc0PrivateKey = data.ADC0Header(keys.PrivateKeyDataType)

func (mod *Module) Run(ctx context.Context) error {
	err := mod.importKeys()
	if err != nil {
		mod.log.Error("error importing keys: %v", err)
	}

	return tasks.Group(
		&IndexerService{Module: mod},
		&Service{Module: mod},
	).Run(ctx)
}

func (mod *Module) CreateKey(alias string) (identity id.Identity, dataID _data.ID, err error) {
	if _, err := mod.node.Tracker().IdentityByAlias(alias); err == nil {
		return identity, dataID, errors.New("alias already in use")
	}

	identity, err = id.GenerateIdentity()
	if err != nil {
		return
	}

	dataID, err = mod.SaveKey(identity)
	if err != nil {
		return
	}

	err = mod.node.Tracker().SetAlias(identity, alias)

	return
}

func (mod *Module) SaveKey(key id.Identity) (_data.ID, error) {
	if key.PrivateKey() == nil {
		return _data.ID{}, errors.New("private key is nil")
	}

	pk := keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: key.PrivateKey().Serialize(),
	}

	w, err := mod.storage.Data().Store(&storage.StoreOpts{Alloc: 70})
	if err != nil {
		return _data.ID{}, err
	}

	err = cslq.Encode(w, "vv", &adc0PrivateKey, &pk)
	if err != nil {
		return _data.ID{}, err
	}

	dataID, err := w.Commit()
	if err != nil {
		mod.log.Errorv(1, "error importing private key %v: %v", key, err)
	}

	return dataID, mod.IndexKey(dataID)
}

func (mod *Module) IndexKey(dataID _data.ID) error {
	dataType, r, err := mod.data.OpenADC0(dataID)
	if err != nil {
		return err
	}
	if dataType != keys.PrivateKeyDataType {
		return errors.New("not a private key file")
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

	var row = dbPrivateKey{
		DataID:    dataID.String(),
		Type:      pk.Type,
		PublicKey: identity.PublicKeyHex(),
	}

	return mod.db.Create(&row).Error
}

func (mod *Module) LoadPrivateKey(dataID _data.ID) (*keys.PrivateKey, error) {
	dataType, r, err := mod.data.OpenADC0(dataID)
	if err != nil {
		return nil, err
	}
	if dataType != keys.PrivateKeyDataType {
		return nil, errors.New("not a private key file")
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

	dataID, err := _data.Parse(row.DataID)
	if err != nil {
		return id.Identity{}, err
	}

	pk, err := mod.LoadPrivateKey(dataID)
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

func (mod *Module) importKeys() error {
	store, err := mod.assets.KeyStore()
	if err != nil {
		return err
	}

	all, err := store.All()
	if err != nil {
		return err
	}

	for _, key := range all {
		if key.PrivateKey() == nil {
			continue
		}

		pk := keys.PrivateKey{
			Type:  keys.KeyTypeIdentity,
			Bytes: key.PrivateKey().Serialize(),
		}

		w, err := mod.storage.Data().Store(&storage.StoreOpts{Alloc: 70})
		if err != nil {
			return err
		}

		err = cslq.Encode(w, "vv", data.ADC0Header(keys.PrivateKeyDataType), &pk)
		if err != nil {
			continue
		}

		_, err = w.Commit()
		if err != nil {
			mod.log.Errorv(1, "error importing private key %v: %v", key, err)
		}
	}

	return nil
}
