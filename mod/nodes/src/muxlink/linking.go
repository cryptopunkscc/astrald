package muxlink

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const DefaultWorkerCount = 8
const DefaultTimeout = time.Minute
const featureMux = "mux"
const featureListFormat = "[s][c]c"

// Open negotiaties a link over the provided conn as the active party
func Open(
	ctx context.Context,
	conn exonet.Conn,
	remoteID id.Identity,
	localID id.Identity,
	localRouter net.Router,
) (link *Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	secureConn, err := noise.HandshakeOutbound(ctx, conn, remoteID, localID)
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	var linkFeatures []string

	err = cslq.Decode(secureConn, featureListFormat, &linkFeatures)
	if err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}

	var muxFound bool
	for _, f := range linkFeatures {
		if f == featureMux {
			muxFound = true
		}
	}
	if !muxFound {
		return nil, errors.New("remote party does not support mux")
	}

	err = cslq.Encode(secureConn, "[c]c", featureMux)
	if err != nil {
		return nil, fmt.Errorf("write mux: %w", err)
	}

	var errCode int
	err = cslq.Decode(secureConn, "c", &errCode)
	if errCode != 0 {
		return nil, errors.New("link feature negotation error")
	}

	return NewLink(secureConn, localRouter), nil
}

// Accept negotiaties a link over the provided conn as the passive party
func Accept(ctx context.Context, conn exonet.Conn, localID id.Identity, localRouter net.Router) (link *Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	secureConn, err := noise.HandshakeInbound(ctx, conn, localID)
	if err != nil {
		return
	}

	var linkFeatures = []string{featureMux}

	err = cslq.Encode(secureConn, featureListFormat, linkFeatures)
	if err != nil {
		return
	}

	for {
		var feature string
		err = cslq.Decode(secureConn, "[c]c", &feature)
		if err != nil {
			return
		}

		switch feature {
		case featureMux:
			cslq.Encode(secureConn, "c", 0)
			return NewLink(secureConn, localRouter), nil

		default:
			cslq.Encode(secureConn, "c", 1)
			return nil, fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				secureConn.RemoteIdentity(),
				secureConn.RemoteEndpoint(),
				feature,
			)
		}
	}
}
