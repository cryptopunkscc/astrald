package astral

import (
	"context"
	"time"
)

type Context struct {
	context.Context
	identity *Identity
	filter   IdentityFilter
	zone     Zone
}

// NewContext returns a new anonymous Context. If ctx is nil, context.Background() is used.
func NewContext(ctx context.Context) *Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Context{Context: ctx, zone: ZoneDefault}
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

// WithZone returns a new context with zone set to z
func (ctx *Context) WithZone(z Zone) *Context {
	c := ctx.clone()
	c.zone = z
	return c
}

// IncludeZone returns a new context with additional zone
func (ctx *Context) IncludeZone(z Zone) *Context {
	c := ctx.clone()
	c.zone = c.zone | z
	return c
}

// ExcludeZone returns a new context with zone z removed
func (ctx *Context) ExcludeZone(z Zone) *Context {
	c := ctx.clone()
	c.zone = c.zone & ^z
	return c
}

// LimitZone returns a new context with zone limited to z
func (ctx *Context) LimitZone(z Zone) *Context {
	c := ctx.clone()
	c.zone = c.zone & z
	return c
}

func (ctx *Context) WithCancel() (clone *Context, cancel context.CancelFunc) {
	clone = ctx.clone()
	clone.Context, cancel = context.WithCancel(ctx.Context)
	return
}

func (ctx *Context) WithCancelCause() (clone *Context, cancel context.CancelCauseFunc) {
	clone = ctx.clone()
	clone.Context, cancel = context.WithCancelCause(ctx.Context)
	return
}

func (ctx *Context) WithTimeout(d time.Duration) (clone *Context, cancel context.CancelFunc) {
	clone = ctx.clone()
	clone.Context, cancel = context.WithTimeout(ctx.Context, d)
	return
}

func (ctx *Context) WithFilter(filter IdentityFilter) *Context {
	clone := ctx.clone()
	clone.filter = filter
	return clone
}

func (ctx *Context) Filter(identity *Identity) bool {
	if ctx.filter == nil {
		return true
	}
	return ctx.filter(identity)
}

func (ctx *Context) clone() *Context {
	return &Context{
		Context:  ctx.Context,
		identity: ctx.identity,
		filter:   ctx.filter,
		zone:     ctx.zone,
	}
}
