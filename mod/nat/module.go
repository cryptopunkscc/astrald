package nat

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/reflectlink"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	"github.com/cryptopunkscc/astrald/node/network"
	"time"
)

const portName = "net.nat.tcp"
const dialTimeout = 5 * time.Second

type Module struct {
	node    node.Node
	mapping natMapping
}

var log = _log.Tag(ModuleName)

func (mod *Module) Run(ctx context.Context) error {
	go func() {
		for event := range mod.node.Events().Subscribe(ctx) {
			switch event := event.(type) {
			case reflectlink.EventLinkReflected:
				// filter out non-inet addresses
				inetAddr, ok := event.Info.ReflectAddr.(inet.Endpoint)
				if !ok {
					continue
				}

				// check if it's one of our detected addresses
				for _, a := range mod.node.Infra().Endpoints() {
					if net.EndpointEqual(a, inetAddr) {
						continue
					}
				}

				// if reflected address is different from the local endpoint it's possibly a nat mapping
				if net.EndpointEqual(event.Link.LocalEndpoint(), event.Info.ReflectAddr) {
					continue
				}

				// is it a public ip?
				if !inetAddr.IsPublicUnicast() {
					continue
				}

				var m natMapping
				if m.intAddr, ok = event.Link.LocalEndpoint().(inet.Endpoint); !ok {
					continue
				}
				m.extAddr = inetAddr

				log.Log("NAT mapping candidate: %s", m.String())

				mod.mapping = m

			case network.EventPeerLinked:
				if event.Link.Network() == inet.DriverName {
					continue
				}
				if !event.Link.Outbound() {
					continue
				}
				if !mod.mapping.extAddr.IsZero() {
					go func() {
						err := mod.query(ctx, event.Peer.Identity())
						if err == nil {
							return
						}
						if err != nil {
							log.Error("query error: %s", err)
						}
					}()
				}
			}
		}
	}()

	go func() {
		mod.runServer(ctx)
	}()

	<-ctx.Done()
	return nil
}
