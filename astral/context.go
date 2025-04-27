package astral

import (
	"context"
	"time"
)

type Context struct {
	context.Context
	identity *Identity
	zone     Zone
}

// NewContext returns a new anonymous Context. If ctx is nil, context.Background() is used.
func NewContext(ctx context.Context) *Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Context{Context: ctx, zone: DefaultZones}
}

func (ctx *Context) Identity() *Identity {
	return ctx.identity
}

func (ctx *Context) Zone() Zone {
	return ctx.zone
}

func (ctx *Context) WithIdentity(id *Identity) *Context {
	c := ctx.clone()
	c.identity = id
	return c
}

// WithZones returns a new context with zones set to z
func (ctx *Context) WithZones(z Zone) *Context {
	c := ctx.clone()
	c.zone = z
	return c
}

// IncludeZones returns a new context with additional zones
func (ctx *Context) IncludeZones(z Zone) *Context {
	c := ctx.clone()
	c.zone = c.zone | z
	return c
}

// ExcludeZones returns a new context with zones z removed
func (ctx *Context) ExcludeZones(z Zone) *Context {
	c := ctx.clone()
	c.zone = c.zone & ^z
	return c
}

// LimitZones returns a new context with zones limited to z
func (ctx *Context) LimitZones(z Zone) *Context {
	c := ctx.clone()
	c.zone = c.zone & z
	return c
}

func (ctx *Context) WithCancel() (*Context, context.CancelFunc) {
	clone := ctx.clone()

	cctx, cancel := context.WithCancel(ctx.Context)
	clone.Context = cctx
	return clone, cancel
}

func (ctx *Context) WithTimeout(d time.Duration) (clone *Context, cancel context.CancelFunc) {
	clone = ctx.clone()
	clone.Context, cancel = context.WithTimeout(ctx.Context, d)
	return
}

func (ctx *Context) clone() *Context {
	return &Context{
		Context:  ctx.Context,
		identity: ctx.identity,
		zone:     ctx.zone,
	}
}
