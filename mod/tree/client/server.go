package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

// NodeOps is a set of tree node operations that can be used with lib/ops
type NodeOps struct {
	Node tree.Node
}

func NewNodeOps(node tree.Node) *NodeOps {
	return &NodeOps{Node: node}
}

type GetArgs struct {
	Path   string
	Follow bool   `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (ops *NodeOps) Get(ctx *astral.Context, q *ops.Query, args GetArgs) (err error) {
	ctx, cancel := ctx.WithCancel()
	defer cancel()

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	node, err := tree.Query(ctx, ops.Node, args.Path, false)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	go func() {
		_, _ = ch.Receive()
		cancel()
	}()
	
	val, err := node.Get(ctx, args.Follow)
	switch {
	case err == nil:
	default:
		return ch.Send(astral.Err(err))
	}

	for object := range val {
		ch.Send(object)
	}

	return ch.Send(&astral.EOS{})
}

type SetArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (ops *NodeOps) Set(ctx *astral.Context, q *ops.Query, args SetArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	node, err := tree.Query(ctx, ops.Node, args.Path, true)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Handle(ctx, func(object astral.Object) {
		err = node.Set(ctx, object)
		if err != nil {
			ch.Send(astral.NewError(err.Error()))
		} else {
			ch.Send(&astral.Ack{})
		}
	})
}

type DeleteArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (ops *NodeOps) Delete(ctx *astral.Context, q *ops.Query, args DeleteArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	node, err := tree.Query(ctx, ops.Node, args.Path, false)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = node.Delete(ctx)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}

type ListArgs struct {
	Path string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (ops *NodeOps) List(ctx *astral.Context, query *ops.Query, args ListArgs) (err error) {
	ch := query.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var path = "/"

	if len(args.Path) > 0 {
		path = args.Path
	}

	node, err := tree.Query(ctx, ops.Node, path, false)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	subs, err := node.Sub(ctx)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	for name := range subs {
		err = ch.Send((*astral.String8)(&name))
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
