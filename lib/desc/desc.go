package desc

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/astral"
)

type Describer[T any] interface {
	Describe(ctx context.Context, object T, opts *Opts) []*Desc
}

type DescribeFunc[T any] func(ctx context.Context, object T, opts *Opts) []*Desc

type Desc struct {
	Source id.Identity
	Data   Data
}

type Data interface {
	Type() string
}

type Opts struct {
	astral.Scope
}

func DefaultOpts() *Opts {
	return &Opts{
		Scope: *astral.DefaultScope(),
	}
}

func Collect[T any](ctx context.Context, object T, opts *Opts, d ...Describer[T]) []*Desc {
	if opts == nil {
		opts = DefaultOpts()
	}

	var descs []*Desc

	for _, describer := range d {
		var items = describer.Describe(ctx, object, opts)
		descs = append(descs, items...)
	}

	return descs
}

type Adapter[T any] struct {
	Func DescribeFunc[T]
}

func (a Adapter[T]) Describe(ctx context.Context, object T, opts *Opts) []*Desc {
	return a.Func(ctx, object, opts)
}

func Func[T any](fn DescribeFunc[T]) Adapter[T] {
	return Adapter[T]{
		Func: fn,
	}
}
