package astral

import "context"

type Context struct {
	context.Context
	identity *Identity
}

func NewContext(ctx context.Context) *Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Context{Context: ctx}
}

func (ctx *Context) Identity() *Identity {
	return ctx.identity
}

func (ctx *Context) WithIdentity(id *Identity) *Context {
	c := ctx.clone()
	c.identity = id
	return c
}

func (ctx *Context) WithCancel() (*Context, context.CancelFunc) {
	clone := ctx.clone()
	
	cctx, cancel := context.WithCancel(ctx.Context)
	clone.Context = cctx
	return clone, cancel
}

func (ctx *Context) clone() *Context {
	return &Context{
		Context:  ctx.Context,
		identity: ctx.identity,
	}
}
