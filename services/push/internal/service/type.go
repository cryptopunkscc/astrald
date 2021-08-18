package service

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serializer"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

type Context struct {
	request.Context
	api.Core
	Ctx context.Context
	Handlers map[byte]Handle
}

type Request struct {
	Context
	*serializer.ReadWriteCloser
	Caller api.Identity
}

type Handle func(r *Request) error

type Handles map[byte]Handle
