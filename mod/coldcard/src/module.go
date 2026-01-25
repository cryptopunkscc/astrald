package coldcard

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/coldcard"
	"github.com/cryptopunkscc/astrald/mod/coldcard/ckcc"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
	Crypto crypto.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	ops    ops.Set
	db     *DB

	devices sig.Map[string, string]
}

func (mod *Module) Run(ctx *astral.Context) error {
	go mod.Scan()
	<-ctx.Done()
	return nil
}

func (mod *Module) Scan() error {
	devices, err := ckcc.List()
	if err != nil {
		return err
	}

	for _, dev := range devices {
		pubKeyHex, err := dev.PubKey(coldcard.BIP44Path)
		if err != nil {
			continue
		}

		mod.devices.Set(dev.Serial, pubKeyHex)
		mod.log.Logv(1, "found coldcard device: %v for key %v", dev.Serial, pubKeyHex)
	}

	return nil
}

func (mod *Module) deviceForPublicKeyHex(keyHex string) *ckcc.Device {
	for serial, key := range mod.devices.Clone() {
		if key == keyHex {
			return ckcc.NewDevice(serial)
		}
	}

	return nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) String() string {
	return coldcard.ModuleName
}
