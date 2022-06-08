package apphost

import (
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
)

type PortEntry struct {
	port   *hub.Port
	target string
}

// checkTarget checks if the port's target is alive
func (entry *PortEntry) checkTarget() bool {
	c, err := ipc.Dial(entry.target)
	if err == nil {
		c.Close()
	}
	return err == nil
}
