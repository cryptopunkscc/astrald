package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
)

const featureMux = "mux"
const featureListFormat = "[s][c]c"

func Open(ctx context.Context, conn net.Conn, remoteID id.Identity, localID id.Identity) (link *CoreLink, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	secureConn, err := auth.HandshakeOutbound(ctx, conn, remoteID, localID)
	if err != nil {
		return
	}

	var linkFeatures []string

	err = cslq.Decode(secureConn, featureListFormat, &linkFeatures)
	if err != nil {
		return
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
		return
	}

	var errCode int
	err = cslq.Decode(secureConn, "c", &errCode)
	if errCode != 0 {
		err = errors.New("link feature negotation error")
		return
	}

	return NewCoreLink(secureConn), nil
}

func Accept(ctx context.Context, conn net.Conn, localID id.Identity) (link *CoreLink, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	secureConn, err := auth.HandshakeInbound(ctx, conn, localID)
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
			return NewCoreLink(secureConn), nil

		default:
			cslq.Encode(secureConn, "c", 1)
			return nil, errors.New("unsupported feature requested by the remote party")
		}
	}
}
