package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/rpc"
)

var es rpc.ErrorSpace

var (
	ErrRejected           = es.NewError(0x01, "rejected")
	ErrDenied             = es.NewError(0x02, "denied")
	ErrUnableToProcess    = es.NewError(0x03, "unable to process")
	ErrInvalidCertificate = es.NewError(0x04, "invalid certificate")
	ErrInvalidRequest     = es.NewError(0xff, "invalid request")
)
