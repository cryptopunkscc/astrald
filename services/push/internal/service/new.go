package service

import (
	"github.com/cryptopunkscc/astrald/services/util/request"
)

func New(
	port string,
	handlers Handles,
) *Context {
	return &Context{
		Context:  request.Context{Port: port},
		Handlers: handlers,
	}
}
