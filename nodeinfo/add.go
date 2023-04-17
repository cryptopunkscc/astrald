package nodeinfo

import (
	"github.com/cryptopunkscc/astrald/node"
	"time"
)

func AddToContacts(nodeInfo *NodeInfo, contacts node.Contacts) error {
	c, err := contacts.FindOrCreate(nodeInfo.Identity)
	if err != nil {
		return err
	}

	if c.Alias() == "" {
		return c.SetAlias(nodeInfo.Alias)
	}

	return nil
}

func AddToTracker(nodeInfo *NodeInfo, tracker node.Tracker) error {
	for _, a := range nodeInfo.Addresses {
		if err := tracker.Add(nodeInfo.Identity, a, time.Now().Add(7*24*time.Hour)); err != nil {
			return err
		}
	}
	return nil
}
