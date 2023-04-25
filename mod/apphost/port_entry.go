package apphost

import (
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"github.com/cryptopunkscc/astrald/node/services"
)

type PortEntry struct {
	port   *services.Service
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
