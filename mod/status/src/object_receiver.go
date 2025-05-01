package status

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ether"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/status"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"time"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	// only receive objects from the local node
	if !drop.SenderID().IsEqual(mod.node.Identity()) {
		return errors.New("object rejected")
	}

	switch object := drop.Object().(type) {
	case *ether.EventBroadcastReceived:
		err := mod.receiveBroadcastEvent(object)
		if err == nil {
			return drop.Accept(false)
		}

	case *tcp.EventNetworkAddressChanged:
		mod.receiveNetworkAddressChanged(object)
		return drop.Accept(false)
	}

	return nil
}

func (mod *Module) receiveBroadcastEvent(event *ether.EventBroadcastReceived) error {
	switch object := event.Object.(type) {
	case *status.Status:
		return mod.receiveStatus(event.SourceID, event.SourceIP, object)

	case *ScanMessage:
		mod.Ether.PushToIP(event.SourceIP, mod.Status(astral.Anyone), nil)
		return errors.New("object rejected")
	}

	return errors.New("object rejected")
}

func (mod *Module) receiveStatus(sourceID *astral.Identity, addr tcp.IP, status *status.Status) error {
	mod.log.Infov(3, "update from %v %v", sourceID, addr)
	mod.cache.Replace(addr.String(), &cache{
		Identity:  sourceID,
		IP:        addr,
		Timestamp: time.Now(),
		Status:    status,
	})

	return errors.New("object rejected")
}

func (mod *Module) receiveNetworkAddressChanged(event *tcp.EventNetworkAddressChanged) {
	if len(event.Added) == 0 {
		return // do nothing if there are no new network addresses
	}

	// broadcast our status
	mod.pushStatus()

	// scan for other nodes
	mod.Scan()
}
