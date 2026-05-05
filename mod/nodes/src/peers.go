package nodes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Peers struct {
	*Module
}

func NewPeers(m *Module) *Peers {
	return &Peers{Module: m}
}

func (mod *Peers) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (_ io.WriteCloser, err error) {
	links := mod.linkPool.links.Select(func(link *Link) bool {
		return link.RemoteIdentity().IsEqual(q.Target)
	})

	if len(links) == 0 {
		return query.RouteNotFound()
	}

	return links[0].GetMux().RouteQuery(ctx, q, w)
}

func (mod *Peers) peers() (peers []*astral.Identity) {
	var r = map[string]struct{}{}

	for _, link := range mod.linkPool.links.Clone() {
		if _, found := r[link.RemoteIdentity().String()]; found {
			continue
		}
		r[link.RemoteIdentity().String()] = struct{}{}
		peers = append(peers, link.RemoteIdentity())
	}

	return
}

func (mod *Peers) addLink(
	link *Link,
) (err error) {
	var (
		dir     = "in"
		netName = "unknown network"
	)

	if link.outbound {
		dir = "out"
	}

	// try to figure out the network name
	switch {
	case link.LocalEndpoint() != nil:
		netName = link.LocalEndpoint().Network()
	case link.RemoteEndpoint() != nil:
		netName = link.RemoteEndpoint().Network()
	}

	err = mod.linkPool.links.Add(link)
	if err != nil {
		return
	}

	link.GetMux().SetRouter(mod.node)

	// log link addition
	mod.log.Infov(1, "added %v-link with %v (%v)", dir, link.RemoteIdentity(), netName)
	linksWithSameIdentity := mod.linkPool.links.Select(func(v *Link) bool {
		return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
	})

	if !link.outbound {
		mod.linkPool.notifyLinkWatchers(link, nil)
	}

	mod.Events.Emit(&nodes.LinkCreatedEvent{
		RemoteIdentity: link.RemoteIdentity(),
		LinkID:         link.id,
		LinkCount:      len(linksWithSameIdentity),
	})

	// handle the link
	go func() {
		mod.readLinkFrames(link)
		// remove the link and its connections
		mod.linkPool.links.Remove(link)
		link.GetMux().closeAllSessions()

		linksWithSameIdentity := mod.linkPool.links.Select(func(v *Link) bool {
			return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
		})

		mod.Events.Emit(&nodes.LinkClosedEvent{
			RemoteIdentity: link.RemoteIdentity(),
			Forced:         false,
			LinkCount:      astral.Int8(len(linksWithSameIdentity)),
		})

		mod.log.Info("closed %v-link with %v (%v): %v", dir, link.RemoteIdentity(), netName, link.Err())
	}()

	// reflect the link
	go mod.reflectLink(link)

	return
}

// reflectLink reflects the observed remote endpoint on inbound links
func (mod *Peers) reflectLink(link *Link) (err error) {
	// reflect only on inbound links with a known remote endpoint
	if link.outbound || link.RemoteEndpoint() == nil {
		return
	}

	// note: rethink maybe switch (?)
	if _, ok := link.RemoteEndpoint().(*gateway.Endpoint); ok {
		// dont reflect gateway endpoints
		return
	}
	// reflect the endpoint
	err = mod.Objects.Push(mod.ctx, link.RemoteIdentity(),
		&nodes.ObservedEndpointMessage{
			Endpoint: link.RemoteEndpoint(),
		})

	// log the result
	if err != nil {
		mod.log.Errorv(2, "Objects.Push(%v, %v): %v", link.RemoteIdentity(), link.RemoteEndpoint(), err)
	} else {
		mod.log.Logv(2, "reflected endpoint %v to %v", link.RemoteEndpoint(), link.RemoteIdentity())
	}

	return
}

// readLinkFrames reads frames from the link until it closes
func (mod *Peers) readLinkFrames(link *Link) {
	for frame := range link.Read() {
		link.GetMux().HandleFrame(frame)
	}
}

// isLinked returns true if there's at least one link with the given remote identity
func (mod *Peers) isLinked(remoteID *astral.Identity) bool {
	for _, link := range mod.linkPool.links.Clone() {
		if link.RemoteIdentity().IsEqual(remoteID) {
			return true
		}
	}
	return false
}

// negotiateOutboundLink reads peer's supported features and the session sessionId.
func (mod *Peers) negotiateOutboundLink(aconn astral.Conn) (features []astral.String8, err error) {
	var featCount astral.Uint16
	if _, err = featCount.ReadFrom(aconn); err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}
	for i := 0; i < int(featCount); i++ {
		var feat astral.String8
		if _, err = feat.ReadFrom(aconn); err != nil {
			return nil, fmt.Errorf("read features: %w", err)
		}
		features = append(features, feat)
	}

	return features, nil
}

// negotiateInboundLink sends our supported features and a fresh session sessionId.
func (mod *Peers) negotiateInboundLink(aconn astral.Conn) (err error) {
	var linkFeatures = []string{featureMux2}
	if _, err = astral.Uint16(len(linkFeatures)).WriteTo(aconn); err != nil {
		return err
	}
	for _, feature := range linkFeatures {
		if _, err = astral.String8(feature).WriteTo(aconn); err != nil {
			return err
		}
	}

	return nil
}

func (mod *Peers) setInboundLinkNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	nonce = astral.NewNonce()
	if _, err = nonce.WriteTo(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Peers) readOutboundLinkNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	if _, err = nonce.ReadFrom(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Peers) EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (_ *Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// get the node's private key
	privKey, err := mod.getPrivateKey()
	if err != nil {
		return nil, err
	}

	// initiate a handshake
	aconn, err := noise.HandshakeOutbound(
		ctx,
		conn,
		remoteID.PublicKey(),
		secp256k1.PrivKeyFromBytes(privKey.Key),
	)
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	// Read peer features and session sessionId
	linkFeatures, err := mod.negotiateOutboundLink(aconn)
	if err != nil {
		return nil, err
	}

	if slices.Contains(linkFeatures, featureMux2) {
		_, err = astral.String8(featureMux2).WriteTo(aconn)
		if err != nil {
			return nil, fmt.Errorf("write: %w", err)
		}

		var errCode astral.Int8
		_, err = errCode.ReadFrom(aconn)
		switch {
		case err != nil:
			return nil, fmt.Errorf("read: %w", err)
		case errCode != 0:
			return nil, errors.New("link feature negotation error")
		}

		nonce, err := mod.readOutboundLinkNonce(aconn)
		if err != nil {
			return nil, fmt.Errorf(`read outbound link nonce: %w`, err)
		}

		link := newLink(mod, aconn, nonce, true)
		err = mod.addLink(link)

		return link, err
	}

	return nil, errors.New("no supported link types found")
}

func (mod *Peers) EstablishInboundLink(ctx context.Context, conn exonet.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// get the node's private key
	privKey, err := mod.getPrivateKey()
	if err != nil {
		return err
	}

	// respond to a handshake
	aconn, err := noise.HandshakeInbound(ctx, conn, secp256k1.PrivKeyFromBytes(privKey.Key))
	if err != nil {
		return
	}

	// Send our features and session sessionId
	err = mod.negotiateInboundLink(aconn)
	if err != nil {
		return err
	}

	for {
		var feature string
		_, err = (*astral.String8)(&feature).ReadFrom(aconn)
		if err != nil {
			return
		}

		switch feature {
		case featureMux2:
			_, err = astral.Uint8(0).WriteTo(aconn)
			if err == nil {
				nonce, err := mod.setInboundLinkNonce(aconn)
				if err != nil {
					return fmt.Errorf("failed to set inbound link nonce: %w", err)
				}

				link := newLink(mod, aconn, nonce, false)
				err = mod.addLink(link)
				if err != nil {
					return fmt.Errorf("failed to add link: %w", err)
				}
			}

			return err
		default:
			_, err = astral.Uint8(1).WriteTo(aconn)
			return fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				aconn.RemoteIdentity(),
				aconn.RemoteEndpoint(),
				feature,
			)
		}
	}
}
