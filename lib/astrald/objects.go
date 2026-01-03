package astrald

import (
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	_ "github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type ObjectsClient struct {
	c        *Client
	targetID *astral.Identity
}

func NewObjectsClient(targetID *astral.Identity, client *Client) *ObjectsClient {
	if client == nil {
		client = DefaultClient()
	}
	return &ObjectsClient{c: client, targetID: targetID}
}

var defaultObjectsClient *ObjectsClient

func Objects() *ObjectsClient {
	if defaultObjectsClient == nil {
		defaultObjectsClient = NewObjectsClient(nil, DefaultClient())
	}
	return defaultObjectsClient
}

func (client *ObjectsClient) Read(ctx *astral.Context, objectID *astral.ObjectID, offset, limit int64) (io.ReadCloser, error) {
	return client.query(ctx, "objects.read", query.Args{
		"id":     objectID,
		"offset": offset,
		"limit":  limit,
		"zone":   "dvn",
	})
}

func (client *ObjectsClient) GetType(ctx *astral.Context, objectID *astral.ObjectID) (string, error) {
	ch, err := client.queryCh(ctx, "objects.get_type", query.Args{
		"id": objectID,
	})
	if err != nil {
		return "", err
	}
	defer ch.Close()

	res, err := ch.Receive()
	if err != nil {
		return "", err
	}

	switch res := res.(type) {
	case *astral.String8:
		return string(*res), nil

	case *astral.ErrorMessage:
		return "", res

	default:
		return "", errors.New("protocol error: unexpected object type")
	}
}

func (client *ObjectsClient) Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	ch, err := client.queryCh(ctx, "objects.describe", query.Args{
		"id": objectID.String(),
	})
	if err != nil {
		return nil, err
	}

	res := make(chan *objects.DescribeResult)

	go func() {
		defer close(res)

		ch.Handle(ctx, func(o astral.Object) {
			switch o := o.(type) {
			case *objects.DescribeResult:
				res <- o

			case *astral.EOS:
				ch.Close()

			default:
				ch.Close()
			}
		})
	}()

	return res, nil
}

func (client *ObjectsClient) Search(ctx *astral.Context, q string) (<-chan *objects.SearchResult, error) {
	ch, err := client.queryCh(ctx, "objects.search", query.Args{
		"q": q,
	})
	if err != nil {
		return nil, err
	}

	res := make(chan *objects.SearchResult)
	go func() {
		defer close(res)

		ch.Handle(ctx, func(o astral.Object) {
			switch o := o.(type) {
			case *objects.SearchResult:
				res <- o

			case *astral.EOS:
				ch.Close()

			default:
				ch.Close()
			}
		})
	}()

	return res, nil
}

func (client *ObjectsClient) Push(ctx *astral.Context, object astral.Object) error {
	ch, err := client.queryCh(ctx, "objects.push", nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Send(object)
	if err != nil {
		return err
	}

	res, err := ch.Receive()
	switch res := res.(type) {
	case *astral.Bool:
		if *res {
			return nil
		} else {
			return errors.New("object rejected")
		}

	case nil:
		return err

	default:
		return fmt.Errorf("unexpected response type: %s", res.ObjectType())
	}
}

func (client *ObjectsClient) Create(ctx *astral.Context, repo string, alloc int) (objects.Writer, error) {
	// prepare arguments
	args := query.Args{}

	if alloc > 0 {
		args["alloc"] = alloc
	}
	if len(repo) > 0 {
		args["repo"] = repo
	}

	// send the query
	ch, err := client.queryCh(ctx, "objects.create", args)
	if err != nil {
		return nil, err
	}

	// handle response
	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *astral.Ack:
		return &writer{ch: ch}, nil

	case *astral.ErrorMessage:
		return nil, msg

	case nil:
		ch.Close()
		return nil, err
	}

	ch.Close()
	return nil, fmt.Errorf("unexpected message type: %s", msg.ObjectType())
}

func (client *ObjectsClient) query(ctx *astral.Context, method string, args any) (astral.Conn, error) {
	return client.c.WithTarget(client.targetID).Query(ctx, method, args)
}

func (client *ObjectsClient) queryCh(ctx *astral.Context, method string, args any) (*channel.Channel, error) {
	return client.c.WithTarget(client.targetID).QueryChannel(ctx, method, args)
}

type writer struct {
	ch *channel.Channel
}

func (w *writer) Write(p []byte) (n int, err error) {
	err = w.ch.Send((*astral.Blob)(&p))
	if err == nil {
		n = len(p)
	}
	return
}

func (w *writer) Commit() (*astral.ObjectID, error) {
	// close the channel after committing
	defer w.ch.Close()

	// send commit message
	err := w.ch.Send(&objects.CommitMsg{})
	if err != nil {
		return nil, err
	}

	// handle response
	o, err := w.ch.Receive()
	switch msg := o.(type) {
	case *astral.ObjectID:
		return msg, nil
	case *astral.ErrorMessage:
		return nil, msg
	case nil:
		return nil, err
	default:
		return nil, fmt.Errorf("unexpected type: %s", msg.ObjectType())
	}
}

func (w *writer) Discard() error {
	return w.ch.Close() // close without committing to discard data
}
