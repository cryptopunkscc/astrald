package nodes

import (
	"fmt"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	switch object := drop.Object().(type) {
	case *nodes.ObservedEndpointMessage:
		err := mod.receiveObservedEndpointMessage(drop.SenderID(), object)
		if err == nil {
			return drop.Accept(false)
		}

	case *events.Event:
		switch e := object.Data.(type) {
		case *nodes.StreamCreatedEvent:
			if e.StreamCount == 1 && slices.ContainsFunc(mod.User.LocalSwarm(),
				e.RemoteIdentity.IsEqual) {

				go func() {
					err := mod.updateNodeEndpoints(mod.ctx, e.RemoteIdentity)
					if err != nil {
						mod.log.Error("updating node endpoints failed: %v", err)
					}
				}()
			}

		}

	}

	return nil
}

func (mod *Module) receiveObservedEndpointMessage(source *astral.Identity, event *nodes.ObservedEndpointMessage) error {
	endpoint := event.Endpoint

	var i ip.IP
	switch e := endpoint.(type) {
	case *tcp.Endpoint:
		i = e.IP
	case *utp.Endpoint:
		i = e.IP
	default:
		// unknown endpoint type
		return nil
	}

	if i.IsPublic() {
		mod.log.Log(`public ip %v reflected from %v`, i, source)
		mod.AddObservedEndpoint(endpoint, i)
	}

	return nil
}

func (mod *Module) updateNodeEndpoints(ctx *astral.Context, identity *astral.Identity) error {
	resolveEndpointsQuery := query.New(ctx.Identity(), identity, nodes.MethodResolveEndpoints, &opResolveEndpointsArgs{
		ID: identity.String(),
	})

	ch, err := query.RouteChan(ctx.IncludeZone(astral.ZoneNetwork), mod.node, resolveEndpointsQuery)
	if err != nil {
		return fmt.Errorf("routing resolve endpoints query: %v", err)
	}

	var endpoints []exonet.Endpoint
queryReadLoop:
	for {
		obj, err := ch.Receive()
		if err != nil {
			return fmt.Errorf("error reading resolved endpoint: %v", err)
		}

		switch obj := obj.(type) {
		case exonet.Endpoint:
			endpoints = append(endpoints, obj)
			err = mod.AddEndpoint(identity, obj)
			if err != nil {
				mod.log.Error("adding resolved endpoint failed %v: %v", obj, err)
			}

			continue queryReadLoop
		case *astral.EOS:
			break queryReadLoop
		case *astral.ErrorMessage:
			return fmt.Errorf("error resolving endpoints: %v", obj.Error())
		default:
			return fmt.Errorf("unexpected response type: %T", obj)
		}
	}

	return nil
}
