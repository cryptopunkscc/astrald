package astrald

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type ObjectsClient struct {
	c      *Client
	target string
}

func NewObjectsClient(c *Client, target string) *ObjectsClient {
	return &ObjectsClient{c: c, target: target}
}

var defaultObjectsClient *ObjectsClient

func Objects() *ObjectsClient {
	if defaultObjectsClient == nil {
		defaultObjectsClient = NewObjectsClient(DefaultClient(), "")
	}
	return defaultObjectsClient
}

func (client *ObjectsClient) Read(objectID *astral.ObjectID, offset, limit uint64) (io.ReadCloser, error) {
	return client.query("objects.read", query.Args{
		"id":     objectID,
		"offset": offset,
		"limit":  limit,
		"zone":   "dvn",
	})
}

func (client *ObjectsClient) GetType(objectID *astral.ObjectID) (string, error) {
	ch, err := client.queryCh("objects.get_type", query.Args{
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

func (client *ObjectsClient) Describe(ctx context.Context, objectID *astral.ObjectID) (<-chan astral.Object, error) {
	ch, err := client.queryCh("objects.describe", query.Args{
		"id": objectID.String(),
	})
	if err != nil {
		return nil, err
	}

	res := make(chan astral.Object)
	done := make(chan struct{})

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			ch.Close()
		}
	}()

	go func() {
		defer close(res)
		defer close(done)

		for {
			obj, err := ch.Receive()
			if err != nil {
				return
			}

			res <- obj
		}
	}()

	return res, nil
}

func (client *ObjectsClient) Search(ctx context.Context, q string) (<-chan *objects.SearchResult, error) {
	ch, err := client.queryCh("objects.search", query.Args{
		"q": q,
	})
	if err != nil {
		return nil, err
	}

	res := make(chan *objects.SearchResult)
	done := make(chan struct{})

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			ch.Close()
		}
	}()

	go func() {
		defer close(res)
		defer close(done)

		for {
			o, _ := ch.Receive()
			switch o := o.(type) {
			case nil:
				return
			case *objects.SearchResult:
				res <- o
			}
		}
	}()

	return res, nil
}

func (client *ObjectsClient) Push(ctx context.Context, object ...astral.Object) error {
	ch, err := client.queryCh("objects.push", nil)
	if err != nil {
		return err
	}

	var errs []error

	for idx, obj := range object {
		err := ch.Send(obj)
		if err != nil {
			errs = append(errs, fmt.Errorf("write at index %d: %w", idx, err))
			break
		}

		res, err := ch.Receive()
		if err != nil {
			errs = append(errs, fmt.Errorf("read error: %w", err))
			break
		}

		switch res.(type) {
		case *astral.Ack:
			continue

		case *astral.ErrorMessage:
			errs = append(errs, fmt.Errorf("push at index %d: %w", idx, err))
		}
	}

	return errors.Join(errs...)
}

func (client *ObjectsClient) Create(repo string, alloc int) (objects.Writer, error) {
	// prepare arguments
	args := query.Args{}

	if alloc > 0 {
		args["alloc"] = alloc
	}
	if len(repo) > 0 {
		args["repo"] = repo
	}

	// send the query
	ch, err := client.queryCh("objects.create", args)
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

func (client *ObjectsClient) query(method string, args any) (*apphost.Conn, error) {
	return client.c.Query(client.target, method, args)
}

func (client *ObjectsClient) queryCh(method string, args any) (*channel.Channel, error) {
	return client.c.QueryChannel(client.target, method, args)
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
