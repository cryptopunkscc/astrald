package user

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
)

var _ user.Module = &Module{}

type Module struct {
	config  Config
	node    node.Node
	log     *log.Logger
	assets  assets.Assets
	db      *gorm.DB
	storage storage.Module
	shares  shares.Module
	content content.Module
	sdp     discovery.Module
	relay   relay.Module
	keys    keys.Module
	admin   admin.Module
	apphost apphost.Module

	identities     sig.Map[string, *Identity]
	profileService *ProfileService
	notifyService  *NotifyService
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (mod *Module) Identities() []id.Identity {
	var list []id.Identity

	for _, i := range mod.identities.Clone() {
		list = append(list, i.identity)
	}

	return list
}

func (mod *Module) Find(identity id.Identity) *Identity {
	v, _ := mod.identities.Get(identity.PublicKeyHex())
	return v
}

func (mod *Module) discoverUsers(ctx context.Context) {
	events.Handle(ctx, mod.node.Events(), func(ctx context.Context, event discovery.EventDiscovered) error {
		// make sure we're not getting our own services
		if event.Identity.IsEqual(mod.node.Identity()) {
			return nil
		}

		for _, cert := range event.Info.Data {
			err := mod.checkCert(event.Identity, cert.Bytes)
			if err != nil {
				mod.log.Errorv(2, "checkCert %v from %v: %v", data.Resolve(cert.Bytes), event.Identity, err)
			}
		}

		for _, service := range event.Info.Services {
			// look only for user profiles
			if service.Type != userProfileServiceType {
				continue
			}

			// and only if they're hosted
			if service.Identity.IsEqual(event.Identity) {
				continue
			}

			mod.log.Infov(2, "user %v discovered on %v", service.Identity, event.Identity)

			if len(service.Extra) == 0 {
				continue
			}
		}

		return nil
	})
}

func (mod *Module) checkCert(relayID id.Identity, certBytes []byte) error {
	var r = bytes.NewReader(certBytes)

	var dataType adc.Header
	err := cslq.Decode(r, "v", &dataType)
	if err != nil {
		return err
	}
	if dataType != relay.RelayCertType {
		return errors.New("invalid data type")
	}

	var cert relay.RelayCert

	err = cslq.Decode(r, "v", &cert)
	if err != nil {
		return err
	}

	err = cert.Validate()
	if err != nil {
		return err
	}

	if !cert.RelayID.IsEqual(relayID) {
		mod.log.Errorv(2, "%v is not %v", cert.RelayID, relayID)
		return errors.New("relay mismatch")
	}

	mod.storage.StoreBytes(certBytes, nil)

	return nil
}
