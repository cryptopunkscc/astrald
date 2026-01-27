package tree

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type Node struct {
	client *Client
	path   []string
}

var _ tree.Node = &Node{}

func (node *Node) Name() string {
	if len(node.path) == 0 {
		return ""
	}
	return node.path[len(node.path)-1]
}

func (node *Node) Get(ctx *astral.Context, follow bool) (<-chan astral.Object, error) {
	ch, err := node.client.queryCh(ctx, "tree.get", query.Args{
		"path":   node.Path(),
		"follow": follow,
	})

	if err != nil {
		return nil, err
	}

	var out = make(chan astral.Object, 1)

	// get the initial value
	obj, err := ch.Receive()
	switch obj := obj.(type) {
	case nil:
		return nil, err
	case *tree.ErrNoValue:
		return nil, obj
	}

	out <- obj

	go func() {
		defer ch.Close()
		defer close(out)
		for {
			obj, _ := ch.Receive()
			switch obj.(type) {
			case nil:
				return
			case *astral.EOS:
				return
			}

			select {
			case <-ctx.Done():
				return
			case out <- obj:
			}
		}
	}()

	return out, nil
}

func (node *Node) Set(ctx *astral.Context, object astral.Object) error {
	ch, err := node.client.queryCh(ctx, "tree.set", query.Args{
		"path": node.Path(),
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Send(object)
	if err != nil {
		return err
	}

	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *astral.Ack:
		return nil
	case nil:
		return err
	case *astral.ErrorMessage:
		return msg
	default:
		return astral.NewErrUnexpectedObject(msg)
	}
}

func (node *Node) Delete(ctx *astral.Context) error {
	ch, err := node.client.queryCh(ctx, "tree.delete", query.Args{
		"path": node.Path(),
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *astral.Ack:
		return nil
	case nil:
		return err
	case *astral.ErrorMessage:
		return msg
	default:
		return astral.NewErrUnexpectedObject(msg)
	}
}

func (node *Node) Sub(ctx *astral.Context) (map[string]tree.Node, error) {
	var sub = make(map[string]tree.Node)

	ch, err := node.client.queryCh(ctx, "tree.list", query.Args{
		"path": node.Path(),
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	err = ch.Switch(
		func(msg *astral.String8) {
			sub[string(*msg)] = &Node{client: node.client, path: append(node.path, string(*msg))}
		},
		channel.StopOnEOS,
		channel.PassErrors,
	)

	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (node *Node) Create(ctx *astral.Context, name string) (tree.Node, error) {
	newPath := "/" + strings.Join(append(node.path, name), "/")

	// calling set without sending any value will still create the node
	ch, err := node.client.queryCh(ctx, "tree.set", query.Args{
		"path": newPath,
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	return &Node{client: node.client, path: append(node.path, name)}, nil
}

func (node *Node) Path() string {
	return "/" + strings.Join(node.path, "/")
}
