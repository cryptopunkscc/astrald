package nodes

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// negotiateOutboundLink reads peer's supported features and the link nonce.
func (mod *Module) negotiateOutboundLink(aconn astral.Conn) (features []astral.String8, err error) {
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

// negotiateInboundLink sends our supported features.
func (mod *Module) negotiateInboundLink(aconn astral.Conn) error {
	linkFeatures := []string{featureMux2}
	if _, err := astral.Uint16(len(linkFeatures)).WriteTo(aconn); err != nil {
		return err
	}
	for _, feature := range linkFeatures {
		if _, err := astral.String8(feature).WriteTo(aconn); err != nil {
			return err
		}
	}

	return nil
}

func (mod *Module) setInboundLinkNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	nonce = astral.NewNonce()
	if _, err = nonce.WriteTo(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Module) readOutboundLinkNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	if _, err = nonce.ReadFrom(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Module) EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (_ nodes.Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	privKey, err := mod.getPrivateKey()
	if err != nil {
		return nil, err
	}

	aconn, err := noise.HandshakeOutbound(
		ctx,
		conn,
		remoteID.PublicKey(),
		secp256k1.PrivKeyFromBytes(privKey.Key),
	)
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

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
			return nil, fmt.Errorf("read outbound link nonce: %w", err)
		}

		link := newLink(mod, aconn, nonce, true)
		err = mod.linkPool.AddLink(link)

		return link, err
	}

	return nil, errors.New("no supported link types found")
}

func (mod *Module) EstablishInboundLink(ctx context.Context, conn exonet.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	privKey, err := mod.getPrivateKey()
	if err != nil {
		return err
	}

	aconn, err := noise.HandshakeInbound(ctx, conn, secp256k1.PrivKeyFromBytes(privKey.Key))
	if err != nil {
		return err
	}

	err = mod.negotiateInboundLink(aconn)
	if err != nil {
		return err
	}

	for {
		var feature string
		_, err = (*astral.String8)(&feature).ReadFrom(aconn)
		if err != nil {
			return err
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
				err = mod.linkPool.AddLink(link)
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
