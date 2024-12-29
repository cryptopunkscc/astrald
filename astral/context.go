package astral

import "context"

type Context interface {
	context.Context
	Identitiy() *Identity
}

var _ Context = &wrappedContext{}

type wrappedContext struct {
	context.Context
	identitiy *Identity
}

func WrapContext(context context.Context, identitiy *Identity) Context {
	return &wrappedContext{Context: context, identitiy: identitiy}
}

func (ctx wrappedContext) Identitiy() *Identity {
	return ctx.identitiy
}
