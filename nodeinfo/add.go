package nodeinfo

import (
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

func AddToContacts(nodeInfo *NodeInfo, contacts *contacts.Manager) error {
	c, err := contacts.FindOrCreate(nodeInfo.Identity)
	if err != nil {
		return err
	}

	if c.Alias() == "" {
		return c.SetAlias(nodeInfo.Alias)
	}

	return nil
}

func AddToTracker(nodeInfo *NodeInfo, tracker *tracker.Tracker) error {
	for _, a := range nodeInfo.Addresses {
		if err := tracker.Add(nodeInfo.Identity, a, time.Now().Add(7*24*time.Hour)); err != nil {
			return err
		}
	}
	return nil
}
