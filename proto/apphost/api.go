package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

type AppHost interface {
	Register(port string, target string) error
	Query(identity id.Identity, query string) (io.ReadWriteCloser, error)
	Resolve(s string) (id.Identity, error)
	NodeInfo(identity id.Identity) (NodeInfo, error)
}

type NodeInfo struct {
	Identity id.Identity
	Name     string
}

func (NodeInfo) FormatCSLQ() string { return "v [c]c" }

var (
	ErrUnknownCommand   = errors.New("unknown command")
	ErrRejected         = errors.New("rejected")
	ErrFailed           = errors.New("failed")
	ErrInvalidErrorCode = errors.New("invalid error code")
)
