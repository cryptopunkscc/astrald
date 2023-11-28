package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/rpc"
)

var es rpc.ErrorSpace

var (
	ErrRejected            = es.NewError(0x01, "rejected")
	ErrDenied              = es.NewError(0x02, "denied")
	ErrCertificateRejected = es.NewError(0x03, "certificate rejected")
	ErrRouteNotFound       = es.NewError(0x05, "route not found")
	ErrInternalError       = es.NewError(0xff, "internal error")
)
